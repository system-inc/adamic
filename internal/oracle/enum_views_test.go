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
	for _, name := range []string{"cast", "forin", "members", "keyed_only", "const"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/enum_views/" + name + ".a", true, false})
	}
}

func enumViewFixture(test *testing.T, name string) string {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enum_views", name))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return path
}

// The namespace cast is admitted only for one string literal K whose export is an enum with an object,
// seen through a readonly view whose value type holds the table's values. Every other cast of the
// namespace stops where it stopped on main: K naming two exports, K naming a plain object, and a view
// narrower than the table's values (a downcast, refused at the cast).
func TestEnumViewCastStops(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"stop_key_union.a", "stop_key_union.a:3:15", "an ESM namespace object used as a value"},
		{"stop_key_object.a", "stop_key_object.a:3:15", "an ESM namespace object used as a value"},
		{"stop_narrow_view.a", "stop_narrow_view.a:4:15", "Adamic 0.1 refuses a cast the runtime can't check"},
		{"stop_narrow_view_checked.ts", "stop_narrow_view_checked.ts:4:15", ""},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			_, loweredError := lowered(test, enumViewFixture(test, probe.file))
			if loweredError == nil ||
				!strings.Contains(loweredError.Error(), probe.at) ||
				!strings.Contains(loweredError.Error(), probe.what) {
				test.Fatalf("want a stop at %s naming %q, got %v", probe.at, probe.what, loweredError)
			}
			test.Logf("stop: %v", loweredError)
		})
	}
}

// tsc refuses these casts itself (TS2352: neither type sufficiently overlaps), as it did on main: a
// string-indexed K, K naming a function or a class, a view of string alone over a numeric enum, and
// Runtime's mutant on #6c4swvt, the view typed number, since the table's reverse keys hold strings. So
// none of them can reach lowering, let alone compile wrong. Language's class instance seen as the view
// is refused the same way: a class has no implicit index signature.
func TestEnumViewCastsTscRefuses(test *testing.T) {
	test.Parallel()
	for _, file := range []string{
		"stop_key_index.a",
		"stop_key_function.a",
		"stop_key_class.a",
		"stop_string_view.a",
		"stop_number_view.a",
		"stop_class_instance.a",
	} {
		test.Run(file, func(test *testing.T) {
			test.Parallel()
			_, loadError := load.Load([]string{enumViewFixture(test, file)})
			var checked *load.CheckError
			if !errors.As(loadError, &checked) || !strings.Contains(loadError.Error(), "error TS2352") {
				test.Fatalf("want tsc's TS2352, got %v", loadError)
			}
		})
	}
}

// A run-time key read of the view arms the program-wide judgment of class getters Reflect.get's read
// does (Lower on #52zja36): a getter that runs code stops at the read, judged by its syntax, and one
// that may give null held bare stops there once every function is lowered. Mutant: the view's read
// not setting dynamicKeyedSite lowers stop_getter_null.a.
func TestEnumViewKeyedReadGetterStops(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{
			"stop_getter_effect.a",
			"stop_getter_effect.a:14:16",
			"a readonly view read by a run-time key in a program with a class getter that does more than return its fields",
		},
		{
			"stop_getter_null.a",
			"stop_getter_null.a:16:16",
			"a readonly view read by a run-time key in a program with a class getter that may give null held bare",
		},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			_, loweredError := lowered(test, enumViewFixture(test, probe.file))
			var notYet *lower.NotYet
			if !errors.As(loweredError, &notYet) ||
				!strings.Contains(loweredError.Error(), probe.at) ||
				!strings.Contains(loweredError.Error(), probe.what) {
				test.Fatalf("want NotYet at %s naming %q, got %v", probe.at, probe.what, loweredError)
			}
		})
	}
}

// Where Node answers what no representation holds, the keyed read stops loudly at run time on both
// backends alike, as Reflect.get's does, and never answers wrong: __proto__ (Object.prototype on
// Node), and an object literal's getter, which Node runs on every read (Language's p02).
func TestEnumViewKeyedReadStopsAlike(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, stdout, stderr string }{
		{
			"stop_proto.a",
			"Identifier number\n",
			"adamic: panic: __proto__ isn't read by a run-time key yet (reading '__proto__')\n",
		},
		{
			"stop_literal_getter.a",
			"",
			"adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'a')\n",
		},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			program, loweredError := lowered(test, enumViewFixture(test, probe.file))
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte(probe.stderr), exitCode: 70}
			compiled, _ := natively(test, program)
			for backend, got := range map[string]run{
				"native":     compiled,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf(
						"%s: %s: exit %d, stdout %q, stderr %q",
						backend,
						difference,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
			if truth := onNode(test, enumViewFixture(test, probe.file)); truth.exitCode != 0 {
				test.Fatalf("Node: %+v", truth)
			}
		})
	}
}

