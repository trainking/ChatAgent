package model

import "time"

type User struct {
	ID                 string     `db:"id" json:"id"`
	Email              string     `db:"email" json:"email"`
	Name               string     `db:"name" json:"name"`
	PasswordHash       string     `db:"password_hash" json:"-"`
	Role               string     `db:"role" json:"role"`
	AvatarURL          string     `db:"avatar_url" json:"avatar_url"`
	Status             string     `db:"status" json:"status"`
	MustChangePassword bool       `db:"must_change_password" json:"must_change_password"`
	LastLoginAt        *time.Time `db:"last_login_at" json:"last_login_at"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}
