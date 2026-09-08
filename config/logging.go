package config

func init() {
	Add("logging", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Log Channel
				|--------------------------------------------------------------------------
				|
				| This option defines the log channel that should be used when writing
				| messages from the application and framework.
				|
			*/
			"default": Env("LOG_CHANNEL", "stack"),

			/*
				|--------------------------------------------------------------------------
				| Log Channels
				|--------------------------------------------------------------------------
				|
				| Here you may configure the log channels available to your application.
				| Each channel declares a driver and any driver-specific options.
				|
			*/
			"channels": map[string]interface{}{
				"stack": map[string]interface{}{
					"driver":   "stack",
					"channels": []interface{}{"single", "error"},
				},
				"single": map[string]interface{}{
					"driver":    "single",
					"level":     Env("APP_LOGGER_LEVEL", "info"),
					"formatter": Env("APP_LOGGER_FORMATTER", "line"),
					"path":      Env("APP_LOGGER_FILE", "storage/logs/app.log"),
				},
				"daily": map[string]interface{}{
					"driver":    "daily",
					"level":     Env("APP_LOGGER_LEVEL", "info"),
					"formatter": Env("APP_LOGGER_FORMATTER", "line"),
					"path":      Env("APP_LOGGER_FILE", "storage/logs/app.log"),
				},
				"error": map[string]interface{}{
					"driver":    Env("ERROR_LOGGER_DRIVER", "daily"),
					"level":     Env("ERROR_LOGGER_LEVEL", "warn"),
					"formatter": Env("ERROR_LOGGER_FORMATTER", "line"),
					"path":      Env("ERROR_LOGGER_FILE", "storage/logs/error/error.log"),
				},
				"stderr": map[string]interface{}{
					"driver":    "stderr",
					"level":     Env("STDERR_LOGGER_LEVEL", "info"),
					"formatter": Env("STDERR_LOGGER_FORMATTER", "line"),
				},
				"stdout": map[string]interface{}{
					"driver":    "stdout",
					"level":     Env("STDOUT_LOGGER_LEVEL", "info"),
					"formatter": Env("STDOUT_LOGGER_FORMATTER", "line"),
				},
				"null": map[string]interface{}{
					"driver": "null",
				},
			},
		}
	})
}
