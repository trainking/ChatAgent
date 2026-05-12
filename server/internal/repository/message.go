package repository

import (
	"time"

	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *model.Message) error {
	now := time.Now()
	msg.CreatedAt = now
	msg.UpdatedAt = now
	if msg.Status == "" {
		msg.Status = "sent"
	}

	return r.db.QueryRow(
		`INSERT INTO messages (conversation_id, sender_type, sender_id, content, message_type, content_type, file_url, file_name, file_size, status, private, metadata, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 RETURNING id`,
		msg.ConversationID, msg.SenderType, msg.SenderID, msg.Content, msg.MessageType, msg.ContentType,
		msg.FileURL, msg.FileName, msg.FileSize, msg.Status, msg.Private, msg.Metadata, msg.CreatedAt, msg.UpdatedAt,
	).Scan(&msg.ID)
}

func (r *MessageRepository) FindByID(id string) (*model.Message, error) {
	var msg model.Message
	err := r.db.Get(&msg, "SELECT * FROM messages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepository) FindByConversation(convID string, beforeID string, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	var msgs []model.Message
	var err error

	if beforeID != "" {
		err = r.db.Select(&msgs,
			`SELECT * FROM messages
			 WHERE conversation_id = $1
			   AND created_at < (SELECT created_at FROM messages WHERE id = $2 AND conversation_id = $1)
			 ORDER BY created_at DESC LIMIT $3`,
			convID, beforeID, limit,
		)
	} else {
		err = r.db.Select(&msgs,
			"SELECT * FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT $2",
			convID, limit,
		)
	}

	if err != nil {
		return nil, err
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

func (r *MessageRepository) GetAfterID(convID string, lastID string) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.Select(&msgs,
		`SELECT * FROM messages
		 WHERE conversation_id = $1
		   AND created_at > (SELECT created_at FROM messages WHERE id = $2 AND conversation_id = $1)
		 ORDER BY created_at ASC`,
		convID, lastID,
	)
	return msgs, err
}

func (r *MessageRepository) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE messages SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}

func (r *MessageRepository) GetLastMessage(convID string) (*model.Message, error) {
	var msg model.Message
	err := r.db.Get(&msg,
		"SELECT * FROM messages WHERE conversation_id = $1 AND private = false ORDER BY created_at DESC LIMIT 1",
		convID,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepository) MarkConversationRead(convID string, userID string) error {
	_, err := r.db.Exec(
		"UPDATE messages SET status = 'read', updated_at = NOW() WHERE conversation_id = $1 AND sender_type = 'contact' AND status != 'read'",
		convID,
	)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("UPDATE conversations SET unread_count = 0 WHERE id = $1", convID)
	return err
}

func (r *MessageRepository) CountUnread(convID string) (int, error) {
	var count int
	err := r.db.Get(&count,
		"SELECT COUNT(*) FROM messages WHERE conversation_id = $1 AND sender_type = 'contact' AND status != 'read'",
		convID,
	)
	return count, err
}

func (r *MessageRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM messages WHERE id = $1", id)
	return err
}

func (r *MessageRepository) CalibrateUnreadCounts() error {
	_, err := r.db.Exec(`
		UPDATE conversations c SET unread_count = (
			SELECT COUNT(*) FROM messages m
			WHERE m.conversation_id = c.id AND m.sender_type = 'contact' AND m.status != 'read'
		) WHERE EXISTS (
			SELECT 1 FROM messages m
			WHERE m.conversation_id = c.id AND m.sender_type = 'contact' AND m.status != 'read'
		)
	`)
	return err
}
