package services

import (
	"time"

	"github.com/AmrSaber/kv/src/common"
)

func SetValue(key string, value string, expiresAt *time.Time, isLocked bool) {
	// Skip write if attempting to write identical values
	currentValue, currentExipry := GetValue(key)
	if common.EqualStringPtrs(currentValue, &value) && common.EqualTimePtrs(currentExipry, expiresAt) {
		return
	}
	_, err := common.GlobalTx.Exec("UPDATE store SET is_latest = 0 WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	_, err = common.GlobalTx.Exec(
		`INSERT INTO store (key, value, is_locked, expires_at) VALUES (?, ?, ?, ?)`,
		key,
		value,
		isLocked,
		common.FormatTimePtr(expiresAt),
	)
	common.FailOn(err)
}

func PruneKey(key string) {
	_, err := common.GlobalTx.Exec("DELETE FROM store WHERE key = ?", key)
	common.FailOn(err)
}

func LockKey(key string, password string) {
	item := GetItem(key)
	if item == nil {
		common.Fail("Key %q does not exist", key)
		return // To shut up the compiler
	}

	if item.IsLocked {
		common.Fail("Key %q is already locked, unlock it first", key)
	}

	// Delete existing record so value is no longer in history
	_, err := common.GlobalTx.Exec("DELETE FROM store WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	encryptedValue, err := common.Encrypt(item.Value, password)
	common.FailOn(err)

	// Set the new encrypted value
	SetValue(key, encryptedValue, item.ExpiresAt, true)
}

func UnlockKey(key string, password string) error {
	item := GetItem(key)
	if item == nil {
		common.Fail("Key %q does not exist", key)
		return nil // To shut up the compiler
	}

	if !item.IsLocked {
		common.Fail("Key %q is not locked", key)
	}

	// Delete existing record so value is no longer in history
	_, err := common.GlobalTx.Exec("DELETE FROM store WHERE key = ? AND is_latest = 1", key)
	common.FailOn(err)

	decryptedValue, err := common.Decrypt(item.Value, password)
	if err != nil {
		return err
	}

	// Set the new encrypted value
	SetValue(key, decryptedValue, item.ExpiresAt, false)
	return nil
}
