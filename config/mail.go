package config

func init() {
	Add("mail", func() map[string]any {
		return map[string]any{
			"host": Env("MAIL_HOST", "127.0.0.1"),
			"port": Env("MAIL_PORT", 1025),
			"from": map[string]any{
				"address": Env("MAIL_FROM_ADDRESS", "noreply@example.com"),
				"name":    Env("MAIL_FROM_NAME", "Prismgo"),
			},
		}
	})
}
