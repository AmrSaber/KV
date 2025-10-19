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
	`
	CREATE TABLE IF NOT EXISTS store (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_latest INTEGER NOT NULL DEFAULT 1,
		expires_at DATETIME DEFAULT NULL
	);
	`,
	`CREATE INDEX IF NOT EXISTS idx_store_latest_key_value ON store(key, is_latest, value);`,
	`CREATE INDEX IF NOT EXISTS idx_store_latest_expire ON store(is_latest, expires_at);`,
}

func StartGlobalTransaction() {
	if GlobalTx != nil {
		panic("Global transaction already started")
	}

	dbPath := getDBPath()
	os.MkdirAll(path.Dir(dbPath), os.ModeDir|os.ModePerm)

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
