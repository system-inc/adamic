package oracle

import (
	"path/filepath"
	"testing"
)

func TestNodeProcessTmpdir(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_process_tmpdir.a")
	for _, environment := range [][]string{
		{"TMPDIR=", "TMP=", "TEMP="},
		{"TMPDIR=/", "TMP=second", "TEMP=third"},
		{"TMPDIR=//", "TMP=second", "TEMP=third"},
		{"TMPDIR=/tmp///", "TMP=second", "TEMP=third"},
		{"TMPDIR=relative/", "TMP=second", "TEMP=third"},
		{"TMPDIR=héllo 🌍/", "TMP=second", "TEMP=third"},
		{"TMPDIR=", "TMP=second/", "TEMP=third/"},
		{"TMPDIR=", "TMP=", "TEMP=third/"},
		{"TMPDIR=bad\xff/", "TMP=", "TEMP="},
	} {
		runner := filepath.Join(repository, "oracle/node.mjs")
		truth := executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", runner, path)
		for _, got := range []run{executeWith(t, environment, binary), executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", runner, script)} {
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%q: %s, Node %q, got %q %q", environment, difference, truth.stdout, got.stdout, got.stderr)
			}
		}
		t.Logf("%q -> %q", environment, truth.stdout)
	}
	path, mutant := nodeProcessMutant(t, "node_process_tmpdir.a", "if (length > 1 && value[length - 1] == '/') { length--; }", "if (length > 1 && value[length - 1] == '/') { length -= 2; }")
	environment := []string{"TMPDIR=relative/", "TMP=", "TEMP="}
	truth := executeWith(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	bad := executeWith(t, environment, mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatal("tmpdir mutant not caught only by Node bytes")
	}
	t.Log("clean tmpdir mutant caught by Node stdout")
}
