package queuedemo

import (
	"context"
	"time"

	"github.com/prismgo/framework/cache"
	contract "github.com/prismgo/framework/contracts/cache"
	"github.com/prismgo/framework/queue"
)

// UniqueJob demonstrates provider-based unique keys, TTLs, stores, and the
// optional release-before-processing policy.
type UniqueJob struct {
	TraceID         string `json:"trace_id"`
	UniqueKey       string `json:"unique_key"`
	Label           string `json:"label"`
	Store           string `json:"store"`
	Block           bool   `json:"block"`
	UntilProcessing bool   `json:"until_processing"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*UniqueJob]()
}

// Handle records processing and optionally waits at the concurrency gate.
func (j *UniqueJob) Handle(ctx context.Context) error {
	RecordTrace(j.TraceID, j.Label+":started")
	if j.Block {
		if err := EnterGate(ctx, j.TraceID); err != nil {
			return err
		}
	}
	RecordTrace(j.TraceID, j.Label+":finished")
	return nil
}

// UniqueID returns the explicit unique key or falls back to the trace ID.
func (j *UniqueJob) UniqueID() string {
	if j.UniqueKey != "" {
		return "queue-demo:" + j.UniqueKey
	}
	return "queue-demo:" + j.TraceID
}

// UniqueFor retains the uniqueness lock for the demo window.
func (j *UniqueJob) UniqueFor() time.Duration { return time.Minute }

// UniqueVia selects the cache store used for uniqueness coordination.
func (j *UniqueJob) UniqueVia() contract.Repository { return cache.Store(j.Store) }

// UniqueUntilProcessing reports whether the lock is released before handling.
func (j *UniqueJob) UniqueUntilProcessing() bool { return j.UntilProcessing }

var (
	_ queue.UniqueIDProvider              = (*UniqueJob)(nil)
	_ queue.UniqueForProvider             = (*UniqueJob)(nil)
	_ queue.UniqueViaProvider             = (*UniqueJob)(nil)
	_ queue.UniqueUntilProcessingProvider = (*UniqueJob)(nil)
)
