package native

import (
	goruntime "runtime"
	"strings"
	"testing"
)

// Not parallel: each variant owns a worker pool and its race detector processes.
func TestRuntimeStaticsHostWorkers(t *testing.T) {
	if goruntime.GOOS != "linux" {
		t.Skip("requires Linux ThreadSanitizer")
	}
	variants := []struct{ name, file, before, after string }{
		{name: "guarded"},
		{"exit_lock_mutant", "host_runtime.c", "pthread_mutex_lock(&exit_lock);", "(void)&exit_lock;"},
		{"raw_output_lock_mutant", "adamic.c", "bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {\n\tpthread_mutex_lock(&output_lock);", "bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {\n\t(void)&output_lock;"},
	}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			before, after := variant.before, variant.after
			if variant.file != "" {
				before = staticsSteadySource(t, variant.file)
				after = strings.ReplaceAll(before, variant.before, variant.after)
				if variant.file == "host_runtime.c" {
					after = strings.ReplaceAll(after, "pthread_mutex_unlock(&exit_lock);", "(void)&exit_lock;")
				} else {
					at := strings.Index(after, "bool adamic_write_raw(")
					after = after[:at] + strings.ReplaceAll(after[at:], "pthread_mutex_unlock(&output_lock);", "(void)&output_lock;")
				}
			}
			library, root := staticsRaceLibrary(t, variant.file, before, after)
			binary := staticsRaceFixture(t, root, library, "host")
			staticsRaceRun(t, binary, variant.file != "")
		})
	}
}
