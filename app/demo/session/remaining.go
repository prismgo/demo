package sessiondemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	cookiepkg "github.com/prismgo/framework/cookie"
	"github.com/prismgo/framework/session"
)

// Compile-time assertions keep the demo extensions aligned with the public contracts.
var (
	_ session.Encryptor = session.NopEncryptor{}
	_ session.Driver    = (*memoryDriver)(nil)
	_ session.Driver    = (*markerDriver)(nil)
)

// demoMemoryDriverName is the driver name registered by the extend scenario.
const demoMemoryDriverName = "demo-memory"

// runBatch executes the Session ID, blocking, cookie, encryption, custom driver,
// and error scenarios that were added after the first coverage batch.
func runBatch(name string) (string, error) {
	switch name {
	case "regenerate":
		return regenerateScenario()
	case "invalidate":
		return invalidateScenario()
	case "blocking":
		return blockingScenario()
	case "file-lock":
		return fileLockScenario()
	case "id-cookie":
		return withGinTestMode(idCookieScenario)
	case "expire-on-close":
		return withGinTestMode(expireOnCloseScenario)
	case "queued-cookies":
		return withGinTestMode(queuedCookiesScenario)
	case "encryption":
		return encryptionScenario()
	case "encryptor":
		return encryptorScenario()
	case "custom-encryptor":
		return customEncryptorScenario()
	case "sensitive-error":
		return sensitiveErrorScenario()
	case "driver-contract":
		return driverContractScenario()
	case "locker-contract":
		return lockerContractScenario()
	case "extend":
		return extendScenario()
	case "extend-validation":
		return extendValidationScenario()
	case "unknown-driver":
		return unknownDriverScenario()
	case "errors":
		return errorConstantsScenario()
	case "recoverable-errors":
		return recoverableErrorsScenario()
	case "laravel-compatibility":
		return laravelCompatibilityScenario()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

// regenerateScenario rotates the ID while keeping business data and destroying the old ID.
func regenerateScenario() (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx := context.Background()
	store := freshStore(manager)
	store.Put("user_id", int64(1001))
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	oldID := store.ID()
	if err := store.Regenerate(ctx); err != nil {
		return "", err
	}
	newID := store.ID()
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	retained, err := restoreSession(manager, newID)
	if err != nil {
		return "", err
	}
	oldRequest, err := restoreSession(manager, oldID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rotated=%t retained=%v old-destroyed=%t",
		oldID != newID, retained.Get("user_id"), oldRequest.ID() != oldID), nil
}

// invalidateScenario clears data and rotates the ID in one step.
func invalidateScenario() (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx := context.Background()
	store := freshStore(manager)
	store.Put("user_id", int64(1001))
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	oldID := store.ID()
	if err := store.Invalidate(ctx); err != nil {
		return "", err
	}
	newID := store.ID()
	cleared := store.Get("user_id") == nil
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	restored, err := restoreSession(manager, newID)
	if err != nil {
		return "", err
	}
	oldRequest, err := restoreSession(manager, oldID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rotated=%t cleared=%t empty=%t old-destroyed=%t",
		oldID != newID, cleared, len(restored.All()) == 0, oldRequest.ID() != oldID), nil
}

// blockingScenario shows that a request holding the session lock forces a second
// request on the same ID to wait until the configured lock wait elapses.
func blockingScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-blocking-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	cfg.Lock.Wait = 250 * time.Millisecond
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	seed, err := manager.Start(ctx, httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	if err != nil {
		return "", err
	}
	seed.Put("counter", int64(1))
	if err := seed.Save(ctx); err != nil {
		return "", err
	}
	id := seed.ID()

	live, err := manager.Start(ctx, requestWithSessionID(id), httptest.NewRecorder())
	if err != nil {
		return "", err
	}
	_, blockedErr := manager.Start(ctx, requestWithSessionID(id), httptest.NewRecorder())
	blocked := errors.Is(blockedErr, session.ErrLockTimeout)
	if err := live.ReleaseRequestLock(ctx); err != nil {
		return "", err
	}
	resumed, err := manager.Start(ctx, requestWithSessionID(id), httptest.NewRecorder())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("blocked-timeout=%t resumed=%t", blocked, resumed.ID() == id), nil
}

// fileLockScenario exercises the file lock token, contention, release, and
// expiration takeover semantics.
func fileLockScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-file-lock-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	driver, err := session.NewFileDriver(cfg)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	id := newDemoSessionID()
	lock, err := driver.Lock(ctx, id, 10*time.Second, 200*time.Millisecond)
	if err != nil {
		return "", err
	}
	_, contendErr := driver.Lock(ctx, id, 10*time.Second, 100*time.Millisecond)
	contention := errors.Is(contendErr, session.ErrLockTimeout)
	if err := lock.Release(ctx); err != nil {
		return "", err
	}
	doubleRelease := errors.Is(lock.Release(ctx), session.ErrLockNotHeld)

	// 文件锁把“超过 TTL 的锁文件”视为过期；持锁方和等待方使用同一 TTL 才能触发接管。
	stale, err := driver.Lock(ctx, id, 100*time.Millisecond, 200*time.Millisecond)
	if err != nil {
		return "", err
	}
	time.Sleep(300 * time.Millisecond)
	fresh, err := driver.Lock(ctx, id, 100*time.Millisecond, 200*time.Millisecond)
	if err != nil {
		return "", err
	}
	staleRelease := errors.Is(stale.Release(ctx), session.ErrLockNotHeld)
	if err := fresh.Release(ctx); err != nil {
		return "", err
	}
	return fmt.Sprintf("held=true contention=%t released=true re-release-not-held=%t expired-takeover=%t stale-release-not-held=%t",
		contention, doubleRelease, fresh != nil, staleRelease), nil
}

// idCookieScenario verifies the full session ID Cookie attribute set and that
// the server-side payload lifetime is independent of the browser attributes.
func idCookieScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-id-cookie-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	cfg.Lifetime = 2 * time.Hour
	cfg.Cookie.Name = "admin_session"
	cfg.Cookie.Path = "/app"
	cfg.Cookie.Domain = "example.test"
	cfg.Cookie.Secure = true
	cfg.Cookie.HTTPOnly = true
	cfg.Cookie.SameSite = "strict"
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	engine := sessionEngine(manager, func(group *gin.RouterGroup) {
		group.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	})
	_, cookies, err := roundTrip(engine, nil, http.MethodGet, "/")
	if err != nil {
		return "", err
	}
	c := firstCookie(cookies, "admin_session")
	if c == nil {
		return "", errors.New("id-cookie scenario: admin_session cookie missing")
	}
	_, statErr := os.Stat(filepath.Join(root, c.Value))
	return fmt.Sprintf("name=%s path=%s domain=%s secure=%t httpOnly=%t sameSite=%s maxAge-positive=%t server=%t",
		c.Name, c.Path, c.Domain, c.Secure, c.HttpOnly, sameSiteName(c.SameSite), c.MaxAge > 0, statErr == nil), nil
}

