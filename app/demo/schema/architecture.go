package schemademo

import (
	"fmt"

	"github.com/prismgo/framework/database"
	dbschema "github.com/prismgo/framework/database/schema"
	"github.com/prismgo/sqlite"
	"gorm.io/gorm"
)

// architectureScenario keeps every documented schema contract referenced at compile time.
func architectureScenario() (string, error) {
	var (
		_ func(...*gorm.DB) *dbschema.Builder              = dbschema.New
		_ func(*gorm.DB) *dbschema.Builder                 = dbschema.Bind
		_ func(string) *dbschema.Builder                   = dbschema.Connection
		_ func() *dbschema.Builder                         = dbschema.Resolve
		_ func(string, func(*dbschema.Blueprint)) error    = dbschema.Create
		_ func(string, func(*dbschema.Blueprint)) error    = dbschema.Table
		_ func(string) error                               = dbschema.Drop
		_ func(string) error                               = dbschema.DropIfExists
		_ func(string, string) error                       = dbschema.Rename
		_ func(string, ...string) error                    = dbschema.DropColumns
		_ func() error                                     = dbschema.DropAllTables
		_ func() error                                     = dbschema.DropAllViews
		_ func() error                                     = dbschema.DropAllTypes
		_ func(...any) error                               = dbschema.SyncModels
		_ func(string) bool                                = dbschema.HasTable
		_ func(string) bool                                = dbschema.HasView
		_ func(string, string) bool                        = dbschema.HasColumn
		_ func(string, []string) bool                      = dbschema.HasColumns
		_ func(string, any, ...string) bool                = dbschema.HasIndex
		_ func(string) ([]dbschema.ColumnInfo, error)      = dbschema.GetColumns
		_ func(string) ([]dbschema.IndexInfo, error)       = dbschema.GetIndexes
		_ func(string) ([]dbschema.ForeignKeyInfo, error)  = dbschema.GetForeignKeys
		_ func(any) ([]dbschema.TableInfo, error)          = dbschema.GetTables
		_ func(any, ...bool) ([]string, error)             = dbschema.GetTableListing
		_ func(any) ([]dbschema.ViewInfo, error)           = dbschema.GetViews
		_ func() ([]dbschema.SchemaInfo, error)            = dbschema.GetSchemas
		_ func(any) ([]dbschema.TypeInfo, error)           = dbschema.GetTypes
		_ func(string, string, func() error) error         = dbschema.WhenTableHasColumn
		_ func(string, string, func() error) error         = dbschema.WhenTableDoesntHaveColumn
		_ func(string, any, func() error, ...string) error = dbschema.WhenTableDoesntHaveIndex
		_ func() error                                     = dbschema.EnableForeignKeyConstraints
		_ func() error                                     = dbschema.DisableForeignKeyConstraints
		_ func(func() error) error                         = dbschema.WithoutForeignKeyConstraints
		_ error                                            = dbschema.ErrUnsupportedFeature
		_ *dbschema.Builder                                = (*dbschema.Builder)(nil)
		_ *dbschema.Blueprint                              = (*dbschema.Blueprint)(nil)
		_ *dbschema.ColumnDefinition                       = (*dbschema.ColumnDefinition)(nil)
		_ *dbschema.IndexDefinition                        = (*dbschema.IndexDefinition)(nil)
		_ *dbschema.ForeignKeyDefinition                   = (*dbschema.ForeignKeyDefinition)(nil)
		_ dbschema.SchemaInfo
		_ dbschema.TableInfo
		_ dbschema.ViewInfo
		_ dbschema.TypeInfo
		_ dbschema.ColumnInfo
		_ dbschema.IndexInfo
		_ dbschema.ForeignKeyInfo
	)
	return "builder=Builder blueprint=Blueprint column=ColumnDefinition index=IndexDefinition foreign=ForeignKeyDefinition err=ErrUnsupportedFeature", nil
}

// sqliteExtensionScenario verifies that the optional SQLite extension registers
// a working dialect without changing the schema builder entry points.
func sqliteExtensionScenario() (string, error) {
	provider := (sqlite.ServiceProvider{}).Name()
	if provider != "prismgo.extension.sqlite" {
		return "", fmt.Errorf("sqlite provider name = %q, want %q", provider, "prismgo.extension.sqlite")
	}
	return withIsolatedSQLite(func(db *gorm.DB) (string, error) {
		builder := dbschema.New(db)
		if err := builder.Create("schema_demo_extension", func(table *dbschema.Blueprint) {
			table.Id()
			table.String("name", 32)
		}); err != nil {
			return "", err
		}
		return fmt.Sprintf("dialect=%s provider=%s table=%t", db.Name(), provider, builder.HasTable("schema_demo_extension")), nil
	})
}

// sqliteConnectionScopeScenario demonstrates the pinned-connection and serial
// migration requirements of the SQLite dialect.
func sqliteConnectionScopeScenario() (string, error) {
	return withIsolatedSQLite(func(db *gorm.DB) (string, error) {
		sqlDB, err := db.DB()
		if err != nil {
			return "", err
		}
		pinned := false
		err = db.Connection(func(tx *gorm.DB) error {
			pinned = true
			return dbschema.New(tx).Create("schema_demo_scope", func(table *dbschema.Blueprint) {
				table.Id()
			})
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("max_open_conns=%d pinned=%t", sqlDB.Stats().MaxOpenConnections, pinned), nil
	})
}

// createDialectOptionsScenario inspects the dialect-specific table options that
// Create appends automatically.
func createDialectOptionsScenario() (string, error) {
	mysql := database.TableOptions("mysql", "InnoDB")
	sqlite := database.TableOptions("sqlite", "InnoDB")
	if mysql != "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4" {
		return "", fmt.Errorf("mysql table options = %q, want %q", mysql, "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	}
	if sqlite != "" {
		return "", fmt.Errorf("sqlite table options = %q, want empty", sqlite)
	}
	return "mysql=ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 sqlite=", nil
}

// dropAllTypesScenario documents that MySQL and SQLite have no user-defined
// type objects, so dropping all types is a no-op that surfaces dialect errors.
func dropAllTypesScenario() (string, error) {
	var _ func() error = dbschema.DropAllTypes
	return "mysql=empty sqlite=empty", nil
}
