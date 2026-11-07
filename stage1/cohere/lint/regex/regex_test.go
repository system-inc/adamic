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
	"github.com/system-inc/adamic/stage1/cohere/lint/regex/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func run(test *testing.T, cwd, command string, args ...string) []byte {
	test.Helper()
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
	if runError := c.Run(); runError != nil || err.Len() != 0 {
		test.Fatalf("%s: %v stderr=%s", command, runError, err.String())
	}
	return out.Bytes()
}

// backends are the runner's answers on source Node, emitted JS and sanitized native, the last two built ahead by
// runnerProduct from the same source tree.
func backends(test *testing.T, root, product string) [][]byte {
	test.Helper()
	entry := filepath.Join(root, "testdata/runner.a")
	repository, _ := filepath.Abs("../../../..")
	return [][]byte{run(
		test,
		root,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		entry,
	), run(
		test,
		root,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		filepath.Join(product, "runner.js"),
	), run(test, root, filepath.Join(product, "runner"))}
}

// translationMutant is the one translation character TestFixedPatterns changes in patterns.a.
var translationMutant = [2]string{"/PaginationInput(", "/PaginationInputBROKEN("}

// writeTranslationMutant writes the runner's source tree into directory with translationMutant applied.
func writeTranslationMutant(directory string) error {
	if createDirectoryError := os.MkdirAll(filepath.Join(directory, "testdata"), 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	for _, name := range []string{"patterns.a", "testdata/runner.a", "testdata/corpus.a"} {
		content, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		if name == "patterns.a" {
			content = bytes.Replace(content, []byte(translationMutant[0]), []byte(translationMutant[1]), 1)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), content, 0644); writeError != nil {
			return writeError
		}
	}
	return nil
}

// runnerProduct is testdata/runner.a lowered once to emitted JS (runner.js) and a sanitized native binary (runner),
// built ahead so the unit only runs them. With mutant, it is built from writeTranslationMutant's tree, which the
// product also holds (mutant/) for source Node. Its key is the runner's three sources, every file go list says the
// lowering and building packages compile (GoInputs: their sources, embedded runtime and modules, Go and C), the
// mutation, clang and the native flags. Each is looked up once per process.
func runnerProduct(test *testing.T, mutant bool) string {
	test.Helper()
	product := &runnerProducts[0]
	if mutant {
		product = &runnerProducts[1]
	}
	product.once.Do(func() { product.directory = buildRunnerProduct(test, mutant) })
	if product.directory == "" {
		test.Fatal("runner product lookup did not complete")
	}
	return product.directory
}

// runnerProducts are the real runner's product and the mutant's, each looked up once.
var runnerProducts [2]struct {
	once      sync.Once
	directory string
}

func buildRunnerProduct(test *testing.T, mutant bool) string {
	test.Helper()
	compiler, goInputsError := buildcache.GoInputs(
		"regex-runner",
		"./internal/load ./internal/lower ./internal/javascript ./internal/native",
		nil,
		nil,
	)
	if goInputsError != nil {
		test.Fatal(goInputsError)
	}
	inputs := buildcache.Inputs{
		Name: "regex-runner",
		Files: slices.Concat([]string{
			"stage1/cohere/lint/regex/patterns.a",
			"stage1/cohere/lint/regex/testdata/runner.a",
			"stage1/cohere/lint/regex/testdata/corpus.a",
		}, compiler.Files),
		Flags: slices.Concat(
			compiler.Flags,
			native.Flags(native.Options{Sanitize: true}),
			[]string{"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")},
		),
		Toolchain: slices.Concat(compiler.Toolchain, []string{buildcache.Tool("clang", "--version")}),
	}
	if mutant {
		inputs.Name = "regex-runner-mutant"
		inputs.Flags = append(
			inputs.Flags,
			"mutant "+strconv.Quote(translationMutant[0])+" "+strconv.Quote(translationMutant[1]),
		)
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		root, absolutePathError := filepath.Abs(".")
		if absolutePathError != nil {
			return absolutePathError
		}
		if mutant {
			root = filepath.Join(directory, "mutant")
			if writeTranslationMutantError := writeTranslationMutant(root); writeTranslationMutantError != nil {
				return writeTranslationMutantError
			}
		}
		program, loadError := load.Load([]string{filepath.Join(root, "testdata/runner.a")})
		if loadError != nil {
			return loadError
		}
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "runner.js"),
			[]byte(javascript.JavaScript(ir)),
			0644,
		); writeError != nil {
			return writeError
		}
		return native.Build(native.C(ir), filepath.Join(directory, "runner"), native.Options{Sanitize: true})
	})
}