// expireOnCloseScenario verifies that expire-on-close only omits browser Cookie
// expiry while the server payload keeps its SESSION_LIFETIME deadline.
func expireOnCloseScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-expire-on-close-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	cfg.Lifetime = 2 * time.Hour
	cfg.ExpireOnClose = true
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	engine := sessionEngine(manager, func(group *gin.RouterGroup) {
		group.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	})
	_, cookies, err := roundTrip(engine, nil, http.MethodGet, "/")
	if err != nil {
		return "", err
	}
	c := firstCookie(cookies, session.DefaultCookieName)
	if c == nil {
		return "", errors.New("expire-on-close scenario: session cookie missing")
	}
	browserSession := c.MaxAge == 0 && c.Expires.IsZero()
	driver, err := session.ResolveDriver("file", manager.Config())
	if err != nil {
		return "", err
	}
	payload, err := driver.Read(context.Background(), c.Value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("browser-session=%t server-persistent=%t", browserSession, payload.ExpiresAt != nil), nil
}

// queuedCookiesScenario verifies that business cookies flush after the session
// ID cookie written by StartSession.
func queuedCookiesScenario() (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	engine := sessionEngine(manager, func(group *gin.RouterGroup) {
		group.GET("/", func(c *gin.Context) {
			if _, err := cookiepkg.QueueMakeFrom(c, "locale", "zh-CN", 60*24,
				cookiepkg.Path("/"), cookiepkg.HTTPOnly(false), cookiepkg.SameSite(cookiepkg.SameSiteLax)); err != nil {
				c.String(http.StatusInternalServerError, "%v", err)
				return
			}
			c.String(http.StatusOK, "queued")
		})
	})
	_, cookies, err := roundTrip(engine, nil, http.MethodGet, "/")
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(cookies))
	for _, c := range cookies {
		names = append(names, c.Name)
	}
	return "order=" + strings.Join(names, ","), nil
}

