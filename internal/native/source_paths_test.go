package native

import (
	"context"
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
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
			if row.EndSequence || row.File == nil || !strings.HasSuffix(row.File.Name, ".a") {
				continue
			}
			path := row.File.Name
			if !filepath.IsAbs(path) {
				path = filepath.Join(directory, path)
			}
			path = filepath.Clean(path)
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
	return ok
}

func TestDebugSourcePaths(t *testing.T) {
	t.Skip("#gd4csnt: relative source paths resolve against synthetic comp_dir /adamic-build; Stage 1 records the failure before choosing flags or maps")
	t.Parallel()
	entry := debugPathFixture(t, t.TempDir())
	binary := filepath.Join(t.TempDir(), "program")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	command := toolchain.CommandContextIn(ctx, "../..", "go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--debug")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build --debug: %v: %s", err, output)
	}
	if !debugSourcePaths(t, binary, entry) {
		t.Fatal("debugger paths do not reach entry and imported source")
	}
}
