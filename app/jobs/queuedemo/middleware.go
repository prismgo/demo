package queuedemo

import (
	"context"
	"errors"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/queue"
)

var errMiddlewareDemo = errors.New("queue demo middleware failure")

// MiddlewareJob selects one documented middleware behavior through Mode.
type MiddlewareJob struct {
	TraceID string `json:"trace_id"`
	Label   string `json:"label"`
	Mode    string `json:"mode"`
	Store   string `json:"store"`
	Block   bool   `json:"block"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*MiddlewareJob]()
}

// Handle records middleware execution and optionally exercises failure paths.
func (j *MiddlewareJob) Handle(ctx context.Context) error {
	if j.Block {
		if err := EnterGate(ctx, j.TraceID); err != nil {
			return err
		}
	}
	RecordTrace(j.TraceID, j.Label+":handled")
	if j.Mode == "throttle" {
		return errMiddlewareDemo
	}
	return nil
}

// Middleware builds the middleware stack selected by Mode.
func (j *MiddlewareJob) Middleware() []queue.Middleware {
	switch j.Mode {
	case "order":
		return []queue.Middleware{queue.MiddlewareFunc(func(ctx context.Context, _ queue.Job, next queue.Next) error {
			RecordTrace(j.TraceID, "job:before")
			err := next(ctx)
			RecordTrace(j.TraceID, "job:after")
			return err
		})}
	case "skip":
		return []queue.Middleware{queue.SkipIf(func(queue.Job) bool { return true })}
	case "custom-skip":
		return []queue.Middleware{queue.MiddlewareFunc(func(context.Context, queue.Job, queue.Next) error {
			return queue.ErrSkipped
		})}
	case "rate":
		return []queue.Middleware{queue.RateLimit("queue-demo:rate:"+j.TraceID, 1, time.Minute)}
	case "overlap":
		return []queue.Middleware{queue.WithoutOverlapping("queue-demo:overlap:" + j.TraceID).
			DontRelease().ExpireAfter(time.Minute).Via(cache.Store(j.Store)).Shared()}
	case "throttle":
		return []queue.Middleware{queue.ThrottlesExceptions(1, time.Minute).
			By("queue-demo:throttle:" + j.TraceID).Backoff(10 * time.Millisecond).
			When(func(err error) bool { return errors.Is(err, errMiddlewareDemo) }).Via(cache.Store(j.Store))}
	default:
		return nil
	}
}

var _ queue.MiddlewareProvider = (*MiddlewareJob)(nil)
