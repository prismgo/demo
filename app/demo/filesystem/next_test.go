package filesystemdemo_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworkcmd "github.com/prismgo/framework/cmd"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/support"
	"github.com/spf13/cobra"

	filesystemdemo "prismgo-demo/app/demo/filesystem"
	demotest "prismgo-demo/app/demo/testing"
)

func TestFilesystemDemoLastModified(t *testing.T) {
	assertFilesystemScenario(t, "last-modified", "modified_recent=true")
}
func TestFilesystemDemoFileInfo(t *testing.T) {
	assertFilesystemScenario(t, "file-info", "size=16; content_type=text/plain; is_dir=false")
}
func TestFilesystemDemoMimeType(t *testing.T) { assertFilesystemScenario(t, "mime-type", "text/plain") }
func TestFilesystemDemoChecksum(t *testing.T) { assertFilesystemScenario(t, "checksum", "sha256=") }
func TestFilesystemDemoPath(t *testing.T)     { assertFilesystemScenario(t, "path", "/example.txt") }
func TestFilesystemDemoMakeDirectory(t *testing.T) {
	assertFilesystemScenario(t, "make-directory", "nested_exists=true")
}
func TestFilesystemDemoFiles(t *testing.T)    { assertFilesystemScenario(t, "files", "count=1") }
func TestFilesystemDemoAllFiles(t *testing.T) { assertFilesystemScenario(t, "all-files", "count=3") }
func TestFilesystemDemoDirectories(t *testing.T) {
	assertFilesystemScenario(t, "directories", "count=1")
}
func TestFilesystemDemoAllDirectories(t *testing.T) {
	assertFilesystemScenario(t, "all-directories", "count=2")
}
func TestFilesystemDemoDeleteDirectory(t *testing.T) {
	assertFilesystemScenario(t, "delete-directory", "nested_exists_after_delete=false; root_protected=true")
}
func TestFilesystemDemoPublicURL(t *testing.T) {
	assertFilesystemScenario(t, "public-url", "/example.txt")
}
func TestFilesystemDemoTemporaryURL(t *testing.T) {
	assertFilesystemScenario(t, "temporary-url", "signed_url=true; expiry_verified=true")
}
func TestFilesystemDemoTemporaryURLCapability(t *testing.T) {
	assertFilesystemScenario(t, "temporary-url-capability", "local=true; public=true")
}
func TestFilesystemDemoTemporaryUploadCapability(t *testing.T) {
	assertFilesystemScenario(t, "temporary-upload-capability", "local=false")
}
func TestFilesystemDemoVerifyTemporaryURL(t *testing.T) {
	assertFilesystemScenario(t, "verify-temporary-url", "tampered_signature_rejected=true")
}
func TestFilesystemDemoLocalVisibility(t *testing.T) {
	assertFilesystemScenario(t, "local-visibility", "visibility=private; switch_rejected=true")
}
func TestFilesystemDemoCustomDriver(t *testing.T) {
	assertFilesystemScenario(t, "custom-driver", "custom driver")
}
func TestFilesystemDemoCustomDriverLifecycle(t *testing.T) {
	assertFilesystemScenario(t, "custom-driver-lifecycle", "closed=true")
}
func TestFilesystemDemoDriverContract(t *testing.T) {
	assertFilesystemScenario(t, "driver-contract", "custom driver implements")
}
func TestFilesystemDemoOptionalDriverCapabilities(t *testing.T) {
	assertFilesystemScenario(t, "optional-driver-capabilities", "temporary_url=false; temporary_upload=false")
}
func TestFilesystemDemoDriverFactoryContext(t *testing.T) {
	assertFilesystemScenario(t, "driver-factory-context", "name=custom; driver=sample; tenant=demo")
}
func TestFilesystemDemoManualManager(t *testing.T) {
	assertFilesystemScenario(t, "manual-manager", "manual manager")
}
func TestFilesystemDemoManagerFromConfig(t *testing.T) {
	assertFilesystemScenario(t, "manager-from-config", "default=local; cloud=oss")
}
func TestFilesystemDemoErrors(t *testing.T) {
	assertFilesystemScenario(t, "errors", "ErrDiskNotFound and ErrEmptyDirectory")
}
func TestFilesystemDemoLocalCapabilities(t *testing.T) {
	assertFilesystemScenario(t, "local-capabilities", "temporary_upload=false")
}
func TestFilesystemDemoLaravelCompatibility(t *testing.T) {
	assertFilesystemScenario(t, "laravel-compatibility", "Storage::disk -> filesystem.Disk")
}
func TestFilesystemDemoBestPractices(t *testing.T) {
	assertFilesystemScenario(t, "best-practices", "separate private and public disks")
}

