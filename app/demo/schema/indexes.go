package schemademo

import (
	"fmt"
	"strings"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// primaryIndexScenario verifies explicit composite primary index creation.
//
// A composite key is used because SQLite only inlines a single-column primary
// key, so a composite index stays observable through the index metadata API.
func primaryIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_primary_index"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("tenant_id", 16)
		blueprint.String("code", 16)
		blueprint.Primary("tenant_id", "code")
	}, func(builder *dbschema.Builder) error {
		value = fmt.Sprintf("primary=%t", builder.HasIndex(table, []string{"tenant_id", "code"}, "primary"))
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// uniqueIndexScenario verifies explicit named unique index creation.
func uniqueIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_unique_index"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("tenant_id", 16)
		blueprint.String("code", 16)
		blueprint.Unique("tenant_id", "code")
	}, func(builder *dbschema.Builder) error {
		value = fmt.Sprintf("unique=%t", builder.HasIndex(table, []string{"tenant_id", "code"}, "unique"))
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// plainIndexScenario verifies plain index creation with the default name.
func plainIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_index"
	want := table + "_status_index"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("status", 16)
		blueprint.Index("status")
	}, func(builder *dbschema.Builder) error {
		names, err := builder.GetIndexListing(table)
		if err != nil {
			return err
		}
		value = fmt.Sprintf("index=%t name=%s", stringSliceContains(names, want), want)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// fulltextIndexScenario verifies full-text index creation.
//
// SQLite has no full-text index kind in its schema compiler, so the declaration
// downgrades to a plain index that keeps the Laravel-style default name.
func fulltextIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_fulltext_index"
	want := table + "_body_fulltext"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.Text("body")
		blueprint.FullText("body")
	}, func(builder *dbschema.Builder) error {
		names, err := builder.GetIndexListing(table)
		if err != nil {
			return err
		}
		value = fmt.Sprintf("index=%t name=%s", stringSliceContains(names, want), want)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// spatialIndexScenario verifies spatial index creation.
//
// MySQL requires a NOT NULL geometry column for a SPATIAL INDEX; SQLite stores
// geometry as text and downgrades the declaration to a plain index.
func spatialIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_spatial_index"
	want := table + "_location_spatial"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.Geometry("location").NotNull()
		blueprint.SpatialIndex("location")
	}, func(builder *dbschema.Builder) error {
		names, err := builder.GetIndexListing(table)
		if err != nil {
			return err
		}
		value = fmt.Sprintf("index=%t name=%s", stringSliceContains(names, want), want)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// namedIndexesScenario verifies explicitly named unique and plain indexes.
func namedIndexesScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_named_indexes"
	const unique = "uix_schema_demo_named_indexes_email"
	const plain = "idx_schema_demo_named_indexes_status"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 64)
		blueprint.String("status", 16)
		blueprint.UniqueNamed(unique, "email")
		blueprint.IndexNamed(plain, "status")
	}, func(builder *dbschema.Builder) error {
		value = fmt.Sprintf("unique=%t index=%t", builder.HasIndex(table, unique), builder.HasIndex(table, plain))
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// indexNamingScenario verifies the Laravel-style default index name and the
// MySQL-compatible length trimming with a SHA1 suffix.
//
// The table name is long enough that the second index name exceeds MySQL's
// 64-character identifier limit while the first one stays under it.
func indexNamingScenario(db *gorm.DB) (string, error) {
	table := "schema_demo_index_naming_" + strings.Repeat("t", 15)
	longColumn := strings.Repeat("x", 30)
	shortWant := table + "_code_index"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("code", 16)
		blueprint.String(longColumn, 16)
		blueprint.Index("code")
		blueprint.Index(longColumn)
	}, func(builder *dbschema.Builder) error {
		names, err := builder.GetIndexListing(table)
		if err != nil {
			return err
		}
		trimmed := false
		for _, name := range names {
			if name != shortWant && len(name) == 64 {
				trimmed = true
			}
		}
		value = fmt.Sprintf("short=%t trimmed=%t", stringSliceContains(names, shortWant), trimmed)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// renameIndexScenario verifies index renaming, which MySQL supports natively.
func renameIndexScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "rename-index"); err != nil {
		return "", err
	}
	const table = "schema_demo_rename_index"
	const from = "idx_schema_demo_rename_index_email"
	const to = "idx_schema_demo_rename_index_contact"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 64)
		blueprint.IndexNamed(from, "email")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.RenameIndex(from, to)
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("renamed=%t source=%t", builder.HasIndex(table, to), builder.HasIndex(table, from)), nil
}

// dropIndexScenario verifies removing a plain index by name.
func dropIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_index"
	const name = "idx_schema_demo_drop_index_email"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 64)
		blueprint.IndexNamed(name, "email")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasIndex(table, name)
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropIndex(name)
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasIndex(table, name)), nil
}

// dropUniqueScenario verifies removing a unique index by name.
func dropUniqueScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_unique"
	const name = "uix_schema_demo_drop_unique_email"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 64)
		blueprint.UniqueNamed(name, "email")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasIndex(table, name)
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropUnique(name)
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasIndex(table, name)), nil
}

// dropPrimaryScenario verifies dropping a composite primary key.
//
// SQLite cannot drop a primary key from an existing table, so this scenario
// requires MySQL.
func dropPrimaryScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "drop-primary"); err != nil {
		return "", err
	}
	const table = "schema_demo_drop_primary"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("tenant_id", 16)
		blueprint.String("code", 16)
		blueprint.Primary("tenant_id", "code")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasIndex(table, []string{"tenant_id", "code"}, "primary")
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropPrimary()
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasIndex(table, []string{"tenant_id", "code"}, "primary")), nil
}

// dropFulltextScenario verifies removing a full-text index by name.
func dropFulltextScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_fulltext"
	name := table + "_body_fulltext"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.Text("body")
		blueprint.FullText("body")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasIndex(table, name)
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropFullText(name)
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasIndex(table, name)), nil
}

// dropSpatialIndexScenario verifies removing a spatial index by name.
func dropSpatialIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_spatial_index"
	name := table + "_location_spatial"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.Geometry("location").NotNull()
		blueprint.SpatialIndex("location")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasIndex(table, name)
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropSpatialIndex(name)
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasIndex(table, name)), nil
}
