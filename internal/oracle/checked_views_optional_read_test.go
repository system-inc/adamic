package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// V1 supports ordinary optional reads but not optional checked reads. Keep the
// source observation and both backend entry points pinned independently.
func TestCheckedViewOptionalReadBoundary(t *testing.T) {
	t.Parallel()
	checked, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane1/optional-read.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("name\nmissing\n")}
	if difference := disagreement(want, onNode(t, checked)); difference != "" {
		t.Fatal("checked source Node: " + difference)
	}
	t.Run("ordinary control", func(t *testing.T) {
		t.Parallel()
		program, path := interfaceFixture(t, "lane1/optional-read-control")
		if difference := disagreement(want, onNode(t, path)); difference != "" {
			t.Fatal("control source Node: " + difference)
		}
		sanitized, binary := nativelyUncached(t, program)
		for name, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(want, got); difference != "" {
				t.Errorf("%s: %s", name, difference)
			}
		}
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	})
	for _, backend := range []string{"native", "JavaScript"} {
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if value := recover(); value != nil {
					t.Errorf("compiler panic instead of the pinned NotYet boundary: %v", value)
				}
			}()
			program, err := lowered(t, checked)
			if err == nil {
				if backend == "native" {
					native.C(program)
				} else {
					javascript.JavaScript(program)
				}
				t.Fatal("optional checked read escaped the V1 NotYet boundary")
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want optional checked-read NotYet, got %v", err)
			}
			expected := checked + ":9:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet"
			if strings.TrimSpace(err.Error()) != expected {
				t.Fatalf("boundary diagnostic: got %q, want %q", err, expected)
			}
		})
	}
}
