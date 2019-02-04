package sinks

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fatih/structs"
)

var site247Alert = Site247Alert{
	MonitorName:          "brand_test",
	MonitorGroupName:     "search",
	SearchPollFrequency:  1,
	MonitorID:            2136797812307,
	FailedLocations:      "Delhi,Bangalore",
	MonitorURL:           "https://localhost:4000/search?query=brand_test",
	IncidentTimeISO:      "2018-07-24T18:43:08+0530",
	MonitorType:          "URL",
	Status:               "DOWN",
	Timezone:             "Asia/Calcutta",
	IncidentTime:         "July 24, 2018 6:43 PM IST",
	IncidentReason:       "Host Unavailable",
	OutageTimeUnixFormat: "1532437988741",
	RCALink:              "https://www.rcalinkdummy.com/somelink",
	Tags:                 []map[string]interface{}{{"tag": "value"}},
}

func TestEventFromSite247(t *testing.T) {
	event := EventFromSite247(site247Alert)
	if event.EventType != fmt.Sprintf("site247.%s.%s.%s", site247Alert.MonitorGroupName, site247Alert.MonitorName, site247Alert.Status) {
		t.Errorf("Event type not matching. expected : %s, got: %s", fmt.Sprintf("site247.%s.%s.%s", site247Alert.MonitorGroupName, site247Alert.MonitorName, site247Alert.Status), event.EventType)
	}
	if !reflect.DeepEqual(event.Data, structs.New(site247Alert).Map()) {
		t.Errorf("Event data not matching. expected : %v, got: %v", site247Alert, event.Data)
	}
	t.Log("TestEventFromSite247 completed")
}
