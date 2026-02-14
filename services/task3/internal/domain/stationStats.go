package domain

type StationStats struct {
	StationID    string
	StationName  string
	Latitude     float64
	Longitude    float64
	VisitCount   int
	Locomotives  []string
	Popularity   float64 // от 0 до 1
}