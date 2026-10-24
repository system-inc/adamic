package native

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestShapeChecksUseDeclaredNames(t *testing.T) {
	t.Parallel()
	e := &emitter{shapes: map[string]string{"z": "named_shape_z", "a": "named_shape_a"}}
	want := "#ifdef ADAMIC_COUNT\n\tadamic_shape_check(&named_shape_a);\n\tadamic_shape_check(&named_shape_z);\n#endif\n"
	if got := e.shapeChecks(); got != want {
		t.Fatalf("startup must use sorted declared names: %q", got)
	}
}

// A public class layout and a named-group layout can be emitted without being
// allocated. Corrupt just that descriptor: allocation checks cannot catch it.
func checkUnallocatedShapeStartup(t *testing.T, source, kind string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "shapes.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	node := inlineArrayRun(t, "node", "--input-type=module-typescript", "--eval", source)
	if node.exit != 0 || string(node.stdout) != "alive\n" || len(node.stderr) != 0 {
		t.Fatalf("Node: %+v", node)
	}
	code := C(program)
	declarations := regexp.MustCompile(`static const adamic_shape (\w+) =`).FindAllStringSubmatch(code, -1)
	checks := regexp.MustCompile(`(?m)^\tadamic_shape_check\(&([A-Za-z_0-9]+)\);$`).FindAllStringSubmatch(code, -1)
	var names, checked []string
	for _, declaration := range declarations {
		names = append(names, declaration[1])
	}
	for _, check := range checks {
		checked = append(checked, check[1])
	}
	sort.Strings(names)
	if len(names) == 0 {
		t.Fatal("missing emitted descriptors")
	}
	if !reflect.DeepEqual(names, checked) {
		t.Logf("declared %v; startup checked %v", names, checked)
	}
	binary := filepath.Join(t.TempDir(), "control")
	if err := Build(code, binary, Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	control := inlineArrayRun(t, binary)
	if control.exit != 0 || string(control.stdout) != "alive\n" || !strings.HasPrefix(string(control.stderr), "adamic: counts:") {
		t.Fatalf("control: %+v", control)
	}
	pattern := regexp.MustCompile(`(?m)(static const adamic_field_kind (\w+)_kinds\[\] = \{)([^}]*)(\};)`)
	targets := pattern.FindAllStringSubmatch(code, -1)
	target := ""
	for _, match := range targets {
		// Public class layouts have exactly one boolean; capture groups are all references.
		if kind == "boolean" && strings.TrimSpace(match[3]) == "adamic_field_boolean" || kind == "reference" && strings.Contains(match[2], "group") {
			target = match[0]
			break
		}
	}
	if target == "" {
		t.Fatal("missing unallocated descriptor", kind)
	}
	before, after := "adamic_field_boolean", "adamic_field_reference"
	if kind == "reference" {
		before, after = after, "adamic_field_number"
	}
	code = strings.Replace(code, target, strings.Replace(target, before, after, 1), 1)
	binary = filepath.Join(t.TempDir(), "bitmap-mutant")
	if err := Build(code, binary, Options{Count: true}); err != nil {
		t.Fatal("mutant must compile:", err)
	}
	mutant := inlineArrayRun(t, binary)
	if mutant.exit == 0 || len(mutant.stdout) != 0 || !strings.Contains(string(mutant.stderr), "inconsistent shape kind for field") {
		t.Fatalf("startup mutant escaped before user code: %+v", mutant)
	}
	t.Log(fmt.Sprintf("unallocated %s bitmap mismatch caught before Node's alive line", kind))
}

func TestCountedPublicShapeStartup(t *testing.T) {
	t.Parallel()
	checkUnallocatedShapeStartup(t, "class Unused { #hidden: number = 1; value: boolean = false; }\nconst instance = new Unused();\nconsole.log('alive');\n", "boolean")
}
func TestCountedCaptureGroupStartup(t *testing.T) {
	t.Parallel()
	checkUnallocatedShapeStartup(t, "const expression = /(?<word>a)/;\nconsole.log('alive');\n", "reference")
}
