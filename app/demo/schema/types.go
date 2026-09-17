package schemademo

import (
	"fmt"
	"strings"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// inspectTable creates a table, runs inspect against its builder, and drops the
// table before returning. The cleanup always runs so destructive scenarios leave
// the shared application database untouched.
func inspectTable(db *gorm.DB, table string, declare func(*dbschema.Blueprint), inspect func(*dbschema.Builder) error) error {
	builder := dbschema.New(db)
	if err := builder.Create(table, declare); err != nil {
		return err
	}
	inspectErr := inspect(builder)
	dropErr := builder.Drop(table)
	if inspectErr != nil {
		return inspectErr
	}
	return dropErr
}

// typeList reports the full column type of each named column as "name=type".
func typeList(db *gorm.DB, table string, declare func(*dbschema.Blueprint), names ...string) (string, error) {
	var value string
	err := inspectTable(db, table, declare, func(builder *dbschema.Builder) error {
		parts := make([]string, 0, len(names))
		for _, name := range names {
			columnType, err := builder.GetColumnType(table, name, true)
			if err != nil {
				return err
			}
			parts = append(parts, name+"="+columnType)
		}
		value = strings.Join(parts, " ")
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// columnFacts returns the metadata of the named columns keyed by column name.
func columnFacts(builder *dbschema.Builder, table string, names ...string) (map[string]dbschema.ColumnInfo, error) {
	columns, err := builder.GetColumns(table)
	if err != nil {
		return nil, err
	}
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	facts := make(map[string]dbschema.ColumnInfo, len(names))
	for _, column := range columns {
		if _, ok := want[column.Name]; ok {
			facts[column.Name] = column
		}
	}
	for _, name := range names {
		if _, ok := facts[name]; !ok {
			return nil, fmt.Errorf("column %s.%s not found", table, name)
		}
	}
	return facts, nil
}

// signedIntegersScenario verifies the signed integer column family.
func signedIntegersScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_signed_integers", func(table *dbschema.Blueprint) {
		table.BigInteger("big")
		table.Integer("int")
		table.MediumInteger("medium")
		table.SmallInteger("small")
		table.TinyInteger("tiny")
	}, "big", "int", "medium", "small", "tiny")
}

// unsignedIntegersScenario verifies the unsigned integer column family.
func unsignedIntegersScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_unsigned_integers", func(table *dbschema.Blueprint) {
		table.UnsignedBigInteger("big")
		table.UnsignedInteger("int")
		table.UnsignedMediumInteger("medium")
		table.UnsignedSmallInteger("small")
		table.UnsignedTinyInteger("tiny")
	}, "big", "int", "medium", "small", "tiny")
}

// stringAndCharScenario verifies variable and fixed-length string columns.
func stringAndCharScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_string_char", func(table *dbschema.Blueprint) {
		table.String("name", 120)
		table.Char("code", 12)
	}, "name", "code")
}

// textTypesScenario verifies the text column family.
func textTypesScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_text_types", func(table *dbschema.Blueprint) {
		table.Text("body")
		table.TinyText("summary")
		table.MediumText("notes")
		table.LongText("archive")
	}, "body", "summary", "notes", "archive")
}

// uuidAndULIDScenario verifies UUID and ULID columns.
func uuidAndULIDScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_uuid_ulid", func(table *dbschema.Blueprint) {
		table.Uuid("identifier")
		table.Ulid("sequence")
	}, "identifier", "sequence")
}

// networkAddressesScenario verifies IP and MAC address columns.
func networkAddressesScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_network_addresses", func(table *dbschema.Blueprint) {
		table.IpAddress("ip")
		table.MacAddress("mac")
	}, "ip", "mac")
}

// rememberTokenScenario verifies the Laravel remember_token convention.
func rememberTokenScenario(db *gorm.DB) (string, error) {
	var value string
	err := inspectTable(db, "schema_demo_remember_token", func(table *dbschema.Blueprint) {
		table.RememberToken()
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, "schema_demo_remember_token", "remember_token")
		if err != nil {
			return err
		}
		column := facts["remember_token"]
		value = fmt.Sprintf("remember_token=%s nullable=%t", column.FullType, column.Nullable)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// booleanScenario verifies boolean columns.
func booleanScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_boolean", func(table *dbschema.Blueprint) {
		table.Boolean("active")
		table.Boolean("archived")
	}, "active", "archived")
}

