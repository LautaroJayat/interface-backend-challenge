package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"messaging-app/e2e/testclient"
	natsAdapter "messaging-app/internal/adapters/nats"
	"messaging-app/internal/adapters/postgres"
	"messaging-app/internal/application"
	"messaging-app/internal/domain"
	"messaging-app/internal/ports"
	"messaging-app/internal/testutils"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite runs the complete application and tests real user journeys
type E2ETestSuite struct {
	suite.Suite

	// Application components
	app      *application.Application
	db       *sql.DB
	natsConn *nats.Conn
	logger   ports.Logger

	// Test configuration
	config  application.FullConfig
	baseURL string
	natsURL string

	// Test clients
	httpManager *testclient.TestUserManager
	natsManager *testclient.NATSTestManager

	// Cleanup function
	cleanup func()
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("Setting up E2E test suite with full application bootstrap...")

	// Load test configuration
	config, err := s.loadTestConfiguration()
	s.Require().NoError(err, "Failed to load test configuration")
	s.config = config

	// Setup test logger
	s.logger = testutils.NewTestLogger(s.T())

	// Verify dependencies are available
	s.verifyDependencies()

	// Initialize database
	s.initializeDatabase()

	// Initialize NATS
	s.initializeNATS()

	// Create and start application
	s.startApplication()

	// Wait for application to be ready
	s.waitForApplicationReady()

	// Initialize test clients
	s.initializeTestClients()

	s.T().Log("E2E test suite setup completed successfully")
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("Tearing down E2E test suite...")

	if s.cleanup != nil {
		s.cleanup()
	}

	// Close test clients
	if s.natsManager != nil {
		s.natsManager.CloseAll()
	}

	// Shutdown application
	if s.app != nil {
		s.app.Shutdown()
	}

	// Close connections
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.db != nil {
		s.db.Close()
	}

	s.T().Log("E2E test suite teardown completed")
}

func (s *E2ETestSuite) TearDownTest() {
	s.T().Log("Cleaning up database after test...")

	// Clean up messages table for test isolation
	_, err := s.db.Exec("TRUNCATE messages")
	s.Require().NoError(err, "Failed to truncate messages table")

	s.T().Log("Database cleanup completed")
}

func (s *E2ETestSuite) loadTestConfiguration() (application.FullConfig, error) {
	// Load base configuration
	config, err := application.LoadConfig()
	if err != nil {
		return config, err
	}

	// Override with test-specific settings
	config.Server.Port = 8081 // Use different port to avoid conflict with NATS WebSocket
	config.Server.Host = "localhost"

	// Use test database settings (assuming docker-compose setup)
	config.Database.Host = "localhost"
	config.Database.Port = 5432
	config.Database.Database = "messaging_app"
	config.Database.User = "postgres"
	config.Database.Password = "posgres" // Note: matches docker-compose.yml typo
	config.Database.SSLMode = "disable"

	// Use test NATS settings
	config.NATS.URL = "ws://localhost:8080"

	// Set test environment
	config.Environment = "test"
	config.Logging.Level = "info"

	return config, nil
}

func (s *E2ETestSuite) verifyDependencies() {
	s.T().Log("Verifying test dependencies (PostgreSQL and NATS)...")

	// Check PostgreSQL
	dbConfig := postgres.Config{
		Host:     s.config.Database.Host,
		Port:     s.config.Database.Port,
		User:     s.config.Database.User,
		Password: s.config.Database.Password,
		Database: s.config.Database.Database,
		SSLMode:  s.config.Database.SSLMode,
	}

	testDB, err := postgres.NewConnection(dbConfig, s.logger)
	s.Require().NoError(err, "PostgreSQL must be available for e2e tests. Run: docker-compose up -d postgres")
	testDB.Close()

	// Check NATS
	testNATS, err := nats.Connect(s.config.NATS.URL)
	s.Require().NoError(err, "NATS must be available for e2e tests. Run: docker-compose up -d nats")
	testNATS.Close()

	s.T().Log("All dependencies verified successfully")
}

