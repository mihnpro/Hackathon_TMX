package handlers

import "github.com/mihnpro/Hackathon_TMX/internal/services"

type Handler struct {

	task1 services.AlgorithmService
	task2 services.MostPopularTripService
	task3 services.VisualizationService
}

func NewHandler(task1 services.AlgorithmService, 
	task2 services.MostPopularTripService, 
	task3 services.VisualizationService) *Handler {

	return &Handler{
		task1: task1,
		task2: task2,
		task3: task3,
	}
}

