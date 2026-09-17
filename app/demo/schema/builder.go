package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// bindScenario verifies schema.Bind binds a migration connection and chains.
func bindScenario(db *gorm.DB) (string, error) {
	builder := dbschema.Bind(db)
	if builder.Bind(db) != builder {
		return "", fmt.Errorf("Builder.Bind reused instance = false, want true")
	}
	if err := builder.Create("schema_demo_bind", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_bind") }()
	return "bound=true chained=true", nil
}

// newScenario verifies the standalone New constructor.
func newScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_new", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_new") }()
	return fmt.Sprintf("new=true created=%t", builder.HasTable("schema_demo_new")), nil
}

// namedConnectionScenario verifies named connection selection and deferred errors.
//
// It selects the same connection the application uses as default, so hermetic
// tests exercise the configured SQLite connection while the Demo command reuses
// the configured MySQL connection instead of writing a stray SQLite file.
func namedConnectionScenario(db *gorm.DB) (string, error) {
	name := db.Name()
	builder := dbschema.Connection(name)
	if builder == nil {
		return "", fmt.Errorf("schema.Connection(%q) returned a nil builder", name)
	}
	defer func() { _ = builder.Close() }()
	if err := builder.Create("schema_demo_named", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	created := builder.HasTable("schema_demo_named")
	_ = builder.Drop("schema_demo_named")

	deferred := dbschema.Connection("schema_demo_missing").Create("schema_demo_missing", func(table *dbschema.Blueprint) {
		table.Id()
	}) != nil
	return fmt.Sprintf("connection=%s created=%t deferred_error=%t", name, created, deferred), nil
}

// facadeScenario verifies the package-level facade delegates to the container builder.
func facadeScenario() (string, error) {
	if err := dbschema.Create("schema_demo_facade", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("name", 32)
	}); err != nil {
		return "", err
	}
	hasTable := dbschema.HasTable("schema_demo_facade")
	hasColumn := dbschema.HasColumn("schema_demo_facade", "name")
	if err := dbschema.Drop("schema_demo_facade"); err != nil {
		return "", err
	}
	return fmt.Sprintf("create=true has_table=%t has_column=%t", hasTable, hasColumn), nil
}

// createScenario verifies declarative table creation and idempotency.
func createScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_create", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("name", 100)
		table.String("code", 32)
		table.TinyInteger("status").Default(1)
		table.Timestamps()
		table.SoftDeletes()
		table.UniqueNamed("uix_schema_demo_create_code", "code")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_create") }()

	if err := builder.Create("schema_demo_create", func(table *dbschema.Blueprint) {
		table.String("late", 32)
	}); err != nil {
		return "", err
	}
	idempotent := !builder.HasColumn("schema_demo_create", "late")
	return fmt.Sprintf("created=true idempotent=%t unique=%t", idempotent, builder.HasIndex("schema_demo_create", "uix_schema_demo_create_code")), nil
}

// createValidationScenario verifies empty blueprints fail and existing tables are skipped.
func createValidationScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	emptyErr := builder.Create("schema_demo_empty", func(*dbschema.Blueprint) {})
	if emptyErr == nil {
		return "", fmt.Errorf("Create with an empty blueprint error = nil, want a validation error")
	}
	if err := builder.Create("schema_demo_validation", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_validation") }()

	if err := builder.Create("schema_demo_validation", func(table *dbschema.Blueprint) {
		table.String("late", 32)
	}); err != nil {
		return "", err
	}
	late := builder.HasColumn("schema_demo_validation", "late")
	return fmt.Sprintf("empty_error=true idempotent=%t late_column=%t", !late, late), nil
}

