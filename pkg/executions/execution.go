package executions

import (
	"time"

	"github.com/myntra/cortex/pkg/events"
)

//go:generate msgp

// Record stores a rules execution state and result
type Record struct {
	ID             string        `json:"id"`
	Bucket         events.Bucket `json:"bucket"`
	ScriptResult   interface{}   `json:"script_result"`
	HookStatusCode int           `json:"hook_status_code"`
	CreatedAt      time.Time     `json:"created_at"`
}
