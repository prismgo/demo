// Package providers registers the demo application's services.
package providers

import (
	"fmt"

	"github.com/prismgo/framework/contracts/container"
	"github.com/prismgo/framework/contracts/provider"
	"gorm.io/gorm"

	"prismgo-demo/app/repositories"
	"prismgo-demo/app/services"
)

// AppServiceProvider registers application-wide services.
type AppServiceProvider struct{}

// Register binds application services before the application boots.
func (p AppServiceProvider) Register(app provider.Application) error {
	c := app.Container()
	bindings := []struct {
		key     string
		factory container.Factory
	}{
		{repositories.UserRepositoryKey, databaseRepositoryFactory(repositories.NewUserRepository)},
		{repositories.OrderRepositoryKey, databaseRepositoryFactory(repositories.NewOrderRepository)},
		{repositories.ReceiptRepositoryKey, databaseRepositoryFactory(repositories.NewReceiptRepository)},
		{services.OrderServiceKey, func(resolver container.Resolver) (any, error) {
			raw, err := resolver.Make(repositories.OrderRepositoryKey)
			if err != nil {
				return nil, err
			}
			orders, ok := raw.(*repositories.OrderRepository)
			if !ok {
				return nil, fmt.Errorf("demo provider: %s has type %T", repositories.OrderRepositoryKey, raw)
			}
			return services.NewOrderService(orders), nil
		}},
	}
	for _, binding := range bindings {
		if c.Bound(binding.key) {
			continue
		}
		if err := c.Singleton(binding.key, binding.factory); err != nil {
			return err
		}
	}
	return nil
}

// Boot runs after all providers have been registered.
func (p AppServiceProvider) Boot(app provider.Application) error {
	return nil
}

func databaseRepositoryFactory[T any](build func(*gorm.DB) T) container.Factory {
	return func(resolver container.Resolver) (any, error) {
		raw, err := resolver.Make("database.default")
		if err != nil {
			return nil, err
		}
		db, ok := raw.(*gorm.DB)
		if !ok {
			return nil, fmt.Errorf("demo provider: database.default has type %T", raw)
		}
		return build(db), nil
	}
}
