package native

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

var programCoreOptions = []Options{
	{ProgramRegion: true, Sanitize: true, Count: true},
	{ProgramRegion: true, Sanitize: true, slabs: true, Count: true},
	{ProgramRegion: true, Count: true},
}

type programCoreMutation struct{ name, file, before, after string }

var programCoreMutations = []programCoreMutation{
	{"aliased-adoption-allowed", "program_region.c", "heap->references != 1", "false"},
	{
		"sharing-allowed",
		"share.c",
		`adamic_panic("Program region members cannot cross into parallel work", sizeof "Program region members cannot cross into parallel work" - 1);`,
		`(void)heap;`,
	},
	{
		"cached-adoption-allowed",
		"program_region.c",
		`adamic_panic("compiler bug: Program adoption must precede canonical caching", sizeof "compiler bug: Program adoption must precede canonical caching" - 1);`,
		`(void)heap;`,
	},
	{"unmovable-adoption-allowed", "program_region.c", "if (unmovable != NULL) {", "if (false) {"},
	{"weak-adoption-allowed", "program_region.c", "adamic_weak_held(heap)", "false"},
	{"weak-cell-adoption-allowed", "program_region.c", "adamic_weak_held(&environment->cells[i])", "false"},
	{
		"member-release-frees",
		"heap.c",
		"if (heap == NULL || adamic_program_is(heap)) { return NULL; }",
		"if (heap == NULL) { return NULL; }\n if (adamic_program_is(heap)) { return heap; }",
	},
	{
		"weak-not-forgotten",
		"program_region.c",
		"  adamic_weak_forget(each);",
		"  /* omitted member weak invalidation */",
	},
	{"cell-weak-not-forgotten", "program_region.c", "adamic_weak_forget(&environment->cells[i]);", "(void)i;"},
	{
		"short-adoption",
		"program_region.c",
		"adamic_heap_program_storage(heap, size)",
		"adamic_heap_program_storage(heap, size - sizeof(size_t))",
	},
}

// ProgramCoreAllocatorCount is the inventory the split verification units cover.
func ProgramCoreAllocatorCount() int { return len(programCoreOptions) }

// ProgramCoreBinary is the direct C control or runtime mutant, built ahead by
// its allocator product. Each shard compiles the unchanged runtime once; mutants
// replace only their changed translation unit. No compiler membership is involved.
func ProgramCoreBinary(test *testing.T, index int, name string) string {
	test.Helper()
	return filepath.Join(programCoreProduct(test, index), name)
}

