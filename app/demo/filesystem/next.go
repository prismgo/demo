package filesystemdemo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/prismgo/framework/config"
	fscontract "github.com/prismgo/framework/contracts/filesystem"
	"github.com/prismgo/framework/filesystem"
)

func isNextCase(name string) bool {
	switch name {
	case "oss-prerequisites", "oss-config", "last-modified", "file-info", "mime-type":
		return true
	case "checksum", "path", "make-directory", "files", "all-files":
		return true
	case "directories", "all-directories", "delete-directory", "public-url", "storage-link":
		return true
	case "storage-link-options", "storage-unlink", "temporary-url", "temporary-url-capability", "temporary-upload-url":
		return true
	case "temporary-upload-capability", "verify-temporary-url", "local-visibility", "oss-visibility", "custom-driver":
		return true
	case "custom-driver-lifecycle", "driver-contract", "optional-driver-capabilities", "driver-factory-context", "manual-manager":
		return true
	case "manager-from-config", "errors", "local-capabilities", "oss-capabilities", "laravel-compatibility":
		return true
	case "best-practices":
		return true
	}
	return false
}

func runNext(ctx context.Context, name string) (result Result, runErr error) {
	if !isNextCase(name) {
		return Result{}, fmt.Errorf("unknown filesystem scenario %q", name)
	}
	result.Case = name
	if isNextConfiguration(name) {
		result.Value, runErr = nextConfiguration(ctx, name)
		return result, runErr
	}
	disk := filesystem.Default()
	if name == "public-url" {
		disk = filesystem.Disk("public")
	}
	if name == "oss-visibility" || name == "temporary-upload-url" {
		disk = filesystem.Disk("oss")
	}
	dir := fmt.Sprintf("demo/filesystem/%s-%d", name, time.Now().UnixNano())
	result.Key = dir + "/example.txt"
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		if err := disk.DeleteDirectory(cleanupCtx, dir); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("clean up filesystem demo %q: %w", dir, err))
		}
	}()
	result.Value, result.Details, runErr = nextOperation(ctx, name, disk, result.Key, dir)
	if runErr != nil {
		runErr = fmt.Errorf("filesystem demo %s: %w", name, runErr)
	}
	return result, runErr
}

func isNextConfiguration(name string) bool {
	switch name {
	case "oss-prerequisites", "oss-config", "storage-link", "storage-link-options", "storage-unlink", "custom-driver", "custom-driver-lifecycle", "driver-contract", "optional-driver-capabilities", "driver-factory-context", "manual-manager", "manager-from-config", "errors", "local-capabilities", "oss-capabilities", "laravel-compatibility", "best-practices", "temporary-upload-capability", "temporary-url-capability":
		return true
	}
	return false
}

