package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// meter-daily's outcome_test.py, run here so a gate runs it: a suite only a hand runs can't fail (#mpesmv7, Correctness
// on #ny0c3gb). It holds which outcome run.sh records for each meter exit, OverBudget and BudgetUnknown apart from
// Failed, and which outcomes submit.sh commits. It needs nothing but python3: temporary report.json files, no network.
// slim_test.py and records_test.py run beside it (#jnx6n47): slim.py's record holds the summary and the rows readers
// read and refuses a log or a file over 1 MB, and no committed record after the last full one holds either.
func TestMeterDailyOutcomes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "meter-daily-outcomes.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// -B writes no bytecode, and -X pycache_prefix reads it only under an empty directory of the test's own, never a
	// __pycache__ in the tree, which no key holds; -s leaves the user site, and a usercustomize there, off sys.path.
	command := toolchain.Command("python3", "-B", "-s", "-X", "pycache_prefix="+t.TempDir(), "-m", "unittest", "-v", "outcome_test", "slim_test", "records_test")
	command.Dir = "meter-daily"
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("meter-daily outcome tests: %v\n%s", runError, data)
	}
	t.Logf("meter-daily outcome tests passed:\n%s", data)
}

// meter-daily/hand-jobs.sh decides whether the daily meter may start on a line box: never beside a Loom hand job,
// which is a child of the box's keeper, `loom control keep` (Loom's condition for the timer on Pupa, Nov 9). Each
// case runs the script under sh with pgrep as a shell function, reading what the case says the process table holds,
// and sleep as one that returns at once, so the 30 minute wait is two one-second reads here.
//
// Mutants: waitForHandJobs returning 0 while the box is held (the held case starts beside the job); handJobHeld
// always false (the same); the script's own variables named step again (run.sh's status step comes back as 1); the
// pgrep function taken out (the box's own pgrep answers, and on a box whose keeper runs, no keeper case reads it).
func TestMeterDailyNeverStartsBesideAHandJob(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, keeper string
		heldReads    int
		wantCode     string
		wantSaid     string
	}{
		{"no keeper", "", 0, "0", "no hand-job keeper ('loom control keep') runs on this box"},
		{"a free box", "4242", 0, "0", "no hand job holds this box (keeper 4242, waited 0 s): running"},
		{"held, then free", "4242", 1, "0", "no hand job holds this box (keeper 4242, waited 1 s): running"},
		{"held throughout", "4242", 1000, "75", "a hand job still holds this box after 2 s (keeper 4242): skipping today's run"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			said := runHandJobWait(t, test.keeper, test.heldReads)
			if !strings.Contains(said, "code="+test.wantCode+" step=wait\n") {
				t.Errorf("want exit %s with run.sh's step still wait, got:\n%s", test.wantCode, said)
			}
			if !strings.Contains(said, test.wantSaid) {
				t.Errorf("want %q said, got:\n%s", test.wantSaid, said)
			}
		})
	}
}

// runHandJobWait runs waitForHandJobs with a fake process table: keeper is the pid pgrep -u answers (none when
// empty), and the keeper has a child for the first heldReads reads of pgrep -P. It returns what the script said, then
// its exit and run.sh's step variable as code=<n> step=<step>.
//
// pgrep and sleep are shell functions the script defines before it sources hand-jobs.sh, never files on PATH. A
// function wins over every PATH entry, so the box's own pgrep is never reached, and nothing is written and then run,
// so there's no exec to fail while another case forks. Fakes written as executables did fail that way: on Nov 16 a
// change that never touched cloud/ went red here (chg_redsxg53), its keeper case reading Eve's real keeper.
func runHandJobWait(t *testing.T, keeper string, heldReads int) string {
	t.Helper()
	counter := filepath.Join(t.TempDir(), "held-reads")
	if err := os.WriteFile(counter, []byte(strings.Repeat("x", heldReads)), 0o644); err != nil {
		t.Fatal(err)
	}
	script := `pgrep() {
	case "$1" in
	-u) [ -n "$FAKE_KEEPER" ] || return 1; echo "$FAKE_KEEPER" ;;
	-P) left=$(cat "$FAKE_HELD_READS"); [ -n "$left" ] || return 1; printf '%s' "${left#x}" > "$FAKE_HELD_READS"; echo 4343 ;;
	*) return 2 ;;
	esac
}
sleep() { return 0; }
. ./meter-daily/hand-jobs.sh
say() { printf '%s\n' "$*"; }
step=wait
code=0
waitForHandJobs say || code=$?
printf 'code=%s step=%s\n' "$code" "$step"
`
	command := toolchain.Command("sh", "-c", script)
	if command.Err != nil {
		t.Fatal(command.Err)
	}
	environment := command.Env
	if environment == nil {
		environment = os.Environ()
	}
	command.Env = append(environment, "FAKE_KEEPER="+keeper, "FAKE_HELD_READS="+counter,
		"METER_HAND_JOB_WAIT=2", "METER_HAND_JOB_STEP=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("sh: %v\n%s", err, output)
	}
	return string(output)
}
