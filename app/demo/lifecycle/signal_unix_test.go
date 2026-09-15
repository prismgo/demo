//go:build unix

package lifecycledemo_test

import "testing"

func TestLifecycleDemoSignalShutdown(t *testing.T) {
	expectValue(t, "signal-shutdown", "sigint=true sigterm=true")
}
