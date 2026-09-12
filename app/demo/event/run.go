// Package eventdemo contains executable event system examples.
package eventdemo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	eventcontract "github.com/prismgo/framework/contracts/event"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/queue"
)

// Result records an observable scenario outcome.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

type message struct {
	EventName string `json:"event_name"`
	Payload   string `json:"payload"`
}

func (m message) Name() string { return m.EventName }

type registeredMessage struct {
	Payload string `json:"payload"`
}

func (*registeredMessage) Name() string { return "demo.event.registered" }

type emptyNameMessage struct{}

func (*emptyNameMessage) Name() string { return "" }

type counter struct{ count *atomic.Int32 }

func (c counter) Handle(_ context.Context, _ event.Event) error { c.count.Add(1); return nil }

type queuedCounter struct{ counter }

func (queuedCounter) ShouldQueue() bool { return true }

type routedCounter struct {
	counter
	connection, queueName string
	delay                 time.Duration
	tries                 int
	backoff               []time.Duration
	timeout               time.Duration
	received              *string
}

func (r routedCounter) Handle(ctx context.Context, ev event.Event) error {
	incoming, ok := ev.(*registeredMessage)
	if !ok {
		return fmt.Errorf("queued event = %T, want *registeredMessage", ev)
	}
	if r.received != nil {
		*r.received = incoming.Payload
	}
	return r.counter.Handle(ctx, ev)
}

func (r routedCounter) ShouldQueue() bool             { return true }
func (r routedCounter) QueueConnection() string       { return r.connection }
func (r routedCounter) QueueName() string             { return r.queueName }
func (r routedCounter) QueueDelay() time.Duration     { return r.delay }
func (r routedCounter) QueueTries() int               { return r.tries }
func (r routedCounter) QueueBackoff() []time.Duration { return r.backoff }
func (r routedCounter) QueueTimeout() time.Duration   { return r.timeout }

type subscriber struct{ count *atomic.Int32 }

func (s subscriber) Subscribe(bus eventcontract.Dispatcher) {
	bus.Listen("demo.event.created", counter{count: s.count})
	bus.ListenFunc("demo.event.updated", func(context.Context, event.Event) error { s.count.Add(1); return nil })
}

// Run executes a documented event scenario using the current application.
func Run(ctx context.Context, name, connection string) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if connection == "" {
		connection = "sync"
	}
	if connection != "sync" && connection != "redis" {
		return Result{}, fmt.Errorf("event demo: unsupported connection %q", connection)
	}
	result := Result{Case: name}
	var err error
	switch name {
	case "architecture":
		var bus eventcontract.Dispatcher = event.New()
		result.Value = fmt.Sprintf("%T implements event.Dispatcher", bus)
	case "event-definition":
		var ev event.Event = message{EventName: "demo.event.created"}
		result.Value = ev.Name()
	case "event-naming":
		result.Value = message{EventName: "demo.event.created"}.Name() + "," + message{EventName: "demo.event.updated"}.Name()
	case "payload-boundaries":
		data, marshalErr := json.Marshal(registeredMessage{Payload: "serializable"})
		if marshalErr != nil {
			return Result{}, fmt.Errorf("marshal event: %w", marshalErr)
		}
		var restored registeredMessage
		if err := json.Unmarshal(data, &restored); err != nil {
			return Result{}, fmt.Errorf("restore event: %w", err)
		}
		result.Value = restored.Payload
	case "queue-prerequisites":
		var marker event.ShouldQueue = queuedCounter{}
		result.Value = fmt.Sprintf("ShouldQueue=%t; factory=%s; worker=required", marker.ShouldQueue(), (&registeredMessage{}).Name())
	case "event-factory":
		event.RegisterEvent[*registeredMessage]()
		result.Value = "registered:" + (&registeredMessage{}).Name()
	case "event-factory-validation":
		result.Value = factoryValidation()
	case "queue-routing-options", "queue-retry-options":
		listener := routedCounter{connection: connection, queueName: "demo-event", tries: 3, backoff: []time.Duration{time.Second, 2 * time.Second}, timeout: 5 * time.Second}
		if name == "queue-routing-options" {
			result.Value = fmt.Sprintf("connection=%s; queue=%s; delay=%s", listener.QueueConnection(), listener.QueueName(), listener.QueueDelay())
		} else {
			result.Value = fmt.Sprintf("tries=%d; backoff=%v; timeout=%s", listener.QueueTries(), listener.QueueBackoff(), listener.QueueTimeout())
		}
		observed, queueErr := runQueued(ctx, name, connection)
		if queueErr != nil {
			return Result{}, fmt.Errorf("event demo %s: %w", name, queueErr)
		}
		result.Value += "; " + observed
	case "should-queue", "queued-wrapper", "queued-sync", "queued-redis", "raw-queued-event":
		result.Value, err = runQueued(ctx, name, connection)
	case "facade":
		result.Value = runFacade(ctx)
	case "async":
		result.Value, err = runAsync(ctx)
	case "async-durability", "app-lifecycle", "provider-lifecycle", "server-lifecycle", "request-lifecycle", "request-finished-ordering", "console-lifecycle", "vendor-publish-event", "listener-error-isolation", "listener-panic-isolation", "async-failure-isolation", "queued-failure-handling", "consistency-boundary", "isolated-testing", "queued-sync-testing", "queued-worker-testing", "event-interface", "listener-interface", "listener-func-interface", "dispatcher-interface", "subscriber-interface", "should-queue-interface", "async-listener-interface", "queue-options-interface", "provider-register", "provider-boot", "laravel-compatibility":
		result.Value, err = runSecondBatch(ctx, name, connection)
	default:
		result.Value, err = runSynchronous(ctx, name)
	}
	if err != nil {
		return Result{}, fmt.Errorf("event demo %s: %w", name, err)
	}
	return result, nil
}

