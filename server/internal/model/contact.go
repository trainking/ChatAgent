package model

import "time"

type Contact struct {
	ID                  string    `db:"id" json:"id"`
	Name                string    `db:"name" json:"name"`
	Email               string    `db:"email" json:"email"`
	Phone               string    `db:"phone" json:"phone"`
	AvatarURL           string    `db:"avatar_url" json:"avatar_url"`
	ContactType         string    `db:"contact_type" json:"contact_type"`
	Blocked             bool      `db:"blocked" json:"blocked"`
	BrowserFingerprints string    `db:"browser_fingerprints" json:"browser_fingerprints"`
	Country             string    `db:"country" json:"country"`
	City                string    `db:"city" json:"city"`
	Browser             string    `db:"browser" json:"browser"`
	OS                  string    `db:"os" json:"os"`
	CustomAttrs         string    `db:"custom_attrs" json:"custom_attrs"`
	LastActivityAt      time.Time `db:"last_activity_at" json:"last_activity_at"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

type ContactInbox struct {
	ContactID     string    `db:"contact_id" json:"contact_id"`
	InboxID       string    `db:"inbox_id" json:"inbox_id"`
	SourceID      string    `db:"source_id" json:"source_id"`
	PubsubToken   string    `db:"pubsub_token" json:"pubsub_token"`
	HMACVerified  bool      `db:"hmac_verified" json:"hmac_verified"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
