// Package services includes models and services that communicate with the database
package services

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
)

// CleanUpDB clears expired values, deletes old history, and prunes old cleared values
func CleanUpDB(tx *sql.Tx) {
	if tx == nil {
		common.RunTx(CleanUpDB)
		return
	}

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
	defer rows.Close()

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		common.FailOn(err)

		SetValue(tx, key, "", nil)
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
					ROW_NUMBER() OVER (PARTITION BY key ORDER BY timestamp DESC) as rn
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
