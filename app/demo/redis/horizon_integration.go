package redisdemo

import (
	"context"
	"fmt"
	"time"

	"github.com/prismgo/horizon"
)

// openHorizonStore boots an isolated application and opens a Redis-backed Horizon
// store under a unique prefix, returning a cleanup that removes every stored key.
func openHorizonStore() (*horizon.RedisStore, string, func(), error) {
	prefix := fmt.Sprintf("prismgo_demo_redis_horizon_%d", time.Now().UnixNano())
	_, base, err := openLiveWith(liveOptions{})
	if err != nil {
		return nil, "", nil, err
	}
	store, err := horizon.NewRedisStore(
		horizon.RedisOptions{Connection: "default"},
		horizon.StoreOptions{Prefix: prefix, HeartbeatTTL: time.Minute, Encoding: "msgpack"},
	)
	if err != nil {
		base()
		return nil, "", nil, fmt.Errorf("create horizon redis store: %w", err)
	}
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if client, err := redisConnectionClient("default"); err == nil {
			keys, err := client.Keys(ctx, prefix+"*").Result()
			if err == nil && len(keys) > 0 {
				_ = client.Del(ctx, keys...).Err()
			}
		}
		base()
	}
	return store, prefix, cleanup, nil
}

// horizonProcessesScenario verifies master, supervisor, and worker heartbeats.
func horizonProcessesScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC()
	master := horizon.MasterState{
		ID: "master-1", Host: "demo", PID: 1001, Status: horizon.MasterRunning,
		StartedAt: now, LastHeartbeatAt: now, SupervisorCount: 1, Environment: "local",
	}
	acquired, err := store.AcquireMasterLease(ctx, master)
	if err != nil || !acquired {
		return "", fmt.Errorf("acquire master lease = %t, err = %v; want true", acquired, err)
	}
	if err := store.HeartbeatMaster(ctx, master); err != nil {
		return "", fmt.Errorf("heartbeat master: %w", err)
	}
	supervisor := horizon.SupervisorState{
		Name: "sup-1", Host: "demo", PID: 1002, MasterID: "master-1", Environment: "local",
		Status: horizon.SupervisorRunning, StartedAt: now, LastHeartbeatAt: now,
		WorkerCount: 1, Connection: "redis", Queues: []string{"default"},
	}
	if acquired, err := store.AcquireSupervisorLease(ctx, supervisor); err != nil || !acquired {
		return "", fmt.Errorf("acquire supervisor lease = %t, err = %v; want true", acquired, err)
	}
	worker := horizon.WorkerState{
		ID: "worker-1", Supervisor: "sup-1", Environment: "local", Host: "demo", PID: 1003,
		Status: horizon.WorkerIdle, StartedAt: now, LastHeartbeatAt: now,
	}
	if err := store.HeartbeatWorker(ctx, worker); err != nil {
		return "", fmt.Errorf("heartbeat worker: %w", err)
	}
	masters, err := store.Masters(ctx, now)
	if err != nil {
		return "", fmt.Errorf("read masters: %w", err)
	}
	supervisors, err := store.Supervisors(ctx, now)
	if err != nil {
		return "", fmt.Errorf("read supervisors: %w", err)
	}
	workers, err := store.Workers(ctx, now)
	if err != nil {
		return "", fmt.Errorf("read workers: %w", err)
	}
	if len(masters) != 1 || len(supervisors) != 1 || len(workers) != 1 {
		return "", fmt.Errorf("masters/supervisors/workers = %d/%d/%d, want 1/1/1", len(masters), len(supervisors), len(workers))
	}
	staleSupervisors, err := store.Supervisors(ctx, now.Add(2*time.Minute))
	if err != nil {
		return "", fmt.Errorf("read stale supervisors: %w", err)
	}
	if len(staleSupervisors) != 1 || staleSupervisors[0].Status != horizon.SupervisorStale {
		return "", fmt.Errorf("stale supervisor status = %#v, want stale", staleSupervisors)
	}
	return fmt.Sprintf("masters=%d supervisors=%d workers=%d stale=%t",
		len(masters), len(supervisors), len(workers), staleSupervisors[0].Status == horizon.SupervisorStale), nil
}

