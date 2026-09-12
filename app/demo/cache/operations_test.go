package cachedemo_test

import "testing"

func TestCacheDemoAdd(t *testing.T) {
	assertScenario(t, "add", "first=true; second=false; value=first")
}
func TestCacheDemoPutMany(t *testing.T) { assertScenario(t, "put-many", "alpha/beta") }
func TestCacheDemoRemember(t *testing.T) {
	assertScenario(t, "remember", "loaded/loaded; loader_calls=1")
}
func TestCacheDemoRememberForever(t *testing.T) {
	assertScenario(t, "remember-forever", "permanent/permanent; loader_calls=1")
}
func TestCacheDemoFlexible(t *testing.T) {
	assertScenario(t, "flexible", "first=version-1; stale=version-1; refreshed=version-2")
}
func TestCacheDemoTouch(t *testing.T) {
	assertScenario(t, "touch", "updated=true; value=retained; missing=false")
}
func TestCacheDemoMany(t *testing.T) { assertScenario(t, "many", "present=one; missing=fallback") }
func TestCacheDemoPull(t *testing.T) { assertScenario(t, "pull", "value=taken; remains=false") }
func TestCacheDemoForget(t *testing.T) {
	assertScenario(t, "forget", "after_forget=false; after_delete=false")
}
func TestCacheDemoForgetMany(t *testing.T) { assertScenario(t, "forget-many", "a=\"\"; b=\"\"") }
func TestCacheDemoFlush(t *testing.T) {
	assertScenario(t, "flush", "after_flush=false; after_clear=false")
}
func TestCacheDemoCounters(t *testing.T) { assertScenario(t, "counters", "1 -> 5 -> 3") }
func TestCacheDemoLock(t *testing.T) {
	assertScenario(t, "lock", "acquired=true; contender=false; released=true")
}
func TestCacheDemoLockCallback(t *testing.T) {
	assertScenario(t, "lock-callback", "acquired=true; callback=true; reacquired=true")
}
func TestCacheDemoLockBlock(t *testing.T) {
	assertScenario(t, "lock-block", "timeout=true; acquired=true; callback=true")
}
func TestCacheDemoLockRestore(t *testing.T) {
	assertScenario(t, "lock-restore", "restored=true; reacquired=true; force_released=true")
}
func TestCacheDemoLockFlush(t *testing.T) {
	assertScenario(t, "lock-flush", "flushed=true; reacquired=true")
}
func TestCacheDemoFunnel(t *testing.T) {
	assertScenario(t, "funnel", "failure_callback=true; entered_after_release=true")
}
func TestCacheDemoWithoutOverlapping(t *testing.T) {
	assertScenario(t, "without-overlapping", "overlap_blocked=true; entered=true; callback=true")
}
func TestCacheDemoMemoryTags(t *testing.T) {
	assertScenario(t, "tags-memory", "before=tagged; after_flush=false")
}
