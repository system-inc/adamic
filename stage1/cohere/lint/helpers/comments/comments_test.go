package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
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
func build(t *testing.T, directory string) string {
	t.Helper()
	start := time.Now()
	defer func() { t.Logf("build native %s wall %.6fs", directory, time.Since(start).Seconds()) }()
	buildDirectory := t.TempDir()
	binary := filepath.Join(buildDirectory, "helpers")
	// Inputs: directory/main.ts and its transitive imports; native runtime; Go and clang toolchains; Sanitize=true.
	product := func(dir string) error {
		stage := time.Now()
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		t.Logf("build load wall %.6fs", time.Since(stage).Seconds())
		if err != nil {
			return err
		}
		stage = time.Now()
		ir, err := lower.Lower(context.Background(), program)
		t.Logf("build lower wall %.6fs", time.Since(stage).Seconds())
		if err != nil {
			return err
		}
		stage = time.Now()
		source := native.C(ir)
		if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(source), 0644); err != nil {
			return err
		}
		t.Logf("build emit wall %.6fs", time.Since(stage).Seconds())
		stage = time.Now()
		err = native.Build(source, filepath.Join(dir, "helpers"), native.Options{Sanitize: true})
		t.Logf("build clang wall %.6fs", time.Since(stage).Seconds())
		return err
	}
	if err := product(buildDirectory); err != nil {
		t.Fatal(err)
	}
	return binary
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

func oracle(t *testing.T) string {
	t.Helper()
	start := time.Now()
	defer func() { t.Logf("build Go oracle wall %.6fs", time.Since(start).Seconds()) }()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/oracle.go")
	exports, _ := filepath.Abs("testdata/exports.go")
	virtual := filepath.Join(root, "adamic_comments_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/comments/adamic_exports.go"): exports}})
	// Inputs: testdata/oracle.go, testdata/exports.go, pinned cohere and
	// TypeScript modules; -overlay; the Go toolchain. No package-local cache.
	directory := t.TempDir()
	product := func(dir string) error {
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := exec.Command("go", "build", "-overlay="+path, "-o", filepath.Join(dir, "go-oracle"), virtual)
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
func TestCommentsMatchCohere(t *testing.T) {
	path, _ := filepath.Abs("testdata/witnesses.json")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	want := run(t, "", oracle(t), path)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", build(t, "."), path), want)
	t.Logf("Go, Node and sanitized native match %d output lines", bytes.Count(want, []byte("\n")))
}

// ADAMIC_TEST_SHARD=i/n (zero-based i) selects deterministic mutant shards; unset runs all.
// Builds are inputs, prepared once before the parallel units. Every unit runs
// the entire witness corpus under the original sanitizers and leak checks.
func TestCommentMutants(t *testing.T) {
	path, _ := filepath.Abs("testdata/witnesses.json")
	// The runtime archive is a separate hashed input shared by all mutant builds.
	start := time.Now()
	if _, err := native.RuntimeLibrary("", native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	t.Logf("build sanitized runtime archive wall %.6fs", time.Since(start).Seconds())
	want := run(t, "", oracle(t), path)
	mutants := []struct{ file, old, new string }{
		{"can_begin_at.ts", "if(position === 0 && text.startsWith('#!'))", "if(false)"},
		{"collect_list_interiors.ts", "if(depth === 1)", "if(false)"},
		{"sort_by_position.ts", "> current.start", "< current.start"},
		{"all.ts", "anchors[node.end] = true;", "anchors[node.end] = false;"},
		{"for_file.ts", "if(!this.ready)", "if(true)"},
	}
	ids := make([]string, len(mutants))
	binaries := make([]string, len(mutants))
	for i, m := range mutants {
		ids[i] = m.file
	}
	checkCommentMutantUnion(t, ids, path)
	for i, m := range mutants {
		directory := t.TempDir()
		for _, file := range []string{"main.ts", "comment.ts", "can_begin_at.ts", "collect_list_interiors.ts", "sort_by_position.ts", "all.ts", "for_file.ts"} {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if file == m.file {
				if strings.Count(text, m.old) != 1 {
					t.Fatalf("mutant anchor count %d", strings.Count(text, m.old))
				}
				text = strings.Replace(text, m.old, m.new, 1)
			}
			parserRoot, _ := filepath.Abs("../../../../typescript")
			options, _ := filepath.Abs("../options_json.ts")
			text = strings.ReplaceAll(text, "../../../../typescript", filepath.ToSlash(parserRoot))
			text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(options))
			if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
		}
		binaries[i] = build(t, directory)
	}
	runCommentMutantShards(t, ids, func(t *testing.T, i int) {
		got := run(t, "", binaries[i], path)
		requireCommentMutantKilled(t, got, want)
		a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, a[i], b[i])
				break
			}
		}
	})
}

