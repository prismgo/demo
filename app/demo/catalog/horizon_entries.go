package catalog

func horizonEntries() []Entry {
	return []Entry{
		baselineEntry("horizon", "Running Horizon", "运行 Horizon", "Running Horizon", "./demo/dev test-horizon", "list", "TestHorizonWithRealQueue", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	}
}
