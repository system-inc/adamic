package oracle

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func TestOct6InheritanceMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name   string
		mutate func(*testing.T, *ir.Program)
	}{
		{"forget override inherited past second level", func(t *testing.T, p *ir.Program) {
			first, second := -1, -1
			for i, f := range p.Functions {
				if f.Name == "First_inherited" {
					first = i
				}
				if f.Name == "Second_inherited" {
					second = i
				}
			}
			if first < 0 || second < 0 {
				t.Fatal("missing method")
			}
			changed := false
			for i, c := range p.Classes {
				if c.Name == "Fourth" {
					for j, target := range c.Methods {
						if target == second {
							p.Classes[i].Methods[j] = first
							changed = true
						}
					}
				}
			}
			if !changed {
				t.Fatal("missing inherited slot")
			}
		}},
		{"lose fourth level ancestry", func(t *testing.T, p *ir.Program) {
			for i, c := range p.Classes {
				if c.Name == "Fourth" {
					p.Classes[i].Base = 0
					return
				}
			}
			t.Fatal("missing fourth class")
		}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_oct6_deep.a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			mutant.mutate(t, p)
			got, _ := natively(t, p)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant must compile and run cleanly: %+v", got)
			}
			if difference := disagreement(want, got); difference != "stdout differs" {
				t.Fatalf("mutant survived: %q", difference)
			}
			t.Logf("caught: Node %q; mutant %q", want.stdout, got.stdout)
		})
	}
}

// Only the leak check can catch this mutant: the same dynamic strings print,
// no invalid access occurs, and compilation must succeed under -Werror.
func TestOct6ReleaseMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_oct6_release.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(p)
	mutated := strings.ReplaceAll(code, "adamic_release(", "(void)(")
	if code == mutated {
		t.Fatal("mutant changed no releases")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutated, binary, native.Options{Sanitize: runtime.GOOS == "linux", Malloc: runtime.GOOS == "darwin"}); err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("only the leak check should catch this: %s, stderr %q", difference, got.stderr)
	}
	var report string
	switch runtime.GOOS {
	case "linux":
		report = leakSanitizer(t, binary)
	case "darwin":
		report = leaksCounted(t, mutated)
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
	if !strings.Contains(report, "LeakSanitizer") && !strings.Contains(report, "leaked") {
		t.Fatalf("leak mutant survived: %s", report)
	}
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, "SUMMARY:") || strings.Contains(line, "leaked") {
			t.Logf("caught by leak check alone: %s", line)
		}
	}
}
