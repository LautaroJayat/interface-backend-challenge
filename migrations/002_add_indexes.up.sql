-- Index for retrieving messages by receiver (inbox queries)
-- Supports: WHERE receiver_id = ? ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_messages_receiver_time
ON messages(receiver_id, created_at DESC);

-- Index for retrieving messages by sender (sent messages)
-- Supports: WHERE sender_id = ? ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_messages_sender_time
ON messages(sender_id, created_at DESC);

-- Composite index for chat queries (both directions)
-- Supports: WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
CREATE INDEX IF NOT EXISTS idx_messages_chat_participants
ON messages(LEAST(sender_id, receiver_id), GREATEST(sender_id, receiver_id), created_at DESC);

-- Index for unread message counts
-- Supports: WHERE receiver_id = ? AND status != 'read'
CREATE INDEX IF NOT EXISTS idx_messages_unread
ON messages(receiver_id, status, created_at)
WHERE status != 'read';

-- Index for status updates
-- Supports: WHERE sender_id = ? AND receiver_id = ? AND created_at IN (...)
CREATE INDEX IF NOT EXISTS idx_messages_status_update
ON messages(sender_id, receiver_id, created_at, status);