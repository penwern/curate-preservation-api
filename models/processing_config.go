package models

import (
	"time"

	transferservice "github.com/penwern/curate-preservation-core/common/proto/a3m/gen/go/a3m/api/transferservice/v1beta1"
)

// PreservationConfig represents a preservation configuration stored in the database
type PreservationConfig struct {
	ID          int64                            `json:"id"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	A3MConfig   transferservice.ProcessingConfig `json:"a3m_config"`
	CreatedAt   time.Time                        `json:"created_at"`
	UpdatedAt   time.Time                        `json:"updated_at"`
}

// NewPreservationConfig creates a new preservation configuration with default values
func NewPreservationConfig(name, description string) *PreservationConfig {
	return &PreservationConfig{
		Name:        name,
		Description: description,
		A3MConfig: transferservice.ProcessingConfig{
			AssignUuidsToDirectories:                     true,
			ExamineContents:                              false,
			GenerateTransferStructureReport:              true,
			DocumentEmptyDirectories:                     true,
			ExtractPackages:                              true,
			DeletePackagesAfterExtraction:                false,
			IdentifyTransfer:                             true,
			IdentifySubmissionAndMetadata:                true,
			IdentifyBeforeNormalization:                  true,
			Normalize:                                    true,
			TranscribeFiles:                              true,
			PerformPolicyChecksOnOriginals:               true,
			PerformPolicyChecksOnPreservationDerivatives: true,
			PerformPolicyChecksOnAccessDerivatives:       true,
			ThumbnailMode:                                transferservice.ProcessingConfig_THUMBNAIL_MODE_GENERATE,
			AipCompressionLevel:                          1,
			AipCompressionAlgorithm:                      transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_S7_COPY,
		},
	}
}

// ToA3MConfig converts the preservation configuration to an A3M preservation config
func (p *PreservationConfig) ToA3MConfig() *transferservice.ProcessingConfig {
	return &transferservice.ProcessingConfig{
		AssignUuidsToDirectories:                     p.A3MConfig.AssignUuidsToDirectories,
		ExamineContents:                              p.A3MConfig.ExamineContents,
		GenerateTransferStructureReport:              p.A3MConfig.GenerateTransferStructureReport,
		DocumentEmptyDirectories:                     p.A3MConfig.DocumentEmptyDirectories,
		ExtractPackages:                              p.A3MConfig.ExtractPackages,
		DeletePackagesAfterExtraction:                p.A3MConfig.DeletePackagesAfterExtraction,
		IdentifyTransfer:                             p.A3MConfig.IdentifyTransfer,
		IdentifySubmissionAndMetadata:                p.A3MConfig.IdentifySubmissionAndMetadata,
		IdentifyBeforeNormalization:                  p.A3MConfig.IdentifyBeforeNormalization,
		Normalize:                                    p.A3MConfig.Normalize,
		TranscribeFiles:                              p.A3MConfig.TranscribeFiles,
		PerformPolicyChecksOnOriginals:               p.A3MConfig.PerformPolicyChecksOnOriginals,
		PerformPolicyChecksOnPreservationDerivatives: p.A3MConfig.PerformPolicyChecksOnPreservationDerivatives,
		PerformPolicyChecksOnAccessDerivatives:       p.A3MConfig.PerformPolicyChecksOnAccessDerivatives,
		ThumbnailMode:                                p.A3MConfig.ThumbnailMode,
		AipCompressionLevel:                          p.A3MConfig.AipCompressionLevel,
		AipCompressionAlgorithm:                      p.A3MConfig.AipCompressionAlgorithm,
	}
}
