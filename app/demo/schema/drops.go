package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// dropRememberTokenScenario verifies the remember token convention drop.
func dropRememberTokenScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_remember_token"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.RememberToken()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumn(table, "remember_token")
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropRememberToken()
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasColumn(table, "remember_token")), nil
}

// dropTimestampsScenario verifies dropping the created_at and updated_at columns.
func dropTimestampsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_timestamps"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.Timestamps()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumns(table, []string{"created_at", "updated_at"})
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropTimestamps()
	}); err != nil {
		return "", err
	}
	after := builder.HasColumns(table, []string{"created_at", "updated_at"})
	return fmt.Sprintf("before=%t after=%t", before, after), nil
}

// dropTimestampsTzScenario verifies the timezone timestamp alias drop.
func dropTimestampsTzScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_timestamps_tz"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.TimestampsTz()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumns(table, []string{"created_at", "updated_at"})
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropTimestampsTz()
	}); err != nil {
		return "", err
	}
	after := builder.HasColumns(table, []string{"created_at", "updated_at"})
	return fmt.Sprintf("before=%t after=%t", before, after), nil
}

// dropSoftDeletesScenario verifies dropping the soft-delete column and its index.
func dropSoftDeletesScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_soft_deletes"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.SoftDeletes()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumn(table, "deleted_at") && builder.HasIndex(table, []string{"deleted_at"})
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropSoftDeletes()
	}); err != nil {
		return "", err
	}
	afterColumn := builder.HasColumn(table, "deleted_at")
	afterIndex := builder.HasIndex(table, []string{"deleted_at"})
	return fmt.Sprintf("before=%t column_after=%t index_after=%t", before, afterColumn, afterIndex), nil
}

// dropSoftDeletesTzScenario verifies the timezone soft-delete alias drop.
func dropSoftDeletesTzScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_soft_deletes_tz"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.SoftDeletesTz()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumn(table, "deleted_at") && builder.HasIndex(table, []string{"deleted_at"})
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropSoftDeletesTz()
	}); err != nil {
		return "", err
	}
	afterColumn := builder.HasColumn(table, "deleted_at")
	afterIndex := builder.HasIndex(table, []string{"deleted_at"})
	return fmt.Sprintf("before=%t column_after=%t index_after=%t", before, afterColumn, afterIndex), nil
}

// dropMorphsScenario verifies dropping polymorphic columns and their index.
func dropMorphsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_morphs"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.Morphs("owner")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumns(table, []string{"owner_id", "owner_type"}) &&
		builder.HasIndex(table, []string{"owner_id", "owner_type"})
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropMorphs("owner")
	}); err != nil {
		return "", err
	}
	afterColumns := builder.HasColumns(table, []string{"owner_id", "owner_type"})
	afterIndex := builder.HasIndex(table, []string{"owner_id", "owner_type"})
	return fmt.Sprintf("before=%t columns_after=%t index_after=%t", before, afterColumns, afterIndex), nil
}

// dropForeignIDForScenario verifies dropping a foreign ID column without its constraint.
func dropForeignIDForScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_drop_foreign_id_for"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignIdFor("owner_id")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	before := builder.HasColumn(table, "owner_id")
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropForeignIdFor("owner_id")
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, builder.HasColumn(table, "owner_id")), nil
}

// dropConstrainedForeignIDScenario verifies dropping a constrained foreign ID
// together with its named foreign key constraint.
func dropConstrainedForeignIDScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "drop-constrained-foreign-id"); err != nil {
		return "", err
	}
	const parent = "schema_demo_drop_customers"
	const table = "schema_demo_drop_orders"
	builder := dbschema.New(db)
	if err := builder.Create(parent, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(parent) }()
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("customer_id").Constrained(parent)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	foreignKeys, err := builder.GetForeignKeys(table)
	if err != nil {
		return "", err
	}
	foreign := false
	for _, foreignKey := range foreignKeys {
		if foreignKey.ForeignTable == parent {
			foreign = true
		}
	}
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.DropConstrainedForeignId("customer_id")
	}); err != nil {
		return "", err
	}
	foreignKeys, err = builder.GetForeignKeys(table)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("foreign=%t column_after=%t foreign_after=%t",
		foreign, builder.HasColumn(table, "customer_id"), len(foreignKeys) > 0), nil
}
