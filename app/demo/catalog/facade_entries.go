package catalog

func facadeEntries() []Entry {
	return []Entry{
		baselineEntry("facade", "Introduction", "简介", "Introduction", "demo:facade list", "introduction", "TestFacadeDemoIntroduction", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "How facades work", "Facade 的工作原理", "How Facades Work", "demo:facade list", "how-facades-work", "TestFacadeDemoHowFacadesWork", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Core facade.Resolve", "核心 Facade：facade.Resolve", "Core Facade: facade.Resolve", "demo:facade list", "core-facade-resolve", "TestFacadeDemoCoreFacadeResolve", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Available facades", "可用 Facade", "Available Facades", "demo:facade list", "list", "TestFacadeDemoAvailableFacades", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Cache facade", "Cache Facade", "Cache Facade", "demo:facade list", "cache-facade", "TestFacadeDemoCacheFacade", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Config facade", "Config Facade", "Config Facade", "demo:facade list", "config-facade", "TestFacadeDemoConfigFacade", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Route facade", "Route Facade", "Route Facade", "demo:facade list", "route-facade", "TestFacadeDemoRouteFacade", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Logger facade", "Logger Facade", "Logger Facade", "demo:facade list", "logger-facade", "TestFacadeDemoLoggerFacade", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Database facade", "Database Facade", "Database Facade", "demo:facade list", "database-facade", "TestFacadeDemoDatabaseFacade", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Filesystem facade", "Filesystem Facade", "Filesystem Facade", "demo:facade list", "filesystem-facade", "TestFacadeDemoFilesystemFacade", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Facade class reference", "Facade 类参考", "Facade Class Reference", "demo:facade list", "class-reference", "TestFacadeDemoClassReference", LevelHermetic, StatusImplemented),
		baselineEntry("facade", "Laravel facade mapping", "Laravel Facade 映射", "Laravel Facade Mapping", "demo:facade list", "laravel-mapping", "TestFacadeDemoLaravelMapping", LevelHermetic, StatusImplemented),
	}
}
