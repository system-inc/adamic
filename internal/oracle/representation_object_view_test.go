package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationObjectFixture = "internal/oracle/testdata/representation_object_view.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationObjectFixture, true, false})
}

func TestRepresentationObjectViewIdentityMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, representationObjectFixture))
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
		if !strings.HasPrefix(function.Name, "holdObject") {
			continue
		}
		for at, statement := range function.Body {
			if result, ok := statement.(ir.Return); ok && result.Value != nil && result.Value.Type() == ir.Union {
				function.Body[at] = ir.Return{Value: ir.Undefined{Of: ir.Union}}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("want one object view function, changed %d", changed)
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
	t.Log("caught by Node stdout; object views lost identity")
}
