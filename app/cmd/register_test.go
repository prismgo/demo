package cmd

import "testing"

func TestCommandFactoriesIncludeDemoCommands(t *testing.T) {
	factories := CommandFactories()
	if len(factories) != 22 {
		t.Fatalf("CommandFactories() returned %d factories, want 22", len(factories))
	}
	wanted := map[string]bool{"demo:list": false, "demo:show": false, "demo:queue": false, "demo:cache": false, "demo:filesystem": false, "demo:event": false, "demo:config": false, "demo:container": false, "demo:cookie": false, "demo:exception": false, "demo:logger": false, "demo:translation": false, "demo:console": false, "demo:commands": false, "demo:http-server": false, "demo:session": false, "demo:redis": false, "demo:lifecycle": false, "demo:route": false, "demo:provider": false, "demo:ratelimit": false, "demo:timer": false}
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
