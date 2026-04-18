-- Base schema - Migrations tracking
CREATE TABLE IF NOT EXISTS migrations (
    migration_number INTEGER PRIMARY KEY,
    migration_name   VARCHAR(255) NOT NULL,
    executed_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT IGNORE INTO migrations (migration_number, migration_name) VALUES (001, '001-base');
