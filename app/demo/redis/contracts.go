package redisdemo

import (
	"context"
	"fmt"
	"strings"

	containercontract "github.com/prismgo/framework/contracts/container"
	rediscontract "github.com/prismgo/framework/contracts/redis"
	"github.com/prismgo/framework/redis"
	goredis "github.com/redis/go-redis/v9"
)

// facadeManagerScenario keeps the Redis facade entry points referenced at compile time.
func facadeManagerScenario() (string, error) {
	var (
		_ func() rediscontract.Factory                      = redis.Resolve
		_ func() *redis.Manager                             = redis.ManagerInstance
		_ func(...string) (rediscontract.Connection, error) = redis.Connection
		_ func(...string) (goredis.UniversalClient, error)  = redis.Client
	)
	return "facade=Resolve,ManagerInstance,Connection,Client", nil
}

// managerCloseOptionScenario verifies the facade close option installs a container closer.
func managerCloseOptionScenario() (string, error) {
	var binding containercontract.Binding
	redis.ManagerCloseOption()(&binding)
	if binding.Closer == nil {
		return "", fmt.Errorf("ManagerCloseOption did not install a container closer")
	}
	return "closer=true", nil
}

// managerContractScenario pins the rediscontract.Factory method set at compile time.
func managerContractScenario() (string, error) {
	var (
		_ rediscontract.Factory                                             = (*redis.Manager)(nil)
		_ func(*redis.Manager, ...string) (rediscontract.Connection, error) = (*redis.Manager).Connection
		_ func(*redis.Manager) (rediscontract.Connection, error)            = (*redis.Manager).DefaultConnection
		_ func(*redis.Manager) map[string]rediscontract.Connection          = (*redis.Manager).Connections
		_ func(*redis.Manager, ...string) error                             = (*redis.Manager).Purge
		_ func(*redis.Manager)                                              = (*redis.Manager).EnableEvents
		_ func(*redis.Manager)                                              = (*redis.Manager).DisableEvents
		_ func(*redis.Manager, context.Context) error                       = (*redis.Manager).Close
	)
	return "factory=Manager methods=7", nil
}

// connectionContractScenario pins the rediscontract.Connection method set at compile time.
func connectionContractScenario() (string, error) {
	var (
		_ rediscontract.Connection                                                           = (*redis.NamedConnection)(nil)
		_ func(*redis.NamedConnection) string                                                = (*redis.NamedConnection).Name
		_ func(*redis.NamedConnection) goredis.UniversalClient                               = (*redis.NamedConnection).Client
		_ func(*redis.NamedConnection, func(context.Context, rediscontract.CommandExecuted)) = (*redis.NamedConnection).Listen
		_ func(*redis.NamedConnection, func(context.Context, rediscontract.CommandFailed))   = (*redis.NamedConnection).ListenForFailures
	)
	return "connection=NamedConnection methods=4", nil
}

// eventAliasesScenario proves the redis event aliases resolve to the contract types.
func eventAliasesScenario() (string, error) {
	var (
		_ rediscontract.CommandExecuted      = redis.CommandExecuted{}
		_ rediscontract.CommandFailed        = redis.CommandFailed{}
		_ rediscontract.CommandSnapshot      = redis.CommandSnapshot{}
		_ rediscontract.CommandBatchExecuted = redis.CommandBatchExecuted{}
		_ rediscontract.CommandBatchFailed   = redis.CommandBatchFailed{}
	)
	return "aliases=CommandExecuted,CommandFailed,CommandSnapshot,CommandBatchExecuted,CommandBatchFailed", nil
}

// eventWrappersScenario reports the bus wrapper names for every Redis event type.
func eventWrappersScenario() (string, error) {
	wrappers := []struct {
		got  string
		want string
	}{
		{redis.CommandExecutedEvent{}.Name(), redis.EventCommandExecuted},
		{redis.CommandFailedEvent{}.Name(), redis.EventCommandFailed},
		{redis.CommandBatchExecutedEvent{}.Name(), redis.EventCommandBatchExecuted},
		{redis.CommandBatchFailedEvent{}.Name(), redis.EventCommandBatchFailed},
	}
	names := make([]string, 0, len(wrappers))
	for _, wrapper := range wrappers {
		if wrapper.got != wrapper.want {
			return "", fmt.Errorf("wrapper name = %q, want %q", wrapper.got, wrapper.want)
		}
		names = append(names, wrapper.got)
	}
	return "wrappers=4 names=" + strings.Join(names, ","), nil
}
