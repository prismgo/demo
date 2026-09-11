package bootstrap

import (
	"github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/oss"
	"github.com/prismgo/rabbitmq"
	"github.com/prismgo/sqlite"
)

// Extensions returns optional framework extensions used by the application.
func Extensions() []provider.ServiceProvider {
	return []provider.ServiceProvider{
		sqlite.ServiceProvider{},
		rabbitmq.ServiceProvider{},
		oss.ServiceProvider{},
	}
}
