package migrations_test

import (
	"testing"

	"prismgo-demo/app/demo/testing"
	"prismgo-demo/database/migrations"

	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func TestDemoOrderMigrations(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	db := database.Resolve()
	for _, migrate := range []func(*gorm.DB) error{
		migrations.CreateUsersTable,
		migrations.CreateDemoOrdersTable,
		migrations.CreateDemoReceiptsTable,
	} {
		if err := migrate(db); err != nil {
			t.Fatalf("run migration: %v", err)
		}
	}
	for _, table := range []string{"users", "demo_orders", "demo_receipts"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("migration did not create %s", table)
		}
	}
	for _, rollback := range []func(*gorm.DB) error{
		migrations.DropDemoReceiptsTable,
		migrations.DropDemoOrdersTable,
		migrations.DropUsersTable,
	} {
		if err := rollback(db); err != nil {
			t.Fatalf("roll back migration: %v", err)
		}
	}
}
