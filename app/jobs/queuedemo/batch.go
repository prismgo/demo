package queuedemo

import (
	"context"
	"errors"

	"github.com/prismgo/framework/queue"
)

// BatchJob is a serializable unit used to demonstrate batch progress.
type BatchJob struct {
	TraceID string `json:"trace_id"`
	Label   string `json:"label"`
	Fail    bool   `json:"fail,omitempty"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*BatchJob]()
}

// Handle records batch progress and optionally marks the job as failed.
func (j *BatchJob) Handle(context.Context) error {
	RecordTrace(j.TraceID, j.Label)
	if j.Fail {
		return queue.Fail(errors.New("queue demo batch failure"))
	}
	return nil
}
