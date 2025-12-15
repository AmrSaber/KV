package services

import (
	"database/sql"
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func SetValue(tx *sql.Tx, key string, value string, expiresAt *time.Time, isLocked bool) {
	// Skip write if attempting to write identical values
	currentValue, currentExipry := GetValue(tx, key)
	if common.EqualStringPtrs(currentValue, &value) && common.EqualTimePtrs(currentExipry, expiresAt) {
		return
	}

	// Get current hidden state to preserve it
	currentItem := GetItem(tx, key)
	isHidden := false
	if currentItem != nil {
		isHidden = currentItem.IsHidden
	}

	_, err := tx.Exec("UPDATE store SET is_latest = 0 WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	_, err = tx.Exec(
		`INSERT INTO store (key, value, is_locked, is_hidden, expires_at) VALUES (?, ?, ?, ?, ?)`,
		key,
		value,
		isLocked,
		isHidden,
		common.FormatTimePtr(expiresAt),
	)
	common.FailOn(err)
}

func PruneKey(tx *sql.Tx, key string) {
	_, err := tx.Exec("DELETE FROM store WHERE key = ?", key)
	common.FailOn(err)
}

func LockKey(tx *sql.Tx, key string, password string) {
	item := GetItem(tx, key)
	if item == nil {
		common.Fail("Key %q does not exist", key)
		return // To shut up the compiler
	}

	if item.IsLocked {
		common.Fail("Key %q is already locked, unlock it first", key)
	}

	// Preserve hidden state before deleting
	isHidden := item.IsHidden

	// Delete existing record so value is no longer in history
	_, err := tx.Exec("DELETE FROM store WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	encryptedValue, err := common.Encrypt(item.Value, password)
	common.FailOn(err)

	// Insert the new encrypted value with preserved hidden state
	_, err = tx.Exec(
		`INSERT INTO store (key, value, is_locked, is_hidden, expires_at) VALUES (?, ?, ?, ?, ?)`,
		key,
		encryptedValue,
		true,
		isHidden,
		common.FormatTimePtr(item.ExpiresAt),
	)
	common.FailOn(err)
}

func UnlockKey(tx *sql.Tx, key string, password string) error {
	item := GetItem(tx, key)
	if item == nil {
		common.Fail("Key %q does not exist", key)
		return nil // To shut up the compiler
	}

	if !item.IsLocked {
		common.Fail("Key %q is not locked", key)
	}

	// Preserve hidden state before deleting
	isHidden := item.IsHidden

	// Delete existing record so value is no longer in history
	_, err := tx.Exec("DELETE FROM store WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	decryptedValue, err := common.Decrypt(item.Value, password)
	if err != nil {
		return err
	}

	// Insert the new decrypted value with preserved hidden state
	_, err = tx.Exec(
		`INSERT INTO store (key, value, is_locked, is_hidden, expires_at) VALUES (?, ?, ?, ?, ?)`,
		key,
		decryptedValue,
		false,
		isHidden,
		common.FormatTimePtr(item.ExpiresAt),
	)
	common.FailOn(err)

	return nil
}

func HideKey(tx *sql.Tx, key string) {
	item := GetItem(tx, key)
	if item == nil {
		common.Fail("Key %q does not exist", key)
		return // To shut up the compiler
	}

	// Idempotent: silently succeed if already hidden
	if item.IsHidden {
		return
	}

	_, err := tx.Exec(
		"UPDATE store SET is_hidden = 1 WHERE key = ? AND is_latest = 1",
		key,
	)
	common.FailOn(err)
}

func ShowKey(tx *sql.Tx, key string) {
	item := GetItem(tx, key)
	if item == nil {
		common.Fail("Key %q does not exist", key)
		return // To shut up the compiler
	}

	// Idempotent: silently succeed if already visible
	if !item.IsHidden {
		return
	}

	_, err := tx.Exec(
		"UPDATE store SET is_hidden = 0 WHERE key = ? AND is_latest = 1",
		key,
	)
	common.FailOn(err)
}

func RenameKey(tx *sql.Tx, oldKey string, newKey string) {
	// Check if old key exists
	oldItem := GetItem(tx, oldKey)
	if oldItem == nil {
		common.Fail("Key %q does not exist", oldKey)
	}

	// Check if new key already exists
	newItem := GetItem(tx, newKey)
	if newItem != nil {
		common.Fail("Key %q already exists", newKey)
	}

	// Rename the key across all history items
	_, err := tx.Exec("UPDATE store SET key = ? WHERE key = ?", newKey, oldKey)
	common.FailOn(err)
}
