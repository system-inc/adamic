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
func TestInspectorProfileDataAgreesWithNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/node_inspector_profile_data.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onSliceNode(t, path)
	want := run{stdout: []byte("file:///project/a.ts\nfile:///elsewhere/b.ts\na.ts\nexternal.js\n")}
	if why := disagreement(want, truth); why != "" {
		t.Fatal(why)
	}
	native, _ := nativelyUncached(t, program)
	for name, actual := range map[string]run{"native": native, "javascript": onSliceJavaScript(t, program)} {
		if why := disagreement(truth, actual); why != "" {
			t.Fatalf("%s: %s", name, why)
		}
	}
}
