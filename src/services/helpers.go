// Package services includes models and services that communicate with the database
package services

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
)

func RunInTransaction(fn func(tx *sql.Tx)) {
	db, err := common.GetDB()
	common.FailOn(err)

	tx, err := common.BeginTarnsaction(db)
	common.FailOn(err)

	defer func() { _ = tx.Rollback() }()

	// Make sure database is cleaned up before running transactions
	cleanUpDB(tx)

	fn(tx)

	err = tx.Commit()
	common.FailOn(err)
}

// cleanUpDB clears expired values, deletes old history, and prunes old cleared values
func cleanUpDB(tx *sql.Tx) {
	clearExpiredValues(tx)
	deleteOldHistory(tx)
	pruneOldClearedValues(tx)
}

func clearExpiredValues(tx *sql.Tx) {
	rows, err := tx.Query(`
		SELECT key
		FROM store
		WHERE
			is_latest = 1 AND
			expires_at IS NOT NULL AND
			datetime(expires_at) < CURRENT_TIMESTAMP
	`)
	common.FailOn(err)
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		common.FailOn(err)

		SetValue(tx, key, "", nil, false)
	}
}

func deleteOldHistory(tx *sql.Tx) {
	config := common.ReadConfig()

	_, err := tx.Exec(`
		DELETE FROM store
		WHERE id IN (
			SELECT id
			FROM (
				SELECT id, 
					ROW_NUMBER() OVER (PARTITION BY key ORDER BY id DESC) as rn
				FROM store
			)
			WHERE rn > ?
		)
		`,
		config.HistoryLength)

	common.FailOn(err)
}

func pruneOldClearedValues(tx *sql.Tx) {
	config := common.ReadConfig()

	rows, err := tx.Query(`
		SELECT key
		FROM store
		WHERE
			is_latest = 1 AND
			value = '' AND
			timestamp < datetime('now', '-' || ? || ' days')
		`,
		config.PruneHistoryAfterDays,
	)
	common.FailOn(err)

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		common.FailOn(err)

		PruneKey(tx, key)
	}
}
