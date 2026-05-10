package common

import "strings"

func ValidateDBName(db string) {
	Assert(db != "kv", "Cannot use 'kv' as DB name")
	Assert(!strings.Contains(db, "@"), "DB name cannot contain '@'")
	Assert(!strings.Contains(db, " "), "DB name cannot contain spaces")
}
