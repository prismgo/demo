package catalog

func starterEntries() []Entry {
	return []Entry{
		manualEntry("starter", "First endpoint", "添加第一个接口", "Add Your First Endpoint", "starter HTTP smoke test", "TestStarterSmoke", LevelScenario, "runs against a clean generated starter"),
	}
}
