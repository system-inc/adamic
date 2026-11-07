package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestWeakDifferentialViewMethodParameter(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_view_method_parameter", "t1 b1\n", "view")
}

func TestWeakDifferentialViewMethod(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_view_method", "true b1\n", "view")
}

func TestWeakDifferentialUnionNarrowed(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "4ddd17f_weak_union_narrowed", "before\nafter: nn\n", "weak")
}

func TestWeakDifferentialSingleNarrowed(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "4ddd17f_weak_single_narrowed", "before\nafter: nn\n", "weak")
}

func TestWeakDifferentialEach(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_narrowed_each", "a1\nb1\n", "callback")
}

func TestWeakDifferentialFind(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_narrowed_find", "b1\nb1\nb1\n", "callback")
}

func TestWeakDifferentialMap(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_narrowed_map", "a1,b1\n", "callback")
}

func TestWeakDifferentialReduce(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_narrowed_reduce", "4\n", "callback")
}

func TestWeakDifferentialSome(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_narrowed_some", "true\n", "callback")
}

func TestWeakDifferentialSort(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "9984394_narrowed_sort", "true\n", "callback")
}

func weakDifferential(test *testing.T, name, output, shape string) {
	test.Helper()
	directory := "internal/oracle/testdata/weak_miscompiles"
	if shape == "view" || shape == "callback" {
		directory += "_refused"
	}
	path, absolutePathError := filepath.Abs(filepath.Join(repository, directory, name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := onNode(test, path)
	expected := run{stdout: []byte(output)}
	if difference := disagreement(expected, node); difference != "" {
		test.Fatalf("source Node: %s: exit=%d stdout=%q stderr=%q", difference, node.exitCode, node.stdout, node.stderr)
	}
	test.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	program, loweredError := lowered(test, path)
	if shape != "weak" && shape != "control" {
		diagnostic := "a StrongTaker seen as a WeakTaker (one keeps something weakly that the other keeps strongly)"
		location := ":19:27:"
		if name == "9984394_view_method" {
			diagnostic = "a StrongGetter seen as a WeakGetter (one keeps something weakly that the other keeps strongly)"
			location = ":19:29:"
		}
		if shape == "callback" {
			method := map[string]string{
				"9984394_narrowed_each":   "forEach",
				"9984394_narrowed_find":   "find",
				"9984394_narrowed_map":    "map",
				"9984394_narrowed_reduce": "reduce",
				"9984394_narrowed_some":   "some",
				"9984394_narrowed_sort":   "sort",
			}[name]
			diagnostic = method +
				" callback over narrowed Weak elements whose parameter needs handle-to-target conversion; use a loop with an explicit narrowed copy"
			location = map[string]string{
				"9984394_narrowed_each":   ":10:2:",
				"9984394_narrowed_find":   ":10:16:",
				"9984394_narrowed_map":    ":10:14:",
				"9984394_narrowed_reduce": ":10:16:",
				"9984394_narrowed_some":   ":10:15:",
				"9984394_narrowed_sort":   ":10:2:",
			}[name]
		}
		if loweredError != nil {
			var pending *lower.NotYet
			want := path + location + " stage 0 can't lower " + diagnostic + " yet"
			if !errors.As(loweredError, &pending) || loweredError.Error() != want || program != nil {
				test.Fatalf(
					"both backends must stop at %q NotYet, got program=%v error=%v",
					want,
					program != nil,
					loweredError,
				)
			}
			test.Logf("both backends: %v", loweredError)
			return
		}
		// Run an accidentally admitted program too. The revert mutants must
		// recover the old runtime fault or wrong output, not a build warning.
		test.Errorf("both backends unexpectedly admit the pinned %s NotYet", shape)
	}
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	js := onJavaScriptBackend(test, program)
	release := released(test, program)
	sanitized, binary := natively(test, program)
	for backend, got := range map[string]run{"javascript": js, "release": release, "sanitized": sanitized} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		want := expected
		if shape == "weak" && backend != "javascript" {
			want = run{
				stdout:   []byte("before\n"),
				stderr:   []byte("adamic: panic: a weak reference was read after what it pointed to was freed\n"),
				exitCode: 70,
			}
		}
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
	if shape == "control" && release.exitCode == 0 {
		if leaked := leaks(test, program, binary); leaked != "" {
			test.Errorf("leaks: %s", leaked)
		}
	}
}

func TestWeakDifferentialMatchingMethods(test *testing.T) {
	test.Parallel()
	weakDifferential(test, "matching_methods", "t1 b1\ntrue b1\n", "control")
}

func TestWeakDifferentialBooleanFilterRelease(test *testing.T) {
	test.Parallel()
	checkWeakControlBackend(test, "boolean_filter", "a1\nb1\n", "native")
}

func TestWeakDifferentialBooleanFilterSanitized(test *testing.T) {
	test.Parallel()
	checkWeakControlBackend(test, "boolean_filter", "a1\nb1\n", "native-sanitized")
}

func TestWeakDifferentialBooleanFilterJavaScript(test *testing.T) {
	test.Parallel()
	checkWeakControlBackend(test, "boolean_filter", "a1\nb1\n", "javascript")
}

func TestWeakDifferentialHeldUnionRelease(test *testing.T) {
	test.Parallel()
	checkWeakControlBackend(test, "held_union", "before\nafter: nn\n", "native")
}

func TestWeakDifferentialHeldUnionSanitized(test *testing.T) {
	test.Parallel()
	checkWeakControlBackend(test, "held_union", "before\nafter: nn\n", "native-sanitized")
}

func TestWeakDifferentialHeldUnionJavaScript(test *testing.T) {
	test.Parallel()
	checkWeakControlBackend(test, "held_union", "before\nafter: nn\n", "javascript")
}

func init() {
	for _, name := range []string{"matching_methods", "boolean_filter", "held_union"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/weak_miscompiles/" + name + ".a", true, false})
	}
	for _, name := range []string{
		"9984394_view_method_parameter",
		"9984394_view_method",
		"9984394_narrowed_each",
		"9984394_narrowed_find",
		"9984394_narrowed_map",
		"9984394_narrowed_reduce",
		"9984394_narrowed_some",
		"9984394_narrowed_sort",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/weak_miscompiles_refused/" + name + ".a", false, false})
	}
	// Weak lifetime is intentionally different from Node tracing. Register each
	// explicit liveness witness separately in main's keyed counts registry.
	for _, name := range []string{"4ddd17f_weak_union_narrowed", "4ddd17f_weak_single_narrowed"} {
		path := "internal/oracle/testdata/weak_miscompiles/" + name + ".a"
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{path, func(test *testing.T) (string, func() string) {
				return counted(test, path, false, nil, false, false)
			}},
		)
	}

}

// Controls have ordinary Node lifetime, so every backend and the leak check can
// be held independently without weakening the explicit Weak stop witnesses.
func checkWeakControlBackend(test *testing.T, name, want, backend string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/weak_miscompiles", name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := onNode(test, path)
	if diff := disagreement(run{stdout: []byte(want)}, node); diff != "" {
		test.Fatal("source Node: " + diff)
	}
	checkCallbackControlBackend(test, path, node, backend)
}

func checkCallbackControlBackend(test *testing.T, path string, node run, backend string) {
	test.Helper()
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	var got run
	switch backend {
	case "native":
		got = released(test, program)
	case "javascript":
		got = onJavaScriptBackend(test, program)
	case "native-sanitized":
		var binary string
		got, binary = natively(test, program)
		if released(test, program).exitCode == 0 {
			if report := leaks(test, program, binary); report != "" {
				test.Errorf("leaks: %s", report)
			}
		}
	default:
		test.Fatalf("unknown backend %q", backend)
	}
	if diff := disagreement(node, got); diff != "" {
		test.Errorf("%s: %s", backend, diff)
	}
}
