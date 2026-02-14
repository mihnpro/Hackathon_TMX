package domain

type LocomotiveDirectionStats struct {
	LocomotiveKey        string
	Model                string
	Number               string
	Depo                 string
	TotalTrips           int
	DirectionVisits      map[string]int // направление -> количество поездок
	MostPopularDirection string
	MaxVisits            int
	PopularDirectionName string
	VisitedDirections    []string // все посещенные направления
}