package regex

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func run(t *testing.T, cwd, command string, args ...string) []byte {
	t.Helper()
	c := exec.Command(command, args...)
	c.Dir = cwd
	if !strings.ContainsRune(command, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment, not PATH's, run
		// in cwd with PWD to match. A program a test built (an oracle, the native runner) runs as named, with this
		// process's environment.
		c = toolchain.CommandIn(cwd, command, args...)
	}
	var out, err bytes.Buffer
	c.Stdout = &out
	c.Stderr = &err
	if e := c.Run(); e != nil || err.Len() != 0 {
		t.Fatalf("%s: %v stderr=%s", command, e, err.String())
	}
	return out.Bytes()
}

// backends are the runner's answers on source Node, emitted JS and sanitized native, the last two built ahead by
// runnerProduct from the same source tree.
func backends(t *testing.T, root, product string) [][]byte {
	t.Helper()
	entry := filepath.Join(root, "testdata/runner.a")
	repository, _ := filepath.Abs("../../../..")
	return [][]byte{run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry), run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(product, "runner.js")), run(t, root, filepath.Join(product, "runner"))}
}

// translationMutant is the one translation character TestFixedPatterns changes in patterns.a.
var translationMutant = [2]string{"/PaginationInput(", "/PaginationInputBROKEN("}

// writeTranslationMutant writes the runner's source tree into directory with translationMutant applied.
func writeTranslationMutant(directory string) error {
	if e := os.MkdirAll(filepath.Join(directory, "testdata"), 0755); e != nil {
		return e
	}
	for _, name := range []string{"patterns.a", "testdata/runner.a", "testdata/corpus.a"} {
		content, e := os.ReadFile(name)
		if e != nil {
			return e
		}
		if name == "patterns.a" {
			content = bytes.Replace(content, []byte(translationMutant[0]), []byte(translationMutant[1]), 1)
		}
		if e := os.WriteFile(filepath.Join(directory, name), content, 0644); e != nil {
			return e
		}
	}
	return nil
}

// runnerProduct is testdata/runner.a lowered once to emitted JS (runner.js) and a sanitized native binary (runner),
// built ahead so the unit only runs them. With mutant, it is built from writeTranslationMutant's tree, which the
// product also holds (mutant/) for source Node. Its key is the runner's three sources, every file go list says the
// lowering and building packages compile (GoInputs: their sources, embedded runtime and modules, Go and C), the
// mutation, clang and the native flags. Each is looked up once per process.
func runnerProduct(t *testing.T, mutant bool) string {
	t.Helper()
	product := &runnerProducts[0]
	if mutant {
		product = &runnerProducts[1]
	}
	product.once.Do(func() { product.directory = buildRunnerProduct(t, mutant) })
	if product.directory == "" {
		t.Fatal("runner product lookup did not complete")
	}
	return product.directory
}

// runnerProducts are the real runner's product and the mutant's, each looked up once.
var runnerProducts [2]struct {
	once      sync.Once
	directory string
}

