package config

func init() {
	Add("filesystem", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Filesystem Disk
				|--------------------------------------------------------------------------
				|
				| Here you may specify the filesystem disk that should be used by
				| default for storing and retrieving files.
				|
			*/
			"default": Env("FILESYSTEM_DISK", "local"),
			"cloud":   Env("FILESYSTEM_CLOUD", "oss"),

			/*
				|--------------------------------------------------------------------------
				| Temporary URL Signing
				|--------------------------------------------------------------------------
				|
				| Local private disks may use signed temporary URLs. When the signing key
				| is empty, the application key may be used by the framework.
				|
			*/
			"temporary_url": map[string]interface{}{
				"signing_key": Env("FILESYSTEM_SIGNING_KEY", ""),
			},

			/*
				|--------------------------------------------------------------------------
				| Filesystem Disks
				|--------------------------------------------------------------------------
				|
				| You may configure as many filesystem disks as necessary. Each disk
				| describes a storage driver and its location or public URL.
				|
			*/
			"disks": map[string]interface{}{
				"local": map[string]interface{}{
					"driver":     "local",
					"root":       Env("FILESYSTEM_LOCAL_ROOT", "storage/app/private"),
					"url":        Env("FILESYSTEM_LOCAL_URL", ""),
					"visibility": Env("FILESYSTEM_LOCAL_VISIBILITY", "private"),
					"serve":      Env("FILESYSTEM_LOCAL_SERVE", true),
				},
				"public": map[string]interface{}{
					"driver":     "local",
					"root":       Env("FILESYSTEM_PUBLIC_ROOT", "storage/app/public"),
					"url":        Env("FILESYSTEM_PUBLIC_URL", "http://localhost:8080/storage"),
					"visibility": Env("FILESYSTEM_PUBLIC_VISIBILITY", "public"),
					"serve":      Env("FILESYSTEM_PUBLIC_SERVE", true),
				},
				"oss": map[string]interface{}{
					"driver":     "oss",
					"bucket":     Env("FILESYSTEM_OSS_BUCKET", ""),
					"endpoint":   Env("FILESYSTEM_OSS_ENDPOINT", ""),
					"access_key": Env("FILESYSTEM_OSS_ACCESS_KEY_ID", ""),
					"secret_key": Env("FILESYSTEM_OSS_ACCESS_KEY_SECRET", ""),
					"prefix":     Env("FILESYSTEM_OSS_PREFIX", ""),
					"url":        Env("FILESYSTEM_OSS_URL", ""),
					"visibility": Env("FILESYSTEM_OSS_VISIBILITY", "private"),
					"timeout":    Env("FILESYSTEM_OSS_TIMEOUT", 30),
				},
			},

			/*
				|--------------------------------------------------------------------------
				| Symbolic Links
				|--------------------------------------------------------------------------
				|
				| Here you may configure the symbolic links that will be created when the
				| `storage:link` command is executed. The map keys should be the locations
				| of the links and the values should be their targets.
				|
			*/
			"links": map[string]interface{}{
				"public/storage": "storage/app/public",
				// "public/images": "storage/app/images",
			},
		}
	})
}
