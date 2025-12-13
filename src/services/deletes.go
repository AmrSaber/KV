package services

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
)

func ClearKeyHistory(tx *sql.Tx, key string) {
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
	_, err := tx.Exec(
		`
			DELETE FROM store
			WHERE key LIKE ? || '%' AND (is_latest = 0 OR value = '')
		`,
		prefix,
	)

	common.FailOn(err)
}
