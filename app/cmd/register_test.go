package cmd

import "testing"

func TestCommandFactoriesIncludeDemoCommands(t *testing.T) {
	factories := CommandFactories()
	if len(factories) != 3 {
		t.Fatalf("CommandFactories() returned %d factories, want 3", len(factories))
	}
	wanted := map[string]bool{"demo:list": false, "demo:show": false, "demo:queue": false}
	for _, factory := range factories {
		command := factory()
		if command != nil {
			if _, ok := wanted[command.Definition().Name]; ok {
				wanted[command.Definition().Name] = true
			}
		}
	}
	for name, found := range wanted {
		if !found {
			t.Fatalf("command %q is not registered", name)
		}
	}
}
