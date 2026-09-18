package schemademo_test

import (
	"testing"

	schemademo "prismgo-demo/app/demo/schema"
	demotest "prismgo-demo/app/demo/testing"
)

// expectValue executes one scenario and asserts its full observable value.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	result, err := schemademo.Run(name)
	if err != nil {
		t.Fatalf("schema demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("schema demo %q case = %q, want %q", name, result.Case, name)
	}
	if result.Value != want {
		t.Fatalf("schema demo %q value = %q, want %q", name, result.Value, want)
	}
}

// newApplication boots an isolated application whose default database is SQLite.
func newApplication(t *testing.T) {
	t.Helper()
	demotest.NewApplication(t, demotest.Options{})
}

func TestSchemaDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "builder=Builder blueprint=Blueprint column=ColumnDefinition index=IndexDefinition foreign=ForeignKeyDefinition err=ErrUnsupportedFeature")
}

func TestSchemaDemoSQLiteExtension(t *testing.T) {
	expectValue(t, "sqlite-extension", "dialect=sqlite provider=prismgo.extension.sqlite table=true")
}

func TestSchemaDemoSQLiteConnectionScope(t *testing.T) {
	expectValue(t, "sqlite-connection-scope", "max_open_conns=1 pinned=true")
}

func TestSchemaDemoCreateDialectOptions(t *testing.T) {
	expectValue(t, "create-dialect-options", "mysql=ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 sqlite=")
}

func TestSchemaDemoDropAllTypes(t *testing.T) {
	expectValue(t, "drop-all-types", "mysql=empty sqlite=empty")
}

func TestSchemaDemoDropAllTables(t *testing.T) {
	expectValue(t, "drop-all-tables", "remaining=0")
}

func TestSchemaDemoDropAllViews(t *testing.T) {
	expectValue(t, "drop-all-views", "views=0")
}

func TestSchemaDemoFacade(t *testing.T) {
	newApplication(t)
	expectValue(t, "facade", "create=true has_table=true has_column=true")
}

func TestSchemaDemoBind(t *testing.T) {
	newApplication(t)
	expectValue(t, "bind", "bound=true chained=true")
}

func TestSchemaDemoNew(t *testing.T) {
	newApplication(t)
	expectValue(t, "new", "new=true created=true")
}

func TestSchemaDemoNamedConnection(t *testing.T) {
	newApplication(t)
	expectValue(t, "named-connection", "connection=sqlite created=true deferred_error=true")
}

func TestSchemaDemoCreate(t *testing.T) {
	newApplication(t)
	expectValue(t, "create", "created=true idempotent=true unique=true")
}

func TestSchemaDemoCreateValidation(t *testing.T) {
	newApplication(t)
	expectValue(t, "create-validation", "empty_error=true idempotent=true late_column=false")
}

func TestSchemaDemoTable(t *testing.T) {
	newApplication(t)
	expectValue(t, "table", "added=true index=true")
}

func TestSchemaDemoAddColumn(t *testing.T) {
	newApplication(t)
	expectValue(t, "add-column", "added=true idempotent=true columns=2")
}

func TestSchemaDemoRenameColumn(t *testing.T) {
	newApplication(t)
	expectValue(t, "rename-column", "renamed=true source=false")
}

func TestSchemaDemoDropColumn(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-column", "dropped=true ignored=true")
}

func TestSchemaDemoDropColumns(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-columns", "dropped=true remaining=2")
}

func TestSchemaDemoRename(t *testing.T) {
	newApplication(t)
	expectValue(t, "rename", "renamed=true source=false")
}

func TestSchemaDemoDrop(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop", "dropped=true idempotent=true")
}

func TestSchemaDemoBuilderDropColumns(t *testing.T) {
	newApplication(t)
	expectValue(t, "builder-drop-columns", "dropped=true remaining=1")
}

func TestSchemaDemoIncrements(t *testing.T) {
	newApplication(t)
	expectValue(t, "increments", "families=5 primary=true generated=true")
}

func TestSchemaDemoSignedIntegers(t *testing.T) {
	newApplication(t)
	expectValue(t, "signed-integers", "big=integer int=integer medium=integer small=integer tiny=integer")
}

func TestSchemaDemoUnsignedIntegers(t *testing.T) {
	newApplication(t)
	expectValue(t, "unsigned-integers", "big=integer int=integer medium=integer small=integer tiny=integer")
}

