package schemademo_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/prismgo/framework/database"
	dbschema "github.com/prismgo/framework/database/schema"

	schemademo "prismgo-demo/app/demo/schema"
	demotest "prismgo-demo/app/demo/testing"
	"prismgo-demo/bootstrap"
)

// newSQLiteIntegrationApplication boots an application whose default database is
// the real SQLite service described by PRISMGO_SQLITE_TEST_DSN. It skips when the
// variable is absent so the hermetic suite stays runnable without the provisioned
// SQLite runtime.
func newSQLiteIntegrationApplication(t *testing.T) {
	t.Helper()
	dsn := demotest.RequireIntegration(t, demotest.ServiceSQLite)
	basePath := t.TempDir()
	paths := map[string]string{
		"CACHE_FILE_PATH":        filepath.Join(basePath, "storage", "framework", "cache", "data"),
		"CACHE_FILE_LOCK_PATH":   filepath.Join(basePath, "storage", "framework", "cache", "locks"),
		"SESSION_FILES":          filepath.Join(basePath, "storage", "framework", "sessions"),
		"FILESYSTEM_LOCAL_ROOT":  filepath.Join(basePath, "storage", "app", "private"),
		"FILESYSTEM_PUBLIC_ROOT": filepath.Join(basePath, "storage", "app", "public"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("prepare sqlite integration path: %v", err)
		}
	}
	environment := map[string]string{
		"APP_ENV": "testing", "APP_DEBUG": "false", "APP_KEY": "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION": "sqlite", "DB_DATABASE": dsn,
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
		t.Fatalf("boot sqlite integration application: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close sqlite integration application: %v", err)
		}
	})
}

// TestSchemaDemoSQLiteIntegrationBatch accepts the batch scenarios that run on
// both dialects against the provisioned SQLite service.
func TestSchemaDemoSQLiteIntegrationBatch(t *testing.T) {
	newSQLiteIntegrationApplication(t)
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "unsigned", want: "count=integer total=integer"},
		{name: "auto-increment", want: "auto_increment=true generated=true"},
		{name: "primary-modifier", want: "code_primary=true label_primary=false"},
		{name: "index-modifier", want: "index=true removed=true"},
		{name: "default", want: "nickname='guest' enabled=1 sort=0 created_at=CURRENT_TIMESTAMP payload="},
		{name: "drop-remember-token", want: "before=true after=false"},
		{name: "drop-timestamps", want: "before=true after=false"},
		{name: "drop-timestamps-tz", want: "before=true after=false"},
		{name: "drop-soft-deletes", want: "before=true column_after=false index_after=false"},
		{name: "drop-soft-deletes-tz", want: "before=true column_after=false index_after=false"},
		{name: "drop-morphs", want: "before=true columns_after=false index_after=false"},
		{name: "drop-foreign-id-for", want: "before=true after=false"},
		{name: "primary-index", want: "primary=true"},
		{name: "unique-index", want: "unique=true"},
		{name: "stored-as", want: "stored_as=placeholder compiled=false"},
		{name: "virtual-as", want: "virtual_as=placeholder compiled=false"},
		{name: "from", want: "from=compatibility affects_sql=false"},
		{name: "instant", want: "instant=compatibility affects_sql=false"},
		{name: "lock", want: "lock=compatibility affects_sql=false"},
		{name: "change-semantics", want: "change=modify_column idempotency=caller"},
		{name: "index", want: "index=true name=schema_demo_index_status_index"},
		{name: "fulltext-index", want: "index=true name=schema_demo_fulltext_index_body_fulltext"},
		{name: "spatial-index", want: "index=true name=schema_demo_spatial_index_location_spatial"},
		{name: "named-indexes", want: "unique=true index=true"},
		{name: "index-naming", want: "short=true trimmed=true"},
		{name: "drop-index", want: "before=true after=false"},
		{name: "drop-unique", want: "before=true after=false"},
		{name: "drop-fulltext", want: "before=true after=false"},
		{name: "drop-spatial-index", want: "before=true after=false"},
		{name: "foreign-dialect", want: "dialect=sqlite foreign_keys=0"},
		{name: "table-view-existence", want: "table=true view=true"},
		{name: "tables", want: "found=true schema_set=true filtered=true"},
		{name: "table-listing", want: "qualified=true bare=true"},
		{name: "views", want: "found=true definition=true"},
		{name: "schemas", want: "count_positive=true name_set=true"},
		{name: "types", want: "count=0 empty=true"},
		{name: "schema-filter", want: "string=true slice=true nil=true"},
		{name: "has-columns", want: "all=true missing=false empty=true"},
		{name: "columns", want: "count=3 names=true nullable=true primary=true"},
		{name: "column-type", want: "short=text full=text missing_error=true"},
		{name: "has-index", want: "name=true columns=true type=true wrong_type=false missing=false"},
		{name: "indexes", want: "names=true unique=true plain=true"},
		{name: "when-has-column", want: "executed=true added=true absent_executed=false"},
		{name: "when-missing-column", want: "present_executed=false missing_executed=true added=true"},
		{name: "when-missing-index", want: "existing_executed=false missing_executed=true created=true"},
		{name: "sync-models", want: "created=true columns=2"},
		{name: "sync-models-columns", want: "created=true added=true"},
		{name: "sync-models-defaults", want: "default=active options=ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"},
		{name: "metadata-types", want: "schema=SchemaInfo table=TableInfo view=ViewInfo type=TypeInfo column=ColumnInfo index=IndexInfo foreign=ForeignKeyInfo"},
		{name: "foreign-key-toggle-dialects", want: "mysql=SET FOREIGN_KEY_CHECKS sqlite=PRAGMA foreign_keys"},
		{name: "sync-models-boundaries", want: "entry=SyncModels creates_tables=true adds_columns=true drops_columns=false auto_migrate=false"},
		{name: "dialect-compatibility", want: "shared=create,index,drop-index mysql_only=rename-index,drop-primary,create-database,sync-models sqlite_only=sqlite-extension"},
		{name: "laravel-compatibility", want: "create=Schema::create table=Schema::table checks=Schema::hasTable columns=Schema::hasColumn indexes=Schema::hasIndex drop=Schema::dropIfExists"},
		{name: "ensure-extension", want: "extension=postgis mysql=unsupported sqlite=unsupported"},
		{name: "ensure-vector-extension", want: "extension=vector mysql=unsupported sqlite=unsupported"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expectValue(t, testCase.name, testCase.want)
		})
	}
}

