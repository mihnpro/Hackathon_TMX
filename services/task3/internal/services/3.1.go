package services

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mihnpro/Hackathon_TMX/internal/domain"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/models/responses"
)

type algorithmService struct {
	dataPath     string
	stationsPath string
	stations     map[string]string // кэш станций: ID -> название
}

type AlgorithmService interface {
	// Для консольного режима
	RunAlgorithm()
	
	// Для API режима
	GetBranchAnalysis() (*responses.Task1Response, error)
	GetDepotBranches(depoCode string) (*responses.DepotBranches, error)
}

func NewAlgorithmService(dataPath, stationsPath string) AlgorithmService {
	svc := &algorithmService{
		dataPath:     dataPath,
		stationsPath: stationsPath,
		stations:     make(map[string]string),
	}
	// Загружаем станции сразу при создании сервиса
	svc.loadStations()
	return svc
}

// loadStations загружает названия станций из CSV
func (a *algorithmService) loadStations() {
	file, err := os.Open(a.stationsPath)
	if err != nil {
		fmt.Printf("Ошибка загрузки станций: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	
	// Читаем заголовок
	header, err := reader.Read()
	if err != nil {
		fmt.Printf("Ошибка чтения заголовка: %v\n", err)
		return
	}
	
	// Находим индексы колонок
	idIdx := -1
	nameIdx := -1
	for i, col := range header {
		col = strings.ToLower(strings.TrimSpace(col))
		if col == "station" {
			idIdx = i
		} else if col == "station_name" {
			nameIdx = i
		}
	}

	if idIdx == -1 || nameIdx == -1 {
		fmt.Printf("Не найдены нужные колонки в файле станций\n")
		return
	}

	// Читаем записи
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Ошибка чтения записей: %v\n", err)
		return
	}

	for _, record := range records {
		if len(record) > idIdx && len(record) > nameIdx {
			stationID := strings.TrimSpace(record[idIdx])
			stationName := strings.TrimSpace(record[nameIdx])
			if stationID != "" && stationName != "" {
				a.stations[stationID] = stationName
			}
		}
	}

	fmt.Printf("Загружено названий станций: %d\n", len(a.stations))
}

// getStationName возвращает название станции по ID
func (a *algorithmService) getStationName(stationID string) string {
	if name, ok := a.stations[stationID]; ok {
		return name
	}
	// Если название не найдено, возвращаем ID
	return stationID
}

// convertStationsToNames преобразует срез ID станций в срез названий
func (a *algorithmService) convertStationsToNames(stationIDs []string) []string {
	names := make([]string, len(stationIDs))
	for i, id := range stationIDs {
		names[i] = a.getStationName(id)
	}
	return names
}

// convertPathToNames преобразует путь с ID в путь с названиями
func (a *algorithmService) convertPathToNames(path []string) []string {
	return a.convertStationsToNames(path)
}

// convertPathsToNames преобразует несколько путей
func (a *algorithmService) convertPathsToNames(paths [][]string) [][]string {
	result := make([][]string, len(paths))
	for i, path := range paths {
		result[i] = a.convertPathToNames(path)
	}
	return result
}

// --- ДОБАВЛЕННЫЕ ФУНКЦИИ ИЗ ОРИГИНАЛЬНОГО КОДА ---

// extractCorePath - извлекает ядро пути (уникальные станции в порядке появления)
func extractCorePath(path []string, depoID string) []string {
	var core []string
	seen := make(map[string]bool)

	for _, s := range path {
		if s == depoID {
			continue // пропускаем депо в ядре
		}
		if !seen[s] {
			core = append(core, s)
			seen[s] = true
		}
	}

	return core
}

// extractUniqueStations - извлекает уникальные станции из пути с сохранением порядка
func extractUniqueStations(path []string) []string {
	var unique []string
	seen := make(map[string]bool)

	for _, s := range path {
		if !seen[s] {
			unique = append(unique, s)
			seen[s] = true
		}
	}

	return unique
}

// clusterPaths - группирует похожие пути в кластеры
func clusterPaths(paths [][]string, depoID string) [][][]string {
	var clusters [][][]string

	for _, path := range paths {
		if len(path) < 2 {
			continue
		}

		corePath := extractCorePath(path, depoID)
		if len(corePath) == 0 {
			continue
		}

		// Ищем подходящий кластер
		found := false
		for i, cluster := range clusters {
			if len(cluster) == 0 {
				continue
			}

			// Проверяем первый путь в кластере
			clusterCore := extractCorePath(cluster[0], depoID)
			if isSimilarCore(corePath, clusterCore) {
				clusters[i] = append(clusters[i], path)
				found = true
				break
			}
		}

		if !found {
			clusters = append(clusters, [][]string{path})
		}
	}

	return clusters
}

// isSimilarCore - проверяет, похожи ли два ядра путей
func isSimilarCore(core1, core2 []string) bool {
	if len(core1) == 0 || len(core2) == 0 {
		return false
	}

	// Проверяем первые 3 станции (начало маршрута)
	minLen := 3
	if len(core1) < minLen || len(core2) < minLen {
		minLen = min(len(core1), len(core2))
	}

	// Проверяем совпадение начала
	for i := 0; i < minLen; i++ {
		if core1[i] != core2[i] {
			return false
		}
	}

	// Проверяем связь конечных станций
	last1 := core1[len(core1)-1]
	last2 := core2[len(core2)-1]

	return last1 == last2 ||
		contains(core1, last2) ||
		contains(core2, last1)
}

// findCoreDirection - находит основное направление кластера
func findCoreDirection(cluster [][]string, depoID string) []string {
	if len(cluster) == 0 {
		return nil
	}

	// Собираем все уникальные станции из всех путей кластера
	allStations := make(map[string]bool)
	for _, path := range cluster {
		for _, s := range path {
			if s != depoID {
				allStations[s] = true
			}
		}
	}

	// Для каждой станции вычисляем среднюю позицию в путях
	stationPositions := make(map[string][]float64)

	for _, path := range cluster {
		uniquePath := extractUniqueStations(path)
		for i, s := range uniquePath {
			if s != depoID {
				stationPositions[s] = append(stationPositions[s], float64(i))
			}
		}
	}

	// Вычисляем средние позиции
	type stationInfo struct {
		station string
		avgPos  float64
	}

	var positions []stationInfo
	for s, posList := range stationPositions {
		sum := 0.0
		for _, p := range posList {
			sum += p
		}
		avg := sum / float64(len(posList))
		positions = append(positions, stationInfo{s, avg})
	}

	// Сортируем по средней позиции
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].avgPos < positions[j].avgPos
	})

	// Извлекаем станции в порядке возрастания
	var direction []string
	for _, p := range positions {
		direction = append(direction, p.station)
	}

	// Если получилось слишком много станций, берем только основные
	if len(direction) > 20 {
		// Оставляем только станции, которые есть в большинстве путей
		threshold := len(cluster) / 2
		var frequent []string

		for _, s := range direction {
			count := 0
			for _, path := range cluster {
				if contains(path, s) {
					count++
				}
			}
			if count >= threshold {
				frequent = append(frequent, s)
			}
		}

		if len(frequent) > 0 {
			return frequent
		}
	}

	return direction
}


