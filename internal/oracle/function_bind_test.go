package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var methodBindFixtures = []string{"method_bind.a", "method_bind_lifetime.a", "bind_cycle_fixed_arrow.a", "bind_cycle_fixed_bound.a", "bind_cycle_host_arrow.a", "bind_cycle_host_bound.a", "bind_cycle_cache_arrow.a", "bind_cycle_cache_bound.a"}

func init() {
	for _, name := range methodBindFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/function_bind.a", true, false})
}

// Runs the new witness only; the family registration also holds it in every ordinary lane.
func TestFunctionBind(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/function_bind.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	backend := onJavaScriptBackend(t, program)
	if diff := disagreement(truth, backend); diff != "" {
		t.Fatalf("javascript: %s; Node stdout %q stderr %q exit %d; backend stdout %q stderr %q exit %d", diff, truth.stdout, truth.stderr, truth.exitCode, backend.stdout, backend.stderr, backend.exitCode)
	}
	actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" {
		t.Fatalf("leaks: %s", leaked)
	}
	if comparable {
		if diff := disagreement(truth, actual); diff != "" {
			t.Fatalf("native: %s", diff)
		}
	}
}

func TestFunctionBindRefusals(t *testing.T) {
	t.Parallel()
	for name, reason := range map[string]string{
		"this_widening":     "this outside a method",
		"inherited_library": "bind of an inherited library function",
		"name":              "a bound function's name",
		"rebind":            "bind of a bound function",
		"new":               "new on a bound function",
		"dynamic_length":    "dynamic function descriptors",
		"dynamic_name":      "dynamic function descriptors",
		"dynamic_bind":      "dynamic function descriptors",
	} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), reason) {
				t.Fatalf("want %q, got %v", reason, err)
			}
		})
	}
	t.Run("opaque_metadata", func(t *testing.T) {
		t.Parallel()
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused/opaque_metadata.ts"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = lowered(t, path)
		if err == nil || !strings.Contains(err.Error(), "a dynamic-shape object in a graph program") || !strings.Contains(err.Error(), "opaque_metadata.ts:4:") {
			t.Fatalf("want the existing graph/dynamic refusal at its source read, got %v", err)
		}
	})
}

func TestFunctionBindWASI(t *testing.T) {
	t.Parallel()
	for index, fixture := range fixtures {
		if fixture.path == "internal/oracle/testdata/function_bind.a" {
			runWASIFixtures(t, []int{index}, nil)
			return
		}
	}
	t.Fatal("bind witness missing from oracle fixtures")
}

func TestMethodBind(t *testing.T) {
	t.Parallel()
	for _, name := range methodBindFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			backend := onJavaScriptBackend(t, program)
			if diff := disagreement(truth, backend); diff != "" {
				t.Fatalf("javascript %s: Node %q (%s), backend %q (%s)", diff, truth.stdout, truth.stderr, backend.stdout, backend.stderr)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if comparable {
				if diff := disagreement(truth, actual); diff != "" {
					t.Fatalf("native %s: Node %q, native %q (%s)", diff, truth.stdout, actual.stdout, actual.stderr)
				}
			}
			if actual := released(t, program); disagreement(truth, actual) != "" {
				t.Fatalf("release native differs: %q (%s)", actual.stdout, actual.stderr)
			}
		})
	}
}

func TestMethodBindWASI(t *testing.T) {
	t.Parallel()
	rows := []int{}
	for index, fixture := range fixtures {
		for _, name := range methodBindFixtures {
			if fixture.path == "internal/oracle/testdata/"+name {
				rows = append(rows, index)
			}
		}
	}
	runWASIFixtures(t, rows, nil)
}

func TestMethodBindRefusals(t *testing.T) {
	t.Parallel()
	for name, reason := range map[string]string{
		"different_receiver": "method bind with a different receiver",
		"call_receiver":      "bind with a call as its method receiver",
		"detached":           "unbound-method",
		"getter_path":        "bind through a receiver path",
		"path":               "bind through a receiver path",
		"getter":             "bind of a getter-backed method value",
		"assigned_receiver":  "method bind whose receiver h isn't a const or a parameter the function never assigns (a parameter the function assigns)",
		"method_write":       "unbound-method",
		"index_receiver":     "bind through an indexed receiver",
		"optional_tail":      "a call through ?.",
		"async_fixed":        "suspension inside method binding operands",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), reason) {
				t.Fatalf("want %q, got %v", reason, err)
			}
		})
	}
}

func TestMethodBindCheckedReceiver(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused/checked_receiver.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a cast the runtime can't check") {
		t.Fatalf("want the existing checked receiver cast refusal, got %v", err)
	}
}

