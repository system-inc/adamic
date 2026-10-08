package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationIntersectionFixture = "internal/oracle/testdata/representation_object_intersection.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationIntersectionFixture, true, false})
}

func TestRepresentationObjectIntersectionIdentityMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, representationIntersectionFixture))
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
		if !strings.HasPrefix(function.Name, "holdRefined") {
			continue
		}
		for at, statement := range function.Body {
			if result, ok := statement.(ir.Return); ok && result.Value != nil && result.Value.Type() == ir.Object {
				function.Body[at] = ir.Return{Value: ir.Undefined{Of: ir.Object}}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("want one intersection function, changed %d", changed)
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
	t.Log("caught by Node stdout; reference-bearing intersection lost identity")
}
