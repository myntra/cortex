package sinks

import (
	"fmt"
	"reflect"
	"testing"
	"github.com/fatih/structs"
)

var azureAlert = AzureAlert{
	SchemaID: "Microsoft.Insights/activityLogs",
	Data: AzureData{
		Status: "Activated",
		Context: AzureContext{
			Activity: AzureActivity{
				Channels:       "Admin, Operation",
				CorrelationID:  "a1be61fd-37ur-ba05-b827-cb874708babf",
				EventSource:    "ResourceHealth",
				EventTimestamp: "2018-09-04T23:09:03.343+00:00",
				Level:          "Informational",
				OperationName:  "Microsoft.Resourcehealth/healthevent/Activated/action",
				OperationID:    "2b37e2d0-7bda-489f-81c6-1447d02265b2",
				Properties: AzureActivityProperty{
					Title:                "Virtual Machine health status changed to unavailable",
					Details:              "Virtual machine has experienced an unexpected event",
					CurrentHealthStatus:  "Unavailable",
					PreviousHealthStatus: "Available",
					Type:                 "Downtime",
					Cause:                "PlatformInitiated",
				},
				ResourceID:           "/subscriptions/<subscription Id>/resourceGroups/<resource group>/providers/Microsoft.Compute/virtualMachines/<resource name>",
				ResourceGroupName:    "<resource group>",
				ResourceProviderName: "Microsoft.Resourcehealth/healthevent/action",
				Status:               "Active",
				SubscriptionID:       "<subscription Id>",
				SubmissionTimestamp:  "2018-09-04T23:11:06.1607287+00:00",
				ResourceType:         "Microsoft.Compute/virtualMachines",
			},
		},
	},
}

func TestEventFromAzure(t *testing.T) {
	event := EventFromAzure(azureAlert)
	if event.EventType != fmt.Sprintf("azure.%s", azureAlert.Data.Context.Activity.ResourceID) {
		t.Errorf("Event type not matching. expected : %s, got: %s", fmt.Sprintf("azure.%s", azureAlert.Data.Context.Activity.ResourceID), event.EventType)
	}
	if !reflect.DeepEqual(event.Data, structs.New(azureAlert).Map()) {
		t.Errorf("Event data not matching. expected : %v, got: %v", azureAlert, event.Data)
	}
	t.Log("TestEventFromAzure completed")
}
