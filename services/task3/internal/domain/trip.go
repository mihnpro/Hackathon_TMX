package domain

import "time"

type Trip struct {
    StartTime   time.Time
    EndTime     time.Time
    Stations    []string      // последовательность станций
    StationNames []string     // названия станций для отображения
    Route       []string      // очищенный маршрут (без повторов)
    DirectionID string        // ID направления этой поездки
}
