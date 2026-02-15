// internal/transport/models/responses/ml.go
package responses

import (
	"time"
	
	"github.com/mihnpro/Hackathon_TMX/internal/domain/ml"
)

// MLPredictionResponse - ответ для API
type MLPredictionResponse struct {
	Success     bool      `json:"success" example:"true"`
	Message     string    `json:"message,omitempty" example:"Successfully processed 10 records"`
	Error       string    `json:"error,omitempty" example:"invalid input data"`
	Predictions []float64 `json:"predictions,omitempty" example:"0.123,0.456,0.789"`
	Count       int       `json:"count,omitempty" example:"10"`
	ProcessedAt time.Time `json:"processed_at,omitempty" example:"2024-01-01T12:00:00Z"`
}

// MLUploadResponse - ответ на загрузку файла
type MLUploadResponse struct {
	Success bool   `json:"success" example:"true"`
	JobID   string `json:"job_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  string `json:"status,omitempty" example:"pending"`
	Message string `json:"message,omitempty" example:"File accepted for processing"`
	Error   string `json:"error,omitempty" example:"file too large"`
}

// MLJobStatusResponse - ответ со статусом задачи
type MLJobStatusResponse struct {
	ID          string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status      string     `json:"status" example:"completed"`
	Filename    string     `json:"filename" example:"data.json"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T12:00:00Z"`
	CompletedAt *time.Time `json:"completed_at,omitempty" example:"2024-01-01T12:00:00Z"`
	RecordCount int        `json:"record_count" example:"100"`
	Result      []float64  `json:"result,omitempty" example:"0.123,0.456,0.789"`
	Error       string     `json:"error,omitempty" example:"processing failed"`
}

// FromDomain преобразует доменную модель в ответ
func (r *MLJobStatusResponse) FromDomain(job *ml.PredictionJob) {
	r.ID = job.ID
	r.Status = string(job.Status)
	r.Filename = job.Filename
	r.CreatedAt = job.CreatedAt
	r.UpdatedAt = job.UpdatedAt
	r.CompletedAt = job.CompletedAt
	r.RecordCount = job.RecordCount
	r.Result = job.Result
	r.Error = job.Error
}

// MLErrorResponse - ответ с ошибкой
type MLErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid request"`
}