package handler

import (
	"github.com/ZnNr/subscription-service/internal/model"
	"github.com/ZnNr/subscription-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// CreateSubscription создает новую подписку
// @Summary Создать подписку
// @Description Создает новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body model.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	var req model.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		ed, err := parseMonthYear(*req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		endDate = &ed
	}

	sub, err := h.service.CreateSubscription(c.Request.Context(), &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetSubscription получает подписку по ID
// @Summary Получить подписку
// @Description Возвращает информацию о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 404 {object} map[string]interface{} "Подписка не найдена"
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	sub, err := h.service.GetSubscription(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription обновляет подписку
// @Summary Обновить подписку
// @Description Обновляет информацию о существующей подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param input body model.UpdateSubscriptionRequest true "Обновленные данные"
// @Success 200 {object} map[string]interface{} "Подписка обновлена"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 404 {object} map[string]interface{} "Подписка не найдена"
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	var req model.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSubscription(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated"})
}

// DeleteSubscription удаляет подписку
// @Summary Удалить подписку
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 204 "Подписка удалена"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 404 {object} map[string]interface{} "Подписка не найдена"
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	if err := h.service.DeleteSubscription(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListSubscriptions возвращает список подписок
// @Summary Список подписок
// @Description Возвращает список подписок с возможностью фильтрации
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "ID пользователя для фильтрации"
// @Param service_name query string false "Название сервиса для фильтрации"
// @Success 200 {array} model.Subscription
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	var userID *uuid.UUID
	if uid := c.Query("user_id"); uid != "" {
		if parsed, err := uuid.Parse(uid); err == nil {
			userID = &parsed
		}
	}

	var serviceName *string
	if sn := c.Query("service_name"); sn != "" {
		serviceName = &sn
	}

	subscriptions, err := h.service.ListSubscriptions(c.Request.Context(), userID, serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// CalculateSummary считает суммарную стоимость подписок
// @Summary Сумма подписок
// @Description Рассчитывает суммарную стоимость подписок за период с фильтрацией
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body model.SummaryRequest true "Параметры расчета"
// @Success 200 {object} model.SummaryResponse
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Router /subscriptions/summary [post]
func (h *Handler) CalculateSummary(c *gin.Context) {
	var req model.SummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	endDate, err := parseMonthYear(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
		return
	}

	summary, err := h.service.CalculateSummary(c.Request.Context(), startDate, endDate, req.UserID, req.ServiceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ValidateMonthYear проверяет формат "MM-YYYY"
func ValidateMonthYear(dateStr string) bool {
	// Регулярное выражение для формата MM-YYYY
	re := regexp.MustCompile(`^(0[1-9]|1[0-2])-(\d{4})$`)
	return re.MatchString(dateStr)
}

// parseMonthYear
func parseMonthYear(dateStr string) (time.Time, error) {
	if !ValidateMonthYear(dateStr) {
		return time.Time{}, &time.ParseError{
			Value:   dateStr,
			Message: "invalid format, expected MM-YYYY",
		}
	}

	month, _ := strconv.Atoi(dateStr[0:2])
	year, _ := strconv.Atoi(dateStr[3:7])

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}
