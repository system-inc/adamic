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

func TestHostAxiomsCheckerSeesOptionalTtyMembers(test *testing.T) {
	test.Parallel()
	program, nodeSourceError := nodeSource(test, `import process from 'node:process';
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
	if nodeSourceError != nil {
		test.Fatal(nodeSourceError)
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
			test.Errorf("%s: checker sees %q, want %q", name, types[name], want)
		}
	}
	for _, name := range []string{"colors", "depth", "windowSize", "rawMode"} {
		if !strings.HasSuffix(types[name], " | undefined") {
			test.Errorf("%s: method is not optional: %q", name, types[name])
		}
	}
}

func TestHostAxiomsPreservePublishedPositions(test *testing.T) {
	test.Parallel()
	rows, hostAxiomsError := hostAxioms()
	if hostAxiomsError != nil {
		test.Fatal(hostAxiomsError)
	}
	files, nodeHostAxiomFilesError := NodeHostAxiomFiles()
	if nodeHostAxiomFilesError != nil {
		test.Fatal(nodeHostAxiomFilesError)
	}
	for _, file := range files {
		data, readFileError := embeddedNodeTypes.ReadFile("node_types/node_modules/@types/node/" + file)
		if readFileError != nil {
			test.Fatal(readFileError)
		}
		served, correctNodeHostAxiomsError := CorrectNodeHostAxioms(file, string(data))
		if correctNodeHostAxiomsError != nil {
			test.Fatal(correctNodeHostAxiomsError)
		}
		before, after := strings.Split(string(data), "\n"), strings.Split(served, "\n")
		if len(before) != len(after) {
			test.Fatalf("%s: line count moved", file)
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
					test.Fatalf("%s:%d: row differs", file, index+1)
				}
			} else if before[index] != after[index] {
				test.Fatalf("%s:%d: uncorrected line moved", file, index+1)
			}
		}
		fs := &nodeLibraryFileSystem{}
		got, ok := fs.ReadFile(tspath.RootedFilePathFromNormalized(nodeTypesDirectory + "/" + file))
		if !ok || got != served {
			test.Fatalf("%s: bundled checker sees different text", file)
		}
	}
}

func TestHostAxiomsRefusePublishedDrift(test *testing.T) {
	test.Parallel()
	row := hostAxiom{File: "tty.d.ts", Line: 2, Published: "  columns: number;", Corrected: "  columns?: number;"}
	for _, one := range []struct{ name, text, reason string }{
		{"missing", "before\n  columns: string;\nafter", "published line"},
		{"moved", "before\nafter\n  columns: number;", "published line"},
		{"twice on line", "before\n  columns: number;  columns: number;\nafter", "published line"},
		{"twice in file", "before\n  columns: number;\n  columns: number;", "occurrences in the file"},
		{"upstream fixed", "before\n  columns?: number;\nafter", "retire the row"},
		{"fixed elsewhere", "before\n  columns: number;\n  columns?: number;", "retire the row"},
	} {
		test.Run(one.name, func(test *testing.T) {
			_, correctNodeHostAxiomsError := correctNodeHostAxioms(row.File, one.text, []hostAxiom{row})
			if correctNodeHostAxiomsError == nil ||
				!strings.Contains(correctNodeHostAxiomsError.Error(), "row 1 (tty.d.ts:2)") ||
				!strings.Contains(correctNodeHostAxiomsError.Error(), one.reason) {
				test.Fatalf("want located drift error, got %v", correctNodeHostAxiomsError)
			}
		})
	}
}

// GoInputs includes go list's EmbedFiles, so this component of GoBuild's content
// key (and Loom's linked compiler input closure) moves when a ruled row changes.
func TestHostAxiomsAreBuildCacheInputs(test *testing.T) {
	test.Parallel()
	inputs, goInputsError := buildcache.GoInputs("host-axioms-load", "./internal/load", nil, nil)
	if goInputsError != nil {
		test.Fatal(goInputsError)
	}
	const table = "internal/load/host_axioms.json"
	if !slices.Contains(inputs.Files, table) {
		test.Fatalf("GoInputs omits %s", table)
	}
	directory := test.TempDir()
	name := filepath.Join(directory, filepath.FromSlash(table))
	if createDirectoryError := os.MkdirAll(filepath.Dir(name), 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	inputs.Files = []string{table}
	key := func(data []byte) string {
		test.Helper()
		if writeError := os.WriteFile(name, data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		value, keyError := buildcache.Key(directory, inputs)
		if keyError != nil {
			test.Fatal(keyError)
		}
		return value
	}
	before := key(hostAxiomsJson)
	changed := strings.Replace(string(hostAxiomsJson), "columns?: number;", "columns: number | undefined;", 1)
	if changed == string(hostAxiomsJson) || key([]byte(changed)) == before {
		test.Fatal("editing a row did not change GoBuild's content key")
	}
}

// Products that run an already linked compiler key compilerinputs.Lowering,
// rather than a later trace of Go compilation. Its directory input includes the
// host table too, even though the loader reads the embedded bytes at run time.
func TestHostAxiomsAreLoweringProductInputs(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	inputs := buildcache.Inputs{Name: "host-axioms-lowering", Files: compilerinputs.Lowering()}
	for _, name := range inputs.Files {
		full := filepath.Join(directory, filepath.FromSlash(name))
		if filepath.Ext(name) == "" {
			if createDirectoryError := os.MkdirAll(full, 0700); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
		} else {
			if createDirectoryError := os.MkdirAll(filepath.Dir(full), 0700); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if writeError := os.WriteFile(full, nil, 0600); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	name := filepath.Join(directory, "internal", "load", "host_axioms.json")
	if createDirectoryError := os.MkdirAll(filepath.Dir(name), 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	key := func(data []byte) string {
		test.Helper()
		if writeError := os.WriteFile(name, data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		got, keyError := buildcache.Key(directory, inputs)
		if keyError != nil {
			test.Fatal(keyError)
		}
		return got
	}
	before := key(hostAxiomsJson)
	changed := strings.Replace(string(hostAxiomsJson), "columns?: number;", "columns: number | undefined;", 1)
	if changed == string(hostAxiomsJson) || key([]byte(changed)) == before {
		test.Fatal("editing a host row did not change the lowering product key")
	}
}

func TestHostAxiomsMissingFileNamesTheRow(test *testing.T) {
	test.Parallel()
	validateNodeHostAxiomsError := validateNodeHostAxioms(fstest.MapFS{})
	if validateNodeHostAxiomsError == nil ||
		!strings.Contains(validateNodeHostAxiomsError.Error(), "row 1 (tty.d.ts:51)") {
		test.Fatalf("missing host axiom file: %v", validateNodeHostAxiomsError)
	}
}
