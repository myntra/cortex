package sinks

import (
	"github.com/myntra/cortex/pkg/events"
	"time"
	"fmt"
	"github.com/fatih/structs"
)

type AzureAlert struct {
	SchemaID string    `json:"schemaId"`
	Data     AzureData `json:"data"`
}

type AzureData struct {
	Status  string       `json:"activated"`
	Context AzureContext `json:"context"`
}

type AzureContext struct {
	Activity AzureActivity `json:"activityLog"`
}

type AzureActivity struct {
	Channels             string                `json:"channels"`
	CorrelationID        string                `json:"correlationId"`
	EventSource          string                `json:"eventSource"`
	EventTimestamp       string                `json:"eventTimestamp"`
	EventDataID          string                `json:"eventDataId"`
	Level                string                `json:"level"`
	OperationName        string                `json:"operationName"`
	OperationID          string                `json:"operationId"`
	Properties           AzureActivityProperty `json:"properties"`
	ResourceID           string                `json:"resourceId"`
	ResourceGroupName    string                `json:"resourceGroupName"`
	ResourceProviderName string                `json:"resourceProviderName"`
	Status               string                `json:"status"`
	SubscriptionID       string                `json:"subscriptionId"`
	SubmissionTimestamp  string                `json:"submissionTimestamp"`
	ResourceType         string                `json:"resourceType"`
}

type AzureActivityProperty struct {
	Title                string `json:"title"`
	Details              string `json:"details"`
	CurrentHealthStatus  string `json:"currentHealthStatus"`
	PreviousHealthStatus string `json:"previousHealthStatus"`
	Type                 string `json:"type"`
	Cause                string `json:"cause"`
}

// EventFromAzure converts alerts sent from azure into cloud events
func EventFromAzure(alert AzureAlert) *events.Event {
	event := events.Event{
		Source:             "azure",
		Data:               structs.New(alert).Map(),
		ContentType:        "application/json",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		SchemaURL:          "",
		EventID:            generateUUID().String(),
		EventTime:          time.Now(),
		EventType:          fmt.Sprintf("azure.%s", alert.Data.Context.Activity.ResourceID),
	}
	return &event
}
