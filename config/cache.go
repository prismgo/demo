package config

func init() {
	Add("cache", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Cache Store
				|--------------------------------------------------------------------------
				|
				| This option controls the cache store that will be used by default by
				| cache operations throughout the application.
				|
			*/
			"default": Env("CACHE_STORE", "memory"),

			/*
				|--------------------------------------------------------------------------
				| Cache Encoding And Prefix
				|--------------------------------------------------------------------------
				|
				| Cache payloads may use a dedicated encoding driver. When this value
				| is empty, Prismgo will inherit the global encoding.default setting.
				|
			*/
			"encoding": Env("CACHE_ENCODING", ""),
			"prefix":   Env("CACHE_PREFIX", "prismgo_cache"),

			/*
				|--------------------------------------------------------------------------
				| Rate Limiter Store
				|--------------------------------------------------------------------------
				|
				| The limiter driver may use a different backend from the default cache
				| store when your application needs isolated rate-limit state.
				|
			*/
			"limiter": map[string]interface{}{
				"driver": Env("CACHE_LIMITER_DRIVER", "memory"),
			},

			/*
				|--------------------------------------------------------------------------
				| Cache Stores
				|--------------------------------------------------------------------------
				|
				| Here you may define all of the cache stores for your application. Each
				| store contains the driver name and driver-specific configuration.
				|
			*/
			"stores": map[string]interface{}{
				"memory": map[string]interface{}{
					"driver":           "memory",
					"prefix":           Env("CACHE_MEMORY_PREFIX", "memory"),
					"default_ttl":      Env("CACHE_MEMORY_TTL", 0),
					"cleanup_interval": Env("CACHE_MEMORY_CLEANUP_INTERVAL", 60),
					"events":           Env("CACHE_MEMORY_EVENTS", true),
				},
				"redis": map[string]interface{}{
					"driver":     "redis",
					"prefix":     Env("CACHE_REDIS_PREFIX", "redis"),
					"connection": Env("CACHE_REDIS_CONNECTION", "cache"),
					"events":     Env("CACHE_REDIS_EVENTS", true),
				},
				"file": map[string]interface{}{
					"driver":      "file",
					"prefix":      Env("CACHE_FILE_PREFIX", "file"),
					"path":        Env("CACHE_FILE_PATH", "storage/framework/cache/data"),
					"lock_path":   Env("CACHE_FILE_LOCK_PATH", "storage/framework/cache/locks"),
					"default_ttl": Env("CACHE_FILE_TTL", 0),
					"events":      Env("CACHE_FILE_EVENTS", true),
				},
				"failover": map[string]interface{}{
					"driver": "failover",
					"stores": Env("CACHE_FAILOVER_STORES", "redis,memory"),
					"events": Env("CACHE_FAILOVER_EVENTS", true),
				},
			},

			/*
				|--------------------------------------------------------------------------
				| Cache Locks And Flexible Refresh
				|--------------------------------------------------------------------------
				|
				| These options configure lock key prefixes, retry cadence, and the
				| timeout used by stale-while-revalidate style cache refreshes.
				|
			*/
			"lock": map[string]interface{}{
				"prefix":         Env("CACHE_LOCK_PREFIX", "locks"),
				"retry_sleep_ms": Env("CACHE_LOCK_RETRY_SLEEP_MS", 50),
			},
			"flexible": map[string]interface{}{
				"refresh_timeout": Env("CACHE_FLEXIBLE_REFRESH_TIMEOUT", 30),
			},
		}
	})
}
