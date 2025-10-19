package services

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
)

func ClearKeyHistory(tx *sql.Tx, key string) {
	if tx == nil {
		common.RunTx(func(tx *sql.Tx) {
			ClearKeyHistory(tx, key)
		})
		return
	}

	_, err := tx.Exec(
		`
			DELETE FROM store
			WHERE key = ? AND (is_latest = 0 OR value = '')
		`,
		key,
	)

	common.FailOn(err)
}

func ClearAllKeysHistory(tx *sql.Tx, prefix string) {
	if tx == nil {
		common.RunTx(func(tx *sql.Tx) {
			ClearAllKeysHistory(tx, prefix)
		})
		return
	}

	_, err := tx.Exec(
		`
			DELETE FROM store
			WHERE key LIKE ? || '%' AND (is_latest = 0 OR value = '')
		`,
		prefix,
	)

	common.FailOn(err)
}
