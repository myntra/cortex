package aggregate

import (
	"encoding/json"

	"github.com/hashicorp/raft"
	"github.com/myntra/aggo/pkg/event"
)

type fsmSnapShot struct {
	db map[string]*event.RuleBucket
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.db)
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