func TestFilesystemDemoStorageLink(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	assertLinkResult(t, "storage-link", "storage:link uses filesystem.links")
	link := support.PublicPath("storage")
	runStorageCommand(t, frameworkcmd.NewStorageLinkCommand(), nil)
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("storage link %q info = %v, error = %v; want symlink", link, info, err)
	}
	runStorageCommand(t, frameworkcmd.NewStorageUnlinkCommand(), nil)
}

func TestFilesystemDemoStorageLinkOptions(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	assertLinkResult(t, "storage-link-options", "--relative")
	link := support.PublicPath("storage")
	runStorageCommand(t, frameworkcmd.NewStorageLinkCommand(), map[string]bool{"relative": true})
	target, err := os.Readlink(link)
	if err != nil || filepath.IsAbs(target) {
		t.Fatalf("relative storage link %q target = %q, error = %v; want relative target", link, target, err)
	}
	runStorageCommand(t, frameworkcmd.NewStorageLinkCommand(), map[string]bool{"force": true})
	target, err = os.Readlink(link)
	if err != nil || !filepath.IsAbs(target) {
		t.Fatalf("forced storage link %q target = %q, error = %v; want absolute target", link, target, err)
	}
	runStorageCommand(t, frameworkcmd.NewStorageUnlinkCommand(), nil)
	if err := os.WriteFile(link, []byte("keep"), 0o600); err != nil {
		t.Fatalf("create ordinary link path %q: %v", link, err)
	}
	if err := storageCommandError(frameworkcmd.NewStorageLinkCommand(), map[string]bool{"force": true}); err == nil {
		t.Fatalf("force storage link on regular file %q error = nil, want rejection", link)
	}
	data, err := os.ReadFile(link)
	if err != nil || string(data) != "keep" {
		t.Fatalf("regular file %q data = %q, error = %v; want keep", link, data, err)
	}
}

func TestFilesystemDemoStorageUnlink(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	assertLinkResult(t, "storage-unlink", "storage:unlink only removes")
	link := support.PublicPath("storage")
	runStorageCommand(t, frameworkcmd.NewStorageLinkCommand(), nil)
	runStorageCommand(t, frameworkcmd.NewStorageUnlinkCommand(), nil)
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("unlinked path %q error = %v, want os.ErrNotExist", link, err)
	}
	runStorageCommand(t, frameworkcmd.NewStorageUnlinkCommand(), nil)
}

func assertLinkResult(t *testing.T, name, want string) {
	t.Helper()
	result, err := filesystemdemo.Run(context.Background(), name)
	if err != nil || !strings.Contains(result.Value, want) {
		t.Fatalf("filesystem scenario %q result = %#v, error = %v; want value containing %q", name, result, err, want)
	}
}

func runStorageCommand(t *testing.T, command console.Command, flags map[string]bool) {
	t.Helper()
	if err := storageCommandError(command, flags); err != nil {
		t.Fatalf("storage command %q error = %v, want nil", command.Definition().Name, err)
	}
}

func storageCommandError(command console.Command, flags map[string]bool) error {
	definition := command.Definition()
	cobraCommand := &cobra.Command{Use: definition.Name}
	for _, name := range []string{"relative", "force"} {
		cobraCommand.Flags().Bool(name, false, "")
		if flags[name] {
			if err := cobraCommand.Flags().Set(name, "true"); err != nil {
				return err
			}
		}
	}
	input := console.NewInput(*definition, cobraCommand, nil)
	commandCtx := console.NewCommandContext(context.Background(), command, *definition, input, console.NewIO(strings.NewReader(""), io.Discard, io.Discard), nil, cobraCommand)
	return command.Handle(commandCtx)
}
