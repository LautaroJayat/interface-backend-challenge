package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"

	"messaging-app/internal/domain"
	"messaging-app/internal/ports"
)

type PostgreSQLMessageRepository struct {
	db     *sql.DB
	logger ports.Logger
}

func NewPostgreSQLMessageRepository(db *sql.DB, logger ports.Logger) *PostgreSQLMessageRepository {
	return &PostgreSQLMessageRepository{
		db:     db,
		logger: logger,
	}
}

// SaveMessage implements ports.MessageRepository
func (r *PostgreSQLMessageRepository) SaveMessage(ctx context.Context, message domain.Message) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	query := `
        INSERT INTO messages (sender_id, receiver_id, created_at, content, status)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := r.db.ExecContext(ctx, query,
		message.SenderID,
		message.ReceiverID,
		message.CreatedAt,
		message.Content,
		message.Status,
	)

	if err != nil {
		// Check for duplicate key error (PostgreSQL error code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrDuplicateMessage
		}
		return fmt.Errorf("failed to save message: %w", err)
	}

	r.logger.Debug("Message saved", "sender", message.SenderID, "receiver", message.ReceiverID)
	return nil
}

// GetMessages implements ports.MessageRepository
func (r *PostgreSQLMessageRepository) GetMessages(ctx context.Context, chatID string, cursor time.Time, limit int) ([]domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	// Parse chat ID to get participants
	participants := strings.Split(chatID, "---")
	if len(participants) != 2 {
		return nil, fmt.Errorf("invalid chat ID format: %s", chatID)
	}

	user1, user2 := participants[0], participants[1]

	var query string
	var args []interface{}

	if cursor.IsZero() {
		// First page - no cursor
		query = `
            SELECT sender_id, receiver_id, created_at, content, status
            FROM messages
            WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
            ORDER BY created_at DESC
            LIMIT $3
        `
		args = []interface{}{user1, user2, limit}
	} else {
		// Subsequent pages - use cursor
		query = `
            SELECT sender_id, receiver_id, created_at, content, status
            FROM messages
            WHERE ((sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1))
              AND created_at < $3
            ORDER BY created_at DESC
            LIMIT $4
        `
		args = []interface{}{user1, user2, cursor, limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.CreatedAt,
			&msg.Content,
			&msg.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	r.logger.Debug("Retrieved messages", "chat_id", chatID, "count", len(messages))
	return messages, nil
}

func (r *PostgreSQLMessageRepository) GetChatSessions(ctx context.Context, userID string) ([]domain.ChatSession, error) {
	// Step 1: Get distinct participants
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT
			CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END as other_participant
		FROM messages
		WHERE sender_id = $1 OR receiver_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get participants: %w", err)
	}
	defer rows.Close()

	var participants []string
	for rows.Next() {
		var participant string
		if err := rows.Scan(&participant); err != nil {
			return nil, fmt.Errorf("scan participant: %w", err)
		}
		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iter participants: %w", err)
	}

	// Step 2: Loop over participants and fetch session info
	sessions := make([]domain.ChatSession, 0, len(participants))

	for _, participant := range participants {
		session := domain.ChatSession{
			OtherParticipant: participant,
			ChatID:           domain.ComputeChatID(userID, participant),
		}

		// Step 2a: Get unread count
		if err := r.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM messages
			WHERE sender_id = $1 AND receiver_id = $2 AND status != 'read'
		`, participant, userID).Scan(&session.UnreadCount); err != nil {
			return nil, fmt.Errorf("get unread count for %s: %w", participant, err)
		}

		// Step 2b: Get last message
		var lastMsg sql.NullString
		var lastBy sql.NullString
		var lastAt sql.NullTime

		err := r.db.QueryRowContext(ctx, `
			SELECT content, sender_id, created_at
			FROM messages
			WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
			ORDER BY created_at DESC
			LIMIT 1
		`, userID, participant).Scan(&lastMsg, &lastBy, &lastAt)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("get last message for %s: %w", participant, err)
		}

		session.LastMessage = lastMsg.String
		session.LastMessageBy = lastBy.String
		session.LastMessageAt = lastAt.Time

		sessions = append(sessions, session)
	}

	// Step 3: Sort by last message timestamp descending
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastMessageAt.After(sessions[j].LastMessageAt)
	})

	r.logger.Debug("Retrieved chat sessions", "user_id", userID, "count", len(sessions))
	return sessions, nil
}

func (r *PostgreSQLMessageRepository) MarkMessagesUpToRead(ctx context.Context, msg domain.MessageID) (int64, error) {

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE messages
		SET status = $4
		WHERE receiver_id = $1
		  AND sender_id = $2
		  AND created_at <= $3
		  AND status != $4
	`

	res, err := tx.ExecContext(ctx, query, msg.ReceiverID, msg.SenderID, msg.CreatedAt, domain.MessageStatusRead)
	if err != nil {
		return 0, fmt.Errorf("update messages: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	r.logger.Debug("Marked messages as read", "receiver", msg.ReceiverID, "sender", msg.SenderID, "count", affected)
	return affected, nil
}

// GetMessageByID implements ports.MessageRepository
func (r *PostgreSQLMessageRepository) GetMessageByID(ctx context.Context, messageID domain.MessageID) (*domain.Message, error) {
	query := `
        SELECT sender_id, receiver_id, created_at, content, status
        FROM messages
        WHERE sender_id = $1 AND receiver_id = $2 AND created_at = $3
    `

	var msg domain.Message
	err := r.db.QueryRowContext(ctx, query, messageID.SenderID, messageID.ReceiverID, messageID.CreatedAt).Scan(
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.CreatedAt,
		&msg.Content,
		&msg.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &msg, nil
}

// GetUnreadCount implements ports.MessageRepository
func (r *PostgreSQLMessageRepository) GetUnreadCount(ctx context.Context, userID, chatID string) (int, error) {
	participants := strings.Split(chatID, "---")
	if len(participants) != 2 {
		return 0, fmt.Errorf("invalid chat ID format: %s", chatID)
	}

	user1, user2 := participants[0], participants[1]
	otherUser := user1
	if user1 == userID {
		otherUser = user2
	}

	query := `
        SELECT COUNT(*)
        FROM messages
        WHERE sender_id = $1 AND receiver_id = $2 AND status != 'read'
    `

	var count int
	err := r.db.QueryRowContext(ctx, query, otherUser, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// MarkChatAsRead implements ports.MessageRepository
func (r *PostgreSQLMessageRepository) MarkChatAsRead(ctx context.Context, userID, chatID string) error {
	participants := strings.Split(chatID, "---")
	if len(participants) != 2 {
		return fmt.Errorf("invalid chat ID format: %s", chatID)
	}

	user1, user2 := participants[0], participants[1]
	otherUser := user1
	if user1 == userID {
		otherUser = user2
	}

	query := `
        UPDATE messages
        SET status = 'read'
        WHERE sender_id = $1 AND receiver_id = $2 AND status != 'read'
    `

	_, err := r.db.ExecContext(ctx, query, otherUser, userID)
	if err != nil {
		return fmt.Errorf("failed to mark chat as read: %w", err)
	}

	r.logger.Debug("Marked chat as read", "user_id", userID, "chat_id", chatID)
	return nil
}
