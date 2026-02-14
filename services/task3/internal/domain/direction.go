package domain

type Direction struct {
	ID       string
	Name     string
	Prefix   string   // префикс станций (первые 2 цифры)
	Stations []string // все станции в направлении
	Depo     string
}