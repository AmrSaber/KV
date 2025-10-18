package services

import (
	"database/sql"
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func GetValue(tx *sql.Tx, key string) (*string, *time.Time) {
	if tx == nil {
		var value *string
		var expiresAt *time.Time
		common.RunTx(func(tx *sql.Tx) { value, expiresAt = GetValue(tx, key) })
		return value, expiresAt
	}

	var value sql.NullString
	var expiresAt sql.NullTime

	err := tx.QueryRow("SELECT value, expires_at FROM store WHERE key = ? AND is_latest = 1", key).Scan(&value, &expiresAt)
	if err != sql.ErrNoRows {
		common.FailOn(err)
	}

	var retValue *string
	var retExpiresAt *time.Time

	if value.Valid {
		retValue = &value.String
	}

	if expiresAt.Valid {
		retExpiresAt = &expiresAt.Time
	}

	return retValue, retExpiresAt
}

func GetKeysMatchingPrefix(tx *sql.Tx, prefix string) []string {
	if tx == nil {
		var keys []string
		common.RunTx(func(tx *sql.Tx) { keys = GetKeysMatchingPrefix(tx, prefix) })
		return keys
	}

	if prefix == "" {
		return getAllKeys(tx)
	}

	rows, err := tx.Query("SELECT key FROM store WHERE key LIKE ? || '%' AND is_latest = 1", prefix)
	common.FailOn(err)

	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		common.FailOn(err)

		keys = append(keys, key)
	}

	return keys
}

func getAllKeys(tx *sql.Tx) []string {
	rows, err := tx.Query("SELECT key FROM store WHERE is_latest = 1")
	common.FailOn(err)

	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		common.FailOn(err)

		keys = append(keys, key)
	}

	return keys
}
