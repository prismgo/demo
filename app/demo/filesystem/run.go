// Package filesystemdemo contains runnable local filesystem examples.
package filesystemdemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/prismgo/framework/config"
	fscontract "github.com/prismgo/framework/contracts/filesystem"
	"github.com/prismgo/framework/filesystem"
)

// Result records the observable result of a filesystem example.
type Result struct {
	Case    string   `json:"case"`
	Key     string   `json:"key,omitempty"`
	Value   string   `json:"value"`
	Details []string `json:"details,omitempty"`
}

// Run executes a documented filesystem example in the current application.
func Run(ctx context.Context, name string) (result Result, runErr error) {
	result = Result{Case: name}
	if isNextCase(name) {
		return runNext(ctx, name)
	}
	if value, ok, err := configurationResult(name); ok || err != nil {
		result.Value = value
		return result, err
	}
	if !operationCase(name) {
		return Result{}, fmt.Errorf("unknown filesystem scenario %q", name)
	}
	dir := fmt.Sprintf("demo/filesystem/%s-%d", name, time.Now().UnixNano())
	disk := filesystem.Default()
	if name == "named-disk" {
		disk = filesystem.Disk("public")
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := disk.DeleteDirectory(cleanupCtx, dir); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("clean up demo directory %q: %w", dir, err))
		}
	}()
	result.Key = dir + "/example.txt"
	value, details, err := operationResult(ctx, name, disk, result.Key, dir)
	if err != nil {
		return Result{}, fmt.Errorf("filesystem demo %s: %w", name, err)
	}
	result.Value, result.Details = value, details
	return result, nil
}

func configurationResult(name string) (string, bool, error) {
	cfg := config.Resolve()
	switch name {
	case "architecture":
		var factory fscontract.Manager = filesystem.Resolve()
		return fmt.Sprintf("%T -> %s", factory, factory.Default().Name()), true, nil
	case "config":
		return fmt.Sprintf("filesystem.default=%s; disks=%d", cfg.GetString("filesystem.default"), len(cfg.GetStringMap("filesystem.disks"))), true, nil
	case "local-prerequisites":
		root := cfg.GetString("filesystem.disks.local.root")
		if root == "" {
			return "", true, fmt.Errorf("local disk root is empty")
		}
		return "local driver: writable root " + root, true, nil
	case "top-level-config":
		return fmt.Sprintf("default=%s; cloud=%s; signing_key_configured=%t", cfg.GetString("filesystem.default"), cfg.GetString("filesystem.cloud"), cfg.GetString("filesystem.temporary_url.signing_key") != ""), true, nil
	case "local-config", "public-config":
		disk := strings.TrimSuffix(name, "-config")
		fields := cfg.GetStringMap("filesystem.disks." + disk)
		if len(fields) == 0 {
			return "", true, fmt.Errorf("disk %q is not configured", disk)
		}
		return fmt.Sprintf("%s: driver=%v; root=%v; visibility=%v; serve=%v", disk, fields["driver"], fields["root"], fields["visibility"], fields["serve"]), true, nil
	case "links-config":
		links := cfg.GetStringMap("filesystem.links")
		return fmt.Sprintf("public/storage=%v", links["public/storage"]), true, nil
	case "disk-config":
		spec := filesystem.DiskConfig{Driver: "local", Root: "storage/app/private", Visibility: filesystem.VisibilityPrivate, Serve: true}
		return fmt.Sprintf("driver=%s; root=%s; visibility=%s; serve=%t", spec.Driver, spec.Root, spec.Visibility, spec.Serve), true, nil
	case "cloud":
		return fmt.Sprintf("cloud=%s; driver=%s", filesystem.CloudName(), cfg.GetString("filesystem.disks."+filesystem.CloudName()+".driver")), true, nil
	case "interface-selection":
		var repo fscontract.Repository = filesystem.Default()
		return fmt.Sprintf("repository=%s; factory=%T", repo.Name(), filesystem.Resolve()), true, nil

	default:
		return "", false, nil
	}
}

