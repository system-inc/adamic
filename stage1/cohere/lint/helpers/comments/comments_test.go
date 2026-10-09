package comments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
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
	if err = childguard.Run(cmd, runGuard); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	captureOutput(t, data, stderr.String(), nil)
	return data
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
	binary := filepath.Join(suiteDirectory, "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestCommentsMatchCohere(t *testing.T) {
	shared := sharedArtifacts(t)
	path, _ := filepath.Abs("testdata/witnesses.json")
	want := run(t, "", shared.oracle, path)
	agreeModes(t, shared.commands, path, want)
	t.Logf("Go, Node, emitted JavaScript and sanitized native match %d output lines", bytes.Count(want, []byte("\n")))
}

// The inventory assumes a common AST adapter. This comparison isolates that
// contract from the separate stage-1 parser, using Go's actual traversal spans.
// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestConsumerCommentHelpers(t *testing.T) {
	shared := sharedArtifacts(t)
	original, _ := filepath.Abs("testdata/consumers.json")
	goOracle := shared.oracle
	adapted := run(t, "", goOracle, "--ast", original)
	path := filepath.Join(t.TempDir(), "adapted.json")
	if err := os.WriteFile(path, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	want := run(t, "", goOracle, original)
	agreeModes(t, shared.commands, path, want)
	t.Logf("Go, Node, emitted JavaScript and sanitized native agree on %d consumer output lines with the stated AST adapter", bytes.Count(want, []byte("\n")))
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestJsxParserGapIsExplicit(t *testing.T) {
	shared := sharedArtifacts(t)
	data, err := os.ReadFile("testdata/parser-gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	goOracle := shared.oracle
	for i, row := range rows {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
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
			// Each mode runs in its own guarded process; wait for all children
			// before comparing their refusal bytes.
			refusals := make([]string, len(shared.commands))
			t.Run("modes", func(t *testing.T) {
				t.Parallel()
				t.Cleanup(func() {
					for i := 1; i < len(refusals); i++ {
						if refusals[i] != refusals[0] {
							t.Fatal("parser refusals differ")
						}
					}
				})
				for i, mode := range shared.commands {
					t.Run(mode.name, func(t *testing.T) {
						t.Parallel()
						_, stderr, err := gapRun(t, mode.withInput(path))
						if !matchesGapRefusal(err, stderr) {
							t.Fatalf("gap silently accepted or failed differently: %v %s", err, stderr)
						}
						refusals[i] = stderr
					})
				}
			})
		})
	}
	t.Logf("%d valid Go JSX inputs explicitly refuse in the independent-parser adapter; AST-adapter helper comparisons cover them separately", len(rows))
}

// Removing the TSX adapter guard must compile and accept the known wrong parse,
// rather than merely crashing on one of the other unsupported JSX forms.
// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestJsxAdapterGuardMutant(t *testing.T) {
	sharedArtifacts(t)
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
	commands := prepare(t, directory, true)
	for _, mode := range commands {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			output, stderr, err := gapRun(t, mode.withInput(path))
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
		})
	}
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
	runErr := childguard.Run(cmd, runGuard)
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	captureOutput(t, output, stderr.String(), runErr)
	return output, stderr.String(), runErr
}

// Optional byte captures make before/after comparisons independent of log order.
func captureOutput(t *testing.T, output []byte, stderr string, runErr error) {
	t.Helper()
	directory := os.Getenv("ADAMIC_COMMENTS_CAPTURE")
	if directory == "" {
		return
	}
	failure := ""
	if runErr != nil {
		failure = runErr.Error()
	}
	data, err := json.Marshal(struct {
		Test          string
		Output        []byte
		Stderr, Error string
	}{t.Name(), output, stderr, failure})
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.CreateTemp(directory, "output-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
