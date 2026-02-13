package services

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Record представляет запись о посещении станции
type Record struct {
	Series  string
	Number  string
	Time    time.Time
	Station string
	Depo    string
}

// Trip представляет одну поездку локомотива
type Trip struct {
	StartTime  time.Time
	EndTime    time.Time
	Stations   []string
	IsComplete bool
}

// Locomotive представляет локомотив и все его записи
type Locomotive struct {
	Key     string
	Series  string
	Number  string
	Depo    string
	Records []Record
	Trips   []Trip
}

// ImprovedBranch - улучшенная структура для хранения ветки
type ImprovedBranch struct {
	Depo          string
	CoreStations  []string      // уникальные станции в порядке от депо
	BranchID      string        // уникальный идентификатор ветки
	AllPaths      [][]string    // все варианты проезда по ветке
	Terminals     map[string]int // конечные станции и частота посещения
	Length        int           // длина ветки в станциях
}

func main() {
	// 1. Загружаем данные
	fmt.Println("Загрузка данных...")
	locomotives := loadData("../data/locomotives_displacement.csv")
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
	depotBranches := analyzeBranchesImproved(locomotives)

	// 4. Выводим результаты
	printImprovedResults(depotBranches)
}

// loadData загружает данные из CSV файла
func loadData(filename string) map[string]Locomotive {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	locomotives := make(map[string]Locomotive)

	// Пропускаем заголовок
	scanner.Scan()

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		parts := strings.Split(line, ",")

		if len(parts) < 5 {
			fmt.Printf("Предупреждение: строка %d имеет меньше 5 полей\n", lineNum)
			continue
		}

		// Очищаем поля от пробелов
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		// Парсим время
		timeStr := parts[2]
		var t time.Time
		var err error

		// Пробуем разные форматы времени
		formats := []string{
			"2006-01-02T15:04:05.000000",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}

		for _, format := range formats {
			t, err = time.Parse(format, timeStr)
			if err == nil {
				break
			}
		}

		if err != nil {
			fmt.Printf("Предупреждение: строка %d, ошибка парсинга времени: %s\n", lineNum, timeStr)
			continue
		}

		record := Record{
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
			locomotives[key] = Locomotive{
				Key:     key,
				Series:  record.Series,
				Number:  record.Number,
				Depo:    record.Depo,
				Records: []Record{record},
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

// splitIntoTrips разбивает записи локомотива на отдельные поездки
func splitIntoTrips(records []Record) []Trip {
	var trips []Trip

	if len(records) == 0 {
		return trips
	}

	currentTrip := Trip{
		StartTime:  records[0].Time,
		Stations:   []string{},
		IsComplete: false,
	}

	for i, rec := range records {
		// Добавляем станцию в текущую поездку
		currentTrip.Stations = append(currentTrip.Stations, rec.Station)

		// Проверяем, вернулись ли в депо
		if rec.Station == rec.Depo {
			currentTrip.EndTime = rec.Time
			currentTrip.IsComplete = true
			trips = append(trips, currentTrip)

			// Начинаем новую поездку, если есть еще записи
			if i < len(records)-1 {
				currentTrip = Trip{
					StartTime:  records[i+1].Time,
					Stations:   []string{},
					IsComplete: false,
				}
			}
		}
	}

	// Если последняя поездка не завершена, добавляем её
	if !currentTrip.IsComplete && len(currentTrip.Stations) > 0 {
		trips = append(trips, currentTrip)
	}

	return trips
}

// analyzeBranchesImproved - улучшенный анализ веток с кластеризацией
func analyzeBranchesImproved(locomotives map[string]Locomotive) map[string][]ImprovedBranch {
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
	depotBranches := make(map[string][]ImprovedBranch)

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

			branch := ImprovedBranch{
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

// cleanStops - убирает только стоянки, оставляя маршрут
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

// generateBranchID - создает уникальный идентификатор ветки
func generateBranchID(coreStations []string) string {
	if len(coreStations) == 0 {
		return "unknown"
	}
	// Берем первую и последнюю станцию для идентификации
	return coreStations[0] + "_to_" + coreStations[len(coreStations)-1]
}

// printImprovedResults - выводит улучшенные результаты
func printImprovedResults(depotBranches map[string][]ImprovedBranch) {
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

		fmt.Printf("Депо %s:\n", depo)
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
				len(branch.CoreStations), branch.CoreStations)

			if len(terminals) > 0 {
				fmt.Printf("    Конечные станции (с частотой):\n")
				for j, term := range terminals {
					if j < 5 { // Показываем только топ-5 конечных
						fmt.Printf("      - %s (%d раз)\n", term, branch.Terminals[term])
					}
				}
			} else {
				fmt.Printf("    Конечные станции: не определены\n")
			}

			// Показываем пример одного пути
			if len(branch.AllPaths) > 0 {
				example := branch.AllPaths[0]
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
			longest = append(longest, longBranch{
				depo:   depo,
				length: b.Length,
				route:  strings.Join(b.CoreStations, " → "),
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


func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}