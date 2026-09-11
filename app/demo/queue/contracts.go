package queuedemo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	encryptioncontract "github.com/prismgo/framework/contracts/encryption"
	queuecontract "github.com/prismgo/framework/contracts/queue"
	"github.com/prismgo/framework/encryption"
	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runEncryptionMissingKey(ctx context.Context, connection string) (result Result, err error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo encryption-missing-key uses an inspecting in-memory transport, got %s", connection)
	}
	_, keyErr := encryption.New(encryption.Config{})
	if !errors.Is(keyErr, encryption.ErrInvalidKey) {
		return Result{}, fmt.Errorf("queue demo encryption-missing-key config: got %v, want %w", keyErr, encryption.ErrInvalidKey)
	}

	connector := &demoMemoryConnector{}
	driver := fmt.Sprintf("demo-encryption-missing-%d", time.Now().UnixNano())
	manager, err := queue.NewManager(queue.Config{
		Default: "missing-key",
		Connections: map[string]queue.ConnectionConfig{
			"missing-key": {Driver: driver, Queue: "demo-encryption-missing-key"},
		},
		PayloadEncrypter: demoRejectedEncrypter{err: keyErr},
	}, queue.NewRegistry())
	if err != nil {
		return Result{}, fmt.Errorf("queue demo encryption-missing-key manager: %w", err)
	}
	manager.Extend(driver, func() (queuecontract.Connector, error) { return connector, nil })
	defer func() {
		if closeErr := manager.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("queue demo encryption-missing-key close manager: %w", closeErr)
		}
	}()

	_, dispatchErr := manager.Dispatch(ctx, &jobs.EncryptionJob{
		TraceID: "encryption-missing-key", Label: "must-not-run", Secret: demoEncryptionSecret, Encrypt: true,
	})
	if !errors.Is(dispatchErr, encryption.ErrInvalidKey) {
		return Result{}, fmt.Errorf("queue demo encryption-missing-key dispatch: got %v, want %w", dispatchErr, encryption.ErrInvalidKey)
	}
	if connector.queue != nil {
		if bodies := connector.queue.recordedBodies(); len(bodies) != 0 {
			return Result{}, fmt.Errorf("queue demo encryption-missing-key recorded %d payloads, want 0", len(bodies))
		}
	}
	steps := []string{"app-key:rejected", "dispatch:rejected", "transport:empty"}
	return Result{Case: "encryption-missing-key", Connection: connection, Queue: "demo-encryption-missing-key", Processed: true, Steps: steps}, nil
}

func runCustomQueueContract(ctx context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo custom-queue-contract is hermetic and selected with sync, got %s", connection)
	}
	transport := &demoContractQueue{}
	if err := transport.Push(ctx, "low", queuecontract.Payload("low-one")); err != nil {
		return Result{}, fmt.Errorf("queue demo custom queue push: %w", err)
	}
	if err := transport.Later(ctx, "high", queuecontract.Payload("high-one"), time.Second); err != nil {
		return Result{}, fmt.Errorf("queue demo custom queue later: %w", err)
	}
	bulk, err := transport.Bulk(ctx, "low", []queuecontract.Payload{[]byte("low-two"), []byte("low-three")})
	if err != nil || bulk.Accepted != 2 {
		return Result{}, fmt.Errorf("queue demo custom queue bulk: accepted=%d err=%v, want accepted=2", bulk.Accepted, err)
	}
	reserved, err := transport.Pop(ctx, []string{"high", "low"}, queuecontract.PopNoWait)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom queue priority pop: %w", err)
	}
	jobID := reserved.ID()
	if string(reserved.Payload()) != "high-one" {
		return Result{}, fmt.Errorf("queue demo custom queue priority payload = %q, want %q", reserved.Payload(), "high-one")
	}
	if err := reserved.Delete(ctx); err != nil {
		return Result{}, fmt.Errorf("queue demo custom queue delete: %w", err)
	}
	if err := transport.Clear(ctx, "low"); err != nil {
		return Result{}, fmt.Errorf("queue demo custom queue clear: %w", err)
	}
	if size, err := transport.Size(ctx, "low"); err != nil || size != 0 {
		return Result{}, fmt.Errorf("queue demo custom queue size after clear: size=%d err=%v, want 0", size, err)
	}
	if err := transport.Close(); err != nil {
		return Result{}, fmt.Errorf("queue demo custom queue close: %w", err)
	}
	if transport.closeCount() != 1 {
		return Result{}, fmt.Errorf("queue demo custom queue close calls = %d, want 1", transport.closeCount())
	}
	steps := []string{"push:accepted", "later:accepted", "bulk:accepted=2", "pop:priority", "clear:size=0", "close:released"}
	return Result{Case: "custom-queue-contract", Connection: "custom", Queue: "high", JobID: jobID, Processed: true, Steps: steps}, nil
}

