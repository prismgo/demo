package redisdemo_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	redisdemo "prismgo-demo/app/demo/redis"
	demotest "prismgo-demo/app/demo/testing"
)

// runIntegration executes one integration scenario against a real Redis service;
// it skips only when PRISMGO_REDIS_TEST_URL is absent.
func runIntegration(t *testing.T, name string) string {
	t.Helper()
	demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	result, err := redisdemo.RunIntegration(name)
	if err != nil {
		t.Fatalf("redis demo %s error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("redis demo %s result case = %q, want %q", name, result.Case, name)
	}
	return result.Value
}

// expectIntegrationValue asserts the full observable value for a deterministic scenario.
func expectIntegrationValue(t *testing.T, name string, want string) {
	t.Helper()
	if got := runIntegration(t, name); got != want {
		t.Fatalf("redis demo %s value = %q, want %q", name, got, want)
	}
}

func TestRedisDemoDefaultClient(t *testing.T) {
	expectIntegrationValue(t, "client", "name=default get=hello")
}

func TestRedisDemoNamedClient(t *testing.T) {
	expectIntegrationValue(t, "named-client", "name=cache get=hello")
}

func TestRedisDemoConnection(t *testing.T) {
	expectIntegrationValue(t, "connection", "name=default client=true listeners=true")
}

func TestRedisDemoStringCommands(t *testing.T) {
	expectIntegrationValue(t, "strings", "get=hello append=10 strlen=10 range=hello mget=[1 2]")
}

func TestRedisDemoHashCommands(t *testing.T) {
	expectIntegrationValue(t, "hashes", "hset=2 hget=alice fields=age=28,name=alice hdel=1")
}

func TestRedisDemoListCommands(t *testing.T) {
	expectIntegrationValue(t, "lists", "rpush=3 range=[a b c] lpop=a llen=2")
}

func TestRedisDemoSetCommands(t *testing.T) {
	expectIntegrationValue(t, "sets", "sadd=3 card=3 member=true members=[a b c] srem=1")
}

func TestRedisDemoSortedSetCommands(t *testing.T) {
	expectIntegrationValue(t, "sorted-sets", "zadd=2 range=[bob alice] score=100 rank=0 incr=105")
}

func TestRedisDemoCounterCommands(t *testing.T) {
	expectIntegrationValue(t, "counters", "incr=1 incrby=4 decrby=3 float=3.5")
}

func TestRedisDemoKeyCommands(t *testing.T) {
	expectIntegrationValue(t, "keys", "exists=true ttl-positive=true persist=true type=string del=1 gone=true")
}

func TestRedisDemoTransaction(t *testing.T) {
	expectIntegrationValue(t, "transaction", "before=0 after=1 ttl-positive=true")
}

func TestRedisDemoLuaScript(t *testing.T) {
	expectIntegrationValue(t, "lua", "first=1 second=0 value=locked")
}

func TestRedisDemoPipeline(t *testing.T) {
	expectIntegrationValue(t, "pipeline", "commands=3 get=1 mget=[1 2]")
}

func TestRedisDemoPublish(t *testing.T) {
	demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	result, err := redisdemo.RunIntegration("publish")
	if err != nil {
		t.Fatalf("redis demo publish error = %v, want nil", err)
	}
	want := regexp.MustCompile(`^subscribers=1 channel=prismgo_demo_redis_pubsub_\d+:notifications payload=order-received$`)
	if !want.MatchString(result.Value) {
		t.Fatalf("redis demo publish value = %q, want subscriber/channel/payload observations", result.Value)
	}
}

func TestRedisDemoSubscribe(t *testing.T) {
	demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	result, err := redisdemo.RunIntegration("subscribe")
	if err != nil {
		t.Fatalf("redis demo subscribe error = %v, want nil", err)
	}
	want := regexp.MustCompile(`^subscribed=true channel=prismgo_demo_redis_pubsub_\d+:notifications payload=hello$`)
	if !want.MatchString(result.Value) {
		t.Fatalf("redis demo subscribe value = %q, want channel/payload observations", result.Value)
	}
}

func TestRedisDemoPatternSubscribe(t *testing.T) {
	demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	result, err := redisdemo.RunIntegration("psubscribe")
	if err != nil {
		t.Fatalf("redis demo psubscribe error = %v, want nil", err)
	}
	want := regexp.MustCompile(`^pattern=prismgo_demo_redis_pubsub_\d+:orders:\* channel=prismgo_demo_redis_pubsub_\d+:orders:1 payload=created$`)
	if !want.MatchString(result.Value) {
		t.Fatalf("redis demo psubscribe value = %q, want pattern/channel/payload observations", result.Value)
	}
}

func TestRedisDemoDefaultConnection(t *testing.T) {
	expectIntegrationValue(t, "default-connection", "name=default same-as-named=true ping=pong")
}

func TestRedisDemoNamedConnection(t *testing.T) {
	expectIntegrationValue(t, "named-connection", "name=cache ping=pong")
}

func TestRedisDemoDefaultConnectionMethod(t *testing.T) {
	expectIntegrationValue(t, "default-connection-method", "name=default same=true")
}

func TestRedisDemoMultipleConnections(t *testing.T) {
	expectIntegrationValue(t, "multiple-connections", "default-db=0 cache-db=5 isolated=true")
}

func TestRedisDemoPurgeRebuild(t *testing.T) {
	expectIntegrationValue(t, "purge-rebuild", "closed=true rebuilt=true ping=pong")
}

func TestRedisDemoCommandExecutedEvent(t *testing.T) {
	expectIntegrationValue(t, "command-executed-event", "event=redis.command_executed command=set connection=default")
}

func TestRedisDemoCommandFailedEvent(t *testing.T) {
	expectIntegrationValue(t, "command-failed-event", "event=redis.command_failed command=incr error=true")
}

func TestRedisDemoBatchExecutedEvent(t *testing.T) {
	expectIntegrationValue(t, "batch-executed-event", "event=redis.command_batch_executed commands=2")
}

func TestRedisDemoBatchFailedEvent(t *testing.T) {
	expectIntegrationValue(t, "batch-failed-event", "event=redis.command_batch_failed commands=2 error=true")
}

func TestRedisDemoEventPayloads(t *testing.T) {
	expectIntegrationValue(t, "event-payloads", "executed=set/default/params:2 failed=incr/default/params:1/error:true")
}

func TestRedisDemoGlobalCommandListener(t *testing.T) {
	expectIntegrationValue(t, "global-command-listener", "registered=true received=1 command=set")
}

func TestRedisDemoGlobalFailureListener(t *testing.T) {
	expectIntegrationValue(t, "global-failure-listener", "registered=true received=1 command=incr")
}

func TestRedisDemoGlobalBatchListener(t *testing.T) {
	expectIntegrationValue(t, "global-batch-listener", "registered=true batches=1 commands=2")
}

func TestRedisDemoConnectionListener(t *testing.T) {
	expectIntegrationValue(t, "connection-listener", "successes=2 sequence=set,get")
}

func TestRedisDemoConnectionFailureListener(t *testing.T) {
	expectIntegrationValue(t, "connection-failure-listener", "failures=1 command=incr error=true")
}

func TestRedisDemoListenerPanicIsolation(t *testing.T) {
	expectIntegrationValue(t, "listener-panic", "command=ok surviving=1 reported=true")
}

func TestRedisDemoDisableEvents(t *testing.T) {
	expectIntegrationValue(t, "disable-events", "events-before=1 events-after=1")
}

func TestRedisDemoEnableEvents(t *testing.T) {
	expectIntegrationValue(t, "enable-events", "existing=2 future=1")
}

// TestRedisDemoIntegrationRequiresURL verifies the runner errors clearly without config.
func TestRedisDemoIntegrationRequiresURL(t *testing.T) {
	_, variable, ok := demotest.LookupIntegration(demotest.ServiceRedisURL)
	if ok {
		t.Skipf("%s is configured, cannot verify missing-URL error here", variable)
	}
	if _, err := redisdemo.RunIntegration("keys"); err == nil || !strings.Contains(err.Error(), variable) {
		t.Fatalf("redis demo keys without URL error = %v, want %s requirement error", err, variable)
	}
}

// TestRedisDemoIntegrationRejectsUnknownScenario keeps the integration dispatcher strict.
func TestRedisDemoIntegrationRejectsUnknownScenario(t *testing.T) {
	if _, err := redisdemo.RunIntegration("no-such-redis-case"); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("unknown redis scenario %q", "no-such-redis-case")) {
		t.Fatalf("redis demo unknown integration scenario error = %v, want unknown scenario error", err)
	}
}
