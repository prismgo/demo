// Package httpserver contains runnable HTTP server configuration and lifecycle examples.
package httpserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/foundation"
	prismhttp "github.com/prismgo/framework/http"

	"prismgo-demo/app/demo/commands"
)

// Result records the server behavior observed by a scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// ginOutputMu serializes the temporary Gin output redirection used by CLI examples.
var ginOutputMu sync.Mutex

// Run executes one HTTP server example against the active application or an isolated Demo process.
func Run(ctx context.Context, executable, name string) (Result, error) {
	ginOutputMu.Lock()
	defer ginOutputMu.Unlock()
	if name != "access-log" {
		previous := gin.DefaultWriter
		gin.DefaultWriter = io.Discard
		defer func() { gin.DefaultWriter = previous }()
	}
	value, err := run(ctx, executable, name)
	if err != nil {
		return Result{}, fmt.Errorf("http-server demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(ctx context.Context, executable, name string) (string, error) {
	cfg := prismhttp.CurrentServerConfig()
	switch name {
	case "bootstrap", "routes":
		if foundation.App == nil {
			return "", errors.New("application is not initialized")
		}
		server, err := foundation.App.NewHTTPServer(ctx, "0")
		if err != nil {
			return "", fmt.Errorf("construct application HTTP server: %w", err)
		}
		if name == "bootstrap" {
			return fmt.Sprintf("application constructed %T at %s", server, server.Addr), nil
		}
		return requestHandler(server.Handler, "/api/health", nil, "")
	case "serve", "port-flag", "stop", "kill", "reload", "restart":
		if executable == "" {
			return "", errors.New("Demo executable is required")
		}
		commandCase := map[string]string{
			"serve": "serve", "port-flag": "serve-port", "stop": "serve-stop",
			"kill": "serve-kill", "reload": "serve-reload", "restart": "serve-restart",
		}[name]
		result, err := commands.Run(ctx, executable, commandCase)
		if err != nil {
			return "", err
		}
		return result.Command + ": " + result.Output, nil
	case "port-flag-scope":
		return portFlagScope(ctx, executable)
	case "host":
		return cfg.Addr(), nil
	case "port":
		return cfg.Port, nil
	case "timeout-fallback":
		return timeoutFallback()
	case "read-timeout":
		return serverField(cfg, func(s *http.Server) any { return s.ReadTimeout })
	case "read-header-timeout":
		return serverField(cfg, func(s *http.Server) any { return s.ReadHeaderTimeout })
	case "write-timeout":
		return serverField(cfg, func(s *http.Server) any { return s.WriteTimeout })
	case "idle-timeout":
		return serverField(cfg, func(s *http.Server) any { return s.IdleTimeout })
	case "shutdown-timeout":
		return gracefulShutdown(ctx, cfg)
	case "max-header-bytes":
		return headerLimit(ctx, cfg)
	case "max-multipart-memory":
		return multipartLimit(cfg)
	case "duration-string":
		return fmt.Sprintf("read=%s header=%s write=%s idle=%s", cfg.ReadTimeout, cfg.ReadHeaderTimeout, cfg.WriteTimeout, cfg.IdleTimeout), nil
	case "duration-seconds":
		return cfg.ReadTimeout.String(), nil
	case "trusted-proxies", "client-ip-headers", "untrusted-client-ip":
		return clientIP(cfg, name)
	case "access-log":
		return accessLog(cfg)
	case "exception-handler", "debug-exception":
		return exceptionResponse(cfg)
	case "pid-file":
		return pidFile(ctx, executable)
	case "missing-pid":
		return missingPID(ctx, executable)
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

func timeoutFallback() (string, error) {
	file, err := os.CreateTemp("", "prismgo-http-timeout-*.env")
	if err != nil {
		return "", fmt.Errorf("create fallback environment: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString("SERVER_TIMEOUT=7\nSERVER_READ_TIMEOUT=invalid\nSERVER_WRITE_TIMEOUT=invalid\nSERVER_SHUTDOWN_TIMEOUT=invalid\n"); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("write fallback environment: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close fallback environment: %w", err)
	}
	if err := config.Resolve().ReloadFromFile(file.Name()); err != nil {
		return "", fmt.Errorf("load fallback environment: %w", err)
	}
	cfg := prismhttp.CurrentServerConfig()
	if err := config.Resolve().Reload(); err != nil {
		return "", fmt.Errorf("restore application environment: %w", err)
	}
	return fmt.Sprintf("read=%s write=%s shutdown=%s", cfg.ReadTimeout, cfg.WriteTimeout, cfg.ShutdownTimeout), nil
}

func serverField(cfg prismhttp.ServerConfig, field func(*http.Server) any) (string, error) {
	server, err := newServer(cfg, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(field(server)), nil
}

func newServer(cfg prismhttp.ServerConfig, route func(*gin.Engine)) (*http.Server, error) {
	return prismhttp.NewApplicationServer("", func(engine *gin.Engine, useInternal func(*gin.Engine)) error {
		useInternal(engine)
		if route != nil {
			route(engine)
		}
		return nil
	}, prismhttp.WithServerConfig(cfg))
}

func requestHandler(handler http.Handler, path string, headers http.Header, remote string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, "http://example.test"+path, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	request.Header = headers
	if remote != "" {
		request.RemoteAddr = remote
	}
	response := newResponseRecorder()
	handler.ServeHTTP(response, request)
	return fmt.Sprintf("status=%d body=%s", response.status, strings.TrimSpace(response.body.String())), nil
}

type responseRecorder struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{header: make(http.Header), status: http.StatusOK}
}
func (r *responseRecorder) Header() http.Header            { return r.header }
func (r *responseRecorder) WriteHeader(status int)         { r.status = status }
func (r *responseRecorder) Write(data []byte) (int, error) { return r.body.Write(data) }

func clientIP(cfg prismhttp.ServerConfig, name string) (string, error) {
	server, err := newServer(cfg, func(engine *gin.Engine) {
		engine.GET("/client-ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })
	})
	if err != nil {
		return "", err
	}
	headers := make(http.Header)
	headers.Set("X-Forwarded-For", "203.0.113.10")
	headers.Set("X-Real-IP", "198.51.100.9")
	remote := "127.0.0.1:12345"
	if name == "untrusted-client-ip" {
		remote = "192.0.2.7:12345"
	}
	return requestHandler(server.Handler, "/client-ip", headers, remote)
}

func multipartLimit(cfg prismhttp.ServerConfig) (string, error) {
	server, err := newServer(cfg, func(engine *gin.Engine) {
		engine.POST("/upload", func(c *gin.Context) {
			if err := c.Request.ParseMultipartForm(engine.MaxMultipartMemory); err != nil {
				c.String(http.StatusBadRequest, "%s", err)
				return
			}
			defer c.Request.MultipartForm.RemoveAll()
			files := c.Request.MultipartForm.File["payload"]
			if len(files) != 1 {
				c.String(http.StatusBadRequest, "file count=%d", len(files))
				return
			}
			opened, err := files[0].Open()
			if err != nil {
				c.String(http.StatusBadRequest, "open upload: %s", err)
				return
			}
			defer opened.Close()
			_, onDisk := opened.(*os.File)
			c.String(http.StatusOK, "memory=%d file-on-disk=%t", engine.MaxMultipartMemory, onDisk)
		})
	})
	if err != nil {
		return "", err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("payload", "example.txt")
	if err != nil {
		return "", fmt.Errorf("create upload part: %w", err)
	}
	if _, err := io.Copy(part, strings.NewReader(strings.Repeat("x", int(cfg.MaxMultipartMemory)+4096))); err != nil {
		return "", fmt.Errorf("write upload part: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close upload form: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, "http://example.test/upload", &body)
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := newResponseRecorder()
	server.Handler.ServeHTTP(response, request)
	return fmt.Sprintf("multipart %s", strings.TrimSpace(response.body.String())), nil
}

func accessLog(cfg prismhttp.ServerConfig) (string, error) {
	var output bytes.Buffer
	previous := gin.DefaultWriter
	gin.DefaultWriter = &output
	defer func() { gin.DefaultWriter = previous }()
	server, err := newServer(cfg, func(engine *gin.Engine) {
		engine.GET("/access", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	})
	if err != nil {
		return "", err
	}
	observed, err := requestHandler(server.Handler, "/access", nil, "127.0.0.1:12345")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("enabled=%t logged=%t %s", cfg.AccessLog, strings.Contains(output.String(), "/access"), observed), nil
}

func exceptionResponse(cfg prismhttp.ServerConfig) (string, error) {
	server, err := newServer(cfg, func(engine *gin.Engine) {
		engine.GET("/error", func(c *gin.Context) { _ = c.Error(errors.New("internal demo detail")) })
	})
	if err != nil {
		return "", err
	}
	return requestHandler(server.Handler, "/error", nil, "127.0.0.1:12345")
}

func gracefulShutdown(ctx context.Context, cfg prismhttp.ServerConfig) (string, error) {
	server, err := newServer(cfg, func(engine *gin.Engine) {
		engine.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ready") })
	})
	if err != nil {
		return "", err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("listen locally: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	defer func() { _ = server.Close(); <-done }()
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return "", fmt.Errorf("shutdown server: %w", err)
	}
	return fmt.Sprintf("graceful shutdown within %s", cfg.ShutdownTimeout), nil
}

func headerLimit(ctx context.Context, cfg prismhttp.ServerConfig) (string, error) {
	server, err := newServer(cfg, func(engine *gin.Engine) {
		engine.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ready") })
	})
	if err != nil {
		return "", err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("listen locally: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	defer func() { _ = server.Close(); <-done }()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+listener.Addr().String()+"/", nil)
	if err != nil {
		return "", fmt.Errorf("create header request: %w", err)
	}
	request.Header.Set("X-Demo-Padding", strings.Repeat("x", server.MaxHeaderBytes+8192))
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(request)
	if err != nil {
		return "", fmt.Errorf("request local server: %w", err)
	}
	defer response.Body.Close()
	return fmt.Sprintf("max=%d oversized-header-status=%d", server.MaxHeaderBytes, response.StatusCode), nil
}

func isolatedRoot() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-http-server-*")
	if err != nil {
		return "", fmt.Errorf("create isolated directory: %w", err)
	}
	if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
		_ = os.RemoveAll(root)
		return "", fmt.Errorf("prepare isolated application: %w", err)
	}
	return root, nil
}

func invoke(ctx context.Context, executable, root string, environment []string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = root
	command.Env = append(os.Environ(), append([]string{
		"APP_ENV=testing", "APP_KEY=base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION=sqlite", "CACHE_STORE=memory", "QUEUE_CONNECTION=sync", "SESSION_DRIVER=file",
	}, environment...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("run %s: %w: %s", strings.Join(args, " "), err, bytes.TrimSpace(output))
	}
	return string(output), nil
}

func portFlagScope(ctx context.Context, executable string) (string, error) {
	if executable == "" {
		return "", errors.New("Demo executable is required")
	}
	root, err := isolatedRoot()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	path := filepath.Join(root, ".env")
	content := []byte("SERVER_PORT=8123\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return "", fmt.Errorf("write isolated environment: %w", err)
	}
	port, err := unusedPort()
	if err != nil {
		return "", err
	}
	_, err = invoke(ctx, executable, root, nil, "serve", "--port="+strconv.Itoa(port), "--stop")
	if err == nil || !strings.Contains(err.Error(), "read pid file failed") {
		return "", fmt.Errorf("port-specific control error = %v, want missing PID", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read isolated environment: %w", err)
	}
	if !bytes.Equal(content, after) {
		return "", fmt.Errorf("environment after --port = %q, want %q", after, content)
	}
	return "--port used transient port; .env remains SERVER_PORT=8123", nil
}

func unusedPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("allocate local port: %w", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func pidFile(ctx context.Context, executable string) (string, error) {
	if executable == "" {
		return "", errors.New("Demo executable is required")
	}
	root, err := isolatedRoot()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	port, err := unusedPort()
	if err != nil {
		return "", err
	}
	portText := strconv.Itoa(port)
	command := exec.CommandContext(ctx, executable, "serve", "--port="+portText)
	command.Dir = root
	command.Env = append(os.Environ(), "APP_ENV=testing", "APP_KEY=base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=", "DB_CONNECTION=sqlite", "CACHE_STORE=memory", "QUEUE_CONNECTION=sync", "SESSION_DRIVER=file", "SERVER_HOST=127.0.0.1", "SERVER_ACCESS_LOG=false")
	logPath := filepath.Join(root, "serve.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return "", fmt.Errorf("create serve log: %w", err)
	}
	defer logFile.Close()
	command.Stdout, command.Stderr = logFile, logFile
	if err := command.Start(); err != nil {
		return "", fmt.Errorf("start serve process: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	defer func() { _ = command.Process.Kill(); <-done }()
	path := filepath.Join(os.TempDir(), "prismgo-serve-"+portText+".pid")
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if parseErr != nil || pid != command.Process.Pid {
				return "", fmt.Errorf("PID file = %q, want process %d", data, command.Process.Pid)
			}
			stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if _, stopErr := invoke(stopCtx, executable, root, nil, "serve", "--port="+portText, "--stop"); stopErr != nil {
				return "", fmt.Errorf("stop PID-owned server: %w", stopErr)
			}
			return fmt.Sprintf("port=%s pid=%d", portText, pid), nil
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline.C:
			output, readErr := os.ReadFile(logPath)
			if readErr != nil {
				return "", fmt.Errorf("PID file %s not created; read serve log: %w", path, readErr)
			}
			return "", fmt.Errorf("PID file %s not created; serve output = %q", path, output)
		case <-ticker.C:
		}
	}
}

func missingPID(ctx context.Context, executable string) (string, error) {
	if executable == "" {
		return "", errors.New("Demo executable is required")
	}
	root, err := isolatedRoot()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	port, err := unusedPort()
	if err != nil {
		return "", err
	}
	output, err := invoke(ctx, executable, root, nil, "serve", "--port="+strconv.Itoa(port), "--stop")
	if err == nil || !strings.Contains(output, "read pid file failed") {
		return "", fmt.Errorf("missing PID output = %q, error = %v, want read pid file failed", output, err)
	}
	return "missing PID reported: " + strings.TrimSpace(output), nil
}
