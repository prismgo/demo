package catalog

func facadeEntries() []Entry {
	return []Entry{
		baselineEntry("facade", "Available facades", "可用 Facade", "Available Facades", "demo:facade list", "list", "TestFacadeDemo", LevelCompile, StatusPlanned),
	}
}
