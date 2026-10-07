package tailwind

import (
	"encoding/json"
	"os"
	"sync"
)

func AdamicSettingsKey(settings ClassLiteralSettings) string { return settings.key() }

var adamicSettingsLock sync.Mutex

// Called only by the capture overlay at the real settings-key entry point.
func AdamicRecordSettings(settings ClassLiteralSettings) {
	path := os.Getenv("ADAMIC_SLOT04_SETTINGS")
	if path == "" {
		return
	}
	adamicSettingsLock.Lock()
	defer adamicSettingsLock.Unlock()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(file).Encode(settings); err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}

var adamicCreated = map[*ClassLiteralReader]ClassLiteralSettings{}
var adamicIdentities = map[*ClassLiteralReader]int{}
var adamicCreateCount int

// The reader-factory overlay records inputs without changing the real result.
func AdamicRecordCreated(reader *ClassLiteralReader, settings ClassLiteralSettings) {
	adamicCreated[reader] = settings
	adamicCreateCount++
}
func AdamicCompiledReader(key string, settings ClassLiteralSettings) (int, int, string) {
	reader := compiledClassLiteralReader(key, settings)
	identity, exists := adamicIdentities[reader]
	if !exists {
		identity = len(adamicIdentities) + 1
		adamicIdentities[reader] = identity
	}
	return identity, adamicCreateCount, adamicCreated[reader].key()
}
