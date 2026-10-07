package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Not parallel: the subprocess repeats the same end-to-end comparison.
func TestEmittedJavaScriptMismatch(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_MISMATCH_PROBE") == "1" {
		directory, err := filepath.Abs(".")
		if err != nil {
			t.Fatal(err)
		}
		path := manifest(t, []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"})
		oracle := goOracle(t)
		binary := buildPort(t, directory, true)
		module := emittedJavaScript(t, directory)
		compareWithJavaScript(t, oracle, binary, directory, path, module)
		data, err := os.ReadFile(module)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, []byte("\nconsole.log('planted emitted JavaScript mismatch');\n")...)
		if err := os.WriteFile(module, data, 0644); err != nil {
			t.Fatal(err)
		}
		compareWithJavaScript(t, oracle, binary, directory, path, module)
		t.Fatal("emitted JavaScript mutant survived")
	}
	log := filepath.Join(t.TempDir(), "mismatch.log")
	output, err := os.Create(log)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(os.Args[0], "-test.run=^TestEmittedJavaScriptMismatch$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_LINT_MISMATCH_PROBE=1")
	command.Stdout, command.Stderr = output, output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if runError == nil || !bytes.Contains(data, []byte("emitted JavaScript:")) || !bytes.Contains(data, []byte("planted emitted JavaScript mismatch")) {
		t.Fatalf("wrong mutant failure: %v\n%s", runError, data)
	}
	t.Logf("ordinary comparison rejected clean-running emitted JavaScript mutant:\n%s", data)
}

func TestDotARename(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"})
	oracle := goOracle(t)
	want := compare(t, oracle, buildPort(t, directory, true), directory, path)
	copied := mutant(t, "", "")
	entry := filepath.Join(copied, "rules/no-var/rule.a")
	before, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(copied, "rules/no-var/rule.ts")
	if err := os.Rename(entry, renamed); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(renamed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("rename changed module bytes")
	}
	got := compare(t, oracle, buildPort(t, copied, true), copied, path)
	if !bytes.Equal(got, want) {
		t.Fatal("rename changed results")
	}
	t.Logf("rename only: .ts and .a identical on all three runtimes against Go (%d bytes)", len(want))
}