func TestSchemaDemoStringAndChar(t *testing.T) {
	newApplication(t)
	expectValue(t, "string-char", "name=text code=text")
}

func TestSchemaDemoTextTypes(t *testing.T) {
	newApplication(t)
	expectValue(t, "text-types", "body=text summary=text notes=text archive=text")
}

func TestSchemaDemoUUIDAndULID(t *testing.T) {
	newApplication(t)
	expectValue(t, "uuid-ulid", "identifier=text sequence=text")
}

func TestSchemaDemoNetworkAddresses(t *testing.T) {
	newApplication(t)
	expectValue(t, "network-addresses", "ip=text mac=text")
}

func TestSchemaDemoRememberToken(t *testing.T) {
	newApplication(t)
	expectValue(t, "remember-token", "remember_token=text nullable=true")
}

func TestSchemaDemoBoolean(t *testing.T) {
	newApplication(t)
	expectValue(t, "boolean", "active=integer archived=integer")
}

func TestSchemaDemoFloatingPoint(t *testing.T) {
	newApplication(t)
	expectValue(t, "floating-point", "ratio=real score=real")
}

func TestSchemaDemoDecimal(t *testing.T) {
	newApplication(t)
	expectValue(t, "decimal", "price=numeric total=numeric")
}

func TestSchemaDemoDate(t *testing.T) {
	newApplication(t)
	expectValue(t, "date", "born_on=datetime")
}

func TestSchemaDemoDateTime(t *testing.T) {
	newApplication(t)
	expectValue(t, "datetime", "seen_at=datetime seen_at_tz=datetime")
}

func TestSchemaDemoTime(t *testing.T) {
	newApplication(t)
	expectValue(t, "time", "starts_at=datetime starts_at_tz=datetime")
}

func TestSchemaDemoTimestamp(t *testing.T) {
	newApplication(t)
	expectValue(t, "timestamp", "published_at=datetime published_at_tz=datetime")
}

func TestSchemaDemoYear(t *testing.T) {
	newApplication(t)
	expectValue(t, "year", "release_year=datetime")
}

func TestSchemaDemoTimestamps(t *testing.T) {
	newApplication(t)
	expectValue(t, "timestamps", "created_at=datetime nullable=true updated_at=datetime nullable=true")
}

func TestSchemaDemoNullableTimestamps(t *testing.T) {
	newApplication(t)
	expectValue(t, "nullable-timestamps", "created_at=datetime nullable=true updated_at=datetime nullable=true")
}

func TestSchemaDemoSoftDeletes(t *testing.T) {
	newApplication(t)
	expectValue(t, "soft-deletes", "deleted_at=datetime nullable=true index=true")
}

func TestSchemaDemoBinary(t *testing.T) {
	newApplication(t)
	expectValue(t, "binary", "payload=blob")
}

func TestSchemaDemoJSON(t *testing.T) {
	newApplication(t)
	expectValue(t, "json", "meta=text meta_b=text")
}

func TestSchemaDemoEnumAndSet(t *testing.T) {
	newApplication(t)
	expectValue(t, "enum-set", "status=text flags=text")
}

func TestSchemaDemoSpatialTypes(t *testing.T) {
	newApplication(t)
	expectValue(t, "spatial-types", "geo=text geog=text pt=text line=text poly=text")
}

func TestSchemaDemoVector(t *testing.T) {
	newApplication(t)
	expectValue(t, "vector", "embedding=text")
}

func TestSchemaDemoForeignID(t *testing.T) {
	newApplication(t)
	expectValue(t, "foreign-id", "owner_id=integer")
}

func TestSchemaDemoForeignIDFor(t *testing.T) {
	newApplication(t)
	expectValue(t, "foreign-id-for", "user_id=integer")
}

func TestSchemaDemoMorphs(t *testing.T) {
	newApplication(t)
	expectValue(t, "morphs", "taggable_id=integer taggable_type=text index=true")
}

func TestSchemaDemoNullableMorphs(t *testing.T) {
	newApplication(t)
	expectValue(t, "nullable-morphs", "taggable_id=integer nullable=true taggable_type=text nullable=true index=true")
}

func TestSchemaDemoNullable(t *testing.T) {
	newApplication(t)
	expectValue(t, "nullable", "nickname_nullable=true bio_nullable=true")
}