func (s *E2ETestSuite) initializeDatabase() {
	s.T().Log("Initializing database connection...")

	dbConfig := postgres.Config{
		Host:            s.config.Database.Host,
		Port:            s.config.Database.Port,
		User:            s.config.Database.User,
		Password:        s.config.Database.Password,
		Database:        s.config.Database.Database,
		SSLMode:         s.config.Database.SSLMode,
		MaxConnections:  s.config.Database.MaxConnections,
		MaxIdleTime:     s.config.Database.MaxIdleTime,
		ConnMaxLifetime: s.config.Database.ConnMaxLifetime,
	}

	db, err := postgres.NewConnection(dbConfig, s.logger)
	s.Require().NoError(err, "Failed to connect to database")
	s.db = db

	// Verify database is ready (migrations should be applied via docker-compose)
	err = s.db.Ping()
	s.Require().NoError(err, "Database ping failed")

	s.T().Log("Database initialized successfully")
}

func (s *E2ETestSuite) initializeNATS() {
	s.T().Log("Initializing NATS connection...")

	natsConfig := natsAdapter.Config{
		URL:             s.config.NATS.URL,
		MaxReconnects:   s.config.NATS.MaxReconnects,
		ReconnectWait:   s.config.NATS.ReconnectWait,
		ConnectTimeout:  s.config.NATS.ConnectTimeout,
		RequestTimeout:  s.config.NATS.RequestTimeout,
		EnableJetStream: s.config.NATS.EnableJetStream,
		ClusterName:     s.config.NATS.ClusterName,
	}

	natsConn, err := natsAdapter.NewConnection(natsConfig, s.logger)
	s.Require().NoError(err, "Failed to connect to NATS")
	s.natsConn = natsConn

	s.T().Log("NATS initialized successfully")
}

func (s *E2ETestSuite) startApplication() {
	s.T().Log("Starting application server...")

	// Initialize adapters
	messageRepo := postgres.NewPostgreSQLMessageRepository(s.db, s.logger)
	publisher := natsAdapter.NewNATSMessagePublisher(s.natsConn, s.logger)

	// Create application
	s.app = application.NewApplication(
		s.config.GetApplicationConfig(),
		s.logger,
		messageRepo,
		publisher,
		s.config.GetHTTPConfig(),
	)

	// Initialize application
	err := s.app.Initialize()
	s.Require().NoError(err, "Failed to initialize application")

	// Start application in background
	go func() {
		if err := s.app.Start(); err != nil {
			s.T().Errorf("Application failed to start: %v", err)
		}
	}()

	// Set cleanup function
	s.cleanup = func() {
		s.app.Shutdown()
	}

	s.T().Log("Application server started")
}

func (s *E2ETestSuite) waitForApplicationReady() {
	s.T().Log("Waiting for application to be ready...")

	s.baseURL = fmt.Sprintf("http://%s:%d", s.config.Server.Host, s.config.Server.Port)
	healthURL := s.baseURL + "/health"

	// Wait up to 30 seconds for server to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.Require().Fail("Application failed to become ready within timeout")
		case <-ticker.C:
			resp, err := http.Get(healthURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				s.T().Logf("Application ready at %s", s.baseURL)
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

func (s *E2ETestSuite) initializeTestClients() {
	s.T().Log("Initializing test clients...")

	// Initialize HTTP test client manager
	s.httpManager = testclient.NewTestUserManager(s.baseURL)

	// Initialize NATS test client manager
	s.natsURL = "ws://localhost:8080" // Use direct NATS connection for now
	s.natsManager = testclient.NewNATSTestManager(s.natsURL)

	// Verify HTTP client works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test client and verify health endpoint
	testClient := testclient.NewClient(testclient.Config{
		BaseURL: s.baseURL,
		Timeout: 5 * time.Second,
	})

	err := testClient.Health(ctx)
	s.Require().NoError(err, "HTTP test client health check failed")

	s.T().Log("Test clients initialized successfully")
}

// Helper method to get a fresh HTTP client for a user
func (s *E2ETestSuite) GetHTTPClient(userID, email, handler string) *testclient.Client {
	config := testclient.Config{
		BaseURL: s.baseURL,
		Timeout: 30 * time.Second,
		UserContext: domain.UserContext{
			UserID:  userID,
			Email:   email,
			Handler: handler,
		},
	}
	return testclient.NewClient(config)
}

// Helper method to get a NATS client for a user
func (s *E2ETestSuite) GetNATSClient(userID string) (*testclient.NATSClient, error) {
	return s.natsManager.GetClient(userID)
}

// Helper method to create test users as needed during journeys
func (s *E2ETestSuite) CreateTestUser(userID, email, handler string) *testclient.Client {
	return s.GetHTTPClient(userID, email, handler)
}
