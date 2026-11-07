package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

const grainCorpusShards = 8

type grainRecipe struct {
	name, probe, command string
	overlay              map[string]string
}
type grainProducts struct{ main, goBinary, backend, sanitized, release string }
type grainPreparation struct {
	once     sync.Once
	products grainProducts
}

var grainRecipes = map[string]grainRecipe{
	"path": {"path", "path_probe.ts", "adamic_path", map[string]string{
		"cmd/adamic_path/main.go":                       "path_go.go",
		"internal/format/markdown/adamic_path_facts.go": "path_facts.go",
		"internal/format/markdown/adamic_path.go":       "path_bridge.go",
		"internal/format/markdown/adamic_ast.go":        "ast_bridge.go",
	}},
	"mdast": {"mdast", "mdast_probe.ts", "adamic_mdast", map[string]string{
		"cmd/adamic_mdast/main.go":                            "mdast_go.go",
		"internal/format/markdown/mdast/adamic_mdast.go":      "mdast_bridge.go",
		"internal/format/markdown/micromark/adamic_events.go": "events_transport.go",
	}},
	"decode": {
		"decode",
		"decode_probe.ts",
		"adamic_decode",
		map[string]string{"cmd/adamic_decode/main.go": "decode_go.go"},
	},
	"width": {"width", "width_probe.ts", "adamic_width", map[string]string{"cmd/adamic_width/main.go": "width_go.go"}},
}
var grainPreparations sync.Map
var grainContexts sync.Map

func grainReady(test *testing.T, name string) grainProducts {
	test.Helper()
	state, _ := grainPreparations.LoadOrStore(name, &grainPreparation{})
	prepared := state.(*grainPreparation)
	prepared.once.Do(func() {
		started := time.Now()
		prepared.products = grainProduct(test, name, "")
		test.Logf("grain setup %.3fs", time.Since(started).Seconds())
	})
	return prepared.products
}

