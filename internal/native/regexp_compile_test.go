package native

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/unicodeproperties"
)

// Enumeration comes from the Go AST; expected entries come from the public Go
// provider, independently of the Python generator's source-layout conversion.
func runtimePropertyAliases(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "../unicodeproperties/tables.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	prefixes := map[string][]string{
		"binaryNames":          {""},
		"generalCategoryNames": {"", "gc=", "General_Category="},
		"scriptNames":          {"sc=", "Script="},
		"scriptExtensionNames": {"scx=", "Script_Extensions="},
		"stringProperties":     {""},
	}
	var aliases []string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, specification := range general.Specs {
			value, ok := specification.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 {
				continue
			}
			selected, ok := prefixes[value.Names[0].Name]
			if !ok {
				continue
			}
			table := value.Values[0].(*ast.CompositeLit)
			for _, element := range table.Elts {
				literal := element.(*ast.KeyValueExpr).Key.(*ast.BasicLit)
				name, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				for _, prefix := range selected {
					aliases = append(aliases, prefix+name)
				}
			}
		}
	}
	sort.Strings(aliases)
	if len(aliases) != 1722 {
		t.Fatalf("alias enumeration: %d, want 1722", len(aliases))
	}
	return aliases
}

func TestRegExpRuntimePropertiesGenerated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tables.h")
	command := exec.Command("python3", "../regexp/testdata/generate-runtime-properties.py", "--output", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generate: %v\n%s", err, output)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile("runtime/regexp_compile_tables.h")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, committed) {
		t.Fatal("runtime property tables are stale; run generate-runtime-properties.py")
	}
}

func runtimePropertyHarness(t *testing.T) string {
	t.Helper()
	aliases := runtimePropertyAliases(t)
	var source, rows strings.Builder
	source.WriteString("#include \"regexp_compile.h\"\n")
	names := map[string]string{}
	rangeEntries, stringEntries := 0, 0
	for _, alias := range aliases {
		property, ok := unicodeproperties.Lookup(alias, true)
		if !ok {
			t.Fatalf("Go provider rejected %q", alias)
		}
		key := property.Name + "=" + property.Value
		name, exists := names[key]
		if !exists {
			name = fmt.Sprintf("expected_%d", len(names))
			names[key] = name
			ranges, texts := "NULL", "NULL"
			rangeCount := 0
			if property.Set != nil {
				rangeCount = len(property.Set.Ranges)
			}
			if rangeCount > 0 {
				ranges = name + "_ranges"
				fmt.Fprintf(&source, "static const adamic_regex_compile_range %s[] = {", ranges)
				for _, interval := range property.Set.Ranges {
					fmt.Fprintf(&source, "{%d,%d},", interval.Start, interval.End)
				}
				source.WriteString("};\n")
				rangeEntries += rangeCount
			}
			if len(property.Sequences) > 0 {
				texts = name + "_texts"
				for index, sequence := range property.Sequences {
					fmt.Fprintf(&source, "static const uint32_t %s_text_%d[] = {", name, index)
					for _, point := range sequence {
						fmt.Fprintf(&source, "%d,", point)
					}
					source.WriteString("};\n")
				}
				fmt.Fprintf(&source, "static const adamic_regex_compile_text %s[] = {", texts)
				for index, sequence := range property.Sequences {
					fmt.Fprintf(&source, "{%s_text_%d,%d},", name, index, len([]rune(sequence)))
				}
				source.WriteString("};\n")
				stringEntries += len(property.Sequences)
			}
			fmt.Fprintf(&source, "static const adamic_regex_compile_property %s = {%s,%d,%s,%d,%t};\n", name, ranges, rangeCount, texts, len(property.Sequences), property.Kind == unicodeproperties.KindStrings)
		}
		fmt.Fprintf(&rows, "{%s,%d,&%s},\n", cString(alias), len(alias), name)
	}
	// Invalid spellings are looked up through Go independently, including
	// embedded NUL and a lone surrogate in the runtime's WTF-8 encoding.
	invalid := []string{"", "=", "ascii", "ASCII=Yes", "SC=Latin", "Script=latin", "Script=Latn\x00x", " RGI_Emoji", "\xed\xa0\x80"}
	for _, alias := range aliases {
		invalid = append(invalid, alias+"!", " "+alias, alias+"\x00")
	}
	for _, alias := range invalid {
		if _, ok := unicodeproperties.Lookup(alias, true); ok {
			t.Fatalf("negative property case unexpectedly valid: %q", alias)
		}
		fmt.Fprintf(&rows, "{%s,%d,NULL},\n", cString(alias), len(alias))
	}
	// Every alias's two Unicode modes and counted-string near misses are checked.
	source.WriteString("typedef struct {const char *alias; size_t length; const adamic_regex_compile_property *expected;} probe;\nstatic const probe probes[] = {\n")
	source.WriteString(rows.String())
	source.WriteString("};\n")
	source.WriteString(`
static bool same(const adamic_regex_compile_property *a, const adamic_regex_compile_property *b) {
    if (a == NULL || b == NULL) return a == b;
    if (a->range_count != b->range_count || a->string_count != b->string_count || a->string_property != b->string_property) return false;
    for (size_t i = 0; i < a->range_count; i++) {
        if (a->ranges[i].first != b->ranges[i].first || a->ranges[i].last != b->ranges[i].last) return false;
    }
    for (size_t i = 0; i < a->string_count; i++) {
        if (a->strings[i].count != b->strings[i].count) return false;
        for (size_t j = 0; j < a->strings[i].count; j++) {
            if (a->strings[i].points[j] != b->strings[i].points[j]) return false;
        }
    }
    return true;
}
static int check(void) {
    for (size_t i = 0; i < sizeof(probes)/sizeof(probes[0]); i++) {
        const probe *p = &probes[i];
        if (!same(adamic_regex_compile_lookup_property((const unsigned char *)p->alias, p->length, true), p->expected)) return (int)i + 1;
        if (!same(adamic_regex_compile_lookup_property((const unsigned char *)p->alias, p->length, false), p->expected != NULL && p->expected->string_property ? NULL : p->expected)) return (int)i + 1;
        // Include the trailing NUL in the length. strcmp would incorrectly accept it.
        if (adamic_regex_compile_lookup_property((const unsigned char *)p->alias, p->length + 1, true) != NULL) return (int)i + 1;
    }
    return 0;
}
#if ADAMIC_TARGET_WASI
void _start(void) { if (check() != 0) __builtin_trap(); }
#else
#include <stdio.h>
int main(void) { int result = check(); if (result != 0) fprintf(stderr, "property identity failed at alias %d\n", result); return result != 0; }
#endif
`)
	t.Logf("entry identity: %d aliases, %d unique properties, %d ranges, %d strings", len(aliases), len(names), rangeEntries, stringEntries)
	return source.String()
}

