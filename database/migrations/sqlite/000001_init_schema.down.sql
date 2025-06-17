-- +migrate Down
DROP TRIGGER IF EXISTS update_preservation_configs_updated_at;
DROP TABLE IF EXISTS preservation_configs; 