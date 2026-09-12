package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func runCronCommand(ctx context.Context, executable, root, name string) (Result, error) {
	path := filepath.Join(root, "cron-output.log")
	outputFile, err := os.Create(path)
	if err != nil {
		return Result{}, fmt.Errorf("create cron output: %w", err)
	}
	cmd := exec.CommandContext(ctx, executable, "cron")
	cmd.Dir = root
	cmd.Env = commandEnvironment(nil)
	cmd.Stdout = outputFile
	cmd.Stderr = outputFile
	if err := cmd.Start(); err != nil {
		_ = outputFile.Close()
		return Result{}, fmt.Errorf("start cron: %w", err)
	}
	waited := false
	defer func() {
		if !waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
		_ = outputFile.Close()
	}()
	var output []byte
	for {
		output, err = os.ReadFile(path)
		if err != nil {
			return Result{}, fmt.Errorf("read cron output: %w", err)
		}
		if strings.Contains(string(output), "cron scheduler started") {
			break
		}
		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf("wait for cron start: %w; output = %q", err, output)
		}
		time.Sleep(20 * time.Millisecond)
	}
	signal := syscall.SIGINT
	if name == "cron-shutdown" {
		signal = syscall.SIGTERM
	}
	if err := cmd.Process.Signal(signal); err != nil {
		return Result{}, fmt.Errorf("signal cron: %w", err)
	}
	waitErr := cmd.Wait()
	waited = true
	if waitErr != nil {
		return Result{}, fmt.Errorf("wait for cron shutdown: %w", waitErr)
	}
	output, err = os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("read completed cron output: %w", err)
	}
	if !strings.Contains(string(output), "app:daily-maintenance") || !strings.Contains(string(output), "cron scheduler stopped") {
		return Result{}, fmt.Errorf("%s output = %q, want registered task and graceful stop", name, output)
	}
	return Result{Case: name, Command: "cron", Output: strings.TrimSpace(string(output))}, nil
}
