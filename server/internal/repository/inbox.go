package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type InboxRepository struct {
	db *sqlx.DB
}

func NewInboxRepository(db *sqlx.DB) *InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Create(inbox *model.Inbox) error {
	query := `INSERT INTO inboxes (name, description, icon, welcome_title, welcome_message, inbox_type, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, inbox.Name, inbox.Description, inbox.Icon, inbox.WelcomeTitle,
		inbox.WelcomeMessage, inbox.InboxType, inbox.Status, inbox.CreatedBy).
		Scan(&inbox.ID, &inbox.CreatedAt, &inbox.UpdatedAt)
}

func (r *InboxRepository) FindByID(id string) (*model.Inbox, error) {
	inbox := &model.Inbox{}
	query := `SELECT id, name, description, icon, welcome_title, welcome_message, inbox_type, status, created_by,
		created_at, updated_at FROM inboxes WHERE id = $1`
	err := r.db.Get(inbox, query, id)
	if err != nil {
		return nil, err
	}
	return inbox, nil
}

func (r *InboxRepository) FindByName(name string) (*model.Inbox, error) {
	inbox := &model.Inbox{}
	query := `SELECT id, name, description, icon, welcome_title, welcome_message, inbox_type, status, created_by,
		created_at, updated_at FROM inboxes WHERE name = $1`
	err := r.db.Get(inbox, query, name)
	if err != nil {
		return nil, err
	}
	return inbox, nil
}

func (r *InboxRepository) ListByUser(userID string, isAdmin bool) ([]model.Inbox, error) {
	var inboxes []model.Inbox
	var query string
	var args []interface{}

	if isAdmin {
		query = `SELECT i.id, i.name, i.description, i.icon, i.welcome_title, i.welcome_message, i.inbox_type, i.status, i.created_by,
			i.created_at, i.updated_at, u.name AS creator_name, u.email AS creator_email
			FROM inboxes i LEFT JOIN users u ON i.created_by = u.id ORDER BY i.created_at DESC`
	} else {
		query = `SELECT DISTINCT i.id, i.name, i.description, i.icon, i.welcome_title, i.welcome_message, i.inbox_type, i.status, i.created_by,
			i.created_at, i.updated_at, u.name AS creator_name, u.email AS creator_email
			FROM inboxes i
			LEFT JOIN users u ON i.created_by = u.id
			LEFT JOIN inbox_collaborators c ON i.id = c.inbox_id
			WHERE i.created_by = $1 OR c.user_id = $1
			ORDER BY i.created_at DESC`
		args = append(args, userID)
	}

	err := r.db.Select(&inboxes, query, args...)
	if err != nil {
		return nil, err
	}
	return inboxes, nil
}

func (r *InboxRepository) Update(inbox *model.Inbox) error {
	_, err := r.db.Exec(
		`UPDATE inboxes SET description=$1, icon=$2, welcome_title=$3, welcome_message=$4, status=$5, updated_at=NOW() WHERE id=$6`,
		inbox.Description, inbox.Icon, inbox.WelcomeTitle, inbox.WelcomeMessage, inbox.Status, inbox.ID,
	)
	return err
}

func (r *InboxRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM inboxes WHERE id = $1", id)
	return err
}

func (r *InboxRepository) AddCollaborator(inboxID, userID string) error {
	query := `INSERT INTO inbox_collaborators (inbox_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, inboxID, userID)
	return err
}

func (r *InboxRepository) RemoveAllCollaborators(inboxID string) error {
	_, err := r.db.Exec("DELETE FROM inbox_collaborators WHERE inbox_id = $1", inboxID)
	return err
}

func (r *InboxRepository) GetCollaborators(inboxID string) ([]model.User, error) {
	var users []model.User
	query := `SELECT u.id, u.email, u.name, u.role, u.avatar_url, u.status, u.must_change_password,
		u.last_login_at, u.online_status, u.created_at, u.updated_at
		FROM users u INNER JOIN inbox_collaborators c ON u.id = c.user_id
		WHERE c.inbox_id = $1 ORDER BY u.name`
	err := r.db.Select(&users, query, inboxID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *InboxRepository) GetCreator(inboxID string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT u.id, u.email, u.name, u.role, u.avatar_url, u.status, u.must_change_password,
		u.last_login_at, u.online_status, u.created_at, u.updated_at
		FROM users u INNER JOIN inboxes i ON u.id = i.created_by
		WHERE i.id = $1`
	err := r.db.Get(user, query, inboxID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