// tableScenario verifies altering an existing table with a column and index.
func tableScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_table", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("name", 64)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_table") }()

	if err := builder.Table("schema_demo_table", func(table *dbschema.Blueprint) {
		table.String("email", 128).Nullable().After("name")
		table.IndexNamed("idx_schema_demo_table_email", "email")
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("added=%t index=%t", builder.HasColumn("schema_demo_table", "email"), builder.HasIndex("schema_demo_table", "idx_schema_demo_table_email")), nil
}

// addColumnScenario verifies idempotent column addition.
func addColumnScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_add_column", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_add_column") }()

	add := func() error {
		return builder.Table("schema_demo_add_column", func(table *dbschema.Blueprint) {
			table.String("code", 32)
		})
	}
	if err := add(); err != nil {
		return "", err
	}
	if err := add(); err != nil {
		return "", err
	}
	columns, err := builder.GetColumnListing("schema_demo_add_column")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("added=true idempotent=%t columns=%d", len(columns) == 2, len(columns)), nil
}

// changeColumnScenario verifies Change compiles to MODIFY COLUMN on MySQL.
func changeColumnScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "change-column"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_change_column", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("phone", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_change_column") }()

	if err := builder.Table("schema_demo_change_column", func(table *dbschema.Blueprint) {
		table.String("phone", 64).Nullable().Change()
	}); err != nil {
		return "", err
	}
	phoneType, err := builder.GetColumnType("schema_demo_change_column", "phone", true)
	if err != nil {
		return "", err
	}
	columns, err := builder.GetColumns("schema_demo_change_column")
	if err != nil {
		return "", err
	}
	nullable := false
	for _, column := range columns {
		if column.Name == "phone" {
			nullable = column.Nullable
		}
	}
	return fmt.Sprintf("type=%s nullable=%t", phoneType, nullable), nil
}

// renameColumnScenario verifies idempotent column renaming.
func renameColumnScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_rename_column", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("name", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_rename_column") }()

	rename := func() error {
		return builder.Table("schema_demo_rename_column", func(table *dbschema.Blueprint) {
			table.RenameColumn("name", "title")
		})
	}
	if err := rename(); err != nil {
		return "", err
	}
	if err := rename(); err != nil {
		return "", err
	}
	return fmt.Sprintf("renamed=%t source=%t", builder.HasColumn("schema_demo_rename_column", "title"), builder.HasColumn("schema_demo_rename_column", "name")), nil
}

// dropColumnScenario verifies variadic column dropping ignores missing columns.
func dropColumnScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_drop_column", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("legacy_code", 32)
		table.String("legacy_level", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_drop_column") }()

	if err := builder.Table("schema_demo_drop_column", func(table *dbschema.Blueprint) {
		table.DropColumn("legacy_code")
		table.DropColumn("missing_column")
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("dropped=%t ignored=%t", !builder.HasColumn("schema_demo_drop_column", "legacy_code"), builder.HasColumn("schema_demo_drop_column", "legacy_level")), nil
}

// dropColumnsScenario verifies slice-based column dropping.
func dropColumnsScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_drop_columns", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("old_name", 32)
		table.String("old_phone", 32)
		table.String("kept", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_drop_columns") }()

	if err := builder.Table("schema_demo_drop_columns", func(table *dbschema.Blueprint) {
		table.DropColumns([]string{"old_name", "old_phone"})
	}); err != nil {
		return "", err
	}
	columns, err := builder.GetColumnListing("schema_demo_drop_columns")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("dropped=true remaining=%d", len(columns)), nil
}

// rawScenario verifies raw structural SQL commands on MySQL.
func rawScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "raw"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_raw", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("name", 100)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_raw") }()

	if err := builder.Table("schema_demo_raw", func(table *dbschema.Blueprint) {
		table.Raw("ALTER TABLE `schema_demo_raw` ADD COLUMN `search_name` varchar(100) GENERATED ALWAYS AS (`name`) STORED")
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("raw=true generated=%t", builder.HasColumn("schema_demo_raw", "search_name")), nil
}

// renameScenario verifies idempotent table renaming.
func renameScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_rename_from", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_rename_to") }()

	if err := builder.Rename("schema_demo_rename_from", "schema_demo_rename_to"); err != nil {
		return "", err
	}
	if err := builder.Rename("schema_demo_rename_from", "schema_demo_rename_to"); err != nil {
		return "", err
	}
	return fmt.Sprintf("renamed=%t source=%t", builder.HasTable("schema_demo_rename_to"), builder.HasTable("schema_demo_rename_from")), nil
}

