package config

func init() {
	Add("session", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Session Driver
				|--------------------------------------------------------------------------
				|
				| This option controls the session backend used to persist session data
				| between requests to your application.
				|
			*/
			"driver": Env("SESSION_DRIVER", "file"),

			/*
				|--------------------------------------------------------------------------
				| Session Lifetime And Payload
				|--------------------------------------------------------------------------
				|
				| These options control session expiration and how server-side session
				| payloads are encrypted or encoded.
				|
			*/
			"lifetime":        Env("SESSION_LIFETIME", 120),
			"expire_on_close": Env("SESSION_EXPIRE_ON_CLOSE", false),
			"encrypt":         Env("SESSION_ENCRYPT", false),
			"encoding":        Env("SESSION_ENCODING", ""),
			"connection":      Env("SESSION_CONNECTION", "default"),
			"prefix":          Env("SESSION_PREFIX", "prismgo_session"),

			/*
				|--------------------------------------------------------------------------
				| Session Cookie
				|--------------------------------------------------------------------------
				|
				| These options determine how the session cookie is named and sent back
				| to the browser for each request.
				|
			*/
			"cookie":    Env("SESSION_COOKIE", "prismgo_session"),
			"path":      Env("SESSION_PATH", "/"),
			"domain":    Env("SESSION_DOMAIN", ""),
			"secure":    Env("SESSION_SECURE_COOKIE", false),
			"http_only": Env("SESSION_HTTP_ONLY", true),
			"same_site": Env("SESSION_SAME_SITE", "lax"),

			/*
				|--------------------------------------------------------------------------
				| Session Storage And Blocking
				|--------------------------------------------------------------------------
				|
				| File sessions use the files path. Blocking requests may use locks to
				| prevent concurrent writes to the same session.
				|
			*/
			"files":        Env("SESSION_FILES", "storage/framework/sessions"),
			"lock_seconds": Env("SESSION_LOCK_SECONDS", 10),
			"lock_wait":    Env("SESSION_LOCK_WAIT_SECONDS", 10),
		}
	})
}
