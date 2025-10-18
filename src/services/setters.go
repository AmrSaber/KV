package services

import (
	"database/sql"
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func SetValue(tx *sql.Tx, key string, value string, expiresAt *time.Time) {
	if tx == nil {
		common.RunTx(func(tx *sql.Tx) {
			SetValue(tx, key, value, expiresAt)
		})
		return
	}

	// Skip write if attempting to write identical values
	currentValue, currentExipry := GetValue(tx, key)
	if common.EqualStringPtrs(currentValue, &value) && common.EqualTimePtrs(currentExipry, expiresAt) {
		return
	}
	_, err := tx.Exec("UPDATE store SET is_latest = 0 WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	_, err = tx.Exec(
		`INSERT INTO store (key, value, expires_at) VALUES (?, ?, ?)`,
		key,
		value,
		common.FormatTimePtr(expiresAt),
	)
	common.FailOn(err)
}

func PruneKey(tx *sql.Tx, key string) {
	if tx == nil {
		common.RunTx(func(tx *sql.Tx) {
			PruneKey(tx, key)
		})
		return
	}

	_, err := tx.Exec("DELETE FROM store WHERE key = ?", key)
	common.FailOn(err)
}
