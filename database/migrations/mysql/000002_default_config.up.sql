-- +migrate Up
INSERT INTO preservation_configs (
    name, description,
    assign_uuids_to_directories, examine_contents, generate_transfer_structure_report,
    document_empty_directories, extract_packages, delete_packages_after_extraction,
    identify_transfer, identify_submission_and_metadata, identify_before_normalization,
    normalize, transcribe_files, perform_policy_checks_on_originals,
    perform_policy_checks_on_preservation_derivatives, perform_policy_checks_on_access_derivatives,
    thumbnail_mode, aip_compression_level, aip_compression_algorithm
) VALUES (
    'Default Configuration', 'Default preservation configuration for your one-click preservation',
    true, false, true,
    true, true, false,
    true, true, true,
    true, true, true,
    true, true,
    1, 1, 5
); 