// floatingPointScenario verifies float and double columns.
func floatingPointScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_floating_point", func(table *dbschema.Blueprint) {
		table.Float("ratio")
		table.Double("score", 10, 4)
	}, "ratio", "score")
}

// decimalScenario verifies signed and unsigned decimal columns.
func decimalScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_decimal", func(table *dbschema.Blueprint) {
		table.Decimal("price", 10, 2)
		table.UnsignedDecimal("total", 12, 3)
	}, "price", "total")
}

// dateScenario verifies date columns.
func dateScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_date", func(table *dbschema.Blueprint) {
		table.Date("born_on")
	}, "born_on")
}

// dateTimeScenario verifies datetime columns and the timezone alias.
func dateTimeScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_datetime", func(table *dbschema.Blueprint) {
		table.DateTime("seen_at")
		table.DateTimeTz("seen_at_tz")
	}, "seen_at", "seen_at_tz")
}

// timeScenario verifies time columns and the timezone alias.
func timeScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_time", func(table *dbschema.Blueprint) {
		table.Time("starts_at")
		table.TimeTz("starts_at_tz")
	}, "starts_at", "starts_at_tz")
}

// timestampScenario verifies timestamp columns and the timezone alias.
func timestampScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_timestamp", func(table *dbschema.Blueprint) {
		table.Timestamp("published_at")
		table.TimestampTz("published_at_tz")
	}, "published_at", "published_at_tz")
}

// yearScenario verifies year columns.
func yearScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_year", func(table *dbschema.Blueprint) {
		table.Year("release_year")
	}, "release_year")
}

// timestampsScenario verifies the created_at and updated_at convenience columns.
func timestampsScenario(db *gorm.DB) (string, error) {
	return timestampFacts(db, "schema_demo_timestamps", func(table *dbschema.Blueprint) {
		table.Timestamps()
	})
}

// nullableTimestampsScenario verifies the nullable timestamp convenience alias.
func nullableTimestampsScenario(db *gorm.DB) (string, error) {
	return timestampFacts(db, "schema_demo_nullable_timestamps", func(table *dbschema.Blueprint) {
		table.NullableTimestamps()
	})
}

// timestampFacts reports the timestamp convention columns and their nullability.
func timestampFacts(db *gorm.DB, table string, declare func(*dbschema.Blueprint)) (string, error) {
	var value string
	err := inspectTable(db, table, declare, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, table, "created_at", "updated_at")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("created_at=%s nullable=%t updated_at=%s nullable=%t",
			facts["created_at"].FullType, facts["created_at"].Nullable,
			facts["updated_at"].FullType, facts["updated_at"].Nullable)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// softDeletesScenario verifies the soft-delete column and its default index.
func softDeletesScenario(db *gorm.DB) (string, error) {
	var value string
	err := inspectTable(db, "schema_demo_soft_deletes", func(table *dbschema.Blueprint) {
		table.SoftDeletes()
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, "schema_demo_soft_deletes", "deleted_at")
		if err != nil {
			return err
		}
		column := facts["deleted_at"]
		value = fmt.Sprintf("deleted_at=%s nullable=%t index=%t",
			column.FullType, column.Nullable, builder.HasIndex("schema_demo_soft_deletes", []string{"deleted_at"}))
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// binaryScenario verifies binary columns.
func binaryScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_binary", func(table *dbschema.Blueprint) {
		table.Binary("payload")
	}, "payload")
}

// jsonScenario verifies JSON and JSONB columns.
func jsonScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_json", func(table *dbschema.Blueprint) {
		table.Json("meta")
		table.Jsonb("meta_b")
	}, "meta", "meta_b")
}

