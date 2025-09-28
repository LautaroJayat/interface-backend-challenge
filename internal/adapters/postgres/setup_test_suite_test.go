package postgres_test

import (
	"database/sql"
	"messaging-app/internal/adapters/postgres"
	"messaging-app/internal/testutils"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	db   *sql.DB
	repo *postgres.PostgreSQLMessageRepository
}

func (s *TestSuite) TearDownTest() {
	_, err := s.db.Exec("TRUNCATE messages")
	s.Require().NoError(err)
}

func (s *TestSuite) SetupSuite() {
	db := setupTestDB(s.T())
	s.db = db
	s.repo = postgres.NewPostgreSQLMessageRepository(s.db, &testutils.TestLogger{T: s.T()})

}

func (s *TestSuite) TearDownSuite() {
	s.db.Close()
}

func TestPostgresAdaptersSuiteIntegration(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