func runCustomReservedJob(ctx context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo custom-reserved-job is hermetic and selected with sync, got %s", connection)
	}
	transport := &demoContractQueue{}
	if err := transport.Push(ctx, "jobs", queuecontract.Payload("payload")); err != nil {
		return Result{}, fmt.Errorf("queue demo custom reserved push: %w", err)
	}
	first, err := transport.Pop(ctx, []string{"jobs"}, queuecontract.PopNoWait)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom reserved first pop: %w", err)
	}
	jobID := first.ID()
	copyOfPayload := first.Payload()
	copyOfPayload[0] = 'X'
	if first.Name() != "demo.contract.Job" || string(first.Payload()) != "payload" || first.Attempts() != 1 {
		return Result{}, fmt.Errorf("queue demo custom reserved first state: name=%q attempts=%d payload=%q, want name=%q attempts=1 payload=%q", first.Name(), first.Attempts(), first.Payload(), "demo.contract.Job", "payload")
	}
	if err := first.Release(ctx, 2*time.Second); err != nil {
		return Result{}, fmt.Errorf("queue demo custom reserved release: %w", err)
	}
	second, err := transport.Pop(ctx, []string{"jobs"}, queuecontract.PopNoWait)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom reserved second pop: %w", err)
	}
	if second.ID() != jobID || second.Attempts() != 2 {
		return Result{}, fmt.Errorf("queue demo custom reserved second state: id=%q attempts=%d, want id=%q attempts=2", second.ID(), second.Attempts(), jobID)
	}
	if err := second.Delete(ctx); err != nil {
		return Result{}, fmt.Errorf("queue demo custom reserved delete: %w", err)
	}
	if transport.deletedCount() != 1 {
		return Result{}, fmt.Errorf("queue demo custom reserved deletes = %d, want 1", transport.deletedCount())
	}
	steps := []string{"metadata:read", "payload:copied", "attempts:1", "release:2s", "attempts:2", "delete:acknowledged"}
	return Result{Case: "custom-reserved-job", Connection: "custom", Queue: "jobs", JobID: jobID, Processed: true, Steps: steps}, nil
}

