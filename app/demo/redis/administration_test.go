package redisdemo_test

import "testing"

func TestRedisDemoProviderRegistration(t *testing.T) {
	expectValue(t, "provider-registration", "name=redis bound=true lazy=true")
}

func TestRedisDemoProviderEventBridge(t *testing.T) {
	expectValue(t, "provider-event-bridge", "bridged=true event=redis.command_failed command=set")
}

func TestRedisDemoContainerFactory(t *testing.T) {
	expectValue(t, "container-factory", "key=redis type=*redis.Manager default=default")
}

func TestRedisDemoContainerConnection(t *testing.T) {
	expectValue(t, "container-connection", "key=redis.connection name=default client=true")
}

func TestRedisDemoContainerNamedConnection(t *testing.T) {
	expectValue(t, "container-named-connection", "factory=redis connection=cache distinct=true")
}

func TestRedisDemoLifecycleClose(t *testing.T) {
	expectValue(t, "lifecycle-close", "resolved=2 remaining=0 client-closed=true")
}

func TestRedisDemoHorizonConfiguration(t *testing.T) {
	expectValue(t, "horizon-config", "store=redis connection=cache prefix=demo_horizon encoding=json ttl=30s env=local")
}

func TestRedisDemoFacadeManagerAccess(t *testing.T) {
	expectValue(t, "facade-manager", "facade=Resolve,ManagerInstance,Connection,Client")
}

func TestRedisDemoManagerCloseOption(t *testing.T) {
	expectValue(t, "manager-close-option", "closer=true")
}

func TestRedisDemoManagerContract(t *testing.T) {
	expectValue(t, "manager-contract", "factory=Manager methods=7")
}

func TestRedisDemoConnectionContract(t *testing.T) {
	expectValue(t, "connection-contract", "connection=NamedConnection methods=4")
}

func TestRedisDemoEventAliases(t *testing.T) {
	expectValue(t, "event-aliases", "aliases=CommandExecuted,CommandFailed,CommandSnapshot,CommandBatchExecuted,CommandBatchFailed")
}

func TestRedisDemoEventWrappers(t *testing.T) {
	expectValue(t, "event-wrappers", "wrappers=4 names=redis.command_executed,redis.command_failed,redis.command_batch_executed,redis.command_batch_failed")
}
