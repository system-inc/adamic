package lint

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The probe runs in a subprocess of this test binary, so the planted mismatch can fail it without failing
// this run.
func TestEmittedJavaScriptMismatch(t *testing.T) {
	t.Parallel()
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
		// The emitted module is the run's shared one, so the planted mismatch goes into a copy of it.
		module = filepath.Join(t.TempDir(), "lint.mjs")
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
	t.Parallel()
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
	t.Parallel()
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
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("second suggestion edit mutant survived on %s", side.name)
		}
		t.Logf("second suggestion edit mutant caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	t.Parallel()
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

const testWitnessScriptKindShards = 1

// ADAMIC_TEST_SHARD=i/n selects deterministic case shards (zero-based); unset runs all.
// The gate invokes each unit as -run '^TestWitnessScriptKind$/^shard-NNN$'.
// Until internal/buildcache lands, setup builds its products once and shares them;
// its measured time includes those builds.
// Build products are prepared once before parallel units. The original enumeration has one case.
func TestWitnessScriptKind(t *testing.T) {
	t.Parallel()
	products := witnessProducts(t)
	ids := []string{"no-debugger-tsx"}
	shards := checkedCaseShards(t, ids, len(ids))
	if len(shards) != testWitnessScriptKindShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testWitnessScriptKindShards)
	}
	for i, cases := range shards {
		if !selectedCaseShard(t, i) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			t.Logf("case ids: %v", cases)
			for range cases {
				if filepath.Ext(products.Source) != ".tsx" {
					t.Fatal("witness script kind lost")
				}
				compareWithJavaScript(t, products.Oracle, products.Native, products.Directory,
					manifest(t, []string{products.Source + "\tno-debugger"}), products.JavaScript)
			}
			if os.Getenv("ADAMIC_LINT_WITNESS_PRODUCTS") == "" {
				t.Run("planted-disagreement", func(t *testing.T) {
					checkWitnessPlantedDisagreement(t, products)
				})
			}
		})
	}
}
