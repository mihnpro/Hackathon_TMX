package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	
	"github.com/mihnpro/Hackathon_TMX/internal/services"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/handlers"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/routes"
)

func main() {
	// Создаем сервисы
	task1Service := services.NewAlgorithmService("./data/locomotives_displacement.csv")
	task2Service := services.NewMostPopularTripService("./data/locomotives_displacement.csv")
	task3Service := services.NewVisualizationService("./data/locomotives_displacement.csv")
	
	// НОВОЕ: создаем ML сервис для интеграции с Python
	mlService := services.NewMLIntegrationService("http://localhost:8000")
	
	// Создаем обработчики
	task1Handler := handlers.NewTask1Handler(task1Service)
	task2Handler := handlers.NewTask2Handler(task2Service)
	task3Handler := handlers.NewTask3Handler(task3Service)
	
	// НОВОЕ: создаем ML обработчик
	mlHandler := handlers.NewMLHandler(mlService)
	
	// Создаем временную директорию для карт
	mapsDir := "./maps"
	if err := os.MkdirAll(mapsDir, 0755); err != nil {
		log.Printf("⚠️ Не удалось создать директорию для карт: %v", err)
	}
	
	// НОВОЕ: создаем директорию для загружаемых файлов
	os.MkdirAll("./uploads", 0755)
	
	// Настраиваем Gin
	router := gin.Default()
	
	// Добавляем CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	// Добавляем middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	
	// Настраиваем все маршруты (API + Frontend) через единый файл routes.go
	// НОВОЕ: передаем mlHandler
	routes.SetupAllRoutes(
		router, 
		task1Handler, 
		task2Handler, 
		task3Handler, 
		mlHandler, // добавляем ML handler
		mapsDir,
	)
	
	// Graceful shutdown для очистки временных файлов
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("🛑 Получен сигнал завершения, очищаем ресурсы...")
		
		// Очищаем временные файлы от сервиса визуализации
		if vs, ok := task3Service.(interface{ Cleanup() }); ok {
			vs.Cleanup()
		}
		
		// Также удаляем локальную директорию maps если она существует
		if err := os.RemoveAll(mapsDir); err != nil {
			log.Printf("⚠️ Ошибка при удалении директории %s: %v", mapsDir, err)
		} else {
			log.Printf("🧹 Директория %s удалена", mapsDir)
		}
		
		// НОВОЕ: удаляем загруженные файлы
		os.RemoveAll("./uploads")
		
		os.Exit(0)
	}()
	
	// Запускаем сервер
	log.Println("🌐 Сервер запущен на :8080")
	log.Println("📚 Доступные endpoints:")
	log.Println("   GET / - Главная страница")
	log.Println("   GET /task1 - Задание 1")
	log.Println("   GET /task2 - Задание 2")
	log.Println("   GET /task3 - Задание 3")
	log.Println()
	log.Println("   🔹 API Задание 1:")
	log.Println("      GET    /api/v1/task1/branches           - все ветки")
	log.Println("      GET    /api/v1/task1/depots             - список депо")
	log.Println("      GET    /api/v1/task1/depots/:depo/branches - ветки депо")
	log.Println()
	log.Println("   🔹 API Задание 2:")
	log.Println("      GET    /api/v1/popular-direction                 - все направления")
	log.Println("      GET    /api/v1/locomotives/:series/:number/popular-direction - направление локомотива")
	log.Println()
	log.Println("   🔹 API Задание 3:")
	log.Println("      GET    /api/v1/task3/depots             - список депо")
	log.Println("      GET    /api/v1/task3/depots/:depo       - информация о депо")
	log.Println("      POST   /api/v1/task3/generate           - генерация карт")
	log.Println("      GET    /maps/*                           - сгенерированные карты")
	log.Println()
	log.Println("   🔹 НОВОЕ: ML Wear Prediction:")
	log.Println("      POST   /api/v1/ml/predict        - предсказание (JSON в теле)")
	log.Println("      POST   /api/v1/ml/upload         - загрузка файла с данными")
	log.Println("      GET    /api/v1/ml/health         - проверка ML сервиса")
	log.Println("      GET    /api/v1/ml/info           - информация о модели")
	log.Println("      GET    /ml                        - веб-интерфейс для ML")
	
	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Ошибка запуска сервера:", err)
	}
}