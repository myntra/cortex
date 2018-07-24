package store

import (
	"encoding/json"

	"github.com/myntra/cortex/pkg/rules"

	"github.com/hashicorp/raft"
	"github.com/myntra/cortex/pkg/events"
)

type db struct {
	buckets map[string]*events.Bucket
	rules   map[string]*rules.Rule
	scripts map[string][]byte
}

type fsmSnapShot struct {
	data *db
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.data)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapShot) Release() {}
