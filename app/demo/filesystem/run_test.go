package filesystemdemo_test

import (
	"context"
	"strings"
	"testing"

	"github.com/prismgo/framework/filesystem"

	filesystemdemo "prismgo-demo/app/demo/filesystem"
	demotest "prismgo-demo/app/demo/testing"
)

func assertFilesystemScenario(t *testing.T, name, want string) {
	t.Helper()
	demotest.NewApplication(t, demotest.Options{})
	result, err := filesystemdemo.Run(context.Background(), name)
	if err != nil {
		t.Fatalf("run filesystem scenario %q error = %v, want nil", name, err)
	}
	if result.Case != name || !strings.Contains(result.Value, want) {
		t.Fatalf("filesystem scenario %q result = %#v, want case %q and value containing %q", name, result, name, want)
	}
	if name == "put-file" && (len(result.Details) != 1 || !strings.HasSuffix(result.Details[0], "/example.txt")) {
		t.Fatalf("filesystem scenario %q details = %v, want saved original filename", name, result.Details)
	}
	if name == "put-file-as" && (len(result.Details) != 1 || !strings.HasSuffix(result.Details[0], "/renamed.txt")) {
		t.Fatalf("filesystem scenario %q details = %v, want saved renamed filename", name, result.Details)
	}
	if name == "upload-fields" && (!strings.Contains(result.Value, "original_name=example.txt") || !strings.Contains(result.Value, "size=16")) {
		t.Fatalf("filesystem scenario %q value = %q, want original name and size", name, result.Value)
	}
	if result.Key != "" {
		disk := filesystem.Default()
		if name == "named-disk" || name == "public-url" {
			disk = filesystem.Disk("public")
		}
		missing, err := disk.Missing(context.Background(), result.Key)
		if err != nil || !missing {
			t.Fatalf("filesystem scenario %q cleaned key %q missing = %t, error = %v; want true, nil", name, result.Key, missing, err)
		}
	}
}

func TestFilesystemDemoArchitecture(t *testing.T) {
	assertFilesystemScenario(t, "architecture", "*filesystem.Manager -> local")
}
func TestFilesystemDemoConfiguration(t *testing.T) {
	assertFilesystemScenario(t, "config", "filesystem.default=local; disks=3")
}
func TestFilesystemDemoLocalPrerequisites(t *testing.T) {
	assertFilesystemScenario(t, "local-prerequisites", "local driver: writable root")
}
func TestFilesystemDemoTopLevelConfiguration(t *testing.T) {
	assertFilesystemScenario(t, "top-level-config", "default=local; cloud=oss")
}
func TestFilesystemDemoLocalConfiguration(t *testing.T) {
	assertFilesystemScenario(t, "local-config", "local: driver=local; root=")
}
func TestFilesystemDemoPublicConfiguration(t *testing.T) {
	assertFilesystemScenario(t, "public-config", "public: driver=local; root=")
}
func TestFilesystemDemoLinksConfiguration(t *testing.T) {
	assertFilesystemScenario(t, "links-config", "public/storage=storage/app/public")
}
func TestFilesystemDemoDiskConfiguration(t *testing.T) {
	assertFilesystemScenario(t, "disk-config", "driver=local; root=storage/app/private")
}
func TestFilesystemDemoFacade(t *testing.T) {
	assertFilesystemScenario(t, "facade", "hello filesystem")
}
func TestFilesystemDemoNamedDisk(t *testing.T) {
	assertFilesystemScenario(t, "named-disk", "hello filesystem")
}
func TestFilesystemDemoCloudDisk(t *testing.T) {
	assertFilesystemScenario(t, "cloud", "cloud=oss; driver=oss")
}
func TestFilesystemDemoInterfaceSelection(t *testing.T) {
	assertFilesystemScenario(t, "interface-selection", "repository=local; factory=*filesystem.Manager")
}
func TestFilesystemDemoGet(t *testing.T)  { assertFilesystemScenario(t, "get", "hello filesystem") }
func TestFilesystemDemoJSON(t *testing.T) { assertFilesystemScenario(t, "json", "PrismGo") }
func TestFilesystemDemoOpenStream(t *testing.T) {
	assertFilesystemScenario(t, "open-stream", "hello filesystem")
}
func TestFilesystemDemoReadStream(t *testing.T) {
	assertFilesystemScenario(t, "read-stream", "hello filesystem")
}
func TestFilesystemDemoDownload(t *testing.T) {
	assertFilesystemScenario(t, "download", "hello filesystem")
}
func TestFilesystemDemoFileExistence(t *testing.T) {
	assertFilesystemScenario(t, "file-existence", "missing_before=true; exists_after=true; file_exists=true")
}
func TestFilesystemDemoDirectoryExistence(t *testing.T) {
	assertFilesystemScenario(t, "directory-existence", "before=false; after=true")
}
func TestFilesystemDemoPut(t *testing.T) { assertFilesystemScenario(t, "put", "hello filesystem") }
func TestFilesystemDemoPutReader(t *testing.T) {
	assertFilesystemScenario(t, "put-reader", "stream alias")
}
func TestFilesystemDemoPrependAppend(t *testing.T) {
	assertFilesystemScenario(t, "prepend-append", "first\nmiddle\nlast")
}
func TestFilesystemDemoPutOptions(t *testing.T) {
	assertFilesystemScenario(t, "put-options", "content_type=text/plain; visibility=private")
}
func TestFilesystemDemoPutFile(t *testing.T) {
	assertFilesystemScenario(t, "put-file", "hello filesystem")
}
func TestFilesystemDemoPutFileAs(t *testing.T) {
	assertFilesystemScenario(t, "put-file-as", "hello filesystem")
}
func TestFilesystemDemoUploadFields(t *testing.T) {
	assertFilesystemScenario(t, "upload-fields", "disk=local; path=")
}
func TestFilesystemDemoCopyMove(t *testing.T) {
	assertFilesystemScenario(t, "copy-move", "hello filesystem")
}
func TestFilesystemDemoCrossDiskGuard(t *testing.T) {
	assertFilesystemScenario(t, "cross-disk-guard", "source_disk=true; other_disk=false")
}
func TestFilesystemDemoDelete(t *testing.T) {
	assertFilesystemScenario(t, "delete", "first_missing=true; second_missing=true")
}
func TestFilesystemDemoSize(t *testing.T) { assertFilesystemScenario(t, "size", "size=16") }

func TestFilesystemDemoUnknownScenario(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	_, err := filesystemdemo.Run(context.Background(), "unknown")
	if err == nil || !strings.Contains(err.Error(), `unknown filesystem scenario "unknown"`) {
		t.Fatalf("unknown scenario error = %v, want descriptive error", err)
	}
}
