package domain


type LocomotiveDirectionStats struct {
    LocomotiveKey       string
    Model               string
    Number              string
    Depo                string
    DepoName            string
    TotalTrips          int
    DirectionVisits     map[string]int    // ID направления -> количество поездок
    Directions          []DirectionInfo   // информация о посещенных направлениях
    MostPopularDirection string
    MostPopularName     string
    MaxVisits           int
}

type DirectionInfo struct {
    ID          string
    Name        string
    Terminal    string
    TerminalName string
    Visits      int
    Percentage  float64
}