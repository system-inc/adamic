package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"903f25b_private_mangled", "8e67677_null_undefined_inline"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/silent_reflection_nullish/" + name + ".a", false, false})
	}
}

// These exact integration witnesses may stop only at their named unsupported
// boundary. If a revert admits either one, execute all modes so Node catches
// the original exit-zero wrong answer, rather than only the missing diagnostic.
func TestSilentReflectionNullish(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ name, stdout, reason string }{
		{"903f25b_private_mangled", "false false true\n", "private class slots are not JavaScript own properties"},
		{"8e67677_null_undefined_inline", "absent: undefined=true null=false\nmiss: undefined=false null=true\npast end: undefined=true null=false\nstored: undefined=false null=true\n", "null and undefined comparison without a tagged reference slot"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/silent_reflection_nullish", fixture.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != fixture.stdout {
				t.Fatalf("source Node: exit %d stdout %q stderr %q", want.exitCode, want.stdout, want.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.HasPrefix(err.Error(), path+":") || !strings.Contains(err.Error(), fixture.reason) {
					t.Fatalf("want the named boundary %q with its path, got %v", fixture.reason, err)
				}
				t.Logf("Node %q; stopped: %v", want.stdout, err)
				return
			}
			sanitized, binary := natively(t, program)
			for _, mode := range []struct {
				name string
				got  run
			}{{"JavaScript", onJavaScriptBackend(t, program)}, {"release", released(t, program)}, {"sanitized", sanitized}} {
				t.Logf("%s: exit %d stdout %q stderr %q; Node %q", mode.name, mode.got.exitCode, mode.got.stdout, mode.got.stderr, want.stdout)
				if difference := disagreement(want, mode.got); difference != "" {
					t.Errorf("%s: %s", mode.name, difference)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Errorf("leaks: %s", report)
			}
		})
	}
}
