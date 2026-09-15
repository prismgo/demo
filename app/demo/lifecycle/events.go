package lifecycledemo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
)

// bestEffortEventsScenario verifies lifecycle dispatch is skipped without an event dispatcher.
func bestEffortEventsScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	if err := app.Container().Instance(foundation.ContainerKeyEventDispatcher, nil); err != nil {
		return "", fmt.Errorf("unregister event dispatcher: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot application without event dispatcher: %w", err)
	}
	closeErr := app.Close()
	if closeErr != nil {
		return "", fmt.Errorf("close application without event dispatcher: %w", closeErr)
	}
	return "boot-ok=true close-ok=true", nil
}

// appBootingEventScenario verifies the app.booting payload and its timing before provider boots.
func appBootingEventScenario(base string) (string, error) {
	app := buildApplication(base, nil, nil)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	args := -1
	bus.Listen(event.EventAppBooting, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if booting, ok := ev.(event.AppBooting); ok {
			args = len(booting.Args)
		}
		log.add("app.booting")
		return nil
	}))
	bus.Listen(event.EventProviderBooting, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("provider.booting")
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot app-booting-event application: %w", err)
	}
	sequence := log.values()
	before := indexOf(sequence, "app.booting") == 0 && indexOf(sequence, "provider.booting") > 0
	return fmt.Sprintf("args=%d before-provider-boot=%t", args, before), nil
}

// appBootedEventScenario verifies the app.booted payload and its timing after provider boots.
func appBootedEventScenario(base string) (string, error) {
	app := buildApplication(base, nil, nil)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	durationNonNegative := false
	cacheBound := false
	bus.Listen(event.EventProviderBooted, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("provider.booted")
		return nil
	}))
	bus.Listen(event.EventAppBooted, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if booted, ok := ev.(event.AppBooted); ok {
			durationNonNegative = booted.Duration >= 0
		}
		cacheBound = app.Container().Bound("cache.manager")
		log.add("app.booted")
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot app-booted-event application: %w", err)
	}
	sequence := log.values()
	after := len(sequence) > 1 && sequence[len(sequence)-1] == "app.booted"
	return fmt.Sprintf("duration-nonnegative=%t after-provider-boot=%t cache-bound=%t",
		durationNonNegative, after, cacheBound), nil
}

// appTerminatingEventScenario verifies the app.terminating payload and context cancellation timing.
func appTerminatingEventScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	reason := ""
	contextCanceled := false
	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventAppTerminating, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if terminating, ok := ev.(event.AppTerminating); ok {
			reason = terminating.Reason
		}
		contextCanceled = app.Context().Err() != nil
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot app-terminating-event application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close app-terminating-event application: %w", err)
	}
	return fmt.Sprintf("reason=%s context-canceled=%t", reason, contextCanceled), nil
}

// appTerminatedEventScenario verifies the app.terminated payload duration and error fields.
func appTerminatedEventScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	durationPositive := false
	closeDurationPositive := false
	errorEmpty := false
	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventAppTerminated, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if terminated, ok := ev.(event.AppTerminated); ok {
			durationPositive = terminated.Duration > 0
			closeDurationPositive = terminated.CloseDuration > 0
			errorEmpty = terminated.Error == ""
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot app-terminated-event application: %w", err)
	}
	time.Sleep(time.Millisecond)
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close app-terminated-event application: %w", err)
	}
	return fmt.Sprintf("duration-positive=%t close-duration-positive=%t error-empty=%t",
		durationPositive, closeDurationPositive, errorEmpty), nil
}

// providerRegisteringEventScenario verifies the registering payload and timing.
func providerRegisteringEventScenario(base string) (string, error) {
	phase, provider, sequence, err := providerPhaseResults(base, "registering")
	if err != nil {
		return "", err
	}
	before := indexOf(sequence, "registering:demo.p1") < indexOf(sequence, "registered:demo.p1")
	return fmt.Sprintf("phase=%s provider=%s before-registered=%t", phase, provider, before), nil
}

