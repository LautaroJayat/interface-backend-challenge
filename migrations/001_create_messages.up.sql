-- Create messages table with composite primary key
CREATE TABLE IF NOT EXISTS messages (
    sender_id TEXT NOT NULL,
    receiver_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    content TEXT NOT NULL,
    status TEXT DEFAULT 'sent' NOT NULL,

    -- Composite primary key for natural idempotency
    PRIMARY KEY (sender_id, receiver_id, created_at),

    -- Constraints
    CONSTRAINT messages_status_check CHECK (status IN ('sent', 'delivered', 'read')),
    CONSTRAINT messages_sender_not_empty CHECK (LENGTH(TRIM(sender_id)) > 0),
    CONSTRAINT messages_receiver_not_empty CHECK (LENGTH(TRIM(receiver_id)) > 0),
    CONSTRAINT messages_content_not_empty CHECK (LENGTH(TRIM(content)) > 0),
    CONSTRAINT messages_different_users CHECK (sender_id != receiver_id)
);

-- Add table comment
COMMENT ON TABLE messages IS 'Stores 1:1 chat messages with composite key for idempotency';
COMMENT ON COLUMN messages.sender_id IS 'User ID of message sender';
COMMENT ON COLUMN messages.receiver_id IS 'User ID of message receiver';
COMMENT ON COLUMN messages.created_at IS 'Message creation timestamp (part of PK)';
COMMENT ON COLUMN messages.content IS 'Message text content';
COMMENT ON COLUMN messages.status IS 'Message delivery status: sent, delivered, read';