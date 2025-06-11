package database

import (
	"database/sql"
	"errors"

	"github.com/penwern/curate-preservation-core-api/models"
)

var (
	ErrNotFound = errors.New("preservation config not found")
)

// CreateConfig creates a new preservation configuration in the database
func (d *Database) CreateConfig(config *models.PreservationConfig) error {
	var query string

	if d.dbType == "sqlite3" {
		query = `
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
			aip_compression_algorithm
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	} else {
		query = `
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
			aip_compression_algorithm
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	}

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
	)

	if err != nil {
		return err
	}

	// Get the auto-generated ID and assign it to the config
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	config.ID = id

	return nil
}

// GetConfig retrieves a preservation configuration by ID
func (d *Database) GetConfig(id int64) (*models.PreservationConfig, error) {
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
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &config, nil
}

// ListConfigs retrieves all preservation configurations
func (d *Database) ListConfigs() ([]*models.PreservationConfig, error) {
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
		created_at,
		updated_at
	FROM preservation_configs
	ORDER BY id`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

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
		aip_compression_algorithm = ?
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
