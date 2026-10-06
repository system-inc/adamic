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
	for _, name := range []string{"class_octal", "class_escapes", "quantifier_bounds", "class_string_duplicates"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/regexp_native_" + name + ".a", true, false,
		})
	}
}

// Not parallel: these deadlines guard algorithmic complexity, not throughput under contention.
// Compile time is excluded. The unfixed duplicate-alternative probe takes about a minute.
func TestRegExpNativeTiming(t *testing.T) {
	for _, name := range []string{"class_string_duplicates"} {
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
			}
		})
	}
}
