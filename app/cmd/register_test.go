package cmd

import "testing"

func TestCommandFactoriesIncludeDemoList(t *testing.T) {
	factories := CommandFactories()
	if len(factories) != 1 {
		t.Fatalf("CommandFactories() returned %d factories, want 1", len(factories))
	}
	command := factories[0]()
	if command == nil || command.Definition().Name != "demo:list" {
		t.Fatalf("registered command = %#v", command)
	}
}
