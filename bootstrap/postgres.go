package bootstrap

import (
	"github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// postgresDemoProvider installs the PostgreSQL dialector for local Demo scenarios.
type postgresDemoProvider struct{}

func (postgresDemoProvider) Name() string { return "demo.extension.postgres" }

func (postgresDemoProvider) Register(provider.Application) error { return nil }

func (postgresDemoProvider) Boot(app provider.Application) error {
	manager, err := database.ManagerFrom(app.Container())
	if err != nil {
		return err
	}
	manager.Extend("postgres", func(ctx database.DriverContext) (gorm.Dialector, error) {
		return postgres.Open(ctx.DSN), nil
	})
	return nil
}

var _ provider.ServiceProvider = postgresDemoProvider{}
var _ provider.NamedProvider = postgresDemoProvider{}
