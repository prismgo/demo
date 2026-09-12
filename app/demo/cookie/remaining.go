package cookiedemo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	cookiecontract "github.com/prismgo/framework/contracts/cookie"
	encryptioncontract "github.com/prismgo/framework/contracts/encryption"
	"github.com/prismgo/framework/cookie"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
)

// Compile-time assertions keep the demo extensions aligned with the public contracts.
var (
	_ cookiecontract.Signer              = (*demoSecurity)(nil)
	_ encryptioncontract.StringEncrypter = (*demoSecurity)(nil)
)

type demoSecurity struct {
	steps []string
	fail  string
}

func (s *demoSecurity) Sign(_ context.Context, name, value string) (string, error) {
	s.steps = append(s.steps, "sign")
	if s.fail == "sign" {
		return "", errors.New("secret signing input")
	}
	return name + "." + value, nil
}

func (s *demoSecurity) Unsign(_ context.Context, name, value string) (string, error) {
	s.steps = append(s.steps, "verify")
	if s.fail == "verify" || !strings.HasPrefix(value, name+".") {
		return "", errors.New("secret untrusted value")
	}
	return strings.TrimPrefix(value, name+"."), nil
}

func (s *demoSecurity) EncryptString(_ context.Context, value string) (string, error) {
	s.steps = append(s.steps, "encrypt")
	if s.fail == "encrypt" {
		return "", errors.New("secret plaintext")
	}
	return "enc-" + value, nil
}

func (s *demoSecurity) DecryptString(_ context.Context, value string) (string, error) {
	s.steps = append(s.steps, "decrypt")
	if s.fail == "decrypt" || !strings.HasPrefix(value, "enc-") {
		return "", errors.New("secret ciphertext")
	}
	return strings.TrimPrefix(value, "enc-"), nil
}

func runRemaining(name string) (string, error) {
	switch name {
	case "request-cookie", "request-not-found", "request-context", "request-signer", "request-encryptor":
		return readScenario(name)
	case "queue-expire", "queue-forget":
		return queuedDeletion(name)
	case "forget", "deletion-scope":
		return directDeletion(name)
	case "same-site-default", "same-site-modes":
		return sameSiteScenario(name)
	case "security-contracts":
		return "signer=true encryptor=true", nil
	case "outgoing-order", "incoming-order", "passthrough", "attach-security", "queue-security", "sensitive-error":
		return securityScenario(name)
	case "errors":
		return errorScenario(), nil
	case "laravel-compatibility":
		return laravelScenario()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

func readScenario(name string) (string, error) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	switch name {
	case "request-cookie":
		r.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
		value, err := cookie.RequestCookie(r, "theme")
		return "value=" + value, err
	case "request-not-found":
		_, err := cookie.RequestCookie(r, "missing")
		return fmt.Sprintf("missing=%t", errors.Is(err, cookie.ErrCookieNotFound)), nil
	case "request-context":
		r.AddCookie(&http.Cookie{Name: "trace", Value: "value"})
		ctx := context.WithValue(context.Background(), contextKey{}, "request-42")
		value, err := cookie.RequestCookie(r, "trace", cookie.RequestWithContext(ctx), cookie.RequestWithSigner(contextReadSigner{}))
		return "value=" + value, err
	case "request-signer":
		r.AddCookie(&http.Cookie{Name: "theme", Value: "theme.dark"})
		value, err := cookie.RequestCookie(r, "theme", cookie.RequestWithSigner(&demoSecurity{}))
		return "value=" + value, err
	case "request-encryptor":
		r.AddCookie(&http.Cookie{Name: "theme", Value: "enc-dark"})
		value, err := cookie.RequestCookie(r, "theme", cookie.RequestWithEncryptor(&demoSecurity{}))
		return "value=" + value, err
	}
	return "", fmt.Errorf("unknown read scenario %q", name)
}

type contextReadSigner struct{}

func (contextReadSigner) Sign(_ context.Context, _ string, value string) (string, error) {
	return value, nil
}

func (contextReadSigner) Unsign(ctx context.Context, _ string, value string) (string, error) {
	trace, _ := ctx.Value(contextKey{}).(string)
	return value + ":" + trace, nil
}

func queuedDeletion(name string) (string, error) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(previousMode)
	engine := gin.New()
	engine.Use(httpmiddleware.QueuedCookies())
	var scenarioErr error
	engine.GET("/", func(c *gin.Context) {
		if name == "queue-expire" {
			_, scenarioErr = cookie.QueueExpireFrom(c, "theme", cookie.Path("/admin"), cookie.Domain("example.test"))
		} else {
			_, scenarioErr = cookie.QueueForgetFrom(c, "theme", cookie.Path("/admin"), cookie.Domain("example.test"))
		}
		c.Status(http.StatusNoContent)
	})
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if scenarioErr != nil {
		return "", scenarioErr
	}
	return deletionResult(w.Result())
}

func directDeletion(name string) (string, error) {
	w := httptest.NewRecorder()
	if name == "deletion-scope" {
		if err := cookie.Make("theme", "dark", 5, cookie.Path("/admin"), cookie.Domain("example.test")).Attach(w); err != nil {
			return "", err
		}
	}
	if err := cookie.Forget("theme", cookie.Path("/admin"), cookie.Domain("example.test")).Attach(w); err != nil {
		return "", err
	}
	if name == "deletion-scope" {
		cookies := w.Result().Cookies()
		if len(cookies) != 2 {
			return "", fmt.Errorf("deletion scope cookies = %d, want 2", len(cookies))
		}
		return fmt.Sprintf("sameScope=%t deleted=%t", cookies[0].Path == cookies[1].Path && cookies[0].Domain == cookies[1].Domain, cookies[1].MaxAge < 0), nil
	}
	return deletionResult(w.Result())
}

