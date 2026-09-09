package queuedemo

import (
	"context"
	"time"

	"github.com/prismgo/framework/contracts/cache"
	"github.com/prismgo/framework/queue"
)

// StrategyJob demonstrates every optional strategy-provider interface. Empty
// unique/debounce keys and ShouldEncrypt=false keep those opt-in behaviors for
// their dedicated demo scenarios.
type StrategyJob struct {
	TraceID string `json:"trace_id"`
}

func init() {
	// Registration lets workers reconstruct the serialized demo payload.
	queue.RegisterType[*StrategyJob]()
}

// Handle records execution of the strategy demo job.
func (j *StrategyJob) Handle(context.Context) error {
	RecordTrace(j.TraceID, "handle")
	return nil
}

// QueueConnection returns a fallback connection overridden by dispatch options.
func (j *StrategyJob) QueueConnection() string { return "missing-provider-connection" }

// QueueName returns a fallback queue overridden by dispatch options.
func (j *StrategyJob) QueueName() string { return "missing-provider-queue" }

// QueueDelay returns a fallback delay overridden by dispatch options.
func (j *StrategyJob) QueueDelay() time.Duration { return 24 * time.Hour }

// Tries returns a fallback attempt limit overridden by dispatch options.
func (j *StrategyJob) Tries() int { return 9 }

// Timeout returns the worker timeout for the strategy demo.
func (j *StrategyJob) Timeout() time.Duration { return time.Minute }

// Backoff returns the retry delays for the strategy demo.
func (j *StrategyJob) Backoff() []time.Duration { return []time.Duration{time.Second, time.Minute} }

// RetryUntil returns the retry deadline for the strategy demo.
func (j *StrategyJob) RetryUntil() time.Time { return time.Now().Add(time.Hour) }

// MaxExceptions returns the maximum exception count.
func (j *StrategyJob) MaxExceptions() int { return 4 }

// FailOnTimeout requests failure recording when the job times out.
func (j *StrategyJob) FailOnTimeout() bool { return true }

// ShouldEncrypt keeps encryption disabled for this strategy example.
func (j *StrategyJob) ShouldEncrypt() bool { return false }

// Tags labels the strategy demo job.
func (j *StrategyJob) Tags() []string { return []string{"queue-demo", "strategies"} }

// Silenced suppresses lifecycle noise in the strategy demo.
func (j *StrategyJob) Silenced() bool { return true }

// UniqueID keeps uniqueness disabled for this strategy example.
func (j *StrategyJob) UniqueID() string { return "" }

// UniqueFor returns the demonstration uniqueness window.
func (j *StrategyJob) UniqueFor() time.Duration { return time.Minute }

// UniqueVia leaves the uniqueness store unspecified.
func (j *StrategyJob) UniqueVia() cache.Repository { return nil }

// UniqueUntilProcessing demonstrates release-before-processing behavior.
func (j *StrategyJob) UniqueUntilProcessing() bool { return true }

// DebounceID keeps debouncing disabled for this strategy example.
func (j *StrategyJob) DebounceID() string { return "" }

// DebounceFor returns the demonstration debounce window.
func (j *StrategyJob) DebounceFor() time.Duration { return time.Minute }

// DebounceVia leaves the debounce store unspecified.
func (j *StrategyJob) DebounceVia() cache.Repository { return nil }

// Failed accepts failure callbacks without adding side effects.
func (j *StrategyJob) Failed(context.Context, error) {}

// Middleware returns an observable middleware around job handling.
func (j *StrategyJob) Middleware() []queue.Middleware {
	return []queue.Middleware{queue.MiddlewareFunc(func(ctx context.Context, _ queue.Job, next queue.Next) error {
		RecordTrace(j.TraceID, "job:before")
		err := next(ctx)
		RecordTrace(j.TraceID, "job:after")
		return err
	})}
}

var (
	_ queue.ConnectionProvider            = (*StrategyJob)(nil)
	_ queue.QueueProvider                 = (*StrategyJob)(nil)
	_ queue.DelayProvider                 = (*StrategyJob)(nil)
	_ queue.TriesProvider                 = (*StrategyJob)(nil)
	_ queue.TimeoutProvider               = (*StrategyJob)(nil)
	_ queue.BackoffProvider               = (*StrategyJob)(nil)
	_ queue.RetryUntilProvider            = (*StrategyJob)(nil)
	_ queue.MaxExceptionsProvider         = (*StrategyJob)(nil)
	_ queue.FailOnTimeoutProvider         = (*StrategyJob)(nil)
	_ queue.EncryptedProvider             = (*StrategyJob)(nil)
	_ queue.TagsProvider                  = (*StrategyJob)(nil)
	_ queue.SilencedProvider              = (*StrategyJob)(nil)
	_ queue.MiddlewareProvider            = (*StrategyJob)(nil)
	_ queue.UniqueIDProvider              = (*StrategyJob)(nil)
	_ queue.UniqueForProvider             = (*StrategyJob)(nil)
	_ queue.UniqueViaProvider             = (*StrategyJob)(nil)
	_ queue.UniqueUntilProcessingProvider = (*StrategyJob)(nil)
	_ queue.DebounceIDProvider            = (*StrategyJob)(nil)
	_ queue.DebounceForProvider           = (*StrategyJob)(nil)
	_ queue.DebounceViaProvider           = (*StrategyJob)(nil)
	_ queue.FailedProvider                = (*StrategyJob)(nil)
)
