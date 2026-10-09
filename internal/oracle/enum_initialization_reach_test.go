package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"native-enum-map", "native-call-before-enum", "safe-helper", "safe-constructor", "callable-selection", "call-graph-24", "unknown-after"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/fixtures/enum-init-reach/" + name + ".a", true, false})
	}
}

func TestEnumInitializationNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"native-enum-map", "native-call-before-enum", "safe-helper", "safe-constructor", "callable-selection", "call-graph-24", "unknown-after"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/enum-init-reach", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if d := disagreement(truth, got); d != "" {
					t.Fatalf("%s: %+v", d, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
	for _, name := range []string{"reaching-direct", "reaching-helper", "reaching-cycle", "reaching-constructor", "reaching-derived"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/enum-init-reach", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), "TypeError") {
				t.Fatalf("Node must fail at pending enum: %+v", truth)
			}
			_, err = lowered(t, path)
			var gap *lower.NotYet
			if !errors.As(err, &gap) || gap.What != "reading an enum before its runtime initialization; move the call after the enum declaration" || !strings.Contains(gap.Where, ":2:") {
				t.Fatalf("lost reaching-call refusal: %v", err)
			}
		})
	}
}

func TestEnumInitializationUnknownPinned(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unknown-before", "unknown-callback", "unknown-property"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/enum-init-reach", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 1 || string(truth.stdout) != "before\n" || !strings.Contains(string(truth.stderr), "TypeError") {
				t.Fatalf("source Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			// Source Node confirms stdout and exit 1; uncaught backend errors omit engine-specific stderr.
			want := run{stdout: truth.stdout, exitCode: truth.exitCode}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if d := disagreement(want, got); d != "" {
					t.Fatalf("%s: %+v", d, got)
				}
			}
		})
	}
}
