package main

import "github.com/mihnpro/Hackathon_TMX/internal/services"

func main() {
	services.NewAlgorithmService().RunAlgorithm()
	services.NewMostPopularTripService().RunkMostPopularTrip()
}
