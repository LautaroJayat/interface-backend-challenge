package nats_test

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"messaging-app/internal/adapters/nats"
	"messaging-app/internal/domain"
	"messaging-app/internal/ports"
	"messaging-app/internal/testutils"
	"messaging-app/testdata"

	natsgo "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	conn      *natsgo.Conn
	publisher *nats.NATSMessagePublisher
}

func (s *TestSuite) TearDownTest() {
	// Nothing to clean up for NATS after each test
}

func (s *TestSuite) SetupSuite() {
	conn := setupTestNATS(s.T())
	s.conn = conn
	s.publisher = nats.NewNATSMessagePublisher(conn, testutils.NewTestLogger(s.T()))
}

func (s *TestSuite) TearDownSuite() {
	s.conn.Close()
}

func TestNATSAdaptersSuiteIntegration(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// --- helpers ---
func setupTestNATS(t *testing.T) *natsgo.Conn {
	t.Helper()

	cfg := nats.DefaultConfig()

	// Allow overriding config via env
	if url := os.Getenv("NATS_URL"); url != "" {
		cfg.URL = url
	}

	if maxReconnectsStr := os.Getenv("NATS_MAX_RECONNECTS"); maxReconnectsStr != "" {
		maxReconnects, err := strconv.Atoi(maxReconnectsStr)
		if err != nil {
			t.Logf("could not convert NATS_MAX_RECONNECTS into integer, %s. err=%s", maxReconnectsStr, err)
			t.FailNow()
		}
		cfg.MaxReconnects = maxReconnects
	}

	conn, err := nats.NewConnection(cfg, testutils.NewTestLogger(t))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	return conn
}

func TestNewConnectionIntegration(t *testing.T) {
	conn := setupTestNATS(t)
	defer conn.Close()

	// Basic connection test
	if !conn.IsConnected() {
		t.Fatalf("expected connection to be established")
	}

	// Test basic publish/subscribe
	subject := "test.connection"
	received := make(chan bool, 1)

	// Subscribe to test subject
	sub, err := conn.Subscribe(subject, func(msg *natsgo.Msg) {
		received <- true
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	// Wait for subscription to be ready
	if err := conn.Flush(); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	// Publish a test message
	if err := conn.Publish(subject, []byte("test message")); err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// Wait for message to be received
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case <-received:
		// Message received successfully
	case <-ctx.Done():
		t.Fatalf("timeout waiting for message")
	}
}

func (s *TestSuite) TestPublishMessage() {
	ctx := context.Background()

	// Use testdata for consistent test users and messages
	alice := testdata.Alice
	bob := testdata.Bob
	validMessages := testdata.ValidMessages()
	testMessage := validMessages[0] // Alice to Bob message

	subject := domain.GetMessageTopic(testMessage.ReceiverID)

	// Subscribe to the message topic to capture published messages
	received := make(chan *domain.MessageEnvelope, 1)
	sub, err := s.conn.Subscribe(subject, func(msg *natsgo.Msg) {
		var envelope domain.MessageEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			s.T().Errorf("failed to unmarshal message envelope: %v", err)
			return
		}
		received <- &envelope
	})
	s.Require().NoError(err)
	defer sub.Unsubscribe()

	// Wait for subscription to be ready
	s.Require().NoError(s.conn.Flush())

	// Publish message using the publisher
	err = s.publisher.PublishMessage(ctx, testMessage)
	s.Require().NoError(err)

	// Wait for message to be received
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case envelope := <-received:
		s.Equal(domain.MessageTypeNewMessage, envelope.Type)
		s.Equal(testMessage.SenderID, envelope.Data.SenderID)
		s.Equal(testMessage.ReceiverID, envelope.Data.ReceiverID)
		s.Equal(testMessage.Content, envelope.Data.Content)
		s.WithinDuration(time.Now().UTC(), envelope.Timestamp, 5*time.Second)

		// Verify we're using the correct test users
		s.Equal(alice.UserID, envelope.Data.SenderID)
		s.Equal(bob.UserID, envelope.Data.ReceiverID)
	case <-ctx.Done():
		s.FailNow("timeout waiting for published message")
	}
}

func (s *TestSuite) TestPublishStatusUpdate() {
	ctx := context.Background()

	// Use testdata for consistent test users and create realistic status update
	alice := testdata.Alice
	validMessages := testdata.ValidMessages()
	baseMessage := validMessages[0] // Alice to Bob message

	userID := alice.UserID
	subject := domain.GetStatusTopic(userID)

	// Create test status update using message from testdata
	testStatusUpdate := ports.StatusUpdate{
		MessageID: domain.MessageID{
			SenderID:   baseMessage.SenderID,
			ReceiverID: baseMessage.ReceiverID,
			CreatedAt:  baseMessage.CreatedAt,
		},
		Status:    domain.MessageStatusDelivered,
		UpdatedBy: "system",
		UpdatedAt: time.Now().UTC(),
	}

	// Subscribe to the status topic to capture published updates
	received := make(chan *domain.StatusUpdateEnvelope, 1)
	sub, err := s.conn.Subscribe(subject, func(msg *natsgo.Msg) {
		var envelope domain.StatusUpdateEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			s.T().Errorf("failed to unmarshal status update envelope: %v", err)
			return
		}
		received <- &envelope
	})
	s.Require().NoError(err)
	defer sub.Unsubscribe()

	// Wait for subscription to be ready
	s.Require().NoError(s.conn.Flush())

	// Publish status update using the publisher
	err = s.publisher.PublishStatusUpdate(ctx, userID, testStatusUpdate)
	s.Require().NoError(err)

	// Wait for status update to be received
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case envelope := <-received:
		s.Equal(domain.MessageTypeStatusUpdate, envelope.Type)
		s.WithinDuration(time.Now().UTC(), envelope.Timestamp, 5*time.Second)

		// Convert the data back to StatusUpdate for comparison
		dataBytes, err := json.Marshal(envelope.Data)
		s.Require().NoError(err)

		var receivedStatusUpdate ports.StatusUpdate
		err = json.Unmarshal(dataBytes, &receivedStatusUpdate)
		s.Require().NoError(err)

		s.Equal(testStatusUpdate.MessageID, receivedStatusUpdate.MessageID)
		s.Equal(testStatusUpdate.Status, receivedStatusUpdate.Status)
		s.Equal(testStatusUpdate.UpdatedBy, receivedStatusUpdate.UpdatedBy)
		s.WithinDuration(testStatusUpdate.UpdatedAt, receivedStatusUpdate.UpdatedAt, time.Second)

		// Verify we're using the correct test user
		s.Equal(alice.UserID, userID)
	case <-ctx.Done():
		s.FailNow("timeout waiting for published status update")
	}
}
