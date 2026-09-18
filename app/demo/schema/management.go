package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// schemaDemoDatabase is the disposable MySQL database reserved for the database
// management scenarios by the local test environment.
const schemaDemoDatabase = "prismgo_schema_demo_test"

// schemaNamesContains reports whether schemas contains one named want.
func schemaNamesContains(schemas []dbschema.SchemaInfo, want string) bool {
	for _, schema := range schemas {
		if schema.Name == want {
			return true
		}
	}
	return false
}

// createDatabaseScenario verifies MySQL database creation and visibility.
func createDatabaseScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "create-database"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	if _, err := builder.DropDatabaseIfExists(schemaDemoDatabase); err != nil {
		return "", err
	}
	defer func() { _, _ = builder.DropDatabaseIfExists(schemaDemoDatabase) }()

	created, err := builder.CreateDatabase(schemaDemoDatabase)
	if err != nil {
		return "", err
	}
	schemas, err := builder.GetSchemas()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("created=%t exists=%t", created, schemaNamesContains(schemas, schemaDemoDatabase)), nil
}

// dropDatabaseScenario verifies MySQL database dropping and removal.
func dropDatabaseScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "drop-database"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	if _, err := builder.CreateDatabase(schemaDemoDatabase); err != nil {
		return "", err
	}
	dropped, err := builder.DropDatabaseIfExists(schemaDemoDatabase)
	if err != nil {
		return "", err
	}
	schemas, err := builder.GetSchemas()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("dropped=%t exists=%t", dropped, schemaNamesContains(schemas, schemaDemoDatabase)), nil
}
