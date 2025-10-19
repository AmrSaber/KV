package services

import (
	"github.com/AmrSaber/kv/src/common"
)

func ClearKeyHistory(key string) {
	_, err := common.GlobalTx.Exec(
		`
			DELETE FROM store
			WHERE key = ? AND (is_latest = 0 OR value = '')
		`,
		key,
	)

	common.FailOn(err)
}

func ClearAllKeysHistory(prefix string) {
	_, err := common.GlobalTx.Exec(
		`
			DELETE FROM store
			WHERE key LIKE ? || '%' AND (is_latest = 0 OR value = '')
		`,
		prefix,
	)

	common.FailOn(err)
}
