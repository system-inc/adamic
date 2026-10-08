package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: the upstream capture and oracle are shared for this run.
func TestCapturedOracleRecovery(t *testing.T) {
	oracle := goOracle(t)
	rows := upstream(t)
	found := false
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		source, err := os.ReadFile(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		if fields[1] != "no-octal-escape" || string(source) != "' \\01'" {
			continue
		}
		found = true
		answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row})).output
		for _, want := range []string{"octalEscapeSequence", "recovery findings only"} {
			if !bytes.Contains(answer, []byte(want)) {
				t.Fatalf("%s: missing %s: %s", fields[0], want, answer)
			}
		}
		if bytes.Contains(answer, []byte("fixed\t")) {
			t.Fatalf("%s: invalid source claimed fixed", fields[0])
		}
	}
	if !found {
		t.Fatal("upstream no-octal-escape fixture ' \\01' absent")
	}
	// This marker is still a port refusal; Go can judge the recovered tree.
	source := filepath.Join(t.TempDir(), "unsupported.ts")
	if err := os.WriteFile(source, []byte("interface I"), 0644); err != nil {
		t.Fatal(err)
	}
	row := source + "\t@typescript-eslint/method-signature-style\t\t\t\t\tunsupported-recovery"
	answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row})).output
	if !bytes.Contains(answer, []byte("recovery findings only")) || bytes.Contains(answer, []byte("fixed\t")) {
		t.Fatalf("%s: wrong recovery boundary: %s", source, answer)
	}
	t.Logf("captured %d upstream cases; reported octal fixture and unsupported Go recovery replay without fixing", len(rows))
}

func TestUnmarkedMalformedOracleInputStillFails(t *testing.T) {
	source := filepath.Join(t.TempDir(), "unmarked.ts")
	if err := os.WriteFile(source, []byte("'\\1'"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source + "\tno-octal-escape"})
	command := exec.Command(goOracle(t), "--manifest", path)
	log := filepath.Join(t.TempDir(), "guard.log")
	output, err := os.Create(log)
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout, command.Stderr = output, output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if runError == nil || !bytes.Contains(data, []byte("panic: invalid corpus "+source)) {
		t.Fatalf("unmarked input guard did not fail by name: %v\n%s", runError, data)
	}
}

func TestRecoveryClassificationPreservesModes(t *testing.T) {
	rows := []string{"clean.ts\trule", "octal.ts\trule", "unsupported.ts\trule\t\t\t\t\tunsupported-recovery"}
	got, err := classifyRecoveryRows(rows, []string{"0", "1", "1"})
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != rows[0] || !strings.HasSuffix(got[1], "\trecovery") || got[2] != rows[2] {
		t.Fatalf("recovery classification changed a caller's boundary: %q", got)
	}
	if _, err := classifyRecoveryRows(rows, []string{"0"}); err == nil {
		t.Fatal("missing diagnostic flags were accepted")
	}
}
