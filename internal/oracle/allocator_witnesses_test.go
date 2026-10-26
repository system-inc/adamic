package oracle

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Each allocator witness is a solo Node/native/JavaScript, WASI and count case.
func init() {
	for _, name := range []string{"slab_sizes", "reuse_handover_fresh"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}

// The ordinary oracle checks malloc under ASan and slabs in release. This
// independent product also poisons slab slots, so class-size errors meet ASan.
func allocatorSlabProduct(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/slab_sizes.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	options := native.Options{Sanitize: true, Count: true}
	request, err := native.RuntimeRequestFor(options, native.Features(code)...)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "allocator-slab-sizes", Files: append(compilerinputs.Lowering(), "internal/oracle/allocator_witnesses_test.go", "internal/oracle/testdata/slab_sizes.a"), Flags: append(request.Portable, code, "ADAMIC_SLABS=1"), Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("node", "--version")}}
	return buildcache.Product(t, inputs, func(output string) error {
		truth := onNode(t, path)
		if truth.exitCode != 0 || len(truth.stderr) != 0 {
			return fmt.Errorf("Node failed: %+v", truth)
		}
		if err := os.WriteFile(filepath.Join(output, "truth"), truth.stdout, 0600); err != nil {
			return err
		}
		snapshot := t.TempDir()
		// Reads: internal/native/runtime/*
		files, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", file.Name()))
			if err != nil {
				return err
			}
			if file.Name() == "adamic.h" || filepath.Ext(file.Name()) == ".c" {
				data = append([]byte("#define ADAMIC_SLABS 1\n"), data...)
			}
			if err := os.WriteFile(filepath.Join(snapshot, file.Name()), data, 0600); err != nil {
				return err
			}
		}
		library, err := native.RuntimeLibraryForSource(snapshot, code, options)
		if err != nil {
			return err
		}
		if err := native.BuildWithRuntime(code, filepath.Join(output, "control"), options, library); err != nil {
			return err
		}
		// Only heap.c is recompiled for the undersize mutant. Keep every other
		// object and the exact shipping headers from the control archive.
		changed := t.TempDir()
		headers, err := os.ReadDir(filepath.Dir(library))
		if err != nil {
			return err
		}
		for _, header := range headers {
			if filepath.Ext(header.Name()) != ".h" {
				continue
			}
			data, err := os.ReadFile(filepath.Join(filepath.Dir(library), header.Name()))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(changed, header.Name()), data, 0600); err != nil {
				return err
			}
		}
		heap, err := os.ReadFile(filepath.Join(snapshot, "heap.c"))
		if err != nil {
			return err
		}
		old := "heap = take((size - 1) / GRANULE, &number);"
		if strings.Count(string(heap), old) != 1 {
			return fmt.Errorf("missing allocation-size seam")
		}
		text := strings.Replace(string(heap), old, "heap = take((size - 1) / GRANULE == 0 ? 0 : (size - 1) / GRANULE - 1, &number);", 1)
		if err := os.WriteFile(filepath.Join(changed, "heap.c"), []byte(text), 0600); err != nil {
			return err
		}
		object := filepath.Join(changed, "heap.o")
		// The scratch directory's name carries the build and a random suffix, so as native compiles its own runtime
		// (library.go) it's mapped out of everything the compile records of a path, under each spelling of it, and heap.c
		// is compiled from inside it by its bare name, which is what the sanitizers' source locations record: either one
		// left out made two builds of one key differ (#d32jbpx).
		flags := append(native.SourceFlags(code, options), "-ffile-prefix-map="+changed+"=/adamic-runtime")
		if real, err := filepath.EvalSymlinks(changed); err == nil && real != changed {
			flags = append(flags, "-ffile-prefix-map="+real+"=/adamic-runtime")
		}
		flags = append(flags, "-I", changed, "-c", "heap.c", "-o", object)
		if data, err := toolchain.CommandIn(changed, request.Compiler, flags...).CombinedOutput(); err != nil {
			return fmt.Errorf("mutant must compile: %w: %s", err, data)
		}
		archive, err := os.ReadFile(library)
		if err != nil {
			return err
		}
		mutant := filepath.Join(changed, "runtime.a")
		if err := os.WriteFile(mutant, archive, 0600); err != nil {
			return err
		}
		if data, err := toolchain.Command(request.Archiver, "r", mutant, object).CombinedOutput(); err != nil {
			return fmt.Errorf("mutant archive: %w: %s", err, data)
		}
		return native.BuildWithRuntime(code, filepath.Join(output, "undersize"), options, mutant)
	})
}
func TestProduct_AllocatorSlabSizes(t *testing.T) { t.Parallel(); allocatorSlabProduct(t) }
func TestAllocatorSlabSizesSanitized(t *testing.T) {
	t.Parallel()
	product := allocatorSlabProduct(t)
	truth, err := os.ReadFile(filepath.Join(product, "truth"))
	if err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, filepath.Join(product, "control"))
	if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
		t.Fatal(report)
	}
	if !bytes.Equal(result.stdout, truth) || len(leakcheck.CountsLine.ReplaceAll(result.stderr, nil)) != 0 {
		t.Fatalf("Node %q, slab result %+v", truth, result)
	}
}
func TestAllocatorSlabSizesUndersizeMutant(t *testing.T) {
	t.Parallel()
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, filepath.Join(allocatorSlabProduct(t), "undersize"))
	if result.exitCode == 0 || !bytes.Contains(result.stderr, []byte("ERROR: AddressSanitizer: use-after-poison")) {
		t.Fatalf("undersize mutant not caught by ASan: %+v", result)
	}
	t.Log("caught undersized slab storage by ASan use-after-poison")
}

// Giving a fresh temporary up before later arguments finish strands it when
// a later argument throws. Plant that compiler defect only through an overlay.
var allocatorFreshMutation = buildcache.Mutation{
	File:   "internal/native/reuse.go",
	Before: "if slices.Contains(e.owned, value) {\n\t\treturn value\n\t}",
	After:  "if index := slices.Index(e.owned, value); index >= 0 {\n\t\te.owned = slices.Delete(e.owned, index, index+1)\n\t\treturn value\n\t}",
}

func allocatorFreshCompiler(t *testing.T) string {
	t.Helper()
	return buildcache.GoBuildInMutated(t, "", "allocator-fresh-mutant", "./cmd/adamic", nil, []buildcache.Mutation{allocatorFreshMutation})
}
func TestProduct_AllocatorFreshHandoverCompiler(t *testing.T) {
	t.Parallel()
	allocatorFreshCompiler(t)
}
func TestAllocatorFreshHandoverMutant(t *testing.T) {
	t.Parallel()
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reuse_handover_fresh.a"))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "fresh-mutant")
	built := executeWith(t, nil, allocatorFreshCompiler(t), "build", fixture, "-o", binary, "--sanitize", "--count")
	if built.exitCode != 0 {
		t.Fatalf("mutant must compile: %+v", built)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if result.exitCode == 0 || !bytes.Contains(result.stderr, []byte("LeakSanitizer: detected memory leaks")) {
		t.Fatalf("early fresh handover not caught by leaks: %+v", result)
	}
	t.Log("caught compiler fresh handover before argument completion by LeakSanitizer")
}
