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

func TestCensusAppendResultProof(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_append_overload.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	for _, constant := range program.Strings {
		if strings.Contains(constant, "of append result:") {
			test.Fatalf("the real append body should prove its resolved results, got check %q", constant)
		}
	}
}

// The original lying program is held to Node's observation; both Adamic
// backends are independently held to the ruled loud stop, including its text.
func TestCensusOverloadResultStop(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_overload_lie.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	node := onNode(test, path)
	if difference := disagreement(run{stdout: []byte("called\nundefined\n")}, node); difference != "" {
		test.Fatalf("Node: %s", difference)
	}
	want := run{
		stdout:   []byte("called\n"),
		stderr:   []byte("adamic: panic: overload 1 of lie result: expected string, got undefined\n"),
		exitCode: 70,
	}
	native, _ := natively(test, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Errorf("%s: %s; got %#v", name, difference, result)
		}
	}
}
