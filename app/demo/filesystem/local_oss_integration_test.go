package filesystemdemo_test

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	filesystemdemo "prismgo-demo/app/demo/filesystem"
	"prismgo-demo/bootstrap"
)

// TestFilesystemDemoLocalOSSIntegration exercises the OSS extension against a local HTTP endpoint.
func TestFilesystemDemoLocalOSSIntegration(t *testing.T) {
	if os.Getenv("PRISMGO_FILESYSTEM_LOCAL_OSS_TEST") != "1" {
		t.Skip("set PRISMGO_FILESYSTEM_LOCAL_OSS_TEST=1 to run local OSS HTTP integration")
	}
	for _, scenario := range []struct{ name, want string }{
		{"oss-prerequisites", "OSS credentials configured"},
		{"oss-config", "driver=oss; bucket_configured=true; endpoint_configured=true"},
		{"temporary-upload-url", "method=PUT; signed_url=true"},
		{"oss-visibility", "visibility=public"},
		{"oss-capabilities", "temporary_url=true; temporary_upload=true"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			state := &localOSSState{objects: make(map[string]localOSSObject)}
			server := httptest.NewServer(state)
			t.Cleanup(server.Close)
			t.Setenv("FILESYSTEM_OSS_BUCKET", "demo-bucket")
			t.Setenv("FILESYSTEM_OSS_ENDPOINT", server.URL)
			t.Setenv("FILESYSTEM_OSS_ACCESS_KEY_ID", "local-key")
			t.Setenv("FILESYSTEM_OSS_ACCESS_KEY_SECRET", "local-secret")
			t.Setenv("FILESYSTEM_OSS_PREFIX", "")
			t.Setenv("FILESYSTEM_OSS_VISIBILITY", "private")
			t.Setenv("FILESYSTEM_SIGNING_KEY", "local-oss-test-signing-key")
			app := bootstrap.NewApplication(t.TempDir())
			if err := app.Boot(); err != nil {
				t.Fatalf("boot local OSS application error = %v, want nil", err)
			}
			t.Cleanup(func() {
				if err := app.Close(); err != nil {
					t.Errorf("close local OSS application error = %v, want nil", err)
				}
			})
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			result, err := filesystemdemo.Run(ctx, scenario.name)
			if err != nil || !strings.Contains(result.Value, scenario.want) {
				t.Fatalf("local OSS scenario %q result = %#v, error = %v; want value containing %q", scenario.name, result, err, scenario.want)
			}
			state.mu.Lock()
			remaining := len(state.objects)
			state.mu.Unlock()
			if remaining != 0 {
				t.Fatalf("local OSS scenario %q remaining objects = %d, want 0", scenario.name, remaining)
			}
		})
	}
}

type localOSSObject struct {
	data     []byte
	acl      string
	modified time.Time
}
type localOSSState struct {
	mu      sync.Mutex
	objects map[string]localOSSObject
}

func (s *localOSSState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	if len(parts) == 0 || parts[0] != "demo-bucket" {
		http.Error(w, "unknown bucket", http.StatusNotFound)
		return
	}
	key := ""
	if len(parts) == 2 {
		key = parts[1]
	}
	switch {
	case r.Method == http.MethodGet && key == "" && r.URL.Query().Get("list-type") == "2":
		s.list(w, r.URL.Query().Get("prefix"))
	case r.Method == http.MethodPut && r.URL.Query().Has("acl"):
		s.mu.Lock()
		object, ok := s.objects[key]
		if ok {
			object.acl = r.Header.Get("x-oss-object-acl")
			object.modified = time.Now().UTC()
			s.objects[key] = object
		}
		s.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodGet && r.URL.Query().Has("acl"):
		s.mu.Lock()
		object, ok := s.objects[key]
		s.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, `<AccessControlPolicy><Owner><ID>1</ID></Owner><AccessControlList><Grant>`+object.acl+`</Grant></AccessControlList></AccessControlPolicy>`)
	case r.Method == http.MethodPut && key != "":
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		acl := r.Header.Get("x-oss-object-acl")
		if acl == "" {
			acl = "private"
		}
		s.mu.Lock()
		s.objects[key] = localOSSObject{data: data, acl: acl, modified: time.Now().UTC()}
		s.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodDelete && key != "":
		s.mu.Lock()
		delete(s.objects, key)
		s.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, fmt.Sprintf("unsupported %s %s", r.Method, r.URL.String()), http.StatusBadRequest)
	}
}

func (s *localOSSState) list(w http.ResponseWriter, prefix string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	type object struct {
		Key          string `xml:"Key"`
		Size         int    `xml:"Size"`
		LastModified string `xml:"LastModified"`
	}
	result := struct {
		XMLName     xml.Name `xml:"ListBucketResult"`
		IsTruncated bool     `xml:"IsTruncated"`
		Objects     []object `xml:"Contents"`
	}{}
	for key, item := range s.objects {
		if strings.HasPrefix(key, prefix) {
			result.Objects = append(result.Objects, object{Key: key, Size: len(item.data), LastModified: item.modified.Format(time.RFC3339)})
		}
	}
	w.Header().Set("Content-Type", "application/xml")
	if err := xml.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "encode list", http.StatusInternalServerError)
	}
}
