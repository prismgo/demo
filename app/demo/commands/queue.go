package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/prismgo/framework/encoding"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/queue/payload"
	queueredis "github.com/prismgo/framework/queue/redis"
	"github.com/prismgo/framework/queue/state"
	"github.com/redis/go-redis/v9"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runQueueCommand(ctx context.Context, executable, root, name string) (result Result, err error) {
	url := strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL"))
	if url == "" {
		return Result{}, fmt.Errorf("%s requires PRISMGO_REDIS_TEST_URL for Redis integration", name)
	}
	options, err := redis.ParseURL(url)
	if err != nil {
		return Result{}, fmt.Errorf("parse commands demo Redis URL: %w", err)
	}
	host, port, err := net.SplitHostPort(options.Addr)
	if err != nil {
		return Result{}, fmt.Errorf("parse commands demo Redis address: %w", err)
	}
	client := redis.NewClient(options)
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close commands demo Redis client: %w", closeErr))
		}
	}()
	if err := client.Ping(ctx).Err(); err != nil {
		return Result{}, fmt.Errorf("connect commands demo Redis: %w", err)
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	failedPrefix := "prismgo_commands_failed_" + suffix
	queuePrefix := "prismgo_commands_queue_" + suffix
	queueName := "commands-" + suffix
	failedID := "commands-failed-" + suffix
	store := queueredis.NewRedisFailedStoreFromClient(client, queueredis.RedisOptions{Prefix: failedPrefix, Codec: encoding.JSON()})
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if cleanupErr := store.Flush(cleanupCtx); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean commands demo failed jobs: %w", cleanupErr))
		}
		if cleanupErr := client.Del(cleanupCtx, queuePrefix+":queues:"+queueName, queuePrefix+":queues:"+queueName+":notify", queuePrefix+":queues:"+queueName+":reserved", queuePrefix+":queues:"+queueName+":delayed").Err(); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean commands demo Redis queue: %w", cleanupErr))
		}
	}()
	environment := []string{
		"QUEUE_CONNECTION=redis", "QUEUE_ENCODING=json", "QUEUE_FAILED_DRIVER=redis", "QUEUE_FAILED_STORE=default",
		"QUEUE_FAILED_PREFIX=" + failedPrefix, "REDIS_URL=" + url, "REDIS_HOST=" + host, "REDIS_PORT=" + port,
		"REDIS_MAIN_DB=0", "REDIS_QUEUE_PREFIX=" + queuePrefix,
	}
	args := []string{"queue", "redis", "--queue=" + queueName, "--stop-when-empty", "--sleep=0"}
	readyJobs := 0
	switch name {
	case "queue":
		if err := enqueueCommandsJob(ctx, client, queuePrefix, queueName, suffix); err != nil {
			return Result{}, err
		}
	case "queue-work":
		args[0] = "queue:work"
		if err := enqueueCommandsJob(ctx, client, queuePrefix, queueName, suffix); err != nil {
			return Result{}, err
		}
	case "queue-failed", "queue-retry", "queue-forget":
		envelope, err := commandsJobEnvelope(queueName, suffix, false)
		if err != nil {
			return Result{}, err
		}
		failed := payload.FailedJob{
			ID: failedID, JobID: envelope.ID, Connection: "redis", Queue: queueName,
			JobName: envelope.Name, Error: "demonstration failure", FailedAt: time.Now(),
			Envelope: envelope,
		}
		if err := store.Record(ctx, failed); err != nil {
			return Result{}, fmt.Errorf("%s record failed job: %w", name, err)
		}
		switch name {
		case "queue-failed":
			args = []string{"queue:failed"}
		case "queue-retry":
			args = []string{"queue:retry", failedID}
		case "queue-forget":
			args = []string{"queue:forget", failedID}
		}
	case "queue-flush":
		for i := range 2 {
			if err := store.Record(ctx, payload.FailedJob{ID: fmt.Sprintf("%s-%d", failedID, i), JobID: failedID, Connection: "redis", Queue: queueName, JobName: "demo", Error: "demonstration failure", FailedAt: time.Now()}); err != nil {
				return Result{}, fmt.Errorf("record %s failed job %d: %w", name, i, err)
			}
		}
		args = []string{"queue:flush"}
	case "queue-restart":
		restartKey := "commands-restart-" + suffix
		environment = append(environment, "QUEUE_RESTART_CACHE=redis", "QUEUE_RESTART_KEY="+restartKey, "CACHE_REDIS_PREFIX=commands-cache-"+suffix)
		args = []string{"queue:restart"}
	case "worker-connection":
		args = []string{"queue", "--queue=" + queueName, "--stop-when-empty", "--sleep=0"}
		environment = append(environment, "QUEUE_CONNECTION=redis")
		readyJobs = 1
	case "worker-queue":
		args = []string{"queue", "redis", "--queue=unused," + queueName, "--stop-when-empty", "--sleep=0"}
		readyJobs = 1
	case "worker-once":
		args = []string{"queue", "redis", "--queue=" + queueName, "--once", "--sleep=0"}
		readyJobs = 2
	case "worker-stop-when-empty":
		// The default argument set exits only after Redis reports an empty queue.
		readyJobs = 1
	case "worker-sleep":
		args = []string{"queue", "redis", "--queue=" + queueName, "--sleep=1", "--max-time=1"}
	case "worker-timeout":
		args = []string{"queue", "redis", "--queue=" + queueName, "--once", "--timeout=1", "--sleep=0"}
		job := &jobs.WorkerJob{TraceID: suffix, Label: "timeout", WaitForCancel: true}
		jobName, err := queue.JobTypeName(job)
		if err != nil {
			return Result{}, fmt.Errorf("identify %s job: %w", name, err)
		}
		body, err := queue.NewRegistry().Marshal(job)
		if err != nil {
			return Result{}, fmt.Errorf("encode %s job: %w", name, err)
		}
		if err := enqueueCommandsEnvelope(ctx, client, queuePrefix, queueName, payload.Envelope{ID: "commands-job-" + suffix, Name: jobName, Queue: queueName, Payload: body, MaxTries: 1}); err != nil {
			return Result{}, fmt.Errorf("enqueue %s job: %w", name, err)
		}
	case "worker-tries":
		args = append(args, "--tries=2")
		if err := enqueueRetryableCommandsJob(ctx, client, queuePrefix, queueName, suffix); err != nil {
			return Result{}, fmt.Errorf("enqueue %s job: %w", name, err)
		}
	case "worker-backoff":
		args = append(args, "--tries=2", "--backoff=0,1")
		if err := enqueueRetryableCommandsJob(ctx, client, queuePrefix, queueName, suffix); err != nil {
			return Result{}, fmt.Errorf("enqueue %s job: %w", name, err)
		}
	case "worker-max-jobs":
		args = []string{"queue", "redis", "--queue=" + queueName, "--max-jobs=1", "--sleep=0"}
		readyJobs = 2
	case "worker-max-time":
		args = []string{"queue", "redis", "--queue=" + queueName, "--max-time=1", "--sleep=0"}
	case "worker-retry-after":
		args = []string{"queue", "redis", "--queue=" + queueName, "--once", "--timeout=10", "--retry-after=2"}
		job := &jobs.WorkerJob{TraceID: suffix, Label: "visibility", WaitForCancel: true}
		jobName, err := queue.JobTypeName(job)
		if err != nil {
			return Result{}, fmt.Errorf("identify %s job: %w", name, err)
		}
		body, err := queue.NewRegistry().Marshal(job)
		if err != nil {
			return Result{}, fmt.Errorf("encode %s job: %w", name, err)
		}
		if err := enqueueCommandsEnvelope(ctx, client, queuePrefix, queueName, payload.Envelope{ID: "commands-job-" + suffix, Name: jobName, Queue: queueName, Payload: body, MaxTries: 1}); err != nil {
			return Result{}, fmt.Errorf("enqueue %s job: %w", name, err)
		}
	}
	if readyJobs > 0 {
		for i := range readyJobs {
			if err := enqueueCommandsJob(ctx, client, queuePrefix, queueName, fmt.Sprintf("%s-%d", suffix, i)); err != nil {
				return Result{}, fmt.Errorf("%s prepare job %d: %w", name, i, err)
			}
		}
	}
	if name == "worker-retry-after" {
		return observeRetryAfter(ctx, executable, root, client, environment, args, queuePrefix+":queues:"+queueName+":reserved")
	}
	started := time.Now()
	output, err := invoke(ctx, executable, root, environment, args...)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", name, err)
	}
	switch name {
	case "queue", "queue-work":
		if !strings.Contains(output, "queue worker started") {
			return Result{}, fmt.Errorf("%s output = %q, want worker started", name, output)
		}
		page, err := store.Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
		if err != nil {
			return Result{}, fmt.Errorf("%s list processed failed jobs: %w", name, err)
		}
		if len(page.Items) != 1 || page.Items[0].JobID != "commands-job-"+suffix || !strings.Contains(page.Items[0].Error, "first attempt failure") {
			return Result{}, fmt.Errorf("%s processed jobs = %#v, want one first-attempt failure for %s", name, page.Items, suffix)
		}
	case "queue-failed":
		if !strings.Contains(output, failedID) || !strings.Contains(output, "demonstration failure") {
			return Result{}, fmt.Errorf("%s output = %q, want failed job %s", name, output, failedID)
		}
	case "queue-retry":
		if !strings.Contains(output, "retried failed job: "+failedID) {
			return Result{}, fmt.Errorf("%s output = %q, want retry confirmation", name, output)
		}
		size, err := client.LLen(ctx, queuePrefix+":queues:"+queueName).Result()
		if err != nil {
			return Result{}, fmt.Errorf("%s inspect requeued job: %w", name, err)
		}
		if size != 1 {
			return Result{}, fmt.Errorf("%s ready jobs = %d, want 1", name, size)
		}
		workerOutput, err := invoke(ctx, executable, root, environment, "queue:work", "redis", "--queue="+queueName, "--once", "--sleep=0")
		if err != nil {
			return Result{}, fmt.Errorf("%s consume retried job: %w", name, err)
		}
		if !strings.Contains(workerOutput, "queue worker started") {
			return Result{}, fmt.Errorf("%s worker output = %q, want worker started", name, workerOutput)
		}
		size, err = client.LLen(ctx, queuePrefix+":queues:"+queueName).Result()
		if err != nil {
			return Result{}, fmt.Errorf("%s inspect consumed retry: %w", name, err)
		}
		if size != 0 {
			return Result{}, fmt.Errorf("%s ready jobs after worker = %d, want 0", name, size)
		}
		fallthrough
	case "queue-forget":
		if name == "queue-forget" && !strings.Contains(output, "forgot failed job: "+failedID) {
			return Result{}, fmt.Errorf("%s output = %q, want forget confirmation", name, output)
		}
		_, err := store.Find(ctx, failedID)
		if !errors.Is(err, queue.ErrEmpty) {
			return Result{}, fmt.Errorf("%s failed job lookup error = %v, want %v", name, err, queue.ErrEmpty)
		}
	case "queue-flush":
		page, err := store.Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
		if err != nil || len(page.Items) != 0 {
			return Result{}, fmt.Errorf("%s failed jobs = %#v, error = %v; want empty", name, page.Items, err)
		}
	case "queue-restart":
		if !strings.Contains(output, "queue restart signal sent") {
			return Result{}, fmt.Errorf("%s output = %q, want restart confirmation", name, output)
		}
		keys, err := client.Keys(ctx, "*commands-restart-"+suffix).Result()
		if err != nil || len(keys) != 1 {
			return Result{}, fmt.Errorf("%s Redis restart keys = %#v, error = %v; want one", name, keys, err)
		}
		if err := client.Del(ctx, keys[0]).Err(); err != nil {
			return Result{}, fmt.Errorf("clean %s restart key: %w", name, err)
		}
	case "worker-connection", "worker-queue", "worker-once", "worker-stop-when-empty", "worker-sleep", "worker-timeout",
		"worker-tries", "worker-backoff", "worker-max-jobs", "worker-max-time", "worker-retry-after":
		if !strings.Contains(output, "queue worker started") {
			return Result{}, fmt.Errorf("%s output = %q, want worker start", name, output)
		}
		if readyJobs > 0 {
			want := int64(0)
			if name == "worker-once" || name == "worker-max-jobs" {
				want = 1
			}
			size, err := client.LLen(ctx, queuePrefix+":queues:"+queueName).Result()
			if err != nil || size != want {
				return Result{}, fmt.Errorf("%s remaining jobs = %d, error = %v; want %d", name, size, err, want)
			}
		}
		if name == "worker-sleep" || name == "worker-max-time" {
			if elapsed := time.Since(started); elapsed < time.Second || elapsed > 5*time.Second {
				return Result{}, fmt.Errorf("%s duration = %s, want 1s to 5s", name, elapsed)
			}
		}
		if name == "worker-timeout" {
			page, err := store.Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
			if err != nil || len(page.Items) != 1 || !strings.Contains(strings.ToLower(page.Items[0].Error), "deadline") {
				return Result{}, fmt.Errorf("%s failed jobs = %#v, error = %v; want one timeout failure", name, page.Items, err)
			}
		}
		if name == "worker-tries" || name == "worker-backoff" {
			page, err := store.Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
			if err != nil || len(page.Items) != 0 {
				return Result{}, fmt.Errorf("%s failed jobs = %#v, error = %v; want retry to succeed", name, page.Items, err)
			}
		}
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: strings.TrimSpace(output)}, nil
}

