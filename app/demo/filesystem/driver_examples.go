package filesystemdemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/prismgo/framework/filesystem"
)

// sampleDriver is an in-memory custom driver used to exercise manager extension and lifecycle.
type sampleDriver struct {
	mu     sync.Mutex
	files  map[string][]byte
	closed bool
}

func (d *sampleDriver) Close() error { d.mu.Lock(); defer d.mu.Unlock(); d.closed = true; return nil }
func (d *sampleDriver) Write(_ context.Context, key string, reader io.Reader, _ filesystem.PutOptions) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read custom driver input: %w", err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return filesystem.ErrManagerClosed
	}
	if d.files == nil {
		d.files = make(map[string][]byte)
	}
	d.files[key] = data
	return nil
}
func (d *sampleDriver) ReadAll(_ context.Context, key string) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.files[key]
	if !ok {
		return nil, fmt.Errorf("read %q: %w", key, os.ErrNotExist)
	}
	return bytes.Clone(value), nil
}
func (d *sampleDriver) Open(ctx context.Context, key string) (io.ReadCloser, filesystem.FileInfo, error) {
	value, err := d.ReadAll(ctx, key)
	if err != nil {
		return nil, filesystem.FileInfo{}, err
	}
	return io.NopCloser(bytes.NewReader(value)), filesystem.FileInfo{Path: key, Size: int64(len(value))}, nil
}
func (d *sampleDriver) Exists(_ context.Context, key string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.files[key]
	return ok, nil
}
func (d *sampleDriver) Delete(_ context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.files, key)
	return nil
}
func (d *sampleDriver) Copy(ctx context.Context, src, dst string) error {
	value, err := d.ReadAll(ctx, src)
	if err != nil {
		return err
	}
	return d.Write(ctx, dst, bytes.NewReader(value), filesystem.PutOptions{})
}
func (d *sampleDriver) Move(ctx context.Context, src, dst string) error {
	if err := d.Copy(ctx, src, dst); err != nil {
		return err
	}
	return d.Delete(ctx, src)
}
func (d *sampleDriver) Stat(ctx context.Context, key string) (filesystem.FileInfo, error) {
	value, err := d.ReadAll(ctx, key)
	if err != nil {
		return filesystem.FileInfo{}, err
	}
	return filesystem.FileInfo{Path: key, Size: int64(len(value)), LastModified: time.Now(), ContentType: "text/plain"}, nil
}
func (d *sampleDriver) List(_ context.Context, prefix string, recursive bool) ([]filesystem.FileInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []filesystem.FileInfo
	for key, value := range d.files {
		if !strings.HasPrefix(key, strings.TrimSuffix(prefix, "/")+"/") {
			continue
		}
		relative := strings.TrimPrefix(key, strings.TrimSuffix(prefix, "/")+"/")
		if !recursive && strings.Contains(relative, "/") {
			continue
		}
		out = append(out, filesystem.FileInfo{Path: key, Size: int64(len(value))})
	}
	return out, nil
}
func (d *sampleDriver) MakeDirectory(context.Context, string) error { return nil }
func (d *sampleDriver) DeleteDirectory(_ context.Context, dir string) error {
	if strings.TrimSpace(dir) == "" {
		return filesystem.ErrEmptyDirectory
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for key := range d.files {
		if strings.HasPrefix(key, strings.TrimSuffix(dir, "/")+"/") {
			delete(d.files, key)
		}
	}
	return nil
}
func (d *sampleDriver) Path(key string) string     { return "memory://" + path.Clean(key) }
func (d *sampleDriver) URL(string) (string, error) { return "", filesystem.ErrPublicURLUnavailable }
func (d *sampleDriver) TemporaryURL(context.Context, string, time.Time) (string, error) {
	return "", filesystem.ErrTemporaryURLDisabled
}
func (d *sampleDriver) SetVisibility(_ context.Context, _ string, visibility string) error {
	if visibility != filesystem.VisibilityPrivate {
		return filesystem.ErrUnsupportedVisibility
	}
	return nil
}
func (d *sampleDriver) GetVisibility(context.Context, string) (string, error) {
	return filesystem.VisibilityPrivate, nil
}

func customDriverScenario(ctx context.Context, name string) (value string, runErr error) {
	cfg := filesystem.Config{Default: "custom", Disks: map[string]filesystem.DiskConfig{"custom": {Driver: "sample", Visibility: filesystem.VisibilityPrivate, Options: map[string]any{"tenant": "demo"}}}}
	manager, err := filesystem.NewManager(cfg)
	if err != nil {
		return "", fmt.Errorf("construct custom manager: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, manager.Close()) }()
	first := &sampleDriver{}
	replacement := &sampleDriver{}
	called := 0
	var factoryContext filesystem.DriverFactoryContext
	manager.Extend("sample", func(fc filesystem.DriverFactoryContext) (filesystem.Driver, error) {
		called++
		factoryContext = fc
		if name == "custom-driver-lifecycle" {
			return replacement, nil
		}
		return first, nil
	})
	disk := manager.Default()
	if err := disk.Put(ctx, "probe.txt", "custom driver"); err != nil {
		return "", err
	}
	data, err := disk.Get(ctx, "probe.txt")
	if err != nil {
		return "", err
	}
	if called != 1 {
		return "", fmt.Errorf("custom driver factories called = %d, want 1", called)
	}
	switch name {
	case "custom-driver":
		return string(data), nil
	case "custom-driver-lifecycle":
		manager.Extend("sample", func(filesystem.DriverFactoryContext) (filesystem.Driver, error) { return first, nil })
		if manager.Default() != disk {
			return "", errors.New("cached custom disk changed after factory replacement")
		}
		if err := manager.Close(); err != nil {
			return "", err
		}
		if !replacement.closed || first.closed {
			return "", errors.New("custom driver lifecycle closed wrong instance")
		}
		return "factory_called_once=true; cached=true; closed=true", nil
	case "optional-driver-capabilities":
		if disk.ProvidesTemporaryURLs() || disk.(*filesystem.Repository).ProvidesTemporaryUploadURLs() {
			return "", errors.New("minimal custom driver unexpectedly advertises URL capability")
		}
		if err := disk.Put(ctx, "nested/probe.txt", "nested content"); err != nil {
			return "", fmt.Errorf("write fallback directory probe: %w", err)
		}
		exists, err := disk.DirectoryExists(ctx, "nested")
		if err != nil {
			return "", fmt.Errorf("check fallback directory: %w", err)
		}
		if !exists {
			return "", errors.New("fallback directory listing did not find nested file")
		}
		return "temporary_url=false; temporary_upload=false; DirectoryExists uses List fallback", nil
	case "driver-factory-context":
		if factoryContext.Name != "custom" || factoryContext.Driver != "sample" || factoryContext.Config.Options["tenant"] != "demo" {
			return "", fmt.Errorf("factory context = %#v, want custom/sample/demo", factoryContext)
		}
		return "name=custom; driver=sample; tenant=demo", nil
	}
	return "", fmt.Errorf("unknown custom driver scenario %q", name)
}

func managerScenario(ctx context.Context, name string) (value string, runErr error) {
	if name == "manager-from-config" {
		closeManager, manager, err := filesystem.NewManagerFromConfig()
		if err != nil {
			return "", fmt.Errorf("manager from application config: %w", err)
		}
		defer func() { runErr = errors.Join(runErr, closeManager()) }()
		key := fmt.Sprintf("demo/filesystem/config-manager-%d.txt", time.Now().UnixNano())
		if err := manager.Default().Put(ctx, key, "configured manager"); err != nil {
			return "", fmt.Errorf("write through configured manager: %w", err)
		}
		defer func() { runErr = errors.Join(runErr, manager.Default().Delete(context.WithoutCancel(ctx), key)) }()
		content, err := manager.Default().Get(ctx, key)
		if err != nil {
			return "", fmt.Errorf("read through configured manager: %w", err)
		}
		if string(content) != "configured manager" {
			return "", fmt.Errorf("configured manager content = %q, want configured manager", content)
		}
		return fmt.Sprintf("default=%s; cloud=%s", manager.DefaultName(), manager.CloudName()), nil
	}
	root, err := os.MkdirTemp("", "prismgo-filesystem-manager-")
	if err != nil {
		return "", fmt.Errorf("create manual manager directory: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, os.RemoveAll(root)) }()
	manager, err := filesystem.NewManager(filesystem.Config{Default: "local", Disks: map[string]filesystem.DiskConfig{"local": {Driver: "local", Root: root, Visibility: filesystem.VisibilityPrivate, Serve: true}}, TemporaryURL: filesystem.TemporaryURLConfig{SigningKey: "demo-secret"}})
	if err != nil {
		return "", fmt.Errorf("construct manual manager: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, manager.Close()) }()
	if err := manager.Default().Put(ctx, "probe.txt", "manual manager"); err != nil {
		return "", err
	}
	data, err := manager.Default().Get(ctx, "probe.txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
