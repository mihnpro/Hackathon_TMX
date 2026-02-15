// internal/services/ml_integration.go
package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"github.com/mihnpro/Hackathon_TMX/internal/domain/ml"
)

// MLIntegrationService - сервис для интеграции с ML
type MLIntegrationService struct {
	mlServiceURL string
	httpClient   *http.Client
	maxItems     int
}

// NewMLIntegrationService создает новый сервис интеграции
func NewMLIntegrationService(mlServiceURL string) *MLIntegrationService {
	return &MLIntegrationService{
		mlServiceURL: mlServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxItems: 1000,
	}
}

// Predict выполняет предсказание для одного элемента
func (s *MLIntegrationService) Predict(input *ml.WheelInput) (float64, error) {
	// Валидация
	if err := validateInput(input); err != nil {
		return 0, fmt.Errorf("validation error: %w", err)
	}

	// Создаем запрос
	req := ml.BatchPredictionRequest{
		Items: []ml.WheelInput{*input},
	}

	// Отправляем в ML сервис
	resp, err := s.sendRequest(req)
	if err != nil {
		return 0, err
	}

	if len(resp.Predictions) == 0 {
		return 0, ml.ErrPredictionFailed
	}

	return resp.Predictions[0], nil
}

// PredictBatch выполняет пакетное предсказание
func (s *MLIntegrationService) PredictBatch(inputs []ml.WheelInput) (*ml.BatchPredictionResponse, error) {
	// Проверка лимита
	if len(inputs) > s.maxItems {
		return nil, ml.ErrTooManyItems
	}

	// Валидация всех элементов
	for i, input := range inputs {
		if err := validateInput(&input); err != nil {
			return nil, fmt.Errorf("item %d validation error: %w", i, err)
		}
	}

	// Создаем запрос
	req := ml.BatchPredictionRequest{
		Items: inputs,
	}

	// Отправляем в ML сервис
	return s.sendRequest(req)
}

// PredictFromFile выполняет предсказание из JSON файла
func (s *MLIntegrationService) PredictFromFile(content []byte) (*ml.BatchPredictionResponse, error) {
	if len(content) == 0 {
		return nil, ml.ErrEmptyFile
	}

	// Парсим JSON
	var inputs []ml.WheelInput
	if err := json.Unmarshal(content, &inputs); err != nil {
		// Пробуем как JSONL (каждая строка - отдельный JSON)
		return s.parseJSONL(content)
	}

	return s.PredictBatch(inputs)
}

// parseJSONL парсит JSON Lines формат
func (s *MLIntegrationService) parseJSONL(content []byte) (*ml.BatchPredictionResponse, error) {
	lines := bytes.Split(content, []byte("\n"))
	inputs := make([]ml.WheelInput, 0, len(lines))

	for i, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		
		var input ml.WheelInput
		if err := json.Unmarshal(line, &input); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", i+1, err)
		}
		inputs = append(inputs, input)
	}

	if len(inputs) == 0 {
		return nil, ml.ErrEmptyFile
	}

	return s.PredictBatch(inputs)
}

// sendRequest отправляет запрос в ML сервис
func (s *MLIntegrationService) sendRequest(req ml.BatchPredictionRequest) (*ml.BatchPredictionResponse, error) {
	// Сериализуем запрос
	jsonData, err := json.Marshal(req.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем HTTP запрос
	httpReq, err := http.NewRequest(
		http.MethodPost,
		s.mlServiceURL+"/predict",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ML service request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Проверяем статус
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", 
			httpResp.StatusCode, string(body))
	}

	// Декодируем ответ
	var predictions []float64
	if err := json.NewDecoder(httpResp.Body).Decode(&predictions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Формируем ответ
	response := &ml.BatchPredictionResponse{
		Predictions: predictions,
		Count:       len(predictions),
		ProcessedAt: time.Now(),
	}

	return response, nil
}

// HealthCheck проверяет доступность ML сервиса
func (s *MLIntegrationService) HealthCheck() error {
	resp, err := s.httpClient.Get(s.mlServiceURL + "/health")
	if err != nil {
		return ml.ErrMLServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ml.ErrMLServiceUnavailable
	}

	return nil
}

// GetModelInfo получает информацию о модели
func (s *MLIntegrationService) GetModelInfo() (map[string]interface{}, error) {
	resp, err := s.httpClient.Get(s.mlServiceURL + "/info")
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode model info: %w", err)
	}

	return info, nil
}

// validateInput валидирует входные данные
func validateInput(input *ml.WheelInput) error {
	if input.LocomotiveSeries == "" {
		return ml.ErrEmptyLocomotiveSeries
	}
	if input.LocomotiveNumber <= 0 {
		return ml.ErrInvalidLocomotiveNumber
	}
	if input.Depo == "" {
		return ml.ErrEmptyDepo
	}
	if input.SteelNum == "" {
		return ml.ErrEmptySteelNum
	}
	if input.MileageStart < 0 {
		return ml.ErrInvalidMileage
	}
	return nil
}