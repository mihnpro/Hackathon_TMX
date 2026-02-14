package responses

type Task1Response struct {
	Depots          []DepotBranches    `json:"depots"`
	OverallStats    OverallStatsTask1  `json:"overall_stats"`
	LongestBranches []LongestBranch    `json:"longest_branches"`
}

type DepotBranches struct {
	DepoCode    string        `json:"depo_code"`
	BranchCount int           `json:"branch_count"`
	Branches    []BranchInfo  `json:"branches"`
}

type BranchInfo struct {
	BranchID      string            `json:"branch_id"`
	CoreStations  []string          `json:"core_stations"`
	StationCount  int               `json:"station_count"`
	Terminals     []TerminalInfo    `json:"terminals"`
	ExamplePath   []string          `json:"example_path"`
}

type TerminalInfo struct {
	Station   string  `json:"station"`
	Visits    int     `json:"visits"`
	Frequency float64 `json:"frequency"` // процент поездок, заканчивающихся на этой станции
}

type OverallStatsTask1 struct {
	TotalDepots     int     `json:"total_depots"`
	TotalBranches   int     `json:"total_branches"`
	TotalTerminals  int     `json:"total_terminals"`
	AvgBranchesPerDepo float64 `json:"avg_branches_per_depo"`
}

type LongestBranch struct {
	DepoCode    string   `json:"depo_code"`
	BranchID    string   `json:"branch_id"`
	Length      int      `json:"length"`
	Route       []string `json:"route"`
	RouteString string   `json:"route_string"`
}