package domain


type ImprovedBranch struct {
	Depo          string
	CoreStations  []string      // уникальные станции в порядке от депо
	BranchID      string        // уникальный идентификатор ветки
	AllPaths      [][]string    // все варианты проезда по ветке
	Terminals     map[string]int // конечные станции и частота посещения
	Length        int           // длина ветки в станциях
}
