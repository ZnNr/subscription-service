package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/subscription-service/internal/model"
	"github.com/google/uuid"

	"time"
)

type Repository interface {
	CreateSubscription(ctx context.Context, sub *model.Subscription) error
	GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	CalculateSummary(ctx context.Context, startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (*model.SummaryResponse, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.CreatedAt, sub.UpdatedAt)

	return err
}

func (r *PostgresRepository) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanSubscription(row)
}

func (r *PostgresRepository) UpdateSubscription(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) error {
	// Build dynamic update query
	query := "UPDATE subscriptions SET updated_at = $1"
	args := []interface{}{time.Now().UTC()}
	argIndex := 2

	if req.ServiceName != nil {
		query += fmt.Sprintf(", service_name = $%d", argIndex)
		args = append(args, *req.ServiceName)
		argIndex++
	}

	if req.Price != nil {
		query += fmt.Sprintf(", price = $%d", argIndex)
		args = append(args, *req.Price)
		argIndex++
	}

	if req.StartDate != nil {
		query += fmt.Sprintf(", start_date = $%d", argIndex)
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		query += fmt.Sprintf(", end_date = $%d", argIndex)
		args = append(args, *req.EndDate)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRepository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM subscriptions WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE 1=1
	`
	var args []interface{}
	argIndex := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, *serviceName)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		sub, err := scanSubscriptionFromRows(rows)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *PostgresRepository) CalculateSummary(ctx context.Context, startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (*model.SummaryResponse, error) {
	query := `
		SELECT COALESCE(SUM(price), 0) as total_amount, COUNT(*) as count
		FROM subscriptions
		WHERE start_date >= $1 AND (end_date IS NULL OR end_date <= $2)
	`
	args := []interface{}{startDate, endDate}
	argIndex := 3

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, *serviceName)
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var summary model.SummaryResponse
	err := row.Scan(&summary.TotalAmount, &summary.Count)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func scanSubscription(row *sql.Row) (*model.Subscription, error) {
	var sub model.Subscription
	var endDate sql.NullTime

	err := row.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}

	return &sub, nil
}

func scanSubscriptionFromRows(rows *sql.Rows) (*model.Subscription, error) {
	var sub model.Subscription
	var endDate sql.NullTime

	err := rows.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}

	return &sub, nil
}
