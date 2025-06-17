-- +migrate Down
ALTER TABLE preservation_configs
DROP COLUMN compress_aip; 