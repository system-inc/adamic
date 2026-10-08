package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationStateFixture = "internal/oracle/testdata/representation_generic_state.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationStateFixture, true, false})
}

func TestRepresentationGenericStateIdentityMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, representationStateFixture))
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
		if !strings.HasPrefix(function.Name, "passState_") {
			continue
		}
		for at, statement := range function.Body {
			if result, ok := statement.(ir.Return); ok && result.Value != nil && result.Value.Type() == ir.Object {
				function.Body[at] = ir.Return{Value: ir.Undefined{Of: ir.Object}}
				changed++
			}
		}
	}
	if changed != 2 {
		t.Fatalf("want both concrete state instantiations, changed %d", changed)
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish without sanitizer failures: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("want stdout comparison to kill identity mutant, got %q", difference)
	}
	t.Log("caught by Node stdout; both specializations lost state identity")
}
