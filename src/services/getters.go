package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func GetValue(key string) (*string, *time.Time) {
	var value sql.NullString
	var expiresAt sql.NullTime

	err := common.GlobalTx.QueryRow("SELECT value, expires_at FROM store WHERE key = ? AND is_latest = 1", key).Scan(&value, &expiresAt)
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

func ListItems(prefix string, matchType MatchType) []KVItem {
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

	rows, err := common.GlobalTx.Query(query, prefix)
	common.FailOn(err)

	return parseKVItems(rows)
}

func ListKeys(prefix string, matchType MatchType) []string {
	items := ListItems(prefix, matchType)

	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}

	return keys
}

func ListKeyHistory(key string) []KVItem {
	rows, err := common.GlobalTx.Query(`
		SELECT key, value, expires_at, timestamp
		FROM store
		WHERE key = ?
		ORDER BY id ASC`,
		key,
	)
	common.FailOn(err)

	return parseKVItems(rows)
}

func GetHistoryItem(key string, steps int) KVItem {
	var item KVItem
	err := common.GlobalTx.QueryRow(`
		SELECT key, value, expires_at, timestamp
		FROM store
		WHERE key = ?
		ORDER BY id DESC
		LIMIT ?, 1`,
		key,
		steps,
	).Scan(&item.Key, &item.Value, &item.ExpiresAt, &item.Timestamp)
	common.FailOn(err)

	return item
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
