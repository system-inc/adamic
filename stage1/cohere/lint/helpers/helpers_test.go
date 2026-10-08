package helpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	f, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = childguard.Run(cmd, childguard.Options{}); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func oracle(t *testing.T) string {
	t.Helper()
	start := time.Now()
	defer func() { t.Logf("build Go helper oracle wall %.6fs", time.Since(start).Seconds()) }()
	root, _ := filepath.Abs("../../../../cohere")
	side, _ := filepath.Abs("testdata/oracle.go")
	catalog, _ := filepath.Abs("testdata/catalog.go")
	descriptors, _ := filepath.Abs("testdata/descriptors.go")
	virtual := filepath.Join(root, "adamic_helper_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side, filepath.Join(root, "policy/adamic_helpers.go"): catalog, filepath.Join(root, "adamic_helper_descriptors.go"): descriptors}})
	// Inputs: Name=helper-Go-oracle; Files=oracle.go, catalog.go,
	// descriptors.go and pinned cohere/TypeScript dependencies; Flags=-overlay;
	// Toolchain=Go. Product outputs are shared before parallel units.
	directory := t.TempDir()
	product := func(dir string) error {
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := exec.Command("go", "build", "-overlay="+path, "-o", filepath.Join(dir, "go-oracle"), virtual, filepath.Join(root, "adamic_helper_descriptors.go"))
		command.Dir = root
		output, err := childguard.CombinedOutput(command, childguard.Options{})
		if err != nil || len(output) != 0 {
			return fmt.Errorf("Go oracle build: %v: %s", err, output)
		}
		return nil
	}
	if err := product(directory); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "go-oracle")
	return binary
}
func build(t *testing.T, directory string) string {
	t.Helper()
	start := time.Now()
	defer func() { t.Logf("build native helper %s wall %.6fs", directory, time.Since(start).Seconds()) }()
	buildDirectory := t.TempDir()
	binary := filepath.Join(buildDirectory, "helpers")
	// Inputs: Name=helper-sanitized-native; Files=directory/main.ts and all
	// transitive imports, compiler sources and native runtime; Flags=Sanitize;
	// Toolchain=Go and clang. No package-local build cache.
	product := func(dir string) error {
		stage := time.Now()
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		source := native.C(ir)
		if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(source), 0644); err != nil {
			return err
		}
		t.Logf("build lowered helper program wall %.6fs", time.Since(stage).Seconds())
		stage = time.Now()
		err = native.Build(source, filepath.Join(dir, "helpers"), native.Options{Sanitize: true})
		t.Logf("build sanitized helper binary wall %.6fs", time.Since(stage).Seconds())
		return err
	}
	if err := product(buildDirectory); err != nil {
		t.Fatal(err)
	}
	return binary
}
func TestHelpersMatchCohere(t *testing.T) {
	goOracle := oracle(t)
	catalog := run(t, "", goOracle, "catalog")
	descriptor := run(t, "", goOracle, "descriptors")
	pinned, err := os.ReadFile("testdata/descriptors.json")
	if err != nil || !bytes.Equal(descriptor, pinned) {
		t.Fatal("regenerate target descriptors against Go cohere")
	}
	current, err := os.ReadFile("testdata/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(catalog, current) {
		t.Fatal("regenerate catalog against pinned Go cohere")
	}
	cases := fixture(t)
	catalogPath, _ := filepath.Abs("testdata/catalog.json")
	want := run(t, "", goOracle, cases)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	node := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, catalogPath)
	compare(t, node, want)
	compare(t, run(t, "", build(t, "."), cases, catalogPath), want)
	t.Logf("%d oracle output lines matched Go, Node and sanitized native", len(strings.Split(strings.TrimSpace(string(want)), "\n")))
}
func compare(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("output line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size: got %d Go %d", len(got), len(want))
}

const testHelperMutantsShards = 4

