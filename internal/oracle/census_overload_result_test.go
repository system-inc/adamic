package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, fixture := range []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/census_append_overload.a", true, false},
		{"internal/oracle/testdata/census_overload_lie.a", true, true},
	} {
		fixtures = append(fixtures, fixture)
	}
}

func TestCensusAppendResultProof(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_append_overload.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, constant := range program.Strings {
		if strings.Contains(constant, "of append result:") {
			t.Fatalf("the real append body should prove its resolved results, got check %q", constant)
		}
	}
}

// The original lying program is held to Node's observation; both Adamic
// backends are independently held to the ruled loud stop, including its text.
func TestCensusOverloadResultStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_overload_lie.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("called\nundefined\n")}, node); difference != "" {
		t.Fatalf("Node: %s", difference)
	}
	want := run{stdout: []byte("called\n"), stderr: []byte("adamic: panic: overload 1 of lie result: expected string, got undefined\n"), exitCode: 70}
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: %s; got %#v", name, difference, result)
		}
	}
}
