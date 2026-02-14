package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
	"github.com/mihnpro/Hackathon_TMX/internal/services"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/models/responses"
)

type Task3Handler struct {
	task3Service services.VisualizationService
}

func NewTask3Handler(task3Service services.VisualizationService) *Task3Handler {
	return &Task3Handler{
		task3Service: task3Service,
	}
}

// GenerateMaps генерирует карты для депо
// @Summary Generate maps for depot
// @Description Generates overview map, heatmap and locomotive maps for a depot
// @Tags task3
// @Accept json
// @Produce json
// @Param request body responses.GenerateMapsRequest true "Generation parameters"
// @Success 200 {object} responses.GenerateMapsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/task3/generate [post]
func (h *Task3Handler) GenerateMaps(c *gin.Context) {
	var req responses.GenerateMapsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем значение по умолчанию
	if req.MaxLocomotives == 0 {
		req.MaxLocomotives = 10
	}

	data, err := h.task3Service.GenerateMapsAPI(req.DepoID, req.MaxLocomotives)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate maps: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetAvailableDepots возвращает список всех депо
// @Summary Get available depots
// @Description Returns list of all depot IDs that have locomotives
// @Tags task3
// @Accept json
// @Produce json
// @Success 200 {object} responses.DepotsListResponse
// @Router /api/v1/task3/depots [get]
func (h *Task3Handler) GetAvailableDepots(c *gin.Context) {
	depots, err := h.task3Service.GetAvailableDepots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get depots: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.DepotsListResponse{
		Total:  len(depots),
		Depots: depots,
	})
}

// GetDepotInfo возвращает информацию о депо
// @Summary Get depot information
// @Description Returns information about a specific depot
// @Tags task3
// @Accept json
// @Produce json
// @Param depo path string true "Depot ID"
// @Success 200 {object} responses.DepotInfo
// @Failure 404 {object} map[string]string
// @Router /api/v1/task3/depots/{depo} [get]
func (h *Task3Handler) GetDepotInfo(c *gin.Context) {
	depoID := c.Param("depo")
	
	info, err := h.task3Service.GetDepotInfo(depoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}