// ADAMIC_TEST_SHARD=i/n selects units locally; unset runs all. The gate can
// select shard-NNN directly. Every unit keeps the entire case corpus and original
// sanitizer/leak checks. Build inputs are shared until internal/buildcache lands.
func TestHelperMutants(t *testing.T) {
	cases := fixture(t)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	want := run(t, "", oracle(t), cases)
	mutants := []struct{ file, old, new string }{{"options_json.ts", "if(char.charCodeAt(0) < 32)", "if(false)"}, {"option_schema.ts", "matched !== 1", "matched === 0"}, {"policy_message.ts", "text = text.split(`{{${name}}}`).join(value);", "text = text;"}, {"strict_options.ts", "if(field < 0) { return false; }", "if(field < 0) { continue; }"}}
	ids := make([]string, len(mutants))
	binaries := make([]string, len(mutants))
	for i, m := range mutants {
		ids[i] = m.file
	}
	checkHelperMutantUnion(t, ids, cases)
	start := time.Now()
	if _, err := native.RuntimeLibrary("", native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	t.Logf("build sanitized helper runtime archive wall %.6fs", time.Since(start).Seconds())
	for i, m := range mutants {
		t.Logf("build inputs: mutant %d %s", i, m.file)
		directory := t.TempDir()
		for _, file := range []string{"main.ts", "options_json.ts", "option_schema.ts", "policy_message.ts", "strict_options.ts"} {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if file == m.file {
				if strings.Count(string(data), m.old) != 1 {
					t.Fatal("mutant anchor changed")
				}
				data = []byte(strings.Replace(string(data), m.old, m.new, 1))
			}
			if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		binaries[i] = build(t, directory)
	}
	runHelperMutantShards(t, ids, func(t *testing.T, i int) {
		got := run(t, "", binaries[i], cases, catalog)
		requireHelperMutantKilled(t, got, want)
		a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Logf("compiled semantic mutant caught at output line %d: got %q, Go %q", i+1, a[i], b[i])
				break
			}
		}
	})
}

func fixture(t *testing.T) string {
	t.Helper()
	f, err := os.Open("testdata/cases.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	path := filepath.Join(t.TempDir(), "cases.json")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(out, reader); err != nil {
		t.Fatal(err)
	}
	if err = out.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func smallFixture(t *testing.T, corpus any) string {
	t.Helper()
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "cases.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestMessageRefusalsMatchGo(t *testing.T) {
	binary := build(t, ".")
	goOracle := oracle(t)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	cases := []map[string]any{
		{"Kind": "message", "Rule": "missing/rule", "Id": "missing", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "missing", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-stuttering-name", "Id": "stutteringName", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-stuttering-name", "Id": "stutteringName", "Values": map[string]string{"other": "x"}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Values": map[string]string{"extra": "x"}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Values": map[string]string{}, "Choices": []map[string]string{{"Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Phrase": "reason", "Name": "missing"}}},
	}
	phraseRule := "structure/consistency-require-matching-file-name"
	phraseId := "requireMatchingFileName"
	choice := map[string]string{"Rule": phraseRule, "Id": phraseId, "Phrase": "alternative", "Name": "none"}
	for _, choices := range [][]map[string]string{nil, {choice, choice}, {{"Rule": "wrong/rule", "Id": phraseId, "Phrase": "alternative", "Name": "none"}}, {{"Rule": phraseRule, "Id": "wrong", "Phrase": "alternative", "Name": "none"}}} {
		cases = append(cases, map[string]any{"Kind": "message", "Rule": phraseRule, "Id": phraseId, "Values": map[string]string{}, "Choices": choices})
	}
	for i, c := range cases {
		path := smallFixture(t, map[string]any{"Definitions": map[string]string{}, "Cases": []any{c}})
		want := strings.TrimSpace(string(run(t, "", goOracle, path)))
		if !strings.HasPrefix(want, "panic: ") {
			t.Fatalf("guard %d: Go did not refuse", i)
		}
		expected := "adamic: " + want + "\n"
		for _, command := range [][]string{{binary, path, catalog}, {"node", "--disable-warning=ExperimentalWarning", runner, entry, path, catalog}} {
			// Refusals are silent until their final panic; protect startup with FirstOutput.
			cmd := exec.Command(command[0], command[1:]...)
			output, err := os.CreateTemp(t.TempDir(), "stdout-")
			if err != nil {
				t.Fatal(err)
			}
			cmd.Stdout = output
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err = childguard.Run(cmd, childguard.Options{})
			output.Close()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || stderr.String() != expected {
				t.Fatalf("guard %d %s: %v stderr %q, Go %q", i, command[0], err, stderr.String(), expected)
			}
		}
	}
	t.Logf("%d message refusal cases match Go on Node and sanitized native", len(cases))
}
func TestKnownGapsAreExplicit(t *testing.T) {
	corpus := map[string]any{"Definitions": map[string]string{
		"regex":  `{"type":"array","items":{"type":"string","pattern":"x"}}`,
		"custom": `{"kind":"unsupported"}`,
		"fold":   `{"kind":"object","fields":{"Ignore":{"tagged":false,"shape":{"kind":"boolean"}}}}`,
	}, "Cases": []map[string]string{
		{"Kind": "schema", "Schema": "regex", "Input": "[]"},
		{"Kind": "decode", "Schema": "custom", "Input": "null"},
		{"Kind": "decode", "Schema": "fold", "Input": "{\"ſgnore\":true}"},
		{"Kind": "decode", "Schema": "fold", "Input": "{}"},
	}}
	path := smallFixture(t, corpus)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	want := []byte("NotYet: unsupported schema keyword pattern\nNotYet: custom or unsupported Go option type\nNotYet: Unicode fold for untagged option keys\nvalid\n")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path, catalog), want)
	compare(t, run(t, "", build(t, "."), path, catalog), want)
}

