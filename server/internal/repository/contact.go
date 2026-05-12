package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type ContactRepository struct {
	db *sqlx.DB
}

func NewContactRepository(db *sqlx.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(contact *model.Contact) error {
	now := time.Now()
	contact.CreatedAt = now
	contact.UpdatedAt = now
	if contact.LastActivityAt.IsZero() {
		contact.LastActivityAt = now
	}

	return r.db.QueryRow(
		`INSERT INTO contacts (name, email, phone, avatar_url, contact_type, blocked, browser_fingerprints, country, city, browser, os, custom_attrs, last_activity_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING id`,
		contact.Name, contact.Email, contact.Phone, contact.AvatarURL, contact.ContactType, contact.Blocked,
		contact.BrowserFingerprints, contact.Country, contact.City, contact.Browser, contact.OS,
		contact.CustomAttrs, contact.LastActivityAt, contact.CreatedAt, contact.UpdatedAt,
	).Scan(&contact.ID)
}

func (r *ContactRepository) FindByID(id string) (*model.Contact, error) {
	var c model.Contact
	err := r.db.Get(&c, "SELECT * FROM contacts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContactRepository) FindByEmail(email string) (*model.Contact, error) {
	var c model.Contact
	err := r.db.Get(&c, "SELECT * FROM contacts WHERE email = $1 AND email != ''", email)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContactRepository) FindByFingerprint(fingerprint string) (*model.Contact, error) {
	var c model.Contact
	err := r.db.Get(&c, "SELECT * FROM contacts WHERE browser_fingerprints::jsonb @> $1 LIMIT 1", fmt.Sprintf(`["%s"]`, fingerprint))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type ContactFilter struct {
	InboxID  string
	Type     string
	Search   string
	Page     int
	PageSize int
}

func (r *ContactRepository) List(filter ContactFilter) ([]model.Contact, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.InboxID != "" {
		where += fmt.Sprintf(" AND id IN (SELECT contact_id FROM contact_inboxes WHERE inbox_id = $%d)", argIdx)
		args = append(args, filter.InboxID)
		argIdx++
	}
	if filter.Type != "" {
		where += fmt.Sprintf(" AND contact_type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", argIdx, argIdx+1, argIdx+2)
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
		argIdx += 3
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM contacts " + where
	if err := r.db.Get(&total, countQ, args...); err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	var contacts []model.Contact
	query := "SELECT * FROM contacts " + where + " ORDER BY last_activity_at DESC LIMIT $" + fmt.Sprintf("%d", argIdx) + " OFFSET $" + fmt.Sprintf("%d", argIdx+1)
	args = append(args, filter.PageSize, offset)
	if err := r.db.Select(&contacts, query, args...); err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (r *ContactRepository) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	setClause := ""
	args := []interface{}{}
	argIdx := 1
	for k, v := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", k, argIdx)
		args = append(args, v)
		argIdx++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE contacts SET %s WHERE id = $%d", setClause, argIdx)
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *ContactRepository) MergeFingerprints(id string, fingerprints []string) error {
	var existing string
	err := r.db.Get(&existing, "SELECT browser_fingerprints FROM contacts WHERE id = $1", id)
	if err != nil {
		return err
	}

	var existingList []string
	if existing != "" {
		json.Unmarshal([]byte(existing), &existingList)
	}

	seen := make(map[string]bool)
	for _, fp := range existingList {
		seen[fp] = true
	}
	for _, fp := range fingerprints {
		if !seen[fp] {
			existingList = append(existingList, fp)
			seen[fp] = true
		}
	}

	merged, _ := json.Marshal(existingList)
	_, err = r.db.Exec("UPDATE contacts SET browser_fingerprints = $1, updated_at = NOW() WHERE id = $2", string(merged), id)
	return err
}

func (r *ContactRepository) Block(id string) error {
	_, err := r.db.Exec("UPDATE contacts SET blocked = true, updated_at = NOW() WHERE id = $1", id)
	return err
}

func (r *ContactRepository) Unblock(id string) error {
	_, err := r.db.Exec("UPDATE contacts SET blocked = false, updated_at = NOW() WHERE id = $1", id)
	return err
}
