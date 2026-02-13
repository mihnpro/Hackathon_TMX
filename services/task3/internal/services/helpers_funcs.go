package services

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mihnpro/Hackathon_TMX/internal/domain"
)

func cleanStops(stations []string) []string {
	if len(stations) == 0 {
		return stations
	}

	var result []string
	for i, s := range stations {
		if i == 0 || s != stations[i-1] {
			result = append(result, s)
		}
	}
	return result
}

func splitIntoTrips(records []domain.Record) []domain.Trip {
	var trips []domain.Trip

	if len(records) == 0 {
		return trips
	}

	currentTrip := domain.Trip{
		StartTime:  records[0].Time,
		Stations:   []string{},
		IsComplete: false,
	}

	for i, rec := range records {
		currentTrip.Stations = append(currentTrip.Stations, rec.Station)

		// Проверяем, вернулись ли в депо
		if rec.Station == rec.Depo {
			currentTrip.EndTime = rec.Time
			currentTrip.IsComplete = true
			trips = append(trips, currentTrip)

			// Начинаем новую поездку
			if i < len(records)-1 {
				currentTrip = domain.Trip{
					StartTime:  records[i+1].Time,
					Stations:   []string{},
					IsComplete: false,
				}
			}
		}
	}

	// Добавляем последнюю незавершенную поездку
	if !currentTrip.IsComplete && len(currentTrip.Stations) > 0 {
		trips = append(trips, currentTrip)
	}

	return trips
}

func loadData(filename string) map[string]domain.Locomotive {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	locomotives := make(map[string]domain.Locomotive)

	// Пропускаем заголовок
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")

		if len(parts) < 5 {
			continue
		}

		// Очищаем поля от пробелов
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		// Парсим время
		timeStr := parts[2]
		t, err := time.Parse("2006-01-02T15:04:05.000000", timeStr)
		if err != nil {
			continue
		}

		record := domain.Record{
			Series:  parts[0],
			Number:  parts[1],
			Time:    t,
			Station: parts[3],
			Depo:    parts[4],
		}

		key := record.Series + "-" + record.Number

		if loc, exists := locomotives[key]; exists {
			loc.Records = append(loc.Records, record)
			locomotives[key] = loc
		} else {
			locomotives[key] = domain.Locomotive{
				Key:     key,
				Series:  record.Series,
				Number:  record.Number,
				Depo:    record.Depo,
				Records: []domain.Record{record},
			}
		}
	}

	// Сортируем записи каждого локомотива по времени
	for key, loc := range locomotives {
		sort.Slice(loc.Records, func(i, j int) bool {
			return loc.Records[i].Time.Before(loc.Records[j].Time)
		})
		locomotives[key] = loc
	}

	return locomotives
}
