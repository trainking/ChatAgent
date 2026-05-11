package model

import "time"

type Message struct {
	ID             string    `db:"id" json:"id"`
	ConversationID string    `db:"conversation_id" json:"conversation_id"`
	SenderType     string    `db:"sender_type" json:"sender_type"`
	SenderID       string    `db:"sender_id" json:"sender_id"`
	Content        string    `db:"content" json:"content"`
	MessageType    string    `db:"message_type" json:"message_type"`
	ContentType    string    `db:"content_type" json:"content_type"`
	FileURL        string    `db:"file_url" json:"file_url"`
	FileName       string    `db:"file_name" json:"file_name"`
	FileSize       int64     `db:"file_size" json:"file_size"`
	Status         string    `db:"status" json:"status"`
	Private        bool      `db:"private" json:"private"`
	Metadata       string    `db:"metadata" json:"metadata"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type MessageListQuery struct {
	BeforeID string `json:"before_id"`
	Limit    int    `json:"limit"`
}
