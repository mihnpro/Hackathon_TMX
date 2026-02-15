package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mihnpro/Hackathon_TMX/internal/domain"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/models/responses"
)

type visualizationService struct {
	dataPath string
	mapsDir  string // директория для карт (./maps)
}

type VisualizationService interface {
	// Существующие методы для консольного режима
	GenerateMap(depoID string, maxLocomotives int) error
	GenerateHeatmap(depoID string) error
	GenerateLocomotiveMap(locomotiveKey string) error
	GenerateAllMaps(depoID string, maxLocomotives int) error

	// Методы для API режима
	GenerateMapsAPI(depoID string, maxLocomotives int) (*responses.GenerateMapsResponse, error)
	GetAvailableDepots() ([]string, error)
	GetDepotInfo(depoID string) (*responses.DepotInfo, error)
	GetMapsDir() string
	Cleanup()
}

// JSStation структура для передачи станций в JavaScript
type JSStation struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Coords []float64 `json:"coords"`
	Size   float64   `json:"size"`
	Visits int       `json:"visits"`
	Color  string    `json:"color"`
}

// JSRoute структура для передачи маршрутов в JavaScript
type JSRoute struct {
	Points     [][]float64 `json:"points"`
	Color      string      `json:"color"`
	Locomotive string      `json:"locomotive"`
}

func NewVisualizationService(dataPath string) VisualizationService {
	// Создаем директорию ./maps если её нет
	mapsDir := "./maps"
	if err := os.MkdirAll(mapsDir, 0755); err != nil {
		fmt.Printf("⚠️ Не удалось создать директорию %s: %v\n", mapsDir, err)
	} else {
		fmt.Printf("📁 Директория для карт: %s\n", mapsDir)
	}

	return &visualizationService{
		dataPath: dataPath,
		mapsDir:  mapsDir,
	}
}

// GetMapsDir возвращает путь к директории с картами
func (v *visualizationService) GetMapsDir() string {
	return v.mapsDir
}

// Cleanup очищает директорию с картами
func (v *visualizationService) Cleanup() {
	fmt.Printf("🧹 Очистка директории %s\n", v.mapsDir)
	if err := os.RemoveAll(v.mapsDir); err != nil {
		fmt.Printf("⚠️ Ошибка при удалении %s: %v\n", v.mapsDir, err)
	} else {
		fmt.Printf("✅ Директория %s удалена\n", v.mapsDir)
	}
}

// GetAvailableDepots возвращает список всех депо
func (v *visualizationService) GetAvailableDepots() ([]string, error) {
	locomotives := loadData(v.dataPath)
	
	depoSet := make(map[string]bool)
	for _, loc := range locomotives {
		depoSet[loc.Depo] = true
	}
	
	depots := make([]string, 0, len(depoSet))
	for d := range depoSet {
		depots = append(depots, d)
	}
	sort.Strings(depots)
	
	return depots, nil
}

// GetDepotInfo возвращает информацию о депо
func (v *visualizationService) GetDepotInfo(depoID string) (*responses.DepotInfo, error) {
	locomotives := loadData(v.dataPath)

	// Считаем локомотивы в депо
	count := 0
	for _, loc := range locomotives {
		if loc.Depo == depoID {
			count++
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("депо %s не найдено", depoID)
	}

	return &responses.DepotInfo{
		DepoID:          depoID,
		Region:          getRegionByDepo(depoID),
		LocomotiveCount: count,
	}, nil
}

// cleanupOldMaps удаляет старые файлы карт для конкретного депо
func (v *visualizationService) cleanupOldMaps(depoID string) {
	pattern := fmt.Sprintf("depot_%s_*.html", depoID)
	files, err := filepath.Glob(filepath.Join(v.mapsDir, pattern))
	if err != nil {
		return
	}
	
	// Также удаляем старые карты локомотивов этого депо
	// (мы не знаем точные имена, поэтому удалим все .html файлы)
	// Но чтобы не удалять чужие, можно удалить только те, которые содержат локомотивы из этого депо
	// Для простоты удалим все .html файлы в директории
	allHTML, _ := filepath.Glob(filepath.Join(v.mapsDir, "*.html"))
	files = append(files, allHTML...)
	
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fmt.Printf("⚠️ Не удалось удалить %s: %v\n", f, err)
		} else {
			fmt.Printf("   🗑️ Удален старый файл: %s\n", filepath.Base(f))
		}
	}
}