// enumAndSetScenario verifies enum and set columns.
func enumAndSetScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_enum_set", func(table *dbschema.Blueprint) {
		table.Enum("status", []string{"draft", "published"})
		table.Set("flags", []string{"featured", "pinned"})
	}, "status", "flags")
}

// spatialTypesScenario verifies geometry, geography, point, line, and polygon columns.
func spatialTypesScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_spatial_types", func(table *dbschema.Blueprint) {
		table.Geometry("geo")
		table.Geography("geog")
		table.Point("pt")
		table.LineString("line")
		table.Polygon("poly")
	}, "geo", "geog", "pt", "line", "poly")
}

// vectorScenario verifies vector columns.
func vectorScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_vector", func(table *dbschema.Blueprint) {
		table.Vector("embedding", 3)
	}, "embedding")
}

// foreignIDScenario verifies foreign ID columns.
func foreignIDScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_foreign_id", func(table *dbschema.Blueprint) {
		table.ForeignId("owner_id")
	}, "owner_id")
}

// foreignIDForScenario verifies the ForeignIdFor alias semantics.
func foreignIDForScenario(db *gorm.DB) (string, error) {
	return typeList(db, "schema_demo_foreign_id_for", func(table *dbschema.Blueprint) {
		table.ForeignIdFor("user_id")
	}, "user_id")
}

// morphsScenario verifies polymorphic ID, type, and index columns.
func morphsScenario(db *gorm.DB) (string, error) {
	var value string
	err := inspectTable(db, "schema_demo_morphs", func(table *dbschema.Blueprint) {
		table.Morphs("taggable")
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, "schema_demo_morphs", "taggable_id", "taggable_type")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("taggable_id=%s taggable_type=%s index=%t",
			facts["taggable_id"].FullType, facts["taggable_type"].FullType,
			builder.HasIndex("schema_demo_morphs", []string{"taggable_id", "taggable_type"}))
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// nullableMorphsScenario verifies nullable polymorphic columns and their index.
func nullableMorphsScenario(db *gorm.DB) (string, error) {
	var value string
	err := inspectTable(db, "schema_demo_nullable_morphs", func(table *dbschema.Blueprint) {
		table.NullableMorphs("taggable")
	}, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, "schema_demo_nullable_morphs", "taggable_id", "taggable_type")
		if err != nil {
			return err
		}
		value = fmt.Sprintf("taggable_id=%s nullable=%t taggable_type=%s nullable=%t index=%t",
			facts["taggable_id"].FullType, facts["taggable_id"].Nullable,
			facts["taggable_type"].FullType, facts["taggable_type"].Nullable,
			builder.HasIndex("schema_demo_nullable_morphs", []string{"taggable_id", "taggable_type"}))
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// nullableScenario verifies the nullable column modifier.
func nullableScenario(db *gorm.DB) (string, error) {
	return nullabilityFacts(db, "schema_demo_nullable", func(table *dbschema.Blueprint) {
		table.String("nickname", 64).Nullable()
		table.Text("bio").Nullable()
	}, "nickname", "bio")
}

// notNullScenario verifies the not-null column modifier, including precedence
// when Nullable and NotNull are chained on the same column.
func notNullScenario(db *gorm.DB) (string, error) {
	return nullabilityFacts(db, "schema_demo_not_null", func(table *dbschema.Blueprint) {
		table.String("nickname", 64).NotNull()
		table.Text("bio")
		table.String("legacy", 32).Nullable().NotNull()
	}, "nickname", "bio", "legacy")
}

// nullabilityFacts reports the full types and nullability of the named columns.
func nullabilityFacts(db *gorm.DB, table string, declare func(*dbschema.Blueprint), names ...string) (string, error) {
	var value string
	err := inspectTable(db, table, declare, func(builder *dbschema.Builder) error {
		facts, err := columnFacts(builder, table, names...)
		if err != nil {
			return err
		}
		parts := make([]string, 0, len(names))
		for _, name := range names {
			parts = append(parts, name+"_nullable="+fmt.Sprint(facts[name].Nullable))
		}
		value = strings.Join(parts, " ")
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}
