package schemademo_test

import (
	"os"
	"path/filepath"
	"testing"

	schemademo "prismgo-demo/app/demo/schema"
	demotest "prismgo-demo/app/demo/testing"
	"prismgo-demo/bootstrap"
)

// newMySQLApplication boots an isolated application whose default database is
// the real MySQL service described by PRISMGO_MYSQL_TEST_DSN. It skips when the
// variable is absent so the hermetic suite stays runnable without Docker.
func newMySQLApplication(t *testing.T) {
	t.Helper()
	dsn := demotest.RequireIntegration(t, demotest.ServiceMySQL)
	basePath := t.TempDir()
	paths := map[string]string{
		"DB_DATABASE":            filepath.Join(basePath, "storage", "database.sqlite"),
		"CACHE_FILE_PATH":        filepath.Join(basePath, "storage", "framework", "cache", "data"),
		"CACHE_FILE_LOCK_PATH":   filepath.Join(basePath, "storage", "framework", "cache", "locks"),
		"SESSION_FILES":          filepath.Join(basePath, "storage", "framework", "sessions"),
		"FILESYSTEM_LOCAL_ROOT":  filepath.Join(basePath, "storage", "app", "private"),
		"FILESYSTEM_PUBLIC_ROOT": filepath.Join(basePath, "storage", "app", "public"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("prepare mysql integration path: %v", err)
		}
	}
	environment := map[string]string{
		"APP_ENV": "testing", "APP_DEBUG": "false", "APP_KEY": "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION": "mysql", "DB_DSN": dsn,
		"CACHE_STORE": "memory", "QUEUE_CONNECTION": "sync", "SESSION_DRIVER": "file",
		"REDIS_URL": "", "REDIS_CACHE_URL": "", "RABBITMQ_URL": "",
		"FILESYSTEM_DISK": "local",
	}
	for key, value := range paths {
		environment[key] = value
	}
	for key, value := range environment {
		t.Setenv(key, value)
	}
	app := bootstrap.NewApplication(basePath)
	if err := app.Boot(); err != nil {
		t.Fatalf("boot mysql application: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close mysql application: %v", err)
		}
	})
}

func TestSchemaDemoMySQLDefaultStringLength(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "default-string-length", "name=varchar(191) code=char(191)")
}

func TestSchemaDemoMySQLDefaultTimePrecision(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "default-time-precision", "datetime=datetime(3) time=time(3) timestamp=timestamp(3)")
}

func TestSchemaDemoMySQLDefaultMorphKeyType(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "default-morph-key-type", "owner_id=bigint unsigned owner_type=varchar(255)")
}

func TestSchemaDemoMySQLMorphUsingUUIDs(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "morph-using-uuids", "owner_id=char(36) owner_type=varchar(255)")
}

func TestSchemaDemoMySQLMorphUsingULIDs(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "morph-using-ulids", "owner_id=char(26) owner_type=varchar(255)")
}

func TestSchemaDemoMySQLExplicitTagPrecedence(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "explicit-tag-precedence", "name=varchar(191) sized=varchar(77) typed=char(10)")
}

func TestSchemaDemoMySQLChangeColumn(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "change-column", "type=varchar(64) nullable=true")
}

func TestSchemaDemoMySQLRaw(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "raw", "raw=true generated=true")
}

// TestSchemaDemoMySQLColumnTypes asserts the exact MySQL column definitions
// produced by the column-type scenarios, which SQLite collapses into a coarse
// type affinity. spatial-types and vector are excluded because this MySQL
// version rejects geography and vector(n).
func TestSchemaDemoMySQLColumnTypes(t *testing.T) {
	newMySQLApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "increments", want: "families=5 primary=true generated=true"},
		{name: "signed-integers", want: "big=bigint int=int medium=mediumint small=smallint tiny=tinyint"},
		{name: "unsigned-integers", want: "big=bigint unsigned int=int unsigned medium=mediumint unsigned small=smallint unsigned tiny=tinyint unsigned"},
		{name: "string-char", want: "name=varchar(120) code=char(12)"},
		{name: "text-types", want: "body=text summary=tinytext notes=mediumtext archive=longtext"},
		{name: "uuid-ulid", want: "identifier=char(36) sequence=char(26)"},
		{name: "network-addresses", want: "ip=varchar(45) mac=varchar(17)"},
		{name: "remember-token", want: "remember_token=varchar(100) nullable=true"},
		{name: "boolean", want: "active=tinyint(1) archived=tinyint(1)"},
		{name: "floating-point", want: "ratio=float(8,2) score=double(10,4)"},
		{name: "decimal", want: "price=decimal(10,2) total=decimal(12,3) unsigned"},
		{name: "date", want: "born_on=date"},
		{name: "datetime", want: "seen_at=datetime seen_at_tz=datetime"},
		{name: "time", want: "starts_at=time starts_at_tz=time"},
		{name: "timestamp", want: "published_at=timestamp published_at_tz=timestamp"},
		{name: "year", want: "release_year=year"},
		{name: "timestamps", want: "created_at=timestamp nullable=true updated_at=timestamp nullable=true"},
		{name: "nullable-timestamps", want: "created_at=timestamp nullable=true updated_at=timestamp nullable=true"},
		{name: "soft-deletes", want: "deleted_at=timestamp nullable=true index=true"},
		{name: "binary", want: "payload=blob"},
		{name: "json", want: "meta=json meta_b=json"},
		{name: "enum-set", want: "status=enum('draft','published') flags=set('featured','pinned')"},
		{name: "foreign-id", want: "owner_id=bigint unsigned"},
		{name: "foreign-id-for", want: "user_id=bigint unsigned"},
		{name: "morphs", want: "taggable_id=bigint unsigned taggable_type=varchar(255) index=true"},
		{name: "nullable-morphs", want: "taggable_id=bigint unsigned nullable=true taggable_type=varchar(255) nullable=true index=true"},
		{name: "nullable", want: "nickname_nullable=true bio_nullable=true"},
		{name: "not-null", want: "nickname_nullable=false bio_nullable=false legacy_nullable=false"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

// TestSchemaDemoMySQLRequiresServer asserts the MySQL-only scenarios fail with a
// clear dialect error instead of silently degrading on SQLite.
func TestSchemaDemoMySQLRequiresServer(t *testing.T) {
	newApplication(t)
	for _, name := range []string{"default-string-length", "change-column", "raw"} {
		if _, err := schemademo.Run(name); err == nil {
			t.Fatalf("schema demo %q on sqlite error = nil, want MySQL requirement", name)
		}
	}
}
