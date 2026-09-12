package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runKeyCommand(ctx context.Context, executable, root, name string) (Result, error) {
	envPath := filepath.Join(root, ".env")
	initial := "APP_ENV=testing\nAPP_KEY=\n"
	if name == "key-force" || name == "key-show" || name == "key-format" {
		initial = "APP_ENV=testing\nAPP_KEY=existing-key\n"
	}
	if err := os.WriteFile(envPath, []byte(initial), 0o600); err != nil {
		return Result{}, fmt.Errorf("prepare %s environment: %w", name, err)
	}
	args := []string{"key:generate"}
	if name == "key-show" || name == "key-format" {
		args = append(args, "--show")
	}
	if name == "key-force" {
		if _, err := invoke(ctx, executable, root, nil, args...); err == nil {
			return Result{}, fmt.Errorf("%s without --force error = nil, want existing APP_KEY rejection", name)
		}
		args = append(args, "--force")
	}
	output, err := invoke(ctx, executable, root, nil, args...)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", name, err)
	}
	contents, err := os.ReadFile(envPath)
	if err != nil {
		return Result{}, fmt.Errorf("read %s environment: %w", name, err)
	}
	key := ""
	if name == "key-show" || name == "key-format" {
		if string(contents) != initial {
			return Result{}, fmt.Errorf("%s environment = %q, want unchanged %q", name, contents, initial)
		}
		key = strings.TrimSpace(output)
	} else {
		for _, line := range strings.Split(string(contents), "\n") {
			if strings.HasPrefix(line, "APP_KEY=") {
				key = strings.TrimPrefix(line, "APP_KEY=")
			}
		}
		if key == "" || (name == "key-force" && key == "existing-key") {
			return Result{}, fmt.Errorf("%s APP_KEY = %q, want a new key", name, key)
		}
	}
	encoded, ok := strings.CutPrefix(key, "base64:")
	if !ok {
		return Result{}, fmt.Errorf("%s key = %q, want base64 prefix", name, key)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) != 32 {
		return Result{}, fmt.Errorf("%s decoded key length = %d, error = %v; want 32 bytes", name, len(raw), err)
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: fmt.Sprintf("verified 32-byte application key; %s", strings.TrimSpace(output))}, nil
}
