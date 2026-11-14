package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const assignByteStore = "  adamic_object_field_types(target)[adamic_slot_index(target, slot)] = adamic_object_field_type(source, index);"

func objectAssignBytesBinary(t *testing.T, mutant string) string {
	t.Helper()
	library, err := os.ReadFile("runtime/library_object.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(library)
	start := strings.Index(source, "void adamic_object_assign(")
	if start < 0 {
		t.Fatal("assign function missing")
	}
	end := strings.Index(source[start:], "\n}\n")
	if end < 0 {
		t.Fatal("assign function end missing")
	}
	// Keep the real store body, renamed so the archive supplies its other helpers.
	source = "#include \"adamic.h\"\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n" + source[start:start+end+3]
	if mutant != "" {
		if strings.Count(source, assignByteStore) != 1 {
			t.Fatal("assign byte-store seam changed")
		}
		replacement := "  /* mutant: leave the target's old representation byte */"
		if mutant == "raw" {
			replacement = "  adamic_object_field_types(target)[adamic_slot_index(target, slot)] = adamic_object_field_types(source)[index];"
		}
		source = strings.Replace(source, assignByteStore, replacement, 1)
	}
	source = strings.ReplaceAll(source, "adamic_object_assign(", "probe_adamic_object_assign(")
	provider, err := os.ReadFile("runtime/json_objects.c")
	if err != nil {
		t.Fatal(err)
	}
	source += "\n" + strings.ReplaceAll(string(provider), "adamic_json_stringify_objects(", "probe_adamic_json_stringify_objects(")
	harness, err := os.ReadFile("testdata/object-assign-bytes/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "assign-bytes")
	if err := Build(source+"\n"+string(harness), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func TestObjectAssignCopiesRepresentationBytes(t *testing.T) {
	t.Parallel()
	binary := objectAssignBytesBinary(t, "")
	for _, each := range []struct{ name, want string }{
		{"null", "12 {\"b\":null}\n"}, {"undefined", "13 {}\n"},
		{"sentinel", "3 {\"b\":\"7\"}\n"}, {"zero", "0\n"},
	} {
		t.Run(each.name, func(t *testing.T) {
			stdout, stderr, exit := jsonObjectsRuntimeRun(t, binary, each.name)
			if exit != 0 || stdout != each.want || stderr != "" {
				t.Fatalf("exit %d stdout %q stderr %q; want %q", exit, stdout, stderr, each.want)
			}
		})
	}
}

// Dropping the byte store reds TestObjectAssignCopiesRepresentationBytes/null
// and /undefined. A raw-byte copy reds /sentinel, where the interpreting read must normalize it.
func TestObjectAssignRepresentationByteMutants(t *testing.T) {
	t.Parallel()
	for _, variant := range []string{"drop", "raw"} {
		t.Run(variant, func(t *testing.T) {
			binary := objectAssignBytesBinary(t, variant)
			modes := []string{"null", "undefined"}
			if variant == "raw" {
				modes = []string{"sentinel"}
			}
			for _, mode := range modes {
				stdout, stderr, exit := jsonObjectsRuntimeRun(t, binary, mode)
				if exit != 3 || stdout != "" || stderr != "" {
					t.Fatalf("%s/%s must fail its byte assertion: exit %d stdout %q stderr %q", variant, mode, exit, stdout, stderr)
				}
				t.Logf("TestObjectAssignCopiesRepresentationBytes/%s reds with %s", mode, variant)
			}
		})
	}
}