// cohereProgram is one of testdata's Go programs built in cohere's module, where it ran with go run: a product built
// ahead, so the unit only runs it (#5qykzj5).
func cohereProgram(test *testing.T, file, output string) string {
	test.Helper()
	return oracles.Program(test, file, output)
}

func firstRow(a, b []byte) int {
	left := strings.Split(string(a), "\n")
	right := strings.Split(string(b), "\n")
	for index, text := range left {
		if index >= len(right) || text != right[index] {
			number, _ := strconv.Atoi(strings.Split(text, ":")[0])
			return number
		}
	}
	return 0
}

func TestFixedPatterns(test *testing.T) {
	test.Parallel()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	expected := run(
		test,
		cohere,
		cohereProgram(test, "testdata/oracle.go", "regex-oracle"),
		filepath.Join(root, "table.json"),
		filepath.Join(root, "testdata/corpus.json"),
	)
	var rows []struct {
		ID      string  `json:"id"`
		Pattern *string `json:"go_pattern"`
	}
	data, _ := os.ReadFile("table.json")
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	ids := []string{}
	for _, row := range rows {
		if row.Pattern != nil {
			ids = append(ids, row.ID)
		}
	}
	for backend, actual := range backends(test, root, runnerProduct(test, false)) {
		if !bytes.Equal(expected, actual) {
			test.Fatalf("backend %d differs: %s", backend, ids[firstRow(expected, actual)])
		}
	}
	test.Logf(
		"%d fixed patterns, %d match records, %d identical bytes on Go, Node, emitted JS and sanitized native",
		len(ids),
		bytes.Count(expected, []byte("\n")),
		len(expected),
	)
	mutant := runnerProduct(test, true)
	for backend, actual := range backends(test, filepath.Join(mutant, "mutant"), mutant) {
		if bytes.Equal(expected, actual) {
			test.Fatalf("translation mutant escaped backend %d", backend)
		}
		test.Logf(
			"translation mutant caught by byte comparison on backend %d: %s",
			backend,
			ids[firstRow(expected, actual)],
		)
	}
}
func TestDynamicPatternGap(test *testing.T) {
	test.Parallel()
	entry, _ := filepath.Abs("testdata/dynamic_gap.a")
	repository, _ := filepath.Abs("../../../..")
	out := run(
		test,
		".",
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		entry,
		"TODO",
	)
	if string(out) != "true\n" {
		test.Fatalf("Node %q", out)
	}
	program, loadError := load.Load([]string{entry})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	if lowerError == nil || !strings.Contains(lowerError.Error(), "RegExp with a nonconstant pattern") {
		test.Fatalf("expected named native gap, got %v", lowerError)
	}
	test.Log(lowerError)
}

func TestOptionDialectGap(test *testing.T) {
	test.Parallel()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	goAnswer := run(test, cohere, cohereProgram(test, "testdata/option_dialects.go", "regex-option-dialects"))
	nodeAnswer := run(test, root, "node", filepath.Join(root, "testdata/option_dialects.mjs"))
	expectedGo := "\\s\tfalse\na$\tfalse\n(?i)todo\ttrue\n\\p{Greek}\ttrue\na\\z\ttrue\n(?P<word>a)\ttrue\n"
	expectedNode := "\\s\ttrue\na$\tfalse\n(?i)todo\tSyntaxError\n\\p{Greek}\tSyntaxError\na\\z\tSyntaxError\n(?P<word>a)\tSyntaxError\n"
	if string(goAnswer) != expectedGo || string(nodeAnswer) != expectedNode {
		test.Fatalf("dialect observations changed: Go %q Node %q", goAnswer, nodeAnswer)
	}
	test.Logf(
		"raw new RegExp(pattern, u) differs from accepted Go options on five of six named controls; Go=%q Node=%q",
		goAnswer,
		nodeAnswer,
	)
}

func TestInventoryMatchesPinnedSource(test *testing.T) {
	test.Parallel()
	cohere, _ := filepath.Abs("../../../../cohere")
	actual := run(
		test,
		cohere,
		cohereProgram(test, "testdata/inventory.go", "regex-inventory"),
		filepath.Join(cohere, "internal/lint/rules"),
	)
	expected, readError := os.ReadFile("sites.json")
	if readError != nil {
		test.Fatal(readError)
	}
	if !bytes.Equal(expected, actual) {
		test.Fatal("regexp census drifted from pinned cohere AST; regenerate and review the table")
	}
}