func TestRegExpRuntimePropertyIdentity(t *testing.T) {
	source := runtimePropertyHarness(t)
	directory := t.TempDir()
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	runtimeDirectory, err := filepath.Abs("runtime")
	if err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(directory, "check")
	arguments := append(Flags(Options{Sanitize: true}), "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-I", runtimeDirectory, main, filepath.Join(runtimeDirectory, "regexp_compile_properties.c"), "-o", executable)
	command := exec.Command("clang", arguments...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, output)
	}
	command = exec.Command(executable)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("C versus Go property identity: %v\n%s", err, output)
	}
}

func TestRegExpRuntimePropertiesWASI(t *testing.T) {
	// This table unit uses no allocation, libc, atomics or runtime host services.
	// WASI runs the same entry comparisons with 32-bit pointers. The Linux
	// identity gate above is the sanitizer gate of record.
	source := runtimePropertyHarness(t)
	directory := t.TempDir()
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	runtimeDirectory, err := filepath.Abs("runtime")
	if err != nil {
		t.Fatal(err)
	}
	wasm := filepath.Join(directory, "check.wasm")
	arguments := []string{"--target=wasm32-wasi", "-DADAMIC_TARGET_WASI=1", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O1", "-fno-builtin", "-nostdlib", "-Wl,--entry=_start", "-Wl,--export=_start", "-Wl,--export-memory", "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-I", runtimeDirectory, main, filepath.Join(runtimeDirectory, "regexp_compile_properties.c"), "-o", wasm}
	command := exec.Command("clang", arguments...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("wasm32 compile: %v\n%s", err, output)
	}
	command = exec.Command("node", "--no-warnings", "-e", `const fs = require('fs'); const {WASI} = require('node:wasi'); const wasi = new WASI({version:'preview1', args:[], env:{}}); WebAssembly.instantiate(fs.readFileSync(process.argv[1]), {wasi_snapshot_preview1:wasi.wasiImport}).then(({instance}) => wasi.start(instance));`, wasm)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("WASI identity: %v\n%s", err, output)
	}
}

func TestRegExpRuntimeCompilerNotLinked(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "empty")
	if err := Build("#include \"adamic.h\"\nint main(void) { return 0; }\n", binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("sanitized empty program: %v\n%s", err, output)
	}
	output, err := exec.Command("nm", "-a", binary).CombinedOutput()
	if err != nil {
		t.Fatalf("symbol list: %v\n%s", err, output)
	}
	if bytes.Contains(output, []byte("regex_compile_")) {
		t.Fatal("a program without dynamic RegExp linked compiler symbols or tables")
	}
	if bytes.Contains(output, []byte("adamic_regex_new_owned")) {
		t.Fatal("compiler ownership linked into a program without dynamic RegExp")
	}
	t.Log("nm: no compiler or ownership symbols in a program without dynamic RegExp")
}
