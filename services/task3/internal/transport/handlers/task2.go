package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mihnpro/Hackathon_TMX/internal/services"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/models/requests"

)

type Task2Handler struct {
	task2Service services.MostPopularTripService
}

func NewTask2Handler(task2Service services.MostPopularTripService) *Task2Handler {
	return &Task2Handler{
		task2Service: task2Service,
	}
}

// GetPopularDirections возвращает результаты анализа популярных направлений
func (h *Task2Handler) GetPopularDirections(c *gin.Context) {
	data, err := h.task2Service.GetPopularDirections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze popular directions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetLocomotivePopularDirection возвращает самое популярное направление для конкретного локомотива
func (h *Task2Handler) GetLocomotivePopularDirection(c *gin.Context) {
	var req requests.LocomotiveDirectionRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.task2Service.GetLocomotivePopularDirection(req.Series, req.Number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze popular directions: " + err.Error(),
		})
		return
	}

	if data == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Locomotive not found",
		})
		return
	}

	c.JSON(http.StatusOK, data)
}
