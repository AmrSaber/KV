package common

import (
	"database/sql"
	"os"
	"path"

	gap "github.com/muesli/go-app-paths"
	_ "modernc.org/sqlite"
)

var db *sql.DB

var pragmas = []string{
	`PRAGMA journal_mode = WAL`,
	`PRAGMA busy_timeout = 5000`,
}

func ClearDB() {
	dbPath := path.Dir(getDBPath())
	err := os.RemoveAll(dbPath)
	FailOn(err)
}

func GetDB() *sql.DB {
	if db == nil {
		db = openDB()
	}

	return db
}

func openDB() *sql.DB {
	dbPath := getDBPath()
	_ = os.MkdirAll(path.Dir(dbPath), os.ModeDir|os.ModePerm)

	db, err := sql.Open("sqlite", dbPath+"?_txlock=immediate")
	FailOn(err)

	db.SetMaxOpenConns(1)

	// Pragmas cannot run in transactions
	for _, pragma := range pragmas {
		err := runPragma(db, pragma)
		FailOn(err)
	}

	// Migrations transaction
	tx, err := BeginTarnsaction(db)
	FailOn(err)

	defer func() { _ = tx.Rollback() }()

	// Run new migration system
	runMigrations(tx)

	err = tx.Commit()
	FailOn(err)

	return db
}

func getDBPath() string {
	scope := gap.NewScope(gap.User, "kv")

	dbPath, err := scope.DataPath("kv.db")
	FailOn(err)

	return dbPath
}