// GenerateMapsAPI - для API режима (генерирует карты в ./maps)
func (v *visualizationService) GenerateMapsAPI(depoID string, maxLocomotives int) (*responses.GenerateMapsResponse, error) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("🚀 ЗАПУСК ГЕНЕРАЦИИ КАРТ ДЛЯ ДЕПО %s\n", depoID)
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	
	fmt.Printf("📁 Директория для карт: %s\n", v.mapsDir)
	
	// Проверяем существование директории
	if err := os.MkdirAll(v.mapsDir, 0755); err != nil {
		return nil, fmt.Errorf("❌ не удалось создать директорию %s: %w", v.mapsDir, err)
	}
	
	// Удаляем старые карты перед генерацией новых
	fmt.Println("🧹 Очистка старых файлов...")
	v.cleanupOldMaps(depoID)
	fmt.Printf("✅ Директория очищена\n")

	// 1. Загружаем данные
	fmt.Println("1️⃣ Загрузка данных...")
	locomotives := loadData(v.dataPath)
	fmt.Printf("   Загружено локомотивов: %d\n", len(locomotives))

	// 2. Разбиваем на поездки
	fmt.Println("2️⃣ Разбиение на поездки...")
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Фильтруем локомотивы выбранного депо
	fmt.Printf("3️⃣ Фильтрация локомотивов депо %s...\n", depoID)
	depoLocomotives := filterLocomotivesByDepo(locomotives, depoID)
	if len(depoLocomotives) == 0 {
		return nil, fmt.Errorf("❌ депо %s не найдено или нет локомотивов", depoID)
	}
	fmt.Printf("   Найдено локомотивов в депо: %d\n", len(depoLocomotives))

	// 4. Получаем координаты станций
	fmt.Println("4️⃣ Получение координат станций...")
	stations := v.getStationCoordinates(depoID)
	fmt.Printf("   Загружено станций с координатами: %d\n", len(stations))

	// 5. Собираем статистику посещений
	fmt.Println("5️⃣ Сбор статистики посещений...")
	stationStats := v.collectStationStats(depoLocomotives, stations)

	// 6. Строим маршруты для локомотивов
	fmt.Println("6️⃣ Построение маршрутов...")
	routes := v.buildLocomotiveRoutes(depoLocomotives, stations)

	// 7. Сортируем локомотивы по активности
	fmt.Println("7️⃣ Сортировка локомотивов...")
	topLocomotives := getTopLocomotives(depoLocomotives, maxLocomotives)
	fmt.Printf("   Топ-%d локомотивов:\n", len(topLocomotives))
	for i, loc := range topLocomotives {
		fmt.Printf("      %d. %s\n", i+1, loc)
	}

	// 8. Генерируем HTML карты
	fmt.Println("8️⃣ Генерация общей карты...")
	overviewURL, err := v.generateHTMLMapAPI(depoID, stationStats, routes, topLocomotives, stations)
	if err != nil {
		return nil, fmt.Errorf("❌ ошибка генерации общей карты: %w", err)
	}
	fmt.Printf("   ✅ Общая карта: %s\n", overviewURL)

	fmt.Println("9️⃣ Генерация тепловой карты...")
	heatmapURL, err := v.generateHeatmapHTMLAPI(depoID, stationStats)
	if err != nil {
		return nil, fmt.Errorf("❌ ошибка генерации тепловой карты: %w", err)
	}
	fmt.Printf("   ✅ Тепловая карта: %s\n", heatmapURL)

	// 9. Генерируем карты для топ локомотивов
	fmt.Println("🔟 Генерация карт локомотивов...")
	var locoMaps []responses.LocomotiveMap
	for i, locKey := range topLocomotives {
		if i >= maxLocomotives {
			break
		}
		
		fmt.Printf("   Генерация для %s... ", locKey)
		loc := depoLocomotives[locKey]
		locoURL, err := v.generateLocomotiveHTMLAPI(locKey, loc, stations)
		if err != nil {
			fmt.Printf("❌ ошибка: %v\n", err)
			continue
		}
		fmt.Printf("✅ %s\n", locoURL)
		
		locoMaps = append(locoMaps, responses.LocomotiveMap{
			Key:       locKey,
			Model:     loc.Series,
			Number:    loc.Number,
			URL:       locoURL,
			TripCount: len(loc.Trips),
		})
	}

	// Проверяем созданные файлы
	fmt.Println("\n📄 Проверка созданных файлов:")
	files, err := os.ReadDir(v.mapsDir)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения директории: %v\n", err)
	} else {
		fmt.Printf("   Найдено файлов: %d\n", len(files))
		for _, f := range files {
			info, _ := f.Info()
			fmt.Printf("   - %s (%d bytes)\n", f.Name(), info.Size())
		}
	}

	// 10. Формируем ответ
	response := &responses.GenerateMapsResponse{
		DepotID:     depoID,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Maps: responses.MapsList{
			Overview:    overviewURL,
			Heatmap:     heatmapURL,
			Locomotives: locoMaps,
		},
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("✅ ГЕНЕРАЦИЯ ЗАВЕРШЕНА УСПЕШНО\n")
	fmt.Printf("📁 Файлы сохранены в: %s\n", v.mapsDir)
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	return response, nil
}

// ==================== Существующие методы (с изменением пути сохранения) ====================

// GenerateMap создает карту для депо (консольный режим, сохраняет в ./maps)
func (v *visualizationService) GenerateMap(depoID string, maxLocomotives int) error {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("ПУНКТ 3: ВИЗУАЛИЗАЦИЯ ДЕПО %s\n", depoID)
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// 1. Загружаем данные
	locomotives := loadData(v.dataPath)

	// 2. Разбиваем на поездки
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Фильтруем локомотивы выбранного депо
	depoLocomotives := filterLocomotivesByDepo(locomotives, depoID)
	if len(depoLocomotives) == 0 {
		return fmt.Errorf("депо %s не найдено или нет локомотивов", depoID)
	}
	fmt.Printf("Найдено локомотивов в депо: %d\n", len(depoLocomotives))

	// 4. Получаем координаты станций
	stations := v.getStationCoordinates(depoID)
	fmt.Printf("Загружено станций с координатами: %d\n", len(stations))

	// 5. Собираем статистику посещений
	stationStats := v.collectStationStats(depoLocomotives, stations)

	// 6. Строим маршруты для локомотивов
	routes := v.buildLocomotiveRoutes(depoLocomotives, stations)

	// 7. Сортируем локомотивы по активности
	topLocomotives := getTopLocomotives(depoLocomotives, maxLocomotives)

	// 8. Генерируем HTML карту (используем старый метод, но он должен сохранять в v.mapsDir)
	err := v.generateHTMLMap(depoID, stationStats, routes, topLocomotives, stations)
	if err != nil {
		return fmt.Errorf("ошибка генерации карты: %w", err)
	}

	fmt.Printf("✅ Карта сохранена: %s/depot_%s_map.html\n", v.mapsDir, depoID)
	return nil
}

// GenerateHeatmap создает тепловую карту (консольный режим)
func (v *visualizationService) GenerateHeatmap(depoID string) error {
	locomotives := loadData(v.dataPath)

	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	depoLocomotives := filterLocomotivesByDepo(locomotives, depoID)
	stations := v.getStationCoordinates(depoID)
	stationStats := v.collectStationStats(depoLocomotives, stations)

	return v.generateHeatmapHTML(depoID, stationStats)
}

