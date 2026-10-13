package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		}{{"go-compose", executable, 0755}, {"compose-cases.txt", inputs, 0644}, {"compose-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}

// goComposeOracle is formatter_comparison over testdata/compose_go.go, with compose's numbers.go opened to it by
// testdata/compose_exports.txt, a product built ahead (#5qykzj5).
func goComposeOracle(t *testing.T) string {
	t.Helper()
	return formatterComparison(t, "go-compose", "compose_go.go", numbersExports(t, "\n\"slices\"\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"", "compose_exports.txt")...)
}

func TestProduct_YAMLComposeGo(t *testing.T) {
	t.Parallel()
	goComposeOracle(t)
}
func composeCases(t *testing.T) (string, int, int) {
	path, files, count := lexCases(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var extra strings.Builder
	for _, version := range []string{"", "%YAML 1.1\n---\n", "%YAML 1.2\n---\n"} {
		for _, tag := range []string{"!!str", "!!int", "!!bool", "!!null", "!!timestamp", "!!binary", "!!merge", "!!set", "!!pairs", "!!omap", "!!map", "!!seq", "!<toString>", "!<constructor>", "!<__proto__>", "!<unknown>"} {
			for _, value := range []string{"word", "2000-01-01", "null", "<<", "{}", "[]", "{a: null}", "{a: 1}", "[{a: 1}, {a: 2}]", "[{}, word]"} {
				extra.WriteString("0\t" + escape.Replace(version+tag+" "+value+"\n") + "\n")
				count++
			}
		}
	}
	for _, text := range []string{"%TAG !e! tag:example.com,2020:\n---\n!e!%E4%B8%AD value\n", "%TAG !e! tag:example.com,2020:\n---\n!e!%FF value\n", "--- a\n...\n# end\n--- b\n", "---\n" + strings.Repeat("k", 1025) + ": v\n"} {
		extra.WriteString("0\t" + escape.Replace(text) + "\n")
		count++
	}
	target := filepath.Join(t.TempDir(), "compose-cases.txt")
	if err := os.WriteFile(target, append(data, []byte(extra.String())...), 0644); err != nil {
		t.Fatal(err)
	}
	return target, files, count
}

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

func TestComposeMutants(t *testing.T) {
	t.Parallel()
	cases, _, _ := composeCases(t)
	expected := goCompose(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for index, mutant := range composeMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		t.Run(mutant.name, func(t *testing.T) {
			product := composeMutants.product(t, index)
			nativeOut := run(t, "", nil, filepath.Join(product, "mutant"), cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "compose_main.ts"), cases)
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