func runCustomPopSession(ctx context.Context, connection string) (result Result, err error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo custom-pop-session is hermetic and selected with sync, got %s", connection)
	}
	transport := &demoPopSessionQueue{demoContractQueue: &demoContractQueue{}}
	manager, err := newDemoContractManager("pop-session", transport)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		if closeErr := manager.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("queue demo custom-pop-session close manager: %w", closeErr)
		}
	}()

	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("custom-pop-session-%d", runID)
	queueName := fmt.Sprintf("demo-custom-pop-session-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "job:handled"}, queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom pop session dispatch: %w", err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{Connection: "custom", Queues: []string{queueName}, Once: true}); err != nil {
		return Result{}, fmt.Errorf("queue demo custom pop session worker: %w", err)
	}
	sessions, pops, closed := transport.sessionCounts()
	trace := jobs.TakeTrace(traceID)
	if sessions != 1 || pops != 1 || closed != 1 || !advancedContainsStep(trace, "job:handled") {
		return Result{}, fmt.Errorf("queue demo custom pop session: sessions=%d pops=%d closed=%d trace=%v, want 1/1/1 and handled", sessions, pops, closed, trace)
	}
	steps := []string{"session:created", "session:pop", "job:handled", "session:closed"}
	return Result{Case: "custom-pop-session", Connection: "custom", Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runCustomConsumerIntent(ctx context.Context, connection string) (result Result, err error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo custom-consumer-intent is hermetic and selected with sync, got %s", connection)
	}
	transport := &demoConsumerIntentQueue{demoContractQueue: &demoContractQueue{}}
	manager, err := newDemoContractManager("consumer-intent", transport)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		if closeErr := manager.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("queue demo custom-consumer-intent close manager: %w", closeErr)
		}
	}()

	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("custom-consumer-intent-%d", runID)
	queueName := fmt.Sprintf("demo-custom-consumer-intent-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "job:handled"}, queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom consumer intent dispatch: %w", err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{Connection: "custom", Queues: []string{queueName}, Once: true}); err != nil {
		return Result{}, fmt.Errorf("queue demo custom consumer intent worker: %w", err)
	}
	acquired, released, queues := transport.intentState()
	trace := jobs.TakeTrace(traceID)
	if acquired != 1 || released != 1 || len(queues) != 1 || queues[0] != queueName || !advancedContainsStep(trace, "job:handled") {
		return Result{}, fmt.Errorf("queue demo custom consumer intent: acquired=%d released=%d queues=%v trace=%v", acquired, released, queues, trace)
	}
	steps := []string{"intent:acquired", "queues:received", "job:handled", "intent:released"}
	return Result{Case: "custom-consumer-intent", Connection: "custom", Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func newDemoContractManager(label string, transport queuecontract.Queue) (*queue.Manager, error) {
	driver := fmt.Sprintf("demo-%s-%d", label, time.Now().UnixNano())
	manager, err := queue.NewManager(queue.Config{
		Default: "custom",
		Connections: map[string]queue.ConnectionConfig{
			"custom": {Driver: driver, Queue: "default"},
		},
	}, queue.NewRegistry())
	if err != nil {
		return nil, fmt.Errorf("queue demo %s manager: %w", label, err)
	}
	manager.Extend(driver, func() (queuecontract.Connector, error) {
		return demoStaticConnector{transport: transport}, nil
	})
	return manager, nil
}

type demoRejectedEncrypter struct {
	err error
}

func (e demoRejectedEncrypter) Encrypt(context.Context, []byte) ([]byte, error) {
	return nil, e.err
}

func (e demoRejectedEncrypter) Decrypt(context.Context, []byte) ([]byte, error) {
	return nil, e.err
}

type demoStaticConnector struct {
	transport queuecontract.Queue
}

func (c demoStaticConnector) Connect(context.Context, string, queuecontract.ConnectorConfig) (queuecontract.Queue, error) {
	return c.transport, nil
}

type demoContractItem struct {
	id       string
	name     string
	queue    string
	body     queuecontract.Payload
	attempts int
}

// demoContractQueue is a small transport used to exercise the public custom-driver contracts.
type demoContractQueue struct {
	mu         sync.Mutex
	items      []demoContractItem
	nextID     int
	deleted    int
	closeCalls int
	closed     bool
}

func (q *demoContractQueue) Push(_ context.Context, queueName string, body queuecontract.Payload) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return queue.ErrConnectionClosed
	}
	q.nextID++
	q.items = append(q.items, demoContractItem{
		id: fmt.Sprintf("contract-%d", q.nextID), name: "demo.contract.Job", queue: queueName,
		body: append(queuecontract.Payload(nil), body...),
	})
	return nil
}

func (q *demoContractQueue) Later(ctx context.Context, queueName string, body queuecontract.Payload, _ time.Duration) error {
	return q.Push(ctx, queueName, body)
}

func (q *demoContractQueue) Bulk(ctx context.Context, queueName string, bodies []queuecontract.Payload) (queuecontract.BulkResult, error) {
	accepted := 0
	for _, body := range bodies {
		if err := q.Push(ctx, queueName, body); err != nil {
			return queuecontract.BulkResult{Accepted: accepted}, err
		}
		accepted++
	}
	return queuecontract.BulkResult{Accepted: accepted}, nil
}

func (q *demoContractQueue) Pop(_ context.Context, queues []string, _ ...queuecontract.PopWaitMode) (queuecontract.ReservedJob, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil, queue.ErrConnectionClosed
	}
	for _, queueName := range queues {
		for index, item := range q.items {
			if item.queue != queueName {
				continue
			}
			q.items = append(q.items[:index], q.items[index+1:]...)
			item.attempts++
			return &demoContractReservedJob{owner: q, item: item}, nil
		}
	}
	return nil, queue.ErrEmpty
}

