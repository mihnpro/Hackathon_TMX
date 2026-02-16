package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mihnpro/Hackathon_TMX/internal/services"
)

func main() {
	// Парсим аргументы командной строки
	var (
		task       = flag.String("task", "all", "Задача для выполнения: 1, 2, 3, all")
		dataPath   = flag.String("data", "./data/locomotives_displacement.csv", "Путь к файлу с данными")
		depoForMap = flag.String("depo", "940006", "ID депо для визуализации (для задачи 3)")
		maxLoco    = flag.Int("max", 10, "Максимальное количество локомотивов на карте")
	)
	flag.Parse()

	// Создаем сервисы
	algorithmSvc := services.NewAlgorithmService(*dataPath, "./data/station_info.csv")
	popularTripSvc := services.NewMostPopularTripService(*dataPath, "./data/station_info.csv")
	visualizationSvc := services.NewVisualizationService(*dataPath)

	// Засекаем время выполнения
	startTime := time.Now()
	fmt.Printf("Запуск анализа. Время: %s\n", startTime.Format("15:04:05"))
	fmt.Printf("Задача: %s, Депо для карты: %s\n", *task, *depoForMap)
	fmt.Printf("Путь к данным: %s\n\n", *dataPath)

	// Проверяем, что для задачи 3 указано корректное депо
	if *task == "3" && *depoForMap == "station_info" {
		fmt.Println("⚠️  Внимание: Используется значение по умолчанию 'station_info' для депо.")
		fmt.Println("   Возможно, вы хотели указать ID депо, например: -depo=940006")
		fmt.Println("   Продолжаем с '940006'...")
		*depoForMap = "940006"
	}

	// Выполняем задачи
	switch *task {
	case "1":
		// Только пункт 1
		algorithmSvc.RunAlgorithm()

	case "2":
		// Только пункт 2
		popularTripSvc.RunMostPopularTrip()

	case "3":
		// Только пункт 3 - визуализация
		fmt.Println("Запуск визуализации...")
		if err := visualizationSvc.GenerateMap(*depoForMap, *maxLoco); err != nil {
			log.Fatalf("Ошибка визуализации: %v", err)
		}

	case "all":
		// Все пункты
		fmt.Println("=== ПУНКТ 1 ===")
		algorithmSvc.RunAlgorithm()

		fmt.Println("\n=== ПУНКТ 2 ===")
		popularTripSvc.RunMostPopularTrip()

		fmt.Println("\n=== ПУНКТ 3 ===")
		if err := visualizationSvc.GenerateAllMaps(*depoForMap, *maxLoco); err != nil {
			log.Fatalf("Ошибка визуализации: %v", err)
		}

	default:
		log.Fatalf("Неизвестная задача: %s. Используйте 1, 2, 3 или all", *task)
	}

	// Итоговое время
	elapsed := time.Since(startTime)
	fmt.Printf("\n✅ Анализ завершен за %s\n", elapsed)
}