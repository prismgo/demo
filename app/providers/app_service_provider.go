package providers

import (
	providercontract "github.com/prismgo/framework/contracts/provider"
)

// AppServiceProvider registers application-wide services.
type AppServiceProvider struct{}

// Register binds application services before the application boots.
func (p AppServiceProvider) Register(app providercontract.Application) error {
	return nil
}

// Boot runs after all providers have been registered.
func (p AppServiceProvider) Boot(app providercontract.Application) error {
	return nil
}
