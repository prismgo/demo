package bootstrap

import (
	appproviders "prismgo-demo/app/providers"

	"github.com/prismgo/framework/provider"
)

// Providers returns application-level service providers.
func Providers() []provider.ServiceProvider {
	return []provider.ServiceProvider{
		appproviders.AppServiceProvider{},
	}
}
