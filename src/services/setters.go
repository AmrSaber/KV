package services

import (
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func SetValue(key string, value string, expiresAt *time.Time) {
	// Skip write if attempting to write identical values
	currentValue, currentExipry := GetValue(key)
	if common.EqualStringPtrs(currentValue, &value) && common.EqualTimePtrs(currentExipry, expiresAt) {
		return
	}
	_, err := common.GlobalTx.Exec("UPDATE store SET is_latest = 0 WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	_, err = common.GlobalTx.Exec(
		`INSERT INTO store (key, value, expires_at) VALUES (?, ?, ?)`,
		key,
		value,
		common.FormatTimePtr(expiresAt),
	)
	common.FailOn(err)
}

func PruneKey(key string) {
	_, err := common.GlobalTx.Exec("DELETE FROM store WHERE key = ?", key)
	common.FailOn(err)
}
