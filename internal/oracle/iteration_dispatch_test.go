package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"class", "object", "break", "return", "labeled_continue", "body_throw", "next_throw", "close_throw_body", "close_throw_break", "cached_next", "completion", "done_throw", "value_throw", "method_getters"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "stage3/fixtures/iteration-dispatch/" + name + ".a", lowers: true})
	}
}

func init() {
	for _, name := range []string{"iterators_override_this", "iterators_hidden_return"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/class_wrong_output_refused/" + name + ".a", lowers: true})
	}
}

// Source observations are pinned while the named ruled-divergence dependency is
// absent. These .a lies must stop before native code; .ts terminal checks remain pending.
func TestIterationDispatchPendingStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, output string }{
		{"non_object_step", "TypeError\n"}, {"non_object_iterator", "TypeError\n"},
		{"non_callable_next", "TypeError\n"}, {"non_object_close", "1\nTypeError\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/iteration-dispatch", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || len(observed.stderr) != 0 || string(observed.stdout) != probe.output {
				t.Fatalf("Node: %+v", observed)
			}
			_, failure := lowered(t, path)
			var gap *lower.NotYet
			var refusal *lower.Refused
			if !errors.As(failure, &gap) && !errors.As(failure, &refusal) {
				t.Fatalf("type lie must stop before code generation: %v", failure)
			}
			t.Logf("pending .ts backend proof on 5f3b2e36; .a stops: %v", failure)
		})
	}
}
