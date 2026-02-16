// internal/transport/models/responses/task2.go

package responses

type DirectionInfo struct {
    ID             string `json:"id"`
    Name           string `json:"name"`
    Terminal       string `json:"terminal"`
    TerminalName   string `json:"terminal_name"`
    Frequency      int    `json:"frequency"`
    LocomotiveCount int    `json:"locomotive_count"`
}

type LocomotiveDirection struct {
    ID           string  `json:"id"`
    Name         string  `json:"name"`
    Terminal     string  `json:"terminal"`
    TerminalName string  `json:"terminal_name"`
    Visits       int     `json:"visits"`
    Percentage   float64 `json:"percentage"`
}

type MostPopularDirection struct {
    DirectionID   string  `json:"direction_id"`
    DirectionName string  `json:"direction_name"`
    Visits        int     `json:"visits"`
    Percentage    float64 `json:"percentage"`
}

type LocomotiveStats struct {
    Model        string                 `json:"model"`
    Number       string                 `json:"number"`
    Depo         string                 `json:"depo"`
    DepoName     string                 `json:"depo_name"`
    TotalTrips   int                    `json:"total_trips"`
    Directions   []LocomotiveDirection  `json:"directions"`
    MostPopular  *MostPopularDirection  `json:"most_popular,omitempty"`
}

type DepotResponse struct {
    DepoCode        string            `json:"depo_code"`
    DepoName        string            `json:"depo_name"`
    LocomotiveCount int               `json:"locomotive_count"`
    Directions      []DirectionInfo   `json:"directions"`
    Locomotives     []LocomotiveStats `json:"locomotives"`
}

type OverallStats struct {
    TotalLocomotives      int     `json:"total_locomotives"`
    TotalTrips            int     `json:"total_trips"`
    AvgTripsPerLocomotive float64 `json:"avg_trips_per_locomotive"`
    LocomotivesWithFavorite int    `json:"locomotives_with_favorite"`
    LocomotivesWithFavoritePercent float64 `json:"locomotives_with_favorite_percent"`
    LocomotivesSingleDirection int `json:"locomotives_single_direction"`
    LocomotivesSingleDirectionPercent float64 `json:"locomotives_single_direction_percent"`
}

type Task2Response struct {
    Depots       []DepotResponse `json:"depots"`
    OverallStats OverallStats    `json:"overall_stats"`
}