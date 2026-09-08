// Package main is the only process entrypoint for the Prismgo application.
package main

import (
	"context"
	"os"
	"prismgo-demo/bootstrap"

	"github.com/prismgo/framework/console"
)

func main() {
	app := bootstrap.NewApplication()

	if err := app.HandleCommand(context.Background(), os.Args); err != nil {
		console.Exit(err.Error())
	}
}