func checkHelperMutantUnion(t *testing.T, ids []string, cases string) {
	t.Helper()
	if len(ids) != testHelperMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(ids), testHelperMutantsShards)
	}
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct{ Cases []json.RawMessage }
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	for _, id := range ids {
		for row := range corpus.Cases {
			key := fmt.Sprintf("%s/case-%06d", id, row)
			if id == "" || expected[key] {
				t.Fatalf("repeated or empty unsplit case %q", key)
			}
			expected[key] = true
		}
	}
	actual := map[string]bool{}
	count := 0
	for shard := range ids {
		for index, id := range ids {
			if index != shard {
				continue
			}
			for row := range corpus.Cases {
				key := fmt.Sprintf("%s/case-%06d", id, row)
				if actual[key] || !expected[key] {
					t.Fatalf("repeated or unexpected shard case %q", key)
				}
				actual[key] = true
				count++
			}
		}
	}
	if count != len(expected) || count != testHelperMutantsShards*len(corpus.Cases) {
		t.Fatalf("union count %d differs from unsplit %d", count, len(expected))
	}
	for id := range expected {
		if !actual[id] {
			t.Fatalf("missing shard case %q", id)
		}
	}
	t.Logf("union: %d mutants x %d cases = %d unique IDs across %d shards", len(ids), len(corpus.Cases), count, len(ids))
}

func runHelperMutantShards(t *testing.T, ids []string, check func(*testing.T, int)) {
	t.Helper()
	if len(ids) != testHelperMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(ids), testHelperMutantsShards)
	}
	selected, count := -1, 1
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		var err error
		selected, err = strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		count, err = strconv.Atoi(parts[1])
		if err != nil || count < 1 || selected < 0 || selected >= count {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
	}
	for i, id := range ids {
		if selected >= 0 && i%count != selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %s exceeds 30s", elapsed)
				}
			}()
			t.Logf("mutant %s; entire corpus", id)
			check(t, i)
		})
	}
}

func requireHelperMutantKilled(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		t.Fatal("compiled mutant survived")
	}
}

// Prepared oracle-equal output plants a surviving mutant in exactly one unit.
// Each subprocess uses the gate's exact -run selector and the real fatal check.
func TestHelperMutantShardCatchesSurvivor(t *testing.T) {
	ids := []string{"options_json.ts", "option_schema.ts", "policy_message.ts", "strict_options.ts"}
	if os.Getenv("ADAMIC_HELPER_SURVIVOR_PROBE") == "1" {
		runHelperMutantShards(t, ids, func(t *testing.T, i int) {
			got := []byte("killed")
			if i == 2 {
				got = []byte("oracle")
			}
			requireHelperMutantKilled(t, got, []byte("oracle"))
		})
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for i := range ids {
		name := fmt.Sprintf("shard-%03d", i)
		command := exec.Command(executable, "-test.run=^TestHelperMutantShardCatchesSurvivor$/^"+name+"$", "-test.v")
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, "ADAMIC_HELPER_SURVIVOR_PROBE=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "ADAMIC_HELPER_SURVIVOR_PROBE=1")
		output, err := command.CombinedOutput()
		if (err != nil) != (i == 2) {
			t.Fatalf("%s unexpected result: %v: %s", name, err, output)
		}
		if !bytes.Contains(output, []byte(name)) {
			t.Fatalf("missing shard name: %s", output)
		}
		if err != nil && !bytes.Contains(output, []byte("compiled mutant survived")) {
			t.Fatalf("wrong failure: %s", output)
		}
		t.Logf("%s: planted survivor caught=%v", name, err != nil)
	}
}