// --- Модифицированные функции ---

// RunAlgorithm - для консольного режима (с названиями станций)
func (a *algorithmService) RunAlgorithm() {
	// 1. Загружаем данные
	fmt.Println("Загрузка данных...")
	locomotives := loadData(a.dataPath)
	fmt.Printf("Загружено локомотивов: %d\n\n", len(locomotives))

	// 2. Разбиваем на поездки
	fmt.Println("Разбиение на поездки...")
	for key, loc := range locomotives {
		trips := splitIntoTrips(loc.Records)
		loc.Trips = trips
		locomotives[key] = loc
		fmt.Printf("  Локомотив %s: %d поездок\n", key, len(trips))
	}
	fmt.Println()

	// 3. УЛУЧШЕННЫЙ анализ веток
	fmt.Println("Кластеризация веток...")
	depotBranches := a.analyzeBranchesImproved(locomotives)

	// 4. Выводим результаты (с названиями)
	a.printImprovedResults(depotBranches)
}

// GetBranchAnalysis - для API режима (полный анализ)
func (a *algorithmService) GetBranchAnalysis() (*responses.Task1Response, error) {
	// 1. Загружаем данные
	locomotives := loadData(a.dataPath)

	// 2. Разбиваем на поездки
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Анализ веток
	depotBranches := a.analyzeBranchesImproved(locomotives)

	// 4. Формируем ответ (с названиями)
	return a.buildTask1Response(depotBranches), nil
}

