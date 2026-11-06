package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// tsc's debug.ts reads 19 enum objects through its own ESM namespace, each cast to a readonly
// string-index view, (ts as Readonly<Record<"SyntaxKind", Readonly<Record<string, string | number>>>>)
// .SyntaxKind, and getEnumMembers walks each with for...in and a run-time key read, caching by the
// object (B2, #52zja36, under Adamic's ruling of Nov 6). Each fixture is held to Node natively under
// the sanitizers and on the JavaScript backend.
func init() {
	for _, name := range []string{"cast"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/enum_views/" + name + ".a", true, false})
	}
}

func enumViewFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enum_views", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// The namespace cast is admitted only for one string literal K whose export is an enum with an object,
// seen through a readonly view whose value type holds the table's values. Every other cast of the
// namespace stops where it stopped on main: K naming two exports, K naming a plain object, a view
// narrower than the table's values (a downcast, refused at the cast), and a const enum, which has no
// object to read.
func TestEnumViewCastStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"stop_key_union.a", "stop_key_union.a:3:15", "an ESM namespace object used as a value"},
		{"stop_key_object.a", "stop_key_object.a:3:15", "an ESM namespace object used as a value"},
		{"stop_narrow_view.a", "stop_narrow_view.a:4:15", "Adamic 0.1 refuses a cast the runtime can't check"},
		{"stop_const_unbuilt.a", "stop_const_unbuilt.a:3:15", "an ESM namespace object used as a value"},
		{"stop_narrow_view_checked.ts", "stop_narrow_view_checked.ts:4:15", ""},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, enumViewFixture(t, probe.file))
			if err == nil || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), probe.what) {
				t.Fatalf("want a stop at %s naming %q, got %v", probe.at, probe.what, err)
			}
			t.Logf("stop: %v", err)
		})
	}
}

// tsc refuses these casts itself (TS2352: neither type sufficiently overlaps), as it did on main: a
// string-indexed K, K naming a function or a class, a view of string alone over a numeric enum, and
// Runtime's mutant on #6c4swvt, the view typed number, since the table's reverse keys hold strings. So
// none of them can reach lowering, let alone compile wrong.
func TestEnumViewCastsTscRefuses(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"stop_key_index.a", "stop_key_function.a", "stop_key_class.a", "stop_string_view.a", "stop_number_view.a"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			_, err := load.Load([]string{enumViewFixture(t, file)})
			var checked *load.CheckError
			if !errors.As(err, &checked) || !strings.Contains(err.Error(), "error TS2352") {
				t.Fatalf("want tsc's TS2352, got %v", err)
			}
		})
	}
}
