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
	for _, name := range []string{
		"default-string-length", "change-column", "raw",
		"unique-modifier", "comment", "first", "after", "charset", "collation",
		"use-current", "use-current-on-update", "invisible", "drop-constrained-foreign-id",
		"rename-index", "drop-primary",
		"constrained", "constrained-explicit", "foreign", "foreign-actions", "cascade-actions",
		"restrict-actions", "null-actions", "no-action-actions", "foreign-name", "drop-foreign",
		"foreign-keys", "disable-foreign-keys", "enable-foreign-keys", "without-foreign-keys",
		"create-database", "drop-database",
	} {
		if _, err := schemademo.Run(name); err == nil {
			t.Fatalf("schema demo %q on sqlite error = nil, want MySQL requirement", name)
		}
	}
}

// TestSchemaDemoMySQLModifierScenarios asserts the modifier scenarios that run on
// both dialects against real MySQL.
func TestSchemaDemoMySQLModifierScenarios(t *testing.T) {
	newMySQLApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "unsigned", want: "count=int unsigned total=bigint unsigned"},
		{name: "auto-increment", want: "auto_increment=true generated=true"},
		{name: "primary-modifier", want: "code_primary=true label_primary=false"},
		{name: "index-modifier", want: "index=true removed=true"},
		{name: "unique-modifier", want: "unique=true enforced=true removed=true"},
		{name: "comment", want: "comment=hello"},
		{name: "first", want: "first=pinned"},
		{name: "after", want: "order=id,name,inserted,tail"},
		{name: "charset", want: "charset=latin1"},
		{name: "collation", want: "collation=utf8mb4_unicode_ci"},
		{name: "use-current", want: "created_at_default=CURRENT_TIMESTAMP"},
		{name: "use-current-on-update", want: "on_update=true"},
		{name: "invisible", want: "invisible=true"},
		{name: "drop-remember-token", want: "before=true after=false"},
		{name: "drop-timestamps", want: "before=true after=false"},
		{name: "drop-timestamps-tz", want: "before=true after=false"},
		{name: "drop-soft-deletes", want: "before=true column_after=false index_after=false"},
		{name: "drop-soft-deletes-tz", want: "before=true column_after=false index_after=false"},
		{name: "drop-morphs", want: "before=true columns_after=false index_after=false"},
		{name: "drop-constrained-foreign-id", want: "foreign=true column_after=false foreign_after=false"},
		{name: "drop-foreign-id-for", want: "before=true after=false"},
		{name: "primary-index", want: "primary=true"},
		{name: "unique-index", want: "unique=true"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

// TestSchemaDemoMySQLDefault asserts the MySQL default value literals, where the
// SQL literal handling differs from SQLite.
func TestSchemaDemoMySQLDefault(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "default", "nickname=guest enabled=1 sort=0 created_at=CURRENT_TIMESTAMP payload=")
}

// TestSchemaDemoMySQLIndexScenarios asserts the index scenarios against real
// MySQL, including index renaming and primary key dropping that SQLite cannot do.
func TestSchemaDemoMySQLIndexScenarios(t *testing.T) {
	newMySQLApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "index", want: "index=true name=schema_demo_index_status_index"},
		{name: "fulltext-index", want: "index=true name=schema_demo_fulltext_index_body_fulltext"},
		{name: "spatial-index", want: "index=true name=schema_demo_spatial_index_location_spatial"},
		{name: "named-indexes", want: "unique=true index=true"},
		{name: "index-naming", want: "short=true trimmed=true"},
		{name: "rename-index", want: "renamed=true source=false"},
		{name: "drop-index", want: "before=true after=false"},
		{name: "drop-unique", want: "before=true after=false"},
		{name: "drop-primary", want: "before=true after=false"},
		{name: "drop-fulltext", want: "before=true after=false"},
		{name: "drop-spatial-index", want: "before=true after=false"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

// TestSchemaDemoMySQLForeignKeyScenarios asserts foreign key creation, actions,
// naming and dropping against real MySQL.
func TestSchemaDemoMySQLForeignKeyScenarios(t *testing.T) {
	newMySQLApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "constrained", want: "table=schema_demo_parents column=id"},
		{name: "constrained-explicit", want: "table=schema_demo_fk_explicit_owners column=id"},
		{name: "foreign", want: "table=schema_demo_fk_manual_tenants column=id"},
		{name: "foreign-actions", want: "on_update=SET NULL on_delete=SET NULL"},
		{name: "cascade-actions", want: "on_update=CASCADE on_delete=CASCADE"},
		{name: "restrict-actions", want: "on_update=RESTRICT on_delete=RESTRICT"},
		{name: "null-actions", want: "on_update=SET NULL on_delete=SET NULL"},
		{name: "no-action-actions", want: "on_update=NO ACTION on_delete=NO ACTION"},
		{name: "foreign-name", want: "name=fk_schema_demo_fk_name_owner"},
		{name: "drop-foreign", want: "dropped=true remaining=0"},
		{name: "foreign-dialect", want: "dialect=mysql foreign_keys=1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

// TestSchemaDemoMySQLInspectionScenarios asserts the metadata inspection
// scenarios against real MySQL.
func TestSchemaDemoMySQLInspectionScenarios(t *testing.T) {
	newMySQLApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "table-view-existence", want: "table=true view=true"},
		{name: "tables", want: "found=true schema_set=true filtered=true"},
		{name: "table-listing", want: "qualified=true bare=true"},
		{name: "views", want: "found=true definition=true"},
		{name: "schemas", want: "count_positive=true name_set=true"},
		{name: "types", want: "count=0 empty=true"},
		{name: "schema-filter", want: "string=true slice=true nil=true"},
		{name: "has-columns", want: "all=true missing=false empty=true"},
		{name: "columns", want: "count=3 names=true nullable=true primary=true"},
		{name: "column-type", want: "short=varchar full=varchar(64) missing_error=true"},
		{name: "has-index", want: "name=true columns=true type=true wrong_type=false missing=false"},
		{name: "indexes", want: "names=true unique=true plain=true"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

// TestSchemaDemoMySQLConditionalAndSyncScenarios asserts the conditional and
// SyncModels scenarios against real MySQL.
func TestSchemaDemoMySQLConditionalAndSyncScenarios(t *testing.T) {
	newMySQLApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "when-has-column", want: "executed=true added=true absent_executed=false"},
		{name: "when-missing-column", want: "present_executed=false missing_executed=true added=true"},
		{name: "when-missing-index", want: "existing_executed=false missing_executed=true created=true"},
		{name: "sync-models", want: "created=true columns=2"},
		{name: "sync-models-columns", want: "created=true added=true"},
		{name: "sync-models-defaults", want: "default=active options=ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

func TestSchemaDemoForeignKeys(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "foreign-keys", "columns=1 foreign=schema_demo_fk_inspect_parents on_update=CASCADE on_delete=CASCADE")
}

func TestSchemaDemoDisableForeignKeys(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "disable-foreign-keys", "before=1 after=0")
}

func TestSchemaDemoEnableForeignKeys(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "enable-foreign-keys", "before=0 after=1")
}

func TestSchemaDemoWithoutForeignKeys(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "without-foreign-keys", "inside=0 restored=1 error_precedence=true")
}

func TestSchemaDemoCreateDatabase(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "create-database", "created=true exists=true")
}

func TestSchemaDemoDropDatabase(t *testing.T) {
	newMySQLApplication(t)
	expectValue(t, "drop-database", "dropped=true exists=false")
}
