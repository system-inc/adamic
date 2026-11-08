package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func schemaCases(test *testing.T) (string, int, int) {
	path, files, count := lexCases(test)
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	var extra strings.Builder
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	values := []string{
		"",
		"~",
		"null",
		"NULL",
		"Null",
		"NuLl",
		"true",
		"True",
		"TRUE",
		"false",
		"yes",
		"Yes",
		"YES",
		"on",
		"OFF",
		"n",
		"00",
		"07",
		"08",
		"0b101",
		"0o17",
		"0xFF",
		"1_000",
		"+1",
		"-0",
		".0",
		"1.",
		".",
		"1e3",
		"1_e+3",
		".Inf",
		"-.inf",
		".nan",
		".NAN",
		"12:34",
		"12:34:56.70",
		"2000-01-01",
		"2000-1-1T0:0:0Z",
		"0000-99-99 24:59:59.123+2:30",
		"<<",
		"2000-01-01x",
	}
	for _, value := range values {
		for _, before := range []string{"", "+", "-", "x"} {
			for _, after := range []string{"", "x", "_", ".0"} {
				extra.WriteString("0\t" + escape.Replace(before+value+after) + "\n")
				count++
			}
		}
	}
	target := filepath.Join(test.TempDir(), "schema-cases.txt")
	if writeError := os.WriteFile(target, append(data, []byte(extra.String())...), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return target, files, count
}

func goSchema(test *testing.T, cases string) []byte {
	test.Helper()
	goBinary := goSchemaOracle(test)
	return run(test, "", nil, goBinary, cases)
}

// goSchemaOracle is formatter_comparison over testdata/schema_go.go, a product built ahead (#5qykzj5).
func goSchemaOracle(test *testing.T) string { return oracles.Schema(test) }

func TestProduct_YamlFindsSchemaGo(test *testing.T) {
	test.Parallel()
	goSchemaOracle(test)
}

func TestSchemaMatchesGo(test *testing.T) {
	test.Parallel()
	binary := yamlSchemaBuild.binary(test, "schema_main.ts")
	emitted := yamlSchemaBuild.emitted(test, "schema_main.ts")
	cases, files, count := schemaCases(test)
	expected := goSchema(test, cases)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs("schema_main.ts")
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
	external := run(test, "", nil, "node", "testdata/schema_library.mjs", library, cases)
	if !bytes.Equal(external, expected) {
		test.Fatal(firstDifference(external, expected))
	}
	test.Logf(
		"%d repository files, %d cases, %d Schema answer bytes identical on all five sides",
		files,
		count,
		len(expected),
	)
}

var schemaMutants = yamlMutants{
	family:  "schema",
	entry:   "schema_main.ts",
	files:   []string{"patternNode.ts", "schemaPattern.ts", "schemaTag.ts", "schemaTags.ts", "schema_main.ts"},
	mutated: "schemaPattern.ts",
	mutations: []yamlMutation{
		{"class range loses last character", "code <= last.charCodeAt(0)", "code < last.charCodeAt(0)"},
		{"plus accepts empty", "minimum = quantifier === '+' ? 1 : 0;", "minimum = quantifier === '+' ? 0 : 0;"},
		{"optional accepts twice", "maximum = quantifier === '?' ? 1 : -1;", "maximum = quantifier === '?' ? 2 : -1;"},
		{
			"nullable prefix loses successors",
			"node.kind === 'concat' && !this.match(child, '', 0).includes(0)",
			"node.kind === 'concat'",
		},
	},
}

func TestProduct_YamlSchemaMutant000(test *testing.T) {
	test.Parallel()
	schemaMutants.product(test, 0)
}
func TestProduct_YamlSchemaMutant001(test *testing.T) {
	test.Parallel()
	schemaMutants.product(test, 1)
}
func TestProduct_YamlSchemaMutant002(test *testing.T) {
	test.Parallel()
	schemaMutants.product(test, 2)
}
func TestProduct_YamlSchemaMutant003(test *testing.T) {
	test.Parallel()
	schemaMutants.product(test, 3)
}

func TestSchemaMutants(test *testing.T) {
	test.Parallel()
	cases, _, _ := schemaCases(test)
	expected := goSchema(test, cases)
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for index, mutant := range schemaMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		test.Run(mutant.name, func(test *testing.T) {
			product := schemaMutants.product(test, index)
			nativeOut := run(test, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(product, "sources", "schema_main.ts"),
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
