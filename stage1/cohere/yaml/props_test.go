package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func goProps(t *testing.T, cases string) []byte {
	t.Helper()
	binary := goPropsOracle(t)
	expected := run(t, "", nil, binary, cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		yamlArtifacts.Lock()
		defer yamlArtifacts.Unlock()
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		executable, err := os.ReadFile(binary)
		if err != nil {
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
		}{{"go-props", executable, 0755}, {"props-cases.txt", inputs, 0644}, {"props-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}

// goPropsOracle is formatter_comparison over testdata/props_go.go, with compose's numbers.go opened to it by
// testdata/props_exports.txt, a product built ahead (#5qykzj5).
func goPropsOracle(t *testing.T) string {
	t.Helper()
	return formatterComparison(t, "go-props", "props_go.go", numbersExports(t, "\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"", "props_exports.txt")...)
}

func TestProduct_YAMLPropsGo(t *testing.T) {
	t.Parallel()
	goPropsOracle(t)
}

// TestPropsMatchGo reads props' port build (port_products_test.go) and writes /tmp/stage1-yaml-props-*, which only it
// writes, on a failure.
func TestPropsMatchGo(t *testing.T) {
	t.Parallel()
	binary := yamlPropsBuild.binary(t, "props_main.ts")
	emitted := yamlPropsBuild.emitted(t, "props_main.ts")
	cases, files, count := lexCases(t)
	expected := goProps(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("props_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	backend := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/props_library.mjs", library, cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", actual}, {"Node", node}, {"emitted JavaScript", backend}, {"yaml@2.9.0", external}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile("/tmp/stage1-yaml-props-expected.txt", expected, 0644)
			os.WriteFile("/tmp/stage1-yaml-props-actual.txt", side.out, 0644)
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	t.Logf("%d repository files, %d cases, %d props answer bytes identical on all five sides", files, count, len(expected))
}

var propsMutants = yamlMutants{family: "props", entry: "props_main.ts", files: []string{"lexer.ts", "cst.ts", "collectionItem.ts", "cstParser.ts", "props.ts", "propsResolver.ts", "scalarResolution.ts", "composeError.ts", "props_main.ts"}, mutated: "propsResolver.ts", mutations: []yamlMutation{
	{"indicator presence lost", "result.found = index;", "result.found = -1;"},
	{"anchor separation unchecked", "requireSpace = true;", "requireSpace = false;"},
	{"tab indentation accepted", "tab = index;", "tab = -1;"},
}}

func TestProduct_YAMLPropsMutant000(t *testing.T) { t.Parallel(); propsMutants.product(t, 0) }
func TestProduct_YAMLPropsMutant001(t *testing.T) { t.Parallel(); propsMutants.product(t, 1) }
func TestProduct_YAMLPropsMutant002(t *testing.T) { t.Parallel(); propsMutants.product(t, 2) }

func TestPropsMutants(t *testing.T) {
	t.Parallel()
	cases, _, _ := lexCases(t)
	expected := goProps(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for index, mutant := range propsMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		t.Run(mutant.name, func(t *testing.T) {
			product := propsMutants.product(t, index)
			nativeOut := run(t, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "props_main.ts"), cases)
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
