package sinks

import (
	"fmt"
	"testing"

	"github.com/cortex/pkg/types"
)

var alert = types.Site247Alert{
	MonitorName:          "nike_test",
	MonitorGroupName:     "search",
	SearchPollFrequency:  1,
	MonitorID:            2136797812307,
	FailedLocations:      "Delhi,Bangalore",
	MonitorURL:           "https://www.myntra.com/search?query=Nike",
	IncidentTimeISO:      "2018-07-24T18:43:08+0530",
	MonitorType:          "URL",
	Status:               "DOWN",
	Timezone:             "Asia/Calcutta",
	IncidentTime:         "July 24, 2018 6:43 PM IST",
	IncidentReason:       "Host Unavailable",
	OutageTimeUnixFormat: 1532437988741,
	RCALink:              "https://www.site24x7.com/rca.do?id=ygaPwszXNqar1O9ZCR3WWw6HsI%2BteZf8Tj7KEfQd5raoQbNDuhxUt%2FQoG0yksN9DXHqtLUl5%2FcRr%0AtIcRB15%2Bxw%3D%3D",
}

func TestEventFromSite247(t *testing.T) {
	event := EventFromSite247(alert)
	if event.EventType != fmt.Sprintf("site247.%s.%s", alert.MonitorGroupName, alert.MonitorName) {
		t.Errorf("Event type not matching. expected : %s, got: %s", fmt.Sprintf("site247.%s.%s", alert.MonitorGroupName, alert.MonitorName), event.EventType)
	}
	if event.Data != alert {
		t.Errorf("Event data not matching. expected : %v, got: %v", alert, event.Data)
	}
	t.Log("TestEventFromSite247 completed")
}
