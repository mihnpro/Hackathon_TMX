// internal/transport/handlers/ml.go
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mihnpro/Hackathon_TMX/internal/domain/ml"
	"github.com/mihnpro/Hackathon_TMX/internal/services"
	"github.com/mihnpro/Hackathon_TMX/internal/transport/models/responses"
)

// MLHandler - обработчик для ML эндпоинтов
type MLHandler struct {
	mlService   *services.MLIntegrationService
	uploadPath  string
	maxFileSize int64
}

// NewMLHandler создает новый обработчик
func NewMLHandler(mlService *services.MLIntegrationService) *MLHandler {
	return &MLHandler{
		mlService:   mlService,
		uploadPath:  "./uploads",
		maxFileSize: 10 << 20, // 10MB
	}
}

func (h *MLHandler) HandlePredictSync(c *gin.Context) {
	// Читаем тело запроса
	var inputs []ml.WheelInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "Invalid JSON format: " + err.Error(),
		})
		return
	}

	// Вызываем ML сервис
	resp, err := h.mlService.PredictBatch(inputs)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ml.ErrTooManyItems {
			status = http.StatusBadRequest
		}
		c.JSON(status, responses.MLErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// ✨ ИСПРАВЛЕНО: отправляем и предсказания, и входные данные
	c.JSON(http.StatusOK, responses.MLPredictionResponse{
		Success:     true,
		Predictions: resp.Predictions,
		Inputs:      inputs, // Передаем исходные данные для таблицы
		Count:       resp.Count,
		ProcessedAt: resp.ProcessedAt,
	})
}

func (h *MLHandler) HandleUploadFile(c *gin.Context) {
	// Получаем файл из формы
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "File required: " + err.Error(),
		})
		return
	}

	// Проверяем размер файла
	if file.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("File too large. Max size: %d bytes", h.maxFileSize),
		})
		return
	}

	// Проверяем расширение
	if !h.isJSONFile(file.Filename) {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "Only JSON files are allowed",
		})
		return
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.MLErrorResponse{
			Success: false,
			Error:   "Failed to open file: " + err.Error(),
		})
		return
	}
	defer src.Close()

	// Читаем содержимое
	content := make([]byte, file.Size)
	if _, err := src.Read(content); err != nil {
		c.JSON(http.StatusInternalServerError, responses.MLErrorResponse{
			Success: false,
			Error:   "Failed to read file: " + err.Error(),
		})
		return
	}

	// Вызываем ML сервис
	resp, err := h.mlService.PredictFromFile(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.MLErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// ✨ Парсим входные данные из файла для ответа
	var inputs []ml.WheelInput
	parseErr := json.Unmarshal(content, &inputs)

	// Если не удалось распарсить как массив, пробуем как JSONL
	if parseErr != nil {
		inputs, _ = h.parseJSONLContent(content)
	}

	// Отправляем ответ с входными данными
	c.JSON(http.StatusOK, responses.MLPredictionResponse{
		Success:     true,
		Message:     fmt.Sprintf("Successfully processed %d records", resp.Count),
		Predictions: resp.Predictions,
		Inputs:      inputs, // Передаем исходные данные из файла
		Count:       resp.Count,
		ProcessedAt: resp.ProcessedAt,
	})
}

// internal/transport/handlers/ml.go

// parseJSONLContent парсит JSONL контент и возвращает входные данные
func (h *MLHandler) parseJSONLContent(content []byte) ([]ml.WheelInput, error) {
	lines := bytes.Split(content, []byte("\n"))
	inputs := make([]ml.WheelInput, 0, len(lines))

	for i, line := range lines {
		trimmedLine := bytes.TrimSpace(line)
		if len(trimmedLine) == 0 {
			continue
		}

		var input ml.WheelInput
		if err := json.Unmarshal(trimmedLine, &input); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", i+1, err)
		}
		inputs = append(inputs, input)
	}

	if len(inputs) == 0 {
		return nil, fmt.Errorf("no valid JSON objects found")
	}

	return inputs, nil
}

// HandleAsyncUpload - асинхронная загрузка файла
// POST /api/ml/upload/async
func (h *MLHandler) HandleAsyncUpload(c *gin.Context) {
	// Получаем файл из формы
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "File required: " + err.Error(),
		})
		return
	}

	// Проверяем размер файла
	if file.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("File too large. Max size: %d bytes", h.maxFileSize),
		})
		return
	}

	// Проверяем расширение
	if !h.isJSONFile(file.Filename) {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "Only JSON files are allowed",
		})
		return
	}

	// Генерируем ID задачи
	jobID := uuid.New().String()

	// Сохраняем файл
	dst := filepath.Join(h.uploadPath, jobID+".json")
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, responses.MLErrorResponse{
			Success: false,
			Error:   "Failed to save file: " + err.Error(),
		})
		return
	}

	// TODO: Отправляем задачу в очередь
	// Здесь должна быть логика отправки в очередь (Redis/Kafka)

	// Отправляем ответ
	c.JSON(http.StatusAccepted, responses.MLUploadResponse{
		Success: true,
		JobID:   jobID,
		Status:  string(ml.JobStatusPending),
		Message: "File accepted for processing",
	})
}

// HandleJobStatus - проверка статуса задачи
// GET /api/ml/jobs/:id
func (h *MLHandler) HandleJobStatus(c *gin.Context) {
	jobID := c.Param("id")

	// TODO: Получаем статус задачи из хранилища
	// Здесь должна быть логика получения статуса из БД/кэша

	// Пример ответа
	c.JSON(http.StatusOK, responses.MLJobStatusResponse{
		ID:          jobID,
		Status:      string(ml.JobStatusCompleted),
		Filename:    "example.json",
		CreatedAt:   c.GetTime("created_at"),
		UpdatedAt:   c.GetTime("updated_at"),
		RecordCount: 100,
		Result:      []float64{0.123, 0.456, 0.789},
	})
}

// HandleHealth - проверка здоровья ML сервиса
// GET /api/ml/health
func (h *MLHandler) HandleHealth(c *gin.Context) {
	err := h.mlService.HealthCheck()

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"service": "ml-integration",
			"status":  "unhealthy",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": "ml-integration",
		"status":  "healthy",
	})
}

// HandleModelInfo - информация о модели
// GET /api/ml/info
func (h *MLHandler) HandleModelInfo(c *gin.Context) {
	info, err := h.mlService.GetModelInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.MLErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// HandleBatchPredict - пакетное предсказание из формы
// POST /api/ml/batch
func (h *MLHandler) HandleBatchPredict(c *gin.Context) {
	// Получаем JSON данные из формы
	jsonData := c.PostForm("data")
	if jsonData == "" {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "data field is required",
		})
		return
	}

	// Парсим JSON
	var inputs []ml.WheelInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, responses.MLErrorResponse{
			Success: false,
			Error:   "Invalid JSON format: " + err.Error(),
		})
		return
	}

	// Вызываем ML сервис
	resp, err := h.mlService.PredictBatch(inputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.MLErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Отправляем ответ
	c.JSON(http.StatusOK, responses.MLPredictionResponse{
		Success:     true,
		Predictions: resp.Predictions,
		Count:       resp.Count,
		ProcessedAt: resp.ProcessedAt,
	})
}

// Вспомогательные методы
func (h *MLHandler) isJSONFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".json" || ext == ".jsonl"
}
