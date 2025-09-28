package postgres_test

import (
	"context"
	"time"

	_ "github.com/lib/pq"

	"messaging-app/internal/domain"
	"messaging-app/testdata"
)

func (s *TestSuite) TestMessageRepositoryIntegration() {
	ctx := context.Background()

	msg := domain.Message{
		SenderID:   testdata.Alice.UserID,
		ReceiverID: testdata.Bob.UserID,
		CreatedAt:  time.Now().UTC().Truncate(time.Microsecond),
		Content:    "Hello, Bob!",
		Status:     domain.MessageStatusSent,
	}

	// SaveMessage
	err := s.repo.SaveMessage(ctx, msg)
	s.Require().NoError(err)

	// GetMessageByID
	got, err := s.repo.GetMessageByID(ctx, domain.MessageID{
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		CreatedAt:  msg.CreatedAt,
	})
	s.Require().NoError(err)
	s.Require().Equal(msg.Content, got.Content)

	// GetMessages
	messages, err := s.repo.GetMessages(ctx, domain.ComputeChatID(testdata.Alice.UserID, testdata.Bob.UserID), time.Time{}, 10)
	s.Require().NoError(err)
	s.Require().Len(messages, 1)

	// GetChatSessions
	sessions, err := s.repo.GetChatSessions(ctx, testdata.Bob.UserID)
	s.Require().NoError(err)
	s.Require().Len(sessions, 1)

	// GetUnreadCount
	count, err := s.repo.GetUnreadCount(ctx, testdata.Bob.UserID, domain.ComputeChatID(testdata.Alice.UserID, testdata.Bob.UserID))
	s.Require().NoError(err)
	s.Require().Equal(1, count)

	// UpdateMessageStatus
	_, err = s.repo.MarkMessagesUpToRead(ctx, domain.MessageID{
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		CreatedAt:  msg.CreatedAt,
	})
	s.Require().NoError(err)

	// Verify status update
	got2, err := s.repo.GetMessageByID(ctx, domain.MessageID{
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		CreatedAt:  msg.CreatedAt,
	})
	s.Require().NoError(err)
	s.Require().Equal(domain.MessageStatusRead, got2.Status)

	// MarkChatAsRead (should not error even if already read)
	err = s.repo.MarkChatAsRead(ctx, testdata.Bob.UserID, domain.ComputeChatID(testdata.Alice.UserID, testdata.Bob.UserID))
	s.Require().NoError(err)
}
