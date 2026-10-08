package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestClassWrongOutput103(t *testing.T) {
	for _, probe := range []struct{ name, output, path string }{
		{"init_super_number", "score NaN\n", "Base.describe -> Derived.score -> Derived.count"},
		{"init_super_getter", "score NaN\n", "Base.summary -> Derived.score -> Derived.count"},
		{"init_super", "caught TypeError\n", "Base.describe -> Derived.name -> Derived.label"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/classfeat_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node: %+v", observed)
			}
			_, err = lowered(t, path)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || refusal.What != "a field initializer reading an uninitialized field through "+probe.path || refusal.Fix != "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set" {
				t.Fatalf("want path and repair, got %v", err)
			}
		})
	}
}
