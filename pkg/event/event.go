package event

import (
	"net/http"

	"github.com/cnf/structhash"

	"github.com/fnproject/cloudevent"
)

// Event wraps cloudevent.CloudEvent
type Event struct {
	*cloudevent.CloudEvent
}

// Hash returns md5 hash string of the type
func (e *Event) Hash() []byte {

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

	return structhash.Md5(data, 1)
}

// FromRequest parses the event from request
func FromRequest(req *http.Request) (*Event, error) {
	ce := &cloudevent.CloudEvent{}
	err := ce.FromRequest(req)
	return &Event{ce}, err
}
