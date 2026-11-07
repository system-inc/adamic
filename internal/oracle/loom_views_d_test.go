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
func TestLoomViewsDOriginal(test *testing.T) {
	test.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/loom_views_d/checked-cast.ts")
	path, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	if want.exitCode != 0 || string(want.stdout) != "8\n" {
		test.Fatalf("Node: %#v", want)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
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

func loomViewsDAgrees(test *testing.T, relative string) {
	test.Helper()
	loomViewsDAgreesBackend(test, relative, "")
}

func loomViewsDAgreesBackend(test *testing.T, relative, backend string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, relative))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	if want.exitCode != 0 {
		test.Fatalf("Node: %#v", want)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	for _, candidate := range []string{"native", "sanitized", "javascript"} {
		if backend != "" && backend != candidate {
			continue
		}
		var got run
		switch candidate {
		case "native":
			got = releasedUncached(test, program)
		case "sanitized":
			got, _ = nativelyUncached(test, program)
		case "javascript":
			got = onJavaScriptBackend(test, program)
		}
		if diff := disagreement(want, got); diff != "" {
			test.Errorf("%s: %s, stderr %q", candidate, diff, got.stderr)
		}
	}
}
func TestLoomViewsDTupleObject(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_tuple_recognition/p68.ts", "native")
}
func TestLoomViewsDTupleUnion(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_tuple_recognition/p69.ts", "native")
}
func TestLoomViewsDWideCallable(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_wrong_aborts/wide.ts", "native")
}
func TestLoomViewsDTupleObjectSanitized(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_tuple_recognition/p68.ts", "sanitized")
}
func TestLoomViewsDTupleObjectJavaScript(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_tuple_recognition/p68.ts", "javascript")
}
func TestLoomViewsDTupleUnionSanitized(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_tuple_recognition/p69.ts", "sanitized")
}
func TestLoomViewsDTupleUnionJavaScript(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_tuple_recognition/p69.ts", "javascript")
}
func TestLoomViewsDWideCallableSanitized(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_wrong_aborts/wide.ts", "sanitized")
}
func TestLoomViewsDWideCallableJavaScript(test *testing.T) {
	test.Parallel()
	loomViewsDAgreesBackend(test, "internal/lower/testdata/fx7_wrong_aborts/wide.ts", "javascript")
}
func TestLoomViewsDCallableUnion(test *testing.T) {
	test.Parallel()
	loomViewsDAgrees(test, "internal/lower/testdata/fx7_wrong_aborts/p20.ts")
}

// A stored function that fits no member of the view's callable target is a cast the checker was told to believe, so
// the read through the view is refused where it stands, as it was before views D, never admitted to abort at run time
// where Node prints 99 (Compiler's ruling (b) on #aecx10a). In .a the cast before it is refused first: its tag test,
// kind === 'Receiver', is one every Root passes (#q97a1rr); its .ts twin keeps the cast a checked view and is refused
// at the read.
func TestLoomViewsDBadCallable(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/lower/testdata/fx7_wrong_aborts/bad_callable.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != "before\nargument\n99\n" {
		test.Fatalf("Node: %#v", node)
	}
	_, loweredError := lowered(test, path)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		!strings.Contains(loweredError.Error(), "bad_callable.a:6:14:") ||
		!strings.Contains(loweredError.Error(), "a downcast to Receiver whose tag test is vacuous") ||
		!strings.Contains(loweredError.Error(), "give Receiver a tag of its own") {
		test.Fatalf("want the located vacuous downcast refusal at 6:14, got %v", loweredError)
	}
	_, loweredError = lowered(test, typeScriptTwin(test, path))
	if !errors.As(loweredError, &refused) ||
		!strings.Contains(loweredError.Error(), "bad_callable.ts:9:16:") ||
		!strings.Contains(
			loweredError.Error(),
			"checked view read of field value with unsupported representation conversion contract",
		) ||
		!strings.Contains(
			loweredError.Error(),
			"prove or implement the representation conversion contract before reading this field",
		) {
		test.Fatalf("twin: want the located representation conversion refusal at 9:16, got %v", loweredError)
	}
}
