package queuedemo

import (
	"context"
	"time"

	"github.com/prismgo/framework/cache"
	contract "github.com/prismgo/framework/contracts/cache"
	"github.com/prismgo/framework/queue"
)

// DebounceJob records which dispatch survives a shared debounce window.
type DebounceJob struct {
	TraceID string        `json:"trace_id"`
	Label   string        `json:"label"`
	Store   string        `json:"store"`
	Window  time.Duration `json:"window"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*DebounceJob]()
}

// Handle records the dispatch that survives the debounce window.
func (j *DebounceJob) Handle(context.Context) error {
	RecordTrace(j.TraceID, j.Label+":handled")
	return nil
}

// DebounceID groups dispatches belonging to the same trace.
func (j *DebounceJob) DebounceID() string { return "queue-demo:" + j.TraceID }

// DebounceFor returns the configured debounce window or a demo default.
func (j *DebounceJob) DebounceFor() time.Duration {
	if j.Window <= 0 {
		return 100 * time.Millisecond
	}
	return j.Window
}

// DebounceVia selects the cache store used for debounce coordination.
func (j *DebounceJob) DebounceVia() contract.Repository { return cache.Store(j.Store) }

var (
	_ queue.DebounceIDProvider  = (*DebounceJob)(nil)
	_ queue.DebounceForProvider = (*DebounceJob)(nil)
	_ queue.DebounceViaProvider = (*DebounceJob)(nil)
)
