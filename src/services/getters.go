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

func MatchExistingKeysByPrefix(tx *sql.Tx, prefix string) []string {
	if tx == nil {
		var keys []string
		common.RunTx(func(tx *sql.Tx) { keys = MatchExistingKeysByPrefix(tx, prefix) })
		return keys
	}

	rows, err := tx.Query("SELECT key FROM store WHERE key LIKE ? || '%' AND is_latest = 1 AND value != ''", prefix)
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

func ListItems(tx *sql.Tx, prefix string) []KVItem {
	if tx == nil {
		var items []KVItem
		common.RunTx(func(tx *sql.Tx) { items = ListItems(tx, prefix) })
		return items
	}

	rows, err := tx.Query(`
		SELECT key, value, expires_at, timestamp
		FROM store
		WHERE key LIKE ? || '%' AND is_latest = 1 AND value != ''
	`,
		prefix,
	)

	common.FailOn(err)

	var items []KVItem
	for rows.Next() {
		var item KVItem
		var expiresAt sql.NullTime

		err = rows.Scan(&item.Key, &item.Value, &expiresAt, &item.Timestamp)
		common.FailOn(err)

		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}

		items = append(items, item)
	}

	return items
}