func buildRunnerProduct(t *testing.T, mutant bool) string {
	t.Helper()
	compiler, e := buildcache.GoInputs("regex-runner", "./internal/load ./internal/lower ./internal/javascript ./internal/native", nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	inputs := buildcache.Inputs{Name: "regex-runner", Files: slices.Concat([]string{"stage1/cohere/lint/regex/patterns.a", "stage1/cohere/lint/regex/testdata/runner.a", "stage1/cohere/lint/regex/testdata/corpus.a"}, compiler.Files), Flags: slices.Concat(compiler.Flags, native.Flags(native.Options{Sanitize: true}), []string{"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}), Toolchain: slices.Concat(compiler.Toolchain, []string{buildcache.Tool("clang", "--version")})}
	if mutant {
		inputs.Name = "regex-runner-mutant"
		inputs.Flags = append(inputs.Flags, "mutant "+strconv.Quote(translationMutant[0])+" "+strconv.Quote(translationMutant[1]))
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		root, e := filepath.Abs(".")
		if e != nil {
			return e
		}
		if mutant {
			root = filepath.Join(directory, "mutant")
			if e := writeTranslationMutant(root); e != nil {
				return e
			}
		}
		p, e := load.Load([]string{filepath.Join(root, "testdata/runner.a")})
		if e != nil {
			return e
		}
		ir, e := lower.Lower(context.Background(), p)
		if e != nil {
			return e
		}
		if e := os.WriteFile(filepath.Join(directory, "runner.js"), []byte(javascript.JavaScript(ir)), 0644); e != nil {
			return e
		}
		return native.Build(native.C(ir), filepath.Join(directory, "runner"), native.Options{Sanitize: true})
	})
}

// cohereProgram is one of testdata's Go programs built in cohere's module, where it ran with go run: a product built
// ahead, so the unit only runs it (#5qykzj5).
func cohereProgram(t *testing.T, file, output string) string {
	t.Helper()
	path, err := filepath.Abs(file)
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.GoBuildIn(t, "cohere", output, path, nil)
}

func firstRow(a, b []byte) int {
	left := strings.Split(string(a), "\n")
	right := strings.Split(string(b), "\n")
	for i, s := range left {
		if i >= len(right) || s != right[i] {
			n, _ := strconv.Atoi(strings.Split(s, ":")[0])
			return n
		}
	}
	return 0
}

func TestFixedPatterns(t *testing.T) {
	t.Parallel()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	expected := run(t, cohere, cohereProgram(t, "testdata/oracle.go", "regex-oracle"), filepath.Join(root, "table.json"), filepath.Join(root, "testdata/corpus.json"))
	var rows []struct {
		ID      string  `json:"id"`
		Pattern *string `json:"go_pattern"`
	}
	data, _ := os.ReadFile("table.json")
	if e := json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	ids := []string{}
	for _, r := range rows {
		if r.Pattern != nil {
			ids = append(ids, r.ID)
		}
	}
	for backend, actual := range backends(t, root, runnerProduct(t, false)) {
		if !bytes.Equal(expected, actual) {
			t.Fatalf("backend %d differs: %s", backend, ids[firstRow(expected, actual)])
		}
	}
	t.Logf("%d fixed patterns, %d match records, %d identical bytes on Go, Node, emitted JS and sanitized native", len(ids), bytes.Count(expected, []byte("\n")), len(expected))
	mutant := runnerProduct(t, true)
	for backend, actual := range backends(t, filepath.Join(mutant, "mutant"), mutant) {
		if bytes.Equal(expected, actual) {
			t.Fatalf("translation mutant escaped backend %d", backend)
		}
		t.Logf("translation mutant caught by byte comparison on backend %d: %s", backend, ids[firstRow(expected, actual)])
	}
}
func TestDynamicPatternGap(t *testing.T) {
	t.Parallel()
	entry, _ := filepath.Abs("testdata/dynamic_gap.a")
	repository, _ := filepath.Abs("../../../..")
	out := run(t, ".", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry, "TODO")
	if string(out) != "true\n" {
		t.Fatalf("Node %q", out)
	}
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	_, e = lower.Lower(context.Background(), p)
	if e == nil || !strings.Contains(e.Error(), "RegExp with a nonconstant pattern") {
		t.Fatalf("expected named native gap, got %v", e)
	}
	t.Log(e)
}

func TestOptionDialectGap(t *testing.T) {
	t.Parallel()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	goAnswer := run(t, cohere, cohereProgram(t, "testdata/option_dialects.go", "regex-option-dialects"))
	nodeAnswer := run(t, root, "node", filepath.Join(root, "testdata/option_dialects.mjs"))
	expectedGo := "\\s\tfalse\na$\tfalse\n(?i)todo\ttrue\n\\p{Greek}\ttrue\na\\z\ttrue\n(?P<word>a)\ttrue\n"
	expectedNode := "\\s\ttrue\na$\tfalse\n(?i)todo\tSyntaxError\n\\p{Greek}\tSyntaxError\na\\z\tSyntaxError\n(?P<word>a)\tSyntaxError\n"
	if string(goAnswer) != expectedGo || string(nodeAnswer) != expectedNode {
		t.Fatalf("dialect observations changed: Go %q Node %q", goAnswer, nodeAnswer)
	}
	t.Logf("raw new RegExp(pattern, u) differs from accepted Go options on five of six named controls; Go=%q Node=%q", goAnswer, nodeAnswer)
}

func TestInventoryMatchesPinnedSource(t *testing.T) {
	t.Parallel()
	cohere, _ := filepath.Abs("../../../../cohere")
	actual := run(t, cohere, cohereProgram(t, "testdata/inventory.go", "regex-inventory"), filepath.Join(cohere, "internal/lint/rules"))
	expected, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, actual) {
		t.Fatal("regexp census drifted from pinned cohere AST; regenerate and review the table")
	}
}
