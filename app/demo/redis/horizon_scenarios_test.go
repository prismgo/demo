package redisdemo_test

import "testing"

func TestRedisDemoHorizonProcesses(t *testing.T) {
	expectIntegrationValue(t, "horizon-processes", "masters=1 supervisors=1 workers=1 stale=true")
}

func TestRedisDemoHorizonControl(t *testing.T) {
	expectIntegrationValue(t, "horizon-control", "paused=paused supervisor-paused=true terminating=terminating cleared=true")
}

func TestRedisDemoHorizonMetrics(t *testing.T) {
	expectIntegrationValue(t, "horizon-metrics", "windows=1 processed=5 rollups=1")
}

func TestRedisDemoHorizonQueueLengths(t *testing.T) {
	expectIntegrationValue(t, "horizon-queue-lengths", "queues=2 first=3 empty=true")
}

func TestRedisDemoHorizonSummaries(t *testing.T) {
	expectIntegrationValue(t, "horizon-summaries", "batch=running total=3 page=1")
}

func TestRedisDemoHorizonJobDiagnostics(t *testing.T) {
	expectIntegrationValue(t, "horizon-job-diagnostics", "details=1 kind=failed found=true")
}

func TestRedisDemoHorizonObservability(t *testing.T) {
	expectIntegrationValue(t, "horizon-observability", "diagnostics=1 reason=buffer_full")
}

func TestRedisDemoHorizonOrphans(t *testing.T) {
	expectIntegrationValue(t, "horizon-orphans", "orphans=1 older=1 forgotten=true")
}
