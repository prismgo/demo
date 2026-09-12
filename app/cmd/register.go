// Package cmd registers the demo application's console commands.
package cmd

import (
	"github.com/prismgo/framework/console"

	"prismgo-demo/app/cmd/demo"
)

// CommandFactories returns application console commands.
func CommandFactories() []console.CommandFactory {
	return []console.CommandFactory{
		func() console.Command { return demo.NewListCommand() },
		func() console.Command { return demo.NewShowCommand() },
		func() console.Command { return demo.NewQueueCommand() },
		func() console.Command { return demo.NewCacheCommand() },
		func() console.Command { return demo.NewTranslationCommand() },
		func() console.Command { return demo.NewConsoleCommand() },
		func() console.Command { return demo.NewCommandsCommand() },
	}
}