func factoryValidation() string {
	rejected := 0
	for _, check := range []func(){func() { event.RegisterEvent[message]() }, func() { event.RegisterEvent[*emptyNameMessage]() }} {
		func() {
			defer func() {
				if recover() != nil {
					rejected++
				}
			}()
			check()
		}()
	}
	return fmt.Sprintf("rejected=%d", rejected)
}

func runSynchronous(ctx context.Context, name string) (string, error) {
	bus := event.New()
	var count atomic.Int32
	ev := message{EventName: "demo.event.created", Payload: "ok"}
	listen := func(pattern string) { bus.Listen(pattern, counter{count: &count}) }
	switch name {
	case "manual-registration", "struct-registration", "struct-listener":
		listen(ev.Name())
	case "closure-listener", "listener-func":
		bus.ListenFunc(ev.Name(), func(_ context.Context, incoming event.Event) error {
			if incoming.Name() == ev.Name() {
				count.Add(1)
			}
			return nil
		})
	case "wildcard-prefix":
		listen("demo.event.*")
		bus.Dispatch(ctx, message{EventName: "other.event.created"})
	case "wildcard-all":
		listen("*")
	case "wildcard-exact-operations":
		listen("demo.event.*")
		listen(ev.Name())
		if !bus.Has(ev.Name()) {
			return "", errors.New("exact listener missing before Forget")
		}
		bus.Forget(ev.Name())
		if bus.Has(ev.Name()) {
			return "", errors.New("exact listener remains after Forget")
		}
	case "propagation", "dispatch", "dispatch-isolation":
		if name == "propagation" || name == "dispatch-isolation" {
			bus.ListenFunc(ev.Name(), func(context.Context, event.Event) error { return errors.New("isolated listener error") })
		}
		listen(ev.Name())
		listen(ev.Name())
	case "nil-dispatch":
		listen(ev.Name())
		bus.Dispatch(nil, nil)
		bus.Dispatch(nil, ev)
		return fmt.Sprintf("nil-safe; handled=%d", count.Load()), nil
	case "subscriber", "subscriber-registration":
		bus.Subscribe(subscriber{count: &count})
		bus.Dispatch(ctx, message{EventName: "demo.event.updated"})
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
	bus.Dispatch(ctx, ev)
	return fmt.Sprintf("handled=%d; exact=%t", count.Load(), bus.Has(ev.Name())), nil
}

func runQueued(ctx context.Context, name, connection string) (string, error) {
	if name == "queued-redis" && connection != "redis" {
		return "", errors.New("queued-redis requires --connection=redis")
	}
	manager := queue.Resolve()
	if manager == nil {
		return "", errors.New("queue manager is nil")
	}
	event.RegisterEvent[*registeredMessage]()
	bus := event.New()
	var count atomic.Int32
	observed := ""
	received := ""
	listener := event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		observed = fmt.Sprintf("%T:%s", ev, ev.Name())
		count.Add(1)
		return nil
	})
	eventName := "demo.event.registered"
	workerQueue := fmt.Sprintf("demo-event-%d", time.Now().UnixNano())
	switch name {
	case "raw-queued-event":
		eventName = "demo.event.raw"
		bus.Listen(eventName, event.Queued(listener))
	case "should-queue":
		bus.Listen(eventName, queuedCounter{counter{count: &count}})
	case "queued-redis":
		bus.Listen(eventName, routedCounter{counter: counter{count: &count}, connection: "redis", queueName: workerQueue, received: &received})
	case "queue-routing-options", "queue-retry-options":
		bus.Listen(eventName, routedCounter{counter: counter{count: &count}, connection: connection, queueName: "demo-event", tries: 3, backoff: []time.Duration{time.Second, 2 * time.Second}, timeout: 5 * time.Second, received: &received})
	default:
		bus.Listen(eventName, event.Queued(listener))
	}
	var outgoing event.Event = &registeredMessage{Payload: "queued"}
	if name == "raw-queued-event" {
		outgoing = message{EventName: eventName, Payload: "raw"}
	}
	bus.Dispatch(ctx, outgoing)
	if name == "queued-redis" {
		if count.Load() != 0 {
			return "", fmt.Errorf("count before worker = %d, want 0", count.Load())
		}
		workerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := queue.NewWorker(manager).Work(workerCtx, queue.WorkerOptions{Connection: "redis", Queues: []string{workerQueue}, Once: true}); err != nil {
			return "", fmt.Errorf("work Redis queue: %w", err)
		}
	}
	if count.Load() != 1 {
		return "", fmt.Errorf("queued listener count = %d, want 1", count.Load())
	}
	if (name == "queued-redis" || name == "queue-routing-options" || name == "queue-retry-options") && received != "queued" {
		return "", fmt.Errorf("restored payload = %q, want queued", received)
	}
	if name == "raw-queued-event" && (!strings.Contains(observed, "rawQueuedEvent") || !strings.Contains(observed, "demo.event.raw")) {
		return "", fmt.Errorf("raw event = %q, want demo.event.raw", observed)
	}
	return fmt.Sprintf("handled=%d; event=%s", count.Load(), outgoing.Name()), nil
}

func runFacade(ctx context.Context) string {
	var count atomic.Int32
	name := "demo.event.facade"
	event.Listen(name, counter{count: &count})
	before := event.Has(name)
	event.Dispatch(ctx, message{EventName: name})
	event.Forget(name)
	return fmt.Sprintf("handled=%d; before=%t; after=%t", count.Load(), before, event.Has(name))
}
func runAsync(ctx context.Context) (string, error) {
	bus := event.New()
	done := make(chan struct{}, 1)
	bus.Listen("demo.event.async", event.Async(func(context.Context, event.Event) error { done <- struct{}{}; return nil }))
	bus.Dispatch(ctx, message{EventName: "demo.event.async"})
	select {
	case <-done:
		return "handled=1; mode=async", nil
	case <-time.After(3 * time.Second):
		return "", errors.New("async listener did not finish")
	}
}

var (
	_ event.Event                = message{}
	_ event.Listener             = counter{}
	_ event.ShouldQueue          = queuedCounter{}
	_ event.QueueOptionsProvider = routedCounter{}
	_ event.Subscriber           = subscriber{}
)
