package responses

type GenerateMapsRequest struct {
	DepoID         string `json:"depo_id" binding:"required"`
	MaxLocomotives int    `json:"max_locomotives" binding:"min=1,max=20"`
}

type GenerateMapsResponse struct {
	DepotID     string    `json:"depot_id"`
	GeneratedAt string    `json:"generated_at"`
	Maps        MapsList  `json:"maps"`
}

type MapsList struct {
	Overview    string          `json:"overview"`     // ссылка на общую карту
	Heatmap     string          `json:"heatmap"`      // ссылка на тепловую карту
	Locomotives []LocomotiveMap `json:"locomotives"`  // карты отдельных локомотивов
}

type LocomotiveMap struct {
	Key       string `json:"key"`        // "ВЛ80С_12453"
	Model     string `json:"model"`
	Number    string `json:"number"`
	URL       string `json:"url"`        // ссылка на HTML карту
	TripCount int    `json:"trip_count"`
}

type DepotInfo struct {
	DepoID          string `json:"depo_id"`
	Region          string `json:"region"`
	LocomotiveCount int    `json:"locomotive_count"`
}

type DepotsListResponse struct {
	Total  int      `json:"total"`
	Depots []string `json:"depots"`
}