package common

import (
	"database/sql"
	"strconv"
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
	latestMigration := len(migrations) - 1

	// Ensure metadata table exists
	ensureMetadataTable(tx)

	// Get migration index
	currentIndex, found := getMigrationIndex(tx)

	if !found {
		// Case 1: No migration index - run all migrations
		executeMigrations(tx, 0, latestMigration)
		setMigrationIndex(tx, latestMigration)
		return
	}

	// Case 2: Migration index exists - run only new migrations
	if currentIndex < latestMigration {
		executeMigrations(tx, currentIndex+1, latestMigration)
		setMigrationIndex(tx, latestMigration)
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

func executeMigrations(tx *sql.Tx, start, end int) {
	for i := start; i <= end; i++ {
		_, err := tx.Exec(migrations[i])
		FailOn(err)
	}
}

func getMigrationIndex(tx *sql.Tx) (int, bool) {
	var value string
	err := tx.QueryRow(`
		SELECT value FROM _kv_metadata WHERE key = 'migration_index'
	`).Scan(&value)

	if err == sql.ErrNoRows {
		return -1, false
	}

	FailOn(err)

	index, err := strconv.Atoi(value)
	FailOn(err)

	return index, true
}

func setMigrationIndex(tx *sql.Tx, index int) {
	_, err := tx.Exec(`
		INSERT INTO _kv_metadata (key, value)
		VALUES ('migration_index', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, strconv.Itoa(index))
	FailOn(err)
}
