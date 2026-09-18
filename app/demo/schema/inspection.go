package schemademo

import (
	"fmt"
	"strings"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// stringSliceContains reports whether values contains want.
func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// quoteName quotes an identifier for the MySQL and SQLite dialects, both of
// which accept backtick quoting.
func quoteName(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// tableInfoContains reports whether tables contains one named want.
func tableInfoContains(tables []dbschema.TableInfo, want string) bool {
	for _, table := range tables {
		if table.Name == want {
			return true
		}
	}
	return false
}

// viewInfoContains reports whether views contains one named want.
func viewInfoContains(views []dbschema.ViewInfo, want string) bool {
	for _, view := range views {
		if view.Name == want {
			return true
		}
	}
	return false
}

// tableViewExistenceScenario verifies table and view existence inspection.
func tableViewExistenceScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_view_existence"
	const view = "schema_demo_view_existence_active"
	builder := dbschema.New(db)
	defer func() { _ = builder.Drop(table) }()

	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("status", 16)
	}); err != nil {
		return "", err
	}
	if err := db.Exec("DROP VIEW IF EXISTS " + quoteName(view)).Error; err != nil {
		return "", err
	}
	defer func() { _ = db.Exec("DROP VIEW IF EXISTS " + quoteName(view)).Error }()
	if err := db.Exec("CREATE VIEW " + quoteName(view) + " AS SELECT " + quoteName("id") + ", " + quoteName("status") + " FROM " + quoteName(table)).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("table=%t view=%t", builder.HasTable(table), builder.HasView(view)), nil
}

// tablesScenario verifies table metadata and schema-filtered listing.
func tablesScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_tables"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	tables, err := builder.GetTables(nil)
	if err != nil {
		return "", err
	}
	schema := ""
	for _, item := range tables {
		if item.Name == table {
			schema = item.Schema
		}
	}
	if schema == "" {
		return "", fmt.Errorf("GetTables did not report a schema for %s", table)
	}
	filtered, err := builder.GetTables([]string{schema})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("found=%t schema_set=%t filtered=%t", tableInfoContains(tables, table), true, tableInfoContains(filtered, table)), nil
}

// tableListingScenario verifies schema-qualified and bare table name listing.
func tableListingScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_table_listing"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	tables, err := builder.GetTables(nil)
	if err != nil {
		return "", err
	}
	schema := ""
	for _, item := range tables {
		if item.Name == table {
			schema = item.Schema
		}
	}
	if schema == "" {
		return "", fmt.Errorf("GetTables did not report a schema for %s", table)
	}
	qualified, err := builder.GetTableListing(nil)
	if err != nil {
		return "", err
	}
	bare, err := builder.GetTableListing(nil, false)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("qualified=%t bare=%t",
		stringSliceContains(qualified, schema+"."+table), stringSliceContains(bare, table)), nil
}

// viewsScenario verifies view metadata inspection.
func viewsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_views"
	const view = "schema_demo_views_active"
	builder := dbschema.New(db)
	defer func() { _ = builder.Drop(table) }()

	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("status", 16)
	}); err != nil {
		return "", err
	}
	if err := db.Exec("DROP VIEW IF EXISTS " + quoteName(view)).Error; err != nil {
		return "", err
	}
	defer func() { _ = db.Exec("DROP VIEW IF EXISTS " + quoteName(view)).Error }()
	if err := db.Exec("CREATE VIEW " + quoteName(view) + " AS SELECT " + quoteName("id") + " FROM " + quoteName(table) + " WHERE " + quoteName("status") + " = 'active'").Error; err != nil {
		return "", err
	}
	views, err := builder.GetViews(nil)
	if err != nil {
		return "", err
	}
	definition := false
	for _, item := range views {
		if item.Name == view && strings.TrimSpace(item.Definition) != "" {
			definition = true
		}
	}
	return fmt.Sprintf("found=%t definition=%t", viewInfoContains(views, view), definition), nil
}

// schemasScenario verifies schema metadata inspection.
func schemasScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	schemas, err := builder.GetSchemas()
	if err != nil {
		return "", err
	}
	nameSet := len(schemas) > 0
	for _, item := range schemas {
		if strings.TrimSpace(item.Name) == "" {
			nameSet = false
		}
	}
	return fmt.Sprintf("count_positive=%t name_set=%t", len(schemas) > 0, nameSet), nil
}

// typesScenario verifies user-defined type inspection.
//
// MySQL has no independent user type objects and SQLite has no user-defined
// types at all, so both dialects report an empty list instead of an error.
func typesScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	types, err := builder.GetTypes(nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("count=%d empty=%t", len(types), len(types) == 0), nil
}

// schemaFilterScenario verifies the string, slice and nil schema filter forms.
func schemaFilterScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_schema_filter"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	tables, err := builder.GetTables(nil)
	if err != nil {
		return "", err
	}
	schema := ""
	for _, item := range tables {
		if item.Name == table {
			schema = item.Schema
		}
	}
	if schema == "" {
		return "", fmt.Errorf("GetTables did not report a schema for %s", table)
	}
	byString, err := builder.GetTables(schema)
	if err != nil {
		return "", err
	}
	bySlice, err := builder.GetTables([]string{schema})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("string=%t slice=%t nil=%t",
		tableInfoContains(byString, table), tableInfoContains(bySlice, table), tableInfoContains(tables, table)), nil
}

