package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func scalarCases(test *testing.T) (string, int, int) {
	path, files, count := lexCases(test)
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	var extra strings.Builder
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	add := func(text string) { extra.WriteString("0\t" + escape.Replace(text) + "\n"); count++ }
	for _, value := range []string{
		`"\0\a\b\e\f\n\r\t\v\N\_\L\P\ \/\\\""`,
		`"\x00\x7f\xff\u0000\ud800\udc00\U0001f600\U0010ffff"`,
		`"\x0"`,
		`"\xgg"`,
		`"\u123"`,
		`"\U00110000"`,
		`"\q"`,
		"\"x\\\n  y\"",
		"\"x\\\r\n\ty\"",
		"'a  \n  b\n\n  c'",
		"\"a\r\n\n b\"",
		"a  \n  b\n\n  c",
	} {
		add(value + " # tail\n")
	}
	for _, mode := range []string{"|", ">"} {
		for _, header := range []string{"", "-", "+", "1", "2-", "-2", "3+", "+3", "0", "++", "12", "+2-"} {
			for _, content := range []string{
				"",
				"\n",
				" \n",
				"  x\n",
				"  x\n\n\n",
				"  a\n    b\n  c\n",
				"  a\n\n  b\n",
				"    \n  x\n",
				"  a\r\n  b\r\n",
				"  \tx\n",
				"x\n",
			} {
				add(mode + header + " # head\n" + content)
				add("a: " + mode + header + " # head\n" + content)
			}
		}
	}
	target := filepath.Join(test.TempDir(), "scalar-cases.txt")
	if writeError := os.WriteFile(target, append(source, []byte(extra.String())...), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return target, files, count
}
func goScalars(test *testing.T, cases string) []byte {
	test.Helper()
	binary := goScalarsOracle(test)
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
		}{
			{"go-scalar", executable, 0755},
			{"scalar-cases.txt", inputs, 0644},
			{"scalar-expected.txt", expected, 0644},
		} {
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

// goScalarsOracle is formatter_comparison over testdata/scalar_go.go, with compose's numbers.go opened to it by
// testdata/scalar_exports.txt, a product built ahead (#5qykzj5).
func goScalarsOracle(test *testing.T) string { return oracles.Scalars(test) }

func TestProduct_YamlFindsScalarsGo(test *testing.T) {
	test.Parallel()
	goScalarsOracle(test)
}

var scalarMutants = yamlMutants{family: "scalar", entry: "scalar_main.ts", files: []string{
	"lexer.ts",
	"cst.ts",
	"collectionItem.ts",
	"cstParser.ts",
	"scalar.ts",
	"scalarText.ts",
	"scalarResolution.ts",
	"composeError.ts",
	"newlineFold.ts",
	"blockLine.ts",
	"scalar_main.ts",
}, mutations: []yamlMutation{
	{
		"escaped line feed becomes carriage return",
		"case 'n':\n            return '\\n';",
		"case 'n':\n            return '\\r';",
	},
	{
		"strip becomes clip",
		"else if(chomp !== '-') result.value += '\\n';",
		"else if(chomp !== '?') result.value += '\\n';",
	},
	{
		"flow line break stops folding",
		"let separator = ' ';\n    let position = first + 1;",
		"let separator = '\\n';\n    let position = first + 1;",
	},
}}

func TestProduct_YamlScalarMutant000(test *testing.T) {
	test.Parallel()
	scalarMutants.product(test, 0)
}
func TestProduct_YamlScalarMutant001(test *testing.T) {
	test.Parallel()
	scalarMutants.product(test, 1)
}
func TestProduct_YamlScalarMutant002(test *testing.T) {
	test.Parallel()
	scalarMutants.product(test, 2)
}

func TestScalarMutants(test *testing.T) {
	test.Parallel()
	cases, _, _ := scalarCases(test)
	expected := goScalars(test, cases)
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for index, mutant := range scalarMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		test.Run(mutant.name, func(test *testing.T) {
			product := scalarMutants.product(test, index)
			nativeOut := run(test, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(product, "sources", "scalar_main.ts"),
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