func nextConfiguration(ctx context.Context, name string) (string, error) {
	cfg := config.Resolve()
	switch name {
	case "oss-prerequisites":
		required := []string{"FILESYSTEM_OSS_BUCKET", "FILESYSTEM_OSS_ENDPOINT", "FILESYSTEM_OSS_ACCESS_KEY_ID", "FILESYSTEM_OSS_ACCESS_KEY_SECRET"}
		missing := make([]string, 0)
		for _, key := range required {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				missing = append(missing, key)
			}
		}
		if len(missing) != 0 {
			return "", fmt.Errorf("OSS prerequisites missing: %s", strings.Join(missing, ", "))
		}
		return "OSS credentials configured", nil
	case "oss-config":
		spec := cfg.GetStringMap("filesystem.disks.oss")
		if spec["driver"] != "oss" {
			return "", fmt.Errorf("oss driver = %v, want oss", spec["driver"])
		}
		return fmt.Sprintf("driver=oss; bucket_configured=%t; endpoint_configured=%t", spec["bucket"] != "", spec["endpoint"] != ""), nil
	case "storage-link", "storage-link-options", "storage-unlink":
		links := cfg.GetStringMap("filesystem.links")
		if len(links) == 0 {
			return "", errors.New("filesystem links are not configured")
		}
		switch name {
		case "storage-link":
			return "storage:link uses filesystem.links to expose public storage", nil
		case "storage-link-options":
			return "storage:link --relative; storage:link --force only replaces symlinks", nil
		default:
			return "storage:unlink only removes configured symlinks", nil
		}
	case "temporary-url-capability":
		return fmt.Sprintf("local=%t; public=%t", filesystem.Disk("local").ProvidesTemporaryURLs(), filesystem.Disk("public").ProvidesTemporaryURLs()), nil
	case "temporary-upload-capability":
		return fmt.Sprintf("local=%t", filesystem.Disk("local").(*filesystem.Repository).ProvidesTemporaryUploadURLs()), nil
	case "driver-contract":
		var driver filesystem.Driver = (*sampleDriver)(nil)
		return fmt.Sprintf("custom driver implements %T", driver), nil
	case "custom-driver", "custom-driver-lifecycle", "optional-driver-capabilities", "driver-factory-context":
		return customDriverScenario(ctx, name)
	case "manual-manager", "manager-from-config":
		return managerScenario(ctx, name)
	case "errors":
		err := filesystem.Disk("absent-demo-disk").Put(ctx, "probe.txt", "x")
		if !errors.Is(err, filesystem.ErrDiskNotFound) {
			return "", fmt.Errorf("missing disk error = %v, want ErrDiskNotFound", err)
		}
		if err := filesystem.Default().DeleteDirectory(ctx, ""); !errors.Is(err, filesystem.ErrEmptyDirectory) {
			return "", fmt.Errorf("empty directory error = %v, want ErrEmptyDirectory", err)
		}
		return "ErrDiskNotFound and ErrEmptyDirectory identified with errors.Is", nil
	case "local-capabilities":
		local := filesystem.Disk("local")
		if !local.ProvidesTemporaryURLs() || local.(*filesystem.Repository).ProvidesTemporaryUploadURLs() {
			return "", errors.New("unexpected local URL capabilities")
		}
		return "read/write/list/checksum; temporary_url=true; temporary_upload=false", nil
	case "oss-capabilities":
		disk := filesystem.Disk("oss")
		if !disk.ProvidesTemporaryURLs() || !disk.(*filesystem.Repository).ProvidesTemporaryUploadURLs() {
			return "", errors.New("OSS temporary URL capabilities unavailable")
		}
		return "temporary_url=true; temporary_upload=true; per-file visibility", nil
	case "laravel-compatibility":
		var repository fscontract.Repository = filesystem.Default()
		if repository.Name() == "" {
			return "", errors.New("default filesystem disk has no name")
		}
		return "Storage::disk -> filesystem.Disk; Storage::get -> filesystem.Get; Storage::put -> filesystem.Put", nil
	case "best-practices":
		if filesystem.Default().Name() == filesystem.Disk("public").Name() {
			return "", errors.New("private and public disks must differ")
		}
		return "use disk-relative keys; separate private and public disks; check returned errors", nil
	}
	return "", fmt.Errorf("unhandled filesystem configuration scenario %q", name)
}

