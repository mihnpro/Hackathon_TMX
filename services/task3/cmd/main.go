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
		depoForMap = flag.String("depo", "940006", "Депо для визуализации (для задачи 3)")
		// maxLoco    = flag.Int("max", 10, "Максимальное количество локомотивов на карте")
	)
	flag.Parse()

	// Создаем сервисы
	algorithmSvc := services.NewAlgorithmService(*dataPath)
	popularTripSvc := services.NewMostPopularTripService(*dataPath)
	visualizationSvc := services.NewVisualizationService(*dataPath)

	// Засекаем время выполнения
	startTime := time.Now()
	fmt.Printf("Запуск анализа. Время: %s\n", startTime.Format("15:04:05"))
	fmt.Printf("Задача: %s, Депо для карты: %s\n\n", *task, *depoForMap)

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
		if err := visualizationSvc.GenerateAllMaps(*depoForMap); err != nil {
			log.Fatalf("Ошибка визуализации: %v", err)
		}

	case "all":
		// Все пункты
		fmt.Println("=== ПУНКТ 1 ===")
		algorithmSvc.RunAlgorithm()

		fmt.Println("\n=== ПУНКТ 2 ===")
		popularTripSvc.RunMostPopularTrip()

		fmt.Println("\n=== ПУНКТ 3 ===")
		if err := visualizationSvc.GenerateAllMaps(*depoForMap); err != nil {
			log.Fatalf("Ошибка визуализации: %v", err)
		}

	default:
		log.Fatalf("Неизвестная задача: %s. Используйте 1, 2, 3 или all", *task)
	}

	// Итоговое время
	elapsed := time.Since(startTime)
	fmt.Printf("\n✅ Анализ завершен за %s\n", elapsed)
}
