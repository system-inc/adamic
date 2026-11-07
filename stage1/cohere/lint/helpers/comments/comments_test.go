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

func run(test *testing.T, dir, name string, args ...string) []byte {
	test.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment, not PATH's, run
		// in dir with PWD to match. A program a test built (the Go oracle, the native helpers) runs as named, with this
		// process's environment.
		cmd = toolchain.CommandIn(dir, name, args...)
	}
	f, err := os.CreateTemp(test.TempDir(), "output-")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = childguard.Run(cmd, childguard.Options{}); err != nil {
		test.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		test.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		test.Fatal(err)
	}
	return data
}

// A nativeMutation is one planted change to the helper sources: old, which must occur in file exactly once, becomes
// new. The zero value is the unmutated program.
type nativeMutation struct{ file, old, new string }

// The sources of the helpers program, copied whole into a mutant's directory.
var helperSources = []string{
	"main.ts",
	"comment.ts",
	"can_begin_at.ts",
	"collect_list_interiors.ts",
	"sort_by_position.ts",
	"all.ts",
	"for_file.ts",
}

// nativeProducts holds one nativeProduct per mutation, so each is got once per process.
var nativeProducts sync.Map

type nativeProduct struct {
	once   sync.Once
	binary string
}

// helpersNative is the sanitized native helpers program built with m planted, as a product the tree build makes ahead
// (TestProduct_CommentsNative and the mutants' declarations), so no comparison pays for clang. native.Build serializes
// its own runtime cache entries, so products build in parallel.
func helpersNative(test *testing.T, m nativeMutation) string {
	test.Helper()
	value, _ := nativeProducts.LoadOrStore(m, &nativeProduct{})
	product := value.(*nativeProduct)
	product.once.Do(func() {
		name := "comments-native"
		if m.file != "" {
			name += "-mutant-" + m.file
		}
		files := []string{
			"go.mod",
			"go.work",
			"internal",
			"stage1/typescript",
			"stage1/cohere/lint/helpers/options_json.ts",
			"stage1/cohere/lint/helpers/comments/comments_test.go",
		}
		for _, file := range helperSources {
			files = append(files, "stage1/cohere/lint/helpers/comments/"+file)
		}
		flags := []string{
			"native.Build",
			"Sanitize=true",
			"mutation=" + m.file + ":" + m.old + "=>" + m.new,
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		}
		tools := []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
		directory := buildcache.Product(test, buildcache.Inputs{
			Name:      name,
			Files:     files,
			Flags:     flags,
			Toolchain: tools,
		}, func(directory string) error {
			source := "."
			if m.file != "" {
				source = test.TempDir()
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
		test.Fatal("helpers native product failed earlier in this process")
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

func TestProduct_CommentsNative(test *testing.T) {
	test.Parallel()
	helpersNative(test, nativeMutation{})
}

// jsxGuardMutant removes the TSX adapter guard.
var jsxGuardMutant = nativeMutation{"main.ts", "if(adapted < 0 && name.endsWith('.tsx'))", "if(false)"}

func TestProduct_CommentsJsxGuardMutant(test *testing.T) {
	test.Parallel()
	helpersNative(test, jsxGuardMutant)
}

func compare(test *testing.T, got, want []byte) {
	test.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			test.Fatalf("output line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	test.Fatalf("output size: got %d Go %d", len(got), len(want))
}

func oracle(test *testing.T) string {
	test.Helper()
	return oracles.Comments(test)
}

func TestProduct_CommentsFindsGoOracle(test *testing.T) {
	test.Parallel()
	oracle(test)
}

func TestCommentsMatchCohere(test *testing.T) {
	test.Parallel()
	path, _ := filepath.Abs("testdata/witnesses.json")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	want := run(test, "", oracle(test), path)
	compare(test, run(test, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(test, run(test, "", helpersNative(test, nativeMutation{}), path), want)
	test.Logf("Go, Node and sanitized native match %d output lines", bytes.Count(want, []byte("\n")))
}

func TestJsxParserGapIsExplicit(test *testing.T) {
	test.Parallel()
	data, err := os.ReadFile("testdata/parser-gaps.json")
	if err != nil {
		test.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		test.Fatal(err)
	}
	binary := helpersNative(test, nativeMutation{})
	goOracle := oracle(test)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	for _, row := range rows {
		corpus, _ := json.Marshal([]any{row})
		path := filepath.Join(test.TempDir(), "gap.json")
		if err := os.WriteFile(path, corpus, 0644); err != nil {
			test.Fatal(err)
		}
		adapted := run(test, "", goOracle, "--ast", path)
		var facts []struct{ GoDiagnostics int }
		if err := json.Unmarshal(adapted, &facts); err != nil {
			test.Fatal(err)
		}
		if len(facts) != 1 || facts[0].GoDiagnostics != 0 {
			test.Fatal("gap must be valid Go TypeScript")
		}
		var previous string
		for _, command := range [][]string{
			{"node", "--disable-warning=ExperimentalWarning", runner, entry, path},
			{binary, path},
		} {
			_, stderr, err := gapRun(test, command)
			if !matchesGapRefusal(err, stderr) {
				test.Fatalf("gap silently accepted or failed differently: %v %s", err, stderr)
			}
			if previous != "" && previous != stderr {
				test.Fatal("Node and native parser refusals differ")
			}
			previous = stderr
		}
	}
	test.Logf(
		"%d valid Go JSX inputs explicitly refuse in the independent-parser adapter; AST-adapter helper comparisons cover them separately",
		len(rows),
	)
}

// Removing the TSX adapter guard must compile and accept the known wrong parse,
// rather than merely crashing on one of the other unsupported JSX forms.
func TestJsxAdapterGuardMutant(test *testing.T) {
	test.Parallel()
	data, err := os.ReadFile("testdata/parser-gaps.json")
	if err != nil {
		test.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		test.Fatal(err)
	}
	corpus, _ := json.Marshal(rows[:1])
	path := filepath.Join(test.TempDir(), "gap.json")
	if err := os.WriteFile(path, corpus, 0644); err != nil {
		test.Fatal(err)
	}
	output, stderr, err := gapRun(test, []string{helpersNative(test, jsxGuardMutant), path})
	if err != nil || stderr != "" {
		test.Fatalf("mutant must finish normally, not crash: %v %s", err, stderr)
	}
	if matchesGapRefusal(err, stderr) {
		test.Fatal("compiled adapter-guard mutant survived the actual refusal check")
	}
	if !bytes.Contains(output, []byte("comments ")) {
		test.Fatal("guard mutant did not finish the wrong parsed tree")
	}
	test.Log(
		"compiled adapter-guard mutant finishes the unsupported JSX input with exit 0; the explicit-refusal check requires exit 70 and catches it",
	)
}

func matchesGapRefusal(err error, stderr string) bool {
	exit, ok := err.(*exec.ExitError)
	return ok && exit.ExitCode() == 70 && strings.Contains(stderr, "NotYet: stage-1 JSX parser adapter")
}
func gapRun(test *testing.T, command []string) ([]byte, string, error) {
	test.Helper()
	cmd := exec.Command(command[0], command[1:]...)
	if !strings.ContainsRune(
		command[0],
		filepath.Separator,
	) { // node, through internal/toolchain; a built binary runs as named
		cmd = toolchain.Command(command[0], command[1:]...)
	}
	out, err := os.CreateTemp(test.TempDir(), "gap-output-")
	if err != nil {
		test.Fatal(err)
	}
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	runErr := testguard.Run(cmd, time.Minute, testguard.Ceiling)
	if err := out.Close(); err != nil {
		test.Fatal(err)
	}
	output, err := os.ReadFile(out.Name())
	if err != nil {
		test.Fatal(err)
	}
	return output, stderr.String(), runErr
}
