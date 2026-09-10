package queuedemo

import (
	"context"
	"errors"

	"github.com/prismgo/framework/queue"
)

// OperationsJob exposes failure, retry, event, and restart behavior to queue demos.
type OperationsJob struct {
	TraceID   string `json:"trace_id"`
	Label     string `json:"label"`
	FailFirst bool   `json:"fail_first,omitempty"`
	Block     bool   `json:"block,omitempty"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*OperationsJob]()
}

// Handle records each attempt and optionally fails once or waits at a deterministic gate.
func (j *OperationsJob) Handle(ctx context.Context) error {
	attempt := NextAttempt(j.TraceID, j.Label)
	RecordTrace(j.TraceID, j.Label+":attempt")
	if j.Block {
		if err := EnterGate(ctx, j.TraceID); err != nil {
			return err
		}
	}
	if j.FailFirst && attempt == 1 {
		return errors.New("queue demo operations first attempt failure")
	}
	RecordTrace(j.TraceID, j.Label+":handled")
	return nil
}

// Failed records the final-failure callback documented by the queue component.
func (j *OperationsJob) Failed(context.Context, error) {
	RecordTrace(j.TraceID, j.Label+":failed-callback")
}
