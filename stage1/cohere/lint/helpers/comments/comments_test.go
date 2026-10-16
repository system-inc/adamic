package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/helpers/comments/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment, not PATH's, run
		// in dir with PWD to match. A program a test built (the Go oracle, the native helpers) runs as named, with this
		// process's environment.
		cmd = toolchain.CommandIn(dir, name, args...)
	}
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

// A nativeMutation is one planted change to the helper sources: old, which must occur in file exactly once, becomes
// new. The zero value is the unmutated program.
type nativeMutation struct{ file, old, new string }

// The sources of the helpers program, copied whole into a mutant's directory.
var helperSources = []string{"main.ts", "comment.ts", "can_begin_at.ts", "collect_list_interiors.ts", "sort_by_position.ts", "all.ts", "for_file.ts"}

// nativeProducts holds one nativeProduct per mutation, so each is got once per process.
var nativeProducts sync.Map

type nativeProduct struct {
	once   sync.Once
	binary string
}

// helpersNative is the sanitized native helpers program built with m planted, as a product the tree build makes ahead
// (TestProduct_CommentsNative and the mutants' declarations), so no comparison pays for clang. native.Build serializes
// its own runtime cache entries, so products build in parallel.
func helpersNative(t *testing.T, m nativeMutation) string {
	t.Helper()
	value, _ := nativeProducts.LoadOrStore(m, &nativeProduct{})
	product := value.(*nativeProduct)
	product.once.Do(func() {
		name := "comments-native"
		if m.file != "" {
			name += "-mutant-" + m.file
		}
		files := []string{"go.mod", "go.work", "internal", "stage1/typescript", "stage1/cohere/lint/helpers/options_json.ts", "stage1/cohere/lint/helpers/comments/comments_test.go"}
		for _, file := range helperSources {
			files = append(files, "stage1/cohere/lint/helpers/comments/"+file)
		}
		flags := []string{"native.Build", "Sanitize=true", "mutation=" + m.file + ":" + m.old + "=>" + m.new, "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}
		tools := []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
		directory := buildcache.Product(t, buildcache.Inputs{Name: name, Files: files, Flags: flags, Toolchain: tools}, func(directory string) error {
			source := "."
			if m.file != "" {
				source = t.TempDir()
				if err := writeMutant(source, m); err != nil {
					return err
				}
			}
			program, err := load.Load([]string{filepath.Join(source, "main.ts")})
			if err != nil {
				return err
			}
			ir, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			return native.Build(native.C(ir), filepath.Join(directory, "helpers"), native.Options{Sanitize: true})
		})
		product.binary = filepath.Join(directory, "helpers")
	})
	if product.binary == "" {
		t.Fatal("helpers native product failed earlier in this process")
	}
	return product.binary
}

// writeMutant copies the helper sources into directory with m planted, their imports from outside the package made
// absolute.
func writeMutant(directory string, m nativeMutation) error {
	parserRoot, err := filepath.Abs("../../../../typescript")
	if err != nil {
		return err
	}
	options, err := filepath.Abs("../options_json.ts")
	if err != nil {
		return err
	}
	for _, file := range helperSources {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		text := string(data)
		if file == m.file {
			if strings.Count(text, m.old) != 1 {
				return fmt.Errorf("mutant anchor count %d in %s", strings.Count(text, m.old), file)
			}
			text = strings.Replace(text, m.old, m.new, 1)
		}
		text = strings.ReplaceAll(text, "../../../../typescript", filepath.ToSlash(parserRoot))
		text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(options))
		if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
			return err
		}
	}
	return nil
}

func TestProduct_CommentsNative(t *testing.T) {
	t.Parallel()
	helpersNative(t, nativeMutation{})
}

// jsxGuardMutant removes the TSX adapter guard.
var jsxGuardMutant = nativeMutation{"main.ts", "if(adapted < 0 && name.endsWith('.tsx'))", "if(false)"}

func TestProduct_CommentsJsxGuardMutant(t *testing.T) {
	t.Parallel()
	helpersNative(t, jsxGuardMutant)
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
	return oracles.Comments(t)
}

func TestProduct_CommentsFindsGoOracle(t *testing.T) {
	t.Parallel()
	oracle(t)
}

func TestCommentsMatchCohere(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("testdata/witnesses.json")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	want := run(t, "", oracle(t), path)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", helpersNative(t, nativeMutation{}), path), want)
	t.Logf("Go, Node and sanitized native match %d output lines", bytes.Count(want, []byte("\n")))
}

func TestJsxParserGapIsExplicit(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/parser-gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	binary := helpersNative(t, nativeMutation{})
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
	t.Parallel()
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
	output, stderr, err := gapRun(t, []string{helpersNative(t, jsxGuardMutant), path})
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
	if !strings.ContainsRune(command[0], filepath.Separator) { // node, through internal/toolchain; a built binary runs as named
		cmd = toolchain.Command(command[0], command[1:]...)
	}
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
