package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const representationMultimapFixture = "internal/oracle/testdata/representation_generic_multimap.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationMultimapFixture, true, false})
}

func TestRepresentationGenericMultimapMissingAddsMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, representationMultimapFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var removeAdds func([]ir.Statement) int
	removeAdds = func(body []ir.Statement) int {
		changed := 0
		for index, statement := range body {
			switch statement := statement.(type) {
			case ir.Evaluate:
				if add, ok := statement.Value.(ir.SetAdd); ok {
					body[index] = ir.Evaluate{Value: add.Set}
					changed++
				}
			case ir.Block:
				changed += removeAdds(statement.Body)
			case ir.If:
				changed += removeAdds(statement.Then) + removeAdds(statement.Else)
			}
		}
		return changed
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if strings.HasPrefix(function.Name, "addToMultimap_") {
			changed += removeAdds(function.Body)
		}
	}
	if changed != 6 {
		t.Fatalf("want two adds in each of three specializations, changed %d", changed)
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
		t.Fatalf("want Node stdout to kill dropped-adds mutant, got %q", difference)
	}
	t.Log("Node stdout caught absent number, string and boolean set entries")
}
