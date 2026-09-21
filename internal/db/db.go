package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func ConfigFromEnv() (Config, error) {
	config := Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}
	if config.Host == "" || config.Port == "" || config.Name == "" || config.User == "" {
		return Config{}, fmt.Errorf("DB_HOST, DB_PORT, DB_NAME, and DB_USER must be set")
	}
	return config, nil
}

func Open(ctx context.Context, config Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.User, config.Password, config.Host, config.Port, config.Name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingContext); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
