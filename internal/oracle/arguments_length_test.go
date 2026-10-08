package oracle

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

var argumentLengthFixtures = []string{
	"native-arguments-length-value.a", "arguments_length_no_reader.a", "arguments_length_unrelated_type.a",
	"arguments_length.a", "arguments_length_value.a", "arguments_length_value_count.a", "arguments_length_spread.a",
	"arguments_length_extended.a", "arguments_length_static_constructor.a",
	"arguments_length_method.a", "arguments_length_static.a", "arguments_length_reduce_left.a", "arguments_length_callbacks.a",
}

func init() {
	for _, name := range argumentLengthFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Both extensions must preserve the same behavior. The source is stored once as
// .a; the .ts spelling exists only in this test's temporary directory.
func TestArgumentsLengthTypeScriptSource(t *testing.T) {
	t.Parallel()
	for _, name := range argumentLengthFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			compiled, sanitized := natively(t, program)
			for label, run := range map[string]run{"native": compiled, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
				if difference := disagreement(truth, run); difference != "" {
					t.Errorf("%s: %s; Node %q, compiled %q", label, difference, truth.stdout, run.stdout)
				}
			}
			if leaked := leaks(t, program, sanitized); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
		})
	}
}