// The inventory assumes a common AST adapter. This comparison isolates that
// contract from the separate stage-1 parser, using Go's actual traversal spans.
func TestConsumerCommentHelpers(t *testing.T) {
	original, _ := filepath.Abs("testdata/consumers.json")
	goOracle := oracle(t)
	adapted := run(t, "", goOracle, "--ast", original)
	path := filepath.Join(t.TempDir(), "adapted.json")
	if err := os.WriteFile(path, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	want := run(t, "", goOracle, original)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", build(t, "."), path), want)
	t.Logf("Go, Node and sanitized native agree on %d consumer output lines with the stated AST adapter", bytes.Count(want, []byte("\n")))
}

func TestJsxParserGapIsExplicit(t *testing.T) {
	data, err := os.ReadFile("testdata/parser-gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	binary := build(t, ".")
	goOracle := oracle(t)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	for _, row := range rows {
		corpus, _ := json.Marshal([]any{row})
		path := filepath.Join(t.TempDir(), "gap.json")
		if err := os.WriteFile(path, corpus, 0644); err != nil {
			t.Fatal(err)
		}
		adapted := run(t, "", goOracle, "--ast", path)
		var facts []struct{ GoDiagnostics int }
		if err := json.Unmarshal(adapted, &facts); err != nil {
			t.Fatal(err)
		}
		if len(facts) != 1 || facts[0].GoDiagnostics != 0 {
			t.Fatal("gap must be valid Go TypeScript")
		}
		var previous string
		for _, command := range [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, entry, path}, {binary, path}} {
			_, stderr, err := gapRun(t, command)
			if !matchesGapRefusal(err, stderr) {
				t.Fatalf("gap silently accepted or failed differently: %v %s", err, stderr)
			}
			if previous != "" && previous != stderr {
				t.Fatal("Node and native parser refusals differ")
			}
			previous = stderr
		}
	}
	t.Logf("%d valid Go JSX inputs explicitly refuse in the independent-parser adapter; AST-adapter helper comparisons cover them separately", len(rows))
}

// Removing the TSX adapter guard must compile and accept the known wrong parse,
// rather than merely crashing on one of the other unsupported JSX forms.
func TestJsxAdapterGuardMutant(t *testing.T) {
	directory := t.TempDir()
	for _, file := range []string{"main.ts", "comment.ts", "can_begin_at.ts", "collect_list_interiors.ts", "sort_by_position.ts", "all.ts", "for_file.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == "main.ts" {
			old := "if(adapted < 0 && name.endsWith('.tsx'))"
			if strings.Count(text, old) != 1 {
				t.Fatal("guard mutant anchor changed")
			}
			text = strings.Replace(text, old, "if(false)", 1)
		}
		parserRoot, _ := filepath.Abs("../../../../typescript")
		options, _ := filepath.Abs("../options_json.ts")
		text = strings.ReplaceAll(text, "../../../../typescript", filepath.ToSlash(parserRoot))
		text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(options))
		if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile("testdata/parser-gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	corpus, _ := json.Marshal(rows[:1])
	path := filepath.Join(t.TempDir(), "gap.json")
	if err := os.WriteFile(path, corpus, 0644); err != nil {
		t.Fatal(err)
	}
	output, stderr, err := gapRun(t, []string{build(t, directory), path})
	if err != nil || stderr != "" {
		t.Fatalf("mutant must finish normally, not crash: %v %s", err, stderr)
	}
	if matchesGapRefusal(err, stderr) {
		t.Fatal("compiled adapter-guard mutant survived the actual refusal check")
	}
	if !bytes.Contains(output, []byte("comments ")) {
		t.Fatal("guard mutant did not finish the wrong parsed tree")
	}
	t.Log("compiled adapter-guard mutant finishes the unsupported JSX input with exit 0; the explicit-refusal check requires exit 70 and catches it")
}

func matchesGapRefusal(err error, stderr string) bool {
	exit, ok := err.(*exec.ExitError)
	return ok && exit.ExitCode() == 70 && strings.Contains(stderr, "NotYet: stage-1 JSX parser adapter")
}
func gapRun(t *testing.T, command []string) ([]byte, string, error) {
	t.Helper()
	cmd := exec.Command(command[0], command[1:]...)
	out, err := os.CreateTemp(t.TempDir(), "gap-output-")
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	runErr := testguard.Run(cmd, time.Minute, testguard.Ceiling)
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	return output, stderr.String(), runErr
}

func requireCommentMutantKilled(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		t.Fatal("compiled semantic mutant survived")
	}
}