// horizonControlScenario verifies global, supervisor, and terminate control flags.
func horizonControlScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC()
	if err := store.SetGlobalPaused(ctx, true); err != nil {
		return "", fmt.Errorf("set global paused: %w", err)
	}
	if err := store.SetSupervisorPaused(ctx, "sup-control", true); err != nil {
		return "", fmt.Errorf("set supervisor paused: %w", err)
	}
	supervisor := horizon.SupervisorState{
		Name: "sup-control", Host: "demo", PID: 2001, Environment: "local",
		Status: horizon.SupervisorRunning, StartedAt: now, LastHeartbeatAt: now,
		Connection: "redis", Queues: []string{"default"},
	}
	if err := store.HeartbeatSupervisor(ctx, supervisor); err != nil {
		return "", fmt.Errorf("heartbeat control supervisor: %w", err)
	}
	control, err := store.Control(ctx)
	if err != nil {
		return "", fmt.Errorf("read control: %w", err)
	}
	if !control.GlobalPaused || !control.PausedSupervisors["sup-control"] {
		return "", fmt.Errorf("control = %#v, want global and supervisor paused", control)
	}
	paused, err := store.StatusSnapshot(ctx, now)
	if err != nil {
		return "", fmt.Errorf("read paused snapshot: %w", err)
	}
	if err := store.RequestTerminate(ctx, now, true); err != nil {
		return "", fmt.Errorf("request terminate: %w", err)
	}
	terminating, err := store.StatusSnapshot(ctx, now)
	if err != nil {
		return "", fmt.Errorf("read terminating snapshot: %w", err)
	}
	if err := store.ClearTerminateRequest(ctx); err != nil {
		return "", fmt.Errorf("clear terminate: %w", err)
	}
	cleared, err := store.Control(ctx)
	if err != nil {
		return "", fmt.Errorf("read cleared control: %w", err)
	}
	return fmt.Sprintf("paused=%s supervisor-paused=%t terminating=%s cleared=%t",
		paused.Status, control.PausedSupervisors["sup-control"], terminating.Status, cleared.TerminateRequestedAt.IsZero()), nil
}

// horizonMetricsScenario appends and reads Redis event metric windows and rollups.
func horizonMetricsScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC().Truncate(time.Minute)
	windows := []horizon.EventMetricWindow{{
		WindowStart: now, WindowEnd: now.Add(time.Minute), Connection: "redis", Queue: "default",
		JobName: "demo", Processed: 5, Failed: 1, RuntimeMS: 250, SampleCount: 6,
		RuntimeSampleCount: 5, EffectiveSampleRate: 1, Quality: horizon.EventMetricQualityExact,
	}}
	if err := store.AppendEventMetricWindows(ctx, windows, time.Hour); err != nil {
		return "", fmt.Errorf("append metric windows: %w", err)
	}
	page, err := store.EventMetricWindows(ctx, horizon.EventMetricWindowQuery{Page: horizon.PageRequest{Page: 1, PageSize: 10}})
	if err != nil {
		return "", fmt.Errorf("read metric windows: %w", err)
	}
	var processed int64
	for _, item := range page.Items {
		processed += item.Processed
	}
	rollups, err := store.EventMetricRollupWindows(ctx, horizon.EventMetricWindowQuery{})
	if err != nil {
		return "", fmt.Errorf("read metric rollups: %w", err)
	}
	if page.Total != 1 || processed != 5 || len(rollups) != 1 || rollups[0].Processed != 5 {
		return "", fmt.Errorf("metrics = windows:%d processed:%d rollups:%d/%d, want 1/5/1/5",
			page.Total, processed, len(rollups), rollupProcessed(rollups))
	}
	return fmt.Sprintf("windows=%d processed=%d rollups=%d", page.Total, processed, len(rollups)), nil
}

// rollupProcessed returns the processed count of the first rollup window, or -1 when absent.
func rollupProcessed(rollups []horizon.EventMetricWindow) int64 {
	if len(rollups) == 0 {
		return -1
	}
	return rollups[0].Processed
}

// horizonQueueLengthsScenario saves and reads a Redis queue length snapshot.
func horizonQueueLengthsScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	snapshot := horizon.QueueLengthSnapshot{
		CapturedAt: time.Now().UTC(),
		Queues: []horizon.QueueLengthBucket{
			{Connection: "redis", Queue: "default", Size: 3},
			{Connection: "redis", Queue: "high", Size: 0},
		},
	}
	if err := store.SaveQueueLengthSnapshot(ctx, snapshot); err != nil {
		return "", fmt.Errorf("save queue length snapshot: %w", err)
	}
	got, err := store.QueueLengthSnapshot(ctx)
	if err != nil {
		return "", fmt.Errorf("read queue length snapshot: %w", err)
	}
	if len(got.Queues) != 2 {
		return "", fmt.Errorf("queue length entries = %d, want 2", len(got.Queues))
	}
	return fmt.Sprintf("queues=%d first=%d empty=%t", len(got.Queues), got.Queues[0].Size, got.Queues[1].Size == 0), nil
}

