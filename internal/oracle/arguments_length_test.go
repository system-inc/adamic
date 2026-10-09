package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
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

// A type-correct implementation can still read the wrong word. Only Node's
// value comparison catches this exact clean 9-versus-1 miscompile.
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and adamic/runtime directories.
func TestArgumentsLengthWrongSlotMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join("testdata", "native-arguments-length-value.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	source := native.C(program)
	original := "= (double)argument_count;"
	if strings.Count(source, original) != 1 {
		t.Fatal("reader binding mutation site changed")
	}
	mutated := strings.Replace(source, original, "= arguments[0].number;", 1)
	for _, sanitize := range []bool{false, true} {
		name := "release"
		if sanitize {
			name = "sanitized"
		}
		// Not parallel: native.Build writes the shared os.UserCacheDir()/adamic/runtime directory.
		t.Run(name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutated, binary, native.Options{Sanitize: sanitize}); err != nil {
				t.Fatal(err)
			}
			got := execute(t, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != "9\n" {
				t.Fatalf("wrong-slot mutant must run cleanly and print 9: %+v", got)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("Node failed to catch the wrong-slot mutant: %q", difference)
			}
			t.Logf("caught clean wrong-slot miscompile: Node %q, native %q", truth.stdout, got.stdout)
		})
	}
}
