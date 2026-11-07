package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var argumentLengthFixtures = []string{
	"native-arguments-length-value.a",
	"arguments_length_no_reader.a",
	"arguments_length_unrelated_type.a",
	"arguments_length.a",
	"arguments_length_value.a",
	"arguments_length_value_count.a",
	"arguments_length_spread.a",
	"arguments_length_extended.a",
	"arguments_length_static_constructor.a",
	"arguments_length_method.a",
	"arguments_length_static.a",
	"arguments_length_reduce_left.a",
	"arguments_length_callbacks.a",
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
func TestArgumentsLengthTypeScriptSource(test *testing.T) {
	test.Parallel()
	for _, name := range argumentLengthFixtures {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				test.Fatal(err)
			}
			path := filepath.Join(test.TempDir(), "main.ts")
			if err := os.WriteFile(path, source, 0644); err != nil {
				test.Fatal(err)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				test.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, path)
			compiled, sanitized := natively(test, program)
			for label, run := range map[string]run{
				"native":     compiled,
				"JavaScript": onJavaScriptBackend(test, program),
				"release":    released(test, program),
			} {
				if difference := disagreement(truth, run); difference != "" {
					test.Errorf("%s: %s; Node %q, compiled %q", label, difference, truth.stdout, run.stdout)
				}
			}
			if leaked := leaks(test, program, sanitized); leaked != "" {
				test.Errorf("leaks: %s", leaked)
			}
		})
	}
}

// A type-correct implementation can still read the wrong word. Only Node's
// value comparison catches this exact clean 9-versus-1 miscompile.
func TestArgumentsLengthWrongSlotMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join("testdata", "native-arguments-length-value.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	source := native.C(program)
	original := "= (double)argument_count;"
	if strings.Count(source, original) != 1 {
		test.Fatal("reader binding mutation site changed")
	}
	mutated := strings.Replace(source, original, "= arguments[0].number;", 1)
	for _, sanitize := range []bool{false, true} {
		name := "release"
		if sanitize {
			name = "sanitized"
		}
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			binary := filepath.Join(test.TempDir(), "mutant")
			if err := native.Build(mutated, binary, native.Options{Sanitize: sanitize}); err != nil {
				test.Fatal(err)
			}
			got := execute(test, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != "9\n" {
				test.Fatalf("wrong-slot mutant must run cleanly and print 9: %+v", got)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				test.Fatalf("Node failed to catch the wrong-slot mutant: %q", difference)
			}
			test.Logf("caught clean wrong-slot miscompile: Node %q, native %q", truth.stdout, got.stdout)
		})
	}
}
