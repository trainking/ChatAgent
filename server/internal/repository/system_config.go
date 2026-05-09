package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type SystemConfigRepository struct {
	db *sqlx.DB
}

func NewSystemConfigRepository(db *sqlx.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

func (r *SystemConfigRepository) GetAll() (map[string]string, error) {
	var items []model.SystemConfig
	if err := r.db.Select(&items, "SELECT key, value FROM system_config"); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, it := range items {
		result[it.Key] = it.Value
	}
	return result, nil
}

func (r *SystemConfigRepository) Set(key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO system_config (key, value, updated_at) VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`,
		key, value,
	)
	return err
}
