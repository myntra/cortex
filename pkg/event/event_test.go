package event

import (
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
		&cloudevent.CloudEvent{
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
		&cloudevent.CloudEvent{
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
		&cloudevent.CloudEvent{
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

	hash1, _ := existingEvent.Hash()
	hash2, _ := incomingEventDuplicate.Hash()
	hash3, _ := incomingEventUnique.Hash()

	if hash1 != hash2 {
		log.Fatal("matching duplicate events failed")
	}

	if hash2 == hash3 {
		log.Fatal("mathed unique events")
	}
}
