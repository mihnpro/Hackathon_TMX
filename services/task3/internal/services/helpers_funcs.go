package services

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mihnpro/Hackathon_TMX/internal/domain"
)

// cleanStops удаляет повторяющиеся станции (стоянки)
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

// splitIntoTrips разбивает записи на поездки
func splitIntoTrips(records []domain.Record) []domain.Trip {
	var trips []domain.Trip

	if len(records) == 0 {
		return trips
	}

	currentTrip := domain.Trip{
		StartTime:  records[0].Timestamp,
		Stations:   []string{},
	}

	for i, rec := range records {
		currentTrip.Stations = append(currentTrip.Stations, rec.Station)

		// Проверяем, вернулись ли в депо
		if rec.Station == rec.Depo {
			currentTrip.EndTime = rec.Timestamp
			trips = append(trips, currentTrip)

			// Начинаем новую поездку, если есть следующие записи
			if i < len(records)-1 {
				currentTrip = domain.Trip{
					StartTime:  records[i+1].Timestamp,
					Stations:   []string{},
				}
			}
		}
	}

	// Добавляем последнюю незавершенную поездку, если в ней есть станции
	if len(currentTrip.Stations) > 0 && (len(trips) == 0 || trips[len(trips)-1].StartTime != currentTrip.StartTime) {
		trips = append(trips, currentTrip)
	}

	return trips
}

// loadData загружает данные о локомотивах из CSV файла
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
			// Пробуем альтернативный формат
			t, err = time.Parse("2006-01-02T15:04:05", timeStr)
			if err != nil {
				continue
			}
		}

		record := domain.Record{
			Series:    parts[0],
			Number:    parts[1],
			Timestamp: t,
			Station:   parts[3],
			Depo:      parts[4],
		}

		key := record.Series + "-" + record.Number

		if loc, exists := locomotives[key]; exists {
			loc.Records = append(loc.Records, record)
			locomotives[key] = loc
		} else {
			locomotives[key] = domain.Locomotive{
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
			return loc.Records[i].Timestamp.Before(loc.Records[j].Timestamp)
		})
		locomotives[key] = loc
	}

	return locomotives
}

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

// loadStationCoordinates загружает координаты станций из CSV файла
func loadStationCoordinates(filename string) (map[string]domain.Station, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Используем csv.Reader для правильной обработки CSV
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1

	// Читаем заголовок
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	stations := make(map[string]domain.Station)
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Предупреждение: ошибка чтения строки %d: %v\n", lineNum, err)
			continue
		}
		lineNum++

		if len(record) < 4 {
			fmt.Printf("Предупреждение: строка %d имеет меньше 4 полей\n", lineNum)
			continue
		}

		stationID := strings.TrimSpace(record[0])
		stationName := strings.TrimSpace(record[1])
		latStr := strings.TrimSpace(record[2])
		lonStr := strings.TrimSpace(record[3])

		// Пропускаем станции без координат
		if latStr == "" || lonStr == "" {
			continue
		}

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