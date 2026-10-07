package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func goCompose(t *testing.T, cases string) []byte {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := filepath.Abs("testdata/compose_go.go")
	if err != nil {
		t.Fatal(err)
	}
	numbers := filepath.Join(root, "cohere/internal/format/yaml/compose/numbers.go")
	original, err := os.ReadFile(numbers)
	if err != nil {
		t.Fatal(err)
	}
	exports, err := os.ReadFile("testdata/compose_exports.txt")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(original), "import (", "import (\n\"slices\"\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"", 1) + string(exports)
	replacement := filepath.Join(t.TempDir(), "numbers.go")
	if err := os.WriteFile(replacement, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): adapter, numbers: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-compose")
	run(t, filepath.Join(root, "cohere"), nil, "go", "build", "-overlay", path, "-o", binary, "./command/formatter_comparison")
	expected := run(t, "", nil, binary, cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
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

func TestComposeMatchGo(t *testing.T) {
	cases, files, count := composeCases(t)
	expected := goCompose(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("compose_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "compose")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	emitted := filepath.Join(t.TempDir(), "compose.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	backend := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until cloud/setup.sh installs it")
	}
	external := run(t, "", nil, "node", "testdata/compose_library.mjs", library, cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", actual}, {"Node", node}, {"emitted JavaScript", backend}, {"yaml@2.9.0", external}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile("/tmp/stage1-yaml-compose-expected.txt", expected, 0644)
			os.WriteFile("/tmp/stage1-yaml-compose-actual.txt", side.out, 0644)
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	t.Logf("%d repository files, %d cases, %d compose answer bytes identical on all five sides", files, count, len(expected))
}

func TestComposeMutants(t *testing.T) {
	cases, _, _ := composeCases(t)
	expected := goCompose(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, from, to string }{
		{"document start marker lost", "this.document.directives.docStart = props.found >= 0;", "this.document.directives.docStart = false;"},
		{"block mapping pair lost", "this.node(result).items.push(this.pair(key, value, index, itemIndex));", "this.node(result).items.push(this.pair(key, -1, index, itemIndex));"},
		{"implicit key limit skipped", "properties.start < this.parser.get(valueProps.found).offset - 1024", "properties.start < this.parser.get(valueProps.found).offset - 999999"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			entries, err := filepath.Glob("*.ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range entries {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "composer.ts" {
					if strings.Count(string(source), mutant.from) != 1 {
						t.Fatal("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, "compose_main.ts")
			program, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(lowered), binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			nativeOut := run(t, "", nil, binary, cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
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
