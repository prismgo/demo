package queuedemo

import (
	"context"

	"github.com/prismgo/framework/queue"
)

// WorkerJob records normal handling or observes worker timeout cancellation.
type WorkerJob struct {
	TraceID       string `json:"trace_id"`
	Label         string `json:"label"`
	WaitForCancel bool   `json:"wait_for_cancel,omitempty"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*WorkerJob]()
}

// Handle records normal work or waits for worker cancellation.
func (j *WorkerJob) Handle(ctx context.Context) error {
	if j.WaitForCancel {
		<-ctx.Done()
		RecordTrace(j.TraceID, j.Label+":cancelled")
		return ctx.Err()
	}
	RecordTrace(j.TraceID, j.Label)
	return nil
}
