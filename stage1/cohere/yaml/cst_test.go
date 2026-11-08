package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func goCst(test *testing.T, cases string) []byte {
	test.Helper()
	goBinary := goCstOracle(test)
	return run(test, "", nil, goBinary, cases)
}

// goCstOracle is formatter_comparison over testdata/cst_go.go, a product built ahead (#5qykzj5).
func goCstOracle(test *testing.T) string { return oracles.ConcreteSyntaxTree(test) }

func TestProduct_YamlFindsCstGo(test *testing.T) {
	test.Parallel()
	goCstOracle(test)
}

func TestCstMatchesGo(test *testing.T) {
	test.Parallel()
	binary := yamlCstBuild.binary(test, "cst_main.ts")
	emitted := yamlCstBuild.emitted(test, "cst_main.ts")
	cases, files, count := lexCases(test)
	expected := goCst(test, cases)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs("cst_main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	nativeOut := run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	backend := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", nativeOut}, {"Node", node}, {"emitted JavaScript", backend}} {
		if !bytes.Equal(side.out, expected) {
			test.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	external := run(test, "", nil, "node", "testdata/cst_library.mjs", library, cases)
	if !bytes.Equal(external, expected) {
		test.Fatal(firstDifference(external, expected))
	}
	test.Logf(
		"%d repository files, %d cases, %d CST answer bytes identical on all five sides",
		files,
		count,
		len(expected),
	)
}

var cstMutants = yamlMutants{
	family:  "cst",
	entry:   "cst_main.ts",
	files:   []string{"lexer.ts", "cst.ts", "collectionItem.ts", "cstParser.ts", "cst_main.ts"},
	mutated: "cstParser.ts",
	mutations: []yamlMutation{
		{
			"null key absent",
			"item.key = key;\n    item.keyPresent = true;",
			"item.key = key;\n    item.keyPresent = false;",
		},
		{
			"source offset advanced",
			"return this.make(this.type, this.offset, this.indent, this.source, true);",
			"return this.make(this.type, this.offset + 1, this.indent, this.source, true);",
		},
		{
			"line start shifted",
			"this.lineStarts.push(this.offset + source.length);",
			"this.lineStarts.push(this.offset + source.length - 1);",
		},
	},
}

func TestProduct_YamlcstMutant000(test *testing.T) { test.Parallel(); cstMutants.product(test, 0) }
func TestProduct_YamlcstMutant001(test *testing.T) { test.Parallel(); cstMutants.product(test, 1) }
func TestProduct_YamlcstMutant002(test *testing.T) { test.Parallel(); cstMutants.product(test, 2) }

func TestCstMutants(test *testing.T) {
	test.Parallel()
	cases, _, _ := lexCases(test)
	expected := goCst(test, cases)
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for index, mutant := range cstMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		test.Run(mutant.name, func(test *testing.T) {
			product := cstMutants.product(test, index)
			nativeOut := run(test, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(product, "sources", "cst_main.ts"),
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
