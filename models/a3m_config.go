// Package models defines data structures for the preservation API.
package models

import (
	transferservice "github.com/penwern/curate-preservation-core/common/proto/a3m/gen/go/a3m/api/transferservice/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
)

// A3MProcessingConfig is a thin wrapper around the generated ProcessingConfig
type A3MProcessingConfig transferservice.ProcessingConfig

// MarshalJSON emits the proto with our options applied
// This is used to ensure that the A3MConfig is marshaled correctly
// and that the 'omitempty' json directives are ignored
// This is called automatically when the A3MProcessingConfig is marshaled to JSON
func (c *A3MProcessingConfig) MarshalJSON() ([]byte, error) {
	a3mJSON, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseEnumNumbers:  true,
	}.Marshal((*transferservice.ProcessingConfig)(c))
	if err != nil {
		return nil, err
	}
	return a3mJSON, nil
}

// UnmarshalJSON parses the JSON data and populates the A3MProcessingConfig
// This is called automatically when the A3MProcessingConfig is unmarshaled from JSON
func (c *A3MProcessingConfig) UnmarshalJSON(data []byte) error {
	var proto transferservice.ProcessingConfig
	err := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(data, &proto)
	if err != nil {
		return err
	}
	*c = *(*A3MProcessingConfig)(&proto)
	return nil
}

// NewA3MProcessingConfig creates a new A3MProcessingConfig with default values
func NewA3MProcessingConfig() A3MProcessingConfig {
	return A3MProcessingConfig{
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
		AipCompressionAlgorithm:                      transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_S7_BZIP2,
	}
}