func deletionResult(response *http.Response) (string, error) {
	cookies := response.Cookies()
	if len(cookies) != 1 {
		return "", fmt.Errorf("deletion cookies = %d, want 1", len(cookies))
	}
	c := cookies[0]
	return fmt.Sprintf("name=%s path=%s domain=%s maxAge=%d value=%q", c.Name, c.Path, c.Domain, c.MaxAge, c.Value), nil
}

func sameSiteScenario(name string) (string, error) {
	if name == "same-site-default" {
		defaultCookie, err := cookie.Make("default", "a", 0).ToHTTP()
		if err != nil {
			return "", err
		}
		disabled, err := cookie.Make("disabled", "b", 0, cookie.SameSite(cookie.SameSiteDisabled)).ToHTTP()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("default=%t disabled=%t", !strings.Contains(defaultCookie.String(), "SameSite"), !strings.Contains(disabled.String(), "SameSite")), nil
	}
	modes := []cookie.SameSiteMode{cookie.SameSiteLax, cookie.SameSiteStrict, cookie.SameSiteNone}
	values := make([]string, 0, len(modes))
	for _, mode := range modes {
		c, err := cookie.Make("theme", "dark", 0, cookie.SameSite(mode)).ToHTTP()
		if err != nil {
			return "", err
		}
		_, value, ok := strings.Cut(c.String(), "SameSite=")
		if !ok {
			return "", fmt.Errorf("SameSite mode %q missing from Set-Cookie header %q", mode, c.String())
		}
		values = append(values, value)
	}
	return "modes=" + strings.Join(values, ","), nil
}

func securityScenario(name string) (string, error) {
	s := &demoSecurity{}
	w := httptest.NewRecorder()
	switch name {
	case "outgoing-order", "attach-security":
		if err := cookie.Make("theme", "dark", 0).Attach(w, cookie.WithEncryptor(s), cookie.WithSigner(s)); err != nil {
			return "", err
		}
		return fmt.Sprintf("value=%s steps=%s", w.Result().Cookies()[0].Value, strings.Join(s.steps, ",")), nil
	case "incoming-order":
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "theme", Value: "theme.enc-dark"})
		value, err := cookie.RequestCookie(r, "theme", cookie.RequestWithSigner(s), cookie.RequestWithEncryptor(s))
		return fmt.Sprintf("value=%s steps=%s", value, strings.Join(s.steps, ",")), err
	case "passthrough":
		if err := cookie.Make("theme", "dark", 0).Attach(w, cookie.WithSigner(cookie.PassthroughSecurity{}), cookie.WithEncryptor(cookie.PassthroughSecurity{})); err != nil {
			return "", err
		}
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(w.Result().Cookies()[0])
		value, err := cookie.RequestCookie(r, "theme", cookie.RequestWithSigner(cookie.PassthroughSecurity{}), cookie.RequestWithEncryptor(cookie.PassthroughSecurity{}))
		return "value=" + value, err
	case "queue-security":
		q := cookie.NewQueue(cookie.WithEncryptor(s), cookie.WithSigner(s))
		q.Make("theme", "dark", 0)
		if err := q.Flush(w); err != nil {
			return "", err
		}
		return fmt.Sprintf("value=%s steps=%s cleared=%t", w.Result().Cookies()[0].Value, strings.Join(s.steps, ","), !q.HasQueued("theme")), nil
	case "sensitive-error":
		s.fail = "verify"
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "theme", Value: "secret-client-value"})
		_, err := cookie.RequestCookie(r, "theme", cookie.RequestWithSigner(s))
		var sensitive cookie.SensitiveError
		return fmt.Sprintf("redacted=%t signature=%t typed=%t", err != nil && !strings.Contains(err.Error(), "secret"), errors.Is(err, cookie.ErrCookieSignature), errors.As(err, &sensitive)), nil
	}
	return "", fmt.Errorf("unknown security scenario %q", name)
}

func errorScenario() string {
	_, invalidErr := cookie.Make("bad name", "value", 0).ToHTTP()
	_, missingErr := cookie.RequestCookie(httptest.NewRequest(http.MethodGet, "/", nil), "missing")
	_, queueErr := cookie.QueueMakeFrom(nil, "theme", "dark", 0)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "theme", Value: "untrusted"})
	_, signatureErr := cookie.RequestCookie(r, "theme", cookie.RequestWithSigner(&demoSecurity{fail: "verify"}))
	encryptionErr := cookie.Make("theme", "dark", 0).Attach(httptest.NewRecorder(), cookie.WithEncryptor(&demoSecurity{fail: "encrypt"}))
	_, decryptionErr := cookie.RequestCookie(r, "theme", cookie.RequestWithEncryptor(&demoSecurity{fail: "decrypt"}))
	return fmt.Sprintf("invalid=%t missing=%t queue=%t signature=%t encryption=%t decryption=%t",
		errors.Is(invalidErr, cookie.ErrInvalidCookieName), errors.Is(missingErr, cookie.ErrCookieNotFound),
		errors.Is(queueErr, cookie.ErrQueueNotFound), errors.Is(signatureErr, cookie.ErrCookieSignature),
		errors.Is(encryptionErr, cookie.ErrCookieEncryption), errors.Is(decryptionErr, cookie.ErrCookieDecryption))
}

func laravelScenario() (string, error) {
	w := httptest.NewRecorder()
	if err := cookie.Make("theme", "dark", 5).Attach(w); err != nil {
		return "", err
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(w.Result().Cookies()[0])
	value, err := cookie.RequestCookie(r, "theme")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("make=%s forever=%d read=%s forget=%d", cookie.Make("theme", "dark", 5).Value, cookie.Forever("theme", "dark").Minutes, value, cookie.Forget("theme").MaxAge), nil
}
