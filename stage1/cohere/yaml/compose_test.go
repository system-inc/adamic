package yaml

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/executable"
	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func goCompose(t *testing.T, cases string) []byte {
	t.Helper()
	binary := goComposeOracle(t)
	expected := run(t, "", nil, binary, cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		yamlArtifacts.Lock()
		defer yamlArtifacts.Unlock()
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		program, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		// A program copied while sibling tests fork goes through the fork-safe writer, or a fork can hold its writable
		// descriptor and a later exec of it fails with ETXTBSY (#6hjazy4).
		if err := executable.Write(filepath.Join(artifacts, "go-compose"), program); err != nil {
			t.Fatal(err)
		}
		inputs, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range []struct {
			name string
			data []byte
			mode os.FileMode
		}{{"compose-cases.txt", inputs, 0644}, {"compose-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}

// goComposeOracle is formatter_comparison over testdata/compose_go.go, with compose's numbers.go opened to it by
// testdata/compose_exports.txt, a product built ahead (#5qykzj5).
func goComposeOracle(t *testing.T) string { return oracles.Composition(t) }

func TestProduct_YamlFindsComposeGo(t *testing.T) {
	t.Parallel()
	goComposeOracle(t)
}
func composeCases(t *testing.T) (string, int, int) { return oracles.CompositionCases(t) }

// composeMatchGo compares compose's port build (port_products_test.go) with Go over composeCases, in two shards.
var composeMatchGo = yamlMatchGo{family: "compose", shards: 2, cases: composeCases, oracle: goComposeOracle, build: yamlComposeBuild, entry: "compose_main.ts"}

func TestComposeMatchGo_000(t *testing.T) { t.Parallel(); composeMatchGoShard(t, 0) }
func TestComposeMatchGo_001(t *testing.T) { t.Parallel(); composeMatchGoShard(t, 1) }

func TestComposeMatchGoUnion(t *testing.T) {
	t.Parallel()
	composeMatchGo.union(t)
}

func TestComposeMatchGoPlantedFailure(t *testing.T) {
	t.Parallel()
	composeMatchGo.planted(t)
}

func composeMatchGoShard(t *testing.T, shard int) {
	t.Helper()
	result := composeMatchGo.run(t, shard)
	expected := result.expected
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/compose_library.mjs", library, result.path)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", result.native}, {"Node", result.node}, {"emitted JavaScript", result.emitted}, {"yaml@2.9.0", external}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile(composeMatchGo.diagnostic(shard, "expected"), expected, 0644)
			os.WriteFile(composeMatchGo.diagnostic(shard, "actual"), side.out, 0644)
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	t.Logf("%d repository files, %d of %d cases, %d compose answer bytes identical on all five sides", result.files, len(result.indices), len(result.all), len(expected))
}

var composeMutants = yamlMutants{family: "compose", entry: "compose_main.ts", mutated: "composer.ts", mutations: []yamlMutation{
	{"document start marker lost", "this.document.directives.docStart = props.found >= 0;", "this.document.directives.docStart = false;"},
	{"block mapping pair lost", "this.node(result).items.push(this.pair(key, value, index, itemIndex));", "this.node(result).items.push(this.pair(key, -1, index, itemIndex));"},
	{"implicit key limit skipped", "properties.start < this.parser.get(valueProps.found).offset - 1024", "properties.start < this.parser.get(valueProps.found).offset - 999999"},
}}

func TestProduct_YAMLComposeMutant000(t *testing.T) { t.Parallel(); composeMutants.product(t, 0) }
func TestProduct_YAMLComposeMutant001(t *testing.T) { t.Parallel(); composeMutants.product(t, 1) }
func TestProduct_YAMLComposeMutant002(t *testing.T) { t.Parallel(); composeMutants.product(t, 2) }

// composeMutantsShards is the count of composeMutants.mutations, each its own top-level test so Loom's planner can cut
// between mutants, which it can't between subtests.
const composeMutantsShards = 3

// composeMutantsAnswers holds what every mutant compares against, made once per process: the composition cases and
// the Go composer's answer to them.
var composeMutantsAnswers struct {
	once            sync.Once
	cases, expected []byte
}

func composeMutantsExpected(t *testing.T) ([]byte, []byte) {
	t.Helper()
	composeMutantsAnswers.once.Do(func() {
		// The cases file lives in the first caller's temporary directory, gone when that test ends, so its bytes are kept.
		path, _, _ := composeCases(t)
		cases, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		composeMutantsAnswers.expected = goCompose(t, path)
		composeMutantsAnswers.cases = cases
	})
	if composeMutantsAnswers.cases == nil {
		t.Fatal("the Go composer's answers failed earlier in this process")
	}
	return composeMutantsAnswers.cases, composeMutantsAnswers.expected
}

// composeMutantsEnumeration reads the top-level wrappers themselves, so a missing, duplicated or miswired shard can't
// leave a mutant unrun while the count still agrees.
func composeMutantsEnumeration(t *testing.T) []int {
	t.Helper()
	if len(composeMutants.mutations) != composeMutantsShards {
		t.Fatalf("%d compose mutants, but composeMutantsShards is %d", len(composeMutants.mutations), composeMutantsShards)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "compose_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var plan []int
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestComposeMutants_") {
			continue
		}
		shard, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestComposeMutants_"))
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "composeMutantsShard" {
				return true
			}
			if len(call.Args) != 2 {
				t.Fatalf("%s: invalid shard call", function.Name.Name)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				t.Fatalf("%s: shard index must be a literal", function.Name.Name)
			}
			if index, err := strconv.Atoi(literal.Value); err != nil || index != shard {
				t.Fatalf("%s runs mutant %s", function.Name.Name, literal.Value)
			}
			calls++
			return true
		})
		if calls != 1 {
			t.Fatalf("%s runs %d mutants, want one", function.Name.Name, calls)
		}
		plan = append(plan, shard)
	}
	return plan
}

