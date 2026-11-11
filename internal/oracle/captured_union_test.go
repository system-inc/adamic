package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// capturedUnionPins are #n7sx9gz's programs: a let whose members are held differently (an object, a
// string literal, undefined), captured by the closures that write and read it. Each is held to Node,
// and to the pinned output too, so an agreement on the wrong text can't pass.
var capturedUnionPins = []struct{ name, want string }{
	{"class", "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"},
	{"object", "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"},
	{"by_value", "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"},
	{"released", "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"},
	{"nested", "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"},
}

// Each pin is committed as .ts and run as .a too, from the same text, so both dialects are held to
// one program. Every build is asked whether it let go of everything: the sanitized one by the shared
// leak check, and the counted one by its own counts, where an overwritten member never released reads
// as a value leaked.
func TestCapturedUnionAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range capturedUnionPins {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(pin.name+extension, func(t *testing.T) {
				t.Parallel()
				source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/captured_union", pin.name+".ts"))
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), pin.name+extension)
				if err := os.WriteFile(path, source, 0o644); err != nil {
					t.Fatal(err)
				}
				want := onNode(t, path)
				if want.exitCode != 0 || string(want.stdout) != pin.want {
					t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", want.exitCode, want.stdout, pin.want, want.stderr)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				t.Run("native", func(t *testing.T) {
					got := releasedUncached(t, program)
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
					}
				})
				t.Run("sanitized", func(t *testing.T) {
					got, leaked, comparable := nativelyCheckingLeaks(t, program, true)
					if leaked != "" {
						t.Errorf("leaks: %s", leaked)
					}
					if !comparable {
						return
					}
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
					}
				})
				t.Run("counted", func(t *testing.T) {
					binary := filepath.Join(t.TempDir(), "counted")
					if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
						t.Fatal(err)
					}
					got := execute(t, binary)
					if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
						t.Errorf("counts: %s", report)
					}
					if string(got.stdout) != pin.want {
						t.Errorf("got stdout %q", got.stdout)
					}
				})
				t.Run("javascript", func(t *testing.T) {
					got := onJavaScriptBackend(t, program)
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
					}
				})
			})
		}
	}
}
