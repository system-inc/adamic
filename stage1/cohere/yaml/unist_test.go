package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func goUnist(test *testing.T, cases string) []byte {
	test.Helper()
	goBinary := goUnistOracle(test)
	return run(test, "", nil, goBinary, cases)
}

// goUnistOracle is formatter_comparison over testdata/unist_go.go, a product built ahead (#5qykzj5).
func goUnistOracle(test *testing.T) string { return oracles.Unist(test) }

func TestProduct_YamlFindsUnistGo(test *testing.T) {
	test.Parallel()
	goUnistOracle(test)
}

// unistMatchGo compares unist's port build (port_products_test.go) with Go over composeCases, in two shards.
var unistMatchGo = yamlMatchGo{
	family: "unist",
	shards: 2,
	cases:  composeCases,
	oracle: goUnistOracle,
	build:  yamlUnistBuild,
	entry:  "unist_main.ts",
}

func TestUnistMatchesGo_000(test *testing.T) { test.Parallel(); unistMatchesGoShard(test, 0) }
func TestUnistMatchesGo_001(test *testing.T) { test.Parallel(); unistMatchesGoShard(test, 1) }

func TestUnistMatchesGoUnion(test *testing.T) {
	test.Parallel()
	unistMatchGo.union(test)
}

func TestUnistMatchesGoPlantedFailure(test *testing.T) {
	test.Parallel()
	unistMatchGo.planted(test)
}

func unistMatchesGoShard(test *testing.T, shard int) {
	test.Helper()
	result := unistMatchGo.run(test, shard)
	expected := result.expected
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", result.native}, {"Node", result.node}, {"emitted JavaScript", result.emitted}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile(unistMatchGo.diagnostic(shard, "expected"), expected, 0644)
			os.WriteFile(unistMatchGo.diagnostic(shard, "actual"), side.out, 0644)
			test.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	external := run(test, "", nil, "node", "testdata/unist_library.mjs", library, result.path)
	if !bytes.Equal(external, expected) {
		test.Fatal(firstDifference(external, expected))
	}
	test.Logf(
		"%d repository files, %d of %d cases, %d unist answer bytes identical on all five sides",
		result.files,
		len(result.indices),
		len(result.all),
		len(expected),
	)
}

var unistMutants = yamlMutants{
	family:  "unist",
	entry:   "unist_main.ts",
	mutated: "unistContext.ts",
	mutations: []yamlMutation{
		{
			"comment prefix lost",
			"this.node(result).value = token.source.slice(1);",
			"this.node(result).value = token.source.slice(2);",
		},
		{"point column shifted", "new UnistPoint(line, column, offset)", "new UnistPoint(line, column + 1, offset)"},
		{
			"document end marker lost",
			"this.node(result).documentEndMarker = data.end >= 0;",
			"this.node(result).documentEndMarker = false;",
		},
	},
}

func TestProduct_YamlUnistMutant000(test *testing.T) { test.Parallel(); unistMutants.product(test, 0) }
func TestProduct_YamlUnistMutant001(test *testing.T) { test.Parallel(); unistMutants.product(test, 1) }
func TestProduct_YamlUnistMutant002(test *testing.T) { test.Parallel(); unistMutants.product(test, 2) }

func TestUnistMutants(test *testing.T) {
	test.Parallel()
	cases, _, _ := composeCases(test)
	expected := goUnist(test, cases)
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for index, mutant := range unistMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		test.Run(mutant.name, func(test *testing.T) {
			product := unistMutants.product(test, index)
			nativeOut := run(test, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(product, "sources", "unist_main.ts"),
				cases,
			)
			for _, side := range []struct {
				name string
				out  []byte
			}{{"native", nativeOut}, {"Node", node}} {
				if bytes.Equal(side.out, expected) {
					test.Fatalf("%s missed mutant", side.name)
				}
				test.Logf(
					"%s successful execution, wrong bytes caught: %s",
					side.name,
					firstDifference(side.out, expected),
				)
			}
		})
	}
}
