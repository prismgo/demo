package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// dropAllTablesScenario verifies destructive table cleanup on a throwaway database.
func dropAllTablesScenario() (string, error) {
	return withIsolatedSQLite(func(db *gorm.DB) (string, error) {
		builder := dbschema.New(db)
		for _, name := range []string{"schema_demo_drop_a", "schema_demo_drop_b"} {
			if err := builder.Create(name, func(table *dbschema.Blueprint) {
				table.Id()
			}); err != nil {
				return "", err
			}
		}
		if err := builder.DropAllTables(); err != nil {
			return "", err
		}
		names, err := builder.GetTableListing(nil, false)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("remaining=%d", len(names)), nil
	})
}

// dropAllViewsScenario verifies destructive view cleanup on a throwaway database.
func dropAllViewsScenario() (string, error) {
	return withIsolatedSQLite(func(db *gorm.DB) (string, error) {
		builder := dbschema.New(db)
		if err := builder.Create("schema_demo_view_base", func(table *dbschema.Blueprint) {
			table.Id()
		}); err != nil {
			return "", err
		}
		if err := db.Exec("CREATE VIEW `schema_demo_view` AS SELECT `id` FROM `schema_demo_view_base`").Error; err != nil {
			return "", err
		}
		if err := builder.DropAllViews(); err != nil {
			return "", err
		}
		views, err := builder.GetViews(nil)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("views=%d", len(views)), nil
	})
}