// providerRegisteredEventScenario verifies the registered payload and timing.
func providerRegisteredEventScenario(base string) (string, error) {
	phase, provider, sequence, err := providerPhaseResults(base, "registered")
	if err != nil {
		return "", err
	}
	after := indexOf(sequence, "registered:demo.p1") > indexOf(sequence, "registering:demo.p1")
	return fmt.Sprintf("phase=%s provider=%s after-registering=%t", phase, provider, after), nil
}

// providerBootingEventScenario verifies the booting payload and timing after all registrations.
func providerBootingEventScenario(base string) (string, error) {
	phase, provider, sequence, err := providerPhaseResults(base, "booting")
	if err != nil {
		return "", err
	}
	afterAllRegisters := indexOf(sequence, "booting:demo.p1") > indexOf(sequence, "registered:demo.p2")
	return fmt.Sprintf("phase=%s provider=%s after-all-registers=%t", phase, provider, afterAllRegisters), nil
}

// providerBootedEventScenario verifies the booted payload and timing after booting.
func providerBootedEventScenario(base string) (string, error) {
	phase, provider, sequence, err := providerPhaseResults(base, "booted")
	if err != nil {
		return "", err
	}
	afterBooting := indexOf(sequence, "booted:demo.p1") > indexOf(sequence, "booting:demo.p1")
	return fmt.Sprintf("phase=%s provider=%s after-booting=%t", phase, provider, afterBooting), nil
}

// listenersScenario verifies lifecycle listeners registered during provider boot receive events.
func listenersScenario(base string) (string, error) {
	var booted, terminated int32
	app := foundation.Configure(base).
		WithProviders(listenerProvider{booted: &booted, terminated: &terminated}).
		Create()
	defer app.Close()

	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot listeners application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close listeners application: %w", err)
	}
	return fmt.Sprintf("booted-listener=%t terminated-listener=%t",
		atomic.LoadInt32(&booted) == 1, atomic.LoadInt32(&terminated) == 1), nil
}

// payloadBoundariesScenario verifies lifecycle payloads stay serializable and runtime-free.
func payloadBoundariesScenario() (string, error) {
	payloads := []event.Event{
		event.AppBooting{Args: []string{"serve"}},
		event.AppBooted{Duration: time.Millisecond},
		event.AppTerminating{Reason: "application shutdown"},
		event.AppTerminated{Duration: time.Second, CloseDuration: time.Millisecond},
		event.ProviderRegistering{Provider: "demo"},
		event.ProviderRegistered{Provider: "demo"},
		event.ProviderBooting{Provider: "demo"},
		event.ProviderBooted{Provider: "demo"},
		event.ServerStarting{Addr: ":0", PID: 1},
		event.ServerStarted{Addr: ":0", PID: 1},
		event.ServerStopping{Addr: ":0", Reason: "context canceled"},
		event.ServerStopped{Addr: ":0", Duration: time.Millisecond},
		event.RequestReceived{Method: "GET", Path: "/", ClientIP: "127.0.0.1", RequestID: "id", ReceivedAt: time.Now()},
		event.RequestHandled{Method: "GET", Path: "/", RequestID: "id", Status: 200, Duration: time.Millisecond},
		event.RequestFailed{Method: "GET", Path: "/", RequestID: "id", Status: 500, Duration: time.Millisecond, Error: "boom"},
		event.RequestFinished{Method: "GET", Path: "/", RequestID: "id", Status: 200, Duration: time.Millisecond},
		event.ConsoleApplicationStarting{KernelName: "PrismGo"},
		event.CommandStarting{Command: "demo", Input: []string{"demo"}},
		event.CommandFinished{Command: "demo", Succeeded: true, Duration: time.Millisecond},
	}
	for _, payload := range payloads {
		if _, err := json.Marshal(payload); err != nil {
			return "", fmt.Errorf("marshal lifecycle payload %s: %w", payload.Name(), err)
		}
	}
	return fmt.Sprintf("payloads=%d serializable=true", len(payloads)), nil
}
