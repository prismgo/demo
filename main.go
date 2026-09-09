// Package main is the only process entrypoint for the Prismgo application.
package main

import (
	"context"
	"os"
	"path/filepath"
	"prismgo-demo/bootstrap"

	"github.com/prismgo/framework/console"
)

func main() {
	app := bootstrap.NewApplication(localApplicationBasePath())

	if err := app.HandleCommand(context.Background(), os.Args); err != nil {
		console.Exit(err.Error())
	}
}

// localApplicationBasePath keeps the demo rooted in demo whether commands are
// run from the workspace root (`go run ./demo`) or from demo (`go run .`).
func localApplicationBasePath() string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return ""
	}

	demoDirectory := filepath.Join(workingDirectory, "demo")
	if _, err := os.Stat(filepath.Join(demoDirectory, "go.mod")); err == nil {
		return demoDirectory
	}

	return ""
}
