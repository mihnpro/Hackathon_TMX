package domain

type LocomotiveRoute struct {
	LocomotiveKey string
	Model         string
	Number        string
	Points        []RoutePoint
	Color         string
	Trips         int
}