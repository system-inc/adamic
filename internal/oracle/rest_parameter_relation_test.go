package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A method seen as one whose parameters are a rest tuple is related position by position, each of its
// parameters against the tuple's element there, never the tuple's whole type (#6xzs6gj). Before, the
// walk compared a listener's first parameter with `[error: Error]` itself, so tsc's typed event overload,
// `on<E extends keyof T>(eventName: E, listener: (...args: T[E]) => void)`, falsely refused
// `(seen: Error) => ...`. r01, a class method taking Error seen through an interface method taking
// `...args: [error: Error]`, and r02, two parameters against a two-element tuple, lower and agree with
// Node on release, sanitized native and the JavaScript backend. n01, a method needing a field its
// element doesn't carry (`{ text; code }` seen as taking `{ text }`), which tsc accepts by method
// bivariance and Node runs printing undefined, is still refused, now at the element. Each
// program's whole stop is pinned in its .stop file, or "lowers". Mutant: the walk relating a rest
// parameter's whole type again, and r01 and r02 stop.
func TestRestParameterRelation(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/rest_parameter_relation"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			expected, readError := os.ReadFile(absolute + ".stop")
			if readError != nil {
				test.Fatal(readError)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, loweredError := lowered(test, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if loweredError == nil {
					test.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(
					loweredError.Error(),
					strings.TrimSuffix(absolute, path),
					"",
				); got != stop {
					test.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if loweredError != nil {
				test.Fatalf("Lower: %v", loweredError)
			}
			sanitized, _ := natively(test, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"release", released(test, program)},
				{"sanitized", sanitized},
				{"JavaScript", onJavaScriptBackend(test, program)},
			} {
				if difference := disagreement(truth, backend.got); difference != "" {
					test.Fatalf(
						"%s: %s: Node %q, %s %q",
						backend.name,
						difference,
						truth.stdout,
						backend.name,
						backend.got.stdout,
					)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 6 {
		test.Fatalf("found %d programs, want 6", programs)
	}
}
