package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mihnpro/Hackathon_TMX/internal/domain"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/models/responses"
)

type mostPopularTripService struct {
	dataPath string
}

type MostPopularTripService interface {
	// Для консольного режима (как сейчас)
	RunMostPopularTrip()
	
	// Для API режима (возвращают данные)
	GetPopularDirections() (*responses.Task2Response, error)
	GetLocomotivePopularDirection(series, number string) (*responses.LocomotiveStats, error)
}

func NewMostPopularTripService(dataPath string) MostPopularTripService {
	return &mostPopularTripService{dataPath: dataPath}
}

// RunMostPopularTrip - для консольного режима (как было)
func (m *mostPopularTripService) RunMostPopularTrip() {
	// 1. Загружаем данные
	fmt.Println("Загрузка данных...")
	locomotives := loadData(m.dataPath)
	fmt.Printf("Загружено локомотивов: %d\n\n", len(locomotives))

	// 2. Разбиваем на поездки
	fmt.Println("Разбиение на поездки...")
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Определяем направления для каждого депо
	fmt.Println("Определение направлений...")
	depotDirections := identifyDirections(locomotives)

	// 4. ПУНКТ 2: Анализ популярных направлений
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ПУНКТ 2: САМЫЕ ПОПУЛЯРНЫЕ НАПРАВЛЕНИЯ ЛОКОМОТИВОВ")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	locomotiveStats := analyzeFavoriteDirections(locomotives, depotDirections)

	// 5. Выводим результаты
	printDirectionStats(locomotiveStats, depotDirections)

	// 6. Дополнительный анализ по депо
	printDepotAnalysis(locomotiveStats, depotDirections)
}

// GetPopularDirections - для API режима
func (m *mostPopularTripService) GetPopularDirections() (*responses.Task2Response, error) {
	// 1. Загружаем данные
	locomotives := loadData(m.dataPath)

	// 2. Разбиваем на поездки
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Определяем направления для каждого депо
	depotDirections := identifyDirections(locomotives)

	// 4. Анализ популярных направлений
	locomotiveStats := analyzeFavoriteDirections(locomotives, depotDirections)

	// 5. Формируем ответ
	return buildTask2Response(locomotiveStats, depotDirections), nil
}

// GetLocomotivePopularDirection - для API режима (конкретный локомотив)
func (m *mostPopularTripService) GetLocomotivePopularDirection(series, number string) (*responses.LocomotiveStats, error) {
	// 1. Загружаем данные
	locomotives := loadData(m.dataPath)

	// 2. Разбиваем на поездки
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Определяем направления для каждого депо
	depotDirections := identifyDirections(locomotives)

	// 4. Анализ популярных направлений
	locomotiveStats := analyzeFavoriteDirections(locomotives, depotDirections)

	// 5. Ищем нужный локомотив
	key := series + "_" + number
	stats, exists := locomotiveStats[key]
	if !exists {
		return nil, nil // not found
	}

	// 6. Формируем ответ для конкретного локомотива
	return buildLocomotiveStatsResponse(stats, depotDirections[stats.Depo]), nil
}

// identifyDirections - определяет направления для каждого депо
func identifyDirections(locomotives map[string]domain.Locomotive) map[string][]domain.Direction {
	// Собираем все станции, которые посещали локомотивы каждого депо
	depotStations := make(map[string]map[string]bool)

	for _, loc := range locomotives {
		if _, exists := depotStations[loc.Depo]; !exists {
			depotStations[loc.Depo] = make(map[string]bool)
		}

		for _, trip := range loc.Trips {
			for _, station := range trip.Stations {
				if station != loc.Depo {
					depotStations[loc.Depo][station] = true
				}
			}
		}
	}

	// Группируем станции по префиксам (направлениям)
	depotDirections := make(map[string][]domain.Direction)

	// Словарь названий направлений
	directionNames := map[string]string{
		"94": "Западное",
		"24": "Восточное",
		"30": "Центральное",
		"31": "Южное",
		"50": "Северное",
		"51": "Северо-восточное",
		"25": "Юго-восточное",
		"58": "Направление 58",
		"59": "Направление 59",
	}

	for depo, stations := range depotStations {
		// Группируем по первым 2 цифрам
		prefixGroups := make(map[string][]string)

		for station := range stations {
			if len(station) >= 2 {
				prefix := station[:2]
				prefixGroups[prefix] = append(prefixGroups[prefix], station)
			}
		}

		// Создаем направления
		var directions []domain.Direction
		for prefix, stationList := range prefixGroups {
			// Сортируем станции
			sort.Strings(stationList)

			// Определяем название
			name := directionNames[prefix]
			if name == "" {
				name = "Направление " + prefix
			}

			direction := domain.Direction{
				ID:       depo + "_dir_" + prefix,
				Name:     name,
				Prefix:   prefix,
				Stations: stationList,
				Depo:     depo,
			}

			directions = append(directions, direction)
		}

		// Сортируем направления по популярности (количеству станций)
		sort.Slice(directions, func(i, j int) bool {
			return len(directions[i].Stations) > len(directions[j].Stations)
		})

		depotDirections[depo] = directions
	}

	return depotDirections
}