// The head's mutant for the materialized table (Library on #52zja36): drop an enum object's reverse
// keys, the array-index names that hold its members' names, and the key, typeof and value rows go red
// against Node on both backends, for an ordinary enum and a const enum the cast reached alike.
func TestEnumViewReverseKeysMutant(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, enum string }{{"members.a", "Ordered"}, {"const.a", "ConstOrdered"}} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			truth := onNode(test, enumViewFixture(test, probe.file))
			program, loweredError := lowered(test, enumViewFixture(test, probe.file))
			if loweredError != nil {
				test.Fatal(loweredError)
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
					if _, parseError := strconv.ParseFloat(field.Name, 64); parseError == nil {
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
				test.Fatalf("dropped %d reverse keys from %s, want 6", dropped, probe.enum)
			}
			compiled, _ := natively(test, program)
			for backend, got := range map[string]run{
				"native":     compiled,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if got.exitCode != 0 || disagreement(truth, got) != "stdout differs" {
					test.Fatalf(
						"%s: the reverse keys' mutant survived: exit %d, stdout %q, stderr %q",
						backend,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
				test.Logf("%s: reverse keys dropped from %s, caught by Node's stdout", backend, probe.enum)
			}
		})
	}
}

// A const enum the program reaches only by member reads has no object and no binding, as rulings.md:218
// has it; one the namespace cast reaches is built once. Mutant: build every const enum regardless of
// reach, and ConstKind's binding appears (TestConstEnumErasesRuntimeObject goes red with it).
func TestEnumViewConstBuiltOnlyWhereReached(test *testing.T) {
	test.Parallel()
	program, loweredError := lowered(test, enumViewFixture(test, "const.a"))
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	built := map[string]int{}
	for _, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok {
			if _, isLiteral := declaration.Value.(ir.ObjectLiteral); isLiteral {
				built[program.Locals[declaration.Local].Name]++
			}
		}
	}
	if built["ConstOrdered"] != 1 || built["ConstKind"] != 0 {
		test.Fatalf("want ConstOrdered built once and ConstKind never, got %v", built)
	}
}

// debug.ts's formatSyntaxKind shape, the unit Correctness's census found stopping at namespaces.go:94
// ("an enum read through its namespace without a lowered binding"): ConstKind, declared in another
// module of an export * barrel and read through the barrel's namespace in a function of the entry, is
// built, once, and agrees with Node on both backends with no leak. Reach is walked from the entry
// lowering lowers, so it sees the cast wherever the entry's modules hold it. Mutant: constEnumReached
// answering false everywhere turns this, and const.a, into the cast's stop, never a wrong read.
func TestEnumViewFormatConstAgrees(test *testing.T) {
	test.Parallel()
	path := enumViewFixture(test, "format_const.a")
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	built := 0
	for _, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "ConstKind" {
			built++
		}
	}
	if built != 1 {
		test.Fatalf("ConstKind built %d times, want once", built)
	}
	truth := onNode(test, path)
	if string(truth.stdout) != "Identifier Count 9\ntscUsesKeys\n" {
		test.Fatalf("Node: %+v", truth)
	}
	compiled, binary := natively(test, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(truth, got); difference != "" {
			test.Fatalf(
				"%s: %s: exit %d, stdout %q, stderr %q",
				backend,
				difference,
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatalf("leaks: %s", report)
	}
}

// The unbuilt pin Lower asked for on #52zja36: const enums read only by member reads, through the
// barrel's namespace and a named import, are never built, and the reads inline and agree with Node on
// both backends. Mutant: building every const enum regardless of reach gives ConstKind and
// ConstOrdered objects here.
func TestEnumViewConstUnbuiltAgrees(test *testing.T) {
	test.Parallel()
	path := enumViewFixture(test, "stop_const_unbuilt.a")
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	for _, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok {
			if name := program.Locals[declaration.Local].Name; name == "ConstKind" || name == "ConstOrdered" {
				test.Fatalf("%s was built, though only its members are read", name)
			}
		}
	}
	truth := onNode(test, path)
	if string(truth.stdout) != "2 3 1 4294967296\n" {
		test.Fatalf("Node: %+v", truth)
	}
	compiled, _ := natively(test, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(truth, got); difference != "" {
			test.Fatalf(
				"%s: %s: exit %d, stdout %q, stderr %q",
				backend,
				difference,
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		}
	}
}

// Lower's question on #52zja36: a const enum read by a literal union key, E[k], never reaches lowering,
// because tsc refuses it (TS2476), so reach needn't count it as a value read.
func TestEnumViewConstKeyUnionTscRefuses(test *testing.T) {
	test.Parallel()
	_, loadError := load.Load([]string{enumViewFixture(test, "stop_const_key_union.a")})
	var checked *load.CheckError
	if !errors.As(loadError, &checked) || !strings.Contains(loadError.Error(), "error TS2476") {
		test.Fatalf("want tsc's TS2476, got %v", loadError)
	}
}
