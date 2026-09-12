package catalog

func installationEntries() []Entry {
	return []Entry{
		manualEntry("installation", "Creating an application", "创建应用", "Creating an Application", "installer smoke test", "TestInstallerSmoke", LevelIntegration, "requires a clean temporary checkout and network access"),
	}
}
