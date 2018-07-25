package sinks

import (
	"fmt"
	"time"

	"github.com/cortex/pkg/events"
	"github.com/cortex/pkg/types"
	"github.com/fnproject/cloudevent"
	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

// EventFromSite247 converts alerts sent from site24x7 into cloud events
func EventFromSite247(alert types.Site247Alert) *events.Event {
	event := events.Event{
		CloudEvent: &cloudevent.CloudEvent{
			Source:             "site247",
			Data:               alert,
			ContentType:        "application/json",
			EventTypeVersion:   "1.0",
			CloudEventsVersion: "0.1",
			SchemaURL:          "",
			EventID:            generateUUID().String(),
			EventTime:          &time.Time{},
			EventType:          fmt.Sprintf("site247.%s.%s", alert.MonitorGroupName, alert.MonitorName),
		},
	}
	return &event
}

func generateUUID() uuid.UUID {
	uid, err := uuid.NewV4()
	if err != nil {
		glog.Infof("Error in creating new UUID for event sink")
		return uuid.UUID{}
	}
	return uid
}
