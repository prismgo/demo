package schemademo

import (
	"fmt"
	"strings"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// mysqlColumnAttributes holds the MySQL column attributes that the generic
// schema metadata API does not expose.
type mysqlColumnAttributes struct {
	Extra     string `gorm:"column:extra"`
	Charset   string `gorm:"column:charset"`
	Collation string `gorm:"column:collation"`
}

// readMySQLColumnAttributes reads the extended attributes of a MySQL column.
func readMySQLColumnAttributes(db *gorm.DB, table, column string) (mysqlColumnAttributes, error) {
	var attributes mysqlColumnAttributes
	err := db.Raw(`
SELECT COALESCE(EXTRA, '') AS extra,
       COALESCE(CHARACTER_SET_NAME, '') AS charset,
       COALESCE(COLLATION_NAME, '') AS collation
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, table, column).Scan(&attributes).Error
	if err != nil {
		return mysqlColumnAttributes{}, fmt.Errorf("read mysql attributes for %s.%s: %w", table, column, err)
	}
	return attributes, nil
}

// unsignedScenario verifies the unsigned column modifier.
func unsignedScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_unsigned", func(blueprint *dbschema.Blueprint) {
		blueprint.Integer("count").Unsigned()
		blueprint.BigInteger("total").Unsigned()
	}, "count", "total")
}

// autoIncrementScenario verifies the auto-increment column modifier generates
// the first key value.
func autoIncrementScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_auto_increment"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Integer("seq").AutoIncrement().Primary()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()
	if err := db.Exec(fmt.Sprintf("INSERT INTO %s (seq) VALUES (NULL)", table)).Error; err != nil {
		return "", err
	}
	var generated int64
	if err := db.Raw(fmt.Sprintf("SELECT seq FROM %s LIMIT 1", table)).Scan(&generated).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("auto_increment=true generated=%t", generated == 1), nil
}

// primaryModifierScenario verifies the primary-key column modifier.
func primaryModifierScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_primary_modifier"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("code", 16).Primary()
		blueprint.String("label", 16)
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, table, "code", "label")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("code_primary=%t label_primary=%t", facts["code"].Primary, facts["label"].Primary)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// uniqueModifierScenario verifies the unique column modifier and its removal.
//
// SQLite stores the inline unique constraint in an unnamed index that the
// metadata API hides, so the full add/enforce/remove cycle needs MySQL.
func uniqueModifierScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "unique-modifier"); err != nil {
		return "", err
	}
	const table = "schema_demo_unique_modifier"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 128).Unique()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	indexed := builder.HasIndex(table, []string{"email"}, "unique")
	if err := db.Exec(fmt.Sprintf("INSERT INTO %s (email) VALUES ('a')", table)).Error; err != nil {
		return "", err
	}
	enforced := db.Exec(fmt.Sprintf("INSERT INTO %s (email) VALUES ('a')", table)).Error != nil
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("email", 128).Unique(false).Change()
	}); err != nil {
		return "", err
	}
	removed := db.Exec(fmt.Sprintf("INSERT INTO %s (email) VALUES ('a')", table)).Error == nil
	return fmt.Sprintf("unique=%t enforced=%t removed=%t", indexed, enforced, removed), nil
}

// indexModifierScenario verifies the plain index column modifier and its removal.
func indexModifierScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_index_modifier"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("email", 128).Index()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	added := builder.HasIndex(table, []string{"email"})
	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("email", 128).Index(false)
	}); err != nil {
		return "", err
	}
	removed := !builder.HasIndex(table, []string{"email"})
	return fmt.Sprintf("index=%t removed=%t", added, removed), nil
}

// defaultScenario verifies default value modifiers and their SQL literal handling.
func defaultScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_default"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("nickname", 32).Default("guest")
		blueprint.Boolean("enabled").Default(true)
		blueprint.Integer("sort").Default(0)
		blueprint.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
		blueprint.Json("payload").Nullable().Default("NULL")
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, table, "nickname", "enabled", "sort", "created_at", "payload")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("nickname=%s enabled=%s sort=%s created_at=%s payload=%s",
			facts["nickname"].Default, facts["enabled"].Default, facts["sort"].Default,
			facts["created_at"].Default, facts["payload"].Default)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// commentScenario verifies the MySQL column comment modifier.
func commentScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "comment"); err != nil {
		return "", err
	}
	const table = "schema_demo_comment"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("note", 64).Comment("hello")
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, table, "note")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("comment=%s", facts["note"].Comment)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// firstColumnScenario verifies the MySQL first-position column modifier.
func firstColumnScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "first"); err != nil {
		return "", err
	}
	const table = "schema_demo_first"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("name", 32)
		blueprint.String("tail", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("pinned", 32).First()
	}); err != nil {
		return "", err
	}
	listing, err := builder.GetColumnListing(table)
	if err != nil {
		return "", err
	}
	if len(listing) == 0 {
		return "", fmt.Errorf("schema demo first: %s has no columns after alter", table)
	}
	return fmt.Sprintf("first=%s", listing[0]), nil
}

// afterColumnScenario verifies the MySQL after-position column modifier.
func afterColumnScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "after"); err != nil {
		return "", err
	}
	const table = "schema_demo_after"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("name", 32)
		blueprint.String("tail", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	if err := builder.Table(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("inserted", 32).After("name")
	}); err != nil {
		return "", err
	}
	listing, err := builder.GetColumnListing(table)
	if err != nil {
		return "", err
	}
	return "order=" + strings.Join(listing, ","), nil
}

// charsetScenario verifies the MySQL column character set modifier.
func charsetScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "charset"); err != nil {
		return "", err
	}
	const table = "schema_demo_charset"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("name", 32).Charset("latin1")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	attributes, err := readMySQLColumnAttributes(db, table, "name")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("charset=%s", attributes.Charset), nil
}

// collationScenario verifies the MySQL column collation modifier.
func collationScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "collation"); err != nil {
		return "", err
	}
	const table = "schema_demo_collation"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.String("code", 32).Collation("utf8mb4_unicode_ci")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	attributes, err := readMySQLColumnAttributes(db, table, "code")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("collation=%s", attributes.Collation), nil
}

// useCurrentScenario verifies the current timestamp default modifier.
func useCurrentScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "use-current"); err != nil {
		return "", err
	}
	const table = "schema_demo_use_current"
	var value string
	err := inspectTable(db, table, func(blueprint *dbschema.Blueprint) {
		blueprint.Timestamp("created_at").UseCurrent()
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, table, "created_at")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("created_at_default=%s", facts["created_at"].Default)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// useCurrentOnUpdateScenario verifies the current timestamp update modifier.
func useCurrentOnUpdateScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "use-current-on-update"); err != nil {
		return "", err
	}
	const table = "schema_demo_use_current_on_update"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Timestamp("updated_at").UseCurrentOnUpdate()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	attributes, err := readMySQLColumnAttributes(db, table, "updated_at")
	if err != nil {
		return "", err
	}
	onUpdate := strings.Contains(strings.ToLower(attributes.Extra), "on update current_timestamp")
	return fmt.Sprintf("on_update=%t", onUpdate), nil
}

// invisibleScenario verifies the MySQL invisible column modifier.
func invisibleScenario(db *gorm.DB) (string, error) {
	if err := requireMySQL(db, "invisible"); err != nil {
		return "", err
	}
	const table = "schema_demo_invisible"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("secret", 32).Invisible()
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	attributes, err := readMySQLColumnAttributes(db, table, "secret")
	if err != nil {
		return "", err
	}
	invisible := strings.Contains(strings.ToUpper(attributes.Extra), "INVISIBLE")
	return fmt.Sprintf("invisible=%t", invisible), nil
}
