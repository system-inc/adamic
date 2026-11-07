package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func goProps(test *testing.T, cases string) []byte {
	test.Helper()
	binary := goPropsOracle(test)
	expected := run(test, "", nil, binary, cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		yamlArtifacts.Lock()
		defer yamlArtifacts.Unlock()
		if createDirectoryError := os.MkdirAll(artifacts, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		executable, readError := os.ReadFile(binary)
		if readError != nil {
			test.Fatal(readError)
		}
		inputs, readError := os.ReadFile(cases)
		if readError != nil {
			test.Fatal(readError)
		}
		for _, file := range []struct {
			name string
			data []byte
			mode os.FileMode
		}{{"go-props", executable, 0755}, {"props-cases.txt", inputs, 0644}, {"props-expected.txt", expected, 0644}} {
			if writeError := os.WriteFile(
				filepath.Join(artifacts, file.name),
				file.data,
				file.mode,
			); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	return expected
}

// goPropsOracle is formatter_comparison over testdata/props_go.go, with compose's numbers.go opened to it by
// testdata/props_exports.txt, a product built ahead (#5qykzj5).
func goPropsOracle(test *testing.T) string { return oracles.Properties(test) }

func TestProduct_YamlFindsPropsGo(test *testing.T) {
	test.Parallel()
	goPropsOracle(test)
}

// TestPropsMatchGo reads props' port build (port_products_test.go) and writes /tmp/stage1-yaml-props-*, which only it
// writes, on a failure.
func TestPropsMatchGo(test *testing.T) {
	test.Parallel()
	binary := yamlPropsBuild.binary(test, "props_main.ts")
	emitted := yamlPropsBuild.emitted(test, "props_main.ts")
	cases, files, count := lexCases(test)
	expected := goProps(test, cases)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs("props_main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	actual := run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	backend := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	external := run(test, "", nil, "node", "testdata/props_library.mjs", library, cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", actual}, {"Node", node}, {"emitted JavaScript", backend}, {"yaml@2.9.0", external}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile("/tmp/stage1-yaml-props-expected.txt", expected, 0644)
			os.WriteFile("/tmp/stage1-yaml-props-actual.txt", side.out, 0644)
			test.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	test.Logf(
		"%d repository files, %d cases, %d props answer bytes identical on all five sides",
		files,
		count,
		len(expected),
	)
}

var propsMutants = yamlMutants{family: "props", entry: "props_main.ts", files: []string{
	"lexer.ts",
	"cst.ts",
	"collectionItem.ts",
	"cstParser.ts",
	"props.ts",
	"propsResolver.ts",
	"scalarResolution.ts",
	"composeError.ts",
	"props_main.ts",
}, mutated: "propsResolver.ts", mutations: []yamlMutation{
	{"indicator presence lost", "result.found = index;", "result.found = -1;"},
	{"anchor separation unchecked", "requireSpace = true;", "requireSpace = false;"},
	{"tab indentation accepted", "tab = index;", "tab = -1;"},
}}

func TestProduct_YAMLPropsMutant000(test *testing.T) { test.Parallel(); propsMutants.product(test, 0) }
func TestProduct_YAMLPropsMutant001(test *testing.T) { test.Parallel(); propsMutants.product(test, 1) }
func TestProduct_YAMLPropsMutant002(test *testing.T) { test.Parallel(); propsMutants.product(test, 2) }

func TestPropsMutants(test *testing.T) {
	test.Parallel()
	cases, _, _ := lexCases(test)
	expected := goProps(test, cases)
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for index, mutant := range propsMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		test.Run(mutant.name, func(test *testing.T) {
			product := propsMutants.product(test, index)
			nativeOut := run(test, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(product, "sources", "props_main.ts"),
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
