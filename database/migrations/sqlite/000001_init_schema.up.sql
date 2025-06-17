-- +migrate Up
CREATE TABLE IF NOT EXISTS preservation_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    assign_uuids_to_directories BOOLEAN DEFAULT TRUE,
    examine_contents BOOLEAN DEFAULT FALSE,
    generate_transfer_structure_report BOOLEAN DEFAULT TRUE,
    document_empty_directories BOOLEAN DEFAULT TRUE,
    extract_packages BOOLEAN DEFAULT TRUE,
    delete_packages_after_extraction BOOLEAN DEFAULT FALSE,
    identify_transfer BOOLEAN DEFAULT TRUE,
    identify_submission_and_metadata BOOLEAN DEFAULT TRUE,
    identify_before_normalization BOOLEAN DEFAULT TRUE,
    normalize BOOLEAN DEFAULT TRUE,
    transcribe_files BOOLEAN DEFAULT TRUE,
    perform_policy_checks_on_originals BOOLEAN DEFAULT TRUE,
    perform_policy_checks_on_preservation_derivatives BOOLEAN DEFAULT TRUE,
    perform_policy_checks_on_access_derivatives BOOLEAN DEFAULT TRUE,
    thumbnail_mode INT DEFAULT 1,
    aip_compression_level INT DEFAULT 1,
    aip_compression_algorithm INT DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS update_preservation_configs_updated_at
AFTER UPDATE ON preservation_configs
BEGIN
    UPDATE preservation_configs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 