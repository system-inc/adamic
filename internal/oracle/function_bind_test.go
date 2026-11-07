package oracle

import (
	"path/filepath"
	"strings"
	"testing"
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
		"length":            "a bound function's length",
		"typeof_length":     "a bound function's length",
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
		"assigned_receiver":  "method bind with an assigned receiver identifier",
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
