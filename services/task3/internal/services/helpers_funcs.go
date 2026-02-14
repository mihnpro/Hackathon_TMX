package services

import (
	"bufio"
	"fmt"
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


// loadStationCoordinates загружает координаты станций из CSV файла с форматом: station,station_name,latitude,longitude
func loadStationCoordinates(filename string) (map[string]domain.Station, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	stations := make(map[string]domain.Station)

	// Читаем заголовок
	scanner.Scan()

	lineNum := 1
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue // пропускаем пустые строки
		}

		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			fmt.Printf("Предупреждение: строка %d имеет меньше 4 полей\n", lineNum)
			continue
		}

		// Очищаем поля от пробелов
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		stationID := parts[0]
		stationName := parts[1]
		latStr := parts[2]
		lonStr := parts[3]

		// Проверяем, есть ли координаты
		if latStr == "" || lonStr == "" {
			// Пропускаем станции без координат
			continue
		}

		// Парсим координаты
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			fmt.Printf("Предупреждение: строка %d, ошибка парсинга широты: %s\n", lineNum, latStr)
			continue
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			fmt.Printf("Предупреждение: строка %d, ошибка парсинга долготы: %s\n", lineNum, lonStr)
			continue
		}

		stations[stationID] = domain.Station{
			ID:        stationID,
			Name:      stationName,
			Latitude:  lat,
			Longitude: lon,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	fmt.Printf("Загружено станций с координатами: %d\n", len(stations))
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

// calculateCenter вычисляет центр карты на основе станций с координатами
func calculateCenter(stations map[string]domain.Station, defaultLat, defaultLon float64) (float64, float64) {
	if len(stations) == 0 {
		return defaultLat, defaultLon
	}

	var sumLat, sumLon float64
	count := 0

	for _, station := range stations {
		// Проверяем, что координаты не нулевые (хотя мы уже отфильтровали)
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