package config

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {

	raftAddr := ":8878"
	httpAddr := ":8879"

	raftListener, err := net.Listen("tcp", raftAddr)
	require.NoError(t, err)
	httpListener, err := net.Listen("tcp", httpAddr)
	require.NoError(t, err)

	cfg := &Config{
		NodeID:               "node0",
		Dir:                  "./data",
		JoinAddr:             "",
		SnapshotInterval:     30,
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

	err = cfg.Validate()
	require.NoError(t, err)

}