// encryptionScenario verifies that SESSION_ENCRYPT hides payload bytes while the
// client cookie keeps carrying only the opaque session ID.
func encryptionScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-encryption-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	cfg.Encrypt = true
	cfg.Encryptor = demoEncryptor{}
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	w := httptest.NewRecorder()
	store, err := manager.Start(ctx, httptest.NewRequest(http.MethodGet, "/", nil), w)
	if err != nil {
		return "", err
	}
	store.Put("user_id", int64(1001))
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(root, store.ID()))
	if err != nil {
		return "", err
	}
	encrypted := !bytes.Contains(raw, []byte("user_id"))
	cookie := firstCookie(w.Result().Cookies(), session.DefaultCookieName)
	opaque := cookie != nil && cookie.Value == store.ID()
	restored, err := restoreSession(manager, store.ID())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("encrypted=%t cookie-opaque=%t restored=%v", encrypted, opaque, restored.Get("user_id")), nil
}

// encryptorScenario pins the NopEncryptor copy semantics for both directions.
func encryptorScenario() (string, error) {
	nop := session.NopEncryptor{}
	ctx := context.Background()
	plain := []byte("secret-value")
	encrypted, err := nop.Encrypt(ctx, plain)
	if err != nil {
		return "", err
	}
	plain[0] = 'X'
	encryptCopy := string(encrypted) == "secret-value"
	decrypted, err := nop.Decrypt(ctx, []byte("payload"))
	if err != nil {
		return "", err
	}
	decryptCopy := string(decrypted) == "payload"
	decrypted[0] = 'Y'
	return fmt.Sprintf("encrypt-copy=%t decrypt-copy=%t", encryptCopy, decryptCopy), nil
}

// customEncryptorScenario injects a custom Encryptor and verifies the round trip.
func customEncryptorScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-custom-encryptor-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	encryptor := markerEncryptor{marker: "enc:"}
	cfg := session.DefaultConfig()
	cfg.Files = root
	cfg.Encrypt = true
	cfg.Encryptor = encryptor
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	store := freshStore(manager)
	store.Put("user_id", int64(1001))
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(root, store.ID()))
	if err != nil {
		return "", err
	}
	restored, err := restoreSession(manager, store.ID())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("marker=%t restored=%v", bytes.HasPrefix(raw, []byte(encryptor.marker)), restored.Get("user_id")), nil
}

// sensitiveErrorScenario verifies SensitiveError redacts encryption and
// decryption failures while remaining detectable via errors.Is/As.
func sensitiveErrorScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-sensitive-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)

	encryptCfg := session.DefaultConfig()
	encryptCfg.Files = root
	encryptCfg.Encrypt = true
	encryptCfg.Encryptor = demoEncryptor{failEncrypt: true}
	encryptDriver, err := session.NewFileDriver(encryptCfg)
	if err != nil {
		return "", err
	}
	encryptID := newDemoSessionID()
	writeErr := encryptDriver.Write(ctx, encryptID, session.Payload{
		ID: encryptID, Values: map[string]any{"token": "secret-plaintext"},
	}, &expires)

	decryptCfg := session.DefaultConfig()
	decryptCfg.Files = root
	decryptCfg.Encrypt = true
	decryptCfg.Encryptor = demoEncryptor{failDecrypt: true}
	decryptDriver, err := session.NewFileDriver(decryptCfg)
	if err != nil {
		return "", err
	}
	decryptID := newDemoSessionID()
	if err := decryptDriver.Write(ctx, decryptID, session.Payload{
		ID: decryptID, Values: map[string]any{"token": "secret-plaintext"},
	}, &expires); err != nil {
		return "", err
	}
	_, readErr := decryptDriver.Read(ctx, decryptID)

	var sensitive session.SensitiveError
	encryptTyped := errors.As(writeErr, &sensitive) && errors.Is(writeErr, session.ErrEncryptionFailed)
	decryptTyped := errors.As(readErr, &sensitive) && errors.Is(readErr, session.ErrDecryptionFailed)
	encryptRedacted := writeErr != nil && !strings.Contains(writeErr.Error(), "secret")
	decryptRedacted := readErr != nil && !strings.Contains(readErr.Error(), "secret")
	return fmt.Sprintf("encrypt-redacted=%t encrypt-typed=%t decrypt-redacted=%t decrypt-typed=%t",
		encryptRedacted, encryptTyped, decryptRedacted, decryptTyped), nil
}

