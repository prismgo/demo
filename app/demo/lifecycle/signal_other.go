//go:build !unix

package lifecycledemo

import "fmt"

// signalShutdownScenario reports that process signal handling is unavailable.
func signalShutdownScenario(string) (string, error) {
	return "", fmt.Errorf("signal shutdown scenario is not supported on this platform")
}
