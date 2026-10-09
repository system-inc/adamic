package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestWeakDifferentialViewMethodParameter(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_view_method_parameter", "t1 b1\n", "view")
}

func TestWeakDifferentialViewMethod(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_view_method", "true b1\n", "view")
}

func TestWeakDifferentialUnionNarrowed(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "4ddd17f_weak_union_narrowed", "before\nafter: nn\n", "weak")
}

func TestWeakDifferentialSingleNarrowed(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "4ddd17f_weak_single_narrowed", "before\nafter: nn\n", "weak")
}

func TestWeakDifferentialEach(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_narrowed_each", "a1\nb1\n", "callback")
}

func TestWeakDifferentialFind(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_narrowed_find", "b1\nb1\nb1\n", "callback")
}

func TestWeakDifferentialMap(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_narrowed_map", "a1,b1\n", "callback")
}

func TestWeakDifferentialReduce(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_narrowed_reduce", "4\n", "callback")
}

func TestWeakDifferentialSome(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_narrowed_some", "true\n", "callback")
}

func TestWeakDifferentialSort(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "9984394_narrowed_sort", "true\n", "callback")
}

func weakDifferential(t *testing.T, name, output, shape string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/weak_miscompiles", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	expected := run{stdout: []byte(output)}
	if difference := disagreement(expected, node); difference != "" {
		t.Fatalf("source Node: %s: exit=%d stdout=%q stderr=%q", difference, node.exitCode, node.stdout, node.stderr)
	}
	t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	program, err := lowered(t, path)
	if shape != "weak" && shape != "control" {
		diagnostic := "a StrongTaker seen as a WeakTaker (one keeps something weakly that the other keeps strongly)"
		location := ":19:27:"
		if name == "9984394_view_method" {
			diagnostic = "a StrongGetter seen as a WeakGetter (one keeps something weakly that the other keeps strongly)"
			location = ":19:29:"
		}
		if shape == "callback" {
			method := map[string]string{"9984394_narrowed_each": "forEach", "9984394_narrowed_find": "find", "9984394_narrowed_map": "map", "9984394_narrowed_reduce": "reduce", "9984394_narrowed_some": "some", "9984394_narrowed_sort": "sort"}[name]
			diagnostic = method + " callback over narrowed Weak elements whose parameter needs handle-to-target conversion; use a loop with an explicit narrowed copy"
			location = map[string]string{"9984394_narrowed_each": ":10:2:", "9984394_narrowed_find": ":10:16:", "9984394_narrowed_map": ":10:14:", "9984394_narrowed_reduce": ":10:16:", "9984394_narrowed_some": ":10:15:", "9984394_narrowed_sort": ":10:2:"}[name]
		}
		if err != nil {
			var pending *lower.NotYet
			want := path + location + " stage 0 can't lower " + diagnostic + " yet"
			if !errors.As(err, &pending) || err.Error() != want || program != nil {
				t.Fatalf("both backends must stop at %q NotYet, got program=%v error=%v", want, program != nil, err)
			}
			t.Logf("both backends: %v", err)
			return
		}
		// Run an accidentally admitted program too. The revert mutants must
		// recover the old runtime fault or wrong output, not a build warning.
		t.Errorf("both backends unexpectedly admit the pinned %s NotYet", shape)
	}
	if err != nil {
		t.Fatal(err)
	}
	js := onJavaScriptBackend(t, program)
	release := released(t, program)
	sanitized, binary := natively(t, program)
	for backend, got := range map[string]run{"javascript": js, "release": release, "sanitized": sanitized} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		want := expected
		if shape == "weak" && backend != "javascript" {
			want = run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: a weak reference was read after what it pointed to was freed\n"), exitCode: 70}
		}
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
	if shape == "control" && release.exitCode == 0 {
		if leaked := leaks(t, program, binary); leaked != "" {
			t.Errorf("leaks: %s", leaked)
		}
	}
}

func TestWeakDifferentialMatchingMethods(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "matching_methods", "t1 b1\ntrue b1\n", "control")
}

func TestWeakDifferentialBooleanFilter(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "boolean_filter", "a1\nb1\n", "control")
}

func TestWeakDifferentialHeldUnion(t *testing.T) {
	t.Parallel()
	weakDifferential(t, "held_union", "before\nafter: nn\n", "control")
}

func init() {
	for _, name := range []string{"matching_methods", "boolean_filter", "held_union"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/weak_miscompiles/" + name + ".a", true, false})
	}
	for _, name := range []string{"9984394_view_method_parameter", "9984394_view_method", "9984394_narrowed_each", "9984394_narrowed_find", "9984394_narrowed_map", "9984394_narrowed_reduce", "9984394_narrowed_some", "9984394_narrowed_sort"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/weak_miscompiles/" + name + ".a", false, false})
	}
	additionalFixtureCounts = append(additionalFixtureCounts, weakDifferentialCounts)

}

// Weak lifetime is intentionally different from Node tracing, so these run in
// their explicit liveness tests and contribute counted native runs separately.
func weakDifferentialCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range []string{"4ddd17f_weak_union_narrowed", "4ddd17f_weak_single_narrowed"} {
		rows = append(rows, counted(t, "internal/oracle/testdata/weak_miscompiles/"+name+".a", false, nil, false, false))
	}
	return rows
}