// GenerateLocomotiveMap создает карту для конкретного локомотива (консольный режим)
func (v *visualizationService) GenerateLocomotiveMap(locomotiveKey string) error {
	locomotives := loadData(v.dataPath)

	loc, exists := locomotives[locomotiveKey]
	if !exists {
		return fmt.Errorf("локомотив %s не найден", locomotiveKey)
	}

	loc.Trips = splitIntoTrips(loc.Records)
	stations := v.getStationCoordinates(loc.Depo)

	return v.generateLocomotiveHTML(locomotiveKey, loc, stations)
}

// GenerateAllMaps генерирует все карты для депо (консольный режим)
func (v *visualizationService) GenerateAllMaps(depoID string, maxLocomotives int) error {
	// Общая карта с топ-10 локомотивами
	if err := v.GenerateMap(depoID, maxLocomotives); err != nil {
		return err
	}

	// Тепловая карта
	if err := v.GenerateHeatmap(depoID); err != nil {
		return err
	}

	// Карты для топ-5 локомотивов
	locomotives := loadData(v.dataPath)
	depoLocomotives := filterLocomotivesByDepo(locomotives, depoID)

	// Сортируем по количеству поездок
	type locActivity struct {
		key   string
		trips int
	}
	var activities []locActivity
	for key, loc := range depoLocomotives {
		activities = append(activities, locActivity{key, len(loc.Trips)})
	}
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].trips > activities[j].trips
	})

	// Генерируем для топ-5
	for i, act := range activities {
		if i >= maxLocomotives {
			break
		}
		if err := v.GenerateLocomotiveMap(act.key); err != nil {
			fmt.Printf("Ошибка для %s: %v\n", act.key, err)
		}
	}

	fmt.Printf("✅ Все карты для депо %s сгенерированы\n", depoID)
	return nil
}

// ==================== Вспомогательные методы ====================

// getStationCoordinates получает координаты станций
func (v *visualizationService) getStationCoordinates(depoID string) map[string]domain.Station {
	stations := make(map[string]domain.Station)

	// Используем station_info.csv
	coordsFile := "./data/station_info.csv"

	// Проверяем существование файла
	if _, err := os.Stat(coordsFile); os.IsNotExist(err) {
		// Пробуем альтернативные пути
		altPaths := []string{
			"./data/station_info.csv",
			"../data/station_info.csv",
			"../../data/station_info.csv",
		}

		for _, path := range altPaths {
			if _, err := os.Stat(path); err == nil {
				coordsFile = path
				break
			}
		}
	}

	fmt.Printf("Загрузка станций из: %s\n", coordsFile)

	if loadedStations, err := loadStationCoordinates(coordsFile); err == nil {
		stations = loadedStations
	} else {
		fmt.Printf("⚠️ Не удалось загрузить station_info.csv: %v\n", err)
		fmt.Println("Используются тестовые координаты")
		stations = v.generateTestCoordinates(depoID)
	}

	return stations
}

// generateTestCoordinates создает тестовые координаты
func (v *visualizationService) generateTestCoordinates(depoID string) map[string]domain.Station {
	stations := make(map[string]domain.Station)

	// Центр для разных депо
	centers := map[string][2]float64{
		"940006": {55.75, 37.62}, // Москва
		"580003": {55.85, 37.95},
		"50009":  {55.65, 37.45},
		"254905": {55.70, 37.80},
		"304606": {55.80, 37.70},
	}

	center, exists := centers[depoID]
	if !exists {
		center = [2]float64{55.75, 37.62}
	}

	// Генерируем станции по направлениям
	directions := []struct {
		angle  float64
		count  int
		prefix string
	}{
		{0, 10, "94"},   // Запад
		{45, 8, "95"},   // Северо-запад
		{90, 12, "24"},  // Восток
		{135, 7, "25"},  // Юго-восток
		{180, 9, "30"},  // Юг
		{225, 6, "31"},  // Юго-запад
		{270, 11, "50"}, // Север
		{315, 5, "51"},  // Северо-восток
	}

	stationID := 1
	for _, dir := range directions {
		for i := 0; i < dir.count; i++ {
			// Станции располагаются на удалении от центра
			distance := 0.05 + float64(i)*0.03
			lat := center[0] + distance*math.Cos(dir.angle*math.Pi/180)
			lon := center[1] + distance*math.Sin(dir.angle*math.Pi/180)

			id := fmt.Sprintf("%s%04d", dir.prefix, stationID)
			stations[id] = domain.Station{
				ID:        id,
				Name:      fmt.Sprintf("Станция %s", id),
				Latitude:  lat,
				Longitude: lon,
			}
			stationID++
		}
	}

	// Добавляем само депо
	stations[depoID] = domain.Station{
		ID:        depoID,
		Name:      fmt.Sprintf("Депо %s", depoID),
		Latitude:  center[0],
		Longitude: center[1],
	}

	return stations
}

// collectStationStats собирает статистику посещений (только для станций с координатами)
func (v *visualizationService) collectStationStats(
	locomotives map[string]domain.Locomotive,
	stations map[string]domain.Station) map[string]*domain.StationStats {

	stats := make(map[string]*domain.StationStats)

	// Инициализируем статистику только для станций с координатами
	for id, station := range stations {
		stats[id] = &domain.StationStats{
			StationID:   id,
			StationName: station.Name,
			Latitude:    station.Latitude,
			Longitude:   station.Longitude,
			VisitCount:  0,
			Locomotives: []string{},
		}
	}

	// Собираем посещения
	maxVisits := 0
	for locKey, loc := range locomotives {
		for _, trip := range loc.Trips {
			seen := make(map[string]bool)
			for _, stationID := range trip.Stations {
				// Проверяем, есть ли станция в нашей карте (имеет ли координаты)
				if stat, exists := stats[stationID]; exists && !seen[stationID] {
					stat.VisitCount++
					seen[stationID] = true
					if stat.VisitCount > maxVisits {
						maxVisits = stat.VisitCount
					}
					// Добавляем локомотив, если еще не добавлен
					found := false
					for _, l := range stat.Locomotives {
						if l == locKey {
							found = true
							break
						}
					}
					if !found {
						stat.Locomotives = append(stat.Locomotives, locKey)
					}
				}
			}
		}
	}

	// Нормализуем популярность
	if maxVisits > 0 {
		for _, stat := range stats {
			stat.Popularity = float64(stat.VisitCount) / float64(maxVisits)
		}
	}

	return stats
}

