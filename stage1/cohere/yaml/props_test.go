package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func goProps(t *testing.T, cases string) []byte {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := filepath.Abs("testdata/props_go.go")
	if err != nil {
		t.Fatal(err)
	}
	numbers := filepath.Join(root, "cohere/internal/format/yaml/compose/numbers.go")
	original, err := os.ReadFile(numbers)
	if err != nil {
		t.Fatal(err)
	}
	exports, err := os.ReadFile("testdata/props_exports.txt")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(original), "import (", "import (\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"", 1) + string(exports)
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
	binary := filepath.Join(t.TempDir(), "go-props")
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
		}{{"go-props", executable, 0755}, {"props-cases.txt", inputs, 0644}, {"props-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}
// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units); fixed filenames in ADAMIC_YAML_ARTIFACTS; fixed /tmp/stage1-yaml-props-* diagnostic files.
func TestPropsMatchGo(t *testing.T) {
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
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "props")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	emitted := filepath.Join(t.TempDir(), "props.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
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

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units); fixed filenames in ADAMIC_YAML_ARTIFACTS.
func TestPropsMutants(t *testing.T) {
	cases, _, _ := lexCases(t)
	expected := goProps(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, from, to string }{
		{"indicator presence lost", "result.found = index;", "result.found = -1;"},
		{"anchor separation unchecked", "requireSpace = true;", "requireSpace = false;"},
		{"tab indentation accepted", "tab = index;", "tab = -1;"},
	} {
		// Not parallel: native.Build writes the shared adamic/runtime or adamic/units cache.
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range []string{"lexer.ts", "cst.ts", "collectionItem.ts", "cstParser.ts", "props.ts", "propsResolver.ts", "scalarResolution.ts", "composeError.ts", "props_main.ts"} {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "propsResolver.ts" {
					if strings.Count(string(source), mutant.from) != 1 {
						t.Fatal("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, "props_main.ts")
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