func nextOperation(ctx context.Context, name string, disk fscontract.Repository, key, dir string) (string, []string, error) {
	const content = "hello filesystem"
	switch name {
	case "make-directory":
		sub := dir + "/nested"
		if err := disk.MakeDirectory(ctx, sub); err != nil {
			return "", nil, err
		}
		exists, err := disk.DirectoryExists(ctx, sub)
		return fmt.Sprintf("nested_exists=%t", exists), nil, err
	case "delete-directory":
		sub := dir + "/nested"
		if err := disk.MakeDirectory(ctx, sub); err != nil {
			return "", nil, err
		}
		if err := disk.Put(ctx, sub+"/child.txt", content); err != nil {
			return "", nil, err
		}
		if err := disk.DeleteDirectory(ctx, sub); err != nil {
			return "", nil, err
		}
		exists, err := disk.DirectoryExists(ctx, sub)
		if err != nil {
			return "", nil, err
		}
		if !errors.Is(disk.DeleteDirectory(ctx, ""), filesystem.ErrEmptyDirectory) {
			return "", nil, errors.New("empty directory deletion was not rejected")
		}
		return fmt.Sprintf("nested_exists_after_delete=%t; root_protected=true", exists), nil, nil
	case "files", "all-files", "directories", "all-directories":
		if err := seedDirectoryTree(ctx, disk, dir); err != nil {
			return "", nil, err
		}
		var paths []string
		var err error
		switch name {
		case "files":
			paths, err = disk.Files(ctx, dir)
		case "all-files":
			paths, err = disk.AllFiles(ctx, dir)
		case "directories":
			paths, err = disk.Directories(ctx, dir)
		case "all-directories":
			paths, err = disk.AllDirectories(ctx, dir)
		}
		sort.Strings(paths)
		return fmt.Sprintf("count=%d", len(paths)), paths, err
	case "temporary-upload-url":
		result, err := disk.(*filesystem.Repository).TemporaryUploadURL(ctx, key, time.Now().Add(30*time.Minute), filesystem.TemporaryUploadURLOptions{ContentType: "text/plain", Visibility: filesystem.VisibilityPrivate})
		if err != nil {
			return "", nil, err
		}
		if result.URL == "" || result.Method != "PUT" {
			return "", nil, fmt.Errorf("upload URL = %q, method = %q; want URL and PUT", result.URL, result.Method)
		}
		return "method=PUT; signed_url=true", nil, nil
	}
	if err := disk.Put(ctx, key, content, filesystem.PutOptions{ContentType: "text/plain"}); err != nil {
		return "", nil, err
	}
	switch name {
	case "last-modified":
		modified, err := disk.LastModified(ctx, key)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("modified_recent=%t", time.Since(modified) < time.Minute), nil, nil
	case "file-info":
		info, err := disk.LastModifiedInfo(ctx, key)
		return fmt.Sprintf("path=%s; size=%d; content_type=%s; is_dir=%t", info.Path, info.Size, info.ContentType, info.IsDir), nil, err
	case "mime-type":
		mime, err := disk.MimeType(ctx, key)
		return mime, nil, err
	case "checksum":
		actual, err := disk.Checksum(ctx, key, filesystem.ChecksumOptions{Algorithm: "sha256"})
		if err != nil {
			return "", nil, err
		}
		expected := sha256.Sum256([]byte(content))
		if actual != hex.EncodeToString(expected[:]) {
			return "", nil, fmt.Errorf("checksum = %q, want %x", actual, expected)
		}
		return "sha256=" + actual, nil, nil
	case "path":
		actual := disk.Path(key)
		if !filepath.IsAbs(actual) {
			return "", nil, fmt.Errorf("local path = %q, want absolute", actual)
		}
		return actual, nil, nil
	case "public-url":
		actual, err := disk.URL(key)
		if err != nil {
			return "", nil, err
		}
		if !strings.Contains(actual, key) {
			return "", nil, fmt.Errorf("public URL = %q, want key %q", actual, key)
		}
		return actual, nil, nil
	case "temporary-url", "verify-temporary-url":
		signed, err := disk.TemporaryURL(ctx, key, time.Now().Add(5*time.Minute))
		if err != nil {
			return "", nil, err
		}
		parsed, err := url.Parse(signed)
		if err != nil {
			return "", nil, err
		}
		expires, err := time.Parse(time.RFC3339, parsed.Query().Get("expires"))
		if err != nil {
			return "", nil, fmt.Errorf("parse signed URL expiry: %w", err)
		}
		signature := parsed.Query().Get("signature")
		if err := filesystem.VerifyTemporaryURL(disk.Name(), key, expires, signature); err != nil {
			return "", nil, err
		}
		if name == "verify-temporary-url" {
			if !errors.Is(filesystem.VerifyTemporaryURL(disk.Name(), key, expires, "bad-signature"), filesystem.ErrTemporaryURLInvalid) {
				return "", nil, errors.New("tampered signature was accepted")
			}
			return "valid_signature=true; tampered_signature_rejected=true", nil, nil
		}
		return "signed_url=true; expiry_verified=true", nil, nil
	case "local-visibility":
		actual, err := disk.GetVisibility(ctx, key)
		if err != nil {
			return "", nil, err
		}
		if !errors.Is(disk.SetVisibility(ctx, key, filesystem.VisibilityPublic), filesystem.ErrUnsupportedVisibility) {
			return "", nil, errors.New("local visibility switch was accepted")
		}
		return "visibility=" + actual + "; switch_rejected=true", nil, nil
	case "oss-visibility":
		if err := disk.SetVisibility(ctx, key, filesystem.VisibilityPublic); err != nil {
			return "", nil, err
		}
		actual, err := disk.GetVisibility(ctx, key)
		if err != nil {
			return "", nil, err
		}
		if actual != filesystem.VisibilityPublic {
			return "", nil, fmt.Errorf("OSS visibility = %q, want public", actual)
		}
		return "visibility=public", nil, nil
	}
	return "", nil, fmt.Errorf("unhandled filesystem operation scenario %q", name)
}

func seedDirectoryTree(ctx context.Context, disk fscontract.Repository, dir string) error {
	for _, sub := range []string{"", "/nested", "/nested/deeper"} {
		if err := disk.MakeDirectory(ctx, dir+sub); err != nil {
			return fmt.Errorf("make directory %q: %w", dir+sub, err)
		}
	}
	for _, key := range []string{dir + "/root.txt", dir + "/nested/child.txt", dir + "/nested/deeper/leaf.txt"} {
		if err := disk.Put(ctx, key, "content"); err != nil {
			return fmt.Errorf("put tree file %q: %w", key, err)
		}
	}
	return nil
}
