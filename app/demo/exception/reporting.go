package exceptiondemo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/exception"
	"github.com/prismgo/framework/support"
)

func reporterOrder() (string, error) {
	marker := fmt.Sprintf("exception demo reporter order %d", time.Now().UnixNano())
	logged := false
	var readErr error
	h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(any, error, map[string]any) {
		var content string
		content, readErr = readErrorLog()
		logged = readErr == nil && strings.Contains(content, marker)
	}))
	h.Report(context.Background(), errors.New(marker), map[string]any{"status": 500})
	if readErr != nil {
		return "", readErr
	}
	if !logged {
		return "", fmt.Errorf("reporter ran before error log contained %q", marker)
	}
	return "logged_before_reporter=true", nil
}

func defaultReport() (string, error) {
	marker := fmt.Sprintf("exception demo default report %d", time.Now().UnixNano())
	exception.New(exception.WithPanicStack(false)).Report(context.Background(), errors.New(marker), map[string]any{"status": 500})
	content, err := readErrorLog()
	if err != nil {
		return "", err
	}
	if !strings.Contains(content, marker) {
		return "", fmt.Errorf("error log does not contain %q", marker)
	}
	return "error_channel=true", nil
}

func packageReport() (string, error) {
	marker := fmt.Sprintf("exception demo package facade %d", time.Now().UnixNano())
	exception.Report(context.Background(), errors.New(marker), map[string]any{"status": 500, "caller": "demo:exception"})
	content, err := readErrorLog()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, marker) {
			if !strings.Contains(line, "demo:exception") {
				return "", fmt.Errorf("error log entry for %q lacks caller field", marker)
			}
			return "logged=true; caller=true", nil
		}
	}
	return "", fmt.Errorf("error log does not contain %q", marker)
}

func logContext() (string, error) {
	var fields map[string]any
	h := exception.New(exception.WithPanicStack(false), exception.WithContext(func(*gin.Context) map[string]any {
		return map[string]any{"tenant_id": "tenant-42"}
	}), exception.WithReporter(func(_ any, _ error, reported map[string]any) {
		fields = reported
	}))
	response, err := serve(h, requestOptions{
		useMiddleware: true,
		register: func(engine *gin.Engine) {
			engine.GET("/failure", func(c *gin.Context) { _ = c.Error(errDemo) })
		},
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("log_tenant=%v; response_tenant=%t", fields["tenant_id"], strings.Contains(response.body, "tenant_id")), nil
}

func readErrorLog() (string, error) {
	path := support.BasePath(config.Get("logging.channels.error.path"))
	if config.Get("logging.channels.error.driver") == "daily" {
		extension := filepath.Ext(path)
		base := strings.TrimSuffix(filepath.Base(path), extension)
		path = filepath.Join(filepath.Dir(path), base+"-"+time.Now().Format("2006-01-02")+extension)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read exception error log %s: %w", path, err)
	}
	return string(content), nil
}
