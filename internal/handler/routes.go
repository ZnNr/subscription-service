package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", h.CreateSubscription)
			subscriptions.GET("", h.ListSubscriptions)
			subscriptions.GET("/:id", h.GetSubscription)
			subscriptions.PUT("/:id", h.UpdateSubscription)
			subscriptions.DELETE("/:id", h.DeleteSubscription)
			subscriptions.POST("/summary", h.CalculateSummary)
		}
	}

}
