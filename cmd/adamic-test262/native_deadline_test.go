package main

import (
	"os"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	boundedrun.Register("test-native-build-clang-ar", func(args []string) (string, error) {
		code, err := os.ReadFile(args[0])
		if err != nil {
			return "", err
		}
		return "", native.Build(string(code), args[1], native.Options{})
	})
}

func boundedNativeBuild(code, binary string) error {
	source := binary + ".c"
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		return err
	}
	defer os.Remove(source)
	// Observed sanitized links peaked at 19s; 2m leaves margin for cold builds.
	_, err := boundedrun.Operation(2*time.Minute, "test-native-build-clang-ar", source, binary)
	return err
}