// analyzeFavoriteDirections - анализ популярных направлений для локомотивов
func analyzeFavoriteDirections(locomotives map[string]domain.Locomotive,
	depotDirections map[string][]domain.Direction) map[string]domain.LocomotiveDirectionStats {

	stats := make(map[string]domain.LocomotiveDirectionStats)

	for key, loc := range locomotives {
		if len(loc.Trips) == 0 {
			continue
		}

		// Получаем направления для депо
		directions, exists := depotDirections[loc.Depo]
		if !exists {
			continue
		}

		// Инициализируем статистику
		locStats := domain.LocomotiveDirectionStats{
			LocomotiveKey:   key,
			Model:           loc.Series,
			Number:          loc.Number,
			Depo:            loc.Depo,
			TotalTrips:      len(loc.Trips),
			DirectionVisits: make(map[string]int),
		}

		// Анализируем каждую поездку
		for _, trip := range loc.Trips {
			if len(trip.Stations) < 2 {
				continue
			}

			// Очищаем путь от стоянок
			cleanPath := cleanStops(trip.Stations)

			// Множество направлений в этой поездке
			visitedInTrip := make(map[string]bool)

			// Проверяем каждое направление
			for _, dir := range directions {
				if hasDirectionIntersection(cleanPath, dir.Stations) {
					visitedInTrip[dir.ID] = true
				}
			}

			// Увеличиваем счетчики
			for dirID := range visitedInTrip {
				locStats.DirectionVisits[dirID]++
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

		// Находим название направления
		if mostPopular != "" {
			for _, dir := range directions {
				if dir.ID == mostPopular {
					locStats.PopularDirectionName = dir.Name
					break
				}
			}
		}

		// Сохраняем все посещенные направления
		for dirID := range locStats.DirectionVisits {
			locStats.VisitedDirections = append(locStats.VisitedDirections, dirID)
		}

		stats[key] = locStats
	}

	return stats
}

// hasDirectionIntersection - проверяет пересечение с направлением
func hasDirectionIntersection(path []string, directionStations []string) bool {
	dirSet := make(map[string]bool)
	for _, s := range directionStations {
		dirSet[s] = true
	}

	for _, station := range path {
		if dirSet[station] {
			return true
		}
	}

	return false
}

// printDirectionStats - выводит статистику по направлениям
func printDirectionStats(stats map[string]domain.LocomotiveDirectionStats,
	depotDirections map[string][]domain.Direction) {

	// Группируем по депо
	byDepot := make(map[string][]domain.LocomotiveDirectionStats)
	for _, stat := range stats {
		byDepot[stat.Depo] = append(byDepot[stat.Depo], stat)
	}

	// Сортируем депо
	depots := make([]string, 0, len(byDepot))
	for d := range byDepot {
		depots = append(depots, d)
	}
	sort.Strings(depots)

	// Для каждого депо выводим примеры
	for _, depo := range depots {
		locStats := byDepot[depo]

		// Сортируем локомотивы по модели и номеру
		sort.Slice(locStats, func(i, j int) bool {
			if locStats[i].Model == locStats[j].Model {
				return locStats[i].Number < locStats[j].Number
			}
			return locStats[i].Model < locStats[j].Model
		})

		fmt.Printf("Депо %s:\n", depo)
		fmt.Printf("  Локомотивов в депо: %d\n", len(locStats))
		fmt.Printf("  Направлений от депо: %d\n", len(depotDirections[depo]))

		// Показываем направления этого депо
		if dirs, exists := depotDirections[depo]; exists {
			fmt.Printf("  Доступные направления:\n")
			for _, d := range dirs {
				fmt.Printf("    - %s (%s)\n", d.Name, d.Prefix)
			}
		}
		fmt.Println()

		// Показываем первые 5 локомотивов
		displayCount := 5
		if len(locStats) < displayCount {
			displayCount = len(locStats)
		}

		for i := 0; i < displayCount; i++ {
			stat := locStats[i]
			fmt.Printf("  Локомотив %s-%s:\n", stat.Model, stat.Number)
			fmt.Printf("    Всего поездок: %d\n", stat.TotalTrips)
			fmt.Printf("    Посещено направлений: %d\n", len(stat.DirectionVisits))

			if stat.MostPopularDirection != "" {
				percentage := float64(stat.MaxVisits) / float64(stat.TotalTrips) * 100
				fmt.Printf("    САМОЕ ПОПУЛЯРНОЕ НАПРАВЛЕНИЕ: %s\n",
					stat.PopularDirectionName)
				fmt.Printf("      Посещений: %d поездок (%.1f%%)\n",
					stat.MaxVisits, percentage)
			}
			fmt.Println()
		}

		if len(locStats) > displayCount {
			fmt.Printf("  ... и еще %d локомотивов\n", len(locStats)-displayCount)
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	// Общая статистика
	printOverallStats(stats)
}

// printOverallStats - общая статистика
func printOverallStats(stats map[string]domain.LocomotiveDirectionStats) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ОБЩАЯ СТАТИСТИКА ПО ПУНКТУ 2")
	fmt.Println(strings.Repeat("=", 80))

	totalLocomotives := len(stats)
	locWithFavorite := 0
	locWithSingleDirection := 0

	for _, stat := range stats {
		if stat.MostPopularDirection != "" {
			locWithFavorite++
		}
		if len(stat.DirectionVisits) == 1 {
			locWithSingleDirection++
		}
	}

	fmt.Printf("Всего локомотивов с данными: %d\n", totalLocomotives)
	fmt.Printf("Локомотивов с любимым направлением: %d\n", locWithFavorite)
	fmt.Printf("Процент: %.1f%%\n", float64(locWithFavorite)/float64(totalLocomotives)*100)
	fmt.Printf("Локомотивов, работающих на одном направлении: %d (%.1f%%)\n",
		locWithSingleDirection,
		float64(locWithSingleDirection)/float64(totalLocomotives)*100)

	// Считаем популярность направлений
	directionPopularity := make(map[string]int)
	directionNames := make(map[string]string)

	for _, stat := range stats {
		if stat.MostPopularDirection != "" {
			directionPopularity[stat.MostPopularDirection]++
			directionNames[stat.MostPopularDirection] = stat.PopularDirectionName
		}
	}

	type pop struct {
		id    string
		name  string
		count int
	}

	var popular []pop
	for did, cnt := range directionPopularity {
		popular = append(popular, pop{
			id:    did,
			name:  directionNames[did],
			count: cnt,
		})
	}

	sort.Slice(popular, func(i, j int) bool {
		return popular[i].count > popular[j].count
	})

	fmt.Println("\nСамые популярные направления среди локомотивов:")
	for i, p := range popular {
		if i >= 10 {
			break
		}
		percentage := float64(p.count) / float64(locWithFavorite) * 100
		fmt.Printf("  %d. %s - %d локомотивов (%.1f%%)\n",
			i+1, p.name, p.count, percentage)
	}
}

// printDepotAnalysis - анализ по депо
func printDepotAnalysis(stats map[string]domain.LocomotiveDirectionStats,
	depotDirections map[string][]domain.Direction) {

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("АНАЛИЗ ПО ДЕПО")
	fmt.Println(strings.Repeat("=", 80))

	// Группируем по депо
	byDepot := make(map[string][]domain.LocomotiveDirectionStats)
	for _, stat := range stats {
		byDepot[stat.Depo] = append(byDepot[stat.Depo], stat)
	}

	for depo, locs := range byDepot {
		// Считаем распределение любимых направлений в депо
		directionCount := make(map[string]int)

		for _, loc := range locs {
			if loc.MostPopularDirection != "" {
				directionCount[loc.MostPopularDirection]++
			}
		}

		fmt.Printf("\nДепо %s:\n", depo)
		fmt.Printf("  Локомотивов: %d\n", len(locs))
		fmt.Printf("  Распределение по любимым направлениям:\n")

		type depoDir struct {
			name  string
			count int
		}
		var dirs []depoDir
		for did, cnt := range directionCount {
			name := did
			for _, d := range depotDirections[depo] {
				if d.ID == did {
					name = d.Name
					break
				}
			}
			dirs = append(dirs, depoDir{name, cnt})
		}

		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].count > dirs[j].count
		})

		for _, d := range dirs {
			percentage := float64(d.count) / float64(len(locs)) * 100
			fmt.Printf("    %s: %d локомотивов (%.1f%%)\n",
				d.name, d.count, percentage)
		}
	}
}

