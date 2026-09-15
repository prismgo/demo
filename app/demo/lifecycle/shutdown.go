package lifecycledemo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/exception"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/logger"
)

// runnerShutdownScenario verifies RunContext closes the application after the runner.
func runnerShutdownScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	if err := app.RunContext(func(context.Context) error { return nil }); err != nil {
		return "", fmt.Errorf("run shutdown lifecycle: %w", err)
	}
	return fmt.Sprintf("runner-returned=true closed=%t", app.Context().Err() != nil), nil
}

// rootContextCancelScenario verifies the root context cancels before app.terminating.
func rootContextCancelScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	canceledBefore := false
	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventAppTerminating, event.ListenerFunc(func(context.Context, event.Event) error {
		canceledBefore = app.Context().Err() != nil
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot root-context application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close root-context application: %w", err)
	}
	return fmt.Sprintf("canceled-before-terminating=%t", canceledBefore), nil
}

// terminatingEventOrderScenario verifies context cancel, event and terminate ordering.
func terminatingEventOrderScenario(base string) (string, error) {
	log := &eventLog{}
	provider := terminableMarker{
		id: "demo.terminable",
		terminate: func(context.Context) error {
			log.add("terminate")
			return nil
		},
	}
	app := foundation.NewApplication(base)
	defer app.Close()
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register terminating provider: %w", err)
	}

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventAppTerminating, event.ListenerFunc(func(context.Context, event.Event) error {
		if app.Context().Err() != nil {
			log.add("context-canceled")
		}
		log.add("terminating")
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot terminating application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close terminating application: %w", err)
	}
	return "order=" + strings.Join(log.values(), ","), nil
}

// terminableProvidersScenario verifies providers terminate in reverse repository order.
func terminableProvidersScenario(base string) (string, error) {
	log := &eventLog{}
	app := foundation.NewApplication(base)
	defer app.Close()
	for _, id := range []string{"demo.p1", "demo.p2", "demo.p3"} {
		provider := terminableMarker{
			id: id,
			terminate: func(context.Context) error {
				log.add(id)
				return nil
			},
		}
		if err := app.RegisterProvider(provider); err != nil {
			return "", fmt.Errorf("register provider %s: %w", id, err)
		}
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot terminable application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close terminable application: %w", err)
	}
	order := strings.Join(log.values(), ",")
	if order != "demo.p3,demo.p2,demo.p1" {
		return "", fmt.Errorf("terminate order = %q, want demo.p3,demo.p2,demo.p1", order)
	}
	return "order=" + order, nil
}

// cleanupFunctionsScenario verifies RegisterCleanup functions run in reverse order.
func cleanupFunctionsScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	log := &eventLog{}
	for _, id := range []string{"demo.c1", "demo.c2", "demo.c3"} {
		app.RegisterCleanup(func(*foundation.Application) error {
			log.add(id)
			return nil
		})
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot cleanup-functions application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close cleanup-functions application: %w", err)
	}
	order := strings.Join(log.values(), ",")
	if order != "demo.c3,demo.c2,demo.c1" {
		return "", fmt.Errorf("cleanup order = %q, want demo.c3,demo.c2,demo.c1", order)
	}
	return "order=" + order, nil
}

// resourceCloseOrderScenario verifies normal container resources close before reporting ones.
func resourceCloseOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	log := &eventLog{}
	register := func(key string, group container.CloseGroup) error {
		return app.Container().Instance(key, &closeProbe{}, container.WithCloser(func(*closeProbe) error {
			log.add(key)
			return nil
		}), container.WithCloseGroup(group))
	}
	if err := register("demo.normal", container.CloseGroupNormal); err != nil {
		return "", fmt.Errorf("register normal resource: %w", err)
	}
	if err := register("demo.reporting", container.CloseGroupReporting); err != nil {
		return "", fmt.Errorf("register reporting resource: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot resource-close-order application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close resource-close-order application: %w", err)
	}
	order := strings.Join(log.values(), ",")
	if order != "demo.normal,demo.reporting" {
		return "", fmt.Errorf("resource close order = %q, want demo.normal,demo.reporting", order)
	}
	return "order=" + order, nil
}

