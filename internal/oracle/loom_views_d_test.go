package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/loom_views_d/checked-cast.ts", true, false})
}
func TestLoomViewsDOriginal(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/loom_views_d/checked-cast.ts")
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "8\n" {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
		}
	}
}

func loomViewsDAgrees(t *testing.T, relative string) {
	t.Helper()
	loomViewsDAgreesBackend(t, relative, "")
}

func loomViewsDAgreesBackend(t *testing.T, relative, backend string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []string{"native", "sanitized", "javascript"} {
		if backend != "" && backend != candidate {
			continue
		}
		var got run
		switch candidate {
		case "native":
			got = releasedUncached(t, program)
		case "sanitized":
			got, _ = nativelyUncached(t, program)
		case "javascript":
			got = onJavaScriptBackend(t, program)
		}
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stderr %q", candidate, diff, got.stderr)
		}
	}
}
func TestLoomViewsDTupleObject(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_tuple_recognition/p68.ts", "native")
}
func TestLoomViewsDTupleUnion(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_tuple_recognition/p69.ts", "native")
}
func TestLoomViewsDWideCallable(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_wrong_aborts/wide.ts", "native")
}
func TestLoomViewsDTupleObjectSanitized(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_tuple_recognition/p68.ts", "sanitized")
}
func TestLoomViewsDTupleObjectJavaScript(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_tuple_recognition/p68.ts", "javascript")
}
func TestLoomViewsDTupleUnionSanitized(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_tuple_recognition/p69.ts", "sanitized")
}
func TestLoomViewsDTupleUnionJavaScript(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_tuple_recognition/p69.ts", "javascript")
}
func TestLoomViewsDWideCallableSanitized(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_wrong_aborts/wide.ts", "sanitized")
}
func TestLoomViewsDWideCallableJavaScript(t *testing.T) {
	t.Parallel()
	loomViewsDAgreesBackend(t, "internal/lower/testdata/fx7_wrong_aborts/wide.ts", "javascript")
}
func TestLoomViewsDCallableUnion(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/lower/testdata/fx7_wrong_aborts/p20.ts")
}

// A stored function that fits no member of the view's callable target is a cast the checker was told to believe, so
// the read through the view is refused where it stands, as it was before views D, never admitted to abort at run time
// where Node prints 99 (Compiler's ruling (b) on #aecx10a). In .a the cast before it is refused first: its tag test,
// kind === 'Receiver', is one every Root passes (#q97a1rr); its .ts twin keeps the cast a checked view and is refused
// at the read.
func TestLoomViewsDBadCallable(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/fx7_wrong_aborts/bad_callable.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "before\nargument\n99\n" {
		t.Fatalf("Node: %#v", node)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "bad_callable.a:6:14:") || !strings.Contains(err.Error(), "a downcast to Receiver whose tag test is vacuous") || !strings.Contains(err.Error(), "give Receiver a tag of its own") {
		t.Fatalf("want the located vacuous downcast refusal at 6:14, got %v", err)
	}
	_, err = lowered(t, typeScriptTwin(t, path))
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "bad_callable.ts:9:16:") || !strings.Contains(err.Error(), "checked view read of field value with unsupported representation conversion contract") || !strings.Contains(err.Error(), "prove or implement the representation conversion contract before reading this field") {
		t.Fatalf("twin: want the located representation conversion refusal at 9:16, got %v", err)
	}
}
