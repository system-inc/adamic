package oracle

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func TestOct6InheritanceMutants(test *testing.T) {
	test.Parallel()
	for _, mutant := range []struct {
		name   string
		mutate func(*testing.T, *ir.Program)
	}{
		{"forget override inherited past second level", func(test *testing.T, p *ir.Program) {
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
				test.Fatal("missing method")
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
				test.Fatal("missing inherited slot")
			}
		}},
		{"lose fourth level ancestry", func(test *testing.T, p *ir.Program) {
			for i, c := range p.Classes {
				if c.Name == "Fourth" {
					p.Classes[i].Base = 0
					return
				}
			}
			test.Fatal("missing fourth class")
		}},
	} {
		test.Run(mutant.name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/class_oct6_deep.a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			p, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want := onNode(test, path)
			mutant.mutate(test, p)
			got, _ := natively(test, p)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("mutant must compile and run cleanly: %+v", got)
			}
			if difference := disagreement(want, got); difference != "stdout differs" {
				test.Fatalf("mutant survived: %q", difference)
			}
			test.Logf("caught: Node %q; mutant %q", want.stdout, got.stdout)
		})
	}
}

// Only the leak check can catch this mutant: the same dynamic strings print,
// no invalid access occurs, and compilation must succeed under -Werror.
func TestOct6ReleaseMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_oct6_release.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	p, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := native.C(p)
	mutated := strings.ReplaceAll(code, "adamic_release(", "(void)(")
	if code == mutated {
		test.Fatal("mutant changed no releases")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(
		mutated,
		binary,
		native.Options{Sanitize: runtime.GOOS == "linux"},
	); buildError != nil {
		test.Fatal(buildError)
	}
	want := onNode(test, path)
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if difference := disagreement(want, got); difference != "" {
		test.Fatalf("only the leak check should catch this: %s, stderr %q", difference, got.stderr)
	}
	report := leakChecked(test, mutated, binary)
	if !strings.Contains(report, "LeakSanitizer") && !strings.Contains(report, "leaked") {
		test.Fatalf("leak mutant survived: %s", report)
	}
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, "SUMMARY:") || strings.Contains(line, "leaked") {
			test.Logf("caught by leak check alone: %s", line)
		}
	}
}
