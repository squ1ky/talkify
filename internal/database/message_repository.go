package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/squ1ky/talkify/internal/models"
	"time"
)

// MessageRepository handles database operations for messages
type MessageRepository struct {
	db *DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message in the database
func (mr *MessageRepository) Create(message *models.Message) error {
	query := `
		INSERT INTO messages (sender_id, receiver_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := mr.db.QueryRow(
		query,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		message.CreatedAt,
	).Scan(&message.ID)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by ID
func (mr *MessageRepository) GetByID(id int) (*models.Message, error) {
	message := &models.Message{}
	query := `
		SELECT id, sender_id, receiver_id, content, created_at
		FROM messages
		WHERE id = $1`

	err := mr.db.QueryRow(query, id).Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	return message, nil
}

// GetConversationHistory returns history between two users
func (mr *MessageRepository) GetConversationHistory(userID1, userID2 int, limit, offset int) ([]models.MessageWithUserResponse, error) {
	query := `
		SELECT
			m.id, m.content, m.created_at,
			s.id as sender_id, s.username as sender_username, s.created_at as sender_created_at,
			r.id as receiver_id, r.username as receiver_username, r.created_at as receiver_created_at
		FROM messages m
		INNER JOIN users s ON m.sender_id = s.id
		INNER JOIN users r ON m.receiver_id = r.id
		WHERE
			(m.sender_id = $1 AND m.receiver_id = $2) OR
			(m.sender_id = $2 AND m.receiver_id = $1
		ORDER BY m.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := mr.db.Query(query, userID1, userID2, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}
	defer rows.Close()

	var messages []models.MessageWithUserResponse
	for rows.Next() {
		var msg models.MessageWithUserResponse
		var sender models.UserResponse
		var receiver models.UserResponse

		err := rows.Scan(
			&msg.ID, &msg.Content, &msg.CreatedAt,
			&sender.ID, &sender.Username, &sender.CreatedAt,
			&receiver.ID, &receiver.Username, &receiver.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		msg.Sender = sender
		msg.Receiver = receiver
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, nil
}

// GetUserMessages returns all messages for a specific user (sent and received)
func (mr *MessageRepository) GetUserMessages(userID int, limit, offset int) ([]models.MessageWithUserResponse, error) {
	query := `
		SELECT
			m.id, m.content, m.created_at,
			s.id as sender_id, s.username as sender_username, s.created_at as sender_created_at,
			r.id as receiver_id, r.username as receiver_username, r.created_at as receiver_created_at
		FROM messages m
		INNER JOIN users s on m.sender_id = s.id
		INNER JOIN users r on m.receiver_id = r.id
		WHERE m.sender_id = $1 OR m.receiver_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := mr.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user messages: %w", err)
	}
	defer rows.Close()

	var messages []models.MessageWithUserResponse
	for rows.Next() {
		var msg models.MessageWithUserResponse
		var sender models.UserResponse
		var receiver models.UserResponse

		err := rows.Scan(
			&msg.ID, &msg.Content, &msg.CreatedAt,
			&sender.ID, &sender.Username, &sender.CreatedAt,
			&receiver.ID, &receiver.Username, &receiver.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		msg.Sender = sender
		msg.Receiver = receiver
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user messages: %w", err)
	}

	return messages, nil
}

// GetRecentConversations returns list of users with recent conversations
func (mr *MessageRepository) GetRecentConversations(userID int, limit int) ([]models.UserResponse, error) {
	query := `
		SELECT DISTINCT
			CASE
				WHEN m.sender_id = $1 THEN u2.id
				ELSE u1.id
			END as user_id,
			CASE
				WHEN m.sender_id = $1 THEN u2.username
				ELSE u1.username
			END as username,
			CASE
				WHEN m.sender_id = $1 THEN u2.created_at
				ELSE u1.created_at
			END as user_created_at
		FROM messages m
		INNER JOIN users u1 on m.sender_id = u1.id
		INNER JOIN users u2 ON m.receiver_id = u2.id
		WHERE m.sender_id = $1 OR m.receiver_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2`

	rows, err := mr.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent conversations: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.UserResponse
		err := rows.Scan(&user.ID, user.Username, &user.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation row: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversation rows: %w", err)
	}

	return users, nil
}

// CountConversationMessages returns total number of messages between two users
func (mr *MessageRepository) CountConversationMessages(userID1, userID2 int) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM messages
		WHERE
			(sender_id = $1 AND receiver_id = $2) OR
			(sender_id = $2 AND receiver_id = $1)`

	err := mr.db.QueryRow(query, userID1, userID2).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count conversation messages: %w", err)
	}

	return count, nil
}

// GetMessagesSince returns messages sent after specific time
func (mr *MessageRepository) GetMessagesSince(userID int, since time.Time) ([]models.MessageWithUserResponse, error) {
	query := `
		SELECT
			m.id, m.content, m.created_at,
			s.id as sender_id, s.username as sender_username, s.created_at as sender_created_at,
			r.id as receiver_id, r.username as receiver_username, r.created_at as receiver_created_at
		FROM messages m
		INNER JOIN users s on m.sender_id = s.id
		INNER JOIN users r on m.receiver_id = r.id
		WHERE
			(m.sender_id = $1 OR m.receiver_id = $1) AND
			m.created_at > $2
		ORDER BY m.created_at ASC`

	rows, err := mr.db.Query(query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages since: %w", err)
	}
	defer rows.Close()

	var messages []models.MessageWithUserResponse
	for rows.Next() {
		var msg models.MessageWithUserResponse
		var sender models.UserResponse
		var receiver models.UserResponse

		err := rows.Scan(
			&msg.ID, &msg.Content, &msg.CreatedAt,
			&sender.ID, &sender.Username, &sender.CreatedAt,
			&receiver.ID, &receiver.Username, &receiver.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		msg.Sender = sender
		msg.Receiver = receiver
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, nil
}

// Delete removes a message by ID
func (mr *MessageRepository) Delete(id int) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := mr.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}