func TestSchemaDemoNotNull(t *testing.T) {
	newApplication(t)
	expectValue(t, "not-null", "nickname_nullable=false bio_nullable=false legacy_nullable=false")
}

func TestSchemaDemoID(t *testing.T) {
	newApplication(t)
	expectValue(t, "id", "default=id named=custom_id primary=true")
}

func TestSchemaDemoStoredAs(t *testing.T) {
	expectValue(t, "stored-as", "stored_as=placeholder compiled=false")
}

func TestSchemaDemoVirtualAs(t *testing.T) {
	expectValue(t, "virtual-as", "virtual_as=placeholder compiled=false")
}

func TestSchemaDemoFrom(t *testing.T) {
	expectValue(t, "from", "from=compatibility affects_sql=false")
}

func TestSchemaDemoInstant(t *testing.T) {
	expectValue(t, "instant", "instant=compatibility affects_sql=false")
}

func TestSchemaDemoLock(t *testing.T) {
	expectValue(t, "lock", "lock=compatibility affects_sql=false")
}

func TestSchemaDemoChangeSemantics(t *testing.T) {
	expectValue(t, "change-semantics", "change=modify_column idempotency=caller")
}

func TestSchemaDemoUnsigned(t *testing.T) {
	newApplication(t)
	expectValue(t, "unsigned", "count=integer total=integer")
}

func TestSchemaDemoAutoIncrement(t *testing.T) {
	newApplication(t)
	expectValue(t, "auto-increment", "auto_increment=true generated=true")
}

func TestSchemaDemoPrimaryModifier(t *testing.T) {
	newApplication(t)
	expectValue(t, "primary-modifier", "code_primary=true label_primary=false")
}

func TestSchemaDemoIndexModifier(t *testing.T) {
	newApplication(t)
	expectValue(t, "index-modifier", "index=true removed=true")
}

func TestSchemaDemoDefault(t *testing.T) {
	newApplication(t)
	expectValue(t, "default", "nickname='guest' enabled=1 sort=0 created_at=CURRENT_TIMESTAMP payload=")
}

func TestSchemaDemoDropRememberToken(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-remember-token", "before=true after=false")
}

func TestSchemaDemoDropTimestamps(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-timestamps", "before=true after=false")
}

func TestSchemaDemoDropTimestampsTz(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-timestamps-tz", "before=true after=false")
}

func TestSchemaDemoDropSoftDeletes(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-soft-deletes", "before=true column_after=false index_after=false")
}

func TestSchemaDemoDropSoftDeletesTz(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-soft-deletes-tz", "before=true column_after=false index_after=false")
}

func TestSchemaDemoDropMorphs(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-morphs", "before=true columns_after=false index_after=false")
}

func TestSchemaDemoDropForeignIDFor(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-foreign-id-for", "before=true after=false")
}

func TestSchemaDemoPrimaryIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "primary-index", "primary=true")
}

func TestSchemaDemoUniqueIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "unique-index", "unique=true")
}

func TestSchemaDemoIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "index", "index=true name=schema_demo_index_status_index")
}

func TestSchemaDemoFulltextIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "fulltext-index", "index=true name=schema_demo_fulltext_index_body_fulltext")
}

func TestSchemaDemoSpatialIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "spatial-index", "index=true name=schema_demo_spatial_index_location_spatial")
}

func TestSchemaDemoNamedIndexes(t *testing.T) {
	newApplication(t)
	expectValue(t, "named-indexes", "unique=true index=true")
}

func TestSchemaDemoIndexNaming(t *testing.T) {
	newApplication(t)
	expectValue(t, "index-naming", "short=true trimmed=true")
}

func TestSchemaDemoDropIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-index", "before=true after=false")
}

func TestSchemaDemoDropUnique(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-unique", "before=true after=false")
}

func TestSchemaDemoDropFulltext(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-fulltext", "before=true after=false")
}

func TestSchemaDemoDropSpatialIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "drop-spatial-index", "before=true after=false")
}

func TestSchemaDemoForeignKeyDialect(t *testing.T) {
	newApplication(t)
	expectValue(t, "foreign-dialect", "dialect=sqlite foreign_keys=0")
}

func TestSchemaDemoTableViewExistence(t *testing.T) {
	newApplication(t)
	expectValue(t, "table-view-existence", "table=true view=true")
}

