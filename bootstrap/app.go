package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
	httpmiddleware "github.com/prismgo/framework/http/middleware"

	appcmd "prismgo-demo/app/cmd"
	apphttp "prismgo-demo/app/http"
	appmiddleware "prismgo-demo/app/http/middleware"
	appschedule "prismgo-demo/app/schedule"

	// Load application migrations and seeders for CLI registration.
	_ "prismgo-demo/database/migrations"
	_ "prismgo-demo/database/seeders"
)

// NewApplication creates the project application instance.
func NewApplication(basePath ...string) *foundation.Application {
	return foundation.Configure(basePath...).
		WithExtensionProviders(Extensions()...).
		WithProviders(Providers()...).
		WithRouting(func(r *foundation.Routing) {
			r.MigrationPaths("database/migrations")
			r.SeedPaths("database/seeders")
			r.Commands(appcmd.CommandFactories()...)
			r.Schedules(appschedule.Register)
			r.Routes(apphttp.RegisterRoutes)
		}).
		WithMiddleware(func(m *foundation.Middleware) {
			m.Prepend(func(engine *gin.Engine) {
				engine.Use(httpmiddleware.RequestID())
				engine.Use(appmiddleware.Example())
			})
		}).
		Create()
}
