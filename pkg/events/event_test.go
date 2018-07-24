package events

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/fnproject/cloudevent"
)

type exampleData struct {
	Alpha string `json:"alpha"`
	Beta  int    `json:"beta"`
}

func tptr(t time.Time) *time.Time { return nil }

func TestEventHashMatch(t *testing.T) {
	existingEvent := &Event{
		CloudEvent: &cloudevent.CloudEvent{
			EventType:          "com.event.fortytwo",
			EventTypeVersion:   "1.0",
			CloudEventsVersion: "0.1",
			Source:             "/sink",
			EventID:            "42",
			EventTime:          tptr(time.Now()),
			SchemaURL:          "http://www.json.org",
			ContentType:        "application/json",
			Data:               &exampleData{Alpha: "julie", Beta: 42},
			Extensions:         map[string]string{"ext1": "value"},
		},
	}

	incomingEventDuplicate := &Event{
		CloudEvent: &cloudevent.CloudEvent{
			EventType:          "com.event.fortytwo",
			EventTypeVersion:   "1.0",
			CloudEventsVersion: "0.1",
			Source:             "/sink",
			EventID:            "43",
			EventTime:          tptr(time.Now()),
			SchemaURL:          "http://www.json.org",
			ContentType:        "application/json",
			Data:               &exampleData{Alpha: "julie", Beta: 42},
			Extensions:         map[string]string{"ext1": "value"},
		},
	}

	incomingEventUnique := &Event{
		CloudEvent: &cloudevent.CloudEvent{
			EventType:          "com.event.fortytwo",
			EventTypeVersion:   "1.0",
			CloudEventsVersion: "0.1",
			Source:             "/sink",
			EventID:            "43",
			EventTime:          tptr(time.Now()),
			SchemaURL:          "http://www.json.org",
			ContentType:        "application/json",
			Data:               &exampleData{Alpha: "bobby", Beta: 100},
			Extensions:         map[string]string{"ext1": "value"},
		},
	}

	hash1 := existingEvent.Hash()
	hash2 := incomingEventDuplicate.Hash()
	hash3 := incomingEventUnique.Hash()

	if !bytes.Equal(hash1, hash2) {
		log.Fatal("matching duplicate events failed")
	}

	if bytes.Equal(hash2, hash3) {
		log.Fatal("mathed unique events")
	}
}
