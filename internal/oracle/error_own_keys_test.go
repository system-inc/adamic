package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Node gives an Error no enumerable own properties: its message is own and not enumerable, and its name is
// its prototype's. Natively both are fields like any other's, so on main Object.keys of a caught Error, and
// Object.keys, values and entries of one through a { message: string } view, each printed 2 where Node
// prints 0, in .a and .ts alike (#bbhwj30, Lower's 10:21Z comment). errorMayHide (prototype.go) now decides
// whether an argument may hold an Error, the test prototypeHazard's Error case makes too, and every such
// road is NotYet. It also fires for an Error the runtime made, from a failing host call, which has no
// new Error node to find: whenever the library's Error fits the argument's type and the program catches.
// The fixtures sit in error_own_keys, since the flow and fresh corpora lower every program at the top of
// testdata. Mutants: errorMayHide dropped from the keys, values and entries paths turns the view and
// runtime rows red; its catch extension dropped turns the runtime rows red.
func TestErrorOwnKeysNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		file, words  string
		madeInSource bool
	}{
		{"keys_caught.a", "Object.keys of a value that may hold an Error", true},
		{"keys_caught.ts", "Object.keys of a value that may hold an Error", true},
		{"keys_view.a", "Object.keys of a value that may hold an Error", true},
		{"keys_view.ts", "Object.keys of a value that may hold an Error", true},
		{"values_view.a", "Object.values of a value that may hold an Error", true},
		{"values_view.ts", "Object.values of a value that may hold an Error", true},
		{"entries_view.a", "Object.entries of a value that may hold an Error", true},
		{"entries_view.ts", "Object.entries of a value that may hold an Error", true},
		{"keys_runtime.a", "Object.keys of a value that may hold an Error", false},
		{"keys_runtime.ts", "Object.keys of a value that may hold an Error", false},
		{"values_runtime.a", "Object.values of a value that may hold an Error", false},
		{"values_runtime.ts", "Object.values of a value that may hold an Error", false},
		{"entries_runtime.a", "Object.entries of a value that may hold an Error", false},
		{"entries_runtime.ts", "Object.entries of a value that may hold an Error", false},
		{"has_own_property_runtime.a", "an Error may be hidden by the view", false},
		{"property_is_enumerable_runtime.a", "an Error may be hidden by the view", false},
		{"detached_has_own_runtime.a", "an Error may be hidden by the view", false},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/error_own_keys", probe.file),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			// A host call's Error carries enumerable fields of its own (code, errno and the like), so
			// only an Error made in the source has Node's count pinned.
			if probe.madeInSource && disagreement(run{stdout: []byte("0\n")}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, loweredError := lowered(test, path)
			var notYet *lower.NotYet
			if !errors.As(loweredError, &notYet) ||
				!strings.Contains(loweredError.Error(), probe.file) ||
				!strings.Contains(loweredError.Error(), probe.words) {
				test.Fatalf("want NotYet with %q, got %v", probe.words, loweredError)
			}
		})
	}
}

// Every other road that reads an Error's own keys or asks whether one exists, through the same view, stops
// at compile time already: getOwnPropertyNames, the descriptors, fromEntries, hasOwn, Reflect's, for...in,
// an object rest, JSON.stringify, a spread and Object.assign. This holds them there.
func TestErrorOwnPropertyRoadsStop(test *testing.T) {
	test.Parallel()
	for _, file := range []string{
		"get_own_property_names_view.a",
		"get_own_property_descriptor_view.a",
		"get_own_property_descriptors_view.a",
		"from_entries_view.a",
		"has_own_view.a",
		"reflect_own_keys_view.a",
		"reflect_has_view.a",
		"for_in_view.a",
		"object_rest_view.a",
		"json_view.a",
		"spread_view.a",
		"assign_view.a",
	} {
		test.Run(file, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/error_own_keys", file),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if truth := onNode(test, path); truth.exitCode != 0 {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			if _, loweredError := lowered(test, path); loweredError == nil {
				test.Fatal("an own-property road through a view that may hold an Error lowered")
			} else {
				test.Logf("stops: %v", loweredError)
			}
		})
	}
}

// The same roads on a plain object through the same view, in a program with no Error and no catch, still
// lower and agree with Node natively, sanitized and released, and on the JavaScript backend. So does
// Object.keys of a const bound to a literal in a program that catches, since a literal is never an Error.
func TestErrorOwnKeysPlainTwinsAgree(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"twin_keys.a", "1\nmessage\n"},
		{"twin_keys.ts", "1\nmessage\n"},
		{"twin_values.a", "x\n"},
		{"twin_values.ts", "x\n"},
		{"twin_entries.a", "message:x\n"},
		{"twin_entries.ts", "message:x\n"},
		{"twin_own.a", "true true\n"},
		{"twin_literal_with_catch.a", "message\nm\n"},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/error_own_keys", probe.file),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if disagreement(run{stdout: []byte(probe.stdout)}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("a plain object's own keys must lower: %v", loweredError)
			}
			native, _ := natively(test, program)
			for name, got := range map[string]run{
				"native":     native,
				"release":    released(test, program),
				"javascript": onJavaScriptBackend(test, program),
			} {
				if problem := disagreement(truth, got); problem != "" {
					test.Errorf(
						"%s: %s: exit %d, stdout %q, stderr %q",
						name,
						problem,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
		})
	}
}
