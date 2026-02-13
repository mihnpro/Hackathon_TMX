package domain

type MapConfig struct {
	DepoID          string
	CenterLat       float64
	CenterLon       float64
	Zoom            int
	ShowHeatmap     bool
	ShowRoutes      bool
	ShowStations    bool
	MaxLocomotives  int
}