package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ZnNr/subscription-service/internal/model"
	"github.com/ZnNr/subscription-service/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService реализует интерфейс service.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	args := m.Called(ctx, sub)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockService) UpdateSubscription(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	args := m.Called(ctx, userID, serviceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Subscription), args.Error(1)
}

func (m *MockService) CalculateSummary(ctx context.Context, startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (*model.SummaryResponse, error) {
	args := m.Called(ctx, startDate, endDate, userID, serviceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SummaryResponse), args.Error(1)
}

var _ service.Service = (*MockService)(nil)

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", handler.CreateSubscription)
			subscriptions.GET("", handler.ListSubscriptions)
			subscriptions.GET("/:id", handler.GetSubscription)
			subscriptions.PUT("/:id", handler.UpdateSubscription)
			subscriptions.DELETE("/:id", handler.DeleteSubscription)
			subscriptions.POST("/summary", handler.CalculateSummary)
		}
	}

	return router
}

func TestCreateSubscriptionHandler(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")

	requestBody := map[string]interface{}{
		"service_name": "Yandex Plus",
		"price":        400,
		"user_id":      userID.String(),
		"start_date":   "07-2025",
	}

	expectedSub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	mockService.On("CreateSubscription", mock.Anything, mock.AnythingOfType("*model.Subscription")).
		Return(expectedSub, nil)

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/subscriptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedSub.ServiceName, response.ServiceName)
	assert.Equal(t, expectedSub.Price, response.Price)
	mockService.AssertExpectations(t)
}

func TestGetSubscriptionHandler(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	subID := uuid.New()
	expectedSub := &model.Subscription{
		ID:          subID,
		ServiceName: "Netflix",
		Price:       599,
		UserID:      uuid.New(),
		StartDate:   time.Now().UTC(),
	}

	mockService.On("GetSubscription", mock.Anything, subID).
		Return(expectedSub, nil)

	req, _ := http.NewRequest("GET", "/api/v1/subscriptions/"+subID.String(), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedSub.ID, response.ID)
	assert.Equal(t, expectedSub.ServiceName, response.ServiceName)
	mockService.AssertExpectations(t)
}

func TestListSubscriptionsHandler(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expectedSubs := []*model.Subscription{
		{
			ID:          uuid.New(),
			ServiceName: "Netflix",
			Price:       599,
			UserID:      userID,
			StartDate:   time.Now().UTC(),
		},
		{
			ID:          uuid.New(),
			ServiceName: "Spotify",
			Price:       299,
			UserID:      userID,
			StartDate:   time.Now().UTC(),
		},
	}

	mockService.On("ListSubscriptions",
		mock.Anything, // context.Context
		&userID,       // *uuid.UUID
		mock.MatchedBy(func(s *string) bool { return s == nil }), // *string (nil)
	).Return(expectedSubs, nil)

	req, _ := http.NewRequest("GET", "/api/v1/subscriptions?user_id="+userID.String(), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []model.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response, 2)
	mockService.AssertExpectations(t)
}

func TestCalculateSummaryHandler(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	requestBody := map[string]interface{}{
		"start_date": "01-2025",
		"end_date":   "12-2025", // Это парсится как 1 декабря 2025!
		"user_id":    userID.String(),
	}

	expectedSummary := &model.SummaryResponse{
		TotalAmount: 1998,
		Count:       3,
	}

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC) // 1 декабря, а не 31!

	// Исправленный мок
	mockService.On("CalculateSummary",
		mock.Anything, // context.Context
		startDate,     // 1 января 2025
		endDate,       // 1 декабря 2025
		&userID,
		(*string)(nil),
	).Return(expectedSummary, nil)

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/subscriptions/summary", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.SummaryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedSummary.TotalAmount, response.TotalAmount)
	assert.Equal(t, expectedSummary.Count, response.Count)
	mockService.AssertExpectations(t)
}

func TestUpdateSubscriptionHandler(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	subID := uuid.New()

	requestBody := map[string]interface{}{
		"price": 699,
	}

	mockService.On("UpdateSubscription",
		mock.Anything, // context.Context
		subID,
		mock.AnythingOfType("*model.UpdateSubscriptionRequest"),
	).Return(nil)

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/subscriptions/"+subID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteSubscriptionHandler(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	subID := uuid.New()

	mockService.On("DeleteSubscription",
		mock.Anything, // context.Context вместо *gin.Context
		subID,
	).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/subscriptions/"+subID.String(), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
