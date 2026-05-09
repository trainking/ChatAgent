package model

type SystemConfig struct {
	Key   string `db:"key" json:"key"`
	Value string `db:"value" json:"value"`
}

type UserTOTP struct {
	UserID  string `db:"user_id" json:"user_id"`
	Secret  string `db:"secret" json:"-"`
	Enabled bool   `db:"enabled" json:"enabled"`
}
