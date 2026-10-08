package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationWrappedFixture = "internal/oracle/testdata/representation_generic_wrapped_value.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationWrappedFixture, true, false})
}

func TestRepresentationGenericWrappedAbsenceMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, representationWrappedFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var removeReturns func([]ir.Statement) int
	removeReturns = func(body []ir.Statement) int {
		changed := 0
		for index, statement := range body {
			switch statement := statement.(type) {
			case ir.Return:
				if statement.Value != nil && statement.Value.Type().IsReference() {
					body[index] = ir.Return{Value: ir.Undefined{Of: statement.Value.Type()}}
					changed++
				}
			case ir.Block:
				changed += removeReturns(statement.Body)
			case ir.If:
				changed += removeReturns(statement.Then) + removeReturns(statement.Else)
			}
		}
		return changed
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if strings.HasPrefix(function.Name, "maybeWrap_") {
			changed += removeReturns(function.Body)
		}
	}
	if changed != 12 {
		t.Fatalf("want three returns in each of four specializations, changed %d", changed)
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("want Node stdout to kill absent-return mutant, got %q", difference)
	}
	t.Log("Node stdout caught present scalar, reference and wrapped values becoming absent")
}
