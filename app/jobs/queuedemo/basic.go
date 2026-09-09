// Package queuedemo contains serializable jobs used by the queue demos.
package queuedemo

import (
	"context"

	"github.com/prismgo/framework/queue"
)

// BasicJob is the smallest useful queue job: payload data plus Handle.
type BasicJob struct {
	Message string `json:"message"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*BasicJob]()
}

// Handle processes the example message.
func (j *BasicJob) Handle(context.Context) error { return nil }
