package model

import "time"

type ActivityLog struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Event     string    `db:"event" json:"event"`
	Module    string    `db:"module" json:"module"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
