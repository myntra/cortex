package sinks

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fatih/structs"
)

var icingaAlert = IcingaAlert{
	HostAlias:              "hostname-alias",
	HostAddress:            "1.2.3.4",
	ServiceState:           "CRITICAL",
	ServiceOutput:          "connect to address 1.2.3.4 and port 5000: Connection refused",
	NotificationAuthorName: "",
	ServiceDisplayName:     "servicename-26378",
	ServiceDescription:     "servicename-26378",
	LongDateTime:           "2032-07-30 19:16:14 +0530",
	NotificationComment:    "",
	HostDisplayName:        "hostname",
	NotificationType:       "PROBLEM",
}

func TestEventFromIcinga(t *testing.T) {
	event := EventFromIcinga(icingaAlert)
	if event.EventType != fmt.Sprintf("icinga.%s.%s.%s", icingaAlert.ServiceDisplayName, icingaAlert.HostDisplayName, icingaAlert.ServiceOutput) {
		t.Errorf("Event type not matching. expected : %s, got: %s", fmt.Sprintf("icinga.%s.%s.%s", icingaAlert.ServiceDisplayName, icingaAlert.HostDisplayName, icingaAlert.ServiceOutput), event.EventType)
	}
	if !reflect.DeepEqual(event.Data, structs.New(icingaAlert).Map()) {
		t.Errorf("Event data not matching. expected : %v, got: %v", icingaAlert, event.Data)
	}
	t.Log("TestEventFromIcinga completed")
}