// buildLocomotiveRoutes строит маршруты локомотивов (только через станции с координатами)
func (v *visualizationService) buildLocomotiveRoutes(
	locomotives map[string]domain.Locomotive,
	stations map[string]domain.Station) map[string][]domain.LocomotiveRoute {

	routes := make(map[string][]domain.LocomotiveRoute)

	for locKey, loc := range locomotives {
		var locRoutes []domain.LocomotiveRoute

		for _, trip := range loc.Trips {
			if len(trip.Stations) < 2 {
				continue
			}

			// Очищаем от стоянок
			var cleanStations []string
			for i, s := range trip.Stations {
				if i == 0 || s != trip.Stations[i-1] {
					cleanStations = append(cleanStations, s)
				}
			}

			// Создаем точки маршрута (только для станций с координатами)
			var points []domain.RoutePoint
			validPoints := 0

			for order, stationID := range cleanStations {
				if station, exists := stations[stationID]; exists {
					points = append(points, domain.RoutePoint{
						StationID: stationID,
						Lat:       station.Latitude,
						Lon:       station.Longitude,
						Order:     order,
					})
					validPoints++
				}
			}

			// Добавляем маршрут только если есть хотя бы 2 точки с координатами
			if validPoints > 1 {
				locRoutes = append(locRoutes, domain.LocomotiveRoute{
					LocomotiveKey: locKey,
					Model:         loc.Series,
					Number:        loc.Number,
					Points:        points,
					Trips:         1,
				})
			}
		}

		if len(locRoutes) > 0 {
			routes[locKey] = locRoutes
		}
	}

	return routes
}

// ==================== Методы генерации HTML для API (сохраняют в ./maps) ====================

// generateHTMLMapAPI создает HTML файл с картой в ./maps
func (v *visualizationService) generateHTMLMapAPI(
	depoID string,
	stationStats map[string]*domain.StationStats,
	routes map[string][]domain.LocomotiveRoute,
	topLocomotives []string,
	stations map[string]domain.Station) (string, error) {

	fmt.Printf("   📍 Генерация HTML карты для депо %s...\n", depoID)
	
	// Проверяем директорию
	if err := os.MkdirAll(v.mapsDir, 0755); err != nil {
		return "", fmt.Errorf("не удалось создать директорию: %w", err)
	}

	// Определяем регион депо
	depoRegion := getRegionByDepo(depoID)

	// Подготавливаем данные для JavaScript
	var jsRoutes []JSRoute
	var jsStations []JSStation

	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7", "#C7B198", "#DFC2C2", "#B2B2B2"}

	// Собираем активные станции
	activeStations := make(map[string]bool)

	for i, locKey := range topLocomotives {
		if routeList, exists := routes[locKey]; exists {
			color := colors[i%len(colors)]
			for _, route := range routeList {
				var points [][]float64
				for _, p := range route.Points {
					points = append(points, []float64{p.Lon, p.Lat})
					activeStations[p.StationID] = true
				}
				if len(points) > 1 {
					jsRoutes = append(jsRoutes, JSRoute{
						Points:     points,
						Color:      color,
						Locomotive: locKey,
					})
				}
			}
		}
	}
	fmt.Printf("      Собрано маршрутов: %d\n", len(jsRoutes))

	// Добавляем только активные станции
	for _, stat := range stationStats {
		if activeStations[stat.StationID] && stat.Latitude != 0 && stat.Longitude != 0 {
			size := 5 + stat.Popularity*25
			color := fmt.Sprintf("hsl(%d, 70%%, 50%%)", int(240*(1-stat.Popularity)))

			jsStations = append(jsStations, JSStation{
				ID:     stat.StationID,
				Name:   stat.StationName,
				Coords: []float64{stat.Longitude, stat.Latitude},
				Size:   size,
				Visits: stat.VisitCount,
				Color:  color,
			})
		}
	}
	fmt.Printf("      Собрано станций: %d\n", len(jsStations))

	// Определяем границы карты
	minLat, maxLat, minLon, maxLon := v.calculateBounds(jsStations, stations, depoID)

	latPadding := (maxLat - minLat) * 0.2
	lonPadding := (maxLon - minLon) * 0.2

	stationsJSON, _ := json.Marshal(jsStations)
	routesJSON, _ := json.Marshal(jsRoutes)

	// HTML шаблон
	html := v.generateMapHTMLTemplate(depoID, depoRegion, len(jsStations), len(jsRoutes),
		len(topLocomotives), topLocomotives, colors,
		minLat-latPadding, minLon-lonPadding, maxLat+latPadding, maxLon+lonPadding,
		stationsJSON, routesJSON)

	// Сохраняем в директорию
	filename := fmt.Sprintf("depot_%s_map.html", depoID)
	fullPath := filepath.Join(v.mapsDir, filename)

	fmt.Printf("      Сохранение в %s... ", fullPath)
	if err := os.WriteFile(fullPath, []byte(html), 0644); err != nil {
		fmt.Printf("❌ %v\n", err)
		return "", err
	}
	fmt.Printf("✅ (%d bytes)\n", len(html))

	// Проверяем что файл создан
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("файл не создан: %w", err)
	}

	return "/maps/" + filename, nil
}

