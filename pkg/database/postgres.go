package database

import (
	"database/sql"
	"fmt"
	"github.com/ZnNr/subscription-service/internal/config"
	"time"

	_ "github.com/lib/pq"
)

// ConnectPostgres устанавливает соединение с PostgreSQL
func ConnectPostgres(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// RunMigrations выполняет миграции базы данных
func RunMigrations(db *sql.DB) error {
	// Здесь можно реализовать выполнение SQL миграций
	// или использовать инструмент миграций, например golang-migrate
	return nil
}
