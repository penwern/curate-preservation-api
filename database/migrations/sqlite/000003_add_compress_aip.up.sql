-- +migrate Up
ALTER TABLE preservation_configs
ADD COLUMN compress_aip BOOLEAN DEFAULT 0; 