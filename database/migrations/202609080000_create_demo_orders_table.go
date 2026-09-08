package migrations

import (
	"github.com/prismgo/framework/database"
	"github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(CreateDemoOrdersTable, DropDemoOrdersTable)
}

// CreateDemoOrdersTable creates the order side of the fulfillment demo.
func CreateDemoOrdersTable(_ *gorm.DB) error {
	return schema.Create("demo_orders", func(table *schema.Blueprint) {
		table.Id()
		table.ForeignId("user_id").Constrained("users").CascadeOnDelete()
		table.String("status", 32).Default("pending").Index()
		table.BigInteger("total_cents")
		table.Timestamp("paid_at").Nullable()
		table.Timestamps()
	})
}

// DropDemoOrdersTable rolls back the demo order table.
func DropDemoOrdersTable(_ *gorm.DB) error { return schema.DropIfExists("demo_orders") }
