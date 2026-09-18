package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// findForeignKey returns the first foreign key on table that references
// foreignTable.
func findForeignKey(builder *dbschema.Builder, table, foreignTable string) (dbschema.ForeignKeyInfo, bool, error) {
	foreignKeys, err := builder.GetForeignKeys(table)
	if err != nil {
		return dbschema.ForeignKeyInfo{}, false, err
	}
	for _, foreignKey := range foreignKeys {
		if foreignKey.ForeignTable == foreignTable {
			return foreignKey, true, nil
		}
	}
	return dbschema.ForeignKeyInfo{}, false, nil
}

// foreignKeyFixture creates a parent table and a child table carrying one
// foreign key, then reports the observed constraint. The returned cleanup drops
// the child before the parent and is safe to call after a partial failure.
func foreignKeyFixture(db *gorm.DB, parent, child string, declare func(*dbschema.Blueprint)) (dbschema.ForeignKeyInfo, bool, func(), error) {
	builder := dbschema.New(db)
	if err := builder.Create(parent, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return dbschema.ForeignKeyInfo{}, false, nil, err
	}
	cleanup := func() { _ = builder.Drop(parent) }
	if err := builder.Create(child, declare); err != nil {
		return dbschema.ForeignKeyInfo{}, false, cleanup, err
	}
	cleanup = func() {
		_ = builder.Drop(child)
		_ = builder.Drop(parent)
	}
	foreignKey, found, err := findForeignKey(builder, child, parent)
	if err != nil {
		return dbschema.ForeignKeyInfo{}, false, cleanup, err
	}
	return foreignKey, found, cleanup, nil
}

// requireForeignKey fails when the child table has no foreign key referencing
// the expected target, so scenarios never report success on a missing key.
func requireForeignKey(child, parent string, found bool) error {
	if !found {
		return fmt.Errorf("foreign key on %s referencing %s not found", child, parent)
	}
	return nil
}

// constrainedScenario verifies the convention-based constrained foreign ID.
func constrainedScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "constrained"); err != nil {
		return "", err
	}
	const parent = "schema_demo_parents"
	const child = "schema_demo_fk_constrained"
	foreignKey, found, cleanup, err := foreignKeyFixture(db, parent, child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("schema_demo_parent_id").Constrained()
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	if len(foreignKey.Columns) != 1 || foreignKey.Columns[0] != "schema_demo_parent_id" || len(foreignKey.ForeignColumns) != 1 || foreignKey.ForeignColumns[0] != "id" {
		return "", fmt.Errorf("constrained foreign key columns = %v -> %v, want [schema_demo_parent_id] -> [id]", foreignKey.Columns, foreignKey.ForeignColumns)
	}
	return fmt.Sprintf("table=%s column=%s", parent, foreignKey.ForeignColumns[0]), nil
}

// constrainedExplicitScenario verifies Constrained with an explicit target.
func constrainedExplicitScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "constrained-explicit"); err != nil {
		return "", err
	}
	const parent = "schema_demo_fk_explicit_owners"
	const child = "schema_demo_fk_explicit"
	foreignKey, found, cleanup, err := foreignKeyFixture(db, parent, child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("owner_id").Constrained(parent)
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	return fmt.Sprintf("table=%s column=%s", parent, foreignKey.ForeignColumns[0]), nil
}

// foreignScenario verifies a manually declared foreign key.
func foreignScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "foreign"); err != nil {
		return "", err
	}
	const parent = "schema_demo_fk_manual_tenants"
	const child = "schema_demo_fk_manual"
	foreignKey, found, cleanup, err := foreignKeyFixture(db, parent, child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("tenant_id")
		blueprint.Foreign("tenant_id").References("id").On(parent)
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	return fmt.Sprintf("table=%s column=%s", parent, foreignKey.ForeignColumns[0]), nil
}

// foreignActionsScenario verifies custom ON UPDATE and ON DELETE actions.
func foreignActionsScenario(db *gorm.DB) (string, error) {
	return actionForeignKeyScenario(db, "foreign-actions", func(foreign *dbschema.ForeignKeyDefinition) *dbschema.ForeignKeyDefinition {
		return foreign.OnUpdate("SET NULL").OnDelete("SET NULL")
	}, "on_update=SET NULL on_delete=SET NULL")
}

// cascadeActionsScenario verifies the cascade action shortcuts.
func cascadeActionsScenario(db *gorm.DB) (string, error) {
	return actionForeignKeyScenario(db, "cascade-actions", func(foreign *dbschema.ForeignKeyDefinition) *dbschema.ForeignKeyDefinition {
		return foreign.CascadeOnUpdate().CascadeOnDelete()
	}, "on_update=CASCADE on_delete=CASCADE")
}

