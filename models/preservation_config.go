package models

import (
	"time"
)

// PreservationConfig represents a preservation configuration stored in the database
type PreservationConfig struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	A3MConfig   A3MProcessingConfig `json:"a3m_config"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// NewPreservationConfig creates a new preservation configuration with default values
func NewPreservationConfig(name, description string) *PreservationConfig {
	return &PreservationConfig{
		Name:        name,
		Description: description,
		A3MConfig:   NewA3MProcessingConfig(),
	}
}

// // ToA3MConfig converts the preservation configuration to an A3M preservation config
// func (p *PreservationConfig) ToA3MConfig() *A3MProcessingConfig {
// 	return &A3MProcessingConfig{
// 		AssignUuidsToDirectories:                     p.A3MConfig.AssignUuidsToDirectories,
// 		ExamineContents:                              p.A3MConfig.ExamineContents,
// 		GenerateTransferStructureReport:              p.A3MConfig.GenerateTransferStructureReport,
// 		DocumentEmptyDirectories:                     p.A3MConfig.DocumentEmptyDirectories,
// 		ExtractPackages:                              p.A3MConfig.ExtractPackages,
// 		DeletePackagesAfterExtraction:                p.A3MConfig.DeletePackagesAfterExtraction,
// 		IdentifyTransfer:                             p.A3MConfig.IdentifyTransfer,
// 		IdentifySubmissionAndMetadata:                p.A3MConfig.IdentifySubmissionAndMetadata,
// 		IdentifyBeforeNormalization:                  p.A3MConfig.IdentifyBeforeNormalization,
// 		Normalize:                                    p.A3MConfig.Normalize,
// 		TranscribeFiles:                              p.A3MConfig.TranscribeFiles,
// 		PerformPolicyChecksOnOriginals:               p.A3MConfig.PerformPolicyChecksOnOriginals,
// 		PerformPolicyChecksOnPreservationDerivatives: p.A3MConfig.PerformPolicyChecksOnPreservationDerivatives,
// 		PerformPolicyChecksOnAccessDerivatives:       p.A3MConfig.PerformPolicyChecksOnAccessDerivatives,
// 		ThumbnailMode:                                p.A3MConfig.ThumbnailMode,
// 		AipCompressionLevel:                          p.A3MConfig.AipCompressionLevel,
// 		AipCompressionAlgorithm:                      p.A3MConfig.AipCompressionAlgorithm,
// 	}
// }