// generateHeatmapHTMLAPI создает тепловую карту в ./maps
func (v *visualizationService) generateHeatmapHTMLAPI(
	depoID string,
	stationStats map[string]*domain.StationStats) (string, error) {

	// Собираем данные для тепловой карты
	var heatData [][]float64
	for _, stat := range stationStats {
		if stat.VisitCount > 0 {
			heatData = append(heatData, []float64{
				stat.Longitude,
				stat.Latitude,
				float64(stat.VisitCount),
			})
		}
	}

	heatDataJSON, _ := json.Marshal(heatData)
	centerLat, centerLon := 55.75, 37.62 // Москва по умолчанию

	// Находим центр по первой станции с координатами
	for _, stat := range stationStats {
		if stat.Latitude != 0 && stat.Longitude != 0 {
			centerLat, centerLon = stat.Latitude, stat.Longitude
			break
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Депо %s - Тепловая карта</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <script src="https://unpkg.com/leaflet.heat/dist/leaflet-heat.js"></script>
    <style>
        body { margin: 0; padding: 0; }
        #map { height: 100vh; width: 100vw; }
        /* Скрываем атрибуцию Leaflet */
        .leaflet-control-attribution {
            display: none !important;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    <script>
        var map = L.map('map').setView([%f, %f], 11);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: ''
        }).addTo(map);
        var heat = L.heatLayer(%s, {
            radius: 30,
            blur: 20,
            maxZoom: 12,
            gradient: {0.2: 'blue', 0.4: 'cyan', 0.6: 'lime', 0.8: 'yellow', 1.0: 'red'}
        }).addTo(map);
    </script>
</body>
</html>`, depoID, centerLat, centerLon, heatDataJSON)

	filename := fmt.Sprintf("depot_%s_heatmap.html", depoID)
	fullPath := filepath.Join(v.mapsDir, filename)

	if err := os.WriteFile(fullPath, []byte(html), 0644); err != nil {
		return "", err
	}

	return "/maps/" + filename, nil
}

// generateLocomotiveHTMLAPI создает карту для конкретного локомотива в ./maps
func (v *visualizationService) generateLocomotiveHTMLAPI(
	locomotiveKey string,
	loc domain.Locomotive,
	stations map[string]domain.Station) (string, error) {

	// Собираем все поездки
	var allPoints [][][]float64
	for _, trip := range loc.Trips {
		var points [][]float64
		cleanTrip := removeDuplicates(trip.Stations)
		for _, stationID := range cleanTrip {
			if station, exists := stations[stationID]; exists {
				points = append(points, []float64{station.Longitude, station.Latitude})
			}
		}
		if len(points) > 1 {
			allPoints = append(allPoints, points)
		}
	}

	// Уникальные станции
	uniqueStations := make(map[string]bool)
	for _, trip := range loc.Trips {
		for _, stationID := range trip.Stations {
			uniqueStations[stationID] = true
		}
	}

	var stationList []map[string]interface{}
	for stationID := range uniqueStations {
		if station, exists := stations[stationID]; exists {
			stationList = append(stationList, map[string]interface{}{
				"id":   stationID,
				"lat":  station.Latitude,
				"lon":  station.Longitude,
				"name": station.Name,
			})
		}
	}

	routesJSON, _ := json.Marshal(allPoints)
	stationsJSON, _ := json.Marshal(stationList)

	centerLat, centerLon := 55.75, 37.62
	if depoStation, exists := stations[loc.Depo]; exists {
		centerLat, centerLon = depoStation.Latitude, depoStation.Longitude
	}

	safeKey := strings.ReplaceAll(locomotiveKey, "-", "_")
	filename := fmt.Sprintf("locomotive_%s.html", safeKey)
	fullPath := filepath.Join(v.mapsDir, filename)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Локомотив %s</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <style>
        body { margin: 0; padding: 0; }
        #map { height: 100vh; width: 100vw; }
        .info {
            position: absolute;
            top: 10px;
            left: 10px;
            background: white;
            padding: 10px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
            z-index: 1000;
        }
        /* Скрываем атрибуцию Leaflet */
        .leaflet-control-attribution {
            display: none !important;
        }
    </style>
</head>
<body>
    <div class="info">
        <h3>Локомотив %s</h3>
        <p>Модель: %s<br>
           Номер: %s<br>
           Депо: %s<br>
           Поездок: %d<br>
           Станций: %d</p>
    </div>
    <div id="map"></div>
    <script>
        var map = L.map('map').setView([%f, %f], 11);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: ''
        }).addTo(map);

        var stations = %s;
        stations.forEach(function(s) {
            L.circleMarker([s.lat, s.lon], {
                radius: 6,
                color: '#3388ff',
                fillColor: '#3388ff',
                fillOpacity: 0.8
            }).bindPopup(s.id + '<br>' + s.name).addTo(map);
        });

        var colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#C7B198'];
        var routes = %s;
        routes.forEach(function(r, i) {
            L.polyline(r, {
                color: colors[i %% colors.length],
                weight: 3,
                opacity: 0.6
            }).addTo(map);
        });
    </script>
</body>
</html>`, locomotiveKey, locomotiveKey, loc.Series, loc.Number, loc.Depo,
		len(loc.Trips), len(uniqueStations), centerLat, centerLon, stationsJSON, routesJSON)

	if err := os.WriteFile(fullPath, []byte(html), 0644); err != nil {
		return "", err
	}

	return "/maps/" + filename, nil
}

// ==================== Методы генерации HTML для консольного режима ====================

