package services

import (
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/mihnpro/Hackathon_TMX/internal/domain"
    "github.com/mihnpro/Hackathon_TMX/internal/transport/models/responses"
)

type mostPopularTripService struct {
    dataPath     string
    stationsPath string
    stations     domain.StationMap
}

type MostPopularTripService interface {
    RunMostPopularTrip()
    GetPopularDirections() (*responses.Task2Response, error)
    GetLocomotivePopularDirection(series, number string) (*responses.LocomotiveStats, error)
}

func NewMostPopularTripService(dataPath, stationsPath string) MostPopularTripService {
    svc := &mostPopularTripService{
        dataPath:     dataPath,
        stationsPath: stationsPath,
    }
    svc.loadStations()
    return svc
}

// loadStations - загружает информацию о станциях
func (m *mostPopularTripService) loadStations() {
    m.stations = make(domain.StationMap)
    
    file, err := os.Open(m.stationsPath)
    if err != nil {
        fmt.Printf("Предупреждение: не удалось загрузить файл станций: %v\n", err)
        return
    }
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = ','
    reader.FieldsPerRecord = -1

    // Пропускаем заголовок
    _, err = reader.Read()
    if err != nil {
        fmt.Printf("Ошибка чтения заголовка: %v\n", err)
        return
    }

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Printf("Ошибка чтения записи: %v\n", err)
            continue
        }

        if len(record) < 2 {
            continue
        }

        code := strings.TrimSpace(record[0])
        name := strings.TrimSpace(record[1])
        
        var lat, lon float64
        if len(record) >= 3 && record[2] != "" {
            lat, _ = strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
        }
        if len(record) >= 4 && record[3] != "" {
            lon, _ = strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
        }

        m.stations[code] = domain.StationInfo{
            Code:      code,
            Name:      name,
            Latitude:  lat,
            Longitude: lon,
        }
    }
    
    fmt.Printf("Загружено станций: %d\n", len(m.stations))
}

// getStationName - получает название станции
func (m *mostPopularTripService) getStationName(code string) string {
    if station, ok := m.stations[code]; ok {
        return station.Name
    }
    return code
}

// loadData - загружает данные о локомотивах
func (m *mostPopularTripService) loadData() map[string]domain.Locomotive {
    locomotives := make(map[string]domain.Locomotive)
    
    file, err := os.Open(m.dataPath)
    if err != nil {
        fmt.Printf("Ошибка открытия файла: %v\n", err)
        return locomotives
    }
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = ','
    
    // Пропускаем заголовок
    _, err = reader.Read()
    if err != nil {
        fmt.Printf("Ошибка чтения заголовка: %v\n", err)
        return locomotives
    }

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Printf("Ошибка чтения записи: %v\n", err)
            continue
        }

        if len(record) < 5 {
            continue
        }

        series := strings.TrimSpace(record[0])
        number := strings.TrimSpace(record[1])
        datetimeStr := strings.TrimSpace(record[2])
        station := strings.TrimSpace(record[3])
        depo := strings.TrimSpace(record[4])

        timestamp, err := time.Parse("2006-01-02T15:04:05.000000", datetimeStr)
        if err != nil {
            timestamp, err = time.Parse("2006-01-02T15:04:05", datetimeStr)
            if err != nil {
                fmt.Printf("Ошибка парсинга даты: %v\n", err)
                continue
            }
        }

        key := series + "_" + number
        loc, exists := locomotives[key]
        if !exists {
            loc = domain.Locomotive{
                Series:  series,
                Number:  number,
                Depo:    depo,
                Records: []domain.Record{},
            }
        }

        loc.Records = append(loc.Records, domain.Record{
            Series:    series,
            Number:    number,
            Timestamp: timestamp,
            Station:   station,
            Depo:      depo,
        })

        locomotives[key] = loc
    }

    // Сортируем записи по времени
    for key, loc := range locomotives {
        sort.Slice(loc.Records, func(i, j int) bool {
            return loc.Records[i].Timestamp.Before(loc.Records[j].Timestamp)
        })
        locomotives[key] = loc
    }

    return locomotives
}

