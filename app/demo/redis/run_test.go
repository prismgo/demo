package redisdemo_test

import (
	"strings"
	"testing"

	redisdemo "prismgo-demo/app/demo/redis"
)

// runCase executes one scenario and fails with locating context when it errors.
func runCase(t *testing.T, name string) string {
	t.Helper()
	result, err := redisdemo.Run(name)
	if err != nil {
		t.Fatalf("redis demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("redis demo %q result case = %q, want %q", name, result.Case, name)
	}
	return result.Value
}

// expectValue asserts the full observable value for a deterministic scenario.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	if got := runCase(t, name); got != want {
		t.Fatalf("redis demo %q value = %q, want %q", name, got, want)
	}
}

func TestRedisDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "factory=Manager connection=NamedConnection events=4 facade=5")
}

func TestRedisDemoConfiguration(t *testing.T) {
	expectValue(t, "config", "client=go default=default connections=cache,default default-addr=127.0.0.1:6379 cache-db=1")
}

func TestRedisDemoClientIdentifiers(t *testing.T) {
	expectValue(t, "client-identifiers", "accepted=4 rejected=true")
}

func TestRedisDemoNamedConfiguration(t *testing.T) {
	expectValue(t, "named-config", "cache-db=7 cache-name=orders-client")
}

func TestRedisDemoAddressPrecedence(t *testing.T) {
	expectValue(t, "address-precedence", "addr=10.0.0.9:7000 host-override=192.168.1.5:7001")
}

func TestRedisDemoDatabasePrecedence(t *testing.T) {
	expectValue(t, "database-precedence", "db=5 url-kept=1 url-override=3")
}

func TestRedisDemoAuthenticationConfiguration(t *testing.T) {
	expectValue(t, "authentication", "override=acl-user/acl-pass url=url-user/url-pass")
}

func TestRedisDemoURLConfiguration(t *testing.T) {
	expectValue(t, "url", "addr=127.0.0.1:6390 db=4 user=user dial=3s read=2s write=5s retries=3")
}

func TestRedisDemoTLSConfiguration(t *testing.T) {
	expectValue(t, "tls", "tls=true plain=false")
}

func TestRedisDemoClientNameConfiguration(t *testing.T) {
	expectValue(t, "client-name", "name=orders-client")
}

func TestRedisDemoTimeoutConfiguration(t *testing.T) {
	expectValue(t, "timeouts", "duration=3s/500ms/2s seconds=4s/1s/2s")
}

func TestRedisDemoMaxRetriesConfiguration(t *testing.T) {
	expectValue(t, "max-retries", "int=7 string=9")
}

func TestRedisDemoEnvironmentConfiguration(t *testing.T) {
	expectValue(t, "environment", "client=go-redis addr=10.2.0.1:6389 db=6 name=env-client dial=4s read=1s write=2s retries=8")
}

func TestRedisDemoNativeClientContract(t *testing.T) {
	expectValue(t, "native-client", "client=UniversalClient commands=7 pipelines=2 pubsub=3")
}

func TestRedisDemoManagerConstruction(t *testing.T) {
	expectValue(t, "manager", "default=default cache=cache resolved=2 invalid-client=rejected")
}

func TestRedisDemoManagerFromRepository(t *testing.T) {
	expectValue(t, "manager-repository", "source=repository connection=default addr=10.9.0.1:6390 db=3")
}

func TestRedisDemoManagerFromApplication(t *testing.T) {
	expectValue(t, "manager-application", "factory=*redis.Manager connection=default")
}

func TestRedisDemoLazyConnectionCreation(t *testing.T) {
	expectValue(t, "lazy-connection", "before=0 after=1 missing=error client=true")
}

func TestRedisDemoConnectionReuse(t *testing.T) {
	expectValue(t, "connection-reuse", "same=true distinct=true")
}

func TestRedisDemoConnectionsSnapshot(t *testing.T) {
	expectValue(t, "connections-snapshot", "count=2 isolated=true")
}

func TestRedisDemoConnectionSnapshotIsLazy(t *testing.T) {
	expectValue(t, "snapshot-lazy", "initial=0 snapshot=0 after-resolve=1")
}

func TestRedisDemoPurge(t *testing.T) {
	expectValue(t, "purge", "purged=true closed=true rebuilt=true")
}

func TestRedisDemoClose(t *testing.T) {
	expectValue(t, "close", "resolved=2 remaining=0 empty-close=true")
}

func TestRedisDemoCloseErrors(t *testing.T) {
	expectValue(t, "close-errors", "first-error=true remaining=1")
}

func TestRedisDemoCloseCancellation(t *testing.T) {
	expectValue(t, "close-cancellation", "cancelled=true retained=1 retried=0")
}

func TestRedisDemoEventSensitiveParameters(t *testing.T) {
	expectValue(t, "event-sensitive-parameters", "command=set parameters=2 sensitive-verbatim=true")
}

func TestRedisDemoRejectsUnknownScenario(t *testing.T) {
	if _, err := redisdemo.Run("no-such-case"); err == nil || !strings.Contains(err.Error(), `unknown scenario "no-such-case"`) {
		t.Fatalf("redis demo unknown scenario error = %v, want unknown scenario error", err)
	}
}
