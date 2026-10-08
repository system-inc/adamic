package oracle

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"class_octal", "class_escapes", "class_string_duplicates", "quadratic_exec", "quadratic_matchall", "quadratic_test", "input_cache"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/regexp_native_" + name + ".a", true, false,
		})
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/regexp_native_refused/quantifier_bounds.a", false, false})
	// A refusal lives in a directory below testdata, which internal/flow (it walks only programs that
	// lower) doesn't glob.
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/regexp_native_refused/write_groups_compound_missing.a", false, false})
}

// Not parallel: these deadlines guard algorithmic complexity, not throughput under contention.
// Compile time is excluded. The unfixed duplicate-alternative probe takes about a minute.
func TestRegExpNativeTiming(t *testing.T) {
	for _, name := range []string{"class_string_duplicates", "quadratic_exec", "quadratic_matchall", "quadratic_test"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_native_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "probe")
			if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			best := 3 * time.Second
			for trial := 0; trial < 5; trial++ {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				start := time.Now()
				got, err := exec.CommandContext(ctx, binary).Output()
				elapsed := time.Since(start)
				cancel()
				if err != nil {
					t.Fatalf("native execution exceeded 3s or failed: %v (%s)", err, elapsed)
				}
				if want.exitCode != 0 || !bytes.Equal(got, want.stdout) {
					t.Fatalf("Node=%q native=%q", want.stdout, got)
				}
				t.Logf("trial %d native=%s", trial+1, elapsed)
				if elapsed < best {
					best = elapsed
				}
			}
			t.Logf("native best of 5=%s", best)
			best = 3 * time.Second
			for trial := 0; trial < 5; trial++ {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				start := time.Now()
				got, err := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path).Output()
				elapsed := time.Since(start)
				cancel()
				if err != nil || !bytes.Equal(got, want.stdout) {
					t.Fatalf("Node timing control: %v stdout=%q", err, got)
				}
				if elapsed < best {
					best = elapsed
				}
			}
			t.Logf("Node best of 5=%s (includes process startup and oracle loader)", best)
		})
	}
}
