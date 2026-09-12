package consoledemo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func isKernelScenario(name string) bool {
	switch name {
	case "call", "call-silently", "call-input", "isolatable", "missing-input", "missing-input-custom":
		return true
	}
	return false
}

func helperProcess(mode, input string) (string, string, error) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return "", "", fmt.Errorf("locate console demo helper source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", ".."))
	path := filepath.Join(filepath.Dir(source), "testdata", "kernel_helper.go")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", path, mode)
	cmd.Dir = root
	cmd.Stdin = strings.NewReader(input)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	if ctx.Err() != nil {
		return out.String(), errOut.String(), fmt.Errorf("console demo helper %s: %w", mode, ctx.Err())
	}
	return out.String(), errOut.String(), err
}

func runKernelScenario(name string) (Result, error) {
	input := ""
	if name == "missing-input" {
		input = "Ada\n"
	}
	if name == "missing-input-custom" {
		input = "Ada\n\n"
	}
	out, errOut, err := helperProcess(name, input)
	if err != nil {
		return Result{}, fmt.Errorf("console demo %s helper: %w: %s", name, err, errOut)
	}
	marker := "SCENARIO_RESULT="
	start := strings.LastIndex(out, marker)
	if start < 0 {
		return Result{}, fmt.Errorf("console demo %s: helper result missing in %q", name, out)
	}
	data := strings.TrimSpace(out[start+len(marker):])
	var values []string
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return Result{}, fmt.Errorf("console demo %s decode helper result: %w", name, err)
	}
	return Result{Case: name, Values: values, Output: out[:start], ErrorOutput: errOut}, nil
}

func runPackageScenario(name string) (Result, error) {
	if name == "package-output" {
		out, errOut, err := helperProcess(name, "")
		if err != nil {
			return Result{}, fmt.Errorf("console demo package-output helper: %w: %s", err, errOut)
		}
		return Result{Case: name, Output: out, ErrorOutput: errOut}, nil
	}
	result := Result{Case: name}
	for _, mode := range []string{"package-exit", "package-exit-if", "package-exit-if-nil"} {
		out, errOut, err := helperProcess(mode, "")
		if mode == "package-exit-if-nil" {
			if err != nil || !strings.Contains(out, "continued") {
				return Result{}, fmt.Errorf("console demo %s: got output %q, error %v; want continued", mode, out, err)
			}
			result.Values = append(result.Values, "nil=continued")
			continue
		}
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return Result{}, fmt.Errorf("console demo %s: got stdout %q, stderr %q, error %v; want exit code 1", mode, out, errOut, err)
		}
		result.Values = append(result.Values, fmt.Sprintf("%s=1:%s", mode, strings.TrimSpace(out)))
	}
	return result, nil
}
