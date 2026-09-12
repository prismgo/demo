package catalog

func encryptionEntries() []Entry {
	return []Entry{
		baselineEntry("encryption", "Key rotation", "轮换密钥", "Rotating Keys", "demo:encryption list", "list", "TestEncryptionDemo", LevelHermetic, StatusPlanned),
	}
}
