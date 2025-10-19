package services

import "time"

type KVItem struct {
	Key       string     `json:"key,omitempty" yaml:"key,omitempty"`
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty" yaml:"expires-at,omitempty"`
	Timestamp time.Time  `json:"timestamp" yaml:"timestamp"`
}
