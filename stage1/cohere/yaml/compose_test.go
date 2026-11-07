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

func goCompose(test *testing.T, cases string) []byte {
	test.Helper()
	binary := goComposeOracle(test)
	expected := run(test, "", nil, binary, cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		yamlArtifacts.Lock()
		defer yamlArtifacts.Unlock()
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			test.Fatal(err)
		}
		program, err := os.ReadFile(binary)
		if err != nil {
			test.Fatal(err)
		}
		// A program copied while sibling tests fork goes through the fork-safe writer, or a fork can hold its writable
		// descriptor and a later exec of it fails with ETXTBSY (#6hjazy4).
		if err := executable.Write(filepath.Join(artifacts, "go-compose"), program); err != nil {
			test.Fatal(err)
		}
		inputs, err := os.ReadFile(cases)
		if err != nil {
			test.Fatal(err)
		}
		for _, file := range []struct {
			name string
			data []byte
			mode os.FileMode
		}{{"compose-cases.txt", inputs, 0644}, {"compose-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				test.Fatal(err)
			}
		}
	}
	return expected
}

// goComposeOracle is formatter_comparison over testdata/compose_go.go, with compose's numbers.go opened to it by
// testdata/compose_exports.txt, a product built ahead (#5qykzj5).
func goComposeOracle(test *testing.T) string { return oracles.Composition(test) }

func TestProduct_YamlFindsComposeGo(test *testing.T) {
	test.Parallel()
	goComposeOracle(test)
}
func composeCases(test *testing.T) (string, int, int) { return oracles.CompositionCases(test) }

// composeMatchGo compares compose's port build (port_products_test.go) with Go over composeCases, in two shards.
var composeMatchGo = yamlMatchGo{
	family: "compose",
	shards: 2,
	cases:  composeCases,
	oracle: goComposeOracle,
	build:  yamlComposeBuild,
	entry:  "compose_main.ts",
}

func TestComposeMatchGo_000(test *testing.T) { test.Parallel(); composeMatchGoShard(test, 0) }
func TestComposeMatchGo_001(test *testing.T) { test.Parallel(); composeMatchGoShard(test, 1) }

func TestComposeMatchGoUnion(test *testing.T) {
	test.Parallel()
	composeMatchGo.union(test)
}

func TestComposeMatchGoPlantedFailure(test *testing.T) {
	test.Parallel()
	composeMatchGo.planted(test)
}

func composeMatchGoShard(test *testing.T, shard int) {
	test.Helper()
	result := composeMatchGo.run(test, shard)
	expected := result.expected
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	external := run(test, "", nil, "node", "testdata/compose_library.mjs", library, result.path)
	for _, side := range []struct {
		name string
		out  []byte
	}{
		{"native ASan/UBSan/LSan", result.native},
		{"Node", result.node},
		{"emitted JavaScript", result.emitted},
		{"yaml@2.9.0", external},
	} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile(composeMatchGo.diagnostic(shard, "expected"), expected, 0644)
			os.WriteFile(composeMatchGo.diagnostic(shard, "actual"), side.out, 0644)
			test.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	test.Logf(
		"%d repository files, %d of %d cases, %d compose answer bytes identical on all five sides",
		result.files,
		len(result.indices),
		len(result.all),
		len(expected),
	)
}

var composeMutants = yamlMutants{
	family:  "compose",
	entry:   "compose_main.ts",
	mutated: "composer.ts",
	mutations: []yamlMutation{
		{
			"document start marker lost",
			"this.document.directives.docStart = props.found >= 0;",
			"this.document.directives.docStart = false;",
		},
		{
			"block mapping pair lost",
			"this.node(result).items.push(this.pair(key, value, index, itemIndex));",
			"this.node(result).items.push(this.pair(key, -1, index, itemIndex));",
		},
		{
			"implicit key limit skipped",
			"properties.start < this.parser.get(valueProps.found).offset - 1024",
			"properties.start < this.parser.get(valueProps.found).offset - 999999",
		},
	},
}

func TestProduct_YAMLComposeMutant000(test *testing.T) {
	test.Parallel()
	composeMutants.product(test, 0)
}
func TestProduct_YAMLComposeMutant001(test *testing.T) {
	test.Parallel()
	composeMutants.product(test, 1)
}
func TestProduct_YAMLComposeMutant002(test *testing.T) {
	test.Parallel()
	composeMutants.product(test, 2)
}

// composeMutantsShards is the count of composeMutants.mutations, each its own top-level test so Loom's planner can cut
// between mutants, which it can't between subtests.
const composeMutantsShards = 3

