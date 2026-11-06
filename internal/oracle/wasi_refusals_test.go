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
	// fs.watch's fixtures, refused at their first host entry WASI lacks.
	"internal/oracle/testdata/host_watch/descriptor_cap.a":        "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/same_path_close.a":       "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/same_path_unref.a":       "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/same_path_fanout.a":      "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/same_path_micro_close.a": "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/delete_recreate.a":       "wasm32-wasi refuses fs.watch: wasi preview1 has no file system event provider; use a native Node host target",
	"internal/oracle/testdata/host_watch/close_twice.a":           "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/create.a":                "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/missing.a":               "wasm32-wasi refuses fs.watch: wasi preview1 has no file system event provider; use a native Node host target",
	"internal/oracle/testdata/host_watch/mkdir.a":                 "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/on_returns.a":            "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/self_close.a":            "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/self_delete.a":           "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/tsc_shape.a":             "wasm32-wasi refuses fs.watch: wasi preview1 has no file system event provider; use a native Node host target",
	"internal/oracle/testdata/host_watch/two_events.a":            "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/unlink.a":                "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/unref.a":                 "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch/write.a":                 "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",

	"internal/oracle/testdata/host_watch_file/missing.a":        "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/appearance.a":     "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/mtime.a":          "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/cancel_one.a":     "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/cancel_all.a":     "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/duplicate.a":      "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/batch.a":          "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/unref.a":          "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/listener_throw.a": "wasm32-wasi refuses fs.unwatchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	"internal/oracle/testdata/host_watch_file/tsc_close.a":      "wasm32-wasi refuses fs.watchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target",
	programHost14Path: "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/closure_convention_host24.a": "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
	"internal/oracle/testdata/node_fs_rm_options.a":        "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target",
}

var wasiProcessRefusedFixtures = map[string]string{
	"host":         "wasm32-wasi refuses process.platform: WASI has no Node host platform; use a native Node host target",
	"os_namespace": "wasm32-wasi refuses process.platform: WASI has no Node host platform; use a native Node host target",
	"terminal":     "process.stdout._handle requires terminal flag control",
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
