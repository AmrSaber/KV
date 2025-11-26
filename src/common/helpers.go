package common

import (
	"database/sql"
	"strings"
	"time"
)

func runPragma(db *sql.DB, pragma string) error {
	return retryOnBusy(func() error {
		_, err := db.Exec(pragma)
		return err
	})
}

func beginTarnsaction(db *sql.DB) (*sql.Tx, error) {
	var tx *sql.Tx

	err := retryOnBusy(func() error {
		var err error
		tx, err = db.Begin()
		return err
	})

	return tx, err
}

func retryOnBusy(operation func() error) error {
	timeWaited := 0 * time.Second
	retryDelay := 5 * time.Millisecond

	for timeWaited < 3*time.Second {
		err := operation()
		if err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "database is locked") && !strings.Contains(err.Error(), "SQLITE_BUSY") {
			return err
		}

		time.Sleep(retryDelay)
		timeWaited += retryDelay
	}

	return operation()
}
