package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type UserTOTPRepository struct {
	db *sqlx.DB
}

func NewUserTOTPRepository(db *sqlx.DB) *UserTOTPRepository {
	return &UserTOTPRepository{db: db}
}

func (r *UserTOTPRepository) GetByUserID(userID string) (*model.UserTOTP, error) {
	t := &model.UserTOTP{}
	err := r.db.Get(t, "SELECT user_id, secret, enabled FROM user_totp WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *UserTOTPRepository) Upsert(userID, secret string) error {
	_, err := r.db.Exec(
		`INSERT INTO user_totp (user_id, secret, enabled, updated_at) VALUES ($1, $2, false, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET secret = $2, enabled = false, updated_at = NOW()`,
		userID, secret,
	)
	return err
}

func (r *UserTOTPRepository) Enable(userID string) error {
	_, err := r.db.Exec("UPDATE user_totp SET enabled = true, updated_at = NOW() WHERE user_id = $1", userID)
	return err
}
