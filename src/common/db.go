package common

import (
	"database/sql"
	"io"
	"os"
	"path"

	_ "modernc.org/sqlite"
)

var CachedDBs = make(map[string]*sql.DB)

var pragmas = []string{
	`PRAGMA journal_mode = WAL`,
	`PRAGMA busy_timeout = 5000`,
}

func CloseDBs() {
	for name, db := range CachedDBs {
		if db != nil {
			_ = db.Close()
			delete(CachedDBs, name)
		}
	}
}

func GetDB(name string) (*sql.DB, error) {
	if CachedDBs[name] == nil {
		db, err := openDB(name)
		if err != nil {
			return nil, err
		}

		CachedDBs[name] = db
	}

	return CachedDBs[name], nil
}

func openDB(name string) (*sql.DB, error) {
	dbPath := GetConfig().GetDBPath(name)
	err := os.MkdirAll(path.Dir(dbPath), 0o755)
	FailOn(err)

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
	tx, err := BeginTransaction(db)
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

func ValidateSqliteFile(path string) error {
	testDB, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		return err
	}

	defer func() { _ = testDB.Close() }()

	// Try to query to ensure it's actually valid
	return testDB.Ping()
}

func GetDefaultBackupPath(name string) string {
	return GetConfig().GetDBPath(name) + ".backup"
}

func BackupDB(name string, writer io.Writer) error {
	db, err := GetDB(name)
	if err != nil {
		return err
	}

	// Vacuum current database to commit all WAL changes to main file
	_, err = db.Exec("VACUUM")
	if err != nil {
		return err
	}

	// Close database connection
	CloseDBs()

	dbFile, err := os.Open(GetConfig().GetDBPath(name))
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

func BackupDBInPlace(name string) {
	backupPath := GetDefaultBackupPath(name)
	backupWriter, err := os.Create(backupPath)
	FailOn(err)

	err = BackupDB(name, backupWriter)
	_ = backupWriter.Close()
	FailOn(err)
}