// composeMutantSurvived is a side's verdict: a mutant whose output still equals the Go composer's was missed.
func composeMutantSurvived(side string, output, expected []byte) error {
	if bytes.Equal(output, expected) {
		return fmt.Errorf("%s missed mutant", side)
	}
	return nil
}

func composeMutantsShard(t *testing.T, index int) {
	t.Helper()
	setup := time.Now()
	mutant := composeMutants.mutations[index]
	product := composeMutants.product(t, index)
	corpus, expected := composeMutantsExpected(t)
	cases := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, corpus, 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s; setup %.3fs", mutant.name, time.Since(setup).Seconds())
	nativeOut := run(t, "", nil, filepath.Join(product, "mutant"), cases)
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "compose_main.ts"), cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native", nativeOut}, {"Node", node}} {
		if err := composeMutantSurvived(side.name, side.out, expected); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
	}
}

func TestComposeMutantsUnion(t *testing.T) {
	t.Parallel()
	visits := make([]int, len(composeMutants.mutations))
	for _, index := range composeMutantsEnumeration(t) {
		if index < 0 || index >= len(visits) {
			t.Fatalf("TestComposeMutants_%03d names no mutant", index)
		}
		visits[index]++
	}
	for index, count := range visits {
		if count != 1 {
			t.Fatalf("mutant %d (%s) run by %d top-level tests, want one", index, composeMutants.mutations[index].name, count)
		}
	}
	t.Logf("union: all %d compose mutants run, each by exactly one top-level test", len(visits))
}

func TestComposeMutantsPlantedFailure(t *testing.T) {
	t.Parallel()
	// Plant a survivor in one mutant's output. Exactly its owner must reject it; every other mutant differs from the Go
	// composer and stays accepted.
	var caught []int
	for _, shard := range composeMutantsEnumeration(t) {
		output := []byte("wrong bytes")
		if shard == 1 {
			output = []byte("Go composer bytes")
		}
		if composeMutantSurvived("native", output, []byte("Go composer bytes")) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 1 {
		t.Fatalf("planted survivor caught by %v, want [1]", caught)
	}
	t.Log("planted survivor caught exactly by TestComposeMutants_001")
}

func TestComposeMutants_000(t *testing.T) { t.Parallel(); composeMutantsShard(t, 0) }
func TestComposeMutants_001(t *testing.T) { t.Parallel(); composeMutantsShard(t, 1) }
func TestComposeMutants_002(t *testing.T) { t.Parallel(); composeMutantsShard(t, 2) }