// splitIntoTrips - разбивает записи на поездки
func (m *mostPopularTripService) splitIntoTrips(records []domain.Record) []domain.Trip {
    var trips []domain.Trip
    if len(records) == 0 {
        return trips
    }

    var currentTrip domain.Trip
    currentTrip.StartTime = records[0].Timestamp
    currentTrip.Stations = []string{records[0].Station}

    for i := 1; i < len(records); i++ {
        current := records[i]
        prev := records[i-1]

        // Если вернулись в депо - завершаем поездку
        if current.Station == current.Depo && prev.Station != current.Depo {
            currentTrip.Stations = append(currentTrip.Stations, current.Station)
            currentTrip.EndTime = current.Timestamp
            
            // Очищаем маршрут от повторов
            currentTrip.Route = m.cleanStops(currentTrip.Stations)
            
            // Добавляем названия станций
            currentTrip.StationNames = make([]string, len(currentTrip.Stations))
            for j, station := range currentTrip.Stations {
                currentTrip.StationNames[j] = m.getStationName(station)
            }
            
            trips = append(trips, currentTrip)
            
            // Начинаем новую поездку, если есть следующие записи
            if i+1 < len(records) {
                currentTrip = domain.Trip{
                    StartTime: records[i+1].Timestamp,
                    Stations:  []string{records[i+1].Station},
                }
                i++ // пропускаем следующую запись, так как уже использовали
            }
        } else {
            currentTrip.Stations = append(currentTrip.Stations, current.Station)
        }
    }

    return trips
}

// cleanStops - удаляет повторяющиеся станции (стоянки)
func (m *mostPopularTripService) cleanStops(stations []string) []string {
    if len(stations) == 0 {
        return stations
    }

    result := []string{stations[0]}
    for i := 1; i < len(stations); i++ {
        if stations[i] != stations[i-1] {
            result = append(result, stations[i])
        }
    }
    return result
}

// identifyDirectionsFromTrips - определяет направления на основе маршрутов поездок
func (m *mostPopularTripService) identifyDirectionsFromTrips(locomotives map[string]domain.Locomotive) map[string][]domain.Direction {
    // Собираем все уникальные маршруты для каждого депо
    depotRoutes := make(map[string]map[string]*domain.Direction)
    
    for _, loc := range locomotives {
        for _, trip := range loc.Trips {
            if len(trip.Route) < 2 {
                continue // пропускаем поездки без маршрута
            }
            
            // Определяем ключ маршрута: от депо до конечной станции
            start := trip.Route[0] // должна быть станция депо
            end := trip.Route[len(trip.Route)-1] // конечная станция
            
            // Проверяем, что начали с депо
            if start != loc.Depo {
                continue
            }
            
            routeKey := fmt.Sprintf("%s->%s", start, end)
            
            if _, exists := depotRoutes[loc.Depo]; !exists {
                depotRoutes[loc.Depo] = make(map[string]*domain.Direction)
            }
            
            if dir, exists := depotRoutes[loc.Depo][routeKey]; exists {
                // Увеличиваем частоту использования
                dir.Frequency++
                if dir.Locomotives == nil {
                    dir.Locomotives = make(map[string]bool)
                }
                dir.Locomotives[loc.Series+"_"+loc.Number] = true
            } else {
                // Создаем новое направление
                routeNames := make([]string, len(trip.Route))
                for i, station := range trip.Route {
                    routeNames[i] = m.getStationName(station)
                }
                
                terminalName := m.getStationName(end)
                
                // Формируем название направления
                directionName := fmt.Sprintf("Маршрут на %s", terminalName)
                if len(routeNames) > 1 {
                    // Если есть промежуточные станции, добавляем их
                    intermediate := routeNames[1 : len(routeNames)-1]
                    if len(intermediate) > 0 {
                        directionName = fmt.Sprintf("Через %s на %s", 
                            strings.Join(intermediate, " → "), 
                            terminalName)
                    }
                }
                
                depotRoutes[loc.Depo][routeKey] = &domain.Direction{
                    ID:          fmt.Sprintf("dir_%s_%s", loc.Depo, end),
                    Name:        directionName,
                    Depo:        loc.Depo,
                    Terminal:    end,
                    TerminalName: terminalName,
                    Route:       trip.Route,
                    RouteNames:  routeNames,
                    Frequency:   1,
                    Locomotives: map[string]bool{loc.Series + "_" + loc.Number: true},
                }
            }
        }
    }
    
    // Преобразуем в нужный формат и сортируем по популярности
    result := make(map[string][]domain.Direction)
    for depo, routes := range depotRoutes {
        directions := make([]domain.Direction, 0, len(routes))
        for _, dir := range routes {
            directions = append(directions, *dir)
        }
        
        // Сортируем по частоте использования
        sort.Slice(directions, func(i, j int) bool {
            return directions[i].Frequency > directions[j].Frequency
        })
        
        result[depo] = directions
    }
    
    return result
}

