package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/lower/testdata/fx7_wrong_aborts/bad_callable.a", true, true})

	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/loom_views_d/checked-cast.a", true, false})
}
func TestLoomViewsDOriginal(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/loom_views_d/checked-cast.a")
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
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
		}
	}
}
func TestLoomViewsDTupleObject(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/lower/testdata/fx7_tuple_recognition/p68.a")
}
func TestLoomViewsDTupleUnion(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/lower/testdata/fx7_tuple_recognition/p69.a")
}
func TestLoomViewsDWideCallable(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/lower/testdata/fx7_wrong_aborts/wide.a")
}
func TestLoomViewsDCallableUnion(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/lower/testdata/fx7_wrong_aborts/p20.a")
}
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
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "before\n" {
			t.Errorf("%s: %#v", backend, got)
		}
	}
}
