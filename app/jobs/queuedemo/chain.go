package queuedemo

import (
	"context"
	"errors"

	"github.com/prismgo/framework/queue"
)

// ChainJob is one observable link in a queue chain.
type ChainJob struct {
	TraceID string `json:"trace_id"`
	Label   string `json:"label"`
	Fail    bool   `json:"fail,omitempty"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*ChainJob]()
}

// Handle records the chain link and optionally stops the chain.
func (j *ChainJob) Handle(context.Context) error {
	RecordTrace(j.TraceID, j.Label)
	if j.Fail {
		return queue.Fail(errors.New("queue demo chain failure"))
	}
	return nil
}
