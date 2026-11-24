package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An explicit toString, toLocaleString or valueOf on a library object whose own answer isn't Object.prototype's
// stops at compile time (#20e4t06, #shqn1d0, Compiler's two-hour stop). On main each lowered to the constant
// "[object Object]" tag (prototype.go's objectPrototypeCallArguments): a Map or Set iterator, where Node prints
// [object Map Iterator] or [object Set Iterator], and a RegExp match's groups, which have a null prototype, so
// Node throws a TypeError at the call (exit 70), and valueOf answered the groups themselves. The template and
// String() roads already stop on main, before ToPrimitive (J-d27ejbm). The real fix, each object's own tag and
// groups' missing prototype, follows with Runtime Library.
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
		{"g1_tostring.a", "3:14", "a RegExp match's groups"},
		{"g4_valueof.a", "3:21", "a RegExp match's groups"},
		{"g5_local.a", "4:14", "a RegExp match's groups"},
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
// Object.prototype.toString.call of a RegExp match's groups (Object's own tag, prototype or none),
// which next makes as a plain object; a Map's own tag and an array's join; and a program's object literal and
// class instance, Object.prototype's own.
func TestLibraryObjectOwnAnswerControlsAgree(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"i4_result.a", "c3_map_array.a", "c1_control.a", "k2_call_groups.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, libraryOwnAnswerDirectory+name)
		})
	}
}

// A Map or Set iterator held by a program's own view, an interface or a type literal with next, stops at its
// toString, toLocaleString or Object.prototype.toString.call (#v414yhr): on main each printed [object Object] on
// every backend at exit 0, where Node prints [object Map Iterator] or [object Set Iterator] (J-a6w4knm). The view
// is the program's own, so libraryObjectOwnAnswer didn't ask the iterator; prototypeHazard's scan now counts the
// call that makes it. Mutant: the scan's new arm dropped (scratch-ts), which lowers every row again.
func TestLibraryObjectViewStops(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, position, library string }{
		{"v1_interface_map_tostring.a", "6:13", "a library MapIterator"},
		{"v2_interface_set_locale.a", "6:13", "a library SetIterator"},
		{"v3_interface_map_call.a", "6:13", "a library MapIterator"},
		{"v4_literal_set_tostring.a", "3:13", "a library SetIterator"},
		{"v5_literal_map_locale.a", "3:13", "a library MapIterator"},
		{"v6_literal_set_call.a", "3:13", "a library SetIterator"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, libraryOwnAnswerFixture(t, stop.name))
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), stop.library+" may be hidden by the view") || !strings.Contains(err.Error(), stop.name+":"+stop.position) {
				t.Fatalf("want a NotYet at %s:%s naming %s hidden by the view, got %v", stop.name, stop.position, stop.library, err)
			}
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