// driverContractScenario exercises the Driver Read, Write, Destroy, and GC
// contract through a custom implementation and asserts the built-in drivers.
func driverContractScenario() (string, error) {
	var (
		_ session.Driver = (*session.FileDriver)(nil)
		_ session.Driver = (*session.RedisDriver)(nil)
	)
	ctx := context.Background()
	driver := newMemoryDriver()
	manager, err := session.NewManager(session.DefaultConfig(), driver)
	if err != nil {
		return "", err
	}
	store := freshStore(manager)
	store.Put("counter", int64(7))
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	restored, err := restoreSession(manager, store.ID())
	if err != nil {
		return "", err
	}
	wrote := restored.Get("counter") == int64(7)
	if err := driver.GC(ctx, time.Now()); err != nil {
		return "", err
	}
	if err := driver.Destroy(ctx, store.ID()); err != nil {
		return "", err
	}
	_, readErr := driver.Read(ctx, store.ID())
	return fmt.Sprintf("file=true redis=true custom=true wrote=%t destroyed=%t",
		wrote, errors.Is(readErr, session.ErrSessionNotFound)), nil
}

// lockerContractScenario asserts the optional Locker and Lock contracts.
func lockerContractScenario() (string, error) {
	var (
		_ session.Locker = (*session.FileDriver)(nil)
		_ session.Locker = (*session.RedisDriver)(nil)
		_ session.Lock   = demoLock{}
	)
	return "locker=file,redis lock-release=true", nil
}

// extendScenario registers a custom driver and resolves it through a manager.
func extendScenario() (string, error) {
	session.Extend(demoMemoryDriverName, func(session.Config) (session.Driver, error) {
		return newMemoryDriver(), nil
	})
	manager, err := session.NewManager(session.Config{Driver: demoMemoryDriverName}, nil)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	store := freshStore(manager)
	store.Put("user_id", int64(1001))
	if err := store.Save(ctx); err != nil {
		return "", err
	}
	restored, err := restoreSession(manager, store.ID())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("driver=%s resolved=true restored=%v", demoMemoryDriverName, restored.Get("user_id")), nil
}

// extendValidationScenario verifies invalid registrations are ignored and that a
// repeated name follows the last-registration-wins rule.
func extendValidationScenario() (string, error) {
	session.Extend("", func(session.Config) (session.Driver, error) {
		return newMemoryDriver(), nil
	})
	_, emptyErr := session.ResolveDriver("", session.Config{})
	emptyIgnored := errors.Is(emptyErr, session.ErrDriverNotFound)

	session.Extend("demo-nil-driver", nil)
	_, nilErr := session.ResolveDriver("demo-nil-driver", session.Config{Driver: "demo-nil-driver"})
	nilIgnored := errors.Is(nilErr, session.ErrDriverNotFound)

	session.Extend("demo-override", func(session.Config) (session.Driver, error) {
		return &markerDriver{label: "first"}, nil
	})
	session.Extend("demo-override", func(session.Config) (session.Driver, error) {
		return &markerDriver{label: "second"}, nil
	})
	resolved, err := session.ResolveDriver("demo-override", session.Config{Driver: "demo-override"})
	if err != nil {
		return "", err
	}
	marker, ok := resolved.(*markerDriver)
	override := "missing"
	if ok {
		override = marker.label
	}
	return fmt.Sprintf("empty-ignored=%t nil-ignored=%t override=%s", emptyIgnored, nilIgnored, override), nil
}

// unknownDriverScenario verifies unknown drivers report ErrDriverNotFound.
func unknownDriverScenario() (string, error) {
	_, resolveErr := session.ResolveDriver("demo-unknown", session.DefaultConfig())
	_, managerErr := session.NewManager(session.Config{Driver: "demo-unknown"}, nil)
	return fmt.Sprintf("resolve=%t manager=%t",
		errors.Is(resolveErr, session.ErrDriverNotFound), errors.Is(managerErr, session.ErrDriverNotFound)), nil
}

