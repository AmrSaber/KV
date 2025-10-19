// Package services includes models and services that communicate with the database
package services

import (
	"github.com/AmrSaber/kv/src/common"
)

// CleanUpDB clears expired values, deletes old history, and prunes old cleared values
func CleanUpDB() {
	clearExpiredValues()
	deleteOldHistory()
	pruneOldClearedValues()
}

func clearExpiredValues() {
	rows, err := common.GlobalTx.Query(`
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

		SetValue(key, "", nil)
	}
}

func deleteOldHistory() {
	config := common.ReadConfig()

	_, err := common.GlobalTx.Exec(`
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

func pruneOldClearedValues() {
	config := common.ReadConfig()

	rows, err := common.GlobalTx.Query(`
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

		PruneKey(key)
	}
}
