// Package common for all common functionality and utilities
package common

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func FailOn(err error) {
	if err != nil {
		panic(err)
	}
}

func Fail(message string, args ...any) {
	panic(fmt.Sprintf(message, args...))
}

func Assert(cond bool, message string, args ...any) {
	if !cond {
		Fail(message, args...)
	}
}

func EqualTimePtrs(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}

	if t1 == nil || t2 == nil {
		return false
	}

	return t1.Equal(*t2)
}

func FormatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}

	formatted := t.UTC().Format(time.DateTime)
	return &formatted
}

func EqualStringPtrs(s1, s2 *string) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	if s1 == nil || s2 == nil {
		return false
	}

	return *s1 == *s2
}

// ParseKey parses the key and returns key and database parts
func ParseKey(key string) (string, string) {
	if _, noParse := os.LookupEnv("KV_NO_PARSE_KEYS"); noParse {
		return key, DefaultDBName
	}

	parts := strings.Split(key, "@")
	Assert(len(parts) <= 2, "Unable to parse key %q with more than one '@' symbol", key)

	currentDB := GetConfig().CurrentDB

	// When there is no DB
	if len(parts) < 2 {
		parts = append(parts, currentDB)
	}

	// When DB is an empty string (e.g. "key@")
	if parts[1] == "" {
		parts[1] = currentDB
	}

	key, db := parts[0], parts[1]
	Assert(key != "", "Key must not be an empty string")
	ValidateDBName(db)

	return key, db
}

func ParseKeys(keys []string) map[string][]string {
	mappedKeys := make(map[string][]string)

	for _, key := range keys {
		key, db := ParseKey(key)
		mappedKeys[db] = append(mappedKeys[db], key)
	}

	return mappedKeys
}
