package store

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func (d *defaultStore) open() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(d.opt.NodeID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", d.opt.GetBindAddr())
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(d.opt.GetBindAddr(), addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(d.opt.Dir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore

	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(d.opt.Dir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}
	logStore = boltDB
	stableStore = boltDB

	// Instantiate the Raft systemd.
	ra, err := raft.NewRaft(config, (*fsm)(d), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	d.raft = ra

	// bootstrap single node configuration
	if d.opt.JoinAddr == "" {
		glog.Infof("starting %v in a single node cluster \n", d.opt.NodeID)
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		f := ra.BootstrapCluster(configuration)
		if f.Error() != nil {
			return f.Error()
		}

		// since in bootstrap mode, block until leadership is attained.
	loop:
		for {
			select {
			case leader := <-d.raft.LeaderCh():
				glog.Info("isLeader ", leader)
				if leader {
					break loop
				}
			}
		}
	}

	go d.flusher()

	return nil
}

func (d *defaultStore) close() error {
	d.quitFlusherChan <- struct{}{}
	f := d.raft.Shutdown()
	if f.Error() != nil {
		return f.Error()
	}
	glog.Flush()
	return nil
}
