package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func init() {
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
		"this_widening":  "this outside a method",
		"length":         "a bound function's length",
		"typeof_length":  "a bound function's length",
		"name":           "a bound function's name",
		"rebind":         "bind of a bound function",
		"new":            "new on a bound function",
		"dynamic_length": "dynamic function descriptors",
		"dynamic_name":   "dynamic function descriptors",
		"dynamic_bind":   "dynamic function descriptors",
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
