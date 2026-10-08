package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestClassFeaturesReviewInitializerRefusals(t *testing.T) {
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
			if !errors.As(err, &refusal) || !strings.Contains(refusal.What, probe.path) || !strings.Contains(refusal.Fix, "constructor after super") {
				t.Fatalf("want path and repair, got %v", err)
			}
		})
	}
}

func TestClassFeaturesReviewStaticRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/classfeat_static_virtual.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 70 || !strings.Contains(string(observed.stderr), "TypeError") {
		t.Fatalf("Node: %+v", observed)
	}
	_, err = lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "static field read before") || !strings.Contains(refusal.Fix, "declare the field earlier") {
		t.Fatalf("want initialization refusal, got %v", err)
	}
}
