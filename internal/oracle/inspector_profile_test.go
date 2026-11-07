package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/node_inspector_profile_data.a", true, false})
}

// The ordinary families also hold this witness on release and WASI; this small
// native/JavaScript unit is the named pin for an injected loud url check.
func TestInspectorProfileDataAgreesWithNode(test *testing.T) {
	test.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/node_inspector_profile_data.a")
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	truth := onSliceNode(test, path)
	want := run{stdout: []byte("file:///project/a.ts\nfile:///elsewhere/b.ts\na.ts\nexternal.js\n")}
	if why := disagreement(want, truth); why != "" {
		test.Fatal(why)
	}
	native, _ := nativelyUncached(test, program)
	for name, actual := range map[string]run{"native": native, "javascript": onSliceJavaScript(test, program)} {
		if why := disagreement(truth, actual); why != "" {
			test.Fatalf("%s: %s", name, why)
		}
	}
}
