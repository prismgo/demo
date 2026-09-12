// Package seeders contains opt-in database seeds for Demo command examples.
package seeders

import (
	"fmt"

	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

type commandSeed struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func (commandSeed) TableName() string { return "commands_demo_seeds" }

func init() {
	database.RegisterSeeder(SeedCommandsDemo)
}

// SeedCommandsDemo inserts a repeatable marker for the command integration example.
func SeedCommandsDemo(db *gorm.DB) error {
	if err := db.AutoMigrate(&commandSeed{}); err != nil {
		return fmt.Errorf("create commands demo seed table: %w", err)
	}
	marker := commandSeed{ID: 1, Name: "commands-demo"}
	if err := db.FirstOrCreate(&marker, commandSeed{ID: 1}).Error; err != nil {
		return fmt.Errorf("insert commands demo seed: %w", err)
	}
	return nil
}
