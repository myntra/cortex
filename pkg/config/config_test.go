package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	cfg := &Config{
		NodeID:               "node0",
		RaftBindPort:         8878,
		Dir:                  "./data",
		JoinAddr:             "",
		DefaultDwell:         3 * 60 * 1000,   // 3 minutes
		DefaultMaxDwell:      6 * 60 * 1000,   // 6 minutes
		DefaultDwellDeadline: 2.5 * 60 * 1000, // 2.5 minutes
		MaxHistory:           1000,
		FlushInterval:        1000,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
