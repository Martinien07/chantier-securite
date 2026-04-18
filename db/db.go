package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
)

//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate

// Open ouvre une connexion MySQL à partir du DSN fourni.
// Format : user:password@tcp(host:3306)/dbname?parseTime=true
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

// RunMigrations détecte si la base est déjà initialisée et skip les migrations.
func RunMigrations(db *sql.DB) error {
	var tableName string
	err := db.QueryRow(
		"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'cameras' LIMIT 1",
	).Scan(&tableName)

	if err == nil && tableName == "cameras" {
		slog.Info("db: base existante détectée, migrations ignorées")
		return nil
	}

	slog.Info("db: base vide, création de la table de suivi")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		migration_number INT PRIMARY KEY,
		migration_name   VARCHAR(255) NOT NULL,
		executed_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}
	return nil
}
