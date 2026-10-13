package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func goCST(t *testing.T, cases string) []byte {
	t.Helper()
	goBinary := goCSTOracle(t)
	return run(t, "", nil, goBinary, cases)
}

// goCSTOracle is formatter_comparison over testdata/cst_go.go, a product built ahead (#5qykzj5).
func goCSTOracle(t *testing.T) string {
	t.Helper()
	return formatterComparison(t, "go-cst", "cst_go.go")
}

func TestProduct_YAMLCSTGo(t *testing.T) {
	t.Parallel()
	goCSTOracle(t)
}

func TestCSTMatchesGo(t *testing.T) {
	t.Parallel()
	binary := yamlCSTBuild.binary(t, "cst_main.ts")
	emitted := yamlCSTBuild.emitted(t, "cst_main.ts")
	cases, files, count := lexCases(t)
	expected := goCST(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("cst_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	nativeOut := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	backend := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", nativeOut}, {"Node", node}, {"emitted JavaScript", backend}} {
		if !bytes.Equal(side.out, expected) {
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/cst_library.mjs", library, cases)
	if !bytes.Equal(external, expected) {
		t.Fatal(firstDifference(external, expected))
	}
	t.Logf("%d repository files, %d cases, %d CST answer bytes identical on all five sides", files, count, len(expected))
}

var cstMutants = yamlMutants{family: "cst", entry: "cst_main.ts", files: []string{"lexer.ts", "cst.ts", "collectionItem.ts", "cstParser.ts", "cst_main.ts"}, mutated: "cstParser.ts", mutations: []yamlMutation{
	{"null key absent", "item.key = key;\n    item.keyPresent = true;", "item.key = key;\n    item.keyPresent = false;"},
	{"source offset advanced", "return this.make(this.type, this.offset, this.indent, this.source, true);", "return this.make(this.type, this.offset + 1, this.indent, this.source, true);"},
	{"line start shifted", "this.lineStarts.push(this.offset + source.length);", "this.lineStarts.push(this.offset + source.length - 1);"},
}}

func TestProduct_YAMLCSTMutant000(t *testing.T) { t.Parallel(); cstMutants.product(t, 0) }
func TestProduct_YAMLCSTMutant001(t *testing.T) { t.Parallel(); cstMutants.product(t, 1) }
func TestProduct_YAMLCSTMutant002(t *testing.T) { t.Parallel(); cstMutants.product(t, 2) }

func TestCSTMutants(t *testing.T) {
	t.Parallel()
	cases, _, _ := lexCases(t)
	expected := goCST(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for index, mutant := range cstMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		t.Run(mutant.name, func(t *testing.T) {
			product := cstMutants.product(t, index)
			nativeOut := run(t, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "cst_main.ts"), cases)
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
