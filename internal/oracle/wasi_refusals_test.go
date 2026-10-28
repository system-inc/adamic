package oracle

import "testing"

// These are asserted outcomes, never skips. Each runner checks both directions:
// a new refusal fails, and a listed fixture that runs must leave its table.
var wasiInputRefusedFixtures = map[string]string{
	"internal/oracle/testdata/node_process_host.a": "wasm32-wasi refuses process.platform: WASI has no Node host platform; use a native Node host target",
}

var wasiFileRefusedFixtures = map[string]string{
	"write_buffer": "file creation permission bits are not supported",
	"write_file":   "file creation permission bits are not supported",
	"mkdir":        "directory creation permission bits are not supported",
}

var wasiRefusedFixtures = map[string]string{
	programHost14Path: "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/closure_convention_host24.a": "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/node_fs_rm_options.a":        "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
}

var wasiProcessRefusedFixtures = map[string]string{
	"host":         "wasm32-wasi refuses process.platform: WASI has no Node host platform; use a native Node host target",
	"os_namespace": "wasm32-wasi refuses process.platform: WASI has no Node host platform; use a native Node host target",
	"terminal":     "wasm32-wasi refuses process.stdout.columns: WASI has no terminal size; use a native Node host target",
	"blocking":     "process.stdout._handle requires terminal flag control",
}

func assertWASIRefusal(t *testing.T, fixture, reason string, refused map[string]string) bool {
	t.Helper()
	want, listed := refused[fixture]
	if reason == "" {
		if listed {
			t.Fatalf("%s: listed wasm32 refusal %q, but it ran", fixture, want)
		}
		return false
	}
	if !listed {
		t.Fatalf("%s: unlisted wasm32 refusal: %s", fixture, reason)
	}
	if reason != want {
		t.Fatalf("%s: wasm32 refusal changed: got %q, want %q", fixture, reason, want)
	}
	t.Logf("REFUSED %s: %s", fixture, reason)
	return true
}
