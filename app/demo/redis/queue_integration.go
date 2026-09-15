package redisdemo

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	queuecontract "github.com/prismgo/framework/contracts/queue"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/queue/payload"
	redisqueue "github.com/prismgo/framework/queue/redis"
	"github.com/prismgo/framework/queue/state"
)

// redisDemoJob is the serializable queue job used by the Redis queue scenarios.
type redisDemoJob struct {
	// Mode selects the handler behavior; "fail" always returns an error.
	Mode string `json:"mode"`
}

// redisDemoJobHits counts successful Redis queue job executions for the current scenario.
var redisDemoJobHits atomic.Int64

// Handle executes the demo job and fails deterministically in fail mode.
func (j *redisDemoJob) Handle(context.Context) error {
	if j.Mode == "fail" {
		return errors.New("redis demo job failed")
	}
	redisDemoJobHits.Add(1)
	return nil
}

// queueScenarioResources holds the unique keys and queue name of one queue scenario.
type queueScenarioResources struct {
	prefix       string
	queue        string
	failedPrefix string
}

// openQueueLive boots an isolated application whose default queue connection is Redis,
// scoped to unique prefixes so scenarios never collide.
func openQueueLive(failed bool, extra map[string]string) (queueScenarioResources, func(), error) {
	suffix := time.Now().UnixNano()
	unique := fmt.Sprintf("prismgo_demo_redis_queue_%d", suffix)
	resources := queueScenarioResources{
		prefix: unique + ":q",
		queue:  fmt.Sprintf("demo-%d", suffix),
	}
	env := map[string]string{
		"QUEUE_CONNECTION":      "redis",
		"REDIS_QUEUE_PREFIX":    resources.prefix,
		"REDIS_QUEUE":           resources.queue,
		"REDIS_QUEUE_BLOCK_FOR": "0",
	}
	if failed {
		resources.failedPrefix = unique + ":failed"
		env["QUEUE_FAILED_DRIVER"] = "redis"
		env["QUEUE_FAILED_STORE"] = "default"
		env["QUEUE_FAILED_PREFIX"] = resources.failedPrefix
	}
	for key, value := range extra {
		env[key] = value
	}
	_, base, err := openLiveWith(liveOptions{env: env})
	if err != nil {
		return queueScenarioResources{}, nil, err
	}
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if rq, err := redisQueueConnection(); err == nil {
			_ = rq.Clear(ctx, resources.queue)
		}
		if resources.failedPrefix != "" {
			if client, err := redisConnectionClient("default"); err == nil {
				keys, err := client.Keys(ctx, resources.failedPrefix+"*").Result()
				if err == nil && len(keys) > 0 {
					_ = client.Del(ctx, keys...).Err()
				}
			}
		}
		base()
	}
	return resources, cleanup, nil
}

// redisQueueConnection resolves the concrete Redis queue transport for key inspection.
func redisQueueConnection() (*redisqueue.RedisQueue, error) {
	conn, err := queue.Resolve().Queue("redis")
	if err != nil {
		return nil, fmt.Errorf("resolve redis queue connection: %w", err)
	}
	rq, ok := conn.(*redisqueue.RedisQueue)
	if !ok {
		return nil, fmt.Errorf("queue connection = %T, want *redisqueue.RedisQueue", conn)
	}
	return rq, nil
}

// queueDriverScenario dispatches and consumes one job through the Redis queue driver.
func queueDriverScenario() (string, error) {
	resources, cleanup, err := openQueueLive(false, nil)
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	redisDemoJobHits.Store(0)
	rq, err := redisQueueConnection()
	if err != nil {
		return "", err
	}
	if _, err := queue.Dispatch(ctx, &redisDemoJob{Mode: "ok"}, queue.OnConnection("redis"), queue.OnQueue(resources.queue)); err != nil {
		return "", fmt.Errorf("dispatch redis job: %w", err)
	}
	queued, err := rq.Size(ctx, resources.queue)
	if err != nil {
		return "", fmt.Errorf("read queued size: %w", err)
	}
	worker := queue.NewWorker(queue.Resolve())
	if err := worker.Work(ctx, queue.WorkerOptions{
		Connection:    "redis",
		Queues:        []string{resources.queue},
		Once:          true,
		StopWhenEmpty: true,
		RetryAfter:    time.Second,
		Tries:         1,
	}); err != nil {
		return "", fmt.Errorf("work redis queue: %w", err)
	}
	remaining, err := rq.Size(ctx, resources.queue)
	if err != nil {
		return "", fmt.Errorf("read remaining size: %w", err)
	}
	return fmt.Sprintf("queued=%d remaining=%d processed=%d", queued, remaining, redisDemoJobHits.Load()), nil
}