func (q *demoContractQueue) Size(_ context.Context, queueName string) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var size int64
	for _, item := range q.items {
		if item.queue == queueName {
			size++
		}
	}
	return size, nil
}

func (q *demoContractQueue) Clear(_ context.Context, queueName string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	kept := q.items[:0]
	for _, item := range q.items {
		if item.queue != queueName {
			kept = append(kept, item)
		}
	}
	q.items = kept
	return nil
}

func (q *demoContractQueue) Close() error {
	q.mu.Lock()
	q.closeCalls++
	q.closed = true
	q.mu.Unlock()
	return nil
}

func (q *demoContractQueue) deletedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.deleted
}

func (q *demoContractQueue) closeCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.closeCalls
}

type demoContractReservedJob struct {
	owner *demoContractQueue
	item  demoContractItem
}

func (j *demoContractReservedJob) ID() string   { return j.item.id }
func (j *demoContractReservedJob) Name() string { return j.item.name }
func (j *demoContractReservedJob) Payload() queuecontract.Payload {
	return append(queuecontract.Payload(nil), j.item.body...)
}
func (j *demoContractReservedJob) Attempts() int { return j.item.attempts }
func (j *demoContractReservedJob) Delete(context.Context) error {
	j.owner.mu.Lock()
	j.owner.deleted++
	j.owner.mu.Unlock()
	return nil
}
func (j *demoContractReservedJob) Release(_ context.Context, _ time.Duration) error {
	j.owner.mu.Lock()
	j.owner.items = append(j.owner.items, j.item)
	j.owner.mu.Unlock()
	return nil
}

type demoPopSessionQueue struct {
	*demoContractQueue
	sessionsMu     sync.Mutex
	sessions       int
	pops           int
	closedSessions int
}

func (q *demoPopSessionQueue) NewPopSession() queuecontract.Queue {
	q.sessionsMu.Lock()
	q.sessions++
	q.sessionsMu.Unlock()
	return &demoPopSessionView{demoContractQueue: q.demoContractQueue, owner: q}
}

func (q *demoPopSessionQueue) sessionCounts() (int, int, int) {
	q.sessionsMu.Lock()
	defer q.sessionsMu.Unlock()
	return q.sessions, q.pops, q.closedSessions
}

type demoPopSessionView struct {
	*demoContractQueue
	owner *demoPopSessionQueue
}

func (v *demoPopSessionView) Pop(ctx context.Context, queues []string, wait ...queuecontract.PopWaitMode) (queuecontract.ReservedJob, error) {
	v.owner.sessionsMu.Lock()
	v.owner.pops++
	v.owner.sessionsMu.Unlock()
	return v.demoContractQueue.Pop(ctx, queues, wait...)
}

func (v *demoPopSessionView) Close() error {
	v.owner.sessionsMu.Lock()
	v.owner.closedSessions++
	v.owner.sessionsMu.Unlock()
	return nil
}

type demoConsumerIntentQueue struct {
	*demoContractQueue
	intentMu     sync.Mutex
	acquired     int
	released     int
	intentQueues []string
}

func (q *demoConsumerIntentQueue) AcquireConsumerIntent(queues []string) (func() error, error) {
	q.intentMu.Lock()
	q.acquired++
	q.intentQueues = append([]string(nil), queues...)
	q.intentMu.Unlock()
	return func() error {
		q.intentMu.Lock()
		q.released++
		q.intentMu.Unlock()
		return nil
	}, nil
}

func (q *demoConsumerIntentQueue) intentState() (int, int, []string) {
	q.intentMu.Lock()
	defer q.intentMu.Unlock()
	return q.acquired, q.released, append([]string(nil), q.intentQueues...)
}

var (
	_ encryptioncontract.Encrypter       = demoRejectedEncrypter{}
	_ queuecontract.Connector            = demoStaticConnector{}
	_ queuecontract.Queue                = (*demoContractQueue)(nil)
	_ queuecontract.ReservedJob          = (*demoContractReservedJob)(nil)
	_ queuecontract.PopSessionProvider   = (*demoPopSessionQueue)(nil)
	_ queuecontract.ConsumerIntentLeaser = (*demoConsumerIntentQueue)(nil)
)
