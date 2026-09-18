package schemademo

import dbschema "github.com/prismgo/framework/database/schema"

// storedAsScenario keeps the stored generated column placeholder referenced and
// documents that it is currently not compiled.
func storedAsScenario() (string, error) {
	var _ func(string) *dbschema.ColumnDefinition = (*dbschema.ColumnDefinition)(nil).StoredAs
	return "stored_as=placeholder compiled=false", nil
}

// virtualAsScenario keeps the virtual generated column placeholder referenced
// and documents that it is currently not compiled.
func virtualAsScenario() (string, error) {
	var _ func(string) *dbschema.ColumnDefinition = (*dbschema.ColumnDefinition)(nil).VirtualAs
	return "virtual_as=placeholder compiled=false", nil
}

// fromScenario keeps the Laravel integer start-value modifier referenced and
// documents that it does not affect the compiled SQL.
func fromScenario() (string, error) {
	var _ func(int) *dbschema.ColumnDefinition = (*dbschema.ColumnDefinition)(nil).From
	return "from=compatibility affects_sql=false", nil
}

// instantScenario keeps the Laravel instant algorithm modifier referenced and
// documents that it does not affect the compiled SQL.
func instantScenario() (string, error) {
	var _ func() *dbschema.ColumnDefinition = (*dbschema.ColumnDefinition)(nil).Instant
	return "instant=compatibility affects_sql=false", nil
}

// lockScenario keeps the Laravel lock modifier referenced and documents that it
// does not affect the compiled SQL.
func lockScenario() (string, error) {
	var _ func(string) *dbschema.ColumnDefinition = (*dbschema.ColumnDefinition)(nil).Lock
	return "lock=compatibility affects_sql=false", nil
}

// changeSemanticsScenario documents that Change marks a column modification and
// leaves idempotency to the caller.
func changeSemanticsScenario() (string, error) {
	var _ func() *dbschema.ColumnDefinition = (*dbschema.ColumnDefinition)(nil).Change
	return "change=modify_column idempotency=caller", nil
}
