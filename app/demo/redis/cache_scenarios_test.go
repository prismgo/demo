package redisdemo_test

import "testing"

func TestRedisDemoCacheDriver(t *testing.T) {
	expectIntegrationValue(t, "cache-driver", "store=redis value=hello shared-key=true")
}

func TestRedisDemoCacheBasicOperations(t *testing.T) {
	expectIntegrationValue(t, "cache-basic", "get=value has=true missing=true")
}

func TestRedisDemoCacheTTL(t *testing.T) {
	expectIntegrationValue(t, "cache-ttl", "initial-positive=true persisted=true retouched-positive=true")
}

func TestRedisDemoCacheAtomicOperations(t *testing.T) {
	expectIntegrationValue(t, "cache-atomic", "increment=1 increment-by=5 decrement=3 add-new=true add-existing=false pulled=3")
}

func TestRedisDemoCacheBulkOperations(t *testing.T) {
	expectIntegrationValue(t, "cache-bulk", "bulk=a=1,b=2,c=3 forgotten=true/true remaining=3")
}

func TestRedisDemoCacheTags(t *testing.T) {
	expectIntegrationValue(t, "cache-tags", "get=value flushed=true")
}

func TestRedisDemoCacheFlush(t *testing.T) {
	expectIntegrationValue(t, "cache-flush", "before=2 after=0")
}
