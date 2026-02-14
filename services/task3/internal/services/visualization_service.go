package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/mihnpro/Hackathon_TMX/internal/domain"
)

type visualizationService struct {
	dataPath string
}

type VisualizationService interface {
	GenerateMap(depoID string, maxLocomotives int) error
	GenerateHeatmap(depoID string) error
	GenerateLocomotiveMap(locomotiveKey string) error
	GenerateAllMaps(depoID string) error
}

func NewVisualizationService(dataPath string) VisualizationService {
	return &visualizationService{
		dataPath: dataPath,
	}
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

// GenerateMap создает карту для депо
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

	// 8. Генерируем HTML карту
	err := v.generateHTMLMap(depoID, stationStats, routes, topLocomotives, stations)
	if err != nil {
		return fmt.Errorf("ошибка генерации карты: %w", err)
	}

	fmt.Printf("✅ Карта сохранена: maps/depot_%s_map.html\n", depoID)
	return nil
}

// GenerateHeatmap создает тепловую карту
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

// GenerateLocomotiveMap создает карту для конкретного локомотива
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

// GenerateAllMaps генерирует все карты для депо
func (v *visualizationService) GenerateAllMaps(depoID string) error {
	// Общая карта с топ-10 локомотивами
	if err := v.GenerateMap(depoID, 10); err != nil {
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
		if i >= 5 {
			break
		}
		if err := v.GenerateLocomotiveMap(act.key); err != nil {
			fmt.Printf("Ошибка для %s: %v\n", act.key, err)
		}
	}

	fmt.Printf("✅ Все карты для депо %s сгенерированы\n", depoID)
	return nil
}

// filterByDepo фильтрует локомотивы по депо
func (v *visualizationService) filterByDepo(locomotives map[string]domain.Locomotive, depoID string) map[string]domain.Locomotive {
	result := make(map[string]domain.Locomotive)
	for key, loc := range locomotives {
		if loc.Depo == depoID {
			result[key] = loc
		}
	}
	return result
}

// getStationCoordinates получает координаты станций
// getStationCoordinates получает координаты станций из файла
// getStationCoordinates получает координаты станций из station_info.csv
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
			"/Users/polzovatel/Desktop/PYTHON/Hackaton/services/task3/data/station_info.csv",
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

// collectStationStats собирает статистику посещений
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

