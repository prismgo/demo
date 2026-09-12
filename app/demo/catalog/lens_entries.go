package catalog

func lensEntries() []Entry {
	return []Entry{
		manualEntry("lens", "Installation", "安装", "Installation", "Lens workflow verification", "TestLensWorkflow", LevelIntegration, "validated by the Lens toolchain rather than an application command"),
	}
}