// GetDepotBranches - для API режима (конкретное депо)
func (a *algorithmService) GetDepotBranches(depoCode string) (*responses.DepotBranches, error) {
	// 1. Загружаем данные
	locomotives := loadData(a.dataPath)

	// 2. Разбиваем на поездки
	for key, loc := range locomotives {
		loc.Trips = splitIntoTrips(loc.Records)
		locomotives[key] = loc
	}

	// 3. Анализ веток
	depotBranches := a.analyzeBranchesImproved(locomotives)

	// 4. Ищем нужное депо
	branches, exists := depotBranches[depoCode]
	if !exists {
		return nil, nil
	}

	// 5. Формируем ответ для конкретного депо (с названиями)
	return a.buildDepotBranchesResponse(depoCode, branches), nil
}

// analyzeBranchesImproved - улучшенный анализ веток с кластеризацией (ID остаются внутри)
func (a *algorithmService) analyzeBranchesImproved(locomotives map[string]domain.Locomotive) map[string][]domain.ImprovedBranch {
	// Собираем все пути по депо
	allPaths := make(map[string][][]string)

	for _, loc := range locomotives {
		for _, trip := range loc.Trips {
			if len(trip.Stations) < 2 {
				continue
			}

			// Очищаем путь от стоянок
			cleanPath := cleanStops(trip.Stations)

			// Добавляем путь в общий список для депо
			allPaths[loc.Depo] = append(allPaths[loc.Depo], cleanPath)
		}
	}

	// Кластеризуем пути по направлениям
	depotBranches := make(map[string][]domain.ImprovedBranch)

	for depo, paths := range allPaths {
		fmt.Printf("  Анализ депо %s: %d путей\n", depo, len(paths))

		// Группируем похожие пути
		clusters := clusterPaths(paths, depo)

		for _, cluster := range clusters {
			if len(cluster) == 0 {
				continue
			}

			// Определяем основное направление
			coreDirection := findCoreDirection(cluster, depo)

			// Пропускаем слишком короткие ветки
			if len(coreDirection) < 1 {
				continue
			}

			// Собираем все конечные станции
			terminals := make(map[string]int)
			for _, path := range cluster {
				if len(path) > 1 {
					terminal := path[len(path)-1]
					terminals[terminal]++
				}
			}

			branch := domain.ImprovedBranch{
				Depo:         depo,
				CoreStations: coreDirection,
				BranchID:     generateBranchID(coreDirection),
				AllPaths:     cluster,
				Terminals:    terminals,
				Length:       len(coreDirection),
			}

			depotBranches[depo] = append(depotBranches[depo], branch)
		}
	}

	return depotBranches
}

