package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// precedenceModel exercises explicit GORM tags against schema defaults.
type precedenceModel struct {
	ID    uint
	Name  string
	Sized string `gorm:"size:77"`
	Typed string `gorm:"type:char(10)"`
}

// TableName pins the model table so the scenario is deterministic.
func (precedenceModel) TableName() string { return "schema_demo_precedence" }

// defaultStringLengthScenario verifies String and Char inherit DefaultStringLength.
func defaultStringLengthScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "default-string-length"); err != nil {
		return "", err
	}
	dbschema.DefaultStringLength(191)
	defer dbschema.DefaultStringLength(255)
	if err := dbschema.New(db).Create("schema_demo_string_length", func(table *dbschema.Blueprint) {
		table.String("name")
		table.Char("code")
	}); err != nil {
		return "", err
	}
	defer func() { _ = dbschema.New(db).Drop("schema_demo_string_length") }()

	name, err := dbschema.New(db).GetColumnType("schema_demo_string_length", "name", true)
	if err != nil {
		return "", err
	}
	code, err := dbschema.New(db).GetColumnType("schema_demo_string_length", "code", true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("name=%s code=%s", name, code), nil
}

// defaultTimePrecisionScenario verifies time columns inherit DefaultTimePrecision.
func defaultTimePrecisionScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "default-time-precision"); err != nil {
		return "", err
	}
	precision := 3
	dbschema.DefaultTimePrecision(&precision)
	defer dbschema.DefaultTimePrecision(nil)
	if err := dbschema.New(db).Create("schema_demo_time_precision", func(table *dbschema.Blueprint) {
		table.DateTime("seen_at")
		table.Time("starts_at")
		table.Timestamp("published_at")
	}); err != nil {
		return "", err
	}
	defer func() { _ = dbschema.New(db).Drop("schema_demo_time_precision") }()

	dateTime, err := dbschema.New(db).GetColumnType("schema_demo_time_precision", "seen_at", true)
	if err != nil {
		return "", err
	}
	timeType, err := dbschema.New(db).GetColumnType("schema_demo_time_precision", "starts_at", true)
	if err != nil {
		return "", err
	}
	timestamp, err := dbschema.New(db).GetColumnType("schema_demo_time_precision", "published_at", true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("datetime=%s time=%s timestamp=%s", dateTime, timeType, timestamp), nil
}

// defaultMorphKeyTypeScenario verifies the default integer polymorphic ID.
func defaultMorphKeyTypeScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "default-morph-key-type"); err != nil {
		return "", err
	}
	dbschema.DefaultMorphKeyType("int")
	defer dbschema.DefaultMorphKeyType("int")
	if err := dbschema.New(db).Create("schema_demo_morph_int", func(table *dbschema.Blueprint) {
		table.Morphs("owner")
	}); err != nil {
		return "", err
	}
	defer func() { _ = dbschema.New(db).Drop("schema_demo_morph_int") }()
	return morphKeyTypes(db, "schema_demo_morph_int")
}

// morphUsingUUIDsScenario verifies the UUID polymorphic key shortcut.
func morphUsingUUIDsScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "morph-using-uuids"); err != nil {
		return "", err
	}
	dbschema.MorphUsingUuids()
	defer dbschema.DefaultMorphKeyType("int")
	if err := dbschema.New(db).Create("schema_demo_morph_uuid", func(table *dbschema.Blueprint) {
		table.Morphs("owner")
	}); err != nil {
		return "", err
	}
	defer func() { _ = dbschema.New(db).Drop("schema_demo_morph_uuid") }()
	return morphKeyTypes(db, "schema_demo_morph_uuid")
}

// morphUsingULIDsScenario verifies the ULID polymorphic key shortcut.
func morphUsingULIDsScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "morph-using-ulids"); err != nil {
		return "", err
	}
	dbschema.MorphUsingUlids()
	defer dbschema.DefaultMorphKeyType("int")
	if err := dbschema.New(db).Create("schema_demo_morph_ulid", func(table *dbschema.Blueprint) {
		table.Morphs("owner")
	}); err != nil {
		return "", err
	}
	defer func() { _ = dbschema.New(db).Drop("schema_demo_morph_ulid") }()
	return morphKeyTypes(db, "schema_demo_morph_ulid")
}

// morphKeyTypes reads the polymorphic ID and type column definitions.
func morphKeyTypes(db *gorm.DB, table string) (string, error) {
	builder := dbschema.New(db)
	idType, err := builder.GetColumnType(table, "owner_id", true)
	if err != nil {
		return "", err
	}
	typeType, err := builder.GetColumnType(table, "owner_type", true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("owner_id=%s owner_type=%s", idType, typeType), nil
}

// explicitTagPrecedenceScenario verifies explicit GORM tags win over defaults.
func explicitTagPrecedenceScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "explicit-tag-precedence"); err != nil {
		return "", err
	}
	dbschema.DefaultStringLength(191)
	defer dbschema.DefaultStringLength(255)
	if err := dbschema.New(db).SyncModels(&precedenceModel{}); err != nil {
		return "", err
	}
	defer func() { _ = dbschema.New(db).Drop("schema_demo_precedence") }()

	builder := dbschema.New(db)
	name, err := builder.GetColumnType("schema_demo_precedence", "name", true)
	if err != nil {
		return "", err
	}
	sized, err := builder.GetColumnType("schema_demo_precedence", "sized", true)
	if err != nil {
		return "", err
	}
	typed, err := builder.GetColumnType("schema_demo_precedence", "typed", true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("name=%s sized=%s typed=%s", name, sized, typed), nil
}
