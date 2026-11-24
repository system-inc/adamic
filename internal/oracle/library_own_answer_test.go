package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An explicit toString or toLocaleString on a library object whose own answer isn't Object.prototype's stops at
// compile time (#20e4t06, Compiler's two-hour stop). On main each lowered to the constant "[object Object]" tag
// (prototype.go's objectPrototypeCallArguments) for a Map or Set iterator, where Node prints [object Map Iterator] or
// [object Set Iterator]. The template and String() roads already stop on main, before ToPrimitive (J-d27ejbm). The
// real fix, each object's own tag, follows with Runtime Library (#eksbz4c's head T).
//
// A RegExp match's groups object has a null prototype, so on Node a direct call of toString, toLocaleString,
// valueOf, hasOwnProperty or propertyIsEnumerable on one throws a TypeError naming the callee as written (exit 70
// uncaught), where a program's own object of the groups type answers as Object.prototype does (#shqn1d0's stop,
// lifted by #eksbz4c's head G). Each such call tests its receiver at run time, by the shape's sentinel natively
// (adamic_groups_methods) and by a null prototype on the JavaScript backend.
const libraryOwnAnswerDirectory = "internal/oracle/testdata/library_own_answer/"

func libraryOwnAnswerFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, libraryOwnAnswerDirectory+name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Each line main printed wrong stops at its own call, as NotYet naming the receiver, the Object.prototype.toString.call
// road's iterator row (k1, Compiler's ask) among them. Mutant: the stop dropped
// (scratch-ts), which lowers every row again.
func TestLibraryObjectOwnAnswerStops(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, position, receiver string }{
		{"i1_entries.a", "2:13", "a library "},
		{"i2_keys_locale.a", "2:13", "a library "},
		{"i3_set_values.a", "2:13", "a library "},
		{"k1_call_iterator.a", "2:13", "a library "},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, libraryOwnAnswerFixture(t, stop.name))
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "whose own ") || !strings.Contains(err.Error(), stop.receiver) || !strings.Contains(err.Error(), stop.name+":"+stop.position) {
				t.Fatalf("want a NotYet at %s:%s naming %s, got %v", stop.name, stop.position, stop.receiver, err)
			}
		})
	}
}

// What keeps its answer and still agrees with Node natively, sanitized and on JavaScript: an iterator result,
// which next makes as a plain object; Object.prototype.toString.call of a RegExp match's groups (Object's own tag,
// prototype or none); a Map's own tag and an array's join; a program's object literal and class instance,
// Object.prototype's own; a program's own object of the groups type (h1, h3), which the run-time test lets
// answer as Object.prototype does; and a groups TypeError caught, plainly (g9) and as an Error whose message prints
// Node's text exactly (g10). Mutant (b), the test always true, throws on h1 and h3.
func TestLibraryObjectOwnAnswerControlsAgree(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"i4_result.a", "c3_map_array.a", "c1_control.a", "k2_call_groups.a", "h1_own.a", "h3_own_has.a", "g9_caught.a", "g10_caught_message.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, libraryOwnAnswerDirectory+name)
		})
	}
}

// Each direct call of an Object.prototype method on a real groups object throws Node's TypeError, uncaught, after
// what the program printed before it, natively, sanitized and on the JavaScript backend: toString (g1, and g5 through
// a local), toLocaleString (g6), valueOf (g4), hasOwnProperty (g7) and propertyIsEnumerable (g8, through a local),
// where main stopped at compile time or panicked natively; and h2, one parameter of the groups type that sees a
// program's own object first, which answers, and then real groups, which throw. Mutants: the run-time test always
// false (a), and internal/regexp/native.go writing adamic_host_methods again (d), each of which answers instead.
func TestRegExpGroupsMissingMethodsThrowAsNode(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, message string }{
		{"g1_tostring.a", "TypeError: match.groups.toString is not a function"},
		{"g4_valueof.a", "TypeError: match.groups.valueOf is not a function"},
		{"g5_local.a", "TypeError: groups.toString is not a function"},
		{"g6_locale.a", "TypeError: match.groups.toLocaleString is not a function"},
		{"g7_has.a", "TypeError: match.groups.hasOwnProperty is not a function"},
		{"g8_enumerable.a", "TypeError: groups.propertyIsEnumerable is not a function"},
		{"h2_param.a", "TypeError: groups.toString is not a function"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			libraryOwnAnswerHeldToNode(t, row.name, 70, row.message)
		})
	}
}

// libraryOwnAnswerHeldToNode runs a fixture on Node, natively, sanitized and on the JavaScript backend, and each
// backend must give Node's stdout, stderr and exit. Node's own run must exit as exit says with message in its stderr,
// so a fixture that stops throwing on Node fails here rather than passing as one that agrees.
func libraryOwnAnswerHeldToNode(t *testing.T, name string, exit int, message string) {
	t.Helper()
	path := libraryOwnAnswerFixture(t, name)
	want := onNode(t, path)
	if want.exitCode != exit || !strings.Contains(string(want.stderr), message) {
		t.Fatalf("Node: want exit %d with %q, got exit %d stdout %q stderr %q", exit, message, want.exitCode, want.stdout, want.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"native", "sanitized", "javascript"} {
		var got run
		switch backend {
		case "native":
			got = releasedUncached(t, program)
		case "sanitized":
			got, _ = nativelyUncached(t, program)
		case "javascript":
			got = onJavaScriptBackend(t, program)
		}
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stdout %q stderr %q", backend, diff, got.stdout, got.stderr)
		}
	}
}
