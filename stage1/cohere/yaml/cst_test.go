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

func goCST(t *testing.T, cases string) []byte {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("testdata/cst_go.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	goBinary := filepath.Join(t.TempDir(), "go-cst")
	run(t, filepath.Join(root, "cohere"), nil, "go", "build", "-overlay", path, "-o", goBinary, "./command/formatter_comparison")
	return run(t, "", nil, goBinary, cases)
}

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units).
func TestCSTMatchesGo(t *testing.T) {
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
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "cst")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	nativeOut := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	emitted := filepath.Join(t.TempDir(), "cst.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
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

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units).
func TestCSTMutants(t *testing.T) {
	cases, _, _ := lexCases(t)
	expected := goCST(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, from, to string }{
		{"null key absent", "item.key = key;\n    item.keyPresent = true;", "item.key = key;\n    item.keyPresent = false;"},
		{"source offset advanced", "return this.make(this.type, this.offset, this.indent, this.source, true);", "return this.make(this.type, this.offset + 1, this.indent, this.source, true);"},
		{"line start shifted", "this.lineStarts.push(this.offset + source.length);", "this.lineStarts.push(this.offset + source.length - 1);"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range []string{"lexer.ts", "cst.ts", "collectionItem.ts", "cstParser.ts", "cst_main.ts"} {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "cstParser.ts" {
					if strings.Count(string(source), mutant.from) != 1 {
						t.Fatal("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, "cst_main.ts")
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
