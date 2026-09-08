// Package config registers application configuration and re-exports runtime helpers.
package config

import cfg "github.com/prismgo/framework/config"

var (
	// Add registers a configuration namespace.
	Add = cfg.Add

	// Env reads an environment variable with a default value.
	Env = cfg.Env

	// Get reads a value by dot path.
	Get = cfg.Get

	// GetString reads a string value by dot path.
	GetString = cfg.GetString

	// GetInt reads an integer value by dot path.
	GetInt = cfg.GetInt

	// GetFloat64 reads a float value by dot path.
	GetFloat64 = cfg.GetFloat64

	// GetInt64 reads an int64 value by dot path.
	GetInt64 = cfg.GetInt64

	// GetUint reads an unsigned integer value by dot path.
	GetUint = cfg.GetUint

	// GetBool reads a boolean value by dot path.
	GetBool = cfg.GetBool

	// GetStringMapString reads a string map value by dot path.
	GetStringMapString = cfg.GetStringMapString

	// GetStringMap reads a map value by dot path.
	GetStringMap = cfg.GetStringMap
)
