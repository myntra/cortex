package events

import (
	"net/http"

	"github.com/cnf/structhash"

	"github.com/fnproject/cloudevent"
)

// Event wraps cloudevent.CloudEvent
type Event struct {
	*cloudevent.CloudEvent
	hash []byte
}

// Hash returns md5 hash string of the type
func (e *Event) Hash() []byte {
	if len(e.hash) == 0 {
		data := new(cloudevent.CloudEvent)
		data.CloudEventsVersion = e.CloudEventsVersion
		data.ContentType = e.ContentType
		data.Data = e.Data
		//data.EventID = e.EventID
		data.EventType = e.EventType
		data.EventTypeVersion = e.EventTypeVersion
		data.Extensions = e.Extensions
		data.SchemaURL = e.SchemaURL
		data.Source = e.Source

		e.hash = structhash.Md5(data, 1)
	}

	return e.hash
}

// FromRequest parses the event from request
func FromRequest(req *http.Request) (*Event, error) {
	ce := &cloudevent.CloudEvent{}
	err := ce.FromRequest(req)
	return &Event{CloudEvent: ce}, err
}
