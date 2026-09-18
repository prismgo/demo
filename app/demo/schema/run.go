// Package schemademo contains runnable examples of the Laravel-style schema builder.
package schemademo

import (
	"fmt"

	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

// Result records one observable schema scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a schema catalog scenario against the application database.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("schema demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

// run dispatches one scenario. SQLite-only scenarios open an isolated database;
// the remaining scenarios use the application's default connection so hermetic
// tests run on SQLite and MySQL integration tests run on MySQL.
func run(name string) (string, error) {
	switch name {
	case "architecture":
		return architectureScenario()
	case "sqlite-extension":
		return sqliteExtensionScenario()
	case "sqlite-connection-scope":
		return sqliteConnectionScopeScenario()
	case "create-dialect-options":
		return createDialectOptionsScenario()
	case "drop-all-types":
		return dropAllTypesScenario()
	case "stored-as":
		return storedAsScenario()
	case "virtual-as":
		return virtualAsScenario()
	case "from":
		return fromScenario()
	case "instant":
		return instantScenario()
	case "lock":
		return lockScenario()
	case "change-semantics":
		return changeSemanticsScenario()
	case "metadata-types":
		return metadataTypesScenario()
	case "foreign-key-toggle-dialects":
		return foreignKeyToggleDialectsScenario()
	case "sync-models-boundaries":
		return syncModelsBoundariesScenario()
	case "dialect-compatibility":
		return dialectCompatibilityScenario()
	case "laravel-compatibility":
		return laravelCompatibilityScenario()
	case "ensure-extension":
		return ensureExtensionScenario()
	case "ensure-vector-extension":
		return ensureVectorExtensionScenario()
	case "drop-all-tables":
		return dropAllTablesScenario()
	case "drop-all-views":
		return dropAllViewsScenario()
	case "spatial-types":
		return withIsolatedSQLite(func(db *gorm.DB) (string, error) {
			return spatialTypesScenario(db)
		})
	case "vector":
		return withIsolatedSQLite(func(db *gorm.DB) (string, error) {
			return vectorScenario(db)
		})
	}
	return withDatabase(func(db *gorm.DB) (string, error) {
		return databaseScenario(db, name)
	})
}

// withDatabase resolves the application database connection used by scenarios.
func withDatabase(fn func(*gorm.DB) (string, error)) (string, error) {
	db := database.Resolve()
	if db == nil {
		return "", fmt.Errorf("application database connection is not available")
	}
	return fn(db)
}

// databaseScenario executes a scenario that needs a database connection.
func databaseScenario(db *gorm.DB, name string) (string, error) {
	switch name {
	case "default-string-length":
		return defaultStringLengthScenario(db)
	case "default-time-precision":
		return defaultTimePrecisionScenario(db)
	case "default-morph-key-type":
		return defaultMorphKeyTypeScenario(db)
	case "morph-using-uuids":
		return morphUsingUUIDsScenario(db)
	case "morph-using-ulids":
		return morphUsingULIDsScenario(db)
	case "explicit-tag-precedence":
		return explicitTagPrecedenceScenario(db)
	case "bind":
		return bindScenario(db)
	case "new":
		return newScenario(db)
	case "named-connection":
		return namedConnectionScenario(db)
	case "facade":
		return facadeScenario()
	case "create":
		return createScenario(db)
	case "create-validation":
		return createValidationScenario(db)
	case "table":
		return tableScenario(db)
	case "add-column":
		return addColumnScenario(db)
	case "change-column":
		return changeColumnScenario(db)
	case "rename-column":
		return renameColumnScenario(db)
	case "drop-column":
		return dropColumnScenario(db)
	case "drop-columns":
		return dropColumnsScenario(db)
	case "raw":
		return rawScenario(db)
	case "rename":
		return renameScenario(db)
	case "drop":
		return dropScenario(db)
	case "builder-drop-columns":
		return builderDropColumnsScenario(db)
	case "id":
		return idScenario(db)
	case "increments":
		return incrementsScenario(db)
	case "signed-integers":
		return signedIntegersScenario(db)
	case "unsigned-integers":
		return unsignedIntegersScenario(db)
	case "string-char":
		return stringAndCharScenario(db)
	case "text-types":
		return textTypesScenario(db)
	case "uuid-ulid":
		return uuidAndULIDScenario(db)
	case "network-addresses":
		return networkAddressesScenario(db)
	case "remember-token":
		return rememberTokenScenario(db)
	case "boolean":
		return booleanScenario(db)
	case "floating-point":
		return floatingPointScenario(db)
	case "decimal":
		return decimalScenario(db)
	case "date":
		return dateScenario(db)
	case "datetime":
		return dateTimeScenario(db)
	case "time":
		return timeScenario(db)
	case "timestamp":
		return timestampScenario(db)
	case "year":
		return yearScenario(db)
	case "timestamps":
		return timestampsScenario(db)
	case "soft-deletes":
		return softDeletesScenario(db)
	case "binary":
		return binaryScenario(db)
	case "json":
		return jsonScenario(db)
	case "enum-set":
		return enumAndSetScenario(db)
	case "foreign-id":
		return foreignIDScenario(db)
	case "foreign-id-for":
		return foreignIDForScenario(db)
	case "morphs":
		return morphsScenario(db)
	case "nullable-morphs":
		return nullableMorphsScenario(db)
	case "nullable-timestamps":
		return nullableTimestampsScenario(db)
	case "nullable":
		return nullableScenario(db)
	case "not-null":
		return notNullScenario(db)
	case "unsigned":
		return unsignedScenario(db)
	case "auto-increment":
		return autoIncrementScenario(db)
	case "primary-modifier":
		return primaryModifierScenario(db)
	case "unique-modifier":
		return uniqueModifierScenario(db)
	case "index-modifier":
		return indexModifierScenario(db)
	case "default":
		return defaultScenario(db)
	case "comment":
		return commentScenario(db)
	case "first":
		return firstColumnScenario(db)
	case "after":
		return afterColumnScenario(db)
	case "charset":
		return charsetScenario(db)
	case "collation":
		return collationScenario(db)
	case "use-current":
		return useCurrentScenario(db)
	case "use-current-on-update":
		return useCurrentOnUpdateScenario(db)
	case "invisible":
		return invisibleScenario(db)
	case "drop-remember-token":
		return dropRememberTokenScenario(db)
	case "drop-timestamps":
		return dropTimestampsScenario(db)
	case "drop-timestamps-tz":
		return dropTimestampsTzScenario(db)
	case "drop-soft-deletes":
		return dropSoftDeletesScenario(db)
	case "drop-soft-deletes-tz":
		return dropSoftDeletesTzScenario(db)
	case "drop-morphs":
		return dropMorphsScenario(db)
	case "drop-constrained-foreign-id":
		return dropConstrainedForeignIDScenario(db)
	case "drop-foreign-id-for":
		return dropForeignIDForScenario(db)
	case "primary-index":
		return primaryIndexScenario(db)
	case "unique-index":
		return uniqueIndexScenario(db)
	case "index":
		return plainIndexScenario(db)
	case "fulltext-index":
		return fulltextIndexScenario(db)
	case "spatial-index":
		return spatialIndexScenario(db)
	case "named-indexes":
		return namedIndexesScenario(db)
	case "index-naming":
		return indexNamingScenario(db)
	case "rename-index":
		return renameIndexScenario(db)
	case "drop-index":
		return dropIndexScenario(db)
	case "drop-unique":
		return dropUniqueScenario(db)
	case "drop-primary":
		return dropPrimaryScenario(db)
	case "drop-fulltext":
		return dropFulltextScenario(db)
	case "drop-spatial-index":
		return dropSpatialIndexScenario(db)
	case "constrained":
		return constrainedScenario(db)
	case "constrained-explicit":
		return constrainedExplicitScenario(db)
	case "foreign":
		return foreignScenario(db)
	case "foreign-actions":
		return foreignActionsScenario(db)
	case "cascade-actions":
		return cascadeActionsScenario(db)
	case "restrict-actions":
		return restrictActionsScenario(db)
	case "null-actions":
		return nullActionsScenario(db)
	case "no-action-actions":
		return noActionActionsScenario(db)
	case "foreign-name":
		return foreignNameScenario(db)
	case "foreign-dialect":
		return foreignDialectScenario(db)
	case "drop-foreign":
		return dropForeignScenario(db)
	case "table-view-existence":
		return tableViewExistenceScenario(db)
	case "tables":
		return tablesScenario(db)
	case "table-listing":
		return tableListingScenario(db)
	case "views":
		return viewsScenario(db)
	case "schemas":
		return schemasScenario(db)
	case "types":
		return typesScenario(db)
	case "schema-filter":
		return schemaFilterScenario(db)
	case "has-columns":
		return hasColumnsScenario(db)
	case "columns":
		return columnsScenario(db)
	case "column-type":
		return columnTypeScenario(db)
	case "has-index":
		return hasIndexScenario(db)
	case "indexes":
		return indexesScenario(db)
	case "foreign-keys":
		return foreignKeysScenario(db)
	case "when-has-column":
		return whenHasColumnScenario(db)
	case "when-missing-column":
		return whenMissingColumnScenario(db)
	case "when-missing-index":
		return whenMissingIndexScenario(db)
	case "disable-foreign-keys":
		return disableForeignKeysScenario(db)
	case "enable-foreign-keys":
		return enableForeignKeysScenario(db)
	case "without-foreign-keys":
		return withoutForeignKeysScenario(db)
	case "create-database":
		return createDatabaseScenario(db)
	case "drop-database":
		return dropDatabaseScenario(db)
	case "sync-models":
		return syncModelsScenario(db)
	case "sync-models-columns":
		return syncModelsColumnsScenario(db)
	case "sync-models-defaults":
		return syncModelsDefaultsScenario(db)
	default:
		return "", fmt.Errorf("unknown schema scenario %q", name)
	}
}

// requireMySQL rejects a dialect-specific scenario on a non-MySQL connection.
//
// The schema defaults only affect MySQL column definitions; SQLite ignores
// length, precision and generated SQL, so those scenarios must run on MySQL.
func requireMySQL(db *gorm.DB, name string) error {
	if db == nil || db.Dialector == nil {
		return fmt.Errorf("schema scenario %q requires a database connection", name)
	}
	if db.Name() != "mysql" {
		return fmt.Errorf("schema scenario %q requires a MySQL connection, got %q", name, db.Name())
	}
	return nil
}
