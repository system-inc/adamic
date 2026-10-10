package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
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
	"path":   {"path", "path_probe.ts", "adamic_path", map[string]string{"cmd/adamic_path/main.go": "path_go.go", "internal/format/markdown/adamic_path_facts.go": "path_facts.go", "internal/format/markdown/adamic_path.go": "path_bridge.go", "internal/format/markdown/adamic_ast.go": "ast_bridge.go"}},
	"mdast":  {"mdast", "mdast_probe.ts", "adamic_mdast", map[string]string{"cmd/adamic_mdast/main.go": "mdast_go.go", "internal/format/markdown/mdast/adamic_mdast.go": "mdast_bridge.go", "internal/format/markdown/micromark/adamic_events.go": "events_transport.go"}},
	"decode": {"decode", "decode_probe.ts", "adamic_decode", map[string]string{"cmd/adamic_decode/main.go": "decode_go.go"}},
	"width":  {"width", "width_probe.ts", "adamic_width", map[string]string{"cmd/adamic_width/main.go": "width_go.go"}},
}
var grainPreparations sync.Map
var grainContexts sync.Map

func grainReady(t *testing.T, name string) grainProducts {
	t.Helper()
	state, _ := grainPreparations.LoadOrStore(name, &grainPreparation{})
	prepared := state.(*grainPreparation)
	prepared.once.Do(func() {
		started := time.Now()
		prepared.products = grainProduct(t, name, "")
		t.Logf("grain setup %.3fs", time.Since(started).Seconds())
	})
	return prepared.products
}

