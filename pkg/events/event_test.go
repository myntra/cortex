package events

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type exampleData struct {
	Alpha string `json:"alpha"`
	Beta  int    `json:"beta"`
}

func TestEventHashMatch(t *testing.T) {
	existingEvent := &Event{

		EventType:          "com.event.fortytwo",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            "42",
		EventTime:          time.Now(),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               &exampleData{Alpha: "julie", Beta: 42},
		Extensions:         map[string]string{"ext1": "value"},
	}

	incomingEventDuplicate := &Event{

		EventType:          "com.event.fortytwo",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            "43",
		EventTime:          time.Now(),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               &exampleData{Alpha: "julie", Beta: 42},
		Extensions:         map[string]string{"ext1": "value"},
	}

	incomingEventUnique := &Event{
		EventType:          "com.event.fortytwo",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            "43",
		EventTime:          time.Now(),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               &exampleData{Alpha: "bobby", Beta: 100},
		Extensions:         map[string]string{"ext1": "value"},
	}

	hash1 := existingEvent.Hash()
	hash2 := incomingEventDuplicate.Hash()
	hash3 := incomingEventUnique.Hash()

	require.True(t, bytes.Equal(hash1, hash2))
	require.False(t, bytes.Equal(hash2, hash3))

}
