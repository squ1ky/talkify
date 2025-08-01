CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_receiver_id ON messages(receiver_id);
CREATE INDEX idx_messages_conversation ON messages(sender_id, receiver_id, created_at);
CREATE INDEX idx_messages_created_at ON messages(created_at);

ALTER TABLE messages
ADD CONSTRAINT chk_message_content_not_empty
CHECK (LENGTH(TRIM(content)) > 0);

ALTER TABLE messages
ADD CONSTRAINT chk_messages_content_length
CHECK (LENGTH(content) <= 1000);