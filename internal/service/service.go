package service

import (
	"context"
	"github.com/ZnNr/subscription-service/internal/model"
	"github.com/ZnNr/subscription-service/internal/repository"
	"github.com/google/uuid"

	"time"
)

type Service interface {
	CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	CalculateSummary(ctx context.Context, startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (*model.SummaryResponse, error)
}

type SubscriptionService struct {
	repo repository.Repository
}

func NewSubscriptionService(repo repository.Repository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	if err := validateSubscription(sub); err != nil {
		return nil, err
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	createdSub, err := s.repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		return nil, err
	}

	return createdSub, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetSubscription(ctx, id)
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) error {
	existing, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}

	if err := validateUpdateRequest(req, existing); err != nil {
		return err
	}

	return s.repo.UpdateSubscription(ctx, id, req)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.DeleteSubscription(ctx, id)
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	return s.repo.ListSubscriptions(ctx, userID, serviceName)
}

func (s *SubscriptionService) CalculateSummary(ctx context.Context, startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (*model.SummaryResponse, error) {
	if startDate.After(endDate) {
		return nil, ErrInvalidPeriod
	}

	return s.repo.CalculateSummary(ctx, startDate, endDate, userID, serviceName)
}

func validateSubscription(sub *model.Subscription) error {
	if sub.ServiceName == "" {
		return ErrServiceNameRequired
	}

	if sub.Price <= 0 {
		return ErrInvalidPrice
	}

	if sub.UserID == uuid.Nil {
		return ErrUserIDRequired
	}

	if sub.StartDate.IsZero() {
		return ErrStartDateRequired
	}

	// Проверка, что end_date не раньше start_date
	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return ErrInvalidEndDate
	}

	return nil
}

func validateUpdateRequest(req *model.UpdateSubscriptionRequest, existing *model.Subscription) error {
	if req.Price != nil && *req.Price <= 0 {
		return ErrInvalidPrice
	}

	if req.StartDate != nil {
		parsedDate, err := time.Parse("01-YYYY", *req.StartDate)
		if err != nil {
			return ErrInvalidDateFormat
		}

		if existing.EndDate != nil && parsedDate.After(*existing.EndDate) {
			return ErrInvalidStartDate
		}
	}

	if req.EndDate != nil {
		parsedDate, err := time.Parse("01-YYYY", *req.EndDate)
		if err != nil {
			return ErrInvalidDateFormat
		}

		if parsedDate.Before(existing.StartDate) {
			return ErrInvalidEndDate
		}
	}

	return nil
}

// Ошибки
var (
	ErrServiceNameRequired = NewServiceError("service name is required")
	ErrInvalidPrice        = NewServiceError("price must be greater than 0")
	ErrUserIDRequired      = NewServiceError("user ID is required")
	ErrStartDateRequired   = NewServiceError("start date is required")
	ErrInvalidEndDate      = NewServiceError("end date cannot be before start date")
	ErrInvalidStartDate    = NewServiceError("start date cannot be after end date")
	ErrInvalidPeriod       = NewServiceError("start date cannot be after end date")
	ErrInvalidDateFormat   = NewServiceError("invalid date format, expected MM-YYYY")
	ErrNotFound            = NewServiceError("subscription not found")
)

type ServiceError struct {
	Message string
}

func (e ServiceError) Error() string {
	return e.Message
}

func NewServiceError(message string) ServiceError {
	return ServiceError{Message: message}
}
