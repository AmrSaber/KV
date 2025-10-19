package services

import (
	"time"

	"gopkg.in/yaml.v2"
)

type KVItem struct {
	Key       string     `json:"key,omitempty" yaml:"key,omitempty"`
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty" yaml:"expires-at,omitempty"`
	Timestamp time.Time  `json:"timestamp" yaml:"timestamp"`
}

func (item KVItem) String() string {
	output, _ := yaml.Marshal(item)
	return string(output)
}
