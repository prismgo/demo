package cmd

import (
	"prismgo-demo/app/cmd/demo"

	"github.com/prismgo/framework/console"
)

// CommandFactories returns application console commands.
func CommandFactories() []console.CommandFactory {
	return []console.CommandFactory{
		func() console.Command { return demo.NewListCommand() },
	}
}
