package services

import (
	"database/sql"
	"os"
	"path"

	"github.com/AmrSaber/kv/src/common"
)

// MigrateOldDBName migrates old 'kv' db into new 'default' DB
// this check will always run in app start, it might be deleted in future releases
func MigrateOldDBName() {
	kvDBPath := path.Join(common.GetDataDirectory(), "kv.db")
	if _, err := os.Stat(kvDBPath); os.IsNotExist(err) {
		return
	}

	defaultDB := common.DefaultDBName
	defaultDBPath := common.GetConfig().GetDBPath(defaultDB)
	if _, err := os.Stat(defaultDBPath); !os.IsNotExist(err) {
		RunInTransaction(defaultDB, func(tx *sql.Tx) {
			if migrated, _ := common.ReadMetadata[bool](tx, common.MetadataKeyDBMigrated); !migrated {
				common.Warn(
					"Warning: You have both 'kv' and 'default' DBs. " +
						"'kv' DB won't get renamed to 'default'. " +
						"You can access 'kv' database through DB access syntax.",
				)

				err = common.WriteMetadata(tx, common.MetadataKeyDBMigrated, true)
				common.FailOn(err)
			}
		})

		return
	}

	common.BackupDBInPlace("kv")

	err := os.Rename(kvDBPath, defaultDBPath)
	if err != nil {
		common.Fail("Failed to rename database file: %v", err)
	}

	// Mark DB migrated
	RunInTransaction(defaultDB, func(tx *sql.Tx) {
		err := common.WriteMetadata(tx, common.MetadataKeyDBMigrated, true)
		common.FailOn(err)
	})
}

// MigrateInvalidKeys displays warning when it detects keys with @ in their names
// this check will always run in app start, it might be deleted in future releases
func MigrateInvalidKeys() {
	RunInTransaction(common.GetConfig().CurrentDB, func(tx *sql.Tx) {
		var keysExist bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM store WHERE key LIKE '%@%')").Scan(&keysExist)
		common.FailOn(err)

		if !keysExist {
			return
		}

		displayedWarning, _ := common.ReadMetadata[bool](tx, common.MetadataKeyKeysWarningDisplayed)

		if !displayedWarning {
			common.Warn(
				"Warning: Some keys have '@' in their names. " +
					"They are only accessible through setting KV_NO_PARSE_KEYS env variable.",
			)

			err = common.WriteMetadata(tx, common.MetadataKeyKeysWarningDisplayed, true)
			common.FailOn(err)
		}
	})
}
