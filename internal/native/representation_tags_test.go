package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Compile the real C table, rather than testing a second parser's interpretation.
func TestRepresentationTagsAgree(t *testing.T) {
	names := []string{"number", "boolean", "string", "object", "array", "map", "maybe_number", "closure", "maybe_boolean", "union", "weak", "null", "undefined", "record", "uint8_array", "int32_array", "float64_array"}
	tags := []ir.Type{ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Map, ir.MaybeNumber, ir.Closure, ir.MaybeBoolean, ir.Union, ir.Weak, ir.NullRepresentation, ir.UndefinedRepresentation, ir.Record, ir.Uint8Array, ir.Int32Array, ir.Float64Array}
	var source, want strings.Builder
	source.WriteString("#include <stdio.h>\n#include \"representation_tags.h\"\nint main(void) {\n")
	for i, name := range names {
		fmt.Fprintf(&source, "printf(\"%%d\\n\", adamic_rep_%s);\n", name)
		fmt.Fprintf(&want, "%d\n", tags[i])
	}
	source.WriteString("return 0; }\n")
	directory := t.TempDir()
	file := filepath.Join(directory, "tags.c")
	binary := filepath.Join(directory, "tags")
	header, err := runtime.ReadFile("runtime/representation_tags.h")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(directory, "representation_tags.h"), header, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("clang", "-std=c11", "-Wall", "-Wextra", "-Werror", file, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile tags: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run tags: %v\n%s", err, out)
	}
	if string(out) != want.String() {
		t.Fatalf("C representation tags differ from Go: got %q, want %q", out, want.String())
	}
}
