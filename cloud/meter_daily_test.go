package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// meter-daily/hand-jobs.sh decides whether the daily meter may start on a line box: never beside a Loom hand job,
// which is a child of the box's keeper, `loom control keep` (Loom's condition for the timer on Pupa, Nov 9). Each
// case runs the script under sh with a fake pgrep, reading what the case says the process table holds, and a sleep
// that returns at once, so the 30 minute wait is two one-second reads here.
//
// Mutants: waitForHandJobs returning 0 while the box is held (the held case starts beside the job); handJobHeld
// always false (the same); the script's own variables named step again (run.sh's status step comes back as 1).
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
func runHandJobWait(t *testing.T, keeper string, heldReads int) string {
	t.Helper()
	directory := t.TempDir()
	bin := filepath.Join(directory, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	counter := filepath.Join(directory, "held-reads")
	if err := os.WriteFile(counter, []byte(strings.Repeat("x", heldReads)), 0o644); err != nil {
		t.Fatal(err)
	}
	pgrep := `#!/bin/sh
case "$1" in
-u) [ -z "$FAKE_KEEPER" ] && exit 1; echo "$FAKE_KEEPER" ;;
-P) left=$(cat "$FAKE_HELD_READS"); [ -n "$left" ] || exit 1; printf '%s' "${left#x}" > "$FAKE_HELD_READS"; echo 4343 ;;
*) exit 2 ;;
esac
`
	for name, body := range map[string]string{"pgrep": pgrep, "sleep": "#!/bin/sh\nexit 0\n"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	script := `. ./meter-daily/hand-jobs.sh
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
	path := os.Getenv("PATH")
	for _, variable := range environment {
		if value, found := strings.CutPrefix(variable, "PATH="); found {
			path = value
		}
	}
	command.Env = append(environment, "PATH="+bin+":"+path, "FAKE_KEEPER="+keeper, "FAKE_HELD_READS="+counter,
		"METER_HAND_JOB_WAIT=2", "METER_HAND_JOB_STEP=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("sh: %v\n%s", err, output)
	}
	return string(output)
}
