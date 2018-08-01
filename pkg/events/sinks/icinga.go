package sinks

import (
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/myntra/cortex/pkg/events"
)

// IcingaAlert structure for Icinga alert
type IcingaAlert struct {
	NotificationType       string `json:"notification_type"`
	ServiceDescription     string `json:"service_description"`
	HostAlias              string `json:"host_alias"`
	HostAddress            string `json:"host_address"`
	ServiceState           string `json:"service_state"`
	LongDateTime           string `json:"long_date_time"`
	ServiceOutput          string `json:"service_output"`
	NotificationAuthorName string `json:"notification_author_name"`
	NotificationComment    string `json:"notification_comment"`
	HostDisplayName        string `json:"host_display_name"`
	ServiceDisplayName     string `json:"service_display_name"`
}

// EventFromIcinga converts alerts sent from icinga into cloud events
func EventFromIcinga(alert IcingaAlert) *events.Event {
	event := events.Event{
		Source:             "icinga",
		Data:               structs.New(alert).Map(),
		ContentType:        "application/json",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		SchemaURL:          "",
		EventID:            generateUUID().String(),
		EventTime:          time.Now(),
		EventType:          fmt.Sprintf("icinga.%s.%s.%s", alert.ServiceDisplayName, alert.HostDisplayName, alert.ServiceOutput),
	}
	return &event
}
