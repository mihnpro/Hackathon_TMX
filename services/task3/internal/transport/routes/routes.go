package routes

import (
	"github.com/gin-gonic/gin"
	
	"github.com/mihnpro/Hackathon_TMX/internal/transport/handlers"
)

// SetupAllRoutes настраивает все маршруты приложения (API + Frontend)
func SetupAllRoutes(
	router *gin.Engine,
	task1Handler *handlers.Task1Handler,
	task2Handler *handlers.Task2Handler,
	task3Handler *handlers.Task3Handler,
	mapsDir string, // временная директория для карт
) {
	// Настраиваем API маршруты
	setupAPIRoutes(router, task1Handler, task2Handler, task3Handler)
	
	// Настраиваем фронтенд маршруты
	setupFrontendRoutes(router)
	
	// Раздаем сгенерированные карты из временной директории
	router.Static("/maps", mapsDir)
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

// setupAPIRoutes настраивает все API маршруты
func setupAPIRoutes(
	router *gin.Engine,
	task1Handler *handlers.Task1Handler,
	task2Handler *handlers.Task2Handler,
	task3Handler *handlers.Task3Handler,
) {
	api := router.Group("/api/v1")
	{
		// ========== ЗАДАНИЕ 1 ==========
		task1 := api.Group("/task1")
		{
			// Получить полный анализ веток
			task1.GET("/branches", task1Handler.GetBranchAnalysis)
			
			// Получить список всех депо
			task1.GET("/depots", task1Handler.GetAllDepots)
			
			// Получить ветки для конкретного депо
			task1.GET("/depots/:depo/branches", task1Handler.GetDepotBranches)
		}
		
		// ========== ЗАДАНИЕ 2 ==========
		// (без префикса /task2 как в вашем коде)
		api.GET("/popular-direction", task2Handler.GetPopularDirections)
		api.GET("/locomotives/:series/:number/popular-direction", task2Handler.GetLocomotivePopularDirection)
		
		// ========== ЗАДАНИЕ 3 ==========
		task3 := api.Group("/task3")
		{
			// Получить список всех депо
			task3.GET("/depots", task3Handler.GetAvailableDepots)
			
			// Получить информацию о конкретном депо
			task3.GET("/depots/:depo", task3Handler.GetDepotInfo)
			
			// Сгенерировать карты для депо
			task3.POST("/generate", task3Handler.GenerateMaps)
		}
	}
}

// setupFrontendRoutes настраивает маршруты для раздачи статических файлов фронтенда
func setupFrontendRoutes(router *gin.Engine) {
	// Главная страница
	router.StaticFile("/", "./frontend/index.html")
	router.StaticFile("/index.html", "./frontend/index.html")
	
	// Страницы заданий
	setupTaskFrontendRoutes(router, "task1")
	setupTaskFrontendRoutes(router, "task2")
	setupTaskFrontendRoutes(router, "task3")
	
	// Общие файлы
	router.Static("/shared/css", "./frontend/shared/css")
	router.Static("/shared/js", "./frontend/shared/js")
}

// setupTaskFrontendRoutes настраивает фронтенд маршруты для конкретного задания
func setupTaskFrontendRoutes(router *gin.Engine, taskName string) {
	// HTML страница
	router.StaticFile("/"+taskName, "./frontend/"+taskName+"/index.html")
	router.StaticFile("/"+taskName+"/", "./frontend/"+taskName+"/index.html")
	
	// Статические файлы (CSS, JS)
	router.Static("/"+taskName+"/css", "./frontend/"+taskName+"/css")
	router.Static("/"+taskName+"/js", "./frontend/"+taskName+"/js")
	router.Static("/"+taskName+"/assets", "./frontend/"+taskName+"/assets")
}