// restrictActionsScenario verifies the restrict action shortcuts.
func restrictActionsScenario(db *gorm.DB) (string, error) {
	return actionForeignKeyScenario(db, "restrict-actions", func(foreign *dbschema.ForeignKeyDefinition) *dbschema.ForeignKeyDefinition {
		return foreign.RestrictOnUpdate().RestrictOnDelete()
	}, "on_update=RESTRICT on_delete=RESTRICT")
}

// nullActionsScenario verifies the set-null action shortcuts.
func nullActionsScenario(db *gorm.DB) (string, error) {
	return actionForeignKeyScenario(db, "null-actions", func(foreign *dbschema.ForeignKeyDefinition) *dbschema.ForeignKeyDefinition {
		return foreign.NullOnUpdate().NullOnDelete()
	}, "on_update=SET NULL on_delete=SET NULL")
}

// noActionActionsScenario verifies the no-action action shortcuts.
func noActionActionsScenario(db *gorm.DB) (string, error) {
	return actionForeignKeyScenario(db, "no-action-actions", func(foreign *dbschema.ForeignKeyDefinition) *dbschema.ForeignKeyDefinition {
		return foreign.NoActionOnUpdate().NoActionOnDelete()
	}, "on_update=NO ACTION on_delete=NO ACTION")
}

// actionForeignKeyScenario runs one foreign key action case and asserts the
// action rules stored by the server.
func actionForeignKeyScenario(db *gorm.DB, name string, configure func(*dbschema.ForeignKeyDefinition) *dbschema.ForeignKeyDefinition, want string) (string, error) {
	if err := requireMySQL(db, name); err != nil {
		return "", err
	}
	const parent = "schema_demo_fk_actions_tenants"
	const child = "schema_demo_fk_actions"
	foreignKey, found, cleanup, err := foreignKeyFixture(db, parent, child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("tenant_id").Nullable()
		configure(blueprint.Foreign("tenant_id").References("id").On(parent))
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	got := fmt.Sprintf("on_update=%s on_delete=%s", foreignKey.OnUpdate, foreignKey.OnDelete)
	if got != want {
		return "", fmt.Errorf("foreign key actions = %q, want %q", got, want)
	}
	return got, nil
}

// foreignNameScenario verifies a custom foreign key constraint name.
func foreignNameScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "foreign-name"); err != nil {
		return "", err
	}
	const parent = "schema_demo_fk_name_tenants"
	const child = "schema_demo_fk_name"
	const name = "fk_schema_demo_fk_name_owner"
	foreignKey, found, cleanup, err := foreignKeyFixture(db, parent, child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("owner_id")
		blueprint.Foreign("owner_id").References("id").On(parent).Name(name).RestrictOnDelete()
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	if foreignKey.Name != name {
		return "", fmt.Errorf("foreign key name = %q, want %q", foreignKey.Name, name)
	}
	return fmt.Sprintf("name=%s", foreignKey.Name), nil
}

// dropForeignScenario verifies dropping a foreign key constraint by name.
func dropForeignScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "drop-foreign"); err != nil {
		return "", err
	}
	const parent = "schema_demo_drop_foreign_tenants"
	const child = "schema_demo_drop_foreign"
	builder := dbschema.New(db)
	if err := builder.Create(parent, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(parent) }()
	if err := builder.Create(child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("tenant_id").Constrained(parent)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(child) }()

	foreignKey, found, err := findForeignKey(builder, child, parent)
	if err != nil {
		return "", err
	}
	if err := requireForeignKey(child, parent, found); err != nil {
		return "", err
	}
	if err := builder.Table(child, func(blueprint *dbschema.Blueprint) {
		blueprint.DropForeign(foreignKey.Name)
	}); err != nil {
		return "", err
	}
	foreignKeys, err := builder.GetForeignKeys(child)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("dropped=true remaining=%d", len(foreignKeys)), nil
}

// foreignDialectScenario demonstrates the dialect boundary of foreign key
// compilation: MySQL emits the constraint, while SQLite skips it in CREATE TABLE.
func foreignDialectScenario(db *gorm.DB) (string, error) {
	const parent = "schema_demo_fk_dialect_tenants"
	const child = "schema_demo_fk_dialect"
	builder := dbschema.New(db)
	if err := builder.Create(parent, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(parent) }()
	if err := builder.Create(child, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.ForeignId("tenant_id")
		blueprint.Foreign("tenant_id").References("id").On(parent)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(child) }()

	foreignKeys, err := builder.GetForeignKeys(child)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("dialect=%s foreign_keys=%d", db.Name(), len(foreignKeys)), nil
}
