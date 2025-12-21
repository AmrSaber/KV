package common

import (
	"database/sql"
	"io"
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

func CloseDB() {
	if db != nil {
		_ = db.Close()
		db = nil
	}
}

func ClearDB() {
	dbPath := path.Dir(GetDBPath())
	err := os.RemoveAll(dbPath)
	FailOn(err)
}

func GetDB() (*sql.DB, error) {
	var err error

	if db == nil {
		db, err = openDB()
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func openDB() (*sql.DB, error) {
	dbPath := GetDBPath()
	_ = os.MkdirAll(path.Dir(dbPath), os.ModeDir|os.ModePerm)

	db, err := sql.Open("sqlite", dbPath+"?_txlock=immediate")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	// Pragmas cannot run in transactions
	for _, pragma := range pragmas {
		err := runPragma(db, pragma)
		if err != nil {
			return nil, err
		}
	}

	// Migrations transaction
	tx, err := BeginTarnsaction(db)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback() }()

	// Run new migration system
	runMigrations(tx)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetDBPath() string {
	scope := gap.NewScope(gap.User, "kv")

	dbPath, err := scope.DataPath("kv.db")
	FailOn(err)

	return dbPath
}

func ValidateSqliteFile(path string) error {
	testDB, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		return err
	}

	defer func() { _ = testDB.Close() }()

	// Try to query to ensure it's actually valid
	return testDB.Ping()
}

func GetDefaultBackupPath() string {
	destPath := GetDBPath()
	return destPath + ".backup"
}

func BackupDB(writer io.Writer) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// Vacuum current database to commit all WAL changes to main file
	_, err = db.Exec("VACUUM")
	if err != nil {
		return err
	}

	// Close database connection
	CloseDB()

	dbFile, err := os.Open(GetDBPath())
	if err != nil {
		return err
	}

	defer func() { _ = dbFile.Close() }()

	// Backup current database if it exists
	_, err = io.Copy(writer, dbFile)
	if err != nil {
		return err
	}

	return nil
}