// dropScenario verifies Drop and DropIfExists idempotency.
func dropScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_drop_me", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	if err := builder.Drop("schema_demo_drop_me"); err != nil {
		return "", err
	}
	if err := builder.DropIfExists("schema_demo_drop_me"); err != nil {
		return "", err
	}
	return fmt.Sprintf("dropped=%t idempotent=true", !builder.HasTable("schema_demo_drop_me")), nil
}

// builderDropColumnsScenario verifies Builder-level column dropping.
func builderDropColumnsScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_builder_drop", func(table *dbschema.Blueprint) {
		table.Id()
		table.String("legacy_code", 32)
		table.String("legacy_level", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_builder_drop") }()

	if err := builder.DropColumns("schema_demo_builder_drop", "legacy_code", "legacy_level"); err != nil {
		return "", err
	}
	columns, err := builder.GetColumnListing("schema_demo_builder_drop")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("dropped=true remaining=%d", len(columns)), nil
}

// incrementsScenario verifies integer auto-increment primary key families.
//
// Each family member is created on its own table with one generated row so the
// auto-increment values are exercised on both dialects before cleanup.
func incrementsScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_increments_pk", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_increments_pk") }()
	primary, err := incrementPrimary(db, "schema_demo_increments_pk", "id")
	if err != nil {
		return "", err
	}
	for _, name := range []string{"int_id", "big_id", "tiny_id", "small_id", "medium_id"} {
		table := "schema_demo_increments_" + name
		if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
			blueprint.Increments(name)
		}); err != nil {
			return "", err
		}
		if err := db.Exec(fmt.Sprintf("INSERT INTO %s (%s) VALUES (NULL)", table, name)).Error; err != nil {
			_ = builder.Drop(table)
			return "", err
		}
		var generated int64
		if err := db.Raw(fmt.Sprintf("SELECT %s FROM %s LIMIT 1", name, table)).Scan(&generated).Error; err != nil {
			_ = builder.Drop(table)
			return "", err
		}
		if err := builder.Drop(table); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("families=5 primary=%t generated=true", primary), nil
}

// incrementPrimary reports whether the given auto-increment column is primary.
func incrementPrimary(db *gorm.DB, table, column string) (bool, error) {
	columns, err := dbschema.New(db).GetColumns(table)
	if err != nil {
		return false, err
	}
	for _, item := range columns {
		if item.Name == column {
			return item.Primary, nil
		}
	}
	return false, fmt.Errorf("column %s.%s not found", table, column)
}

// idScenario verifies default and named primary key columns.
func idScenario(db *gorm.DB) (string, error) {
	builder := dbschema.New(db)
	if err := builder.Create("schema_demo_id_default", func(table *dbschema.Blueprint) {
		table.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_id_default") }()
	if err := builder.Create("schema_demo_id_named", func(table *dbschema.Blueprint) {
		table.Id("custom_id")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop("schema_demo_id_named") }()

	primary := func(table, column string) (bool, error) {
		columns, err := builder.GetColumns(table)
		if err != nil {
			return false, err
		}
		for _, item := range columns {
			if item.Name == column {
				return item.Primary, nil
			}
		}
		return false, fmt.Errorf("column %s.%s not found", table, column)
	}
	defaultPrimary, err := primary("schema_demo_id_default", "id")
	if err != nil {
		return "", err
	}
	namedPrimary, err := primary("schema_demo_id_named", "custom_id")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("default=id named=custom_id primary=%t", defaultPrimary && namedPrimary), nil
}
