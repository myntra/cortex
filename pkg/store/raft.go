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

	bindAddr := d.opt.GetBindAddr()
	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stderr)
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

	return nil
}

func (d *defaultStore) close() error {
	d.quitFlusherChan <- struct{}{}
	f := d.raft.Shutdown()
	if f.Error() != nil {
		return f.Error()
	}
	glog.Info("raft shut down")
	glog.Flush()
	return nil
}

func (d *defaultStore) acceptJoin(nodeID, addr string) error {
	glog.Infof("received join request for remote node %s at %s", nodeID, addr)

	configFuture := d.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		glog.Infof("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				glog.Infof("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := d.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := d.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	glog.Infof("node %s at %s joined successfully", nodeID, addr)
	return nil

}

func (d *defaultStore) acceptLeave(nodeID string) error {

	glog.Infof("received leave request for remote node %s", nodeID)

	cf := d.raft.GetConfiguration()

	if err := cf.Error(); err != nil {
		glog.Infof("failed to get raft configuration")
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeID) {
			f := d.raft.RemoveServer(server.ID, 0, 0)
			if err := f.Error(); err != nil {
				glog.Infof("failed to remove server %s", nodeID)
				return err
			}

			glog.Infof("node %s left successfully", nodeID)
			return nil
		}
	}

	glog.Infof("node %s not exists in raft group", nodeID)

	return nil

}
