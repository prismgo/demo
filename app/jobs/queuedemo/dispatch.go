package queuedemo

import (
	"context"
	"errors"

	"github.com/prismgo/framework/queue"
)

// DispatchJob records the public effects of the dispatch examples.
type DispatchJob struct {
	TraceID       string `json:"trace_id"`
	Label         string `json:"label"`
	ReturnError   bool   `json:"return_error,omitempty"`
	CheckDeadline bool   `json:"check_deadline,omitempty"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*DispatchJob]()
}

// Handle records dispatch behavior and optionally returns a demo error.
func (j *DispatchJob) Handle(ctx context.Context) error {
	if j.ReturnError {
		return errors.New("queue demo sync dispatch failure")
	}
	if j.CheckDeadline {
		if _, ok := ctx.Deadline(); ok {
			RecordTrace(j.TraceID, j.Label+":deadline")
			return nil
		}
		RecordTrace(j.TraceID, j.Label+":no-deadline")
		return nil
	}
	RecordTrace(j.TraceID, j.Label+":handled")
	return nil
}
