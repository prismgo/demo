package cachedemo_test

import "testing"

func TestCacheDemoUnsupportedTags(t *testing.T) {
	assertScenario(t, "tags-unsupported", "ErrTagsUnsupported")
}
func TestCacheDemoMemo(t *testing.T) {
	assertScenario(t, "memo", "first=first; memo=first; after_write=third")
}
func TestCacheDemoFailover(t *testing.T) {
	assertScenario(t, "failover", "value=survived; fallback=survived")
}
func TestCacheDemoCustomDriver(t *testing.T) {
	assertScenario(t, "custom-driver", "value=custom-value; prefix=demo:tenant")
}
func TestCacheDemoResourceLifecycle(t *testing.T) {
	assertScenario(t, "resource-lifecycle", "manager_closed=true")
}
func TestCacheDemoEvents(t *testing.T) {
	assertScenario(t, "events", "cache.writing,cache.written,cache.retrieving,cache.hit")
}
func TestCacheDemoEventContract(t *testing.T) {
	assertScenario(t, "event-contract", "event=cache.written; store=memory")
}
func TestCacheDemoDeferred(t *testing.T) {
	assertScenario(t, "deferred", "stale=refreshed; after_deferred=updated")
}
func TestCacheDemoKeyPrefixes(t *testing.T) {
	assertScenario(t, "key-prefixes", "cache=prismgo_cache:memory; lock=prismgo_cache:memory:locks")
}
func TestCacheDemoEncoding(t *testing.T) { assertScenario(t, "encoding", "number=42; name=Ada") }
func TestCacheDemoErrors(t *testing.T) {
	assertScenario(t, "errors", "ErrCacheMiss,ErrStoreNotFound,ErrTagsUnsupported")
}
func TestCacheDemoMemoryCapabilities(t *testing.T) {
	assertScenario(t, "memory-capabilities", "store=memory; touch=true; atomic=true; bulk=true; lock=true; tags=true")
}
func TestCacheDemoFileCapabilities(t *testing.T) {
	assertScenario(t, "file-capabilities", "store=file; touch=true; atomic=true; bulk=true; lock=true; tags=false")
}
func TestCacheDemoFailoverCapabilities(t *testing.T) {
	assertScenario(t, "failover-capabilities", "store=failover; touch=true; atomic=true; bulk=true; lock=true; tags=true")
}
func TestCacheDemoLaravelCompatibility(t *testing.T) {
	assertScenario(t, "laravel-compatibility", "Cache::store/get/put -> prismgo")
}
