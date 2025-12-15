package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func GetValue(tx *sql.Tx, key string) (*string, *time.Time) {
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

func GetItem(tx *sql.Tx, key string) *KVItem {
	var item KVItem
	var expiresAt sql.NullTime

	err := tx.QueryRow(`
		SELECT value, timestamp, is_locked, is_hidden, expires_at
		FROM store
		WHERE key = ? AND is_latest = 1 AND value != ''`,
		key,
	).Scan(&item.Value, &item.Timestamp, &item.IsLocked, &item.IsHidden, &expiresAt)
	if err == sql.ErrNoRows {
		return nil
	} else {
		common.FailOn(err)
	}

	if expiresAt.Valid {
		item.ExpiresAt = &expiresAt.Time
	}

	return &item
}

type MatchType int

const (
	MatchAll MatchType = iota
	MatchExisting
	MatchDeleted
)

func ListItems(tx *sql.Tx, prefix string, matchType MatchType) []KVItem {
	query := `
		SELECT key, value, expires_at, timestamp, is_locked, is_hidden
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
	if tx == nil {
		var keys []string
		RunInTransaction(func(tx *sql.Tx) {
			keys = ListKeys(tx, prefix, matchType)
		})

		return keys
	}

	items := ListItems(tx, prefix, matchType)

	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}

	return keys
}

func ListKeyHistory(tx *sql.Tx, key string) []KVItem {
	rows, err := tx.Query(`
		SELECT key, value, expires_at, timestamp, is_locked, is_hidden
		FROM store
		WHERE key = ?
		ORDER BY id ASC`,
		key,
	)
	common.FailOn(err)

	return parseKVItems(rows)
}

func GetHistoryItem(tx *sql.Tx, key string, steps int) KVItem {
	var item KVItem
	err := tx.QueryRow(`
		SELECT key, value, timestamp, is_locked, is_hidden
		FROM store
		WHERE key = ?
		ORDER BY id DESC
		LIMIT ?, 1`,
		key,
		steps,
	).Scan(&item.Key, &item.Value, &item.Timestamp, &item.IsLocked, &item.IsHidden)
	common.FailOn(err)

	return item
}

func parseKVItems(rows *sql.Rows) []KVItem {
	var items []KVItem
	for rows.Next() {
		var item KVItem
		var expiresAt sql.NullTime

		err := rows.Scan(&item.Key, &item.Value, &expiresAt, &item.Timestamp, &item.IsLocked, &item.IsHidden)
		common.FailOn(err)

		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}

		items = append(items, item)
	}

	return items
}
