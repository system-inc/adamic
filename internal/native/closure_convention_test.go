package native

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The ruling explicitly requires this mutant to die in the C compiler.
func TestClosureConventionDropCount(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	if err := Build(source, filepath.Join(t.TempDir(), "valid"), Options{}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(source, "\n")
	mutated := false
	for i, line := range lines {
		if !strings.Contains(line, "= adamic_closure_call(") {
			continue
		}
		last := strings.LastIndex(line, ", ")
		close := strings.LastIndex(line, ");")
		if last < 0 || close < last {
			t.Fatal("counted call has unexpected form")
		}
		lines[i] = line[:last] + line[close:]
		mutated = true
		break
	}
	if !mutated {
		t.Fatal("no counted call to mutate")
	}
	err = Build(strings.Join(lines, "\n"), filepath.Join(t.TempDir(), "mutant"), Options{})
	if err == nil || !strings.Contains(err.Error(), "too few arguments to function call") || !strings.Contains(err.Error(), "expected 3, have 2") {
		t.Fatalf("drop-count mutant was not rejected for typed arity: %v", err)
	}
	t.Log("drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}

// The runtime callback site must obey the same compiler-enforced arity.
func TestClosureConventionRuntimeDropCount(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "closure_convention_regexp_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join("runtime", "regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	original := "adamic_closure_call(callback, packed, argument_count)"
	if strings.Count(string(runtime), original) != 1 {
		t.Fatal("runtime mutation site moved")
	}
	changed := strings.Replace(string(runtime), original, "adamic_closure_call(callback, packed)", 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	source := changed + "\n" + strings.ReplaceAll(C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	err = Build(source, filepath.Join(t.TempDir(), "mutant"), Options{})
	if err == nil || !strings.Contains(err.Error(), "expected 3, have 2") {
		t.Fatalf("runtime drop-count mutant escaped the typed convention: %v", err)
	}
	t.Log("runtime callback drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}

func TestClosureConventionRuntimeFeaturesIgnoreLiterals(t *testing.T) {
	program := &ir.Program{Strings: []string{"adamic_regex_replace_callback", "adamic_fs_file_host_cwd"}}
	library, err := RuntimeLibraryForSource("", C(program), Options{})
	if err != nil {
		t.Fatal(err)
	}
	header, err := os.ReadFile(filepath.Join(filepath.Dir(library), "adamic.h"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"#define ADAMIC_REGEXP_REPLACE_CALLBACK 1", "#define ADAMIC_NODE_HOST 1"} {
		if strings.Contains(string(header), marker) {
			t.Fatalf("a source literal selected unused runtime support: %s", marker)
		}
	}
}

// Definition order, call order and explicit signature-erasing casts each fail
// independently. A constructor check alone misses a never-stored definition.
func TestClosureConventionWrongOrder(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	old := "adamic_value *arguments, size_t argument_count"
	if !strings.Contains(source, old) {
		t.Fatal("no counted definition")
	}
	swapped := strings.Replace(source, old, "size_t argument_count, adamic_value *arguments", 1)
	t.Run("definition", func(t *testing.T) {
		err := Build(swapped, filepath.Join(t.TempDir(), "mutant"), Options{})
		if err == nil || !strings.Contains(err.Error(), "conflicting types") {
			t.Fatalf("swapped definition escaped the header-derived declaration: %v", err)
		}
		t.Log(err)
	})
	t.Run("call", func(t *testing.T) {
		lines := strings.Split(source, "\n")
		changed := false
		for i, line := range lines {
			start := strings.Index(line, "= adamic_closure_call(")
			if start < 0 {
				continue
			}
			first := strings.Index(line[start:], ", ") + start
			last := strings.LastIndex(line, ", ")
			end := strings.LastIndex(line, ");")
			if first >= last || end < last {
				t.Fatal("mutation site changed")
			}
			lines[i] = line[:first+2] + line[last+2:end] + ", " + line[first+2:last] + line[end:]
			changed = true
			break
		}
		if !changed {
			t.Fatal("no counted call")
		}
		err := Build(strings.Join(lines, "\n"), filepath.Join(t.TempDir(), "mutant"), Options{})
		if err == nil || (!strings.Contains(err.Error(), "int-conversion") && !strings.Contains(err.Error(), "incompatible")) {
			t.Fatalf("swapped call escaped typed parameters: %v", err)
		}
		t.Log(err)
	})
	for _, probe := range []struct{ name, code, diagnostic string }{
		{"function-pointer-cast", "static adamic_value wrong(adamic_closure *self, size_t count, adamic_value *arguments) { return (adamic_value){.number = (double)count}; }\nstatic adamic_counted_code erased = (adamic_counted_code)wrong;\n", "cast-function-type-strict"},
		{"void-pointer-cast", "static adamic_value wrong(adamic_closure *self, size_t count, adamic_value *arguments) { return (adamic_value){.number = (double)count}; }\nstatic adamic_closure *erased(void) { return adamic_counted_closure_new((void *)wrong, 0); }\n", "generic association"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			err := Build(strings.Replace(source, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+probe.code, 1), filepath.Join(t.TempDir(), "mutant"), Options{})
			if err == nil || !strings.Contains(err.Error(), probe.diagnostic) {
				t.Fatalf("signature-erasing cast escaped compiler enforcement: %v", err)
			}
			t.Log(err)
		})
	}
}
