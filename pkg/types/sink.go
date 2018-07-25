package types

type Site247Alert struct {
	MonitorName          string `json:"MONITORNAME,omitempty"`
	MonitorGroupName     string `json:"MONITOR_GROUPNAME,omitempty"`
	SearchPollFrequency  int    `json:"SEARCH POLLFREQUENCY,omitempty"`
	MonitorID            int    `json:"MONITOR_ID,omitempty"`
	FailedLocations      string `json:"FAILED_LOCATIONS,omitempty"`
	MonitorURL           string `json:"MONITORURL,omitempty"`
	IncidentTimeISO      string `json:"INCIDENT_TIME_ISO,omitempty"`
	MonitorType          string `json:"MONITORTYPE,omitempty"`
	Status               string `json:"STATUS,omitempty"`
	Timezone             string `json:"TIMEZONE,omitempty"`
	IncidentTime         string `json:"INCIDENT_TIME,omitempty"`
	IncidentReason       string `json:"INCIDENT_REASON,omitempty"`
	OutageTimeUnixFormat int    `json:"OUTAGE_TIME_UNIX_FORMAT,omitempty"`
	RCALink              string `json:"RCA_LINK,omitempty"`
}