// Every product declaration uses this recipe, exactly as independent shard preparation does.
func grainProduct(test *testing.T, name, target string) grainProducts {
	test.Helper()
	recipe := grainRecipes[name]
	root := markdownLayoutProductRoot(test)
	main := filepath.Join(root, "stage1/cohere/markdownblocks/testdata", recipe.probe)
	products := grainProducts{main: main}
	goTools := []string{buildcache.Tool("go", "version")}
	nativeTools := []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")}
	// GoBuild cannot key a newly synthesized overlay target yet: GoInputs stats that
	// nonexistent target. Rule 7 permits this Go build as a Product, with all inputs
	// and reproducible flags declared. Generator/formatter products use GoBuild below.
	if target == "" || target == "go" {
		products.goBinary = oracles.GrainGoOracle(test, name)
	}

	if target == "go" {
		return products
	}
	// Keyed by what lowering reads, the package's and markdowninline's TypeScript, the checker and the compiler, as
	// the pilot families' lowering is; the whole package directory would rebuild every grain port on any test edit.
	lowerDirectory := buildcache.Product(test, buildcache.Inputs{
		Name:      "markdown-grain-" + name + "-lowered",
		Files:     whitespaceLayoutBuildFiles(),
		Toolchain: goTools,
	}, func(dir string) error {
		program, loweredResultError := loweredResult(main)
		if loweredResultError != nil {
			return loweredResultError
		}
		if writeError := os.WriteFile(
			filepath.Join(dir, "program.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	products.backend = filepath.Join(lowerDirectory, "program.mjs")
	if target == "lowered" {
		return products
	}
	source, readError := os.ReadFile(filepath.Join(lowerDirectory, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	for _, sanitize := range []bool{true, false} {
		if target != "" && ((sanitize && target != "sanitized") || (!sanitize && target != "release")) {
			continue
		}
		options := native.Options{Sanitize: sanitize}
		flags := append(
			native.Flags(options),
			fmt.Sprintf("source=%x", sha256.Sum256(source)),
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		)
		dir := buildcache.Product(test, buildcache.Inputs{
			Name:      fmt.Sprintf("markdown-grain-%s-native-%t", name, sanitize),
			Files:     []string{"internal/native"},
			Flags:     flags,
			Toolchain: nativeTools,
		}, func(dir string) error { return native.Build(string(source), filepath.Join(dir, "port"), options) })
		if sanitize {
			products.sanitized = filepath.Join(dir, "port")
		} else {
			products.release = filepath.Join(dir, "port")
		}
	}
	return products
}

// grainGoInputs keys name's Go oracle. The build runs go under toolchain.Env (textGoCommand), so the key names the C
// and C++ compilers that environment gives go, and the C compiler by its report: the caller's CC and CXX are empty on a
// box where go runs clang. The other variables pass to go as the caller set them.
func grainGoInputs(name string, compilers markdownCompilers) buildcache.Inputs {
	return oracles.GrainGoProductInputs(name, compilers.c, compilers.cPlusPlus, compilers.report)
}

func grainOwn(test *testing.T) func() {
	test.Helper()
	started := time.Now()
	timeoutContext, cancel := context.WithTimeout(test.Context(), 90*time.Second)
	grainContexts.Store(test, timeoutContext)
	return func() {
		cancel()
		grainContexts.Delete(test)
		test.Logf("grain own %.3fs", time.Since(started).Seconds())
	}
}

// grainCommand returns the command's context with it, for childguard.Capture to end the reads by.
func grainCommand(test *testing.T, name string, args ...string) (*exec.Cmd, context.Context) {
	ctx := test.Context()
	if value, ok := grainContexts.Load(test); ok {
		ctx = value.(context.Context)
	}
	return textCommand(test, ctx, name, args...), ctx
}

// grainExecute reads a run's output through childguard.Capture for textExecute's reason. Only an exit error is an
// answer: an error that came with a ProcessState was taken for one, so an exec.ErrWaitDelay passed as exit 0 with its
// output cut short.
func grainExecute(test *testing.T, env []string, name string, args ...string) run {
	test.Helper()
	command, ctx := grainCommand(test, name, args...)
	if env != nil {
		command.Env = append(os.Environ(), env...)
	}
	defer textActiveCommands.Delete(command)
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatal(captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}
func grainNode(test *testing.T, main string, args ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return grainExecute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, main}, args...)...,
	)
}
func grainLeaks(test *testing.T, products grainProducts, args ...string) string {
	var run run
	if runtime.GOOS == "darwin" {
		run = grainExecute(test, nil, "leaks", append([]string{"--atExit", "--", products.release}, args...)...)
	} else {
		run = grainExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, args...)
	}
	if run.exitCode != 0 {
		return fmt.Sprintf("exit %d\n%s\n%s", run.exitCode, run.stdout, run.stderr)
	}
	return ""
}
func grainSelect(inputs []auditInput, shard int) []auditInput {
	selected := []auditInput{}
	for index, input := range inputs {
		if index%grainCorpusShards == shard {
			selected = append(selected, input)
		}
	}
	return selected
}
func grainDifference(got, want []byte) error {
	if bytes.Equal(got, want) {
		return nil
	}
	return fmt.Errorf("output differs: %d/%d bytes", len(got), len(want))
}
func grainPlanted(test *testing.T) {
	test.Helper()
	want := []byte("expected\n")
	got := append([]byte(nil), want...)
	got[0] = '!'
	if grainDifference(got, want) == nil {
		test.Fatal("planted output failure survived")
	}
	if os.Getenv("ADAMIC_GRAIN_PLANT") == test.Name() {
		test.Fatalf("planted output disagreement caught: %v", grainDifference(got, want))
	}
}

// Empty groups expect zero catches; each nonempty child plants its own disagreement.
func grainCheckUnion(test *testing.T, inputs []auditInput) {
	test.Helper()
	seen := make([]int, len(inputs))
	for shard := 0; shard < grainCorpusShards; shard++ {
		selected := grainSelect(inputs, shard)
		caught := 0
		for local, input := range selected {
			index := shard + local*grainCorpusShards
			if index >= len(inputs) || input != inputs[index] {
				test.Fatal("shard membership differs")
			}
			seen[index]++
			want := []byte(input.Text)
			got := append([]byte(nil), want...)
			if local == 0 {
				got = append(got, '!')
			}
			if grainDifference(got, want) != nil {
				caught++
			}
		}
		expected := 0
		if len(selected) > 0 {
			expected = 1
		}
		if caught != expected {
			test.Fatalf("shard %d catches %d want %d", shard, caught, expected)
		}
	}
	for index, count := range seen {
		if count != 1 {
			test.Fatalf("case %d occurs %d times", index, count)
		}
	}
	test.Logf("union=%d each exactly once", len(inputs))
}
func grainUnion(test *testing.T, inputs []auditInput, children int, family, helper string) {
	grainVerifyChildren(test, family, helper, children)
	grainCheckUnion(test, inputs)
	grainCheckUnion(test, nil)
	grainCheckUnion(test, []auditInput{{Name: "singleton", Text: "one"}})
	for child := 0; child < children; child++ {
		grainPlanted(test)
	}
}

func TestProduct_GrainMarkdown_path_lowered(test *testing.T) {
	test.Parallel()
	grainProduct(test, "path", "lowered")
}

func TestProduct_GrainMarkdown_path_sanitized(test *testing.T) {
	test.Parallel()
	grainProduct(test, "path", "sanitized")
}

func TestProduct_GrainMarkdown_path_release(test *testing.T) {
	test.Parallel()
	grainProduct(test, "path", "release")
}

func TestProduct_GrainMarkdown_mdast_lowered(test *testing.T) {
	test.Parallel()
	grainProduct(test, "mdast", "lowered")
}

func TestProduct_GrainMarkdown_mdast_sanitized(test *testing.T) {
	test.Parallel()
	grainProduct(test, "mdast", "sanitized")
}

func TestProduct_GrainMarkdown_mdast_release(test *testing.T) {
	test.Parallel()
	grainProduct(test, "mdast", "release")
}

func TestProduct_GrainMarkdown_decode_lowered(test *testing.T) {
	test.Parallel()
	grainProduct(test, "decode", "lowered")
}

func TestProduct_GrainMarkdown_decode_sanitized(test *testing.T) {
	test.Parallel()
	grainProduct(test, "decode", "sanitized")
}

func TestProduct_GrainMarkdown_decode_release(test *testing.T) {
	test.Parallel()
	grainProduct(test, "decode", "release")
}

func TestProduct_GrainMarkdown_width_lowered(test *testing.T) {
	test.Parallel()
	grainProduct(test, "width", "lowered")
}

func TestProduct_GrainMarkdown_width_sanitized(test *testing.T) {
	test.Parallel()
	grainProduct(test, "width", "sanitized")
}

func TestProduct_GrainMarkdown_width_release(test *testing.T) {
	test.Parallel()
	grainProduct(test, "width", "release")
}

type grainToolPreparation struct {
	once sync.Once
	path string
}

var grainToolPreparations sync.Map

func grainTool(test *testing.T, name string) string {
	test.Helper()
	state, _ := grainToolPreparations.LoadOrStore(name, &grainToolPreparation{})
	prepared := state.(*grainToolPreparation)
	prepared.once.Do(func() { prepared.path = grainBuildTool(test, name) })
	return prepared.path
}

func grainBuildTool(test *testing.T, name string) string { return oracles.GrainTool(test, name) }

func grainGenerator(test *testing.T, name string) string {
	if name == "decode" {
		name = "entities"
	}
	return grainTool(test, name)
}
func grainRegenerate(test *testing.T, name, generator, formatter string) {
	command, ctx := grainCommand(test, generator, "-check", "-formatter", formatter)
	command.Dir = markdownLayoutProductRoot(test)
	if output, textCombinedOutputError := textCombinedOutput(ctx, command); textCombinedOutputError != nil {
		test.Fatalf("%s regeneration: %v\n%s", name, textCombinedOutputError, output)
	}
}

// Inspect dispatchers, rather than trusting a second hand-written list of leaves.
func grainVerifyChildren(test *testing.T, family, helper string, children int) {
	test.Helper()
	files, globError := filepath.Glob("grain30_*_test.go")
	if globError != nil {
		test.Fatal(globError)
	}
	seen := map[int]bool{}
	for _, file := range files {
		tree, parseFileError := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		for _, decl := range tree.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			suffix, ok := strings.CutPrefix(fn.Name.Name, family+"_")
			if !ok {
				continue
			}
			index, parseError := strconv.Atoi(suffix)
			if parseError != nil || len(suffix) != 3 || index < 0 || index >= children || seen[index] {
				test.Fatalf("invalid child %s", fn.Name.Name)
			}
			if len(fn.Body.List) != 2 {
				test.Fatalf("invalid child body %s", fn.Name.Name)
			}
			parallel, ok := fn.Body.List[0].(*ast.ExprStmt)
			if !ok {
				test.Fatal("missing t.Parallel")
			}
			call, ok := parallel.X.(*ast.CallExpr)
			if !ok {
				test.Fatal("missing parallel call")
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Parallel" {
				test.Fatal("parallel must be first")
			}
			statement, ok := fn.Body.List[1].(*ast.ExprStmt)
			if !ok {
				test.Fatal("invalid dispatch")
			}
			dispatch, ok := statement.X.(*ast.CallExpr)
			if !ok || len(dispatch.Args) < 2 {
				test.Fatal("invalid dispatch call")
			}
			name, ok := dispatch.Fun.(*ast.Ident)
			if !ok || name.Name != helper {
				test.Fatal("wrong child helper")
			}
			literal, ok := dispatch.Args[len(dispatch.Args)-1].(*ast.BasicLit)
			if !ok || literal.Value != strconv.Itoa(index) {
				test.Fatal("child name and membership disagree")
			}
			seen[index] = true
		}
	}
	if len(seen) != children {
		test.Fatalf("enumerated %d children, want %d", len(seen), children)
	}
}
