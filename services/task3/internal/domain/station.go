package domain

type Station struct {
	ID          string
	Name        string
	Latitude    float64
	Longitude   float64
	VisitCount  int
	Locomotives map[string]int
}
