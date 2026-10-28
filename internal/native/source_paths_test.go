package native

import (
	"bytes"
	"context"
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func debugPathFixture(t *testing.T, root string) string {
	t.Helper()
	entry := filepath.Join(root, "entry", "fixture.a")
	for name, contents := range map[string]string{
		entry:                                 "import { imported } from '../lib/value.a';\nfunction marked(n: number): number {\n if (n > 0) return marked(n - 1) + n;\n return imported(n);\n}\nconsole.log(marked(3).toString());\n",
		filepath.Join(root, "lib", "value.a"): "export function imported(n: number): number {\n if (n > 0) return imported(n - 1) + n;\n return n + 2;\n}\n",
	} {
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return entry
}

// Check complete paths, not basenames: a correct line number is insufficient if
// the debugger cannot open the source. Go's LineReader already resolves DWARF
// file-table directory indexes; only a remaining relative name needs comp_dir.
func debugSourcePaths(t *testing.T, binary, entry string) bool {
	t.Helper()
	var data *dwarf.Data
	var err error
	if object, openErr := elf.Open(binary); openErr == nil {
		defer object.Close()
		data, err = object.DWARF()
	} else {
		object, openErr := macho.Open(binary)
		if openErr != nil {
			t.Fatal(openErr)
		}
		defer object.Close()
		data, err = object.DWARF()
	}
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string]bool{entry: false, filepath.Clean(filepath.Join(filepath.Dir(entry), "../lib/value.a")): false}
	seen := map[string]bool{}
	ok := true
	generatedRows, headerRows := 0, 0
	reader := data.Reader()
	for {
		unit, err := reader.Next()
		if err != nil {
			t.Fatal(err)
		}
		if unit == nil {
			break
		}
		if unit.Tag != dwarf.TagCompileUnit {
			continue
		}
		directory, _ := unit.Val(dwarf.AttrCompDir).(string)
		lines, err := data.LineReader(unit)
		if err != nil {
			t.Fatal(err)
		}
		if lines == nil {
			continue
		}
		var row dwarf.LineEntry
		for {
			err := lines.Next(&row)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if row.EndSequence || row.File == nil {
				continue
			}
			path := row.File.Name
			// DWARF 5 file zero names the CU's primary source. Go's reader
			// joins dir zero even when that filename is already absolute.
			// Use the CU name for it so this check follows debugger semantics.
			if files := lines.Files(); len(files) > 0 && row.File == files[0] {
				if name, _ := unit.Val(dwarf.AttrName).(string); filepath.IsAbs(name) {
					path = name
				}
			}
			if !filepath.IsAbs(path) {
				path = filepath.Join(directory, path)
			}
			path = filepath.Clean(path)
			if strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".h") {
				if strings.HasPrefix(path, filepath.Dir(entry)+string(filepath.Separator)) {
					t.Errorf("generated C/header row resolves beside entry, possibly to a decoy: %s", path)
					ok = false
				}
				if strings.HasSuffix(path, ".h") {
					headerRows++
				}
				if filepath.Base(path) == "main.c" {
					generatedRows++
				}
				continue
			}
			if !strings.HasSuffix(path, ".a") {
				continue
			}
			if seen[path] {
				continue
			}
			seen[path] = true
			_, statErr := os.Stat(path)
			t.Logf("comp_dir=%q row=%q resolved=%q os.Stat=%v", directory, row.File.Name, path, statErr)
			if statErr != nil {
				ok = false
			}
			if _, exists := wanted[path]; exists && statErr == nil {
				wanted[path] = true
			}
		}
	}
	for path, found := range wanted {
		if !found {
			t.Logf("real source not reached: %s", path)
			ok = false
		}
	}
	if generatedRows == 0 {
		t.Error("no generated main.c rows exercised")
		ok = false
	}
	if headerRows == 0 {
		t.Error("no runtime header rows exercised")
		ok = false
	}
	t.Logf("checked %d generated main.c rows and %d header rows against entry-directory decoys", generatedRows, headerRows)
	return ok
}

func TestDebugSourcePaths(t *testing.T)      { t.Parallel(); checkDebugSourcePaths(t, false) }
func TestDebugSourcePathsSplit(t *testing.T) { t.Parallel(); checkDebugSourcePaths(t, true) }
func checkDebugSourcePaths(t *testing.T, split bool) {
	t.Helper()
	entry := debugPathFixture(t, t.TempDir())
	for _, relative := range []string{"main.c", "runtime/adamic.h"} {
		path := filepath.Join(filepath.Dir(entry), relative)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("decoy, never generated code\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(t.TempDir(), "program")
	debugCLI(t, entry, binary, "ADAMIC_NATIVE_SPLIT="+map[bool]string{false: "0", true: "1"}[split])
	if !debugSourcePaths(t, binary, entry) {
		t.Fatal("debugger paths do not reach entry and imported source")
	}
}

func debugCLI(t *testing.T, entry, binary string, environment ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(debugCommandProduct(t, false), "command.test"), "-test.run=^TestBuildDebugDriver$", "--", "build", entry, "-o", binary, "--debug")
	command.Dir = buildcache.Absolute("<repository>/cmd/adamic")
	command.Env = debugCacheEnvironment(t, append(append(os.Environ(), "ADAMIC_BUILD_DEBUG_DRIVER=1"), environment...))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build --debug: %v: %s", err, output)
	}
}

func TestDebugRuntimeArchiveShared(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile(filepath.Join(debugSharingProduct(t, false), "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	var outcomes, keys []string
	for _, line := range strings.Split(string(contents), "\n") {
		match := buildLogLine.FindStringSubmatch(line)
		if match != nil && match[1] == "cArchive" {
			outcomes = append(outcomes, match[4])
			keys = append(keys, match[3])
			t.Log(line)
		}
	}
	if !reflect.DeepEqual(outcomes, []string{"miss", "hit"}) || len(keys) != 2 || keys[0] != keys[1] {
		t.Fatalf("runtime archives: outcomes=%v keys=%v, want one miss then hit under the same key", outcomes, keys)
	}
}

func TestDebugSanitizedUnitInputs(t *testing.T) {
	t.Parallel()
	options := Options{Debug: true, Sanitize: true, Split: true, DebugDirectory: t.TempDir()}
	source := "#include \"adamic.h\"\nstatic int adamic_value = 1;\nstatic int *adamic_probe(void) { return &adamic_value; }\nint main(int argc, char **argv) { (void)argv; return *adamic_probe() == argc ? 0 : 1; }\n"
	binary := filepath.Join(t.TempDir(), "program")
	if err := buildUnitsWithLibrary(source, binary, options, debugArchiveProduct(t, options, false), ""); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	// ASan's global metadata preserves the input spelling even when debug prefix
	// maps remove the build directory elsewhere in the binary.
	if bytes.Contains(contents, []byte("adamic-units-")) {
		t.Fatal("sanitized debug inputs were absolute: ASan recorded a temporary unit directory")
	}
	if !bytes.Contains(contents, []byte(options.DebugDirectory)) {
		t.Fatal("sanitized debug compile lost the entry comp_dir")
	}
	t.Log("sanitized debug split binary keeps entry comp_dir and relative generated inputs")
}
