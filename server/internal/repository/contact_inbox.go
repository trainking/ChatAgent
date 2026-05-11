package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type ContactInboxRepository struct {
	db *sqlx.DB
}

func NewContactInboxRepository(db *sqlx.DB) *ContactInboxRepository {
	return &ContactInboxRepository{db: db}
}

func (r *ContactInboxRepository) FindByPubsubToken(token string) (*model.ContactInbox, error) {
	var ci model.ContactInbox
	err := r.db.Get(&ci, "SELECT * FROM contact_inboxes WHERE pubsub_token = $1", token)
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *ContactInboxRepository) FindByContactInbox(contactID string, inboxID string) (*model.ContactInbox, error) {
	var ci model.ContactInbox
	err := r.db.Get(&ci, "SELECT * FROM contact_inboxes WHERE contact_id = $1 AND inbox_id = $2", contactID, inboxID)
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *ContactInboxRepository) FindByContact(contactID string) ([]model.ContactInbox, error) {
	var cis []model.ContactInbox
	err := r.db.Select(&cis, "SELECT * FROM contact_inboxes WHERE contact_id = $1", contactID)
	return cis, err
}
