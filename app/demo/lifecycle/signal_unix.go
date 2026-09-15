//go:build unix

package lifecycledemo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/prismgo/framework/foundation"
)

// signalShutdownScenario verifies SIGINT and SIGTERM trigger graceful shutdown.
//
// The signals are delivered to the current process only after RegisterShutdownSignals
// has installed its handler, so the test binary is never terminated by the default action.
func signalShutdownScenario(base string) (string, error) {
	sigint, err := awaitSignalShutdown(base, syscall.SIGINT)
	if err != nil {
		return "", err
	}
	sigterm, err := awaitSignalShutdown(base, syscall.SIGTERM)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sigint=%t sigterm=%t", sigint, sigterm), nil
}

// awaitSignalShutdown boots an application, raises sig and reports the shutdown cause.
func awaitSignalShutdown(base string, sig syscall.Signal) (bool, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	app.RegisterShutdownSignals()
	runErr := app.RunContext(func(ctx context.Context) error {
		raise := time.AfterFunc(50*time.Millisecond, func() {
			_ = syscall.Kill(os.Getpid(), sig)
		})
		defer raise.Stop()
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(3 * time.Second):
			return errors.New("shutdown signal was not received within 3s")
		}
	})
	cause := context.Cause(app.Context())
	return runErr == nil && cause != nil && strings.Contains(cause.Error(), sig.String()), nil
}
