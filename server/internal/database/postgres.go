package database

import (
	"fmt"
	"time"

	"github.com/chatagent/server/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgres(cfg config.DBConfig) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	for i := 0; i < 15; i++ {
		db, err = sqlx.Connect("postgres", cfg.DSN())
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
