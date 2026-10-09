//go:build linux

package oracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Linux-specific descriptor, resource-limit and signal observations. Python's POSIX helpers
// change only children, never the parallel Go test process's umask, limits or dispositions.
func TestOutputEdges(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ mode, fixture string }{
		{"fsize", "output_edges/fsize.a"},
		{"fsize_out", "output_edges/fsize_out.a"},
		{"nonblock", "large_output.a"},
		{"closed", "output_edges/closed.a"},
		{"signals", "killed_after_output.a"},
		{"ignored", "killed_after_output.a"},
		{"usr1", "output_edges/usr1.a"},
		{"panic", "output_edges/panic_surrogate.a"},
	} {
		t.Run(probe.mode, func(t *testing.T) {
			t.Parallel()
			path, binary, script := sanitized(t, "internal/oracle/testdata/"+probe.fixture)
			runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			commands := [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, path}, {binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}}
			data, _ := json.Marshal(commands)
			result := execute(t, "python3", "-I", "testdata/output_edges/probe.py", probe.mode, string(data))
			t.Logf("%s", result.stdout)
			if result.exitCode != 0 {
				t.Fatalf("edge probe: exit %d\n%s", result.exitCode, result.stderr)
			}
		})
	}
}

// Timers are outside Adamic's library. This driver calls the same runtime line writer as emitted
// C, ten times 150 ms apart. Node's driver uses its own console.log and timer on the same targets.
func TestOutputLinesArriveWhileRunning(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
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
		t.Fatal(err)
	}
	commands := [][]string{{"node", "--input-type=module", "-e", `for(let i=0;i<10;i++){console.log('line');await new Promise(r=>setTimeout(r,150));}`}, {binary}}
	data, _ := json.Marshal(commands)
	for _, mode := range []string{"terminal", "file"} {
		t.Run(mode, func(t *testing.T) {
			result := execute(t, "python3", "-I", "testdata/output_edges/probe.py", mode, string(data))
			t.Logf("%s", result.stdout)
			if result.exitCode != 0 {
				t.Fatalf("timed probe: exit %d\n%s", result.exitCode, result.stderr)
			}
		})
	}
}

// Linux's public realtime signals (34 through 64 with glibc) terminate by default too.
func TestRealtimeSignalsLeaveWhatWasPrinted(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/killed_after_output.a")
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	commands := [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, path}, {binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}}
	data, _ := json.Marshal(commands)
	// Ask the platform, rather than assuming glibc's numbering on other Linux C libraries.
	limits := execute(t, "python3", "-c", "import signal; print(signal.SIGRTMIN, signal.SIGRTMAX)")
	var first, last int
	if _, err := fmt.Sscan(string(limits.stdout), &first, &last); err != nil {
		t.Fatal(err)
	}
	for stop := first; stop <= last; stop++ {
		t.Run(strconv.Itoa(stop), func(t *testing.T) {
			t.Parallel()
			result := execute(t, "python3", "-I", "testdata/output_edges/probe.py", "realtime", string(data), strconv.Itoa(stop))
			t.Logf("%s", result.stdout)
			if result.exitCode != 0 {
				t.Fatalf("realtime probe: exit %d\n%s", result.exitCode, result.stderr)
			}
		})
	}
}

// UBSan's null-store check would exit before the hardware fault. Disable that check only in
// crash, leaving ASan enabled, so this tests the sanitizer's SIGSEGV handler after startup.
func TestStartupPreservesSanitizerSegvReport(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "fault")
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
		t.Fatal(err)
	}
	report := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:handle_segv=1"}, binary)
	if report.exitCode == 0 || !bytes.Contains(report.stderr, []byte("ERROR: AddressSanitizer: SEGV")) ||
		!bytes.Contains(report.stderr, []byte("#0")) || !bytes.Contains(report.stderr, []byte("crash")) {
		t.Fatalf("want sanitizer SEGV report with crash stack after adamic_start: exit %d, stderr %s", report.exitCode, report.stderr)
	}
	t.Logf("sanitizer report: exit %d\n%s", report.exitCode, report.stderr)
}
