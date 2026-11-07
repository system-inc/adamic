package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// throw new TypeError(message) lowers as new Error(message) does, with "TypeError" in the error's name slot
// (Adamic's ruling on #axochk, Nov 1: an Error subclass is admitted, and docs/0.1.md's exceptions row, ruled
// October 19, makes any other thrown value NotYet, never Refused). Caught, its name and message are Node's;
// uncaught, its panic line is. Both are held to Node natively, sanitized and released, and on the JavaScript
// backend, and so are RangeError, ReferenceError and SyntaxError, the other built-in Error classes the
// runtime names. The fixtures sit in throw_type_error, since the flow and fresh corpora lower every program
// at the top of testdata. Mutant: throwStatement refusing a TypeError again, as it did ("throwing a
// TypeError"); with it applied, this test goes red at lowering.
func TestThrownErrorClassesAgreeWithNode(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, stdout, stderr string }{
		{"caught.a", caughtTypeError, ""},
		{"caught.ts", caughtTypeError, ""},
		{"uncaught.a", "before\n", "adamic: panic: TypeError: x\n"},
		{"uncaught.ts", "before\n", "adamic: panic: TypeError: x\n"},
		{"value.a", "TypeError\nvalue\n", ""},
		{"caught_classes.a", "RangeError: range 10\nReferenceError: reference 11\nSyntaxError: syntax 12\n", ""},
		{"uncaught_range_error.a", "before\n", "adamic: panic: RangeError: x\n"},
		{"uncaught_reference_error.a", "before\n", "adamic: panic: ReferenceError: x\n"},
		{"uncaught_syntax_error.a", "before\n", "adamic: panic: SyntaxError: x\n"},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/throw_type_error", probe.file),
			)
			if err != nil {
				test.Fatal(err)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte(probe.stderr)}
			if probe.stderr != "" {
				want.exitCode = 70
			}
			truth := onNode(test, path)
			if disagreement(want, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatalf("a thrown built-in Error must lower: %v", err)
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

// A caught value tested against TypeError stays where main stops it: instanceof against a library value
// isn't lowered, so a caught TypeError isn't read as one, and the test isn't lowered to true.
func TestCaughtInstanceOfTypeErrorNotYet(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/throw_type_error/instanceof.a"))
	if err != nil {
		test.Fatal(err)
	}
	if truth := onNode(test, path); disagreement(run{stdout: []byte("true\n")}, truth) != "" {
		test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err = lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) ||
		!strings.Contains(err.Error(), "instanceof.a:6:17") ||
		!strings.Contains(err.Error(), "instanceof against a value that isn't a declared class") {
		test.Fatalf("want instanceof TypeError NotYet at the test, got %v", err)
	}
}

// Object.keys of an Error is none on Node, since its own properties aren't enumerable, but natively its
// message and name are fields, and the JavaScript backend's named error has its name enumerable. On main a
// caught Error's keys printed 2 natively (keys_caught.a); a made TypeError's would have too, and 1 on the
// JavaScript backend (keys.a). Each is NotYet at the call, by errorMayHide, and so is a made TypeError seen
// through a structural view in a program with no catch (keys_view.a), which only holds while errorMayHide
// reads every member of errorClasses as made in the source.
func TestErrorKeysNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, at string }{
		{"keys.a", "keys.a:4:16"},
		{"keys_caught.a", "keys_caught.a:7:18"},
		{"keys_view.a", "keys_view.a:5:17"},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/throw_type_error", probe.file),
			)
			if err != nil {
				test.Fatal(err)
			}
			if truth := onNode(test, path); disagreement(run{stdout: []byte("0\n")}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(test, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) ||
				!strings.Contains(err.Error(), probe.at) ||
				!strings.Contains(err.Error(), "Object.keys of a value that may hold an Error") {
				test.Fatalf("want Object.keys of a value that may hold an Error NotYet at %s, got %v", probe.at, err)
			}
		})
	}
}

// propertyOf, exactly as adaptation 94 (compiler/axolotl-94 d42771c3) appends it to adaptation 47's
// src/compiler/hostErrors.ts, lowered as the census lowers a root. On main it was refused at its throw
// (hostErrors.ts:14:15, "throwing a TypeError"). Now the throw lowers, and the first stop is inside its
// message, String(value) of a value that is null or undefined there, which isn't built yet.
func TestPropertyOfLowersPastItsThrow(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/throw_type_error/hostErrors.ts"))
	if err != nil {
		test.Fatal(err)
	}
	_, err = lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) ||
		!strings.Contains(err.Error(), "hostErrors.ts:14:65") ||
		!strings.Contains(err.Error(), "String conversion of an object, array, map or function") {
		test.Fatalf("want propertyOf NotYet at String(value) in its message, got %v", err)
	}
}

// caughtTypeError is what caught.a and caught.ts print on Node.
const caughtTypeError = "TypeError\nbad 3\nTypeError: bad 3\nTypeError\ndirect 4\nTypeError stored\n"

// No Error subclass lowers today: class Bad extends Error stops at its base, "a computed class base", before any
// road reads it (Lower's probe on #bbhwj30: 54 programs, keys, values, entries, propertyIsEnumerable and message,
// directly and through a view, all stopping there). errorMayHide's made test knows the library's Error classes
// by name and walks no base chain, so a made Bad seen as { message: string } with no catch in the program would
// slip past it and print 2 natively where Node prints 0. This program is the tripwire: whichever change lets a
// class extend Error goes red here, and must teach errorMayHide the base chain before it re-pins this.
func TestErrorSubclassStaysNotYet(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/throw_type_error/subclass_keys_view.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	if truth := onNode(test, path); disagreement(run{stdout: []byte("0\n")}, truth) != "" {
		test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err = lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) ||
		!strings.Contains(err.Error(), "subclass_keys_view.a:3:11") ||
		!strings.Contains(err.Error(), "a computed class base") {
		test.Fatalf(
			"an Error subclass no longer stops at its base (got %v): before re-pinning this, errorMayHide's made test (prototype.go) must walk the base chain, so an instance of a class extending any errorClasses member counts as an Error made in the source",
			err,
		)
	}
}
