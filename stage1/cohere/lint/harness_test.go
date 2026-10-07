package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func serializationPort(t *testing.T) string {
	directory := mutant(t, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	return directory
}

func TestCompleteSuggestionSerialization(t *testing.T) {
	directory := serializationPort(t)
	source := filepath.Join(t.TempDir(), "suggestions.ts")
	if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source + "\tno-debugger"})
	oracle := goOracleFrom(t, directory)
	want := compare(t, oracle, buildPort(t, directory, true), directory, path)
	for _, field := range []string{"suggestion\tfirst", "suggestion\tsecond", "suggestion\tempty", "suggestion-edit\t8 9", "fixed\t/*"} {
		if !bytes.Contains(want, []byte(field)) {
			t.Fatalf("missing field %q: %s", field, want)
		}
	}
	changed := filepath.Join(directory, "rules/no-debugger/rule.a")
	data, err := os.ReadFile(changed)
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte("start + 1, start + 2, ''"), []byte("start + 1, start + 3, ''"), 1)
	if err := os.WriteFile(changed, data, 0644); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
		{"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("second suggestion edit mutant survived on %s", side.name)
		}
		t.Logf("second suggestion edit mutant caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	directory := serializationPort(t)
	for _, change := range []struct{ name, from, to string }{
		{"rule.a", "debuggerMessage, '', '', ''", "debuggerMessage, 'fix', ';', ''"},
		{"oracle.go", "d.Fixes = nil", "d.Fixes[0].Text = \";\""},
	} {
		path := filepath.Join(directory, "rules/no-debugger", change.name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Count(data, []byte(change.from)) != 1 {
			t.Fatal("automatic fix anchor changed")
		}
		if err := os.WriteFile(path, bytes.Replace(data, []byte(change.from), []byte(change.to), 1), 0644); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(t.TempDir(), "mixed.ts")
	if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source + "\tno-debugger"})
	got := compare(t, goOracleFrom(t, directory), buildPort(t, directory, true), directory, path)
	if !bytes.Contains(got, []byte("fixed\t/*\\ud83d\\ude00*/;\\u000a")) {
		t.Fatalf("automatic fix lost: %s", got)
	}
	t.Log("automatic fix remains applied while all three suggestions remain unapplied and serialized")
}

func TestWitnessScriptKind(t *testing.T) {
	directory := mutant(t, "", "")
	witness := filepath.Join(directory, "rules/no-debugger/testdata/witness.ts.txt")
	if err := os.Rename(witness, strings.TrimSuffix(witness, ".ts.txt")+".tsx.txt"); err != nil {
		t.Fatal(err)
	}
	witness = strings.TrimSuffix(witness, ".ts.txt") + ".tsx.txt"
	if err := os.WriteFile(witness, []byte("const node = 1; debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sources := ownedWitnesses(t, directory, "no-debugger")
	if filepath.Ext(sources[0]) != ".tsx" {
		t.Fatal("witness script kind lost")
	}
	compare(t, goOracleFrom(t, directory), buildPort(t, directory, true), directory, manifest(t, []string{sources[0] + "\tno-debugger"}))
}
