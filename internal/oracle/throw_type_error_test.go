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
func TestThrownErrorClassesAgreeWithNode(t *testing.T) {
	t.Parallel()
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
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/throw_type_error", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte(probe.stderr)}
			if probe.stderr != "" {
				want.exitCode = 70
			}
			truth := onNode(t, path)
			if disagreement(want, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a thrown built-in Error must lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// A caught value tested against TypeError stays where main stops it: instanceof against a library value
// isn't lowered, so a caught TypeError isn't read as one, and the test isn't lowered to true.
func TestCaughtInstanceOfTypeErrorNotYet(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/throw_type_error/instanceof.a"))
	if err != nil {
		t.Fatal(err)
	}
	if truth := onNode(t, path); disagreement(run{stdout: []byte("true\n")}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "instanceof.a:6:17") || !strings.Contains(err.Error(), "instanceof against a value that isn't a declared class") {
		t.Fatalf("want instanceof TypeError NotYet at the test, got %v", err)
	}
}

// Object.keys of an Error is none on Node, since its own properties aren't enumerable, but natively its
// message and name are fields, and the JavaScript backend's named error has its name enumerable. On main a
// caught Error's keys printed 2 natively (keys_caught.a); a made TypeError's would have too, and 1 on the
// JavaScript backend (keys.a). Both are NotYet at the call.
func TestErrorKeysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at string }{
		{"keys.a", "keys.a:4:16"},
		{"keys_caught.a", "keys_caught.a:7:18"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/throw_type_error", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(run{stdout: []byte("0\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), "Object.keys of an Error") {
				t.Fatalf("want Object.keys of an Error NotYet at %s, got %v", probe.at, err)
			}
		})
	}
}

// propertyOf, exactly as adaptation 94 (compiler/axolotl-94 d42771c3) appends it to adaptation 47's
// src/compiler/hostErrors.ts, lowered as the census lowers a root. On main it was refused at its throw
// (hostErrors.ts:14:15, "throwing a TypeError"). Now the throw lowers, and the first stop is inside its
// message, String(value) of a value that is null or undefined there, which isn't built yet.
func TestPropertyOfLowersPastItsThrow(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/throw_type_error/hostErrors.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "hostErrors.ts:14:65") || !strings.Contains(err.Error(), "String conversion of an object, array, map or function") {
		t.Fatalf("want propertyOf NotYet at String(value) in its message, got %v", err)
	}
}

// caughtTypeError is what caught.a and caught.ts print on Node.
const caughtTypeError = "TypeError\nbad 3\nTypeError: bad 3\nTypeError\ndirect 4\nTypeError stored\n"
