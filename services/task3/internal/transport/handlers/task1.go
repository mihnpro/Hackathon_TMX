package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
	"github.com/mihnpro/Hackathon_TMX/internal/services"
)

type Task1Handler struct {
	task1Service services.AlgorithmService
}

func NewTask1Handler(task1Service services.AlgorithmService) *Task1Handler {
	return &Task1Handler{
		task1Service: task1Service,
	}
}

// GetBranchAnalysis возвращает анализ веток депо
// @Summary Get branch analysis
// @Description Returns analysis of depot branches and terminal stations
// @Tags task1
// @Accept json
// @Produce json
// @Success 200 {object} responses.Task1Response
// @Failure 500 {object} map[string]string
// @Router /api/v1/task1/branches [get]
func (h *Task1Handler) GetBranchAnalysis(c *gin.Context) {
	data, err := h.task1Service.GetBranchAnalysis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze branches: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetDepotBranches возвращает ветки для конкретного депо
// @Summary Get depot branches
// @Description Returns branches for a specific depot
// @Tags task1
// @Accept json
// @Produce json
// @Param depo path string true "Depot code"
// @Success 200 {object} responses.DepotBranches
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/task1/depots/{depo}/branches [get]
func (h *Task1Handler) GetDepotBranches(c *gin.Context) {
	depoCode := c.Param("depo")
	
	data, err := h.task1Service.GetDepotBranches(depoCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze branches: " + err.Error(),
		})
		return
	}

	if data == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Depot not found",
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetAllDepots возвращает список всех депо
// @Summary Get all depots
// @Description Returns list of all depot codes
// @Tags task1
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/task1/depots [get]
func (h *Task1Handler) GetAllDepots(c *gin.Context) {
	data, err := h.task1Service.GetBranchAnalysis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get depots: " + err.Error(),
		})
		return
	}

	depots := make([]string, 0, len(data.Depots))
	for _, depot := range data.Depots {
		depots = append(depots, depot.DepoCode)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_depots": len(depots),
		"depots":       depots,
	})
}