// horizonSummariesScenario saves and reads Redis batch-safe summaries.
func horizonSummariesScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC()
	summary := horizon.BatchSummary{
		ID: "batch-1", Name: "demo-batch", Status: horizon.BatchStatusRunning,
		Total: 3, Pending: 1, Processed: 2, CreatedAt: now, UpdatedAt: now,
	}
	if err := store.SaveBatchSummary(ctx, summary); err != nil {
		return "", fmt.Errorf("save batch summary: %w", err)
	}
	got, ok, err := store.Batch(ctx, "batch-1")
	if err != nil || !ok {
		return "", fmt.Errorf("read batch summary ok=%t err=%v", ok, err)
	}
	page, err := store.BatchesPage(ctx, "", horizon.PageRequest{Page: 1, PageSize: 10})
	if err != nil {
		return "", fmt.Errorf("page batch summaries: %w", err)
	}
	if got.Status != horizon.BatchStatusRunning || got.Total != 3 || page.Total != 1 {
		return "", fmt.Errorf("batch = %s/%d page:%d, want running/3/1", got.Status, got.Total, page.Total)
	}
	return fmt.Sprintf("batch=%s total=%d page=%d", got.Status, got.Total, page.Total), nil
}

// horizonJobDiagnosticsScenario saves and reads high-value job diagnostics.
func horizonJobDiagnosticsScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC()
	detail := horizon.HighValueJobDetail{
		ID: "detail-1", Kind: horizon.HighValueDetailFailed, Connection: "redis", Queue: "default",
		JobID: "job-1", JobName: "demo", RuntimeMS: 1200, ErrorSummary: "boom", OccurredAt: now,
	}
	if err := store.SaveHighValueDetails(ctx, []horizon.HighValueJobDetail{detail}, time.Hour); err != nil {
		return "", fmt.Errorf("save high value details: %w", err)
	}
	page, err := store.HighValueDetails(ctx, horizon.HighValueDetailQuery{
		Page: horizon.PageRequest{Page: 1, PageSize: 10}, Kind: horizon.HighValueDetailFailed,
	})
	if err != nil {
		return "", fmt.Errorf("read high value details: %w", err)
	}
	single, ok, err := store.HighValueDetail(ctx, "detail-1")
	if err != nil || !ok {
		return "", fmt.Errorf("read high value detail ok=%t err=%v", ok, err)
	}
	if page.Total != 1 || single.Kind != horizon.HighValueDetailFailed {
		return "", fmt.Errorf("high value details = total:%d kind:%s, want 1/failed", page.Total, single.Kind)
	}
	return fmt.Sprintf("details=%d kind=%s found=true", page.Total, single.Kind), nil
}

// horizonObservabilityScenario saves and reads Redis observability diagnostics.
func horizonObservabilityScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	diagnostic := horizon.ObservabilityDiagnostic{
		Reason: horizon.MemoryDropBufferFull, Count: 3, ObservedAt: time.Now().UTC(),
		Description: "buffer full", Gap: horizon.ObservabilityGapQuantifiable,
	}
	if err := store.SaveObservabilityDiagnostics(ctx, []horizon.ObservabilityDiagnostic{diagnostic}, time.Hour); err != nil {
		return "", fmt.Errorf("save observability diagnostics: %w", err)
	}
	page, err := store.ObservabilityDiagnostics(ctx, horizon.PageRequest{Page: 1, PageSize: 10})
	if err != nil {
		return "", fmt.Errorf("read observability diagnostics: %w", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Reason != horizon.MemoryDropBufferFull {
		return "", fmt.Errorf("observability diagnostics = %#v, want one buffer_full", page.Items)
	}
	return fmt.Sprintf("diagnostics=%d reason=%s", page.Total, page.Items[0].Reason), nil
}

// horizonOrphansScenario records, queries, and forgets orphan processes.
func horizonOrphansScenario() (string, error) {
	store, _, cleanup, err := openHorizonStore()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC()
	if err := store.RecordOrphanProcess(ctx, "master-1", 4242, now.Add(-10*time.Minute)); err != nil {
		return "", fmt.Errorf("record orphan process: %w", err)
	}
	all, err := store.OrphanProcesses(ctx, "master-1")
	if err != nil {
		return "", fmt.Errorf("read orphan processes: %w", err)
	}
	older, err := store.OrphanProcessesOlderThan(ctx, "master-1", 5*time.Minute, now)
	if err != nil {
		return "", fmt.Errorf("read older orphan processes: %w", err)
	}
	if err := store.ForgetOrphanProcess(ctx, "master-1", 4242); err != nil {
		return "", fmt.Errorf("forget orphan process: %w", err)
	}
	after, err := store.OrphanProcesses(ctx, "master-1")
	if err != nil {
		return "", fmt.Errorf("read orphan processes after forget: %w", err)
	}
	if len(all) != 1 || len(older) != 1 || len(after) != 0 {
		return "", fmt.Errorf("orphans = all:%d older:%d after:%d, want 1/1/0", len(all), len(older), len(after))
	}
	return fmt.Sprintf("orphans=%d older=%d forgotten=true", len(all), len(older)), nil
}
