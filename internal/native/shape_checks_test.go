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

func TestShapeChecksUseDeclaredNames(test *testing.T) {
	test.Parallel()
	emitter := &emitter{shapes: map[string]string{"z": "named_shape_z", "a": "named_shape_a"}}
	want := "#ifdef ADAMIC_COUNT\n\tadamic_shape_check(&named_shape_a);\n\tadamic_shape_check(&named_shape_z);\n#endif\n"
	if got := emitter.shapeChecks(); got != want {
		test.Fatalf("startup must use sorted declared names: %q", got)
	}
}

// A public class layout and a named-group layout can be emitted without being
// allocated. Corrupt just that descriptor: allocation checks cannot catch it.
func checkUnallocatedShapeStartup(test *testing.T, source, kind string) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "shapes.a")
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	node := inlineArrayRun(test, "node", "--input-type=module-typescript", "--eval", source)
	if node.exit != 0 || string(node.stdout) != "alive\n" || len(node.stderr) != 0 {
		test.Fatalf("Node: %+v", node)
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
		test.Fatal("missing emitted descriptors")
	}
	if !reflect.DeepEqual(names, checked) {
		test.Logf("declared %v; startup checked %v", names, checked)
	}
	binary := filepath.Join(test.TempDir(), "control")
	if buildError := Build(code, binary, Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	control := inlineArrayRun(test, binary)
	if control.exit != 0 ||
		string(control.stdout) != "alive\n" ||
		!strings.HasPrefix(string(control.stderr), "adamic: counts:") {
		test.Fatalf("control: %+v", control)
	}
	pattern := regexp.MustCompile(`(?m)(static const adamic_field_kind (\w+)_kinds\[\] = \{)([^}]*)(\};)`)
	targets := pattern.FindAllStringSubmatch(code, -1)
	target := ""
	for _, match := range targets {
		// Public class layouts have exactly one boolean; capture groups are all references.
		if kind == "boolean" && strings.TrimSpace(match[3]) == "adamic_field_boolean" ||
			kind == "reference" && strings.Contains(match[2], "group") {
			target = match[0]
			break
		}
	}
	if target == "" {
		test.Fatal("missing unallocated descriptor", kind)
	}
	before, after := "adamic_field_boolean", "adamic_field_reference"
	if kind == "reference" {
		before, after = after, "adamic_field_number"
	}
	code = strings.Replace(code, target, strings.Replace(target, before, after, 1), 1)
	binary = filepath.Join(test.TempDir(), "bitmap-mutant")
	if buildError := Build(code, binary, Options{Count: true}); buildError != nil {
		test.Fatal("mutant must compile:", buildError)
	}
	mutant := inlineArrayRun(test, binary)
	if mutant.exit == 0 ||
		len(mutant.stdout) != 0 ||
		!strings.Contains(string(mutant.stderr), "inconsistent shape kind for field") {
		test.Fatalf("startup mutant escaped before user code: %+v", mutant)
	}
	test.Log(fmt.Sprintf("unallocated %s bitmap mismatch caught before Node's alive line", kind))
}

func TestCountedPublicShapeStartup(test *testing.T) {
	test.Parallel()
	checkUnallocatedShapeStartup(
		test,
		"class Unused { #hidden: number = 1; value: boolean = false; }\nconst instance = new Unused();\nconsole.log('alive');\n",
		"boolean",
	)
}
func TestCountedCaptureGroupStartup(test *testing.T) {
	test.Parallel()
	checkUnallocatedShapeStartup(test, "const expression = /(?<word>a)/;\nconsole.log('alive');\n", "reference")
}
