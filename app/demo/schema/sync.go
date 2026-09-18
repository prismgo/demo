package schemademo

import (
	"fmt"

	"github.com/prismgo/framework/database"
	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// syncDemoUser is the model used by the basic SyncModels scenario.
type syncDemoUser struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:64"`
}

// TableName pins the demo table so the scenario is independent of GORM naming.
func (syncDemoUser) TableName() string { return "schema_demo_sync_users" }

// syncDemoColumnV1 is the initial model used by the missing-column scenario.
type syncDemoColumnV1 struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:32"`
}

// TableName pins both syncDemoColumn versions to the same demo table.
func (syncDemoColumnV1) TableName() string { return "schema_demo_sync_columns" }

// syncDemoColumnV2 adds a field to syncDemoColumnV1 so SyncModels has to add a
// missing column to the existing table.
type syncDemoColumnV2 struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:32"`
	Email string `gorm:"size:64"`
}

// TableName pins both syncDemoColumn versions to the same demo table.
func (syncDemoColumnV2) TableName() string { return "schema_demo_sync_columns" }

// syncDemoDefault is the model used by the defaults scenario.
type syncDemoDefault struct {
	ID     uint   `gorm:"primaryKey"`
	Status string `gorm:"size:16;default:active"`
}

// TableName pins the demo table so the scenario is independent of GORM naming.
func (syncDemoDefault) TableName() string { return "schema_demo_sync_defaults" }

// syncModelsScenario verifies SyncModels creates a model table with its columns.
func syncModelsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_sync_users"
	builder := dbschema.New(db)
	if err := builder.SyncModels(syncDemoUser{}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	columns, err := builder.GetColumnListing(table)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("created=%t columns=%d", builder.HasTable(table), len(columns)), nil
}

// syncModelsColumnsScenario verifies SyncModels creates a missing table and then
// adds a newly declared column to the existing table.
func syncModelsColumnsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_sync_columns"
	builder := dbschema.New(db)
	if err := builder.SyncModels(syncDemoColumnV1{}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	created := builder.HasTable(table) && builder.HasColumn(table, "name")
	if err := builder.SyncModels(syncDemoColumnV2{}); err != nil {
		return "", err
	}
	return fmt.Sprintf("created=%t added=%t", created, builder.HasColumn(table, "email")), nil
}

// syncModelsDefaultsScenario verifies model defaults survive into the created
// table and that MySQL table options are appended by SyncModels.
func syncModelsDefaultsScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_sync_defaults"
	builder := dbschema.New(db)
	if err := builder.SyncModels(syncDemoDefault{}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	facts, err := columnFacts(builder, table, "status")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("default=%s options=%s", facts["status"].Default, database.TableOptions("mysql", "InnoDB")), nil
}