// hasColumnsScenario verifies single and multiple column existence inspection.
func hasColumnsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_has_columns"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("code", 16)
		blueprint.Integer("level")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	all := builder.HasColumns(table, []string{"code", "level"})
	missing := builder.HasColumns(table, []string{"code", "missing"})
	empty := builder.HasColumns(table, nil)
	return fmt.Sprintf("all=%t missing=%t empty=%t", all, missing, empty), nil
}

// columnsScenario verifies column metadata and name listing.
func columnsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_columns"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("nickname", 64).Nullable()
		blueprint.Integer("level").Default(1)
	}, func(builder *dbschema.Builder) error {
		columns, err := builder.GetColumns(table)
		if err != nil {
			return err
		}
		listing, err := builder.GetColumnListing(table)
		if err != nil {
			return err
		}
		facts, err := columnFacts(builder, table, "id", "nickname", "level")
		if err != nil {
			return err
		}
		named := len(listing) == 3 &&
			stringSliceContains(listing, "id") && stringSliceContains(listing, "nickname") && stringSliceContains(listing, "level")
		value = fmt.Sprintf("count=%d names=%t nullable=%t primary=%t", len(columns), named, facts["nickname"].Nullable, facts["id"].Primary)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// columnTypeScenario verifies short and full column type inspection.
func columnTypeScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_column_type"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("name", 64)
	}, func(builder *dbschema.Builder) error {
		short, err := builder.GetColumnType(table, "name")
		if err != nil {
			return err
		}
		full, err := builder.GetColumnType(table, "name", true)
		if err != nil {
			return err
		}
		missing, missingErr := builder.GetColumnType(table, "missing")
		if missingErr == nil {
			return fmt.Errorf("GetColumnType(%s.missing) = %q, want an error", table, missing)
		}
		value = fmt.Sprintf("short=%s full=%s missing_error=true", short, full)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// hasIndexScenario verifies index existence by name, columns, and kind.
func hasIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_has_index"
	const plain = "idx_schema_demo_has_index_status"
	const unique = "uix_schema_demo_has_index_email"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 128)
		blueprint.String("status", 16)
		blueprint.UniqueNamed(unique, "email")
		blueprint.IndexNamed(plain, "status")
	}, func(builder *dbschema.Builder) error {
		byName := builder.HasIndex(table, plain)
		byColumns := builder.HasIndex(table, []string{"status"})
		byType := builder.HasIndex(table, []string{"email"}, "unique")
		wrongType := builder.HasIndex(table, []string{"email"}, "index")
		missing := builder.HasIndex(table, []string{"missing"})
		value = fmt.Sprintf("name=%t columns=%t type=%t wrong_type=%t missing=%t", byName, byColumns, byType, wrongType, missing)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// indexesScenario verifies index metadata and name listing.
func indexesScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_indexes"
	const unique = "uix_schema_demo_indexes_email"
	const plain = "idx_schema_demo_indexes_status"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 128)
		blueprint.String("status", 16)
		blueprint.UniqueNamed(unique, "email")
		blueprint.IndexNamed(plain, "status")
	}, func(builder *dbschema.Builder) error {
		indexes, err := builder.GetIndexes(table)
		if err != nil {
			return err
		}
		names, err := builder.GetIndexListing(table)
		if err != nil {
			return err
		}
		uniqueKind := false
		plainKind := false
		for _, index := range indexes {
			switch index.Name {
			case unique:
				uniqueKind = index.Unique && index.Type == "unique" && len(index.Columns) == 1 && index.Columns[0] == "email"
			case plain:
				plainKind = !index.Unique && index.Type == "index" && len(index.Columns) == 1 && index.Columns[0] == "status"
			}
		}
		value = fmt.Sprintf("names=%t unique=%t plain=%t",
			stringSliceContains(names, unique) && stringSliceContains(names, plain), uniqueKind, plainKind)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// foreignKeysScenario verifies foreign key metadata inspection.
//
// SQLite reports foreign keys through PRAGMA but the framework treats the
// inspection as MySQL-only, so this scenario requires a real MySQL server.
func foreignKeysScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "foreign-keys"); err != nil {
		return "", err
	}
	const parent = "schema_demo_fk_inspect_parents"
	const child = "schema_demo_fk_inspect"
	foreignKey, found, cleanup, err := foreignKeyFixture(db, parent, child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("parent_id").Nullable()
		blueprint.Foreign("parent_id").References("id").On(parent).CascadeOnUpdate().CascadeOnDelete()
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	return fmt.Sprintf("columns=%d foreign=%s on_update=%s on_delete=%s",
		len(foreignKey.Columns), foreignKey.ForeignTable, foreignKey.OnUpdate, foreignKey.OnDelete), nil
}