// buildTask1Response - формирует полный ответ для API (с названиями)
func (a *algorithmService) buildTask1Response(depotBranches map[string][]domain.ImprovedBranch) *responses.Task1Response {
	response := &responses.Task1Response{
		Depots:          make([]responses.DepotBranches, 0),
		LongestBranches: make([]responses.LongestBranch, 0),
	}

	// Сортируем депо
	depots := make([]string, 0, len(depotBranches))
	for depo := range depotBranches {
		depots = append(depots, depo)
	}
	sort.Strings(depots)

	totalBranches := 0
	totalTerminals := 0

	// Формируем информацию по каждому депо
	for _, depo := range depots {
		branches := depotBranches[depo]
		depotResponse := a.buildDepotBranchesResponse(depo, branches)
		
		response.Depots = append(response.Depots, *depotResponse)
		totalBranches += len(branches)
		
		for _, b := range branches {
			totalTerminals += len(b.Terminals)
			
			// Для списка самых длинных веток
			routeNames := a.convertStationsToNames(b.CoreStations)
			response.LongestBranches = append(response.LongestBranches, responses.LongestBranch{
				DepoCode:    depo,
				BranchID:    b.BranchID,
				Length:      b.Length,
				Route:       routeNames,
				RouteString: strings.Join(routeNames, " → "),
			})
		}
	}

	// Общая статистика
	response.OverallStats = responses.OverallStatsTask1{
		TotalDepots:        len(depots),
		TotalBranches:      totalBranches,
		TotalTerminals:     totalTerminals,
		AvgBranchesPerDepo: float64(totalBranches) / float64(len(depots)),
	}

	// Сортируем самые длинные ветки
	sort.Slice(response.LongestBranches, func(i, j int) bool {
		return response.LongestBranches[i].Length > response.LongestBranches[j].Length
	})

	// Оставляем топ-20
	if len(response.LongestBranches) > 20 {
		response.LongestBranches = response.LongestBranches[:20]
	}

	return response
}

// buildDepotBranchesResponse - формирует ответ для конкретного депо (с названиями)
func (a *algorithmService) buildDepotBranchesResponse(depoCode string, branches []domain.ImprovedBranch) *responses.DepotBranches {
	// Сортируем ветки по длине
	sort.Slice(branches, func(i, j int) bool {
		return len(branches[i].CoreStations) > len(branches[j].CoreStations)
	})

	depotResponse := &responses.DepotBranches{
		DepoCode:    depoCode,
		BranchCount: len(branches),
		Branches:    make([]responses.BranchInfo, 0, len(branches)),
	}

	for _, branch := range branches {
		// Собираем конечные станции
		terminals := make([]responses.TerminalInfo, 0, len(branch.Terminals))
		
		// Считаем общее количество поездок в этой ветке
		totalTripsInBranch := 0
		for _, count := range branch.Terminals {
			totalTripsInBranch += count
		}

		// Сортируем терминалы по частоте
		terminalList := make([]string, 0, len(branch.Terminals))
		for t := range branch.Terminals {
			terminalList = append(terminalList, t)
		}
		sort.Slice(terminalList, func(i, j int) bool {
			return branch.Terminals[terminalList[i]] > branch.Terminals[terminalList[j]]
		})

		// Берем топ-10 терминалов
		for i, term := range terminalList {
			if i >= 10 {
				break
			}
			frequency := float64(branch.Terminals[term]) / float64(totalTripsInBranch) * 100
			terminals = append(terminals, responses.TerminalInfo{
				Station:   a.getStationName(term), // Преобразуем ID в название
				Visits:    branch.Terminals[term],
				Frequency: frequency,
			})
		}

		// Пример пути (преобразуем ID в названия)
		examplePath := []string{}
		if len(branch.AllPaths) > 0 {
			examplePath = a.convertPathToNames(branch.AllPaths[0])
			if len(examplePath) > 10 {
				examplePath = examplePath[:10]
			}
		}

		branchInfo := responses.BranchInfo{
			BranchID:      branch.BranchID,
			CoreStations:  a.convertStationsToNames(branch.CoreStations), // Преобразуем ID в названия
			StationCount:  len(branch.CoreStations),
			Terminals:     terminals,
			ExamplePath:   examplePath,
		}

		depotResponse.Branches = append(depotResponse.Branches, branchInfo)
	}

	return depotResponse
}

