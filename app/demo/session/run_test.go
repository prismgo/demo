package sessiondemo_test

import (
	"strings"
	"testing"

	sessiondemo "prismgo-demo/app/demo/session"
)

// runCase executes one scenario and fails with locating context when it errors.
func runCase(t *testing.T, name string) string {
	t.Helper()
	result, err := sessiondemo.Run(name)
	if err != nil {
		t.Fatalf("session demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("session demo %q result case = %q, want %q", name, result.Case, name)
	}
	return result.Value
}

// expectValue asserts the full observable value for a deterministic scenario.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	if got := runCase(t, name); got != want {
		t.Fatalf("session demo %q value = %q, want %q", name, got, want)
	}
}

func TestSessionDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "contracts=driver,locker,lock,encryptor")
}

func TestSessionDemoConfig(t *testing.T) {
	want := "driver=redis lifetime=45m0s expireOnClose=true encrypt=true encoding=json " +
		"connection=demo prefix=demo_session cookie=demo_session_cookie path=/demo domain=demo.test " +
		"secure=true httpOnly=false sameSite=strict lockTTL=7s lockWait=3s"
	expectValue(t, "config", want)
}

func TestSessionDemoFileDriver(t *testing.T) {
	expectValue(t, "file-driver", "file=true mode=-rw------- restored=1001")
}

func TestSessionDemoTopLevelConfig(t *testing.T) {
	expectValue(t, "top-level-config", "driver=file lifetime=2h0m0s expireOnClose=false encrypt=false encoding=msgpack")
}

func TestSessionDemoCookieConfig(t *testing.T) {
	expectValue(t, "cookie-config", `cookie=prismgo_session path=/ domain="" secure=false httpOnly=true sameSite=lax`)
}

func TestSessionDemoFileConfig(t *testing.T) {
	expectValue(t, "file-config", "files=storage/framework/sessions")
}

func TestSessionDemoRedisConfig(t *testing.T) {
	expectValue(t, "redis-config", "connection=default prefix=prismgo_session")
}

func TestSessionDemoLockConfig(t *testing.T) {
	expectValue(t, "lock-config", "lockSeconds=10s lockWait=10s")
}

func TestSessionDemoManagerConfig(t *testing.T) {
	expectValue(t, "manager-config", "lifetime=30m0s cookie=admin_session secure=true files=true invalid-encoding=true")
}

func TestSessionDemoMiddleware(t *testing.T) {
	expectValue(t, "middleware", `cookie=prismgo_session httpOnly=true sameSiteLax=true persistent=true read="user=1001 notice=saved"`)
}

func TestSessionDemoRecovery(t *testing.T) {
	expectValue(t, "recovery", "missing=true expired=true corrupted=true invalid=true")
}

func TestSessionDemoResponseBuffering(t *testing.T) {
	expectValue(t, "response-buffering", `delivered="chunk-1|chunk-2" save-failed-status=500 body-suppressed=true`)
}

func TestSessionDemoStoreFrom(t *testing.T) {
	expectValue(t, "store-from", "store=[id-len=43 facade-get=1001] bare=[default=guest found=false put=true]")
}

func TestSessionDemoCustomManager(t *testing.T) {
	expectValue(t, "custom-manager", `cookie=admin_session r1="scope=admin" r2="scope=admin" default=resolved=true persisted=true`)
}

func TestSessionDemoGet(t *testing.T) {
	expectValue(t, "get", "value=alice default=guest absent=<nil>")
}

func TestSessionDemoAll(t *testing.T) {
	expectValue(t, "all", "count=2 isolated=true")
}

func TestSessionDemoSubsets(t *testing.T) {
	expectValue(t, "subsets", "only=2 except=2")
}

func TestSessionDemoHas(t *testing.T) {
	expectValue(t, "has", "has=true nil=false")
}

func TestSessionDemoExists(t *testing.T) {
	expectValue(t, "exists", "exists=true missing=false")
}

func TestSessionDemoMissing(t *testing.T) {
	expectValue(t, "missing", "missing=true present=false")
}

func TestSessionDemoPut(t *testing.T) {
	expectValue(t, "put", "scalar=1001 map=true slice=true")
}

func TestSessionDemoCounters(t *testing.T) {
	expectValue(t, "counters", "first=1 second=3 remaining=-2 invalid=true")
}

func TestSessionDemoFlash(t *testing.T) {
	expectValue(t, "flash", "flash r1:status=ok r2:status=ok r3:status=-")
}

