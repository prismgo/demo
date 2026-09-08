package config

func init() {
	Add("app", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Application Name
				|--------------------------------------------------------------------------
				|
				| This value is the name of your application. Prismgo uses it when the
				| framework needs to display or reference the application name.
				|
			*/
			"name": Env("APP_NAME", "Prismgo"),

			/*
				|--------------------------------------------------------------------------
				| Application Environment
				|--------------------------------------------------------------------------
				|
				| This value determines the environment your application is running in.
				| It may influence how services are configured by the framework.
				|
			*/
			"env": Env("APP_ENV", "production"),

			/*
				|--------------------------------------------------------------------------
				| Application Key
				|--------------------------------------------------------------------------
				|
				| This key may be used by encryption services and should be set to a
				| random value before the application is deployed to production.
				|
			*/
			"key": Env("APP_KEY", ""),

			/*
				|--------------------------------------------------------------------------
				| Application Debug Mode
				|--------------------------------------------------------------------------
				|
				| When debug mode is enabled, Prismgo may expose more detailed errors.
				| Disable this option in production environments.
				|
			*/
			"debug": Env("APP_DEBUG", false),

			/*
				|--------------------------------------------------------------------------
				| Application URL
				|--------------------------------------------------------------------------
				|
				| This URL is used by the framework when it needs to generate absolute
				| URLs from the command line or other non-HTTP contexts.
				|
			*/
			"url": Env("APP_URL", "http://localhost:8080"),

			/*
				|--------------------------------------------------------------------------
				| Application Locale Configuration
				|--------------------------------------------------------------------------
				|
				| The application locale determines the default language used by
				| translation services. The fallback locale is used when a line is
				| not available in the current locale.
				|
			*/
			"timezone":        Env("APP_TIMEZONE", "UTC"),
			"locale":          Env("APP_LOCALE", "en"),
			"fallback_locale": Env("APP_FALLBACK_LOCALE", "en"),

			/*
				|--------------------------------------------------------------------------
				| HTTP Server Configuration
				|--------------------------------------------------------------------------
				|
				| These options configure the built-in HTTP server, including timeouts,
				| request limits, proxy trust, access logging, and exception handling.
				|
			*/
			"server": map[string]interface{}{
				"host":                 Env("SERVER_HOST", ""),
				"port":                 Env("SERVER_PORT", 8080),
				"timeout":              Env("SERVER_TIMEOUT", 15),
				"read_timeout":         Env("SERVER_READ_TIMEOUT", "15s"),
				"read_header_timeout":  Env("SERVER_READ_HEADER_TIMEOUT", "5s"),
				"write_timeout":        Env("SERVER_WRITE_TIMEOUT", "30s"),
				"idle_timeout":         Env("SERVER_IDLE_TIMEOUT", "60s"),
				"shutdown_timeout":     Env("SERVER_SHUTDOWN_TIMEOUT", "15s"),
				"max_header_bytes":     Env("SERVER_MAX_HEADER_BYTES", 1048576),
				"max_multipart_memory": Env("SERVER_MAX_MULTIPART_MEMORY", 33554432),
				"trusted_proxies":      Env("SERVER_TRUSTED_PROXIES", ""),
				"client_ip_headers":    Env("SERVER_CLIENT_IP_HEADERS", "X-Forwarded-For,X-Real-IP"),
				"access_log":           Env("SERVER_ACCESS_LOG", true),
				"exception_handler":    Env("SERVER_EXCEPTION_HANDLER", true),
			},
		}
	})
}
