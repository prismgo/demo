package migrations

import (
	"github.com/prismgo/framework/database"
	"github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(CreateDemoReceiptsTable, DropDemoReceiptsTable)
}

// CreateDemoReceiptsTable creates generated receipt metadata.
func CreateDemoReceiptsTable(_ *gorm.DB) error {
	return schema.Create("demo_receipts", func(table *schema.Blueprint) {
		table.Id()
		table.ForeignId("order_id").Constrained("demo_orders").CascadeOnDelete()
		table.String("path", 512)
		table.Timestamp("generated_at")
		table.Timestamps()
		table.Unique("order_id")
	})
}

// DropDemoReceiptsTable rolls back generated receipt metadata.
func DropDemoReceiptsTable(_ *gorm.DB) error { return schema.DropIfExists("demo_receipts") }