func TestSessionDemoNow(t *testing.T) {
	expectValue(t, "now", "now r1:preview=once r2:preview=-")
}

func TestSessionDemoReflash(t *testing.T) {
	expectValue(t, "reflash", "reflash r1:a=one b=two r2:a=one b=two r3:a=one b=two r4:a=- b=-")
}

func TestSessionDemoKeep(t *testing.T) {
	expectValue(t, "keep", "keep r1:kept=v1 dropped=v2 r2:kept=v1 dropped=v2 r3:kept=v1 dropped=- r4:kept=- dropped=-")
}

func TestSessionDemoForget(t *testing.T) {
	expectValue(t, "forget", "gone=true gone=true")
}

func TestSessionDemoFlush(t *testing.T) {
	expectValue(t, "flush", "data=0 flash=<nil> id-same=true")
}

func TestSessionDemoPull(t *testing.T) {
	expectValue(t, "pull", "pulled=hello exists=false fallback=fallback")
}

func TestSessionDemoRegenerate(t *testing.T) {
	expectValue(t, "regenerate", "rotated=true retained=1001 old-destroyed=true")
}

func TestSessionDemoInvalidate(t *testing.T) {
	expectValue(t, "invalidate", "rotated=true cleared=true empty=true old-destroyed=true")
}

func TestSessionDemoBlocking(t *testing.T) {
	expectValue(t, "blocking", "blocked-timeout=true resumed=true")
}

func TestSessionDemoFileLock(t *testing.T) {
	expectValue(t, "file-lock", "held=true contention=true released=true re-release-not-held=true expired-takeover=true stale-release-not-held=true")
}

func TestSessionDemoIDCookie(t *testing.T) {
	expectValue(t, "id-cookie", "name=admin_session path=/app domain=example.test secure=true httpOnly=true sameSite=strict maxAge-positive=true server=true")
}

func TestSessionDemoExpireOnClose(t *testing.T) {
	expectValue(t, "expire-on-close", "browser-session=true server-persistent=true")
}

func TestSessionDemoQueuedCookies(t *testing.T) {
	expectValue(t, "queued-cookies", "order=prismgo_session,locale")
}

func TestSessionDemoEncryption(t *testing.T) {
	expectValue(t, "encryption", "encrypted=true cookie-opaque=true restored=1001")
}

func TestSessionDemoEncryptor(t *testing.T) {
	expectValue(t, "encryptor", "encrypt-copy=true decrypt-copy=true")
}

func TestSessionDemoCustomEncryptor(t *testing.T) {
	expectValue(t, "custom-encryptor", "marker=true restored=1001")
}

func TestSessionDemoSensitiveError(t *testing.T) {
	expectValue(t, "sensitive-error", "encrypt-redacted=true encrypt-typed=true decrypt-redacted=true decrypt-typed=true")
}

func TestSessionDemoDriverContract(t *testing.T) {
	expectValue(t, "driver-contract", "file=true redis=true custom=true wrote=true destroyed=true")
}

func TestSessionDemoLockerContract(t *testing.T) {
	expectValue(t, "locker-contract", "locker=file,redis lock-release=true")
}

func TestSessionDemoExtend(t *testing.T) {
	expectValue(t, "extend", "driver=demo-memory resolved=true restored=1001")
}

func TestSessionDemoExtendValidation(t *testing.T) {
	expectValue(t, "extend-validation", "empty-ignored=true nil-ignored=true override=second")
}

func TestSessionDemoUnknownDriver(t *testing.T) {
	expectValue(t, "unknown-driver", "resolve=true manager=true")
}

func TestSessionDemoErrors(t *testing.T) {
	expectValue(t, "errors", "invalid-config=true invalid-id=true invalid-expires=true driver-not-found=true invalid-type=true lock-not-held=true")
}

func TestSessionDemoRecoverableErrors(t *testing.T) {
	expectValue(t, "recoverable-errors", "recovered=6 fatal=true")
}

func TestSessionDemoLaravelCompatibility(t *testing.T) {
	expectValue(t, "laravel-compatibility", "mapped=19 blocking=true")
}

func TestSessionDemoRejectsUnknownScenario(t *testing.T) {
	if _, err := sessiondemo.Run("no-such-case"); err == nil || !strings.Contains(err.Error(), `unknown scenario "no-such-case"`) {
		t.Fatalf("session demo unknown scenario error = %v, want unknown scenario error", err)
	}
}