// buildTask2Response - формирует структурированный ответ для API
func buildTask2Response(stats map[string]domain.LocomotiveDirectionStats, 
	depotDirections map[string][]domain.Direction) *responses.Task2Response {

	response := &responses.Task2Response{
		Depots:           make([]responses.DepotResponse, 0),
		OverallStats:     responses.OverallStats{},
		DirectionPopularity: make([]responses.PopularityItem, 0),
	}

	// Группируем по депо
	byDepot := make(map[string][]domain.LocomotiveDirectionStats)
	for _, stat := range stats {
		byDepot[stat.Depo] = append(byDepot[stat.Depo], stat)
	}

	// Сортируем депо
	depots := make([]string, 0, len(byDepot))
	for d := range byDepot {
		depots = append(depots, d)
	}
	sort.Strings(depots)

	// Формируем ответ по каждому депо
	for _, depo := range depots {
		locStats := byDepot[depo]
		
		// Сортируем локомотивы
		sort.Slice(locStats, func(i, j int) bool {
			if locStats[i].Model == locStats[j].Model {
				return locStats[i].Number < locStats[j].Number
			}
			return locStats[i].Model < locStats[j].Model
		})

		depotResponse := responses.DepotResponse{
			DepoCode:            depo,
			LocomotiveCount:     len(locStats),
			DirectionsCount:     len(depotDirections[depo]),
			AvailableDirections: make([]responses.DirectionInfo, 0),
			Locomotives:         make([]responses.LocomotiveStats, 0),
		}

		// Добавляем доступные направления
		if dirs, exists := depotDirections[depo]; exists {
			for _, d := range dirs {
				depotResponse.AvailableDirections = append(depotResponse.AvailableDirections, responses.DirectionInfo{
					ID:     d.ID,
					Name:   d.Name,
					Prefix: d.Prefix,
				})
			}
		}

		// Добавляем локомотивы
		for _, stat := range locStats {
			locStatsResp := buildLocomotiveStatsResponse(stat, depotDirections[depo])
			depotResponse.Locomotives = append(depotResponse.Locomotives, *locStatsResp)
		}

		response.Depots = append(response.Depots, depotResponse)
	}

	// Общая статистика
	totalLocomotives := len(stats)
	locWithFavorite := 0
	locWithSingleDirection := 0

	for _, stat := range stats {
		if stat.MostPopularDirection != "" {
			locWithFavorite++
		}
		if len(stat.DirectionVisits) == 1 {
			locWithSingleDirection++
		}
	}

	response.OverallStats = responses.OverallStats{
		TotalLocomotives:                totalLocomotives,
		LocomotivesWithFavorite:         locWithFavorite,
		LocomotivesWithFavoritePercent:  calculatePercentage(locWithFavorite, totalLocomotives),
		LocomotivesSingleDirection:      locWithSingleDirection,
		LocomotivesSingleDirectionPercent: calculatePercentage(locWithSingleDirection, totalLocomotives),
	}

	// Считаем популярность направлений
	directionPopularity := make(map[string]int)
	directionNames := make(map[string]string)

	for _, stat := range stats {
		if stat.MostPopularDirection != "" {
			directionPopularity[stat.MostPopularDirection]++
			directionNames[stat.MostPopularDirection] = stat.PopularDirectionName
		}
	}

	for dirID, count := range directionPopularity {
		response.DirectionPopularity = append(response.DirectionPopularity, responses.PopularityItem{
			DirectionID:   dirID,
			DirectionName: directionNames[dirID],
			Count:         count,
			Percentage:    calculatePercentage(count, locWithFavorite),
		})
	}

	// Сортируем по популярности
	sort.Slice(response.DirectionPopularity, func(i, j int) bool {
		return response.DirectionPopularity[i].Count > response.DirectionPopularity[j].Count
	})

	return response
}

