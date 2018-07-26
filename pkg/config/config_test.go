package config

import (
	"net"
	"testing"
)

func TestConfig(t *testing.T) {

	raftAddr := ":8878"
	httpAddr := ":8879"

	raftListener, err := net.Listen("tcp", raftAddr)
	if err != nil {
		t.Fatal(err)
	}

	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		NodeID:               "node0",
		Dir:                  "./data",
		JoinAddr:             "",
		DefaultDwell:         3 * 60 * 1000,   // 3 minutes
		DefaultMaxDwell:      6 * 60 * 1000,   // 6 minutes
		DefaultDwellDeadline: 2.5 * 60 * 1000, // 2.5 minutes
		MaxHistory:           1000,
		FlushInterval:        1000,
		HTTPAddr:             httpAddr,
		RaftAddr:             raftAddr,
		HTTPListener:         httpListener,
		RaftListener:         raftListener,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
