// Package common for all common functionality and utilities
package common

import (
	"os"
	"time"
)

func FailOn(err error) {
	if err != nil {
		if GlobalTx != nil {
			GlobalTx.Rollback()
		}

		panic(err)
	}
}

func Fail(message string, args ...any) {
	if GlobalTx != nil {
		GlobalTx.Rollback()
	}

	Stderr.Printf(red(message)+"\n", args...)
	os.Exit(1)
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
