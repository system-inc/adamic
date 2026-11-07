package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/loom_views_c/snapshot-original.ts", true, false})
}
func TestLoomViewsCOriginal(test *testing.T) {
	test.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/loom_views_c/snapshot-original.ts")
	path, err := filepath.Abs(path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	if want.exitCode != 0 || string(want.stdout) != "true\nundefined\n" {
		test.Fatalf("Node: %#v", want)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	sanitized, _ := nativelyUncached(test, program)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  sanitized,
		"javascript": onJavaScriptBackend(test, program),
	} {
		if diff := disagreement(want, got); diff != "" {
			test.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
		}
	}
}

// A producer identical to the first callable member, and assignable to the
// second, is carried by the first, natively and on the JavaScript backend.
func TestLoomViewsCExactCallableMember(test *testing.T) {
	test.Parallel()
	loomViewsDAgrees(test, "internal/lower/testdata/callable_producers/p135_exact_member.ts")
}