// composeMutantsAnswers holds what every mutant compares against, made once per process: the composition cases and
// the Go composer's answer to them.
var composeMutantsAnswers struct {
	once            sync.Once
	cases, expected []byte
}

func composeMutantsExpected(test *testing.T) ([]byte, []byte) {
	test.Helper()
	composeMutantsAnswers.once.Do(func() {
		// The cases file lives in the first caller's temporary directory, gone when that test ends, so its bytes are
		// kept.
		path, _, _ := composeCases(test)
		cases, err := os.ReadFile(path)
		if err != nil {
			test.Fatal(err)
		}
		composeMutantsAnswers.expected = goCompose(test, path)
		composeMutantsAnswers.cases = cases
	})
	if composeMutantsAnswers.cases == nil {
		test.Fatal("the Go composer's answers failed earlier in this process")
	}
	return composeMutantsAnswers.cases, composeMutantsAnswers.expected
}

// composeMutantsEnumeration reads the top-level wrappers themselves, so a missing, duplicated or miswired shard can't
// leave a mutant unrun while the count still agrees.
func composeMutantsEnumeration(test *testing.T) []int {
	test.Helper()
	if len(composeMutants.mutations) != composeMutantsShards {
		test.Fatalf(
			"%d compose mutants, but composeMutantsShards is %d",
			len(composeMutants.mutations),
			composeMutantsShards,
		)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "compose_test.go", nil, 0)
	if err != nil {
		test.Fatal(err)
	}
	var plan []int
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestComposeMutants_") {
			continue
		}
		shard, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestComposeMutants_"))
		if err != nil {
			test.Fatal(err)
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
				test.Fatalf("%s: invalid shard call", function.Name.Name)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				test.Fatalf("%s: shard index must be a literal", function.Name.Name)
			}
			if index, err := strconv.Atoi(literal.Value); err != nil || index != shard {
				test.Fatalf("%s runs mutant %s", function.Name.Name, literal.Value)
			}
			calls++
			return true
		})
		if calls != 1 {
			test.Fatalf("%s runs %d mutants, want one", function.Name.Name, calls)
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

func composeMutantsShard(test *testing.T, index int) {
	test.Helper()
	setup := time.Now()
	mutant := composeMutants.mutations[index]
	product := composeMutants.product(test, index)
	corpus, expected := composeMutantsExpected(test)
	cases := filepath.Join(test.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, corpus, 0644); err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	test.Logf("%s; setup %.3fs", mutant.name, time.Since(setup).Seconds())
	nativeOut := run(test, "", nil, filepath.Join(product, "mutant"), cases)
	node := run(
		test,
		"",
		nil,
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(product, "sources", "compose_main.ts"),
		cases,
	)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native", nativeOut}, {"Node", node}} {
		if err := composeMutantSurvived(side.name, side.out, expected); err != nil {
			test.Fatal(err)
		}
		test.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
	}
}

func TestComposeMutantsUnion(test *testing.T) {
	test.Parallel()
	visits := make([]int, len(composeMutants.mutations))
	for _, index := range composeMutantsEnumeration(test) {
		if index < 0 || index >= len(visits) {
			test.Fatalf("TestComposeMutants_%03d names no mutant", index)
		}
		visits[index]++
	}
	for index, count := range visits {
		if count != 1 {
			test.Fatalf(
				"mutant %d (%s) run by %d top-level tests, want one",
				index,
				composeMutants.mutations[index].name,
				count,
			)
		}
	}
	test.Logf("union: all %d compose mutants run, each by exactly one top-level test", len(visits))
}

func TestComposeMutantsPlantedFailure(test *testing.T) {
	test.Parallel()
	// Plant a survivor in one mutant's output. Exactly its owner must reject it; every other mutant differs from the Go
	// composer and stays accepted.
	var caught []int
	for _, shard := range composeMutantsEnumeration(test) {
		output := []byte("wrong bytes")
		if shard == 1 {
			output = []byte("Go composer bytes")
		}
		if composeMutantSurvived("native", output, []byte("Go composer bytes")) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 1 {
		test.Fatalf("planted survivor caught by %v, want [1]", caught)
	}
	test.Log("planted survivor caught exactly by TestComposeMutants_001")
}

func TestComposeMutants_000(test *testing.T) { test.Parallel(); composeMutantsShard(test, 0) }
func TestComposeMutants_001(test *testing.T) { test.Parallel(); composeMutantsShard(test, 1) }
func TestComposeMutants_002(test *testing.T) { test.Parallel(); composeMutantsShard(test, 2) }
