package domain

import "time"

type Record struct {
	Series  string
	Number  string
	Time    time.Time
	Station string
	Depo    string
}
