package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type PermissionRepository struct {
	db *sqlx.DB
}

func NewPermissionRepository(db *sqlx.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) ListAll() ([]model.Permission, error) {
	var perms []model.Permission
	query := `SELECT code, name, description FROM permissions ORDER BY code`
	err := r.db.Select(&perms, query)
	return perms, err
}

func (r *PermissionRepository) GetRolePermissions(role string) ([]string, error) {
	var codes []string
	query := `SELECT permission_code FROM role_permissions WHERE role = $1 ORDER BY permission_code`
	err := r.db.Select(&codes, query, role)
	return codes, err
}

func (r *PermissionRepository) SetRolePermissions(role string, codes []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM role_permissions WHERE role = $1", role); err != nil {
		return err
	}

	for _, code := range codes {
		if _, err := tx.Exec(
			"INSERT INTO role_permissions (role, permission_code) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			role, code,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
