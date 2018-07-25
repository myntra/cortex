package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func (d *defaultStore) open() error {

	id := d.opt.NodeID

	if id == "" {
		data, err := ioutil.ReadFile(filepath.Join(d.opt.Dir, "node.id"))
		id = strings.TrimSpace(string(data))
		if os.IsNotExist(err) || id == "" {
			var data [4]byte
			if _, err := rand.Read(data[:]); err != nil {
				panic("random error: " + err.Error())
			}
			id = hex.EncodeToString(data[:])[:7]
			err = ioutil.WriteFile(filepath.Join(d.opt.Dir, "node.id"), []byte(id+"\n"), 0600)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	glog.Info("opening raft store \n")
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", d.opt.RaftAddr)
	if err != nil {
		return err
	}

	//raft.NewTCPTransportWithConfig
	transport, err := NewTCPTransport(d.opt.RaftListener, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	glog.Info("created raft transport \n")
	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(d.opt.Dir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	glog.Info("created snapshot store \n")

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore

	glog.Info("raft.db => ", filepath.Join(d.opt.Dir, "raft.db"))
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(d.opt.Dir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}
	logStore = boltDB
	stableStore = boltDB

	glog.Info("created boltdb store \n")
	// Instantiate the Raft systemd.
	ra, err := raft.NewRaft(config, (*fsm)(d), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	d.raft = ra
	d.boltDB = boltDB

	glog.Info("created raft systemd \n")

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
		ra.BootstrapCluster(configuration)

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
	} else {
		// join a remote node
		glog.Infof("join a remote node %v\n", d.opt.JoinAddr)
		err := httpRaftJoin(d.opt.JoinAddr, d.opt.NodeID, d.opt.RaftAddr)
		if err != nil {
			return err
		}
	}

	go d.flusher()

	return nil
}

func (d *defaultStore) snapshot() error {
	f := d.raft.Snapshot()
	return f.Error()
}

func (d *defaultStore) close() error {
	d.quitFlusherChan <- struct{}{}
	f := d.raft.Shutdown()
	if f.Error() != nil {
		return f.Error()
	}

	// close the raft database
	if d.boltDB != nil {
		d.boltDB.Close()
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
