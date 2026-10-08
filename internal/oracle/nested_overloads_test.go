package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, fixture := range []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/nested_overloads.a", true, false},
		{"internal/oracle/testdata/nested_overload_lie.a", true, true},
	} {
		fixtures = append(fixtures, fixture)
	}
}

// Overload signatures inside a function body once crashed lowering. A lying
// nested overload is held to the same loud stop as one at module scope.
func TestNestedOverloadResultStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/nested_overload_lie.a"))
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
	want := run{stdout: []byte("called\n"), stderr: []byte("adamic: panic: overload 1 of lookup result: expected Map<string, number>, got undefined\n"), exitCode: 70}
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: %s; got %#v", name, difference, result)
		}
	}
}
