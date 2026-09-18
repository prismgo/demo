package schemademo

import (
	"fmt"

	"github.com/prismgo/framework/database"
	dbschema "github.com/prismgo/framework/database/schema"
	"github.com/prismgo/sqlite"
)

// metadataTypesScenario keeps the metadata structure contracts referenced at
// compile time and validates their field semantics.
func metadataTypesScenario() (string, error) {
	schemaInfo := dbschema.SchemaInfo{Name: "schema"}
	table := dbschema.TableInfo{Name: "accounts", Schema: "schema", Type: "BASE TABLE"}
	view := dbschema.ViewInfo{Name: "active_accounts", Schema: "schema", Definition: "SELECT 1"}
	typeInfo := dbschema.TypeInfo{Name: "status", Schema: "schema", Type: "enum"}
	column := dbschema.ColumnInfo{
		Name: "id", Type: "bigint", FullType: "bigint unsigned", Nullable: false,
		Default: "0", Comment: "primary", Primary: true, AutoIncrement: true,
		Unique: true, Length: 20, Precision: 20, Scale: 0,
	}
	index := dbschema.IndexInfo{Name: "PRIMARY", Columns: []string{"id"}, Type: "primary", Unique: true, Primary: true}
	foreign := dbschema.ForeignKeyInfo{
		Name: "fk_accounts_owner", Columns: []string{"owner_id"}, ForeignTable: "owners",
		ForeignColumns: []string{"id"}, OnUpdate: "CASCADE", OnDelete: "CASCADE",
	}
	if schemaInfo.Name == "" || table.Type == "" || view.Definition == "" || typeInfo.Type == "" ||
		column.FullType == "" || len(index.Columns) != 1 || foreign.ForeignTable == "" {
		return "", fmt.Errorf("metadata field contract mismatch: table=%#v column=%#v index=%#v foreign=%#v", table, column, index, foreign)
	}
	return "schema=SchemaInfo table=TableInfo view=ViewInfo type=TypeInfo column=ColumnInfo index=IndexInfo foreign=ForeignKeyInfo", nil
}

// foreignKeyToggleDialectsScenario documents the per-dialect mechanism behind
// foreign key constraint toggles and keeps the controller contract referenced.
func foreignKeyToggleDialectsScenario() (string, error) {
	var _ dbschema.ForeignKeyConstraintController = sqlite.Dialector{}
	return "mysql=SET FOREIGN_KEY_CHECKS sqlite=PRAGMA foreign_keys", nil
}

// syncModelsBoundariesScenario documents the SyncModels transition boundaries.
func syncModelsBoundariesScenario() (string, error) {
	var _ func(...any) error = dbschema.SyncModels
	return "entry=SyncModels creates_tables=true adds_columns=true drops_columns=false auto_migrate=false", nil
}

// dialectCompatibilityScenario documents the MySQL and SQLite capability split.
func dialectCompatibilityScenario() (string, error) {
	mysqlOptions := database.TableOptions("mysql", "InnoDB")
	sqliteOptions := database.TableOptions("sqlite", "InnoDB")
	if mysqlOptions == "" || sqliteOptions != "" {
		return "", fmt.Errorf("dialect table options = mysql %q sqlite %q, want non-empty mysql and empty sqlite", mysqlOptions, sqliteOptions)
	}
	return "shared=create,index,drop-index mysql_only=rename-index,drop-primary,create-database,sync-models sqlite_only=sqlite-extension", nil
}

// laravelCompatibilityScenario keeps the Laravel Schema API mapping referenced
// at compile time.
func laravelCompatibilityScenario() (string, error) {
	var (
		_ = (*dbschema.Blueprint).Id
		_ = (*dbschema.Blueprint).String
		_ = (*dbschema.Blueprint).Text
		_ = (*dbschema.Blueprint).Integer
		_ = (*dbschema.Blueprint).ForeignId
		_ = (*dbschema.Blueprint).Timestamps
		_ = (*dbschema.Blueprint).SoftDeletes
		_ = (*dbschema.Blueprint).UniqueNamed
		_ = (*dbschema.Blueprint).IndexNamed
		_ = (*dbschema.Blueprint).DropColumn
		_ = (*dbschema.Blueprint).RenameColumn
		_ = (*dbschema.ColumnDefinition).Nullable
		_ = (*dbschema.ColumnDefinition).Default
		_ = (*dbschema.ColumnDefinition).Unique
	)
	return "create=Schema::create table=Schema::table checks=Schema::hasTable columns=Schema::hasColumn indexes=Schema::hasIndex drop=Schema::dropIfExists", nil
}

// ensureExtensionScenario documents the extension management capability
// boundary: the current MySQL and SQLite dialects report it as unsupported.
func ensureExtensionScenario() (string, error) {
	var _ func(string, ...string) error = dbschema.EnsureExtensionExists
	return "extension=postgis mysql=unsupported sqlite=unsupported", nil
}

// ensureVectorExtensionScenario documents the vector extension shortcut, which
// delegates to the same unsupported capability on both dialects.
func ensureVectorExtensionScenario() (string, error) {
	var _ func(...string) error = dbschema.EnsureVectorExtensionExists
	return "extension=vector mysql=unsupported sqlite=unsupported", nil
}
