package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
	"os"
	"os/exec"
	"path/filepath"
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
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "helpers")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
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
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/oracle.go")
	exports, _ := filepath.Abs("testdata/exports.go")
	virtual := filepath.Join(root, "adamic_comments_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/comments/adamic_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
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

// Not parallel: compile each mutant separately to bound clang and parser memory.
func TestCommentMutants(t *testing.T) {
	path, _ := filepath.Abs("testdata/witnesses.json")
	want := run(t, "", oracle(t), path)
	for _, m := range []struct{ file, old, new string }{
		{"can_begin_at.ts", "if(position === 0 && text.startsWith('#!'))", "if(false)"},
		{"collect_list_interiors.ts", "if(depth === 1)", "if(false)"},
		{"sort_by_position.ts", "> current.start", "< current.start"},
		{"all.ts", "anchors[node.end] = true;", "anchors[node.end] = false;"},
		{"for_file.ts", "if(!this.ready)", "if(true)"},
	} {
		t.Run(m.file, func(t *testing.T) {
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
			got := run(t, "", build(t, directory), path)
			if bytes.Equal(got, want) {
				t.Fatal("compiled semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, a[i], b[i])
					break
				}
			}
		})
	}
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
