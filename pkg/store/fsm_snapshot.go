package store

import (
	"github.com/golang/glog"

	"github.com/hashicorp/raft"
)

type fsmSnapShot struct {
	data *DB
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	glog.Info("persist =>")

	// Encode data.
	b, err := f.data.MarshalMsg(nil)
	if err != nil {
		return err
	}

	// Write data to sink.
	if _, err := sink.Write(b); err != nil {
		glog.Info("persist => err ", err)
		sink.Cancel()
		return err
	}

	glog.Infof("persisted len => %v", len(b))

	return nil
}

func (f *fsmSnapShot) Release() {
	glog.Info("release =>")
}