// queueReadyScenario verifies immediate jobs land on the Redis ready list.
func queueReadyScenario() (string, error) {
	resources, cleanup, err := openQueueLive(false, nil)
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for i := 0; i < 2; i++ {
		if _, err := queue.Dispatch(ctx, &redisDemoJob{Mode: "ok"}, queue.OnConnection("redis"), queue.OnQueue(resources.queue)); err != nil {
			return "", fmt.Errorf("dispatch ready job %d: %w", i, err)
		}
	}
	rq, err := redisQueueConnection()
	if err != nil {
		return "", err
	}
	ready, err := rq.Client().LLen(ctx, rq.ReadyKey(resources.queue)).Result()
	if err != nil {
		return "", fmt.Errorf("read ready list: %w", err)
	}
	delayed, err := rq.Client().ZCard(ctx, rq.DelayedKey(resources.queue)).Result()
	if err != nil {
		return "", fmt.Errorf("read delayed set: %w", err)
	}
	return fmt.Sprintf("ready=%d delayed=%d", ready, delayed), nil
}

// queueDelayedScenario verifies delayed jobs land on the Redis sorted set.
func queueDelayedScenario() (string, error) {
	resources, cleanup, err := openQueueLive(false, nil)
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := queue.Dispatch(ctx, &redisDemoJob{Mode: "ok"},
		queue.OnConnection("redis"), queue.OnQueue(resources.queue), queue.Delay(2*time.Second)); err != nil {
		return "", fmt.Errorf("dispatch delayed job: %w", err)
	}
	rq, err := redisQueueConnection()
	if err != nil {
		return "", err
	}
	ready, err := rq.Client().LLen(ctx, rq.ReadyKey(resources.queue)).Result()
	if err != nil {
		return "", fmt.Errorf("read ready list: %w", err)
	}
	delayed, err := rq.Client().ZCard(ctx, rq.DelayedKey(resources.queue)).Result()
	if err != nil {
		return "", fmt.Errorf("read delayed set: %w", err)
	}
	_, popErr := rq.Pop(ctx, []string{resources.queue})
	notDue := errors.Is(popErr, redisqueue.ErrEmpty)
	if !notDue {
		return "", fmt.Errorf("pop before due = %v, want empty", popErr)
	}
	return fmt.Sprintf("delayed=%d ready=%d not-due=%t", delayed, ready, notDue), nil
}

// queueBlockingPopScenario verifies the blocking pop waits for a late dispatch.
func queueBlockingPopScenario() (string, error) {
	resources, cleanup, err := openQueueLive(false, map[string]string{"REDIS_QUEUE_BLOCK_FOR": "2"})
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := queue.Resolve().Queue("redis")
	if err != nil {
		return "", fmt.Errorf("resolve redis queue: %w", err)
	}
	results := make(chan queuecontract.ReservedJob, 1)
	failures := make(chan error, 1)
	go func() {
		reserved, err := conn.Pop(ctx, []string{resources.queue}, queuecontract.PopWaitAvailable)
		if err != nil {
			failures <- err
			return
		}
		results <- reserved
	}()
	time.Sleep(200 * time.Millisecond)
	select {
	case err := <-failures:
		return "", fmt.Errorf("blocking pop before dispatch: %w", err)
	case <-results:
		return "", fmt.Errorf("blocking pop returned before dispatch, want a waiting pop")
	default:
	}
	if _, err := queue.Dispatch(ctx, &redisDemoJob{Mode: "ok"}, queue.OnConnection("redis"), queue.OnQueue(resources.queue)); err != nil {
		return "", fmt.Errorf("dispatch blocking job: %w", err)
	}
	select {
	case err := <-failures:
		return "", fmt.Errorf("blocking pop: %w", err)
	case reserved := <-results:
		if reserved == nil {
			return "", fmt.Errorf("blocking pop returned a nil reserved job")
		}
		if err := reserved.Delete(ctx); err != nil {
			return "", fmt.Errorf("delete reserved job: %w", err)
		}
		return "blocked=true popped=true", nil
	case <-ctx.Done():
		return "", fmt.Errorf("blocking pop timed out: %w", ctx.Err())
	}
}

// queueFailedScenario verifies failed Redis jobs are stored in the Redis failed store.
func queueFailedScenario() (string, error) {
	resources, cleanup, err := openQueueLive(true, nil)
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	id, err := queue.Dispatch(ctx, &redisDemoJob{Mode: "fail"},
		queue.OnConnection("redis"), queue.OnQueue(resources.queue), queue.Tries(1))
	if err != nil {
		return "", fmt.Errorf("dispatch failing job: %w", err)
	}
	worker := queue.NewWorker(queue.Resolve())
	if err := worker.Work(ctx, queue.WorkerOptions{
		Connection:    "redis",
		Queues:        []string{resources.queue},
		Once:          true,
		StopWhenEmpty: true,
		RetryAfter:    time.Second,
		Tries:         1,
	}); err != nil {
		return "", fmt.Errorf("work failing queue: %w", err)
	}
	page, err := queue.Failed().Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
	if err != nil {
		return "", fmt.Errorf("page failed jobs: %w", err)
	}
	var recorded *payload.FailedJob
	for index := range page.Items {
		if page.Items[index].JobID == id {
			recorded = &page.Items[index]
			break
		}
	}
	if recorded == nil || recorded.Error == "" {
		return "", fmt.Errorf("failed jobs = %#v, want job %q with an error", page.Items, id)
	}
	if err := queue.Failed().Forget(ctx, recorded.ID); err != nil {
		return "", fmt.Errorf("forget failed job: %w", err)
	}
	return "stored=true error=true", nil
}