func commentMutantSelected(t *testing.T, index int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q: want zero-based i/n", value)
	}
	selected, first := strconv.Atoi(parts[0])
	count, second := strconv.Atoi(parts[1])
	if first != nil || second != nil || count < 1 || selected < 0 || selected >= count {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index%count == selected
}

func runCommentMutantShards(t *testing.T, ids []string, check func(*testing.T, int)) {
	t.Helper()
	for i, id := range ids {
		if !commentMutantSelected(t, i) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%02d-%s", i, id), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: wall %s exceeds 30s", elapsed)
				}
			}()
			check(t, i)
		})
	}
}

// Count and compare the full mutant/witness Cartesian product, independently
// of gate selection, so selecting one box never hides an enumeration error.
func checkCommentMutantUnion(t *testing.T, ids []string, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Name string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	expected := make(map[string]bool)
	for _, id := range ids {
		for rowIndex, row := range rows {
			key := fmt.Sprintf("%s/%04d/%s", id, rowIndex, row.Name)
			if row.Name == "" || expected[key] {
				t.Fatalf("repeated or empty unsplit case id %q", key)
			}
			expected[key] = true
		}
	}
	actual := make(map[string]bool)
	count := 0
	for shard := range ids {
		for index, id := range ids {
			if index != shard {
				continue
			}
			for rowIndex, row := range rows {
				key := fmt.Sprintf("%s/%04d/%s", id, rowIndex, row.Name)
				if actual[key] || !expected[key] {
					t.Fatalf("repeated or unexpected shard case %q", key)
				}
				actual[key] = true
				count++
			}
		}
	}
	if count != len(expected) || count != len(ids)*len(rows) {
		t.Fatalf("union count %d, unsplit %d", count, len(expected))
	}
	for key := range expected {
		if !actual[key] {
			t.Fatalf("missing shard case %q", key)
		}
	}
	t.Logf("union: %d mutants x %d witnesses = %d unique cases across %d shards", len(ids), len(rows), count, len(ids))
}

// A prepared output equal to the oracle plants a survivor in exactly one unit.
// Subprocesses exercise the same parallel runner and fatal check as the real test.
func TestCommentMutantShardCatchesSurvivor(t *testing.T) {
	ids := []string{"can_begin_at.ts", "collect_list_interiors.ts", "sort_by_position.ts", "all.ts", "for_file.ts"}
	if os.Getenv("ADAMIC_COMMENT_SURVIVOR_PROBE") == "1" {
		runCommentMutantShards(t, ids, func(t *testing.T, i int) {
			got := []byte("killed")
			if i == 2 {
				got = []byte("oracle")
			}
			requireCommentMutantKilled(t, got, []byte("oracle"))
		})
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for i := range ids {
		command := exec.Command(executable, "-test.run=^TestCommentMutantShardCatchesSurvivor$", "-test.v")
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, "ADAMIC_COMMENT_SURVIVOR_PROBE=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "ADAMIC_COMMENT_SURVIVOR_PROBE=1", fmt.Sprintf("ADAMIC_TEST_SHARD=%d/%d", i, len(ids)))
		output, err := command.CombinedOutput()
		failed := err != nil
		if failed != (i == 2) {
			t.Fatalf("shard %d: unexpected result %v: %s", i, err, output)
		}
		name := fmt.Sprintf("shard-%02d-%s", i, ids[i])
		if !bytes.Contains(output, []byte(name)) {
			t.Fatalf("shard name missing: %s", output)
		}
		if failed && !bytes.Contains(output, []byte("compiled semantic mutant survived")) {
			t.Fatalf("wrong failure: %s", output)
		}
		t.Logf("%s: planted survivor caught=%v", name, failed)
	}
}
