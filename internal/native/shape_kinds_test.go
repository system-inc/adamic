package native

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestShapeKindsDistinguishScalarLayouts(t *testing.T) {
	e := &emitter{}
	number := e.shapeOf([]string{"value"}, []ir.Type{ir.Number})
	boolean := e.shapeOf([]string{"value"}, []ir.Type{ir.Boolean})
	reference := e.shapeOf([]string{"value"}, []ir.Type{ir.String})
	if number == boolean || number == reference || boolean == reference {
		t.Fatal("different field kinds share a shape")
	}
	if e.shapeOf([]string{"value"}, []ir.Type{ir.Number}) != number {
		t.Fatal("identical layouts do not share a shape")
	}
	source := strings.Join(e.declarations, "\n")
	for _, kind := range []string{"adamic_field_number", "adamic_field_boolean", "adamic_field_reference"} {
		if !strings.Contains(source, kind) {
			t.Fatalf("missing %s", kind)
		}
	}
}

// Mutate a runtime stat field, keeping its existing ownership bitmap unchanged.
// The counted allocator must reject the inconsistent shape before any slots are used.
func TestCountedShapeKindsCatchRuntimeMutant(t *testing.T) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid", true: "mutant"}[mutant], func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range files {
				contents := string(file.contents)
				if mutant && file.name == "directory.c" {
					before := "status_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_number, adamic_field_boolean}"
					after := "status_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_boolean}"
					if strings.Count(contents, before) != 1 {
						t.Fatal("stat mutant target missing")
					}
					contents = strings.Replace(contents, before, after, 1)
				}
				if err := os.WriteFile(filepath.Join(directory, file.name), []byte(contents), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			options := Options{Count: true}
			library, err := RuntimeLibrary(directory, options)
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(directory, "main.c")
			code := `#include "adamic.h"
_Static_assert(sizeof(adamic_field_kind) == 1, "field kinds must occupy one byte");
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 static adamic_string path = ADAMIC_STRING(".");
 adamic_object *result = adamic_file_status(&path);
 if (result->shape->kinds[2] != adamic_field_number || result->shape->kinds[3] != adamic_field_boolean) return 1;
 adamic_release(result);
 return 0;
}
`
			if err := os.WriteFile(source, []byte(code), 0o644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "probe")
			arguments := append(LinkFlags(options), "-I", filepath.Dir(library), "-o", binary, source)
			arguments = append(arguments, RuntimeLinkFlags(library)...)
			arguments = append(arguments, "-lm")
			if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
				t.Fatalf("compile: %v\n%s", err, output)
			}
			output, err := exec.Command(binary).CombinedOutput()
			if mutant {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatalf("mutant did not terminate: %v\n%s", err, output)
				}
				status, ok := exit.Sys().(syscall.WaitStatus)
				if !ok || !status.Signaled() || status.Signal() != syscall.SIGABRT || !strings.Contains(string(output), "inconsistent shape kind for field size") {
					t.Fatalf("mutant escaped shape check: %v\n%s", err, output)
				}
			} else if err != nil {
				t.Fatalf("valid shape: %v\n%s", err, output)
			}
		})
	}
}