// analyzeFavoriteDirections - анализ популярных направлений для каждого локомотива
func (m *mostPopularTripService) analyzeFavoriteDirections(
    locomotives map[string]domain.Locomotive,
    depotDirections map[string][]domain.Direction) map[string]domain.LocomotiveDirectionStats {

    stats := make(map[string]domain.LocomotiveDirectionStats)

    for key, loc := range locomotives {
        if len(loc.Trips) == 0 {
            continue
        }

        // Получаем все возможные направления для этого депо
        directions, exists := depotDirections[loc.Depo]
        if !exists {
            continue
        }

        // Создаем карту направлений для быстрого доступа
        dirMap := make(map[string]domain.Direction)
        for _, dir := range directions {
            dirMap[dir.ID] = dir
        }

        locStats := domain.LocomotiveDirectionStats{
            LocomotiveKey:   key,
            Model:           loc.Series,
            Number:          loc.Number,
            Depo:            loc.Depo,
            DepoName:        m.getStationName(loc.Depo),
            TotalTrips:      len(loc.Trips),
            DirectionVisits: make(map[string]int),
            Directions:      make([]domain.DirectionInfo, 0),
        }

        // Анализируем каждую поездку
        for _, trip := range loc.Trips {
            if len(trip.Route) < 2 {
                continue
            }

            // Определяем, какому направлению соответствует эта поездка
            matchedDir := m.matchTripToDirection(trip, directions)
            if matchedDir != "" {
                locStats.DirectionVisits[matchedDir]++
            }
        }

        // Находим самое популярное направление
        maxVisits := 0
        mostPopular := ""
        for dirID, visits := range locStats.DirectionVisits {
            if visits > maxVisits {
                maxVisits = visits
                mostPopular = dirID
            }
        }

        locStats.MostPopularDirection = mostPopular
        locStats.MaxVisits = maxVisits
        
        if mostPopular != "" {
            if dir, ok := dirMap[mostPopular]; ok {
                locStats.MostPopularName = dir.Name
            }
        }

        // Формируем детальную информацию о посещенных направлениях
        for dirID, visits := range locStats.DirectionVisits {
            if dir, ok := dirMap[dirID]; ok {
                info := domain.DirectionInfo{
                    ID:          dirID,
                    Name:        dir.Name,
                    Terminal:    dir.Terminal,
                    TerminalName: dir.TerminalName,
                    Visits:      visits,
                    Percentage:  float64(visits) / float64(locStats.TotalTrips) * 100,
                }
                locStats.Directions = append(locStats.Directions, info)
            }
        }

        // Сортируем направления по популярности
        sort.Slice(locStats.Directions, func(i, j int) bool {
            return locStats.Directions[i].Visits > locStats.Directions[j].Visits
        })

        stats[key] = locStats
    }

    return stats
}

// matchTripToDirection - определяет, какому направлению соответствует поездка
func (m *mostPopularTripService) matchTripToDirection(trip domain.Trip, directions []domain.Direction) string {
    if len(trip.Route) < 2 {
        return ""
    }

    // Конечная станция поездки
    tripEnd := trip.Route[len(trip.Route)-1]
    
    // Ищем направление с такой же конечной станцией
    for _, dir := range directions {
        if dir.Terminal == tripEnd {
            return dir.ID
        }
    }
    
    // Если не нашли точного совпадения, ищем частичное совпадение маршрута
    bestMatch := ""
    bestScore := 0
    
    for _, dir := range directions {
        score := m.calculateRouteSimilarity(trip.Route, dir.Route)
        if score > bestScore && score > 50 { // минимум 50% совпадения
            bestScore = score
            bestMatch = dir.ID
        }
    }
    
    return bestMatch
}

// calculateRouteSimilarity - вычисляет процент совпадения маршрутов
func (m *mostPopularTripService) calculateRouteSimilarity(route1, route2 []string) int {
    if len(route1) == 0 || len(route2) == 0 {
        return 0
    }
    
    // Создаем множества станций
    set1 := make(map[string]bool)
    for _, s := range route1 {
        set1[s] = true
    }
    
    set2 := make(map[string]bool)
    for _, s := range route2 {
        set2[s] = true
    }
    
    // Считаем пересечение
    intersection := 0
    for s := range set1 {
        if set2[s] {
            intersection++
        }
    }
    
    // Считаем объединение
    union := len(set1) + len(set2) - intersection
    if union == 0 {
        return 0
    }
    
    return intersection * 100 / union
}

