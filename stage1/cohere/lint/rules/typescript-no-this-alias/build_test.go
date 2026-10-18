package assertions

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Measurement only (#mntyb57), never to land: each phase's seconds, cold then warm, reported by failing.
func TestCompileProfiles(t *testing.T) {
	entry, err := filepath.Abs("profile.a")
	if err != nil {
		t.Fatal(err)
	}
	var report []string
	for _, pass := range []string{"cold", "warm"} {
		directory := t.TempDir()
		started := time.Now()
		lap := func(name string) {
			report = append(report, fmt.Sprintf("%s %s %.2fs", pass, name, time.Since(started).Seconds()))
			started = time.Now()
		}
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		lap("load")
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			t.Fatal(err)
		}
		lap("lower")
		if err := native.Build(native.C(lowered), filepath.Join(directory, "native"), native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		lap("sanitized")
		if err := native.Build(native.C(lowered), filepath.Join(directory, "release"), native.Options{}); err != nil {
			t.Fatal(err)
		}
		lap("release")
		if err := os.WriteFile(filepath.Join(directory, "emitted.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
			t.Fatal(err)
		}
		lap("javascript")
	}
	t.Fatalf("PROFILE-TIMINGS %s", strings.Join(report, " | "))
}
