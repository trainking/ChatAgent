package repository

import (
	"github.com/chatagent/server/internal/model"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `INSERT INTO users (email, name, password_hash, role, avatar_url, must_change_password)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, user.Email, user.Name, user.PasswordHash, user.Role, user.AvatarURL, user.MustChangePassword).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, name, password_hash, role, avatar_url, status, must_change_password,
		created_at, updated_at FROM users WHERE email = $1`
	err := r.db.Get(user, query, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, name, password_hash, role, avatar_url, status, must_change_password,
		created_at, updated_at FROM users WHERE id = $1`
	err := r.db.Get(user, query, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM users")
	return count, err
}

func (r *UserRepository) UpdatePassword(id, hash string) error {
	_, err := r.db.Exec(
		"UPDATE users SET password_hash = $1, must_change_password = false, updated_at = NOW() WHERE id = $2",
		hash, id,
	)
	return err
}

func (r *UserRepository) AdminResetPassword(id, hash string) error {
	_, err := r.db.Exec(
		"UPDATE users SET password_hash = $1, must_change_password = true, updated_at = NOW() WHERE id = $2",
		hash, id,
	)
	return err
}

func (r *UserRepository) List() ([]model.User, error) {
	var users []model.User
	query := `SELECT id, email, name, password_hash, role, avatar_url, status, must_change_password,
		created_at, updated_at FROM users ORDER BY created_at DESC`
	err := r.db.Select(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (r *UserRepository) Update(user *model.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET name=$1, role=$2, status=$3, must_change_password=$4, updated_at=NOW() WHERE id=$5`,
		user.Name, user.Role, user.Status, user.MustChangePassword, user.ID,
	)
	return err
}