// TestSchemaDemoSQLiteIntegrationDialectBoundary asserts the MySQL-only batch
// scenarios fail with a clear dialect error instead of silently degrading on
// SQLite.
func TestSchemaDemoSQLiteIntegrationDialectBoundary(t *testing.T) {
	newSQLiteIntegrationApplication(t)
	for _, name := range []string{
		"unique-modifier", "comment", "first", "after", "charset", "collation",
		"use-current", "use-current-on-update", "invisible", "drop-constrained-foreign-id",
		"rename-index", "drop-primary",
		"constrained", "constrained-explicit", "foreign", "foreign-actions", "cascade-actions",
		"restrict-actions", "null-actions", "no-action-actions", "foreign-name", "drop-foreign",
		"foreign-keys", "disable-foreign-keys", "enable-foreign-keys", "without-foreign-keys",
		"create-database", "drop-database",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := schemademo.Run(name); err == nil {
				t.Fatalf("schema demo %q on sqlite error = nil, want MySQL requirement", name)
			}
		})
	}
}

// TestSchemaDemoSQLiteIntegrationColumnUnique guards the field-level Unique
// modifier on SQLite: Unique() registers a default-named unique index that the
// metadata API can see, and the database enforces it.
func TestSchemaDemoSQLiteIntegrationColumnUnique(t *testing.T) {
	newSQLiteIntegrationApplication(t)
	db := database.Resolve()
	if db == nil {
		t.Fatal("application database connection is not available")
	}
	builder := dbschema.New(db)
	const table = "schema_demo_sqlite_unique"
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 128).Unique()
	}); err != nil {
		t.Fatalf("create %s: %v", table, err)
	}
	t.Cleanup(func() { _ = builder.Drop(table) })

	if !builder.HasIndex(table, []string{"email"}, "unique") {
		t.Errorf("field Unique() index visibility on sqlite = false, want true")
	}
	if err := db.Exec(fmt.Sprintf("INSERT INTO %s (email) VALUES ('a')", table)).Error; err != nil {
		t.Fatalf("insert first row into %s: %v", table, err)
	}
	if err := db.Exec(fmt.Sprintf("INSERT INTO %s (email) VALUES ('a')", table)).Error; err == nil {
		t.Errorf("duplicate insert into %s error = nil, want unique violation", table)
	}
}
