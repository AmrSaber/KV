package common

import (
	"database/sql"
)

const (
	MetadataKeyMigrationIndex       = "migration_index"
	MetadataKeyDBMigrated           = "kv_db_migrated"
	MetadataKeyKeysWarningDisplayed = "keys_warning_displayed"
)

var migrations = []string{
	`
	CREATE TABLE IF NOT EXISTS store (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		is_locked INTEGER NOT NULL,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_latest INTEGER NOT NULL DEFAULT 1,
		expires_at DATETIME DEFAULT NULL
	);
	`,

	// Ensure only one latest record per key
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_store_unique_latest_key ON store(key) WHERE is_latest = 1;`,

	// Index for listing operations with prefix matching
	`CREATE INDEX IF NOT EXISTS idx_store_latest_key_value ON store(key, is_latest, value);`,

	// Index for TTL cleanup queries
	`CREATE INDEX IF NOT EXISTS idx_store_latest_expire ON store(is_latest, expires_at);`,

	// Index for history queries ordered by timestamp
	`CREATE INDEX IF NOT EXISTS idx_store_key_timestamp ON store(key, timestamp);`,

	// Index for history queries ordered by id (more efficient than timestamp)
	`CREATE INDEX IF NOT EXISTS idx_store_key_id ON store(key, id);`,

	// Add is_hidden column (replaces previous hack)
	`ALTER TABLE store ADD COLUMN is_hidden INTEGER NOT NULL DEFAULT 0`,
}

func runMigrations(tx *sql.Tx) {
	// Create metadata table if it's not created
	ensureMetadataTable(tx)

	// Get migration index
	currentIndex, found := getMigrationIndex(tx)

	if !found {
		// No migration index - run all migrations
		executeMigrations(tx, 0)
		updateMigrationIndex(tx)
		return
	}

	// Migration index exists - run only new migrations
	if currentIndex < len(migrations)-1 {
		executeMigrations(tx, currentIndex+1)
		updateMigrationIndex(tx)
	}
}

func ensureMetadataTable(tx *sql.Tx) {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS _kv_metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)

	FailOn(err)
}

func executeMigrations(tx *sql.Tx, start int) {
	for _, query := range migrations[start:] {
		_, err := tx.Exec(query)
		FailOn(err)
	}
}

func ReadMetadata[T any](tx *sql.Tx, key string) (T, error) {
	var value T
	err := tx.QueryRow("SELECT value FROM _kv_metadata WHERE key = ?", key).Scan(&value)
	if err != nil {
		return value, err
	}

	return value, err
}

func WriteMetadata(tx *sql.Tx, key string, value any) error {
	_, err := tx.Exec(`
		INSERT INTO _kv_metadata (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)

	return err
}

func getMigrationIndex(tx *sql.Tx) (int, bool) {
	index, err := ReadMetadata[int](tx, MetadataKeyMigrationIndex)
	if err == sql.ErrNoRows {
		return -1, false
	}

	FailOn(err)
	return index, true
}

func updateMigrationIndex(tx *sql.Tx) {
	err := WriteMetadata(tx, MetadataKeyMigrationIndex, len(migrations)-1)
	FailOn(err)
}
