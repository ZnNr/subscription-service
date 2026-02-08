package database

import (
	"database/sql"
	"fmt"
	"github.com/ZnNr/subscription-service/internal/config"
	"log"
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
	migrations := []string{
		// Миграция 1: Создание таблицы subscriptions
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			service_name VARCHAR(255) NOT NULL,
			price INTEGER NOT NULL CHECK (price > 0),
			user_id UUID NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Миграция 2: Создание индексов
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_dates ON subscriptions(start_date, end_date)`,

		// Миграция 3: Таблица для отслеживания миграций (опционально)
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Выполняем каждую миграцию
	for i, migration := range migrations {
		log.Printf("Applying migration %d", i+1)

		if _, err := tx.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %v", i+1, err)
		}
	}

	// Записываем версию миграции
	_, err = tx.Exec(`
		INSERT INTO schema_migrations (version) 
		VALUES (1) 
		ON CONFLICT (version) DO NOTHING
	`)
	if err != nil {
		log.Printf("Warning: could not record migration version: %v", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}
