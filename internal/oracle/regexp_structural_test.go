package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

var regexpStructuralFixtures = []struct{ name, path, stdout string }{
	{"method", "internal/oracle/testdata/review/agree/fxspptb_8e67677_structural_method.a", "true\n"},
	{"optional", "internal/oracle/testdata/review/agree/fxspptb_8e67677_structural_optional.a", "has test\nno test\n"},
	{"optional_calls", "internal/oracle/testdata/regexp_structural/optional_calls.a", "true\ntrue\nundefined\ntrue\ntrue\n"},
}

func init() {
	for _, fixture := range regexpStructuralFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{fixture.path, true, false})
	}
}

// Emptying the served table reds method and optional at their view relation. Calls retain
// the allocation as their receiver; value reads are held separately by TestRegExpServedValueRoads.
func TestRegExpStructuralMethods(t *testing.T) {
	t.Parallel()
	for _, fixture := range regexpStructuralFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			want := run{stdout: []byte(fixture.stdout)}
			if difference := disagreement(want, truth); difference != "" {
				t.Fatalf("Node: %s: %+v", difference, truth)
			}
			native, binary := nativelyUncached(t, program)
			for name, got := range map[string]run{"sanitized": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Error(report)
			}
		})
	}
}

func TestRegExpStructuralInKeepsHostStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_structural/in.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("true\nfalse\n")}, truth); difference != "" {
		t.Fatalf("Node: %s: %+v", difference, truth)
	}
	want := run{stderr: []byte("adamic: panic: this runtime value's properties aren't read by a run-time key yet (reading 'test')\n"), exitCode: 70}
	native, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s: %+v", name, difference, got)
		}
	}
}

func TestRegExpStructuralRequiredDetachIsRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_structural/required_detached.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	want := ":3:11: Adamic 0.1 refuses a method read as a value (test would lose its object, and this with it); call it in an arrow that keeps the object: (text) => t.test(text) (unbound-method)"
	if err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %q, got %v", want, err)
	}
}