// printDirectionStats - выводит статистику
func (m *mostPopularTripService) printDirectionStats(
    stats map[string]domain.LocomotiveDirectionStats,
    depotDirections map[string][]domain.Direction) {

    // Группируем по депо
    byDepot := make(map[string][]domain.LocomotiveDirectionStats)
    for _, stat := range stats {
        byDepot[stat.Depo] = append(byDepot[stat.Depo], stat)
    }

    depots := make([]string, 0, len(byDepot))
    for d := range byDepot {
        depots = append(depots, d)
    }
    sort.Strings(depots)

    for _, depo := range depots {
        locStats := byDepot[depo]
        depoName := m.getStationName(depo)
        
        fmt.Printf("\n%s\n", strings.Repeat("=", 80))
        fmt.Printf("ДЕПО: %s (код: %s)\n", depoName, depo)
        fmt.Printf("%s\n", strings.Repeat("=", 80))
        
        fmt.Printf("\n📊 ОБЩАЯ ИНФОРМАЦИЯ:\n")
        fmt.Printf("  • Локомотивов в депо: %d\n", len(locStats))
        
        // Показываем популярные направления из этого депо
        if dirs, exists := depotDirections[depo]; exists && len(dirs) > 0 {
            fmt.Printf("\n🚂 ПОПУЛЯРНЫЕ НАПРАВЛЕНИЯ ИЗ ДЕПО:\n")
            for i, dir := range dirs {
                if i >= 5 {
                    fmt.Printf("  • ... и еще %d направлений\n", len(dirs)-5)
                    break
                }
                fmt.Printf("  %d. %s\n", i+1, dir.Name)
                fmt.Printf("     Маршрут: %s\n", strings.Join(dir.RouteNames, " → "))
                fmt.Printf("     Используют: %d локомотивов, %d поездок\n", 
                    len(dir.Locomotives), dir.Frequency)
            }
        }
        
        fmt.Printf("\n📈 АНАЛИЗ ЛОКОМОТИВОВ:\n")
        
        // Сортируем локомотивы по модели и номеру
        sort.Slice(locStats, func(i, j int) bool {
            if locStats[i].Model == locStats[j].Model {
                return locStats[i].Number < locStats[j].Number
            }
            return locStats[i].Model < locStats[j].Model
        })

        // Показываем первые 10 локомотивов
        displayCount := 10
        if len(locStats) < displayCount {
            displayCount = len(locStats)
        }

        for i := 0; i < displayCount; i++ {
            stat := locStats[i]
            fmt.Printf("\n  🔹 Локомотив %s-%s:\n", stat.Model, stat.Number)
            fmt.Printf("     Всего поездок: %d\n", stat.TotalTrips)
            
            if len(stat.Directions) > 0 {
                fmt.Printf("     Посещенные направления:\n")
                for j, dir := range stat.Directions {
                    if j >= 3 {
                        fmt.Printf("       • ... и еще %d направлений\n", len(stat.Directions)-3)
                        break
                    }
                    fmt.Printf("       %d. %s\n", j+1, dir.Name)
                    fmt.Printf("          Конечная: %s\n", dir.TerminalName)
                    fmt.Printf("          Поездок: %d (%.1f%%)\n", dir.Visits, dir.Percentage)
                }
            }
            
            if stat.MostPopularDirection != "" {
                fmt.Printf("\n     ⭐ САМОЕ ПОПУЛЯРНОЕ НАПРАВЛЕНИЕ:\n")
                fmt.Printf("        %s\n", stat.MostPopularName)
                fmt.Printf("        Поездок: %d из %d (%.1f%%)\n", 
                    stat.MaxVisits, stat.TotalTrips,
                    float64(stat.MaxVisits)/float64(stat.TotalTrips)*100)
            }
        }

        if len(locStats) > displayCount {
            fmt.Printf("\n  ... и еще %d локомотивов\n", len(locStats)-displayCount)
        }
    }

    m.printOverallStats(stats)
}

