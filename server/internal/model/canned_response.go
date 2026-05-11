package model

import "time"

type CannedResponse struct {
	ID        string    `db:"id" json:"id"`
	InboxID   *string   `db:"inbox_id" json:"inbox_id"`
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	CreatedBy string    `db:"created_by" json:"created_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
