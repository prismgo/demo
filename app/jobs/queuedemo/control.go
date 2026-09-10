package queuedemo

import (
	"context"
	"errors"
	"time"

	"github.com/prismgo/framework/queue"
)

// ControlJob demonstrates the public fail, release, and skip error controls.
type ControlJob struct {
	TraceID      string        `json:"trace_id"`
	Label        string        `json:"label"`
	Mode         string        `json:"mode"`
	ReleaseDelay time.Duration `json:"release_delay,omitempty"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*ControlJob]()
}

// Handle returns the selected queue control error and records each attempt.
func (j *ControlJob) Handle(ctx context.Context) error {
	attempt := NextAttempt(j.TraceID, j.Label)
	RecordTrace(j.TraceID, j.Label+":attempt")
	switch j.Mode {
	case "fail":
		return queue.Fail(errors.New("queue demo requested immediate failure"))
	case "release":
		if attempt == 1 {
			return queue.ReleaseAfter(j.ReleaseDelay, errors.New("queue demo requested delayed release"))
		}
	case "skip":
		return queue.ErrSkipped
	case "error":
		return errors.New("queue demo requested retryable failure")
	case "timeout":
		<-ctx.Done()
		RecordTrace(j.TraceID, j.Label+":cancelled")
		return ctx.Err()
	}
	RecordTrace(j.TraceID, j.Label+":handled")
	return nil
}

// Failed records the final-failure callback for the immediate failure path.
func (j *ControlJob) Failed(context.Context, error) {
	RecordTrace(j.TraceID, j.Label+":failed-callback")
}

var _ queue.FailedProvider = (*ControlJob)(nil)