func observeRetryAfter(ctx context.Context, executable, root string, client *redis.Client, environment, args []string, reservedKey string) (Result, error) {
	cmd := exec.Command(executable, args...)
	cmd.Dir = root
	cmd.Env = commandEnvironment(environment)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("start worker-retry-after: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()
	for {
		reserved, err := client.ZRangeWithScores(ctx, reservedKey, 0, 0).Result()
		if err != nil {
			return Result{}, fmt.Errorf("inspect worker-retry-after reservation: %w", err)
		}
		if len(reserved) == 1 {
			remaining := time.Duration(reserved[0].Score-float64(time.Now().UnixMilli())) * time.Millisecond
			if remaining < 500*time.Millisecond || remaining > 2500*time.Millisecond {
				return Result{}, fmt.Errorf("worker-retry-after visibility = %s, want about 2s", remaining)
			}
			return Result{Case: "worker-retry-after", Command: strings.Join(args, " "), Output: fmt.Sprintf("Redis reservation visible after %s", remaining)}, nil
		}
		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf("wait for worker-retry-after reservation: %w", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func enqueueCommandsJob(ctx context.Context, client *redis.Client, prefix, queueName, suffix string) error {
	envelope, err := commandsJobEnvelope(queueName, suffix, true)
	if err != nil {
		return err
	}
	return enqueueCommandsEnvelope(ctx, client, prefix, queueName, envelope)
}

func enqueueRetryableCommandsJob(ctx context.Context, client *redis.Client, prefix, queueName, suffix string) error {
	envelope, err := commandsJobEnvelope(queueName, suffix, true)
	if err != nil {
		return err
	}
	envelope.MaxTries = 0
	return enqueueCommandsEnvelope(ctx, client, prefix, queueName, envelope)
}

func enqueueCommandsEnvelope(ctx context.Context, client *redis.Client, prefix, queueName string, envelope payload.Envelope) error {
	encoded, err := encoding.JSON().Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode commands demo envelope: %w", err)
	}
	key := prefix + ":queues:" + queueName
	pipe := client.TxPipeline()
	pipe.RPush(ctx, key, encoded)
	pipe.LPush(ctx, key+":notify", "1")
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("enqueue commands demo job: %w", err)
	}
	return nil
}

func commandsJobEnvelope(queueName, suffix string, failFirst bool) (payload.Envelope, error) {
	job := &jobs.OperationsJob{TraceID: suffix, Label: "commands", FailFirst: failFirst}
	name, err := queue.JobTypeName(job)
	if err != nil {
		return payload.Envelope{}, fmt.Errorf("identify commands demo job: %w", err)
	}
	body, err := queue.NewRegistry().Marshal(job)
	if err != nil {
		return payload.Envelope{}, fmt.Errorf("encode commands demo job: %w", err)
	}
	return payload.Envelope{
		ID: "commands-job-" + suffix, Name: name, Queue: queueName, Payload: body, MaxTries: 1,
	}, nil
}
