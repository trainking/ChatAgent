package model

import "time"

type Conversation struct {
	ID            string     `db:"id" json:"id"`
	DisplayID     int        `db:"display_id" json:"display_id"`
	InboxID       string     `db:"inbox_id" json:"inbox_id"`
	ContactID     string     `db:"contact_id" json:"contact_id"`
	AssigneeID    *string    `db:"assignee_id" json:"assignee_id"`
	Status        string     `db:"status" json:"status"`
	Priority      string     `db:"priority" json:"priority"`
	Subject       string     `db:"subject" json:"subject"`
	UnreadCount   int        `db:"unread_count" json:"unread_count"`
	LastMessageAt time.Time  `db:"last_message_at" json:"last_message_at"`
	WaitingSince  *time.Time `db:"waiting_since" json:"waiting_since"`
	FirstReplyAt  *time.Time `db:"first_reply_at" json:"first_reply_at"`
	ResolvedAt    *time.Time `db:"resolved_at" json:"resolved_at"`
	SnoozedUntil  *time.Time `db:"snoozed_until" json:"snoozed_until"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`

	Contact       *Contact `db:"-" json:"contact,omitempty"`
	Assignee      *User    `db:"-" json:"assignee,omitempty"`
	LastMessage   *Message `db:"-" json:"last_message,omitempty"`
}
