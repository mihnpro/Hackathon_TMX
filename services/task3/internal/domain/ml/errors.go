// internal/domain/ml/errors.go
package ml

import "errors"

var (
	ErrEmptyLocomotiveSeries   = errors.New("locomotive series cannot be empty")
	ErrInvalidLocomotiveNumber = errors.New("locomotive number must be positive")
	ErrEmptyDepo               = errors.New("depo cannot be empty")
	ErrEmptySteelNum           = errors.New("steel number cannot be empty")
	ErrInvalidMileage          = errors.New("mileage must be non-negative")
	ErrTooManyItems            = errors.New("too many items (max 1000)")
	ErrEmptyFile               = errors.New("file is empty")
	ErrInvalidJSON             = errors.New("invalid JSON format")
	ErrMLServiceUnavailable    = errors.New("ML service is unavailable")
	ErrPredictionFailed        = errors.New("prediction failed")
)