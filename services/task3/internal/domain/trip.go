package domain

import "time"

type Trip struct {
	StartTime  time.Time
	EndTime    time.Time
	Stations   []string
	IsComplete bool
}
