//go:build linux

package oracle

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Linux-specific descriptor, resource-limit and signal observations. Python's POSIX helpers
// change only children, never the parallel Go test process's umask, limits or dispositions.
func TestOutputEdges(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ mode, fixture string }{
		{"ignored", "killed_after_output.a"},
	} {
		test.Run(probe.mode, func(test *testing.T) {
			test.Parallel()
			path, binary, script := sanitized(test, "internal/oracle/testdata/"+probe.fixture)
			runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			commands := [][]string{
				{"node", "--disable-warning=ExperimentalWarning", runner, path},
				{binary},
				{"node", "--disable-warning=ExperimentalWarning", runner, script},
			}
			data, _ := json.Marshal(commands)
			result := execute(test, "python3", "-I", "testdata/output_edges/probe.py", probe.mode, string(data))
			test.Logf("%s", result.stdout)
			if result.exitCode != 0 {
				test.Fatalf("edge probe: exit %d\n%s", result.exitCode, result.stderr)
			}
		})
	}
}

// Timers are outside Adamic's library. This driver calls the same runtime line writer as emitted
// C, ten times 150 ms apart. Node's driver uses its own console.log and timer on the same targets.
func TestOutputLinesArriveWhileRunning(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	binary := filepath.Join(directory, "timed")
	source := `
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <time.h>
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 adamic_string text = ADAMIC_STRING("line");
 for (int index = 0; index < 10; index++) {
  adamic_write_line(adamic_stdout, &text);
  struct timespec delay = {0, 150000000};
  nanosleep(&delay, NULL);
 }
 return 0;
}
`
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	commands := [][]string{{
		"node",
		"--input-type=module",
		"-e",
		`for(let i=0;i<10;i++){console.log('line');await new Promise(r=>setTimeout(r,150));}`,
	}, {binary}}
	data, _ := json.Marshal(commands)
	for _, mode := range []string{"terminal"} {
		test.Run(mode, func(test *testing.T) {
			result := execute(test, "python3", "-I", "testdata/output_edges/probe.py", mode, string(data))
			test.Logf("%s", result.stdout)
			if result.exitCode != 0 {
				test.Fatalf("timed probe: exit %d\n%s", result.exitCode, result.stderr)
			}
		})
	}
}

// UBSan's null-store check would exit before the hardware fault. Disable that check only in
// crash, leaving ASan enabled, so this tests the sanitizer's SIGSEGV handler after startup.
func TestStartupPreservesSanitizerSegvReport(test *testing.T) {
	test.Parallel()
	binary := filepath.Join(test.TempDir(), "fault")
	source := `
#include "adamic.h"
__attribute__((no_sanitize("undefined"), noinline))
static void crash(volatile int *pointer) {
 *pointer = 1;
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 crash(NULL);
 return 0;
}
`
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	report := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0:handle_segv=1"}, binary)
	if report.exitCode == 0 || !bytes.Contains(report.stderr, []byte("ERROR: AddressSanitizer: SEGV")) ||
		!bytes.Contains(report.stderr, []byte("#0")) || !bytes.Contains(report.stderr, []byte("crash")) {
		test.Fatalf(
			"want sanitizer SEGV report with crash stack after adamic_start: exit %d, stderr %s",
			report.exitCode,
			report.stderr,
		)
	}
	test.Logf("sanitizer report: exit %d\n%s", report.exitCode, report.stderr)
}