func programCoreProduct(test *testing.T, index int) string {
	test.Helper()
	options := programCoreOptions[index]
	files, flags, compiler, _, runtimeInputsError := runtimeInputs(
		"",
		options,
		[]string{"-DADAMIC_CANONICAL_CLOSURES=1"},
	)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
	}
	inputs := buildcache.Inputs{
		Name: fmt.Sprintf("program-region-core-%d", index),
		Files: []string{
			"internal/native/runtime",
			"internal/native/testdata/program_region",
			"internal/native/program_region_products_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/target.go",
			"internal/native/split_units_test.go",
			"internal/native/tsgo.go",
		},
		Flags: portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{
			buildcache.Tool(compiler, "--version"),
			buildcache.Tool(archiverName(compiler), "--version"),
		},
	}
	// Portable parts keep the product independent of a tool installation path.
	for _, mutation := range programCoreMutations {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf(
			"%s %s %x %x",
			mutation.name,
			mutation.file,
			sha256.Sum256([]byte(mutation.before)),
			sha256.Sum256([]byte(mutation.after)),
		))
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		scratch, temporaryDirectoryError := os.MkdirTemp("", "adamic-program-core-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(scratch)
		within := filepath.Join(scratch, "runtime")
		if createDirectoryError := os.MkdirAll(within, 0755); createDirectoryError != nil {
			return createDirectoryError
		}
		var sources []string
		for _, file := range files {
			if writeError := os.WriteFile(filepath.Join(within, file.name), file.contents, 0644); writeError != nil {
				return writeError
			}
			if strings.HasSuffix(file.name, ".c") {
				sources = append(sources, file.name)
			}
		}
		compile := func(root, source string) error {
			args := append(slices.Clone(flags), prefixMaps(root, "/adamic-runtime")...)
			args = append(args, "-c", source, "-o", strings.TrimSuffix(source, ".c")+".o")
			command := toolchain.CommandIn(root, compiler, args...)
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("compile %s: %w\n%s", source, combinedOutputError, output)
			}
			return nil
		}
		if eachInParallelError := eachInParallel(
			len(sources),
			func(index int) error { return compile(within, sources[index]) },
		); eachInParallelError != nil {
			return eachInParallelError
		}
		link := func(name, changed, object, source string) error {
			if writeError := os.WriteFile(filepath.Join(scratch, "main.c"), []byte(source), 0644); writeError != nil {
				return writeError
			}
			args := append(slices.Clone(flags), prefixMaps(scratch, "/adamic-build")...)
			args = append(args, "-I", "runtime", "-o", filepath.Join(directory, name), "main.c")
			for _, file := range sources {
				if file == changed {
					args = append(args, object)
				} else {
					args = append(args, filepath.Join("runtime", strings.TrimSuffix(file, ".c")+".o"))
				}
			}
			args = append(args, objectPrefixFlags()...)
			args = append(args, "-lm")
			if output, combinedOutputError := toolchain.CommandIn(
				scratch,
				compiler,
				args...,
			).CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("link %s: %w\n%s", name, combinedOutputError, output)
			}
			return nil
		}
		code, readError := os.ReadFile("testdata/program_region/core.c")
		if readError != nil {
			return readError
		}
		if linkError := link("control", "", "", string(code)); linkError != nil {
			return linkError
		}
		unsafe := strings.Replace(
			string(code),
			"root->slots[0].reference = adamic_graph_hold(root, child);",
			"root->slots[0].reference = child;",
			1,
		)
		if index == 0 {
			if unsafe == string(code) {
				return fmt.Errorf("missing fallback retain mutant")
			}
			if linkError := link("fallback-retain-dropped", "", "", unsafe); linkError != nil {
				return linkError
			}
		}
		if index != 0 {
			return nil
		}
		for _, mutation := range programCoreMutations {
			changed := filepath.Join(scratch, mutation.name)
			if createDirectoryError := os.MkdirAll(changed, 0755); createDirectoryError != nil {
				return createDirectoryError
			}
			for _, file := range files {
				if file.name != mutation.file && !strings.HasSuffix(file.name, ".h") {
					continue
				}
				data := file.contents
				if file.name == mutation.file {
					if strings.Count(string(data), mutation.before) != 1 {
						return fmt.Errorf("ambiguous mutant %s", mutation.name)
					}
					data = []byte(strings.Replace(string(data), mutation.before, mutation.after, 1))
				}
				if writeError := os.WriteFile(filepath.Join(changed, file.name), data, 0644); writeError != nil {
					return writeError
				}
			}
			if compileError := compile(changed, mutation.file); compileError != nil {
				return compileError
			}
			object := filepath.Join(mutation.name, strings.TrimSuffix(mutation.file, ".c")+".o")
			if linkError := link(mutation.name, mutation.file, object, string(code)); linkError != nil {
				return linkError
			}
		}
		return nil
	})
}

func TestProduct_ProgramRegionCoreASan(test *testing.T) { test.Parallel(); programCoreProduct(test, 0) }
func TestProduct_ProgramRegionCoreASanSlabs(test *testing.T) {
	test.Parallel()
	programCoreProduct(test, 1)
}
func TestProduct_ProgramRegionCoreCounted(test *testing.T) {
	test.Parallel()
	programCoreProduct(test, 2)
}
