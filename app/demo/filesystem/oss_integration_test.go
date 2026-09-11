package filesystemdemo_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prismgo/framework/filesystem"

	"prismgo-demo/bootstrap"
)

func TestFilesystemDemoOSSDriver(t *testing.T) {
	required := []string{
		"FILESYSTEM_OSS_BUCKET",
		"FILESYSTEM_OSS_ENDPOINT",
		"FILESYSTEM_OSS_ACCESS_KEY_ID",
		"FILESYSTEM_OSS_ACCESS_KEY_SECRET",
	}
	for _, name := range required {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			t.Skipf("%s is not set; skipping real OSS integration test", name)
		}
	}

	prefix := fmt.Sprintf("prismgo-demo-tests/0004-%d", time.Now().UnixNano())
	t.Setenv("FILESYSTEM_OSS_PREFIX", prefix)
	t.Setenv("FILESYSTEM_OSS_VISIBILITY", filesystem.VisibilityPublic)

	app := bootstrap.NewApplication()
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close application: %v", err)
		}
	})
	if err := app.Boot(); err != nil {
		t.Fatalf("boot application with OSS extension: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	disk := filesystem.Disk("oss")
	key := "probe.txt"
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := disk.Delete(cleanupCtx, key); err != nil {
			t.Errorf("cleanup OSS object %s/%s: %v", prefix, key, err)
			return
		}
		exists, err := disk.Exists(cleanupCtx, key)
		if err != nil {
			t.Errorf("verify OSS cleanup %s/%s: %v", prefix, key, err)
			return
		}
		if exists {
			t.Errorf("OSS cleanup %s/%s exists = true, want false", prefix, key)
		}
	})

	const want = "hello from PrismGo OSS demo"
	if err := disk.Put(ctx, key, want, filesystem.PutOptions{ContentType: "text/plain"}); err != nil {
		t.Fatalf("put OSS object %s/%s: %v", prefix, key, err)
	}
	actual, err := disk.Get(ctx, key)
	if err != nil {
		t.Fatalf("get OSS object %s/%s: %v", prefix, key, err)
	}
	if string(actual) != want {
		t.Fatalf("OSS object %s/%s content = %q, want %q", prefix, key, actual, want)
	}
	publicURL, err := disk.URL(key)
	if err != nil || strings.TrimSpace(publicURL) == "" {
		t.Fatalf("OSS object %s/%s URL = %q, error = %v; want non-empty URL", prefix, key, publicURL, err)
	}
	temporaryURL, err := disk.TemporaryURL(ctx, key, time.Now().Add(5*time.Minute))
	if err != nil || strings.TrimSpace(temporaryURL) == "" {
		t.Fatalf("OSS object %s/%s temporary URL = %q, error = %v; want non-empty URL", prefix, key, temporaryURL, err)
	}
	if err := disk.Delete(ctx, key); err != nil {
		t.Fatalf("delete OSS object %s/%s: %v", prefix, key, err)
	}
	exists, err := disk.Exists(ctx, key)
	if err != nil {
		t.Fatalf("verify deleted OSS object %s/%s: %v", prefix, key, err)
	}
	if exists {
		t.Fatalf("deleted OSS object %s/%s exists = true, want false", prefix, key)
	}
}
