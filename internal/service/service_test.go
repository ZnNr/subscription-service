package service

import (
	"context"
	"github.com/ZnNr/subscription-service/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository для тестирования сервиса
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockRepository) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockRepository) UpdateSubscription(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockRepository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	args := m.Called(ctx, userID, serviceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Subscription), args.Error(1)
}

func (m *MockRepository) CalculateSummary(ctx context.Context, startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (*model.SummaryResponse, error) {
	args := m.Called(ctx, startDate, endDate, userID, serviceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SummaryResponse), args.Error(1)
}

func TestCreateSubscription(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       599,
		UserID:      userID,
		StartDate:   startDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Настраиваем мок
	mockRepo.On("CreateSubscription", ctx, sub).Return(nil)
	mockRepo.On("GetSubscription", ctx, sub.ID).Return(sub, nil)

	// Вызываем метод
	result, err := service.CreateSubscription(ctx, sub)

	// Проверяем
	assert.NoError(t, err)
	assert.Equal(t, sub.ServiceName, result.ServiceName)
	assert.Equal(t, sub.Price, result.Price)
	mockRepo.AssertExpectations(t)
}

func TestCreateSubscription_InvalidPrice(t *testing.T) {
	service := NewSubscriptionService(nil)
	ctx := context.Background()

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       0, // Невалидная цена
		UserID:      uuid.New(),
		StartDate:   time.Now().UTC(),
	}

	_, err := service.CreateSubscription(ctx, sub)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPrice, err)
}

func TestGetSubscription(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	subID := uuid.New()
	expectedSub := &model.Subscription{
		ID:          subID,
		ServiceName: "Netflix",
		Price:       599,
		UserID:      uuid.New(),
		StartDate:   time.Now().UTC(),
	}

	// Настраиваем мок
	mockRepo.On("GetSubscription", ctx, subID).Return(expectedSub, nil)

	// Вызываем метод
	result, err := service.GetSubscription(ctx, subID)

	// Проверяем
	assert.NoError(t, err)
	assert.Equal(t, expectedSub, result)
	mockRepo.AssertExpectations(t)
}

func TestUpdateSubscription(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	subID := uuid.New()
	existingSub := &model.Subscription{
		ID:          subID,
		ServiceName: "Netflix",
		Price:       599,
		UserID:      uuid.New(),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	updateReq := &model.UpdateSubscriptionRequest{
		Price: &[]int{699}[0],
	}

	// Настраиваем моки
	mockRepo.On("GetSubscription", ctx, subID).Return(existingSub, nil)
	mockRepo.On("UpdateSubscription", ctx, subID, updateReq).Return(nil)

	// Вызываем метод
	err := service.UpdateSubscription(ctx, subID, updateReq)

	// Проверяем
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteSubscription(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	subID := uuid.New()
	existingSub := &model.Subscription{
		ID: subID,
	}

	// Настраиваем моки
	mockRepo.On("GetSubscription", ctx, subID).Return(existingSub, nil)
	mockRepo.On("DeleteSubscription", ctx, subID).Return(nil)

	// Вызываем метод
	err := service.DeleteSubscription(ctx, subID)

	// Проверяем
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCalculateSummary(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSubscriptionService(mockRepo)
	ctx := context.Background()

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	expectedSummary := &model.SummaryResponse{
		TotalAmount: 1298,
		Count:       2,
	}

	// Настраиваем мок
	mockRepo.On("CalculateSummary", ctx, startDate, endDate, &userID, mock.Anything).Return(expectedSummary, nil)

	// Вызываем метод
	result, err := service.CalculateSummary(ctx, startDate, endDate, &userID, nil)

	// Проверяем
	assert.NoError(t, err)
	assert.Equal(t, expectedSummary, result)
	mockRepo.AssertExpectations(t)
}

func TestCalculateSummary_InvalidPeriod(t *testing.T) {
	service := NewSubscriptionService(nil)
	ctx := context.Background()

	startDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // startDate > endDate

	_, err := service.CalculateSummary(ctx, startDate, endDate, nil, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPeriod, err)
}
