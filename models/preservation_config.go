package models

import (
	"time"
)

// PreservationConfig represents a preservation configuration stored in the database
type PreservationConfig struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	CompressAIP bool                `json:"compress_aip"`
	A3MConfig   A3MProcessingConfig `json:"a3m_config"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// NewPreservationConfig creates a new preservation configuration with default values
func NewPreservationConfig(name, description string) *PreservationConfig {
	return &PreservationConfig{
		Name:        name,
		Description: description,
		CompressAIP: false,
		A3MConfig:   NewA3MProcessingConfig(),
	}
}
