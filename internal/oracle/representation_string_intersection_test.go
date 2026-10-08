package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationStringFixture = "internal/oracle/testdata/representation_string_intersection.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationStringFixture, true, false})
}

func TestRepresentationStringIntersectionMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, representationStringFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	replacement := len(program.Strings)
	program.Strings = append(program.Strings, "mutant")
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if !strings.HasPrefix(function.Name, "hold") {
			continue
		}
		for at, statement := range function.Body {
			if result, ok := statement.(ir.Return); ok && result.Value != nil && result.Value.Type() == ir.String {
				function.Body[at] = ir.Return{Value: ir.StringConstant{Index: replacement}}
				changed++
			}
		}
	}
	if changed != 2 {
		t.Fatalf("want both string intersection functions, changed %d", changed)
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
	t.Log("caught by Node stdout; string intersection results lost their string payload")
}
