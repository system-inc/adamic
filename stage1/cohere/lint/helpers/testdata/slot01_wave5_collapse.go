package tailwind

func AdamicWave5Named(value string) bool { return isValidNamedValue(value) }

func AdamicWave5Bucket(value string) string { return breakpointBucket(value) }
