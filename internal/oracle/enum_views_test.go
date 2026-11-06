package oracle

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// tsc's debug.ts reads 19 enum objects through its own ESM namespace, each cast to a readonly
// string-index view, (ts as Readonly<Record<"SyntaxKind", Readonly<Record<string, string | number>>>>)
// .SyntaxKind, and getEnumMembers walks each with for...in and a run-time key read, caching by the
// object (B2, #52zja36, under Adamic's ruling of Nov 6). Each fixture is held to Node natively under
// the sanitizers and on the JavaScript backend.
func init() {
	for _, name := range []string{"cast", "forin", "members", "keyed_only"} {
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
// none of them can reach lowering, let alone compile wrong. Language's class instance seen as the view
// is refused the same way: a class has no implicit index signature.
func TestEnumViewCastsTscRefuses(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"stop_key_index.a", "stop_key_function.a", "stop_key_class.a", "stop_string_view.a", "stop_number_view.a", "stop_class_instance.a"} {
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

// A run-time key read of the view arms the program-wide judgment of class getters Reflect.get's read
// does (Lower on #52zja36): a getter that runs code stops at the read, judged by its syntax, and one
// that may give null held bare stops there once every function is lowered. Mutant: the view's read
// not setting dynamicKeyedSite lowers stop_getter_null.a.
func TestEnumViewKeyedReadGetterStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"stop_getter_effect.a", "stop_getter_effect.a:14:16", "a readonly view read by a run-time key in a program with a class getter that does more than return its fields"},
		{"stop_getter_null.a", "stop_getter_null.a:16:16", "a readonly view read by a run-time key in a program with a class getter that may give null held bare"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, enumViewFixture(t, probe.file))
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), probe.what) {
				t.Fatalf("want NotYet at %s naming %q, got %v", probe.at, probe.what, err)
			}
		})
	}
}

// Where Node answers what no representation holds, the keyed read stops loudly at run time on both
// backends alike, as Reflect.get's does, and never answers wrong: __proto__ (Object.prototype on
// Node), and an object literal's getter, which Node runs on every read (Language's p02).
func TestEnumViewKeyedReadStopsAlike(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout, stderr string }{
		{"stop_proto.a", "Identifier number\n", "adamic: panic: __proto__ isn't read by a run-time key yet (reading '__proto__')\n"},
		{"stop_literal_getter.a", "", "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'a')\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			program, err := lowered(t, enumViewFixture(t, probe.file))
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte(probe.stderr), exitCode: 70}
			compiled, _ := natively(t, program)
			for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if truth := onNode(t, enumViewFixture(t, probe.file)); truth.exitCode != 0 {
				t.Fatalf("Node: %+v", truth)
			}
		})
	}
}

// The head's mutant for the materialized table (Library on #52zja36): drop an enum object's reverse
// keys, the array-index names that hold its members' names, and the key, typeof and value rows go red
// against Node on both backends.
func TestEnumViewReverseKeysMutant(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, enum string }{{"members.a", "Ordered"}} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			truth := onNode(t, enumViewFixture(t, probe.file))
			program, err := lowered(t, enumViewFixture(t, probe.file))
			if err != nil {
				t.Fatal(err)
			}
			dropped := 0
			for position, statement := range program.Main {
				declaration, ok := statement.(ir.Declare)
				literal, isLiteral := declaration.Value.(ir.ObjectLiteral)
				if !ok || !isLiteral || program.Locals[declaration.Local].Name != probe.enum {
					continue
				}
				kept := []ir.Field{}
				for _, field := range literal.Fields {
					if _, err := strconv.ParseFloat(field.Name, 64); err == nil {
						dropped++
						continue
					}
					kept = append(kept, field)
				}
				literal.Fields = kept
				declaration.Value = literal
				program.Main[position] = declaration
			}
			if dropped != 6 {
				t.Fatalf("dropped %d reverse keys from %s, want 6", dropped, probe.enum)
			}
			compiled, _ := natively(t, program)
			for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
				if got.exitCode != 0 || disagreement(truth, got) != "stdout differs" {
					t.Fatalf("%s: the reverse keys' mutant survived: exit %d, stdout %q, stderr %q", backend, got.exitCode, got.stdout, got.stderr)
				}
				t.Logf("%s: reverse keys dropped from %s, caught by Node's stdout", backend, probe.enum)
			}
		})
	}
}
