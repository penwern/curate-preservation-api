-- +migrate Down
DELETE FROM preservation_configs WHERE name = 'Default Configuration'; 