package config

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	DB *sqlx.DB
}

func NewConfig() (*Config, error) {
	db, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")),
	)
	if err != nil {
		return nil, err
	}
	return &Config{DB: db}, nil
}