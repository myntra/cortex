package store

import (
	"github.com/golang/glog"

	"github.com/hashicorp/raft"
)

type fsmSnapShot struct {
	msgs *Messages
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	glog.Info("persist =>")

	// Encode message.
	b, err := f.msgs.MarshalMsg(nil)
	if err != nil {
		return err
	}

	// Write messages to sink.
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
