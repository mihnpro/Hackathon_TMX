package responses

type Task2Response struct {
	Depots              []DepotResponse      `json:"depots"`
	OverallStats        OverallStats         `json:"overall_stats"`
	DirectionPopularity []PopularityItem     `json:"direction_popularity"`
}

type DepotResponse struct {
	DepoCode            string          `json:"depo_code"`
	LocomotiveCount     int             `json:"locomotive_count"`
	DirectionsCount     int             `json:"directions_count"`
	AvailableDirections []DirectionInfo `json:"available_directions"`
	Locomotives         []LocomotiveStats `json:"locomotives"`
}

type DirectionInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

type LocomotiveStats struct {
	Model             string            `json:"model"`
	Number            string            `json:"number"`
	Depo              string            `json:"depo"`
	TotalTrips        int               `json:"total_trips"`
	VisitedDirections []string          `json:"visited_directions"`
	DirectionVisits   []DirectionVisit  `json:"direction_visits"`
	MostPopular       *MostPopularDirection `json:"most_popular,omitempty"`
}

type DirectionVisit struct {
	DirectionID   string  `json:"direction_id"`
	DirectionName string  `json:"direction_name"`
	Visits        int     `json:"visits"`
	Percentage    float64 `json:"percentage"`
}

type MostPopularDirection struct {
	DirectionID   string  `json:"direction_id"`
	DirectionName string  `json:"direction_name"`
	Visits        int     `json:"visits"`
	Percentage    float64 `json:"percentage"`
}

type OverallStats struct {
	TotalLocomotives                int     `json:"total_locomotives"`
	LocomotivesWithFavorite         int     `json:"locomotives_with_favorite"`
	LocomotivesWithFavoritePercent  float64 `json:"locomotives_with_favorite_percent"`
	LocomotivesSingleDirection      int     `json:"locomotives_single_direction"`
	LocomotivesSingleDirectionPercent float64 `json:"locomotives_single_direction_percent"`
}

type PopularityItem struct {
	DirectionID   string  `json:"direction_id"`
	DirectionName string  `json:"direction_name"`
	Count         int     `json:"count"`
	Percentage    float64 `json:"percentage"`
}

// Для запроса конкретного локомотива
type LocomotiveDirectionRequest struct {
	Series string `uri:"series" binding:"required"`
	Number string `uri:"number" binding:"required"`
}