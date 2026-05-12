package common

import "strings"

func ValidateDBName(db string) {
	Assert(!strings.Contains(db, "@"), "DB name cannot contain '@'")
	Assert(!strings.Contains(db, " "), "DB name cannot contain spaces")
}