// errorConstantsScenario maps each documented error constant to a reachable call.
func errorConstantsScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-errors-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	driver, err := session.NewFileDriver(cfg)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	id := newDemoSessionID()
	_, invalidIDErr := driver.Read(ctx, "short-id")
	invalidExpiresErr := driver.Write(ctx, id, session.Payload{ID: id}, nil)
	_, driverErr := session.ResolveDriver("demo-missing", cfg)
	var zeroManager session.Manager
	_, invalidConfigErr := zeroManager.Start(ctx, httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	store := freshStore(manager)
	store.Put("count", "text")
	_, invalidTypeErr := store.Increment("count")
	lock, err := driver.Lock(ctx, id, time.Second, 100*time.Millisecond)
	if err != nil {
		return "", err
	}
	if err := lock.Release(ctx); err != nil {
		return "", err
	}
	notHeldErr := lock.Release(ctx)
	return fmt.Sprintf("invalid-config=%t invalid-id=%t invalid-expires=%t driver-not-found=%t invalid-type=%t lock-not-held=%t",
		errors.Is(invalidConfigErr, session.ErrInvalidConfig), errors.Is(invalidIDErr, session.ErrInvalidSessionID),
		errors.Is(invalidExpiresErr, session.ErrInvalidExpiresAt), errors.Is(driverErr, session.ErrDriverNotFound),
		errors.Is(invalidTypeErr, session.ErrInvalidValueType), errors.Is(notHeldErr, session.ErrLockNotHeld)), nil
}

// recoverableErrorsScenario verifies Manager.Start recovers each recoverable read
// error into a fresh session while returning fatal read errors.
func recoverableErrorsScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-recoverable-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	recoverable := []error{
		session.ErrSessionNotFound,
		session.ErrSessionExpired,
		session.ErrPayloadMalformed,
		session.ErrPayloadDeserialize,
		session.ErrDecryptionFailed,
		session.ErrInvalidSessionID,
	}
	recovered := 0
	for _, injected := range recoverable {
		manager, err := session.NewManager(session.Config{Files: root}, scriptedDriver{readErr: injected})
		if err != nil {
			return "", err
		}
		requested := newDemoSessionID()
		store, err := restoreSession(manager, requested)
		if err != nil {
			return "", fmt.Errorf("recover %v: %w", injected, err)
		}
		if store.ID() != requested {
			recovered++
		}
	}
	fatalManager, err := session.NewManager(session.Config{Files: root}, scriptedDriver{readErr: errors.New("fatal read")})
	if err != nil {
		return "", err
	}
	_, fatalErr := restoreSession(fatalManager, newDemoSessionID())
	return fmt.Sprintf("recovered=%d fatal=%t", recovered, fatalErr != nil), nil
}

// laravelCompatibilityScenario references every documented Laravel mapping and
// reports the lock-based blocking correspondence.
func laravelCompatibilityScenario() (string, error) {
	var (
		_ = session.Get
		_ = session.Put
		_ = session.Has
		_ = session.Exists
		_ = session.Missing
		_ = session.Pull
		_ = session.Forget
		_ = session.Flush
		_ = session.Flash
		_ = session.Now
		_ = session.Reflash
		_ = session.Keep
		_ = session.Regenerate
		_ = session.Invalidate
		_ = (*session.Store).All
		_ = (*session.Store).Only
		_ = (*session.Store).Except
		_ = (*session.Store).Increment
		_ = (*session.Store).Decrement
	)
	cfg := session.DefaultConfig()
	blocking := cfg.Lock.TTL > 0 && cfg.Lock.Wait > 0
	return fmt.Sprintf("mapped=19 blocking=%t", blocking), nil
}

// demoEncryptor is a reversible XOR encryptor used to observe payload encryption.
type demoEncryptor struct {
	failEncrypt bool
	failDecrypt bool
}

// Encrypt XORs the payload, or fails when failEncrypt is set.
func (e demoEncryptor) Encrypt(_ context.Context, plaintext []byte) ([]byte, error) {
	if e.failEncrypt {
		return nil, errors.New("secret plaintext")
	}
	return xorBytes(plaintext), nil
}

// Decrypt XORs the payload, or fails when failDecrypt is set.
func (e demoEncryptor) Decrypt(_ context.Context, ciphertext []byte) ([]byte, error) {
	if e.failDecrypt {
		return nil, errors.New("secret ciphertext")
	}
	return xorBytes(ciphertext), nil
}

