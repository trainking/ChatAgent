package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type ActivityRepository struct {
	db *sqlx.DB
}

func NewActivityRepository(db *sqlx.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(log *model.ActivityLog) error {
	query := `INSERT INTO activity_logs (user_id, event, module, content)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, log.UserID, log.Event, log.Module, log.Content).
		Scan(&log.ID, &log.CreatedAt)
}

func (r *ActivityRepository) FindByUserID(userID string, limit, offset int) ([]model.ActivityLog, int64, error) {
	var total int64
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM activity_logs WHERE user_id = $1", userID); err != nil {
		return nil, 0, err
	}

	var logs []model.ActivityLog
	query := `SELECT id, user_id, event, module, content, created_at
		FROM activity_logs WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.Select(&logs, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
