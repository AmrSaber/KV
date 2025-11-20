package common

import (
	"database/sql"
	"os"
	"path"

	gap "github.com/muesli/go-app-paths"
	_ "modernc.org/sqlite"
)

var GlobalTx *sql.Tx

var migrations = []string{
	`PRAGMA journal_mode=WAL`,
	`PRAGMA busy_timeout = 5000`,
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
}

func StartGlobalTransaction() {
	if GlobalTx != nil {
		panic("Global transaction already started")
	}

	dbPath := getDBPath()
	_ = os.MkdirAll(path.Dir(dbPath), os.ModeDir|os.ModePerm)

	db, err := sql.Open("sqlite", dbPath)
	FailOn(err)

	for _, query := range migrations {
		_, err := db.Exec(query)
		FailOn(err)
	}

	GlobalTx, err = db.Begin()
	FailOn(err)
}

func ClearDB() {
	dbPath := path.Dir(getDBPath())
	err := os.RemoveAll(dbPath)
	FailOn(err)
}

func getDBPath() string {
	scope := gap.NewScope(gap.User, "kv")

	dbPath, err := scope.DataPath("kv.db")
	FailOn(err)

	return dbPath
}