// generateHTMLMap создает HTML файл с картой (консольный режим)
func (v *visualizationService) generateHTMLMap(
	depoID string,
	stationStats map[string]*domain.StationStats,
	routes map[string][]domain.LocomotiveRoute,
	topLocomotives []string,
	stations map[string]domain.Station) error {

	// Создаем директорию для карт
	if err := os.MkdirAll(v.mapsDir, 0755); err != nil {
		return err
	}

	// Определяем регион депо по его ID
	depoRegion := getRegionByDepo(depoID)
	fmt.Printf("📍 Депо %s находится в регионе: %s\n", depoID, depoRegion)

	// Подготавливаем данные для JavaScript
	var jsRoutes []JSRoute
	var jsStations []JSStation

	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7", "#C7B198", "#DFC2C2", "#B2B2B2"}

	// Сначала собираем все станции, которые реально посещаются в маршрутах
	activeStations := make(map[string]bool)

	for i, locKey := range topLocomotives {
		if routeList, exists := routes[locKey]; exists {
			color := colors[i%len(colors)]
			for _, route := range routeList {
				var points [][]float64
				for _, p := range route.Points {
					points = append(points, []float64{p.Lon, p.Lat})
					activeStations[p.StationID] = true
				}
				if len(points) > 1 {
					jsRoutes = append(jsRoutes, JSRoute{
						Points:     points,
						Color:      color,
						Locomotive: locKey,
					})
				}
			}
		}
	}

	// Добавляем только активные станции (которые есть в маршрутах)
	for _, stat := range stationStats {
		if activeStations[stat.StationID] && stat.Latitude != 0 && stat.Longitude != 0 {
			// Размер от 5 до 30 пикселей
			size := 5 + stat.Popularity*25
			// Цвет от синего к красному
			color := fmt.Sprintf("hsl(%d, 70%%, 50%%)", int(240*(1-stat.Popularity)))

			jsStations = append(jsStations, JSStation{
				ID:     stat.StationID,
				Name:   stat.StationName,
				Coords: []float64{stat.Longitude, stat.Latitude},
				Size:   size,
				Visits: stat.VisitCount,
				Color:  color,
			})
		}
	}

	// Определяем границы карты по активным станциям
	minLat, maxLat := 90.0, -90.0
	minLon, maxLon := 180.0, -180.0

	if len(jsStations) == 0 {
		// Если нет активных станций, используем координаты депо
		if depo, exists := stations[depoID]; exists {
			minLat, maxLat = depo.Latitude-0.5, depo.Latitude+0.5
			minLon, maxLon = depo.Longitude-0.5, depo.Longitude+0.5
		} else {
			// Запасной вариант
			minLat, maxLat = 55.0, 56.0
			minLon, maxLon = 37.0, 38.0
		}
	} else {
		for _, stat := range jsStations {
			if stat.Coords[1] < minLat {
				minLat = stat.Coords[1]
			}
			if stat.Coords[1] > maxLat {
				maxLat = stat.Coords[1]
			}
			if stat.Coords[0] < minLon {
				minLon = stat.Coords[0]
			}
			if stat.Coords[0] > maxLon {
				maxLon = stat.Coords[0]
			}
		}
	}

	// Добавляем отступы 20% для лучшего обзора
	latPadding := (maxLat - minLat) * 0.2
	lonPadding := (maxLon - minLon) * 0.2

	// Конвертируем в JSON
	stationsJSON, _ := json.Marshal(jsStations)
	routesJSON, _ := json.Marshal(jsRoutes)

	// HTML шаблон
	html := v.generateMapHTMLTemplate(depoID, depoRegion, len(jsStations), len(jsRoutes),
		len(topLocomotives), topLocomotives, colors,
		minLat-latPadding, minLon-lonPadding, maxLat+latPadding, maxLon+lonPadding,
		stationsJSON, routesJSON)

	// Сохраняем файл
	filename := filepath.Join(v.mapsDir, fmt.Sprintf("depot_%s_map.html", depoID))
	return os.WriteFile(filename, []byte(html), 0644)
}

