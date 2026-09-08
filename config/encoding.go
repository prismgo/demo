package config

func init() {
	Add("encoding", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Encoding Driver
				|--------------------------------------------------------------------------
				|
				| This option controls the encoder used by framework features that need
				| to serialize payloads for storage or transport.
				|
			*/
			"default": Env("PRISMGO_ENCODING", "msgpack"),
		}
	})
}