// buildLocomotiveStatsResponse - формирует ответ для конкретного локомотива
func buildLocomotiveStatsResponse(stat domain.LocomotiveDirectionStats, 
	directions []domain.Direction) *responses.LocomotiveStats {

	locStatsResp := &responses.LocomotiveStats{
		Model:             stat.Model,
		Number:            stat.Number,
		Depo:              stat.Depo,
		TotalTrips:        stat.TotalTrips,
		VisitedDirections: stat.VisitedDirections,
		DirectionVisits:   make([]responses.DirectionVisit, 0),
	}

	// Добавляем информацию о посещениях направлений
	for dirID, visits := range stat.DirectionVisits {
		// Находим название направления
		dirName := dirID
		for _, dir := range directions {
			if dir.ID == dirID {
				dirName = dir.Name
				break
			}
		}

		locStatsResp.DirectionVisits = append(locStatsResp.DirectionVisits, responses.DirectionVisit{
			DirectionID:   dirID,
			DirectionName: dirName,
			Visits:        visits,
			Percentage:    calculatePercentage(visits, stat.TotalTrips),
		})
	}

	// Добавляем самое популярное направление
	if stat.MostPopularDirection != "" {
		locStatsResp.MostPopular = &responses.MostPopularDirection{
			DirectionID:   stat.MostPopularDirection,
			DirectionName: stat.PopularDirectionName,
			Visits:        stat.MaxVisits,
			Percentage:    calculatePercentage(stat.MaxVisits, stat.TotalTrips),
		}
	}

	return locStatsResp
}

// calculatePercentage - вспомогательная функция для расчета процентов
func calculatePercentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}