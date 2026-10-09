package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Counted startup checks must use the stable names declared by the emitter,
// even when layouts have identical fields with different scalar kinds.
func TestCountedStableShapeChecksCatchMutant(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "shapes.a")
	source := `const number = { value: 7 };
const boolean = { value: false };
console.log(number.value.toString());
console.log(boolean.value ? 'true' : 'false');
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	declarations := regexp.MustCompile(`static const adamic_shape ([A-Za-z_0-9]+) =`).FindAllStringSubmatch(generated, -1)
	checks := regexp.MustCompile(`(?m)^\tadamic_shape_check\(&([A-Za-z_0-9]+)\);$`).FindAllStringSubmatch(generated, -1)
	wanted, got := []string{}, []string{}
	for _, declaration := range declarations {
		wanted = append(wanted, declaration[1])
	}
	for _, check := range checks {
		got = append(got, check[1])
	}
	sort.Strings(wanted)
	if len(wanted) < 2 || !reflect.DeepEqual(got, wanted) {
		t.Fatalf("startup checks %v; declared shapes %v", got, wanted)
	}
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", oracle, path)
	binary := filepath.Join(directory, "counted")
	if err := Build(generated, binary, Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).Output()
	if err != nil || string(output) != want {
		t.Fatalf("counted: %v: %q; Node %q", err, output, want)
	}
	mutant := generated
	for index, check := range checks {
		mutant = strings.Replace(mutant, check[0], fmt.Sprintf("\tadamic_shape_check(&adamic_shape_%d);", index), 1)
	}
	if err := Build(mutant, filepath.Join(directory, "mutant"), Options{Count: true}); err == nil || !strings.Contains(err.Error(), "undeclared identifier 'adamic_shape_0'") {
		t.Fatalf("numeric-name mutant escaped: %v", err)
	}
}