func TestSchemaDemoTables(t *testing.T) {
	newApplication(t)
	expectValue(t, "tables", "found=true schema_set=true filtered=true")
}

func TestSchemaDemoTableListing(t *testing.T) {
	newApplication(t)
	expectValue(t, "table-listing", "qualified=true bare=true")
}

func TestSchemaDemoViews(t *testing.T) {
	newApplication(t)
	expectValue(t, "views", "found=true definition=true")
}

func TestSchemaDemoSchemas(t *testing.T) {
	newApplication(t)
	expectValue(t, "schemas", "count_positive=true name_set=true")
}

func TestSchemaDemoTypes(t *testing.T) {
	newApplication(t)
	expectValue(t, "types", "count=0 empty=true")
}

func TestSchemaDemoSchemaFilter(t *testing.T) {
	newApplication(t)
	expectValue(t, "schema-filter", "string=true slice=true nil=true")
}

func TestSchemaDemoHasColumns(t *testing.T) {
	newApplication(t)
	expectValue(t, "has-columns", "all=true missing=false empty=true")
}

func TestSchemaDemoColumns(t *testing.T) {
	newApplication(t)
	expectValue(t, "columns", "count=3 names=true nullable=true primary=true")
}

func TestSchemaDemoColumnType(t *testing.T) {
	newApplication(t)
	expectValue(t, "column-type", "short=text full=text missing_error=true")
}

func TestSchemaDemoHasIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "has-index", "name=true columns=true type=true wrong_type=false missing=false")
}

func TestSchemaDemoIndexes(t *testing.T) {
	newApplication(t)
	expectValue(t, "indexes", "names=true unique=true plain=true")
}

func TestSchemaDemoMetadataTypes(t *testing.T) {
	expectValue(t, "metadata-types", "schema=SchemaInfo table=TableInfo view=ViewInfo type=TypeInfo column=ColumnInfo index=IndexInfo foreign=ForeignKeyInfo")
}

func TestSchemaDemoWhenHasColumn(t *testing.T) {
	newApplication(t)
	expectValue(t, "when-has-column", "executed=true added=true absent_executed=false")
}

func TestSchemaDemoWhenMissingColumn(t *testing.T) {
	newApplication(t)
	expectValue(t, "when-missing-column", "present_executed=false missing_executed=true added=true")
}

func TestSchemaDemoWhenMissingIndex(t *testing.T) {
	newApplication(t)
	expectValue(t, "when-missing-index", "existing_executed=false missing_executed=true created=true")
}

func TestSchemaDemoForeignKeyToggleDialects(t *testing.T) {
	expectValue(t, "foreign-key-toggle-dialects", "mysql=SET FOREIGN_KEY_CHECKS sqlite=PRAGMA foreign_keys")
}

func TestSchemaDemoEnsureExtension(t *testing.T) {
	expectValue(t, "ensure-extension", "extension=postgis mysql=unsupported sqlite=unsupported")
}

func TestSchemaDemoEnsureVectorExtension(t *testing.T) {
	expectValue(t, "ensure-vector-extension", "extension=vector mysql=unsupported sqlite=unsupported")
}

func TestSchemaDemoSyncModels(t *testing.T) {
	newApplication(t)
	expectValue(t, "sync-models", "created=true columns=2")
}

func TestSchemaDemoSyncModelsColumns(t *testing.T) {
	newApplication(t)
	expectValue(t, "sync-models-columns", "created=true added=true")
}

func TestSchemaDemoSyncModelsDefaults(t *testing.T) {
	newApplication(t)
	expectValue(t, "sync-models-defaults", "default=active options=ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
}

func TestSchemaDemoSyncModelsBoundaries(t *testing.T) {
	expectValue(t, "sync-models-boundaries", "entry=SyncModels creates_tables=true adds_columns=true drops_columns=false auto_migrate=false")
}

func TestSchemaDemoDialectCompatibility(t *testing.T) {
	expectValue(t, "dialect-compatibility", "shared=create,index,drop-index mysql_only=rename-index,drop-primary,create-database,sync-models sqlite_only=sqlite-extension")
}

func TestSchemaDemoLaravelCompatibility(t *testing.T) {
	expectValue(t, "laravel-compatibility", "create=Schema::create table=Schema::table checks=Schema::hasTable columns=Schema::hasColumn indexes=Schema::hasIndex drop=Schema::dropIfExists")
}