// generateHTMLMap создает HTML файл с картой
func (v *visualizationService) generateHTMLMap(
	depoID string,
	stationStats map[string]*domain.StationStats,
	routes map[string][]domain.LocomotiveRoute,
	topLocomotives []string,
	stations map[string]domain.Station) error {

	// Создаем директорию для карт
	if err := os.MkdirAll("maps", 0755); err != nil {
		return err
	}

	// Подготавливаем данные для JavaScript
	var jsRoutes []JSRoute
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7", "#C7B198", "#DFC2C2", "#B2B2B2"}

	for i, locKey := range topLocomotives {
		if routeList, exists := routes[locKey]; exists {
			color := colors[i%len(colors)]
			for _, route := range routeList {
				var points [][]float64
				for _, p := range route.Points {
					points = append(points, []float64{p.Lon, p.Lat})
				}
				jsRoutes = append(jsRoutes, JSRoute{
					Points:     points,
					Color:      color,
					Locomotive: locKey,
				})
			}
		}
	}

	// Подготавливаем станции
	var jsStations []JSStation
	for _, stat := range stationStats {
		if stat.Latitude != 0 && stat.Longitude != 0 {
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

	// Конвертируем в JSON
	stationsJSON, _ := json.Marshal(jsStations)
	routesJSON, _ := json.Marshal(jsRoutes)
	centerLat, centerLon := calculateCenter(stations, 55.75, 37.62)

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
            width: 250px;
        }
        .legend {
            margin-top: 10px;
            padding: 10px 0;
        }
        .legend-item {
            display: flex;
            align-items: center;
            margin: 5px 0;
        }
        .color-box {
            width: 20px;
            height: 20px;
            margin-right: 8px;
            border-radius: 4px;
        }
        .station-info {
            position: absolute;
            bottom: 30px;
            left: 10px;
            background: white;
            padding: 10px;
            border-radius: 4px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.2);
            z-index: 1000;
            font-size: 12px;
            max-width: 300px;
        }
        .controls {
            position: absolute;
            top: 10px;
            left: 10px;
            background: white;
            padding: 10px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
            z-index: 1000;
        }
        button {
            margin: 2px;
            padding: 5px 10px;
            cursor: pointer;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    
    <div class="info-panel">
        <h3>Депо %s</h3>
        <p>Станций: %d<br>
           Маршрутов: %d<br>
           Топ-%d локомотивов</p>
        
        <div class="legend">
            <h4>Цвета маршрутов:</h4>`, depoID, depoID, len(jsStations), len(jsRoutes), len(topLocomotives))

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
    </div>

    <div class="controls">
        <button onclick="toggleHeatmap()">Тепловая карта</button>
        <button onclick="toggleRoutes()">Маршруты</button>
        <button onclick="toggleStations()">Станции</button>
        <button onclick="resetView()">Сброс вида</button>
    </div>

    <div id="stationInfo" class="station-info" style="display: none;"></div>

    <script>
        // Инициализация карты
        var map = L.map('map').setView([%f, %f], 11);
        
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '© OpenStreetMap contributors'
        }).addTo(map);

        // Слои
        var stationLayer = L.layerGroup();
        var routeLayer = L.layerGroup();
        var heatLayer = null;

        // Добавляем станции
        var stations = %s;
        stations.forEach(function(s) {
            var marker = L.circleMarker(s.coords, {
                radius: s.size,
                color: s.color,
                fillColor: s.color,
                fillOpacity: 0.8,
                weight: 1
            }).bindPopup('<b>' + s.id + '</b><br>' + s.name + '<br>Посещений: ' + s.visits);
            
            marker.on('mouseover', function() {
                document.getElementById('stationInfo').style.display = 'block';
                document.getElementById('stationInfo').innerHTML = '<b>' + s.id + '</b><br>' + s.name + '<br>Посещений: ' + s.visits;
            });
            
            marker.on('mouseout', function() {
                document.getElementById('stationInfo').style.display = 'none';
            });
            
            stationLayer.addLayer(marker);
        });
        stationLayer.addTo(map);

        // Добавляем маршруты
        var routes = %s;
        routes.forEach(function(r) {
            var polyline = L.polyline(r.points, {
                color: r.color,
                weight: 3,
                opacity: 0.7
            }).bindPopup('Локомотив: ' + r.locomotive);
            routeLayer.addLayer(polyline);
        });
        routeLayer.addTo(map);

        // Тепловая карта
        var heatData = stations.map(function(s) {
            return [s.coords[1], s.coords[0], s.visits];
        });
        heatLayer = L.heatLayer(heatData, {
            radius: 25,
            blur: 15,
            maxZoom: 10,
            gradient: {0.4: 'blue', 0.6: 'lime', 0.8: 'red'}
        });

        // Управление слоями
        function toggleHeatmap() {
            if (map.hasLayer(heatLayer)) {
                map.removeLayer(heatLayer);
            } else {
                heatLayer.addTo(map);
            }
        }

        function toggleRoutes() {
            if (map.hasLayer(routeLayer)) {
                map.removeLayer(routeLayer);
            } else {
                routeLayer.addTo(map);
            }
        }

        function toggleStations() {
            if (map.hasLayer(stationLayer)) {
                map.removeLayer(stationLayer);
            } else {
                stationLayer.addTo(map);
            }
        }

        function resetView() {
            map.setView([%f, %f], 11);
        }

        // Добавляем масштаб
        L.control.scale().addTo(map);
    </script>
</body>
</html>`, centerLat, centerLon, stationsJSON, routesJSON, centerLat, centerLon)

	// Сохраняем файл
	filename := fmt.Sprintf("maps/depot_%s_map.html", depoID)
	return os.WriteFile(filename, []byte(html), 0644)
}

// generateHeatmapHTML создает тепловую карту
func (v *visualizationService) generateHeatmapHTML(
	depoID string,
	stationStats map[string]*domain.StationStats) error {

	filename := fmt.Sprintf("maps/depot_%s_heatmap.html", depoID)

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
	centerLat, centerLon := calculateCenter(nil, 55.75, 37.62)

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
    </style>
</head>
<body>
    <div id="map"></div>
    <script>
        var map = L.map('map').setView([%f, %f], 11);
        
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png').addTo(map);

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

// generateLocomotiveHTML создает карту для конкретного локомотива
func (v *visualizationService) generateLocomotiveHTML(
	locomotiveKey string,
	loc domain.Locomotive,
	stations map[string]domain.Station) error {

	filename := fmt.Sprintf("maps/locomotive_%s.html", strings.ReplaceAll(locomotiveKey, "-", "_"))

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
	centerLat, centerLon := calculateCenter(stations, 55.75, 37.62)
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
        
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png').addTo(map);

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
