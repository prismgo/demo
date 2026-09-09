package queuedemo

import (
	"context"
	"fmt"
	"sync"
)

var demoTrace = struct {
	sync.Mutex
	items map[string][]string
}{items: make(map[string][]string)}

type traceGate struct {
	started chan struct{}
	release chan struct{}
}

var demoGates = struct {
	sync.Mutex
	items map[string]*traceGate
}{items: make(map[string]*traceGate)}

// ResetTrace starts an empty trace for one demo execution.
func ResetTrace(id string) {
	demoTrace.Lock()
	demoTrace.items[id] = nil
	demoTrace.Unlock()
}

// RecordTrace appends an observable step from a deserialized job or middleware.
func RecordTrace(id, step string) {
	demoTrace.Lock()
	demoTrace.items[id] = append(demoTrace.items[id], step)
	demoTrace.Unlock()
}

// TakeTrace returns the completed trace and releases its process-local storage.
func TakeTrace(id string) []string {
	demoTrace.Lock()
	defer demoTrace.Unlock()
	steps := append([]string(nil), demoTrace.items[id]...)
	delete(demoTrace.items, id)
	return steps
}

// PrepareGate creates a deterministic blocking point for concurrency examples.
func PrepareGate(id string) {
	demoGates.Lock()
	demoGates.items[id] = &traceGate{started: make(chan struct{}), release: make(chan struct{})}
	demoGates.Unlock()
}

// WaitGateStarted waits until the demo job has entered Handle.
func WaitGateStarted(ctx context.Context, id string) error {
	demoGates.Lock()
	gate := demoGates.items[id]
	demoGates.Unlock()
	if gate == nil {
		return fmt.Errorf("queue demo trace gate %q is not prepared", id)
	}
	select {
	case <-gate.started:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// EnterGate signals that Handle started and waits for the scenario to release it.
func EnterGate(ctx context.Context, id string) error {
	demoGates.Lock()
	gate := demoGates.items[id]
	demoGates.Unlock()
	if gate == nil {
		return fmt.Errorf("queue demo trace gate %q is not prepared", id)
	}
	select {
	case <-gate.started:
	default:
		close(gate.started)
	}
	select {
	case <-gate.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseGate unblocks the demo job and removes the gate.
func ReleaseGate(id string) {
	demoGates.Lock()
	gate := demoGates.items[id]
	delete(demoGates.items, id)
	demoGates.Unlock()
	if gate == nil {
		return
	}
	select {
	case <-gate.release:
	default:
		close(gate.release)
	}
}
