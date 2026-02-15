package ml

import (
	"time"
)

// WheelInput - входные данные для ML модели
type WheelInput struct {
	LocomotiveSeries string  `json:"locomotive_series"`
	LocomotiveNumber int     `json:"locomotive_number"`
	Depo             string  `json:"depo"`
	SteelNum         string  `json:"steel_num"`
	MileageStart     float64 `json:"mileage_start"`
}

// PredictionResult - результат предсказания
type PredictionResult struct {
	Input      WheelInput `json:"input"`
	Prediction float64    `json:"prediction"`
	Error      string     `json:"error,omitempty"`
}

// BatchPredictionRequest - запрос на пакетное предсказание
type BatchPredictionRequest struct {
	Items    []WheelInput `json:"items"`
	Filename string       `json:"filename,omitempty"`
}

// BatchPredictionResponse - ответ на пакетное предсказание
type BatchPredictionResponse struct {
	Predictions []float64 `json:"predictions"`
	Count       int       `json:"count"`
	ProcessedAt time.Time `json:"processed_at"`
}

// JobStatus - статус задачи (для асинхронной обработки)
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// PredictionJob - задача на предсказание
type PredictionJob struct {
	ID          string     `json:"id"`
	Filename    string     `json:"filename"`
	Status      JobStatus  `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	RecordCount int        `json:"record_count"`
	Result      []float64  `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	UserID      string     `json:"user_id,omitempty"`
}

// Validation методы
func (w *WheelInput) Validate() error {
	if w.LocomotiveSeries == "" {
		return ErrEmptyLocomotiveSeries
	}
	if w.LocomotiveNumber <= 0 {
		return ErrInvalidLocomotiveNumber
	}
	if w.Depo == "" {
		return ErrEmptyDepo
	}
	if w.SteelNum == "" {
		return ErrEmptySteelNum
	}
	if w.MileageStart < 0 {
		return ErrInvalidMileage
	}
	return nil
}
