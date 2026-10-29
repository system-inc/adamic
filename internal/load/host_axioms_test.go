package load

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
)

func TestHostAxiomsCheckerSeesOptionalTTYMembers(t *testing.T) {
	t.Parallel()
	program, err := nodeSource(t, `import process from 'node:process';
 const outputTTY = process.stdout.isTTY;
 const outputColumns = process.stdout.columns;
 const outputRows = process.stdout.rows;
 const inputTTY = process.stdin.isTTY;
 const inputRaw = process.stdin.isRaw;
 const errorTTY = process.stderr.isTTY;
 const errorColumns = process.stderr.columns;
 const errorRows = process.stderr.rows;
 const colors = process.stdout.hasColors;
 const depth = process.stdout.getColorDepth;
 const windowSize = process.stdout.getWindowSize;
 const rawMode = process.stdin.setRawMode;
 if (process.stdout.hasColors) {
  process.stdout.hasColors();
  process.stdout.hasColors(256);
  process.stdout.hasColors({TERM:'dumb'});
  process.stdout.hasColors(256, {TERM:'dumb'});
 }
 `)
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]string{}
	for _, declaration := range program.Declarations(context.Background()) {
		if declaration.Kind == "variable" {
			types[declaration.Name] = declaration.Type
		}
	}
	for name, want := range map[string]string{
		"outputTTY": "boolean | undefined", "outputColumns": "number | undefined", "outputRows": "number | undefined",
		"inputTTY": "boolean | undefined", "inputRaw": "boolean | undefined",
		"errorTTY": "boolean | undefined", "errorColumns": "number | undefined", "errorRows": "number | undefined",
	} {
		if types[name] != want {
			t.Errorf("%s: checker sees %q, want %q", name, types[name], want)
		}
	}
	for _, name := range []string{"colors", "depth", "windowSize", "rawMode"} {
		if !strings.HasSuffix(types[name], " | undefined") {
			t.Errorf("%s: method is not optional: %q", name, types[name])
		}
	}
}

func TestHostAxiomsPreservePublishedPositions(t *testing.T) {
	t.Parallel()
	rows, err := hostAxioms()
	if err != nil {
		t.Fatal(err)
	}
	files, err := NodeHostAxiomFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := embeddedNodeTypes.ReadFile("node_types/node_modules/@types/node/" + file)
		if err != nil {
			t.Fatal(err)
		}
		served, err := CorrectNodeHostAxioms(file, string(data))
		if err != nil {
			t.Fatal(err)
		}
		before, after := strings.Split(string(data), "\n"), strings.Split(served, "\n")
		if len(before) != len(after) {
			t.Fatalf("%s: line count moved", file)
		}
		corrected := map[int]hostAxiom{}
		for _, row := range rows {
			if row.File == file {
				corrected[row.Line] = row
			}
		}
		for index := range before {
			if row, changed := corrected[index+1]; changed {
				if after[index] != strings.Replace(before[index], row.Published, row.Corrected, 1) {
					t.Fatalf("%s:%d: row differs", file, index+1)
				}
			} else if before[index] != after[index] {
				t.Fatalf("%s:%d: uncorrected line moved", file, index+1)
			}
		}
		fs := &nodeLibraryFS{}
		got, ok := fs.ReadFile(tspath.RootedFilePathFromNormalized(nodeTypesDirectory + "/" + file))
		if !ok || got != served {
			t.Fatalf("%s: bundled checker sees different text", file)
		}
	}
}

func TestHostAxiomsRefusePublishedDrift(t *testing.T) {
	t.Parallel()
	row := hostAxiom{File: "tty.d.ts", Line: 2, Published: "  columns: number;", Corrected: "  columns?: number;"}
	for _, one := range []struct{ name, text, reason string }{
		{"missing", "before\n  columns: string;\nafter", "published line"},
		{"moved", "before\nafter\n  columns: number;", "published line"},
		{"twice on line", "before\n  columns: number;  columns: number;\nafter", "published line"},
		{"twice in file", "before\n  columns: number;\n  columns: number;", "occurrences in the file"},
		{"upstream fixed", "before\n  columns?: number;\nafter", "retire the row"},
		{"fixed elsewhere", "before\n  columns: number;\n  columns?: number;", "retire the row"},
	} {
		t.Run(one.name, func(t *testing.T) {
			_, err := correctNodeHostAxioms(row.File, one.text, []hostAxiom{row})
			if err == nil || !strings.Contains(err.Error(), "row 1 (tty.d.ts:2)") || !strings.Contains(err.Error(), one.reason) {
				t.Fatalf("want located drift error, got %v", err)
			}
		})
	}
}

// GoInputs includes go list's EmbedFiles, so this component of GoBuild's content
// key (and Loom's linked compiler input closure) moves when a ruled row changes.
func TestHostAxiomsAreBuildCacheInputs(t *testing.T) {
	t.Parallel()
	inputs, err := buildcache.GoInputs("host-axioms-load", "./internal/load", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	const table = "internal/load/host_axioms.json"
	if !slices.Contains(inputs.Files, table) {
		t.Fatalf("GoInputs omits %s", table)
	}
	directory := t.TempDir()
	name := filepath.Join(directory, filepath.FromSlash(table))
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	inputs.Files = []string{table}
	key := func(data []byte) string {
		t.Helper()
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
		value, err := buildcache.Key(directory, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := key(hostAxiomsJSON)
	changed := strings.Replace(string(hostAxiomsJSON), "columns?: number;", "columns: number | undefined;", 1)
	if changed == string(hostAxiomsJSON) || key([]byte(changed)) == before {
		t.Fatal("editing a row did not change GoBuild's content key")
	}
}

// Products that run an already linked compiler key compilerinputs.Lowering,
// rather than a later trace of Go compilation. Its directory input includes the
// host table too, even though the loader reads the embedded bytes at run time.
func TestHostAxiomsAreLoweringProductInputs(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	inputs := buildcache.Inputs{Name: "host-axioms-lowering", Files: compilerinputs.Lowering()}
	for _, name := range inputs.Files {
		full := filepath.Join(directory, filepath.FromSlash(name))
		if filepath.Ext(name) == "" {
			if err := os.MkdirAll(full, 0700); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, nil, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	name := filepath.Join(directory, "internal", "load", "host_axioms.json")
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	key := func(data []byte) string {
		t.Helper()
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
		got, err := buildcache.Key(directory, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	before := key(hostAxiomsJSON)
	changed := strings.Replace(string(hostAxiomsJSON), "columns?: number;", "columns: number | undefined;", 1)
	if changed == string(hostAxiomsJSON) || key([]byte(changed)) == before {
		t.Fatal("editing a host row did not change the lowering product key")
	}
}

func TestHostAxiomsMissingFileNamesTheRow(t *testing.T) {
	t.Parallel()
	err := validateNodeHostAxioms(fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "row 1 (tty.d.ts:51)") {
		t.Fatalf("missing host axiom file: %v", err)
	}
}
