package schemademo

import (
	"errors"
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// foreignKeyChecks reads the MySQL FOREIGN_KEY_CHECKS session variable.
func foreignKeyChecks(db *gorm.DB) (int, error) {
	var value int
	if err := db.Raw("SELECT @@FOREIGN_KEY_CHECKS").Scan(&value).Error; err != nil {
		return 0, err
	}
	return value, nil
}

// disableForeignKeysScenario verifies DisableForeignKeyConstraints turns off
// MySQL foreign key checks and restores them before returning.
func disableForeignKeysScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "disable-foreign-keys"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	if err := builder.EnableForeignKeyConstraints(); err != nil {
		return "", err
	}
	before, err := foreignKeyChecks(db)
	if err != nil {
		return "", err
	}
	if err := builder.DisableForeignKeyConstraints(); err != nil {
		return "", err
	}
	defer func() { _ = builder.EnableForeignKeyConstraints() }()
	after, err := foreignKeyChecks(db)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%d after=%d", before, after), nil
}

// enableForeignKeysScenario verifies EnableForeignKeyConstraints restores MySQL
// foreign key checks after they were disabled.
func enableForeignKeysScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "enable-foreign-keys"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	if err := builder.DisableForeignKeyConstraints(); err != nil {
		return "", err
	}
	defer func() { _ = builder.EnableForeignKeyConstraints() }()
	before, err := foreignKeyChecks(db)
	if err != nil {
		return "", err
	}
	if err := builder.EnableForeignKeyConstraints(); err != nil {
		return "", err
	}
	after, err := foreignKeyChecks(db)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%d after=%d", before, after), nil
}

// withoutForeignKeysScenario verifies WithoutForeignKeyConstraints scopes the
// toggle and returns the callback error instead of a restore error.
func withoutForeignKeysScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "without-foreign-keys"); err != nil {
		return "", err
	}
	builder := dbschema.New(db)
	inside := -1
	sentinel := errors.New("schema demo callback failure")
	err := builder.WithoutForeignKeyConstraints(func() error {
		checks, checksErr := foreignKeyChecks(db)
		if checksErr != nil {
			return checksErr
		}
		inside = checks
		return sentinel
	})
	restored, restoreErr := foreignKeyChecks(db)
	if restoreErr != nil {
		return "", restoreErr
	}
	if !errors.Is(err, sentinel) {
		return "", fmt.Errorf("WithoutForeignKeyConstraints error = %v, want callback error %v", err, sentinel)
	}
	return fmt.Sprintf("inside=%d restored=%d error_precedence=%t", inside, restored, true), nil
}