// printImprovedResults - выводит улучшенные результаты (с названиями)
func (a *algorithmService) printImprovedResults(depotBranches map[string][]domain.ImprovedBranch) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("УЛУЧШЕННЫЙ АНАЛИЗ: РЕАЛЬНЫЕ ВЕТКИ ДЕПО")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	// Сортируем депо для вывода
	depots := make([]string, 0, len(depotBranches))
	for depo := range depotBranches {
		depots = append(depots, depo)
	}
	sort.Strings(depots)

	totalBranches := 0
	totalTerminals := 0

	for _, depo := range depots {
		branches := depotBranches[depo]

		// Сортируем ветки по длине
		sort.Slice(branches, func(i, j int) bool {
			return len(branches[i].CoreStations) < len(branches[j].CoreStations)
		})

		fmt.Printf("Депо %s:\n", a.getStationName(depo))
		fmt.Printf("  Реальных веток: %d\n", len(branches))
		fmt.Println()

		for i, branch := range branches {
			// Собираем конечные станции
			terminals := make([]string, 0, len(branch.Terminals))
			terminalCounts := make([]int, 0, len(branch.Terminals))
			for t, cnt := range branch.Terminals {
				terminals = append(terminals, t)
				terminalCounts = append(terminalCounts, cnt)
			}

			// Сортируем конечные станции по частоте
			sort.Slice(terminals, func(i2, j2 int) bool {
				return branch.Terminals[terminals[i2]] > branch.Terminals[terminals[j2]]
			})

			fmt.Printf("  Ветка %d (ID: %s):\n", i+1, branch.BranchID)
			fmt.Printf("    Основной маршрут (%d станций): %v\n",
				len(branch.CoreStations), a.convertStationsToNames(branch.CoreStations))

			if len(terminals) > 0 {
				fmt.Printf("    Конечные станции (с частотой):\n")
				for j, term := range terminals {
					if j < 5 { // Показываем только топ-5 конечных
						fmt.Printf("      - %s (%d раз)\n", a.getStationName(term), branch.Terminals[term])
					}
				}
			} else {
				fmt.Printf("    Конечные станции: не определены\n")
			}

			// Показываем пример одного пути
			if len(branch.AllPaths) > 0 {
				example := a.convertPathToNames(branch.AllPaths[0])
				if len(example) > 15 {
					example = example[:15]
				}
				fmt.Printf("    Пример пути: %v\n", example)
			}
			fmt.Println()
		}

		totalBranches += len(branches)
		for _, b := range branches {
			totalTerminals += len(b.Terminals)
		}

		fmt.Println(strings.Repeat("-", 60))
	}

	// Итоговая статистика
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ИТОГОВАЯ СТАТИСТИКА")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Всего депо: %d\n", len(depots))
	fmt.Printf("Всего реальных веток: %d\n", totalBranches)
	fmt.Printf("Всего конечных станций: %d\n", totalTerminals)
	fmt.Printf("Среднее количество веток на депо: %.1f\n",
		float64(totalBranches)/float64(len(depots)))

	// Анализ самых длинных веток
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("САМЫЕ ДЛИННЫЕ ВЕТКИ")
	fmt.Println(strings.Repeat("=", 80))

	type longBranch struct {
		depo   string
		length int
		route  string
	}

	var longest []longBranch
	for depo, branches := range depotBranches {
		for _, b := range branches {
			routeNames := a.convertStationsToNames(b.CoreStations)
			longest = append(longest, longBranch{
				depo:   a.getStationName(depo),
				length: b.Length,
				route:  strings.Join(routeNames, " → "),
			})
		}
	}

	sort.Slice(longest, func(i, j int) bool {
		return longest[i].length > longest[j].length
	})

	for i, lb := range longest {
		if i >= 10 {
			break
		}
		fmt.Printf("%d. Депо %s: %d станций\n", i+1, lb.depo, lb.length)
		fmt.Printf("   Маршрут: %s\n", lb.route)
	}
}