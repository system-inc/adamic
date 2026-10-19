package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// productCompilerArguments names every source and include directory independently of
// where the builder placed it. The working directory also reaches DW_AT_comp_dir.
// Keep this on compiler calls only: ar accepts no compiler flags.
func productCompilerArguments(t *testing.T, working string, arguments []string) []string {
	t.Helper()
	mapped := append([]string{}, arguments...)
	mapped = append(mapped, prefixMaps(working, "/adamic-test-build")...)
	include := 0
	source := 0
	for index, argument := range arguments {
		directory, name := "", ""
		if argument == "-I" && index+1 < len(arguments) {
			directory = arguments[index+1]
			name = fmt.Sprintf("/adamic-test-include-%d", include)
			include++
		} else if len(argument) > 2 && argument[:2] == "-I" {
			directory = argument[2:]
			name = fmt.Sprintf("/adamic-test-include-%d", include)
			include++
		} else if filepath.Ext(argument) == ".c" {
			directory = filepath.Dir(argument)
			name = fmt.Sprintf("/adamic-test-source-%d", source)
			source++
		}
		if directory != "" {
			absolute, err := filepath.Abs(directory)
			if err != nil {
				t.Fatal(err)
			}
			mapped = append(mapped, prefixMaps(absolute, name)...)
		}
	}
	return mapped
}

// Sanitizer metadata keeps source and header filenames even with prefix maps.
// Compile through fixed relative aliases too, as the shipping builder does, in
// a private directory so neither the source tree nor cached headers are changed.
func productCompilerCommand(t *testing.T, ctx context.Context, compiler string, arguments ...string) *exec.Cmd {
	t.Helper()
	directory := t.TempDir()
	mapped := productCompilerArguments(t, directory, arguments)
	include, source := 0, 0
	alias := func(path, name string) string {
		t.Helper()
		absolute, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(absolute, filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
		return name
	}
	for index, argument := range arguments {
		if argument == "-I" && index+1 < len(arguments) {
			mapped[index+1] = alias(arguments[index+1], fmt.Sprintf("include-%d", include))
			include++
		} else if len(argument) > 2 && argument[:2] == "-I" {
			mapped[index] = "-I" + alias(argument[2:], fmt.Sprintf("include-%d", include))
			include++
		} else if filepath.Ext(argument) == ".c" {
			absolute, err := filepath.Abs(argument)
			if err != nil {
				t.Fatal(err)
			}
			relative := filepath.Join(alias(filepath.Dir(absolute), fmt.Sprintf("source-%d", source)), filepath.Base(absolute))
			mapped[index] = relative
			source++
		}
	}
	command := exec.CommandContext(ctx, compiler, mapped...)
	command.Dir = directory
	return command
}

// A tiny sanitized product exercises the same edited-unit compile, archive update
// and harness link as remoteScanProduct, without depending on a warm runtime cache.
func TestProductCompilerBytesAreIndependentOfDirectory(t *testing.T) {
	t.Parallel()
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	working, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var products []map[string][]byte
	var directoryLengths []int
	for _, name := range []string{"short", "a-much-longer-build-directory"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			directoryLengths = append(directoryLengths, len(root))
			include := filepath.Join(root, "cache")
			extra := filepath.Join(root, "extra")
			scratch := filepath.Join(root, "scratch")
			for _, directory := range []string{include, extra, scratch} {
				if err := os.MkdirAll(directory, 0755); err != nil {
					t.Fatal(err)
				}
			}
			write := func(name, text string) {
				t.Helper()
				if err := os.WriteFile(name, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			write(filepath.Join(include, "adamic.h"), "static const char *header_file(void) { return __FILE__; }\nstatic unsigned header_shift(unsigned n) { return 1u << n; }\nconst char *heap_file(void);\n")
			write(filepath.Join(extra, "extra.h"), "static const char *extra_file(void) { return __FILE__; }\nstatic int extra_divide(int n) { return 42 / n; }\n")
			write(filepath.Join(scratch, "heap.c"), "#include \"adamic.h\"\nconst char *heap_file(void) { volatile unsigned n = 1; return header_shift(n) ? __FILE__ : header_file(); }\n")
			flags := append(Flags(Options{Sanitize: true}), "-I", include, "-I"+extra)
			object := filepath.Join(scratch, "heap.o")
			staticsCompile(t, compiler, append(append([]string{}, flags...), "-c", filepath.Join(scratch, "heap.c"), "-o", object)...)
			library := filepath.Join(scratch, "runtime.a")
			staticsCompile(t, archiverName(compiler), "rcs", library, object)
			// Exercise the replacement operation used by the remote-scan mutants too.
			staticsCompile(t, archiverName(compiler), "r", library, object)
			binary := buildHarnessWithLibrary(t, "#include \"adamic.h\"\n#include \"extra.h\"\n#include <stdio.h>\nint main(int count, char **arguments) { (void)arguments; puts(__FILE__); puts(header_file()); puts(extra_file()); puts(heap_file()); printf(\"%u %d\\n\", header_shift((unsigned)count), extra_divide(count)); return 0; }\n", flags, compiler, library)
			product := map[string][]byte{}
			for _, file := range []struct{ name, path string }{{"heap.o", object}, {"runtime.a", library}, {"harness", binary}, {"main.c", filepath.Join(filepath.Dir(binary), "main.c")}} {
				data, err := os.ReadFile(file.path)
				if err != nil {
					t.Fatal(err)
				}
				for _, directory := range []string{root, filepath.Dir(binary), working} {
					if bytes.Contains(data, []byte(directory)) {
						t.Errorf("%s records an unmapped build path", file.name)
					}
				}
				product[file.name] = data
			}
			products = append(products, product)
		})
	}
	if len(products) != 2 {
		t.Fatal("both product builds must finish")
	}
	if directoryLengths[0] == directoryLengths[1] {
		t.Fatal("build directories must have different lengths")
	}
	for _, name := range []string{"heap.o", "runtime.a", "harness", "main.c"} {
		a, b := products[0][name], products[1][name]
		if !bytes.Equal(a, b) {
			offset := 0
			for offset < min(len(a), len(b)) && a[offset] == b[offset] {
				offset++
			}
			t.Errorf("%s differs at byte %d (sizes %d and %d)", name, offset+1, len(a), len(b))
		}
	}
}
