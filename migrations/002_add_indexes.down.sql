-- Drop all indexes
DROP INDEX IF EXISTS idx_messages_receiver_time;
DROP INDEX IF EXISTS idx_messages_sender_time;
DROP INDEX IF EXISTS idx_messages_chat_participants;
DROP INDEX IF EXISTS idx_messages_unread;
DROP INDEX IF EXISTS idx_messages_status_update;