// Every product declaration uses this recipe, exactly as independent shard preparation does.
func grainProduct(t *testing.T, name, target string) grainProducts {
	t.Helper()
	recipe := grainRecipes[name]
	root := markdownLayoutProductRoot(t)
	main := filepath.Join(root, "stage1/cohere/markdownblocks/testdata", recipe.probe)
	p := grainProducts{main: main}
	goTools := []string{buildcache.Tool("go", "version")}
	nativeTools := []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")}
	// GoBuild cannot key a newly synthesized overlay target yet: GoInputs stats that
	// nonexistent target. Rule 7 permits this Go build as a Product, with all inputs
	// and reproducible flags declared. Generator/formatter products use GoBuild below.
	if target == "" || target == "go" {
		flags := []string{"-trimpath", "-buildvcs=false", "-ldflags=-buildid="}
		inputs := buildcache.Inputs{Name: "markdown-grain-" + name + "-go-v2", Files: []string{"go.mod", "go.work", "cohere", "stage1/cohere/markdownblocks/testdata", "stage1/cohere/markdownblocks/grain30_products_test.go"}, Toolchain: goTools, Flags: append([]string(nil), flags...)}
		for _, variable := range []string{"GOFLAGS", "CGO_ENABLED", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS"} {
			inputs.Flags = append(inputs.Flags, variable+"="+os.Getenv(variable))
		}
		inputs.Flags = append(inputs.Flags, buildcache.Tool("go", "env", "GOWORK"))
		dir := buildcache.Product(t, inputs, func(dir string) error {
			replace := map[string]string{}
			for to, from := range recipe.overlay {
				replace[filepath.Join(root, "cohere", to)] = filepath.Join(root, "stage1/cohere/markdownblocks/testdata", from)
			}
			data, err := json.Marshal(map[string]any{"Replace": replace})
			if err != nil {
				return err
			}
			overlay := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlay, data, 0644); err != nil {
				return err
			}
			args := append(append([]string{"build"}, flags...), "-overlay="+overlay, "-o", filepath.Join(dir, "oracle"), filepath.Join(root, "cohere/cmd", recipe.command, "main.go"))
			cmd := exec.CommandContext(t.Context(), "go", args...)
			cmd.Dir = filepath.Join(root, "cohere")
			if output, err := combinedOutput(cmd); err != nil {
				return fmt.Errorf("Go oracle: %w\n%s", err, output)
			}
			return nil
		})
		p.goBinary = filepath.Join(dir, "oracle")
	}

	if target == "go" {
		return p
	}
	lowerDir := buildcache.Product(t, buildcache.Inputs{Name: "markdown-grain-" + name + "-lowered", Files: []string{"go.mod", "internal", "stage1/cohere/markdownblocks"}, Toolchain: goTools}, func(dir string) error {
		program, err := loweredResult(main)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	p.backend = filepath.Join(lowerDir, "program.mjs")
	if target == "lowered" {
		return p
	}
	source, err := os.ReadFile(filepath.Join(lowerDir, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sanitize := range []bool{true, false} {
		if target != "" && ((sanitize && target != "sanitized") || (!sanitize && target != "release")) {
			continue
		}
		options := native.Options{Sanitize: sanitize}
		flags := append(native.Flags(options), fmt.Sprintf("source=%x", sha256.Sum256(source)), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		dir := buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprintf("markdown-grain-%s-native-%t", name, sanitize), Files: []string{"internal/native"}, Flags: flags, Toolchain: nativeTools}, func(dir string) error { return native.Build(string(source), filepath.Join(dir, "port"), options) })
		if sanitize {
			p.sanitized = filepath.Join(dir, "port")
		} else {
			p.release = filepath.Join(dir, "port")
		}
	}
	return p
}

func grainOwn(t *testing.T) func() {
	t.Helper()
	started := time.Now()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	grainContexts.Store(t, ctx)
	return func() { cancel(); grainContexts.Delete(t); t.Logf("grain own %.3fs", time.Since(started).Seconds()) }
}
func grainCommand(t *testing.T, name string, args ...string) *exec.Cmd {
	ctx := t.Context()
	if value, ok := grainContexts.Load(t); ok {
		ctx = value.(context.Context)
	}
	return textCommand(t, ctx, name, args...)
}
func grainExecute(t *testing.T, env []string, name string, args ...string) run {
	t.Helper()
	cmd := grainCommand(t, name, args...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	var out, errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	defer textActiveCommands.Delete(cmd)
	err := cmd.Run()
	if err != nil && cmd.ProcessState == nil {
		t.Fatal(err)
	}
	return run{stdout: out.Bytes(), stderr: errout.Bytes(), exitCode: cmd.ProcessState.ExitCode()}
}
func grainNode(t *testing.T, main string, args ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return grainExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, main}, args...)...)
}
func grainLeaks(t *testing.T, p grainProducts, args ...string) string {
	var r run
	if runtime.GOOS == "darwin" {
		r = grainExecute(t, nil, "leaks", append([]string{"--atExit", "--", p.release}, args...)...)
	} else {
		r = grainExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, p.sanitized, args...)
	}
	if r.exitCode != 0 {
		return fmt.Sprintf("exit %d\n%s\n%s", r.exitCode, r.stdout, r.stderr)
	}
	return ""
}
func grainSelect(inputs []auditInput, shard int) []auditInput {
	selected := []auditInput{}
	for i, input := range inputs {
		if i%grainCorpusShards == shard {
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
func grainPlanted(t *testing.T) {
	t.Helper()
	want := []byte("expected\n")
	got := append([]byte(nil), want...)
	got[0] = '!'
	if grainDifference(got, want) == nil {
		t.Fatal("planted output failure survived")
	}
	if os.Getenv("ADAMIC_GRAIN_PLANT") == t.Name() {
		t.Fatalf("planted output disagreement caught: %v", grainDifference(got, want))
	}
}

// Empty groups expect zero catches; each nonempty child plants its own disagreement.
func grainCheckUnion(t *testing.T, inputs []auditInput) {
	t.Helper()
	seen := make([]int, len(inputs))
	for shard := 0; shard < grainCorpusShards; shard++ {
		selected := grainSelect(inputs, shard)
		caught := 0
		for local, input := range selected {
			index := shard + local*grainCorpusShards
			if index >= len(inputs) || input != inputs[index] {
				t.Fatal("shard membership differs")
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
			t.Fatalf("shard %d catches %d want %d", shard, caught, expected)
		}
	}
	for index, count := range seen {
		if count != 1 {
			t.Fatalf("case %d occurs %d times", index, count)
		}
	}
	t.Logf("union=%d each exactly once", len(inputs))
}
func grainUnion(t *testing.T, inputs []auditInput, children int, family, helper string) {
	grainVerifyChildren(t, family, helper, children)
	grainCheckUnion(t, inputs)
	grainCheckUnion(t, nil)
	grainCheckUnion(t, []auditInput{{Name: "singleton", Text: "one"}})
	for child := 0; child < children; child++ {
		grainPlanted(t)
	}
}

func TestProduct_GrainMarkdown_path_go(t *testing.T) { t.Parallel(); grainProduct(t, "path", "go") }

func TestProduct_GrainMarkdown_path_lowered(t *testing.T) {
	t.Parallel()
	grainProduct(t, "path", "lowered")
}

func TestProduct_GrainMarkdown_path_sanitized(t *testing.T) {
	t.Parallel()
	grainProduct(t, "path", "sanitized")
}

func TestProduct_GrainMarkdown_path_release(t *testing.T) {
	t.Parallel()
	grainProduct(t, "path", "release")
}

func TestProduct_GrainMarkdown_mdast_go(t *testing.T) { t.Parallel(); grainProduct(t, "mdast", "go") }

func TestProduct_GrainMarkdown_mdast_lowered(t *testing.T) {
	t.Parallel()
	grainProduct(t, "mdast", "lowered")
}

func TestProduct_GrainMarkdown_mdast_sanitized(t *testing.T) {
	t.Parallel()
	grainProduct(t, "mdast", "sanitized")
}

func TestProduct_GrainMarkdown_mdast_release(t *testing.T) {
	t.Parallel()
	grainProduct(t, "mdast", "release")
}

func TestProduct_GrainMarkdown_decode_go(t *testing.T) { t.Parallel(); grainProduct(t, "decode", "go") }

func TestProduct_GrainMarkdown_decode_lowered(t *testing.T) {
	t.Parallel()
	grainProduct(t, "decode", "lowered")
}

func TestProduct_GrainMarkdown_decode_sanitized(t *testing.T) {
	t.Parallel()
	grainProduct(t, "decode", "sanitized")
}

func TestProduct_GrainMarkdown_decode_release(t *testing.T) {
	t.Parallel()
	grainProduct(t, "decode", "release")
}

func TestProduct_GrainMarkdown_width_go(t *testing.T) { t.Parallel(); grainProduct(t, "width", "go") }

func TestProduct_GrainMarkdown_width_lowered(t *testing.T) {
	t.Parallel()
	grainProduct(t, "width", "lowered")
}

func TestProduct_GrainMarkdown_width_sanitized(t *testing.T) {
	t.Parallel()
	grainProduct(t, "width", "sanitized")
}

func TestProduct_GrainMarkdown_width_release(t *testing.T) {
	t.Parallel()
	grainProduct(t, "width", "release")
}

type grainToolPreparation struct {
	once sync.Once
	path string
}

var grainToolPreparations sync.Map

func grainTool(t *testing.T, name string) string {
	t.Helper()
	state, _ := grainToolPreparations.LoadOrStore(name, &grainToolPreparation{})
	prepared := state.(*grainToolPreparation)
	prepared.once.Do(func() { prepared.path = grainBuildTool(t, name) })
	return prepared.path
}

func grainBuildTool(t *testing.T, name string) string {
	pkg := "./stage1/cohere/markdownblocks/tools/generate_" + name
	if name == "formatter" {
		pkg = "github.com/system-inc/cohere/command/cohere"
	}
	return buildcache.GoBuild(t, "grain-"+name, pkg, nil)
}

func grainGenerator(t *testing.T, name string) string {
	if name == "decode" {
		name = "entities"
	}
	return grainTool(t, name)
}
func grainRegenerate(t *testing.T, name, generator, formatter string) {
	cmd := grainCommand(t, generator, "-check", "-formatter", formatter)
	cmd.Dir = markdownLayoutProductRoot(t)
	if output, err := combinedOutput(cmd); err != nil {
		t.Fatalf("%s regeneration: %v\n%s", name, err, output)
	}
}
func TestProduct_GrainMarkdownEntitiesGenerator(t *testing.T) { t.Parallel(); grainTool(t, "entities") }
func TestProduct_GrainMarkdownWidthGenerator(t *testing.T)    { t.Parallel(); grainTool(t, "width") }
func TestProduct_GrainMarkdownFormatter(t *testing.T)         { t.Parallel(); grainTool(t, "formatter") }

// Inspect dispatchers, rather than trusting a second hand-written list of leaves.
func grainVerifyChildren(t *testing.T, family, helper string, children int) {
	t.Helper()
	files, err := filepath.Glob("grain30_*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int]bool{}
	for _, file := range files {
		tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
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
			index, err := strconv.Atoi(suffix)
			if err != nil || len(suffix) != 3 || index < 0 || index >= children || seen[index] {
				t.Fatalf("invalid child %s", fn.Name.Name)
			}
			if len(fn.Body.List) != 2 {
				t.Fatalf("invalid child body %s", fn.Name.Name)
			}
			parallel, ok := fn.Body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Fatal("missing t.Parallel")
			}
			call, ok := parallel.X.(*ast.CallExpr)
			if !ok {
				t.Fatal("missing parallel call")
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Parallel" {
				t.Fatal("parallel must be first")
			}
			statement, ok := fn.Body.List[1].(*ast.ExprStmt)
			if !ok {
				t.Fatal("invalid dispatch")
			}
			dispatch, ok := statement.X.(*ast.CallExpr)
			if !ok || len(dispatch.Args) < 2 {
				t.Fatal("invalid dispatch call")
			}
			name, ok := dispatch.Fun.(*ast.Ident)
			if !ok || name.Name != helper {
				t.Fatal("wrong child helper")
			}
			literal, ok := dispatch.Args[len(dispatch.Args)-1].(*ast.BasicLit)
			if !ok || literal.Value != strconv.Itoa(index) {
				t.Fatal("child name and membership disagree")
			}
			seen[index] = true
		}
	}
	if len(seen) != children {
		t.Fatalf("enumerated %d children, want %d", len(seen), children)
	}
}
