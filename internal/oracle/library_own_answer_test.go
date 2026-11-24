package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An explicit toString or toLocaleString on a library object whose own answer isn't Object.prototype's stopped at
// compile time (#20e4t06, Compiler's two-hour stop): each had lowered to the constant "[object Object]" tag
// (prototype.go's objectPrototypeCallArguments) for a Map or Set iterator, where Node prints [object Map Iterator] or
// [object Set Iterator]. Now a library object with a tag of its own that a program can hold (ir.LibraryObjectTags)
// answers the tag its value has, read at run time, never the one its type names (#eksbz4c's head T): natively by its
// shape's address or its kind (adamic_object_tag), and on the JavaScript backend by Object.prototype.toString.call of
// the value, whose iterator wrappers carry their tag. Any other library interface still stops. The template and
// String() roads stop before ToPrimitive (J-d27ejbm), as they did.
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

// A library object whose tag lowering doesn't read stops at its call, as NotYet naming the receiver: b1, a value
// typed as an ArrayIterator, which no program makes yet, though Node prints [object Map Iterator] for the Map's
// iterator it holds, so its type names no tag the value has.
func TestLibraryObjectOwnAnswerStops(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, position, receiver string }{
		{"b1_array_typed.a", "3:13", "a library ArrayIterator"},
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

// Each library object with a tag of its own answers it, natively, sanitized and on the JavaScript backend, and Node's
// own run prints the tag, so a row can't pass with both printing [object Object]: a Map's and a Set's iterators
// directly, through toLocaleString and by Object.prototype.toString.call (i1, i2, i3, k1, which stopped at compile time
// before), a WeakMap, a WeakSet and a Promise by each road (w1, w2, p1), a matchAll iterator by each (m1), and b2, a
// program's own class typed as a MapIterator, whose value has no tag, so it answers [object Object] beside Node. Mutants:
// a tag changed in adamic_object_tag (a), the Map's and the Set's iterator shapes swapped in its answer (b), and the
// JavaScript wrappers' [Symbol.toStringTag] dropped (c).
func TestLibraryObjectsAnswerTheirOwnTag(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, tag string }{
		{"i1_entries.a", "[object Map Iterator]"},
		{"i2_keys_locale.a", "[object Map Iterator]"},
		{"i3_set_values.a", "[object Set Iterator]"},
		{"k1_call_iterator.a", "[object Map Iterator]"},
		{"w1_weak_map.a", "[object WeakMap]"},
		{"w2_weak_set.a", "[object WeakSet]"},
		{"p1_promise.a", "[object Promise]"},
		{"m1_match_all.a", "[object RegExp String Iterator]"},
		{"b2_own_typed.a", "[object Object]"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			libraryOwnTagAgrees(t, row.name, row.tag)
		})
	}
}

// A Map or Set iterator held by a program's own view, an interface or a type literal with next, answers its own tag
// through the view's toString, toLocaleString or Object.prototype.toString.call, natively, sanitized and on the
// JavaScript backend (#v414yhr's stop, lifted by #eksbz4c's head T): before #v414yhr each printed [object Object], and
// then each stopped at compile time, since the view is the program's own and its type names no tag. v7 is one
// parameter that sees a Map's iterator, a Set's and a program's own object in turn, each answering its own. Mutant (d),
// prototypeHazard's program-wide view stop restored for these iterators, stops every row again.
func TestLibraryObjectViewsAnswerTheValuesTag(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, tag string }{
		{"v1_interface_map_tostring.a", "[object Map Iterator]"},
		{"v2_interface_set_locale.a", "[object Set Iterator]"},
		{"v3_interface_map_call.a", "[object Map Iterator]"},
		{"v4_literal_set_tostring.a", "[object Set Iterator]"},
		{"v5_literal_map_locale.a", "[object Map Iterator]"},
		{"v6_literal_set_call.a", "[object Set Iterator]"},
		{"v7_interface_both.a", "[object Set Iterator]"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			libraryOwnTagAgrees(t, row.name, row.tag)
		})
	}
}

// libraryOwnTagAgrees holds a fixture to Node natively, sanitized and on the JavaScript backend, after checking that
// Node's own run prints tag.
func libraryOwnTagAgrees(t *testing.T, name, tag string) {
	t.Helper()
	if truth := onNode(t, libraryOwnAnswerFixture(t, name)); !strings.Contains(string(truth.stdout), tag+"\n") {
		t.Fatalf("Node: want %s printed, got exit %d stdout %q stderr %q", tag, truth.exitCode, truth.stdout, truth.stderr)
	}
	loomViewsDAgrees(t, libraryOwnAnswerDirectory+name)
}

// prototypeHazard's program-wide view stop keeps counting every value libraryObjectOwnAnswer names whose answer
// isn't read at run time, and drops only the tagged ones (a deliberate narrowing of Runtime's should on #eksbz4c, which
// asked for the stop to come out whole): a Date's toString on Node is its own date text, never a tag, so a Date behind
// a program's own view would print [object Object] without it. No Date view lowers today, since a view naming any
// Date method stops at the view first (d1, J-9jt9xde's probes), so d1 pins that earlier stop, and the day it lifts d1
// fails here and shows the arm's Date case. s1, a Stats behind a view of its size, is the arm's one live witness today:
// it stops there (Node prints [object Object], so the arm is conservative for it). Mutant (e), the stop dropped
// whole, lowers s1.
func TestLibraryObjectViewStops(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, position, words string }{
		{"d1_date_view.a", "4:22", "a view naming getTime, which lib declares on Date"},
		{"s1_stats_view.a", "6:13", "a library Stats may be hidden by the view"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, libraryOwnAnswerFixture(t, stop.name))
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), stop.words) || !strings.Contains(err.Error(), stop.name+":"+stop.position) {
				t.Fatalf("want a NotYet at %s:%s naming %q, got %v", stop.name, stop.position, stop.words, err)
			}
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

// What a view holds that isn't a library object with its own answer keeps lowering and agrees with Node natively,
// sanitized and on JavaScript: a program's own object literal and class instance behind the same interface, an
// iterator result behind a view of its done, and a program that makes a Map iterator its view can't hold.
func TestLibraryObjectViewControlsAgree(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"c4_view_own.a", "c5_view_class.a", "c6_view_result.a", "c7_view_unrelated.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, libraryOwnAnswerDirectory+name)
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
