package migrations

import (
	"github.com/prismgo/framework/database"
	"github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(CreateUsersTable, DropUsersTable)
}

// CreateUsersTable creates the default users table for new applications.
func CreateUsersTable(db *gorm.DB) error {
	return schema.Create("users", func(table *schema.Blueprint) {
		table.Id()
		table.String("name")
		table.String("email").Unique()
		table.Timestamp("email_verified_at").Nullable()
		table.String("password")
		table.RememberToken()
		table.Timestamps()
		table.SoftDeletes()
	})
}

// DropUsersTable rolls back the default users table.
func DropUsersTable(db *gorm.DB) error {
	return schema.DropIfExists("users")
}
