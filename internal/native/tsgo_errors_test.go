package native

import (
	"os"
	"strings"
	"testing"
)

func TestTSGoRefusesWASI(t *testing.T) {
	options := Options{Target: "wasm32-wasi"}
	for name, build := range map[string]func(string, string, string, Options) error{"whole": BuildTSGo, "split": BuildSplitTSGo} {
		t.Run(name, func(t *testing.T) {
			err := build("", "unused.wasm", "missing-checker-archive", options)
			if err == nil || !strings.Contains(err.Error(), "tsgo is not supported for wasm32-wasi") {
				t.Fatalf("checker target refusal: %v", err)
			}
		})
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("set WASI_SYSROOT to build every runtime unit")
	}
	if _, err := RuntimeLibrary("", options); err != nil {
		t.Fatalf("WASI runtime archive: %v", err)
	}
}