// generateHeatmapHTML создает тепловую карту (консольный режим)
func (v *visualizationService) generateHeatmapHTML(
	depoID string,
	stationStats map[string]*domain.StationStats) error {

	// Создаем директорию для карт
	if err := os.MkdirAll(v.mapsDir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(v.mapsDir, fmt.Sprintf("depot_%s_heatmap.html", depoID))

	// Собираем данные для тепловой карты
	var heatData [][]float64
	for _, stat := range stationStats {
		if stat.VisitCount > 0 {
			heatData = append(heatData, []float64{
				stat.Longitude,
				stat.Latitude,
				float64(stat.VisitCount),
			})
		}
	}

	heatDataJSON, _ := json.Marshal(heatData)
	centerLat, centerLon := 55.75, 37.62

	// Находим центр (берем первую станцию с координатами)
	for _, stat := range stationStats {
		if stat.Latitude != 0 && stat.Longitude != 0 {
			centerLat, centerLon = stat.Latitude, stat.Longitude
			break
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Депо %s - Тепловая карта</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <script src="https://unpkg.com/leaflet.heat/dist/leaflet-heat.js"></script>
    <style>
        body { margin: 0; padding: 0; }
        #map { height: 100vh; width: 100vw; }
        /* Скрываем атрибуцию Leaflet */
        .leaflet-control-attribution {
            display: none !important;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    <script>
        var map = L.map('map').setView([%f, %f], 11);
        
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: ''
        }).addTo(map);

        var heat = L.heatLayer(%s, {
            radius: 30,
            blur: 20,
            maxZoom: 12,
            gradient: {0.2: 'blue', 0.4: 'cyan', 0.6: 'lime', 0.8: 'yellow', 1.0: 'red'}
        }).addTo(map);
    </script>
</body>
</html>`, depoID, centerLat, centerLon, heatDataJSON)

	return os.WriteFile(filename, []byte(html), 0644)
}

// generateLocomotiveHTML создает карту для конкретного локомотива (консольный режим)
func (v *visualizationService) generateLocomotiveHTML(
	locomotiveKey string,
	loc domain.Locomotive,
	stations map[string]domain.Station) error {

	// Создаем директорию для карт
	if err := os.MkdirAll(v.mapsDir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(v.mapsDir, fmt.Sprintf("locomotive_%s.html", strings.ReplaceAll(locomotiveKey, "-", "_")))

	// Собираем все поездки
	var allPoints [][][]float64
	for _, trip := range loc.Trips {
		var points [][]float64
		// Очищаем от стоянок
		cleanTrip := removeDuplicates(trip.Stations)
		for _, stationID := range cleanTrip {
			if station, exists := stations[stationID]; exists {
				points = append(points, []float64{station.Longitude, station.Latitude})
			}
		}
		if len(points) > 1 {
			allPoints = append(allPoints, points)
		}
	}

	// Уникальные станции
	uniqueStations := make(map[string]bool)
	for _, trip := range loc.Trips {
		for _, stationID := range trip.Stations {
			uniqueStations[stationID] = true
		}
	}

	var stationList []map[string]interface{}
	for stationID := range uniqueStations {
		if station, exists := stations[stationID]; exists {
			stationList = append(stationList, map[string]interface{}{
				"id":   stationID,
				"lat":  station.Latitude,
				"lon":  station.Longitude,
				"name": station.Name,
			})
		}
	}

	// Конвертируем в JSON
	routesJSON, _ := json.Marshal(allPoints)
	stationsJSON, _ := json.Marshal(stationList)

	// Центр карты
	centerLat, centerLon := 55.75, 37.62
	if depoStation, exists := stations[loc.Depo]; exists {
		centerLat, centerLon = depoStation.Latitude, depoStation.Longitude
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Локомотив %s</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <style>
        body { margin: 0; padding: 0; }
        #map { height: 100vh; width: 100vw; }
        .info {
            position: absolute;
            top: 10px;
            left: 10px;
            background: white;
            padding: 10px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
            z-index: 1000;
        }
        /* Скрываем атрибуцию Leaflet */
        .leaflet-control-attribution {
            display: none !important;
        }
    </style>
</head>
<body>
    <div class="info">
        <h3>Локомотив %s</h3>
        <p>Модель: %s<br>
           Номер: %s<br>
           Депо: %s<br>
           Поездок: %d<br>
           Станций: %d</p>
    </div>
    <div id="map"></div>
    <script>
        var map = L.map('map').setView([%f, %f], 11);
        
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: ''
        }).addTo(map);

        // Станции
        var stations = %s;
        stations.forEach(function(s) {
            L.circleMarker([s.lat, s.lon], {
                radius: 6,
                color: '#3388ff',
                fillColor: '#3388ff',
                fillOpacity: 0.8
            }).bindPopup(s.id + '<br>' + s.name).addTo(map);
        });

        // Маршруты (разные цвета для разных поездок)
        var colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#C7B198'];
        var routes = %s;
        routes.forEach(function(r, i) {
            L.polyline(r, {
                color: colors[i %% colors.length],
                weight: 3,
                opacity: 0.6
            }).addTo(map);
        });
    </script>
</body>
</html>`, locomotiveKey, locomotiveKey, loc.Series, loc.Number, loc.Depo, len(loc.Trips), len(uniqueStations),
		centerLat, centerLon, stationsJSON, routesJSON)

	return os.WriteFile(filename, []byte(html), 0644)
}

// ==================== Общие вспомогательные функции ====================

// getRegionByDepo определяет регион по ID депо
func getRegionByDepo(depoID string) string {
	// По первой цифре ID депо можно примерно определить регион
	if len(depoID) < 2 {
		return "Неизвестно"
	}

	prefix := depoID[:2]

	regionMap := map[string]string{
		"94": "Забайкалье",
		"58": "Ростовская область",
		"51": "Ростовская область",
		"52": "Краснодарский край",
		"59": "Ростовская область",
		"60": "Пермский край",
		"61": "Свердловская область",
		"17": "Смоленская область",
		"78": "Свердловская область",
		"20": "Центральный регион",
		"21": "Центральный регион",
		"30": "Ярославская область",
		"31": "Ярославская область",
		"40": "Ленинградская область",
		"41": "Ленинградская область",
		"50": "Тверская область",
		"25": "Татарстан",
		"24": "Нижегородская область",
	}

	if region, exists := regionMap[prefix]; exists {
		return region
	}

	// По долготе депо (примерно)
	switch {
	case depoID >= "940000" && depoID < "950000":
		return "Забайкалье"
	case depoID >= "580000" && depoID < "590000":
		return "Ростовская область"
	case depoID >= "500000" && depoID < "510000":
		return "Тверская область"
	case depoID >= "200000" && depoID < "210000":
		return "Центральный регион"
	default:
		return "Россия"
	}
}

// calculateBounds определяет границы карты
func (v *visualizationService) calculateBounds(
	stations []JSStation,
	stationMap map[string]domain.Station,
	depoID string) (minLat, maxLat, minLon, maxLon float64) {

	minLat, maxLat = 90.0, -90.0
	minLon, maxLon = 180.0, -180.0

	if len(stations) == 0 {
		if depo, exists := stationMap[depoID]; exists {
			return depo.Latitude - 0.5, depo.Latitude + 0.5,
				depo.Longitude - 0.5, depo.Longitude + 0.5
		}
		return 55.0, 56.0, 37.0, 38.0
	}

	for _, s := range stations {
		if s.Coords[1] < minLat {
			minLat = s.Coords[1]
		}
		if s.Coords[1] > maxLat {
			maxLat = s.Coords[1]
		}
		if s.Coords[0] < minLon {
			minLon = s.Coords[0]
		}
		if s.Coords[0] > maxLon {
			maxLon = s.Coords[0]
		}
	}

	return minLat, maxLat, minLon, maxLon
}

// generateMapHTMLTemplate создает HTML шаблон карты
func (v *visualizationService) generateMapHTMLTemplate(
	depoID, depoRegion string,
	stationsCount, routesCount, topCount int,
	topLocomotives, colors []string,
	minLat, minLon, maxLat, maxLon float64,
	stationsJSON, routesJSON []byte) string {

	// HTML шаблон
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Депо %s - Карта маршрутов</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <script src="https://unpkg.com/leaflet.heat/dist/leaflet-heat.js"></script>
    <style>
        body { margin: 0; padding: 0; font-family: Arial; }
        #map { height: 100vh; width: 100vw; }
        /* Скрываем атрибуцию Leaflet */
        .leaflet-control-attribution {
            display: none !important;
        }
        .info-panel {
            position: absolute;
            top: 10px;
            right: 10px;
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
            z-index: 1000;
            max-height: 80vh;
            overflow-y: auto;
            width: 300px;
        }
        .legend { margin-top: 10px; padding: 10px 0; border-top: 1px solid #eee; }
        .legend-item { display: flex; align-items: center; margin: 5px 0; }
        .color-box { width: 20px; height: 20px; margin-right: 8px; border-radius: 4px; }
        .station-info {
            position: absolute; bottom: 30px; left: 10px; background: white;
            padding: 10px; border-radius: 4px; box-shadow: 0 2px 5px rgba(0,0,0,0.2);
            z-index: 1000; font-size: 12px; max-width: 300px; display: none;
        }
        .controls {
            position: absolute; top: 10px; left: 10px; background: white;
            padding: 10px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.2);
            z-index: 1000;
        }
        button {
            margin: 2px; padding: 5px 10px; cursor: pointer;
            background: #f0f0f0; border: 1px solid #ccc; border-radius: 4px;
        }
        button:hover { background: #e0e0e0; }
        .stats { font-size: 12px; color: #666; margin-top: 10px; }
    </style>
</head>
<body>
    <div id="map"></div>
    
    <div class="info-panel">
        <h3>Депо %s</h3>
        <p>Регион: %s</p>
        <p>Станций на карте: %d<br>
           Маршрутов: %d<br>
           Топ-%d локомотивов</p>
        
        <div class="legend">
            <h4>Цвета маршрутов:</h4>`, depoID, depoID, depoRegion, stationsCount, routesCount, topCount)

	// Добавляем легенду для каждого локомотива
	for i, locKey := range topLocomotives {
		html += fmt.Sprintf(`
            <div class="legend-item">
                <div class="color-box" style="background: %s"></div>
                <span>%s</span>
            </div>`, colors[i%len(colors)], locKey)
	}

	html += fmt.Sprintf(`
        </div>
        
        <div class="legend">
            <h4>Популярность станций:</h4>
            <div class="legend-item"><div class="color-box" style="background: #ff0000"></div> Высокая</div>
            <div class="legend-item"><div class="color-box" style="background: #ffaa00"></div> Средняя</div>
            <div class="legend-item"><div class="color-box" style="background: #0000ff"></div> Низкая</div>
        </div>
        
        <div class="stats">
            <p>💡 Показаны только станции, которые посещают локомотивы депо.</p>
        </div>
    </div>

    <div class="controls">
        <button onclick="toggleHeatmap()">🔥 Тепловая карта</button>
        <button onclick="toggleRoutes()">🛤️ Маршруты</button>
        <button onclick="toggleStations()">📍 Станции</button>
        <button onclick="resetView()">🗺️ Сброс вида</button>
    </div>

    <div id="stationInfo" class="station-info"></div>

    <script>
        var map = L.map('map').fitBounds([[%f, %f], [%f, %f]]);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: ''
        }).addTo(map);

        var stationLayer = L.layerGroup();
        var routeLayer = L.layerGroup();
        var heatLayer = null;

        var stations = %s;
        stations.forEach(function(s) {
            var marker = L.circleMarker([s.coords[1], s.coords[0]], {
                radius: s.size,
                color: s.color,
                fillColor: s.color,
                fillOpacity: 0.8,
                weight: 1
            }).bindPopup('<b>' + s.id + '</b><br>' + s.name + '<br>Посещений: ' + s.visits);
            
            marker.on('mouseover', function(e) {
                document.getElementById('stationInfo').style.display = 'block';
                document.getElementById('stationInfo').innerHTML = '<b>' + s.id + '</b><br>' + s.name + '<br>Посещений: ' + s.visits;
                document.getElementById('stationInfo').style.left = (e.originalEvent.pageX + 10) + 'px';
                document.getElementById('stationInfo').style.top = (e.originalEvent.pageY - 40) + 'px';
            });
            
            marker.on('mouseout', function() {
                document.getElementById('stationInfo').style.display = 'none';
            });
            
            stationLayer.addLayer(marker);
        });
        stationLayer.addTo(map);

        var routes = %s;
        routes.forEach(function(r) {
            var points = r.points.map(function(p) { return [p[1], p[0]]; });
            var polyline = L.polyline(points, {
                color: r.color,
                weight: 3,
                opacity: 0.7
            }).bindPopup('Локомотив: ' + r.locomotive);
            routeLayer.addLayer(polyline);
        });
        routeLayer.addTo(map);

        var heatData = stations.map(function(s) {
            return [s.coords[1], s.coords[0], s.visits];
        });
        heatLayer = L.heatLayer(heatData, {
            radius: 20, blur: 15, maxZoom: 12,
            gradient: {0.2: 'blue', 0.4: 'cyan', 0.6: 'lime', 0.8: 'yellow', 1.0: 'red'}
        });

        function toggleHeatmap() {
            if (map.hasLayer(heatLayer)) map.removeLayer(heatLayer);
            else heatLayer.addTo(map);
        }

        function toggleRoutes() {
            if (map.hasLayer(routeLayer)) map.removeLayer(routeLayer);
            else routeLayer.addTo(map);
        }

        function toggleStations() {
            if (map.hasLayer(stationLayer)) map.removeLayer(stationLayer);
            else stationLayer.addTo(map);
        }

        function resetView() {
            map.fitBounds([[%f, %f], [%f, %f]]);
        }

        L.control.scale().addTo(map);
    </script>
</body>
</html>`, minLat, minLon, maxLat, maxLon, stationsJSON, routesJSON, minLat, minLon, maxLat, maxLon)

	return html
}