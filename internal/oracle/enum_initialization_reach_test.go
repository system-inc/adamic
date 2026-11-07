package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{
		"native-enum-map",
		"native-call-before-enum",
		"safe-helper",
		"safe-constructor",
		"callable-selection",
		"call-graph-24",
		"unknown-after",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/fixtures/enum-init-reach/" + name + ".a", true, false})
	}
}

func TestEnumInitializationNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"native-enum-map",
		"native-call-before-enum",
		"safe-helper",
		"safe-constructor",
		"callable-selection",
		"call-graph-24",
		"unknown-after",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "stage3/fixtures/enum-init-reach", name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onNode(test, path)
			if truth.exitCode != 0 {
				test.Fatalf("Node: %+v", truth)
			}
			sanitized, binary := nativelyUncached(test, program)
			for _, got := range []run{sanitized, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if d := disagreement(truth, got); d != "" {
					test.Fatalf("%s: %+v", d, got)
				}
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
	for _, name := range []string{
		"reaching-direct",
		"reaching-helper",
		"reaching-cycle",
		"reaching-constructor",
		"reaching-derived",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "stage3/fixtures/enum-init-reach", name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), "TypeError") {
				test.Fatalf("Node must fail at pending enum: %+v", truth)
			}
			_, loweredError := lowered(test, path)
			var gap *lower.NotYet
			if !errors.As(loweredError, &gap) ||
				gap.What != "reading an enum before its runtime initialization; move the call after the enum declaration" ||
				!strings.Contains(gap.Where, ":2:") {
				test.Fatalf("lost reaching-call refusal: %v", loweredError)
			}
		})
	}
}

func TestEnumInitializationUnknownPinned(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"unknown-before", "unknown-callback", "unknown-property"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "stage3/fixtures/enum-init-reach", name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if truth.exitCode == 0 ||
				string(truth.stdout) != "before\n" ||
				!strings.Contains(string(truth.stderr), "TypeError") {
				test.Fatalf("source Node: %+v", truth)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want := run{
				stdout:   []byte("before\n"),
				stderr:   []byte("adamic: panic: ReferenceError: Cannot access 'Pending' before initialization\n"),
				exitCode: 70,
			}
			sanitized, _ := nativelyUncached(test, program)
			for _, got := range []run{sanitized, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if d := disagreement(want, got); d != "" {
					test.Fatalf("%s: %+v", d, got)
				}
			}
		})
	}
}
