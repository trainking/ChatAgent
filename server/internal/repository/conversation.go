package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chatagent/server/internal/model"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

type ConversationRepository struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewConversationRepository(db *sqlx.DB, rdb *redis.Client) *ConversationRepository {
	return &ConversationRepository{db: db, rdb: rdb}
}

func (r *ConversationRepository) Create(conv *model.Conversation) error {
	now := time.Now()
	conv.CreatedAt = now
	conv.UpdatedAt = now
	conv.LastMessageAt = now

	if conv.Status == "" {
		conv.Status = "open"
	}
	if conv.Priority == "" {
		conv.Priority = "medium"
	}

	ctx := context.Background()
	displayID, err := r.rdb.Incr(ctx, fmt.Sprintf("display_id:%s", conv.InboxID)).Result()
	if err != nil {
		return err
	}
	conv.DisplayID = int(displayID)

	return r.db.QueryRow(
		`INSERT INTO conversations (display_id, inbox_id, contact_id, assignee_id, status, priority, subject, unread_count, last_message_at, waiting_since, first_reply_at, resolved_at, snoozed_until, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING id`,
		conv.DisplayID, conv.InboxID, conv.ContactID, conv.AssigneeID, conv.Status, conv.Priority, conv.Subject, conv.UnreadCount,
		conv.LastMessageAt, conv.WaitingSince, conv.FirstReplyAt, conv.ResolvedAt, conv.SnoozedUntil, conv.CreatedAt, conv.UpdatedAt,
	).Scan(&conv.ID)
}

func (r *ConversationRepository) FindByID(id string) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.Get(&conv, "SELECT * FROM conversations WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

type ConversationFilter struct {
	InboxID    string
	Status     string
	Priority   string
	AssigneeID string
	Search     string
	Page       int
	PageSize   int
}

func (r *ConversationRepository) List(filter ConversationFilter) ([]model.Conversation, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.InboxID != "" {
		where += fmt.Sprintf(" AND c.inbox_id = $%d", argIdx)
		args = append(args, filter.InboxID)
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND c.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Priority != "" {
		where += fmt.Sprintf(" AND c.priority = $%d", argIdx)
		args = append(args, filter.Priority)
		argIdx++
	}
	if filter.AssigneeID == "unassigned" {
		where += " AND c.assignee_id IS NULL"
	} else if filter.AssigneeID == "me" {
		where += " AND c.assignee_id IS NOT NULL"
	} else if filter.AssigneeID != "" {
		where += fmt.Sprintf(" AND c.assignee_id = $%d", argIdx)
		args = append(args, filter.AssigneeID)
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (co.name ILIKE $%d OR c.subject ILIKE $%d)", argIdx, argIdx+1)
		s := "%" + filter.Search + "%"
		args = append(args, s, s)
		argIdx += 2
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM conversations c LEFT JOIN contacts co ON c.contact_id = co.id " + where
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

	var convs []model.Conversation
	query := "SELECT c.* FROM conversations c LEFT JOIN contacts co ON c.contact_id = co.id " + where +
		fmt.Sprintf(" ORDER BY c.last_message_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)
	if err := r.db.Select(&convs, query, args...); err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

func (r *ConversationRepository) UpdateStatus(id string, status string) error {
	var updates map[string]interface{}
	switch status {
	case "resolved":
		updates = map[string]interface{}{
			"status":      status,
			"resolved_at": time.Now(),
		}
	case "snoozed":
		updates = map[string]interface{}{
			"status": status,
		}
	default:
		updates = map[string]interface{}{
			"status": status,
		}
	}
	return r.updateFields(id, updates)
}

func (r *ConversationRepository) UpdateAssignee(id string, assigneeID *string) error {
	return r.updateFields(id, map[string]interface{}{
		"assignee_id": assigneeID,
	})
}

func (r *ConversationRepository) UpdatePriority(id string, priority string) error {
	return r.updateFields(id, map[string]interface{}{
		"priority": priority,
	})
}

func (r *ConversationRepository) UpdateSnoozedUntil(id string, snoozedUntil time.Time) error {
	return r.updateFields(id, map[string]interface{}{
		"status":        "snoozed",
		"snoozed_until": snoozedUntil,
	})
}

func (r *ConversationRepository) OnNewMessage(convID string, senderType string, fromContact bool) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_message_at": now,
		"updated_at":      now,
	}

	if fromContact {
		updates["waiting_since"] = now
	} else {
		_, err := r.db.Exec(
			"UPDATE conversations SET first_reply_at = $1 WHERE id = $2 AND first_reply_at IS NULL",
			now, convID,
		)
		if err != nil {
			return err
		}
	}

	return r.updateFields(convID, updates)
}

func (r *ConversationRepository) IncrementUnread(id string) error {
	_, err := r.db.Exec("UPDATE conversations SET unread_count = unread_count + 1 WHERE id = $1", id)
	return err
}

func (r *ConversationRepository) ResetUnread(id string) error {
	_, err := r.db.Exec("UPDATE conversations SET unread_count = 0 WHERE id = $1", id)
	return err
}

func (r *ConversationRepository) GetUnreadCount(userID string, inboxIDs []string) (int64, error) {
	if len(inboxIDs) == 0 {
		return 0, nil
	}

	query, args, _ := sqlx.In(
		"SELECT COALESCE(SUM(unread_count), 0) FROM conversations WHERE inbox_id IN (?) AND (assignee_id = ? OR assignee_id IS NULL)",
		inboxIDs, userID,
	)
	query = r.db.Rebind(query)

	var total int64
	err := r.db.Get(&total, query, args...)
	return total, err
}

func (r *ConversationRepository) FindByContact(contactID string) ([]model.Conversation, error) {
	var convs []model.Conversation
	err := r.db.Select(&convs, "SELECT * FROM conversations WHERE contact_id = $1 ORDER BY last_message_at DESC", contactID)
	return convs, err
}

func (r *ConversationRepository) updateFields(id string, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE conversations SET %s WHERE id = $%d", setClause, argIdx)
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *ConversationRepository) ReopenFromResolved(id string) error {
	return r.updateFields(id, map[string]interface{}{
		"status":       "open",
		"resolved_at":  nil,
		"waiting_since": time.Now(),
	})
}
