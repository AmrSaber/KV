package services

import (
	"database/sql"
	"fmt"
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

type MatchType int

const (
	MatchAll MatchType = iota
	MatchExisting
	MatchDeleted
)

func ListItems(tx *sql.Tx, prefix string, matchType MatchType) []KVItem {
	if tx == nil {
		var items []KVItem
		common.RunTx(func(tx *sql.Tx) { items = ListItems(tx, prefix, matchType) })
		return items
	}

	query := `
		SELECT key, value, expires_at, timestamp
		FROM store
		WHERE key LIKE ? || '%' AND is_latest = 1
	`
	switch matchType {
	case MatchExisting:
		query += " AND value != ''"
	case MatchDeleted:
		query += " AND value = ''"
	case MatchAll:
		// Do nothing
	default:
		panic(fmt.Sprintf("Match type %q is not supported", matchType))
	}

	rows, err := tx.Query(query, prefix)
	common.FailOn(err)

	return parseKVItems(rows)
}

func ListKeys(tx *sql.Tx, prefix string, matchType MatchType) []string {
	items := ListItems(tx, prefix, matchType)

	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}

	return keys
}

func ListKeyHistory(tx *sql.Tx, key string) []KVItem {
	if tx == nil {
		var items []KVItem
		common.RunTx(func(tx *sql.Tx) { items = ListKeyHistory(tx, key) })
		return items
	}

	rows, err := tx.Query(`
		SELECT key, value, expires_at, timestamp
		FROM store
		WHERE key = ?
		ORDER BY timestamp ASC`,
		key,
	)
	common.FailOn(err)

	return parseKVItems(rows)
}

func parseKVItems(rows *sql.Rows) []KVItem {
	var items []KVItem
	for rows.Next() {
		var item KVItem
		var expiresAt sql.NullTime

		err := rows.Scan(&item.Key, &item.Value, &expiresAt, &item.Timestamp)
		common.FailOn(err)

		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}

		items = append(items, item)
	}

	return items
}