func operationCase(name string) bool {
	switch name {
	case "facade", "named-disk", "get", "json", "open-stream", "read-stream", "download", "file-existence", "directory-existence", "put", "put-reader", "prepend-append", "put-options", "put-file", "put-file-as", "upload-fields", "copy-move", "cross-disk-guard", "delete", "size":
		return true
	}
	return false
}

func operationResult(ctx context.Context, name string, disk fscontract.Repository, key, dir string) (string, []string, error) {
	const content = "hello filesystem"
	switch name {
	case "facade":
		if err := filesystem.Put(ctx, key, content); err != nil {
			return "", nil, err
		}
		data, err := filesystem.Get(ctx, key)
		return string(data), []string{"disk:" + filesystem.Name()}, err
	case "named-disk":
		if err := disk.Put(ctx, key, content); err != nil {
			return "", nil, err
		}
		data, err := disk.Get(ctx, key)
		return string(data), []string{"disk:" + disk.Name()}, err
	case "put":
		for _, item := range []struct {
			key   string
			value any
		}{{key + ".string", content}, {key + ".bytes", []byte(content)}, {key, strings.NewReader(content)}} {
			if err := disk.Put(ctx, item.key, item.value, filesystem.PutOptions{ContentType: "text/plain"}); err != nil {
				return "", nil, err
			}
		}
		data, err := disk.Get(ctx, key)
		return string(data), []string{"input:string,bytes,reader"}, err
	case "put-reader":
		if err := disk.PutReader(ctx, key, strings.NewReader(content), filesystem.PutOptions{ContentType: "text/plain"}); err != nil {
			return "", nil, err
		}
		if err := disk.WriteStream(ctx, key, strings.NewReader("stream alias"), filesystem.PutOptions{ContentType: "text/plain"}); err != nil {
			return "", nil, err
		}
		data, err := disk.Get(ctx, key)
		return string(data), []string{"aliases:PutReader,WriteStream"}, err
	case "prepend-append":
		if err := disk.Put(ctx, key, "middle"); err != nil {
			return "", nil, err
		}
		if err := disk.Prepend(ctx, key, "first"); err != nil {
			return "", nil, err
		}
		if err := disk.Append(ctx, key, "last"); err != nil {
			return "", nil, err
		}
		data, err := disk.Get(ctx, key)
		return string(data), nil, err
	case "put-options":
		if err := disk.Put(ctx, key, content, filesystem.PutOptions{ContentType: "text/plain", Visibility: filesystem.VisibilityPrivate}); err != nil {
			return "", nil, err
		}
		info, err := disk.LastModifiedInfo(ctx, key)
		if err != nil {
			return "", nil, err
		}
		visibility, err := disk.GetVisibility(ctx, key)
		return fmt.Sprintf("content_type=%s; visibility=%s", info.ContentType, visibility), nil, err
	case "put-file", "put-file-as", "upload-fields":
		header, err := uploadHeader("example.txt", content)
		if err != nil {
			return "", nil, err
		}
		saved := ""
		if name == "put-file" || name == "upload-fields" {
			saved, err = disk.PutFile(ctx, dir, header)
		} else {
			saved, err = disk.PutFileAs(ctx, dir, header, "renamed.txt")
		}
		if err != nil {
			return "", nil, err
		}
		data, err := disk.Get(ctx, saved)
		if err != nil {
			return "", nil, err
		}
		if name == "upload-fields" {
			info, err := disk.LastModifiedInfo(ctx, saved)
			if err != nil {
				return "", nil, err
			}
			return fmt.Sprintf("disk=%s; path=%s; original_name=%s; mime_type=%s; size=%d", disk.Name(), saved, header.Filename, info.ContentType, info.Size), nil, nil
		}
		return string(data), []string{"saved:" + saved}, nil
	case "copy-move":
		if err := disk.Put(ctx, key, content); err != nil {
			return "", nil, err
		}
		copied, moved := dir+"/copy.txt", dir+"/moved.txt"
		if err := disk.Copy(ctx, key, copied); err != nil {
			return "", nil, err
		}
		if err := disk.Move(ctx, copied, moved); err != nil {
			return "", nil, err
		}
		exists, err := disk.Exists(ctx, copied)
		if err != nil {
			return "", nil, err
		}
		data, err := disk.Get(ctx, moved)
		return string(data), []string{fmt.Sprintf("source_copy_exists:%t", exists)}, err
	case "cross-disk-guard":
		if err := disk.Put(ctx, key, content); err != nil {
			return "", nil, err
		}
		target := dir + "/copy.txt"
		if err := disk.Copy(ctx, key, target); err != nil {
			return "", nil, err
		}
		onSource, err := disk.Exists(ctx, target)
		if err != nil {
			return "", nil, err
		}
		onOther, err := filesystem.Disk("public").Exists(ctx, target)
		return fmt.Sprintf("source_disk=%t; other_disk=%t", onSource, onOther), []string{"copy stays on " + disk.Name()}, err
	case "delete":
		if err := disk.Put(ctx, key, content); err != nil {
			return "", nil, err
		}
		other := dir + "/second.txt"
		if err := disk.Put(ctx, other, content); err != nil {
			return "", nil, err
		}
		if err := disk.Delete(ctx, key, other); err != nil {
			return "", nil, err
		}
		first, err := disk.Missing(ctx, key)
		if err != nil {
			return "", nil, err
		}
		second, err := disk.Missing(ctx, other)
		return fmt.Sprintf("first_missing=%t; second_missing=%t", first, second), nil, err
	case "file-existence":
		missing, err := disk.Missing(ctx, key)
		if err != nil {
			return "", nil, err
		}
		if err := disk.Put(ctx, key, content); err != nil {
			return "", nil, err
		}
		exists, err := disk.Exists(ctx, key)
		if err != nil {
			return "", nil, err
		}
		alias, err := disk.FileExists(ctx, key)
		return fmt.Sprintf("missing_before=%t; exists_after=%t; file_exists=%t", missing, exists, alias), nil, err
	case "directory-existence":
		before, err := disk.DirectoryExists(ctx, dir)
		if err != nil {
			return "", nil, err
		}
		if err := disk.MakeDirectory(ctx, dir); err != nil {
			return "", nil, err
		}
		after, err := disk.DirectoryExists(ctx, dir)
		return fmt.Sprintf("before=%t; after=%t", before, after), nil, err
	}
	if err := disk.Put(ctx, key, content); err != nil {
		return "", nil, err
	}
	switch name {
	case "get":
		data, err := disk.Get(ctx, key)
		return string(data), nil, err
	case "json":
		if err := disk.Put(ctx, key, `{"name":"PrismGo"}`); err != nil {
			return "", nil, err
		}
		var record struct {
			Name string `json:"name"`
		}
		if err := disk.JSON(ctx, key, &record); err != nil {
			return "", nil, err
		}
		return record.Name, nil, nil
	case "open-stream":
		reader, info, err := disk.OpenStream(ctx, key)
		if err != nil {
			return "", nil, err
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		return string(data), []string{fmt.Sprintf("size:%d", info.Size)}, err
	case "read-stream":
		reader, err := disk.ReadStream(ctx, key)
		if err != nil {
			return "", nil, err
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		return string(data), nil, err
	case "download":
		var out bytes.Buffer
		err := disk.Download(ctx, key, &out)
		return out.String(), nil, err
	case "size":
		size, err := disk.Size(ctx, key)
		return fmt.Sprintf("size=%d", size), nil, err
	}
	return "", nil, errors.New("unreachable filesystem scenario")
}

func uploadHeader(filename, content string) (*multipart.FileHeader, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create upload form: %w", err)
	}
	if _, err := io.WriteString(part, content); err != nil {
		return nil, fmt.Errorf("write upload form: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close upload form: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, "http://localhost/upload", &body)
	if err != nil {
		return nil, fmt.Errorf("create upload request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(1 << 20); err != nil {
		return nil, fmt.Errorf("parse upload form: %w", err)
	}
	headers := request.MultipartForm.File["file"]
	if len(headers) != 1 {
		return nil, fmt.Errorf("upload headers = %d, want 1", len(headers))
	}
	return headers[0], nil
}
