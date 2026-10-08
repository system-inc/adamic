package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Not parallel: this test explicitly compares the object cache with its bypass on the same input.
func TestSplitTSGoAgrees(t *testing.T) {
	archive := os.Getenv("ADAMIC_CLANG_TSGO_ARCHIVE")
	if archive == "" {
		t.Skip("set ADAMIC_CLANG_TSGO_ARCHIVE to a built checker archive")
	}
	program, err := load.Load([]string{"../../stage1/cohere/typeaware/main.ts"})
	if err != nil {
		t.Fatal(err)
	}
	program.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source, err := TSGoC(lowered)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	input := filepath.Join(directory, "sample.a")
	config := filepath.Join(directory, "tsconfig.json")
	manifest := filepath.Join(directory, "manifest")
	for name, contents := range map[string]string{
		input:    "const value = 1; const negative = -value;\n",
		config:   `{"compilerOptions":{"strict":true,"target":"esnext"},"files":["sample.a"]}`,
		manifest: input + "\n",
	} {
		if err := os.WriteFile(name, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	whole := filepath.Join(directory, "whole")
	options := Options{Sanitize: true, Jobs: 5}
	if err := BuildTSGo(source, whole, archive, options); err != nil {
		t.Fatal(err)
	}
	want, err := exec.Command(whole, config, manifest).CombinedOutput()
	if err != nil {
		t.Fatalf("whole checker: %v\n%s", err, want)
	}
	for _, uncached := range []string{"0", "1"} {
		t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		split := filepath.Join(directory, "split")
		if err := BuildSplitTSGo(source, split, archive, options); err != nil {
			t.Fatal(err)
		}
		got, err := exec.Command(split, config, manifest).CombinedOutput()
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("checker uncached=%s: %v\nwhole=%s\nsplit=%s", uncached, err, want, got)
		}
		t.Logf("checker uncached=%s: %d identical bytes: %s", uncached, len(got), got)
	}
}

// Compile-only coverage of the split checker path: no checker archive or runtime
// link is needed to detect disagreements between generated units and runtime types.
func TestSplitTSGoSmoke(t *testing.T) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang is not installed")
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "smoke.ts")
	// arguments.length selects the struct method-entry convention; the class
	// ensures that convention is exercised by an emitted method table.
	const input = `class Counter {
 value = 1;
 step(): number { this.value += 1; return this.value; }
}
function count(): number { return arguments.length; }
console.log(new Counter().step().toString() + count().toString());
`
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	loaded.EnableTSGo()
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	source, err := TSGoC(program)
	if err != nil {
		t.Fatal(err)
	}
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "const adamic_method_entry") {
		t.Fatal("smoke program did not emit a method table")
	}
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	common := []runtimeFile{{"units.h", []byte(header)}, {"tsgo.h", bridge.Header}}
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") {
			common = append(common, file)
		}
	}
	// Match splitTSGoRuntime's checker flags and buildUnitsWithLibrary's per-unit
	// path. Always bypass the object cache so the fast gate actually compiles.
	flags := append(sourceFlags(source, Options{}), "-DADAMIC_TSGO")
	for _, unit := range units {
		snapshot := append(append([]runtimeFile{}, common...), runtimeFile{unit.name, []byte(unit.source)})
		if _, err := compileUnit(unit, snapshot, flags, compiler, string(version), filepath.Join(directory, "cache"), directory, true); err != nil {
			t.Error(err)
		}
	}
	t.Logf("attempted compilation of %d split checker units", len(units))
}
