package redisdemo_test

import "testing"

func TestRedisDemoQueueDriver(t *testing.T) {
	expectIntegrationValue(t, "queue-driver", "queued=1 remaining=0 processed=1")
}

func TestRedisDemoQueueReadyList(t *testing.T) {
	expectIntegrationValue(t, "queue-ready", "ready=2 delayed=0")
}

func TestRedisDemoQueueDelayedSet(t *testing.T) {
	expectIntegrationValue(t, "queue-delayed", "delayed=1 ready=0 not-due=true")
}

func TestRedisDemoQueueBlockingPop(t *testing.T) {
	expectIntegrationValue(t, "queue-blocking-pop", "blocked=true popped=true")
}

func TestRedisDemoQueueFailedStore(t *testing.T) {
	expectIntegrationValue(t, "queue-failed", "stored=true error=true")
}
