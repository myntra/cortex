package rules

// Rule is the array of related service events
type Rule struct {
	Title               string   `json:"title"`
	ID                  string   `json:"id"`
	ScriptID            string   `json:"script_id"`             // javascript script which is called before hookEndPoint is called.
	HookEndpoint        string   `json:"hook_endpoint"`         // endpoint which accepts a POST json objects
	HookRetry           int      `json:"hook_retry"`            // number of retries while attempting to post
	EventTypes          []string `json:"event_types"`           // a list of event types to look for. a regex pattern is also allowed
	WaitWindow          uint64   `json:"wait_window"`           // wait duration in milliseconds for events to arrive
	WaitWindowThreshold uint64   `json:"wait_window_threshold"` // wait duration threshold after which arriving events expand the wait window
	MaxWaitWindow       uint64   `json:"max_wait_window"`       // maximum wait duration until which WaitWindow can expand
}
