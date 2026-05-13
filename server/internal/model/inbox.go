package model

import "time"

type Inbox struct {
	ID             string    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Description    string    `db:"description" json:"description"`
	Icon           string    `db:"icon" json:"icon"`
	WelcomeTitle   string    `db:"welcome_title" json:"welcome_title"`
	WelcomeMessage string    `db:"welcome_message" json:"welcome_message"`
	InboxType      string    `db:"inbox_type" json:"inbox_type"`
	Status         string    `db:"status" json:"status"`
	CreatedBy      string    `db:"created_by" json:"created_by"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
	CreatorName    string    `db:"creator_name" json:"creator_name"`
	CreatorEmail   string    `db:"creator_email" json:"creator_email"`
	// Populated on detail read
	Creator       *User  `db:"-" json:"creator,omitempty"`
	Collaborators []User `db:"-" json:"collaborators,omitempty"`
}

type InboxCollaborator struct {
	InboxID string `db:"inbox_id" json:"inbox_id"`
	UserID  string `db:"user_id" json:"user_id"`
}
