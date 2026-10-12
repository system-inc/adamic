package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func goUnist(t *testing.T, cases string) []byte {
	t.Helper()
	goBinary := goUnistOracle(t)
	return run(t, "", nil, goBinary, cases)
}

// goUnistOracle is formatter_comparison over testdata/unist_go.go, a product built ahead (#5qykzj5).
func goUnistOracle(t *testing.T) string {
	t.Helper()
	return formatterComparison(t, "go-unist", "unist_go.go")
}

func TestProduct_YAMLUnistGo(t *testing.T) {
	t.Parallel()
	goUnistOracle(t)
}

// unistMatchGo compares unist's port build (port_products_test.go) with Go over composeCases, in two shards.
var unistMatchGo = yamlMatchGo{family: "unist", shards: 2, cases: composeCases, oracle: goUnistOracle, build: yamlUnistBuild, entry: "unist_main.ts"}

func TestUnistMatchesGo_000(t *testing.T) { t.Parallel(); unistMatchesGoShard(t, 0) }
func TestUnistMatchesGo_001(t *testing.T) { t.Parallel(); unistMatchesGoShard(t, 1) }

func TestUnistMatchesGoUnion(t *testing.T) {
	t.Parallel()
	unistMatchGo.union(t)
}

func TestUnistMatchesGoPlantedFailure(t *testing.T) {
	t.Parallel()
	unistMatchGo.planted(t)
}

func unistMatchesGoShard(t *testing.T, shard int) {
	t.Helper()
	result := unistMatchGo.run(t, shard)
	expected := result.expected
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", result.native}, {"Node", result.node}, {"emitted JavaScript", result.emitted}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile(unistMatchGo.diagnostic(shard, "expected"), expected, 0644)
			os.WriteFile(unistMatchGo.diagnostic(shard, "actual"), side.out, 0644)
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/unist_library.mjs", library, result.path)
	if !bytes.Equal(external, expected) {
		t.Fatal(firstDifference(external, expected))
	}
	t.Logf("%d repository files, %d of %d cases, %d unist answer bytes identical on all five sides", result.files, len(result.indices), len(result.all), len(expected))
}

var unistMutants = yamlMutants{family: "unist", entry: "unist_main.ts", mutated: "unistContext.ts", mutations: []yamlMutation{
	{"comment prefix lost", "this.node(result).value = token.source.slice(1);", "this.node(result).value = token.source.slice(2);"},
	{"point column shifted", "new UnistPoint(line, column, offset)", "new UnistPoint(line, column + 1, offset)"},
	{"document end marker lost", "this.node(result).documentEndMarker = data.end >= 0;", "this.node(result).documentEndMarker = false;"},
}}

func TestProduct_YAMLUnistMutant000(t *testing.T) { t.Parallel(); unistMutants.product(t, 0) }
func TestProduct_YAMLUnistMutant001(t *testing.T) { t.Parallel(); unistMutants.product(t, 1) }
func TestProduct_YAMLUnistMutant002(t *testing.T) { t.Parallel(); unistMutants.product(t, 2) }

func TestUnistMutants(t *testing.T) {
	t.Parallel()
	cases, _, _ := composeCases(t)
	expected := goUnist(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for index, mutant := range unistMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		t.Run(mutant.name, func(t *testing.T) {
			product := unistMutants.product(t, index)
			nativeOut := run(t, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "unist_main.ts"), cases)
			for _, side := range []struct {
				name string
				out  []byte
			}{{"native", nativeOut}, {"Node", node}} {
				if bytes.Equal(side.out, expected) {
					t.Fatalf("%s missed mutant", side.name)
				}
				t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
			}
		})
	}
}
