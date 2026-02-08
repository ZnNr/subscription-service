package repository

import (
	"context"
	"database/sql"
	"github.com/ZnNr/subscription-service/internal/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PostgresRepositoryTestSuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo Repository
	ctx  context.Context
}

func (s *PostgresRepositoryTestSuite) SetupTest() {
	var err error
	s.db, s.mock, err = sqlmock.New()
	assert.NoError(s.T(), err)

	s.repo = NewPostgresRepository(s.db)
	s.ctx = context.Background()
}

func (s *PostgresRepositoryTestSuite) TearDownTest() {
	s.db.Close()
}

func (s *PostgresRepositoryTestSuite) TestCreateSubscription() {
	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       599,
		UserID:      uuid.New(),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     nil,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Ожидаем SQL запрос
	s.mock.ExpectExec(`INSERT INTO subscriptions`).
		WithArgs(
			sub.ID, sub.ServiceName, sub.Price, sub.UserID,
			sub.StartDate, sub.EndDate, sub.CreatedAt, sub.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызываем метод
	err := s.repo.CreateSubscription(s.ctx, sub)

	// Проверяем
	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *PostgresRepositoryTestSuite) TestGetSubscription() {
	subID := uuid.New()
	expectedSub := &model.Subscription{
		ID:          subID,
		ServiceName: "Netflix",
		Price:       599,
		UserID:      uuid.New(),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     nil,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Ожидаем SQL запрос
	rows := sqlmock.NewRows([]string{
		"id", "service_name", "price", "user_id",
		"start_date", "end_date", "created_at", "updated_at",
	}).AddRow(
		expectedSub.ID, expectedSub.ServiceName, expectedSub.Price, expectedSub.UserID,
		expectedSub.StartDate, expectedSub.EndDate, expectedSub.CreatedAt, expectedSub.UpdatedAt,
	)

	s.mock.ExpectQuery(`SELECT .* FROM subscriptions WHERE id = \$1`).
		WithArgs(subID).
		WillReturnRows(rows)

	// Вызываем метод
	result, err := s.repo.GetSubscription(s.ctx, subID)

	// Проверяем
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedSub.ID, result.ID)
	assert.Equal(s.T(), expectedSub.ServiceName, result.ServiceName)
	assert.Equal(s.T(), expectedSub.Price, result.Price)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *PostgresRepositoryTestSuite) TestGetSubscription_NotFound() {
	subID := uuid.New()

	s.mock.ExpectQuery(`SELECT .* FROM subscriptions WHERE id = \$1`).
		WithArgs(subID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "service_name", "price", "user_id",
			"start_date", "end_date", "created_at", "updated_at",
		}))

	result, err := s.repo.GetSubscription(s.ctx, subID)

	// Просто проверяем что есть ошибка, без проверки текста
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *PostgresRepositoryTestSuite) TestUpdateSubscription() {
	subID := uuid.New()
	updateReq := &model.UpdateSubscriptionRequest{
		Price:       &[]int{699}[0],
		ServiceName: &[]string{"Netflix Premium"}[0],
	}

	// Ожидаем SQL запрос
	s.mock.ExpectExec(`UPDATE subscriptions SET updated_at = \$1, service_name = \$2, price = \$3 WHERE id = \$4`).
		WithArgs(sqlmock.AnyArg(), "Netflix Premium", 699, subID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Вызываем метод
	err := s.repo.UpdateSubscription(s.ctx, subID, updateReq)

	// Проверяем
	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *PostgresRepositoryTestSuite) TestDeleteSubscription() {
	subID := uuid.New()

	// Ожидаем SQL запрос
	s.mock.ExpectExec(`DELETE FROM subscriptions WHERE id = \$1`).
		WithArgs(subID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Вызываем метод
	err := s.repo.DeleteSubscription(s.ctx, subID)

	// Проверяем
	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *PostgresRepositoryTestSuite) TestListSubscriptions() {
	userID := uuid.New()
	serviceName := "Netflix"

	expectedSubs := []*model.Subscription{
		{
			ID:          uuid.New(),
			ServiceName: "Netflix",
			Price:       599,
			UserID:      userID,
			StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     nil,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	// Ожидаем SQL запрос
	rows := sqlmock.NewRows([]string{
		"id", "service_name", "price", "user_id",
		"start_date", "end_date", "created_at", "updated_at",
	}).AddRow(
		expectedSubs[0].ID, expectedSubs[0].ServiceName, expectedSubs[0].Price, expectedSubs[0].UserID,
		expectedSubs[0].StartDate, expectedSubs[0].EndDate, expectedSubs[0].CreatedAt, expectedSubs[0].UpdatedAt,
	)

	s.mock.ExpectQuery(`SELECT .* FROM subscriptions WHERE 1=1 AND user_id = \$1 AND service_name = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, serviceName).
		WillReturnRows(rows)

	// Вызываем метод
	result, err := s.repo.ListSubscriptions(s.ctx, &userID, &serviceName)

	// Проверяем
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), expectedSubs[0].ServiceName, result[0].ServiceName)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *PostgresRepositoryTestSuite) TestCalculateSummary() {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	userID := uuid.New()

	// Ожидаем SQL запрос
	rows := sqlmock.NewRows([]string{"total_amount", "count"}).
		AddRow(1298, 2)

	s.mock.ExpectQuery(`SELECT COALESCE\(SUM\(price\), 0\) as total_amount, COUNT\(\*\) as count FROM subscriptions WHERE start_date >= \$1 AND \(end_date IS NULL OR end_date <= \$2\) AND user_id = \$3`).
		WithArgs(startDate, endDate, userID).
		WillReturnRows(rows)

	// Вызываем метод
	result, err := s.repo.CalculateSummary(s.ctx, startDate, endDate, &userID, nil)

	// Проверяем
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1298, result.TotalAmount)
	assert.Equal(s.T(), 2, result.Count)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestPostgresRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepositoryTestSuite))
}
