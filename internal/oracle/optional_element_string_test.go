package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/optional_element_string.a", true, false})
}

func TestOptionalStringElementMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_element_string.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "optional_string_element" {
			continue
		}
		value := function.Body[0].(ir.Return).Value.(ir.Conditional)
		function.Body[0] = ir.Return{Value: value.WhenNot}
		changed++
	}
	if changed == 0 {
		t.Fatal("mutant changed no optional string elements")
	}
	want := onNode(t, path)
	backend := onJavaScriptBackend(t, program)
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": backend} {
		if difference := disagreement(want, got); difference == "" {
			t.Errorf("%s accepted the missing short-circuit mutant", name)
		} else {
			t.Logf("%s caught mutant: %s; Node stdout %q; mutant stdout %q, exit %d", name, difference, want.stdout, got.stdout, got.exitCode)
		}
	}
}

// Nullable string types are not represented yet. Inject null into the same fixture's IR and
// source to hold the element lowering to Node without changing the type representation rules.
func TestOptionalStringElementNull(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_element_string.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	line := program.Main[0].(ir.WriteLine)
	call := line.Value.(ir.Call)
	call.Arguments[0] = ir.Null{Of: ir.String}
	line.Value = call
	program.Main[0] = line
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Replace(string(source), "word: string | undefined", "word: string | null | undefined", 1))
	path = filepath.Join(t.TempDir(), "null.a")
	if err := os.WriteFile(path, []byte(strings.Replace(string(source), "first(undefined)", "first(null)", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	backend := onJavaScriptBackend(t, program)
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": backend} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s null receiver: %s; stdout %q, stderr %q", name, difference, got.stdout, got.stderr)
		}
	}
}
