package config

func init() {
	Add("queue", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Queue Connection Name
				|--------------------------------------------------------------------------
				|
				| Prismgo's queue API supports multiple backends. This value determines
				| which connection should be used by default.
				|
			*/
			"default": Env("QUEUE_CONNECTION", "sync"),

			/*
				|--------------------------------------------------------------------------
				| Queue Payload Encoding
				|--------------------------------------------------------------------------
				|
				| Queue envelopes, failed jobs, batches, and chain metadata may use a
				| dedicated encoding driver. Empty values inherit encoding.default.
				|
			*/
			"encoding": Env("QUEUE_ENCODING", ""),

			/*
				|--------------------------------------------------------------------------
				| Failed Jobs, Batches, And Restart Signals
				|--------------------------------------------------------------------------
				|
				| These options configure where failed jobs and batch metadata are
				| stored, plus the cache key used to broadcast queue restarts.
				|
			*/
			"failed": map[string]interface{}{
				"driver": Env("QUEUE_FAILED_DRIVER", "memory"),
				"store":  Env("QUEUE_FAILED_STORE", "default"),
				"prefix": Env("QUEUE_FAILED_PREFIX", "prismgo_queue"),
				"ttl":    Env("QUEUE_FAILED_TTL", 0),
			},
			"batching": map[string]interface{}{
				"driver": Env("QUEUE_BATCHING_DRIVER", "memory"),
				"store":  Env("QUEUE_BATCHING_STORE", "default"),
				"prefix": Env("QUEUE_BATCHING_PREFIX", "prismgo_queue"),
				"ttl":    Env("QUEUE_BATCHING_TTL", 0),
			},
			"restart": map[string]interface{}{
				"cache": Env("QUEUE_RESTART_CACHE", ""),
				"key":   Env("QUEUE_RESTART_KEY", "prismgo:queue:restart"),
			},

			/*
				|--------------------------------------------------------------------------
				| Queue Connections
				|--------------------------------------------------------------------------
				|
				| Each queue connection may use a different driver. The sync connection
				| runs jobs immediately, while Redis and RabbitMQ run jobs asynchronously.
				|
			*/
			"connections": map[string]interface{}{
				"sync": map[string]interface{}{
					"driver": "sync",
					"queue":  Env("SYNC_QUEUE", "default"),
				},
				"database": map[string]interface{}{
					"driver": "database",
					"table":  Env("DATABASE_QUEUE_TABLE", "jobs"),
					"queue":  Env("DATABASE_QUEUE", "default"),
				},
				"redis": map[string]interface{}{
					"driver":      "redis",
					"queue":       Env("REDIS_QUEUE", "default"),
					"prefix":      Env("REDIS_QUEUE_PREFIX", "prismgo_queue"),
					"connection":  Env("REDIS_QUEUE_CONNECTION", "default"),
					"retry_after": Env("REDIS_QUEUE_RETRY_AFTER", 90),
					"block_for":   Env("REDIS_QUEUE_BLOCK_FOR", 0),
				},
				"rabbitmq": map[string]interface{}{
					"driver":                     "rabbitmq",
					"queue":                      Env("RABBITMQ_QUEUE", "default"),
					"block_for":                  Env("RABBITMQ_BLOCK_FOR", 1),
					"url":                        Env("RABBITMQ_URL", ""),
					"scheme":                     Env("RABBITMQ_SCHEME", "amqp"),
					"host":                       Env("RABBITMQ_HOST", "127.0.0.1"),
					"port":                       Env("RABBITMQ_PORT", "5672"),
					"username":                   Env("RABBITMQ_USERNAME", ""),
					"password":                   Env("RABBITMQ_PASSWORD", ""),
					"vhost":                      Env("RABBITMQ_VHOST", "/"),
					"exchange":                   Env("RABBITMQ_EXCHANGE", "prismgo.queue"),
					"exchange_type":              Env("RABBITMQ_EXCHANGE_TYPE", "direct"),
					"declare":                    Env("RABBITMQ_DECLARE", true),
					"exchange_durable":           Env("RABBITMQ_EXCHANGE_DURABLE", true),
					"queue_durable":              Env("RABBITMQ_QUEUE_DURABLE", true),
					"queue_max_priority":         Env("RABBITMQ_QUEUE_MAX_PRIORITY", 0),
					"message_persistent":         Env("RABBITMQ_MESSAGE_PERSISTENT", true),
					"auto_delete":                Env("RABBITMQ_AUTO_DELETE", false),
					"exclusive":                  Env("RABBITMQ_EXCLUSIVE", false),
					"no_wait":                    Env("RABBITMQ_NO_WAIT", false),
					"confirm":                    Env("RABBITMQ_CONFIRM", true),
					"delay_mode":                 Env("RABBITMQ_DELAY_MODE", "plugin"),
					"delay_buckets":              Env("RABBITMQ_DELAY_BUCKETS", "5,10,30,60,300,900,3600"),
					"prefetch":                   Env("RABBITMQ_PREFETCH", 1),
					"heartbeat":                  Env("RABBITMQ_HEARTBEAT", 10),
					"publish_timeout":            Env("RABBITMQ_PUBLISH_TIMEOUT", 5),
					"publish_channels":           Env("RABBITMQ_PUBLISH_CHANNELS", 1),
					"reconnect_min_delay":        Env("RABBITMQ_RECONNECT_MIN_DELAY", "100ms"),
					"reconnect_max_delay":        Env("RABBITMQ_RECONNECT_MAX_DELAY", "5s"),
					"restart_queue":              Env("RABBITMQ_RESTART_QUEUE", "prismgo.queue.restart"),
					"restart_enabled":            Env("RABBITMQ_RESTART_ENABLED", true),
					"restart_poll_interval":      Env("RABBITMQ_RESTART_POLL_INTERVAL", "1s"),
					"topology_cache_ttl":         Env("RABBITMQ_TOPOLOGY_CACHE_TTL", 0),
					"topology_cache_max_entries": Env("RABBITMQ_TOPOLOGY_CACHE_MAX_ENTRIES", 0),
				},
			},
		}
	})
}
