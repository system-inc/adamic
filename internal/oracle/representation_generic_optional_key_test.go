package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationOptionalKeyFixture = "internal/oracle/testdata/representation_generic_optional_key.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationOptionalKeyFixture, true, false})
}

func TestRepresentationGenericOptionalKeyPresenceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, representationOptionalKeyFixture))
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
		if !strings.HasPrefix(function.Name, "firstKey_") {
			continue
		}
		for at, statement := range function.Body {
			if result, ok := statement.(ir.Return); ok && result.Value != nil {
				of := result.Value.Type()
				if of.IsMaybe() {
					function.Body[at] = ir.Return{Value: ir.MaybeOf{Of: of}}
				} else if of.IsReference() {
					function.Body[at] = ir.Return{Value: ir.Undefined{Of: of}}
				} else {
					t.Fatalf("unexpected optional key return type %v", of)
				}
				changed++
			}
		}
	}
	if changed != 4 {
		t.Fatalf("want number, string, boolean and object specializations, changed %d", changed)
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
	t.Log("caught by Node stdout; zero, empty string, false and an object became missing keys")
}
