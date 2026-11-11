package native

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A timer warning names the process that printed it, as Node's does: (node:<pid>) is the running
// program's own id. The oracle's timeoutWarningLine rule makes those digits a token (#zhgmbwz), so no
// family checks them; this reads them against the id of the process that ran (Correctness's should,
// #7h02qb7). Mutant: timer_warn prints a fixed 1 in adamic_node_pid's place, and this goes red while
// the timers_source families stay green.
func TestTimerWarningNamesItsOwnProcess(t *testing.T) {
	t.Parallel()
	const source = "import 'node:timers';\nconsole.log('before');\nsetTimeout(() => { console.log('fired'); }, Number('not a number'));\nconsole.log('scheduled');\n"
	directory := t.TempDir()
	path := filepath.Join(directory, "delay_nan.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "delay_nan")
	if err := Build(C(program), binary, Options{}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := programCommand(ctx, directory, binary)
	stdout, stderr, err := childguard.Capture(ctx, command)
	if err != nil {
		t.Fatalf("%v\n%s", err, stderr)
	}
	if string(stdout) != "before\nscheduled\nfired\n" {
		t.Fatalf("stdout %q", stdout)
	}
	match := regexp.MustCompile(`^\(node:([0-9]+)\) TimeoutNaNWarning: NaN is not a number\.\n`).FindSubmatch(stderr)
	if match == nil {
		t.Fatalf("no warning line first on stderr: %q", stderr)
	}
	if printed, err := strconv.Atoi(string(match[1])); err != nil || printed != command.Process.Pid {
		t.Fatalf("the warning names process %s, the program ran as %d", match[1], command.Process.Pid)
	}
}