// shutdownErrorReportingScenario verifies close errors are reported while reporting resources live.
func shutdownErrorReportingScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	closeErr := errors.New("demo resource close failure")
	reported := false
	reportingClosed := false
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot shutdown-error-reporting application: %w", err)
	}
	// Register the reporting resources after boot so boot-time exception handler and
	// logger construction cannot overwrite the overrides that observe the close error.
	manager, err := logger.NewManager(logger.Config{
		Default:  "null",
		Channels: map[string]logger.ChannelOptions{"null": {Driver: "null", Level: "debug"}},
	})
	if err != nil {
		return "", fmt.Errorf("build reporting logger manager: %w", err)
	}
	if err := app.Container().Instance("logger.manager", manager, container.WithCloseGroup(container.CloseGroupReporting)); err != nil {
		return "", fmt.Errorf("register reporting logger manager: %w", err)
	}
	if err := app.Container().Instance(foundation.ContainerKeyExceptionHandler, exception.New(
		exception.WithPanicStack(false),
		exception.WithReporter(func(_ any, err error, fields map[string]any) {
			if errors.Is(err, closeErr) && fields["phase"] == "application.close" {
				reported = true
			}
		}),
	), container.WithCloseGroup(container.CloseGroupReporting)); err != nil {
		return "", fmt.Errorf("register reporting exception handler: %w", err)
	}
	if err := app.Container().Instance("demo.failing.resource", &closeProbe{}, container.WithCloser(func(*closeProbe) error {
		return closeErr
	})); err != nil {
		return "", fmt.Errorf("register failing resource: %w", err)
	}
	if err := app.Container().Instance("demo.reporting.resource", &closeProbe{}, container.WithCloser(func(*closeProbe) error {
		reportingClosed = true
		return nil
	}), container.WithCloseGroup(container.CloseGroupReporting)); err != nil {
		return "", fmt.Errorf("register reporting resource: %w", err)
	}
	closeResult := app.Close()
	if !errors.Is(closeResult, closeErr) {
		return "", fmt.Errorf("close error = %v, want %v", closeResult, closeErr)
	}
	if !reported {
		return "", fmt.Errorf("close error was not reported before reporting resources closed")
	}
	return fmt.Sprintf("reported=%t reporting-closed=%t error=%t",
		reported, reportingClosed, errors.Is(closeResult, closeErr)), nil
}

// terminatedEventOrderScenario verifies terminated event timing and aggregated close error.
func terminatedEventOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	cleanupErr := errors.New("cleanup failed")
	log := &eventLog{}
	durationPositive := false
	closeDurationPositive := false
	gotError := ""
	provider := terminableMarker{
		id: "demo.terminable",
		terminate: func(context.Context) error {
			log.add("terminate")
			return nil
		},
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register terminating provider: %w", err)
	}
	app.RegisterCleanup(func(*foundation.Application) error {
		log.add("cleanup")
		return cleanupErr
	})

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventAppTerminated, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		log.add("terminated")
		if terminated, ok := ev.(event.AppTerminated); ok {
			durationPositive = terminated.Duration > 0
			closeDurationPositive = terminated.CloseDuration > 0
			gotError = terminated.Error
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot terminated-event application: %w", err)
	}
	if err := app.Close(); !errors.Is(err, cleanupErr) {
		return "", fmt.Errorf("close error = %v, want %v", err, cleanupErr)
	}
	order := strings.Join(log.values(), ",")
	if order != "terminate,cleanup,terminated" {
		return "", fmt.Errorf("terminated event order = %q, want terminate,cleanup,terminated", order)
	}
	return fmt.Sprintf("order=%s duration-positive=%t close-duration-positive=%t error=%s",
		order, durationPositive, closeDurationPositive, gotError), nil
}

// closeRetryScenario verifies CloseContext retries only unfinished resource closures.
func closeRetryScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	closeErr := errors.New("demo transient close failure")
	var failingCalls, stableCalls int
	if err := app.Container().Instance("demo.failing", &closeProbe{}, container.WithCloser(func(*closeProbe) error {
		failingCalls++
		if failingCalls == 1 {
			return closeErr
		}
		return nil
	})); err != nil {
		return "", fmt.Errorf("register failing resource: %w", err)
	}
	if err := app.Container().Instance("demo.stable", &closeProbe{}, container.WithCloser(func(*closeProbe) error {
		stableCalls++
		return nil
	})); err != nil {
		return "", fmt.Errorf("register stable resource: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot close-retry application: %w", err)
	}
	firstErr := app.Close()
	retryErr := app.Close()
	if !errors.Is(firstErr, closeErr) {
		return "", fmt.Errorf("first close error = %v, want %v", firstErr, closeErr)
	}
	if retryErr != nil {
		return "", fmt.Errorf("retry close error = %v, want nil", retryErr)
	}
	return fmt.Sprintf("first-error=%t retry-clean=%t stable-once=%t failing-twice=%t",
		true, retryErr == nil, stableCalls == 1, failingCalls == 2), nil
}
