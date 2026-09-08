package config

func init() {
	Add("database", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Database Connection Name
				|--------------------------------------------------------------------------
				|
				| Here you may specify which of the database connections below should be
				| used as your default connection for database operations.
				|
			*/
			"default": Env("DB_CONNECTION", "sqlite"),

			/*
				|--------------------------------------------------------------------------
				| Database Connections
				|--------------------------------------------------------------------------
				|
				| Each connection listed here may be selected by name. Prismgo ships with
				| common local defaults so new applications can run quickly.
				|
			*/
			"connections": map[string]interface{}{
				"sqlite": map[string]interface{}{
					"driver":             "sqlite",
					"database":           Env("DB_DATABASE", "storage/database.sqlite"),
					"max_open_conns":     Env("DB_MAX_OPEN_CONNS", 1),
					"max_idle_conns":     Env("DB_MAX_IDLE_CONNS", 1),
					"conn_max_lifetime":  Env("DB_CONN_MAX_LIFETIME", "1h"),
					"conn_max_idle_time": Env("DB_CONN_MAX_IDLE_TIME", "10m"),
				},
				"mysql": map[string]interface{}{
					"driver":             "mysql",
					"host":               Env("DB_HOST", "127.0.0.1"),
					"port":               Env("DB_PORT", 3306),
					"database":           Env("DB_DATABASE", "prismgo"),
					"username":           Env("DB_USERNAME", "root"),
					"password":           Env("DB_PASSWORD", ""),
					"charset":            Env("DB_CHARSET", "utf8mb4"),
					"timezone":           Env("DB_TIMEZONE", ""),
					"collation":          Env("DB_COLLATION", ""),
					"parse_time":         Env("DB_PARSE_TIME", true),
					"loc":                Env("DB_LOC", "Local"),
					"unix_socket":        Env("DB_UNIX_SOCKET", ""),
					"prefix":             Env("DB_PREFIX", ""),
					"strict":             Env("DB_STRICT", false),
					"modes":              Env("DB_MODES", ""),
					"isolation_level":    Env("DB_ISOLATION_LEVEL", ""),
					"engine":             Env("DB_ENGINE", ""),
					"dsn":                Env("DB_DSN", ""),
					"max_open_conns":     Env("DB_MAX_OPEN_CONNS", 30),
					"max_idle_conns":     Env("DB_MAX_IDLE_CONNS", 10),
					"conn_max_lifetime":  Env("DB_CONN_MAX_LIFETIME", "1h"),
					"conn_max_idle_time": Env("DB_CONN_MAX_IDLE_TIME", "10m"),
					"ssl": map[string]interface{}{
						"ca":                   Env("DB_SSL_CA", ""),
						"cert":                 Env("DB_SSL_CERT", ""),
						"key":                  Env("DB_SSL_KEY", ""),
						"insecure_skip_verify": Env("DB_SSL_INSECURE_SKIP_VERIFY", false),
					},
					"options": map[string]interface{}{
						"timeout":       Env("DB_OPTION_TIMEOUT", ""),
						"read_timeout":  Env("DB_OPTION_READ_TIMEOUT", ""),
						"write_timeout": Env("DB_OPTION_WRITE_TIMEOUT", ""),
					},
				},
			},

			/*
				|--------------------------------------------------------------------------
				| Redis Databases
				|--------------------------------------------------------------------------
				|
				| Redis connections are configured here so cache, queue, session, and
				| Horizon may refer to a shared connection by name.
				|
			*/
			"redis": map[string]interface{}{
				"client":  Env("REDIS_CLIENT", "go"),
				"options": map[string]interface{}{},
				"default": map[string]interface{}{
					"url":           Env("REDIS_URL", ""),
					"scheme":        Env("REDIS_SCHEME", ""),
					"host":          Env("REDIS_HOST", "127.0.0.1"),
					"port":          Env("REDIS_PORT", "6379"),
					"username":      Env("REDIS_USERNAME", ""),
					"password":      Env("REDIS_PASSWORD", ""),
					"database":      Env("REDIS_MAIN_DB", 1),
					"name":          Env("REDIS_NAME", ""),
					"timeout":       Env("REDIS_TIMEOUT", ""),
					"read_timeout":  Env("REDIS_READ_TIMEOUT", ""),
					"write_timeout": Env("REDIS_WRITE_TIMEOUT", ""),
					"max_retries":   Env("REDIS_MAX_RETRIES", ""),
				},
				"cache": map[string]interface{}{
					"url":           Env("REDIS_CACHE_URL", ""),
					"scheme":        Env("REDIS_CACHE_SCHEME", ""),
					"host":          Env("REDIS_HOST", "127.0.0.1"),
					"port":          Env("REDIS_PORT", "6379"),
					"username":      Env("REDIS_USERNAME", ""),
					"password":      Env("REDIS_PASSWORD", ""),
					"database":      Env("REDIS_CACHE_DB", 0),
					"name":          Env("REDIS_CACHE_NAME", ""),
					"timeout":       Env("REDIS_CACHE_TIMEOUT", ""),
					"read_timeout":  Env("REDIS_CACHE_READ_TIMEOUT", ""),
					"write_timeout": Env("REDIS_CACHE_WRITE_TIMEOUT", ""),
					"max_retries":   Env("REDIS_CACHE_MAX_RETRIES", ""),
				},
			},
		}
	})
}
