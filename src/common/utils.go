// Package common for all common functionality and utilities
package common

import "time"

func FailOn(err error) {
	if err != nil {
		panic(err)
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

	formatted := t.Format(time.RFC3339)
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