// printOverallStats - общая статистика
func (m *mostPopularTripService) printOverallStats(stats map[string]domain.LocomotiveDirectionStats) {
    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("ОБЩАЯ СТАТИСТИКА")
    fmt.Println(strings.Repeat("=", 80))

    totalLocomotives := len(stats)
    locWithFavorite := 0
    locWithSingleDirection := 0
    totalTrips := 0

    for _, stat := range stats {
        totalTrips += stat.TotalTrips
        if stat.MostPopularDirection != "" {
            locWithFavorite++
        }
        if len(stat.Directions) == 1 {
            locWithSingleDirection++
        }
    }

    fmt.Printf("\n📊 ОБЩИЕ ПОКАЗАТЕЛИ:\n")
    fmt.Printf("  • Всего локомотивов: %d\n", totalLocomotives)
    fmt.Printf("  • Всего поездок: %d\n", totalTrips)
    fmt.Printf("  • Среднее число поездок на локомотив: %.1f\n", 
        float64(totalTrips)/float64(totalLocomotives))
    
    fmt.Printf("\n📈 СТАТИСТИКА ПО НАПРАВЛЕНИЯМ:\n")
    fmt.Printf("  • Локомотивов с любимым направлением: %d (%.1f%%)\n",
        locWithFavorite, float64(locWithFavorite)/float64(totalLocomotives)*100)
    fmt.Printf("  • Локомотивов, работающих на одном направлении: %d (%.1f%%)\n",
        locWithSingleDirection, 
        float64(locWithSingleDirection)/float64(totalLocomotives)*100)
}

// RunMostPopularTrip - основной метод для консольного режима
func (m *mostPopularTripService) RunMostPopularTrip() {
    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("ЗАГРУЗКА ДАННЫХ")
    fmt.Println(strings.Repeat("=", 80))
    
    locomotives := m.loadData()
    fmt.Printf("✓ Загружено локомотивов: %d\n", len(locomotives))

    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("РАЗБИЕНИЕ НА ПОЕЗДКИ")
    fmt.Println(strings.Repeat("=", 80))
    
    totalTrips := 0
    for key, loc := range locomotives {
        loc.Trips = m.splitIntoTrips(loc.Records)
        totalTrips += len(loc.Trips)
        locomotives[key] = loc
    }
    fmt.Printf("✓ Выделено поездок: %d\n", totalTrips)

    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("ОПРЕДЕЛЕНИЕ НАПРАВЛЕНИЙ")
    fmt.Println(strings.Repeat("=", 80))
    
    depotDirections := m.identifyDirectionsFromTrips(locomotives)
    
    totalDirections := 0
    for _, dirs := range depotDirections {
        totalDirections += len(dirs)
    }
    fmt.Printf("✓ Определено направлений: %d\n", totalDirections)

    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("АНАЛИЗ ПОПУЛЯРНЫХ НАПРАВЛЕНИЙ")
    fmt.Println(strings.Repeat("=", 80))
    
    locomotiveStats := m.analyzeFavoriteDirections(locomotives, depotDirections)
    
    m.printDirectionStats(locomotiveStats, depotDirections)
}

// GetPopularDirections - для API режима
func (m *mostPopularTripService) GetPopularDirections() (*responses.Task2Response, error) {
    locomotives := m.loadData()

    for key, loc := range locomotives {
        loc.Trips = m.splitIntoTrips(loc.Records)
        locomotives[key] = loc
    }

    depotDirections := m.identifyDirectionsFromTrips(locomotives)
    locomotiveStats := m.analyzeFavoriteDirections(locomotives, depotDirections)

    return m.buildTask2Response(locomotiveStats, depotDirections), nil
}

// GetLocomotivePopularDirection - для API режима
func (m *mostPopularTripService) GetLocomotivePopularDirection(series, number string) (*responses.LocomotiveStats, error) {
    locomotives := m.loadData()

    for key, loc := range locomotives {
        loc.Trips = m.splitIntoTrips(loc.Records)
        locomotives[key] = loc
    }

    depotDirections := m.identifyDirectionsFromTrips(locomotives)
    locomotiveStats := m.analyzeFavoriteDirections(locomotives, depotDirections)

    key := series + "_" + number
    stats, exists := locomotiveStats[key]
    if !exists {
        return nil, fmt.Errorf("локомотив %s-%s не найден", series, number)
    }

    return m.buildLocomotiveStatsResponse(stats, depotDirections[stats.Depo]), nil
}

