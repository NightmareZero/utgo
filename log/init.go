package log

func init() {
	InitWithConfig(LogConfig{
		Sync:      true,
		Level:     "debug",
		Caller:    true,
		Console:   true,
		NotToFile: true,
	})
}
