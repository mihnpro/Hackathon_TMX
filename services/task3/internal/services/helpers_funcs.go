package services

import (
	"bufio"
	"os"
	"sort"
	"strconv"
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

// --- НОВЫЕ ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ---

// filterLocomotivesByDepo фильтрует локомотивы по заданному депо
func filterLocomotivesByDepo(locomotives map[string]domain.Locomotive, depoID string) map[string]domain.Locomotive {
	filtered := make(map[string]domain.Locomotive)
	for key, loc := range locomotives {
		if loc.Depo == depoID {
			filtered[key] = loc
		}
	}
	return filtered
}

// countVisits считает количество посещений каждой станции
func countVisits(locomotives map[string]domain.Locomotive) map[string]int {
	visits := make(map[string]int)

	for _, loc := range locomotives {
		for _, trip := range loc.Trips {
			// Убираем стоянки для подсчета уникальных посещений за поездку
			seen := make(map[string]bool)
			for _, station := range trip.Stations {
				if !seen[station] {
					visits[station]++
					seen[station] = true
				}
			}
		}
	}

	return visits
}

// getUniqueStations возвращает все уникальные станции из поездок
func getUniqueStations(locomotives map[string]domain.Locomotive) map[string]bool {
	stations := make(map[string]bool)

	for _, loc := range locomotives {
		for _, trip := range loc.Trips {
			for _, station := range trip.Stations {
				stations[station] = true
			}
		}
	}

	return stations
}

// contains проверяет, есть ли элемент в срезе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// removeDuplicates удаляет дубликаты из среза (сохраняя порядок)
func removeDuplicates(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	result := []string{slice[0]}
	for i := 1; i < len(slice); i++ {
		if slice[i] != slice[i-1] {
			result = append(result, slice[i])
		}
	}
	return result
}

// getStationPrefix извлекает префикс станции (первые 2 цифры)
func getStationPrefix(stationID string) string {
	if len(stationID) >= 2 {
		return stationID[:2]
	}
	return stationID
}

// groupByPrefix группирует станции по префиксу
func groupByPrefix(stations []string) map[string][]string {
	groups := make(map[string][]string)

	for _, station := range stations {
		prefix := getStationPrefix(station)
		groups[prefix] = append(groups[prefix], station)
	}

	// Сортируем каждую группу
	for prefix := range groups {
		sort.Strings(groups[prefix])
	}

	return groups
}

// loadStationCoordinates загружает координаты станций из CSV файла
func loadStationCoordinates(filename string) (map[string]domain.Station, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	stations := make(map[string]domain.Station)

	// Пропускаем заголовок
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")

		if len(parts) < 4 {
			continue
		}

		// Очищаем поля
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		// Парсим координаты
		lat, _ := strconv.ParseFloat(parts[2], 64)
		lon, _ := strconv.ParseFloat(parts[3], 64)

		stations[parts[0]] = domain.Station{
			ID:        parts[0],
			Name:      parts[1],
			Latitude:  lat,
			Longitude: lon,
		}
	}

	return stations, nil
}

// generateBranchID создает ID ветки из первой и последней станции
func generateBranchID(stations []string) string {
	if len(stations) == 0 {
		return "unknown"
	}
	if len(stations) == 1 {
		return stations[0]
	}
	return stations[0] + "_to_" + stations[len(stations)-1]
}

// getTopLocomotives возвращает топ N локомотивов по количеству поездок
func getTopLocomotives(locomotives map[string]domain.Locomotive, n int) []string {
	type locActivity struct {
		key   string
		count int
	}

	var activities []locActivity
	for key, loc := range locomotives {
		activities = append(activities, locActivity{key, len(loc.Trips)})
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].count > activities[j].count
	})

	var result []string
	for i, act := range activities {
		if i >= n {
			break
		}
		result = append(result, act.key)
	}

	return result
}

// calculateCenter вычисляет центр карты на основе станций
func calculateCenter(stations map[string]domain.Station, defaultLat, defaultLon float64) (float64, float64) {
	if len(stations) == 0 {
		return defaultLat, defaultLon
	}

	var sumLat, sumLon float64
	count := 0

	for _, station := range stations {
		if station.Latitude != 0 && station.Longitude != 0 {
			sumLat += station.Latitude
			sumLon += station.Longitude
			count++
		}
	}

	if count > 0 {
		return sumLat / float64(count), sumLon / float64(count)
	}

	return defaultLat, defaultLon
}
