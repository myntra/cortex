package store

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/myntra/cortex/pkg/executions"

	"github.com/myntra/cortex/pkg/rules"

	"github.com/hashicorp/raft"
	"github.com/myntra/cortex/pkg/events"
)

// DB boltdb storage
type DB struct {
	Buckets map[string]*events.Bucket     `json:"buckets"`
	Rules   map[string]*rules.Rule        `json:"rules"`
	History map[string]*executions.Record `json:"history"`
	Scripts map[string][]byte             `json:"script"`
}

type fsmSnapShot struct {
	data *DB
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	glog.Info("persist =>")
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.data)
		if err != nil {
			return err
		}

		glog.Infof("persist => %v %v", len(b), string(b))
		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		glog.Info("persist => err ", err)
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapShot) Release() {
	glog.Info("release =>")
}
