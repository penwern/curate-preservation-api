package database

import (
	"database/sql"
	"errors"

	"github.com/penwern/curate-preservation-api/models"
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

var (
	// ErrNotFound is returned when a preservation config is not found in the database
	ErrNotFound = errors.New("preservation config not found")
)

// CreateConfig creates a new preservation configuration in the database
func (d *Database) CreateConfig(config *models.PreservationConfig) error {
	logger.Debug("Creating new preservation config: %s", config.Name)

	query := `
	INSERT INTO preservation_configs (
		name, description, 
		assign_uuids_to_directories,
		examine_contents,
		generate_transfer_structure_report,
		document_empty_directories,
		extract_packages,
		delete_packages_after_extraction,
		identify_transfer,
		identify_submission_and_metadata,
		identify_before_normalization,
		normalize,
		transcribe_files,
		perform_policy_checks_on_originals,
		perform_policy_checks_on_preservation_derivatives,
		perform_policy_checks_on_access_derivatives,
		thumbnail_mode,
		aip_compression_level,
		aip_compression_algorithm,
		compress_aip
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := d.db.Exec(
		query,
		config.Name,
		config.Description,
		config.A3MConfig.AssignUuidsToDirectories,
		config.A3MConfig.ExamineContents,
		config.A3MConfig.GenerateTransferStructureReport,
		config.A3MConfig.DocumentEmptyDirectories,
		config.A3MConfig.ExtractPackages,
		config.A3MConfig.DeletePackagesAfterExtraction,
		config.A3MConfig.IdentifyTransfer,
		config.A3MConfig.IdentifySubmissionAndMetadata,
		config.A3MConfig.IdentifyBeforeNormalization,
		config.A3MConfig.Normalize,
		config.A3MConfig.TranscribeFiles,
		config.A3MConfig.PerformPolicyChecksOnOriginals,
		config.A3MConfig.PerformPolicyChecksOnPreservationDerivatives,
		config.A3MConfig.PerformPolicyChecksOnAccessDerivatives,
		config.A3MConfig.ThumbnailMode,
		config.A3MConfig.AipCompressionLevel,
		config.A3MConfig.AipCompressionAlgorithm,
		config.CompressAIP,
	)

	if err != nil {
		logger.Error("Failed to create preservation config '%s': %v", config.Name, err)
		return err
	}

	// Get the auto-generated ID and assign it to the config
	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("Failed to get last insert ID for config '%s': %v", config.Name, err)
		return err
	}
	config.ID = id

	logger.Debug("Successfully created preservation config '%s' with ID: %d", config.Name, config.ID)
	return nil
}

// GetConfig retrieves a preservation configuration by ID
func (d *Database) GetConfig(id int64) (*models.PreservationConfig, error) {
	logger.Debug("Fetching preservation config with ID: %d", id)

	query := `
	SELECT 
		id, name, description, 
		assign_uuids_to_directories,
		examine_contents,
		generate_transfer_structure_report,
		document_empty_directories,
		extract_packages,
		delete_packages_after_extraction,
		identify_transfer,
		identify_submission_and_metadata,
		identify_before_normalization,
		normalize,
		transcribe_files,
		perform_policy_checks_on_originals,
		perform_policy_checks_on_preservation_derivatives,
		perform_policy_checks_on_access_derivatives,
		thumbnail_mode,
		aip_compression_level,
		aip_compression_algorithm,
		compress_aip,
		created_at,
		updated_at
	FROM preservation_configs
	WHERE id = ?`

	var config models.PreservationConfig
	err := d.db.QueryRow(query, id).Scan(
		&config.ID,
		&config.Name,
		&config.Description,
		&config.A3MConfig.AssignUuidsToDirectories,
		&config.A3MConfig.ExamineContents,
		&config.A3MConfig.GenerateTransferStructureReport,
		&config.A3MConfig.DocumentEmptyDirectories,
		&config.A3MConfig.ExtractPackages,
		&config.A3MConfig.DeletePackagesAfterExtraction,
		&config.A3MConfig.IdentifyTransfer,
		&config.A3MConfig.IdentifySubmissionAndMetadata,
		&config.A3MConfig.IdentifyBeforeNormalization,
		&config.A3MConfig.Normalize,
		&config.A3MConfig.TranscribeFiles,
		&config.A3MConfig.PerformPolicyChecksOnOriginals,
		&config.A3MConfig.PerformPolicyChecksOnPreservationDerivatives,
		&config.A3MConfig.PerformPolicyChecksOnAccessDerivatives,
		&config.A3MConfig.ThumbnailMode,
		&config.A3MConfig.AipCompressionLevel,
		&config.A3MConfig.AipCompressionAlgorithm,
		&config.CompressAIP,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("Preservation config not found: %d", id)
			return nil, ErrNotFound
		}
		logger.Error("Failed to fetch preservation config %d: %v", id, err)
		return nil, err
	}

	logger.Debug("Successfully fetched preservation config: %s (ID: %d)", config.Name, config.ID)
	return &config, nil
}

// ListConfigs retrieves all preservation configurations
func (d *Database) ListConfigs() ([]*models.PreservationConfig, error) {
	logger.Debug("Fetching all preservation configs")

	query := `
	SELECT 
		id, name, description, 
		assign_uuids_to_directories,
		examine_contents,
		generate_transfer_structure_report,
		document_empty_directories,
		extract_packages,
		delete_packages_after_extraction,
		identify_transfer,
		identify_submission_and_metadata,
		identify_before_normalization,
		normalize,
		transcribe_files,
		perform_policy_checks_on_originals,
		perform_policy_checks_on_preservation_derivatives,
		perform_policy_checks_on_access_derivatives,
		thumbnail_mode,
		aip_compression_level,
		aip_compression_algorithm,
		compress_aip,
		created_at,
		updated_at
	FROM preservation_configs
	ORDER BY id`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows: %v", err)
		}
	}()

	var configs []*models.PreservationConfig
	for rows.Next() {
		var config models.PreservationConfig
		err := rows.Scan(
			&config.ID,
			&config.Name,
			&config.Description,
			&config.A3MConfig.AssignUuidsToDirectories,
			&config.A3MConfig.ExamineContents,
			&config.A3MConfig.GenerateTransferStructureReport,
			&config.A3MConfig.DocumentEmptyDirectories,
			&config.A3MConfig.ExtractPackages,
			&config.A3MConfig.DeletePackagesAfterExtraction,
			&config.A3MConfig.IdentifyTransfer,
			&config.A3MConfig.IdentifySubmissionAndMetadata,
			&config.A3MConfig.IdentifyBeforeNormalization,
			&config.A3MConfig.Normalize,
			&config.A3MConfig.TranscribeFiles,
			&config.A3MConfig.PerformPolicyChecksOnOriginals,
			&config.A3MConfig.PerformPolicyChecksOnPreservationDerivatives,
			&config.A3MConfig.PerformPolicyChecksOnAccessDerivatives,
			&config.A3MConfig.ThumbnailMode,
			&config.A3MConfig.AipCompressionLevel,
			&config.A3MConfig.AipCompressionAlgorithm,
			&config.CompressAIP,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan preservation config row: %v", err)
			return nil, err
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over preservation config rows: %v", err)
		return nil, err
	}

	logger.Debug("Successfully fetched %d preservation configs", len(configs))
	return configs, nil
}

// UpdateConfig updates an existing preservation configuration
func (d *Database) UpdateConfig(config *models.PreservationConfig) error {
	// First check if the config exists
	_, err := d.GetConfig(config.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	query := `
	UPDATE preservation_configs SET
		name = ?,
		description = ?,
		assign_uuids_to_directories = ?,
		examine_contents = ?,
		generate_transfer_structure_report = ?,
		document_empty_directories = ?,
		extract_packages = ?,
		delete_packages_after_extraction = ?,
		identify_transfer = ?,
		identify_submission_and_metadata = ?,
		identify_before_normalization = ?,
		normalize = ?,
		transcribe_files = ?,
		perform_policy_checks_on_originals = ?,
		perform_policy_checks_on_preservation_derivatives = ?,
		perform_policy_checks_on_access_derivatives = ?,
		thumbnail_mode = ?,
		aip_compression_level = ?,
		aip_compression_algorithm = ?,
		compress_aip = ?
	WHERE id = ?`

	_, err = d.db.Exec(
		query,
		config.Name,
		config.Description,
		config.A3MConfig.AssignUuidsToDirectories,
		config.A3MConfig.ExamineContents,
		config.A3MConfig.GenerateTransferStructureReport,
		config.A3MConfig.DocumentEmptyDirectories,
		config.A3MConfig.ExtractPackages,
		config.A3MConfig.DeletePackagesAfterExtraction,
		config.A3MConfig.IdentifyTransfer,
		config.A3MConfig.IdentifySubmissionAndMetadata,
		config.A3MConfig.IdentifyBeforeNormalization,
		config.A3MConfig.Normalize,
		config.A3MConfig.TranscribeFiles,
		config.A3MConfig.PerformPolicyChecksOnOriginals,
		config.A3MConfig.PerformPolicyChecksOnPreservationDerivatives,
		config.A3MConfig.PerformPolicyChecksOnAccessDerivatives,
		config.A3MConfig.ThumbnailMode,
		config.A3MConfig.AipCompressionLevel,
		config.A3MConfig.AipCompressionAlgorithm,
		config.CompressAIP,
		config.ID,
	)

	return err
}

// DeleteConfig deletes a preservation configuration by ID
func (d *Database) DeleteConfig(id int64) error {
	// Check if the config exists
	_, err := d.GetConfig(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Delete the config
	query := `DELETE FROM preservation_configs WHERE id = ?`
	_, err = d.db.Exec(query, id)
	return err
}
