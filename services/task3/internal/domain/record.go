package domain

import "time"

type Record struct {
    Series    string
    Number    string
    Timestamp time.Time
    Station   string
    Depo      string
}