// markerEncryptor prefixes payloads with a fixed marker to prove injection.
type markerEncryptor struct {
	marker string
}

// Encrypt prefixes the plaintext with the marker.
func (e markerEncryptor) Encrypt(_ context.Context, plaintext []byte) ([]byte, error) {
	out := make([]byte, 0, len(e.marker)+len(plaintext))
	out = append(out, e.marker...)
	return append(out, plaintext...), nil
}

// Decrypt strips the marker or fails when it is absent.
func (e markerEncryptor) Decrypt(_ context.Context, ciphertext []byte) ([]byte, error) {
	if !bytes.HasPrefix(ciphertext, []byte(e.marker)) {
		return nil, errors.New("missing payload marker")
	}
	return ciphertext[len(e.marker):], nil
}

// xorBytes applies a symmetric XOR mask.
func xorBytes(input []byte) []byte {
	out := make([]byte, len(input))
	for i, b := range input {
		out[i] = b ^ 0x5a
	}
	return out
}

// memoryDriver is a minimal in-process Driver used by the extend and contract demos.
type memoryDriver struct {
	mu     sync.Mutex
	values map[string]session.Payload
}

// newMemoryDriver creates an empty in-process driver.
func newMemoryDriver() *memoryDriver {
	return &memoryDriver{values: make(map[string]session.Payload)}
}

// Read returns the stored payload or ErrSessionNotFound.
func (d *memoryDriver) Read(_ context.Context, id string) (session.Payload, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	payload, ok := d.values[id]
	if !ok {
		return session.Payload{}, session.ErrSessionNotFound
	}
	return payload, nil
}

// Write stores a copy of the payload.
func (d *memoryDriver) Write(_ context.Context, id string, payload session.Payload, expiresAt *time.Time) error {
	if expiresAt == nil || payload.ID != id {
		return session.ErrInvalidSessionID
	}
	payload.ExpiresAt = expiresAt
	d.mu.Lock()
	defer d.mu.Unlock()
	d.values[id] = payload
	return nil
}

// Destroy removes the stored payload; a missing record is not an error.
func (d *memoryDriver) Destroy(_ context.Context, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.values, id)
	return nil
}

// GC removes expired records.
func (d *memoryDriver) GC(_ context.Context, before time.Time) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for id, payload := range d.values {
		if payload.ExpiresAt != nil && !payload.ExpiresAt.After(before) {
			delete(d.values, id)
		}
	}
	return nil
}

// markerDriver is a labeled Driver used to observe last-registration-wins.
type markerDriver struct {
	label string
}

// Read reports no stored payload.
func (d *markerDriver) Read(context.Context, string) (session.Payload, error) {
	return session.Payload{}, session.ErrSessionNotFound
}

// Write accepts and drops the payload.
func (d *markerDriver) Write(context.Context, string, session.Payload, *time.Time) error {
	return nil
}

// Destroy accepts a missing record.
func (d *markerDriver) Destroy(context.Context, string) error { return nil }

// GC is a no-op.
func (d *markerDriver) GC(context.Context, time.Time) error { return nil }

// scriptedDriver returns a fixed read error to exercise recovery classification.
type scriptedDriver struct {
	readErr error
}

// Read returns the scripted error.
func (d scriptedDriver) Read(context.Context, string) (session.Payload, error) {
	return session.Payload{}, d.readErr
}

// Write accepts and drops the payload.
func (d scriptedDriver) Write(context.Context, string, session.Payload, *time.Time) error {
	return nil
}

// Destroy accepts a missing record.
func (d scriptedDriver) Destroy(context.Context, string) error { return nil }

// GC is a no-op.
func (d scriptedDriver) GC(context.Context, time.Time) error { return nil }

// requestWithSessionID builds a request that carries the given session ID.
func requestWithSessionID(id string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: session.DefaultCookieName, Value: id})
	return r
}

// sameSiteName renders an http.SameSite value for stable scenario output.
func sameSiteName(mode http.SameSite) string {
	switch mode {
	case http.SameSiteDefaultMode:
		return "default"
	case http.SameSiteLaxMode:
		return "lax"
	case http.SameSiteStrictMode:
		return "strict"
	case http.SameSiteNoneMode:
		return "none"
	}
	return "unknown"
}