// buildTask2Response - формирует ответ для API
func (m *mostPopularTripService) buildTask2Response(
    stats map[string]domain.LocomotiveDirectionStats,
    depotDirections map[string][]domain.Direction) *responses.Task2Response {

    response := &responses.Task2Response{
        Depots:      make([]responses.DepotResponse, 0),
        OverallStats: responses.OverallStats{},
    }

    // Группируем по депо
    byDepot := make(map[string][]domain.LocomotiveDirectionStats)
    for _, stat := range stats {
        byDepot[stat.Depo] = append(byDepot[stat.Depo], stat)
    }

    depots := make([]string, 0, len(byDepot))
    for d := range byDepot {
        depots = append(depots, d)
    }
    sort.Strings(depots)

    for _, depo := range depots {
        locStats := byDepot[depo]
        
        depotResponse := responses.DepotResponse{
            DepoCode:        depo,
            DepoName:        m.getStationName(depo),
            LocomotiveCount: len(locStats),
            Directions:      make([]responses.DirectionInfo, 0),
            Locomotives:     make([]responses.LocomotiveStats, 0),
        }

        // Добавляем направления
        if dirs, exists := depotDirections[depo]; exists {
            for _, d := range dirs {
                depotResponse.Directions = append(depotResponse.Directions, responses.DirectionInfo{
                    ID:          d.ID,
                    Name:        d.Name,
                    Terminal:    d.Terminal,
                    TerminalName: d.TerminalName,
                    Frequency:   d.Frequency,
                    LocomotiveCount: len(d.Locomotives),
                })
            }
        }

        // Добавляем локомотивы
        for _, stat := range locStats {
            locStatsResp := m.buildLocomotiveStatsResponse(stat, depotDirections[depo])
            depotResponse.Locomotives = append(depotResponse.Locomotives, *locStatsResp)
        }

        response.Depots = append(response.Depots, depotResponse)
    }

    // Общая статистика
    totalLocomotives := len(stats)
    locWithFavorite := 0
    locWithSingleDirection := 0
    totalTrips := 0

    for _, stat := range stats {
        totalTrips += stat.TotalTrips
        if stat.MostPopularDirection != "" {
            locWithFavorite++
        }
        if len(stat.Directions) == 1 {
            locWithSingleDirection++
        }
    }

    response.OverallStats = responses.OverallStats{
        TotalLocomotives:      totalLocomotives,
        TotalTrips:            totalTrips,
        AvgTripsPerLocomotive: float64(totalTrips) / float64(totalLocomotives),
        LocomotivesWithFavorite: locWithFavorite,
        LocomotivesWithFavoritePercent: float64(locWithFavorite) / float64(totalLocomotives) * 100,
        LocomotivesSingleDirection: locWithSingleDirection,
        LocomotivesSingleDirectionPercent: float64(locWithSingleDirection) / float64(totalLocomotives) * 100,
    }

    return response
}

// buildLocomotiveStatsResponse - формирует ответ для конкретного локомотива
func (m *mostPopularTripService) buildLocomotiveStatsResponse(
    stat domain.LocomotiveDirectionStats,
    directions []domain.Direction) *responses.LocomotiveStats {

    locStatsResp := &responses.LocomotiveStats{
        Model:        stat.Model,
        Number:       stat.Number,
        Depo:         stat.Depo,
        DepoName:     stat.DepoName,
        TotalTrips:   stat.TotalTrips,
        Directions:   make([]responses.LocomotiveDirection, 0),
    }

    // Добавляем информацию о посещенных направлениях
    for _, dir := range stat.Directions {
        locStatsResp.Directions = append(locStatsResp.Directions, responses.LocomotiveDirection{
            ID:          dir.ID,
            Name:        dir.Name,
            Terminal:    dir.Terminal,
            TerminalName: dir.TerminalName,
            Visits:      dir.Visits,
            Percentage:  dir.Percentage,
        })
    }

    // Добавляем самое популярное направление
    if stat.MostPopularDirection != "" {
        locStatsResp.MostPopular = &responses.MostPopularDirection{
            DirectionID:   stat.MostPopularDirection,
            DirectionName: stat.MostPopularName,
            Visits:        stat.MaxVisits,
            Percentage:    float64(stat.MaxVisits) / float64(stat.TotalTrips) * 100,
        }
    }

    return locStatsResp
}