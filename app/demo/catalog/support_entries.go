package catalog

func supportEntries() []Entry {
	return []Entry{
		baselineEntry("support", "Feature overview", "功能概览", "Feature Overview", "demo:support list", "list", "TestSupportDemo", LevelHermetic, StatusPlanned),
	}
}