func TestMethodBindRecordBivarianceRefusal(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_runtime/record_nonfunction.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	want := path + ":4:3: Adamic 0.1 refuses a function taking Dictionary seen as one taking Record<string, string | (() => void)> " +
		"(tsc relates a method's parameters both ways), so it can be handed what it can't take; " +
		"write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, " +
		"or take the wider type in the method (method-signature-style)"
	if err == nil || err.Error() != want {
		t.Fatalf("want %q, got %v", want, err)
	}
}

// Expected native failures use run/disagreement, separately from Node's TypeError.
func TestMethodBindRecordNonfunction(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_runtime/record_nonfunction_after_call.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), "TypeError") {
		t.Fatalf("source Node must throw TypeError: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: TypeError: record key 'm' is not a function\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %+v", diff, got)
		}
	}
}

// A bound closure calls its target with the arguments its bound type names, while JavaScript's bind hands
// on every argument the bound function receives (#04f5s0s, Language's J-xet5sem). seen declares an index,
// and held as (shape) => boolean, then bound, find's index never reached it: narrow_type printed "seen at
// undefined" twice on both backends where Node prints "seen at 0" and "seen at 1", and narrow_default,
// its defaulted twin, prints -1. Each stops at the bind, naming both counts, until the bound closure
// forwards every argument.
// own_type, bound as seen's own type, and narrow_direct, passed with no bind, agree with Node natively
// under the sanitizers, released, and on the JavaScript backend. The stored .a is Node's source, and its
// .ts copy is what lowers, as predicate_parameter's pins do: the .a refuses find's `array[i]!` first.
// The count asks only the functions the allocation flow says can be the bound value, so a bind whose
// reach names its functions counts those alone. The four optional-property fixtures have unknown
// allocation reach, but the name scan reads every initialization and write of f, so all four now
// lower and agree. known_reach binds only one beside two, another Unary with two parameters; the
// known allocation reach still takes precedence. Mutants are recorded beside Lower's tests.
func TestFunctionBindExtraParameters(t *testing.T) {
	t.Parallel()
	const seen = "seen at 0\nseen at 1\ncircle\n"
	const narrowCount = "bind of a function that declares 2 parameters where the type it's bound as names 1"
	for _, probe := range []struct{ name, nodeOut, site, what string }{
		{"narrow_type", seen, "narrow.bind(undefined)", narrowCount},
		{"narrow_default", seen, "narrow.bind(undefined)", narrowCount},
		{"own_type", seen, "", ""},
		{"narrow_direct", seen, "", ""},
		{"known_reach", "2\n2\n", "", ""},
		{"absent_property", "10\nundefined\n", "", ""},
		{"present_property", "10\nfunction\n42\n", "", ""},
		{"absent_fixed_property", "10\nundefined\n", "", ""},
		{"null_property", "10\nundefined\n", "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			witness, checked := predicateInputFrom(t, "internal/oracle/testdata/bind_extra_parameters", probe.name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatalf("source Node: %s, stdout %q", difference, node.stdout)
			}
			program, err := lowered(t, checked)
			if probe.what == "" {
				if err != nil {
					t.Fatal(err)
				}
				sanitized, _ := nativelyUncached(t, program)
				for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
					}
				}
				return
			}
			if err == nil {
				sanitized, _ := nativelyUncached(t, program)
				javaScript := onJavaScriptBackend(t, program)
				t.Fatalf("lowered: sanitized native stdout %q (%s), the JavaScript backend stdout %q (%s)", sanitized.stdout, disagreement(node, sanitized), javaScript.stdout, disagreement(node, javaScript))
			}
			source, readErr := os.ReadFile(checked)
			if readErr != nil {
				t.Fatal(readErr)
			}
			site := ""
			for number, line := range strings.Split(string(source), "\n") {
				if column := strings.Index(line, probe.site); column >= 0 {
					site = fmt.Sprintf("%s:%d:%d", checked, number+1, column+1)
				}
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.what || notYet.Where != site {
				t.Fatalf("want NotYet %q at %s, got %v", probe.what, site, err)
			}
		})
	}
}

// The name fallback admits a readable f beside unrelated copies and computed writes, even
// though two is held as Unary elsewhere and the allocation-flow projection is unknown.
func TestFunctionBindNameReach(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_name_reach.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("2\n3\n4\n")}, truth); difference != "" {
		t.Fatalf("source Node: %s", difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" || !comparable {
		t.Fatalf("sanitized native: leaked %q, comparable %t", leaked, comparable)
	}
	for index, actual := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatalf("%s: %s", []string{"sanitized native", "release native", "JavaScript backend"}[index], difference)
		}
	}
}
