package domain

type Station struct {
	ID          string
	Name        string
	Latitude    float64
	Longitude   float64
	VisitCount  int
	Locomotives map[string]int
}

type StationInfo struct {
    Code      string
    Name      string
    Latitude  float64
    Longitude float64
}

// StationMap для быстрого доступа к информации о станциях
type StationMap map[string]StationInfo