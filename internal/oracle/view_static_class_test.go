package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const viewStaticClassDirectory = "internal/oracle/testdata/view_static_class/"

// viewStaticClassAgreeing and viewStaticClassKeys are the directory's programs the tests below don't name in
// their own tables; TestViewStaticClassProgramsAreAllRun holds the directory to every name.
var viewStaticClassAgreeing = []string{"union_typeof.ts", "tagged_count.ts", "string_static.ts", "optional_absent.ts", "widened_write.ts", "inherited_field.ts", "shadowed.ts", "three_level.ts", "inherited_optional.ts", "call_inherited.ts", "name_declared.ts"}

const viewStaticClassKeys = "object_keys.ts"

// viewStaticClassStop is a program that keeps a stop, though Node runs it: at lowering, in lowering's words,
// or at run time, on every backend, with stdout and stderr.
type viewStaticClassStop struct {
	name, stdout, stderr string
	// lowering is the words lowering refuses the program with, where it stops before any backend runs.
	lowering string
}

const viewStaticClassNoMember = "adamic: panic: field read failed: viewed.value matches no member of number | Inner; expected number | Inner, found function\n"

const viewStaticClassWrite = "adamic: panic: field write failed: a write to count through a checked view that holds a class constructor is not yet lowered\n"

// viewStaticClassNotInner is the stop of a read at a field typed Inner itself, natively where a class constructor is an object.
const viewStaticClassNotInner = "adamic: panic: field read failed: viewed.value is not a Inner; expected Inner, found function\n"

// viewStaticClassNarrowed is the stop the check lowering puts on a union local typeof narrowed gives a class constructor
// (localRead, internal/lower/locals.go), which asks typeof === 'object' and is answered 'function' (#775qbxz).
const viewStaticClassNarrowed = "adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n"

var viewStaticClassStops = []viewStaticClassStop{
	// Lowering stops a static getter in a checked field contract before the run time is asked, a base's too:
	// a subclass's allocation owns its bases' accessors (accessorReach).
	{name: "static_getter.ts", lowering: "stage 0 can't lower a getter in a checked field contract yet"},
	{name: "base_getter.ts", lowering: "stage 0 can't lower a getter in a checked field contract yet"},
	// The selection admits Sub, whose count Base holds, and Klass, which owns call: each stops past it, at the
	// read of a local typeof narrowed, as narrowed_read does. Before, each stopped at the selection, found function.
	{name: "inherited_count.ts", stdout: "before\n", stderr: viewStaticClassNarrowed},
	{name: "call_field.ts", stdout: "before\n", stderr: viewStaticClassNarrowed},
	// A required field named for a builtin or for Function's prototype is admitted at lowering, so the run time
	// walk refuses it: Klass declares no name and no call, and Node reads Function's.
	{name: "name_field.ts", stdout: "before\n", stderr: viewStaticClassNoMember},
	{name: "call_undeclared.ts", stdout: "before\n", stderr: viewStaticClassNoMember},
	// Base declares name, but Node reads Sub's own builtin name, which only a static Sub writes would shadow.
	{name: "name_inherited.ts", stdout: "before\n", stderr: viewStaticClassNotInner},
	// An optional field named for Function's prototype leaves StaticsAdmitted false: absent, it would read
	// undefined where Node reads Function's call.
	{name: "call_optional_undeclared.ts", stdout: "before\n", stderr: viewStaticClassNotInner},
	// Lowering refuses the read of a union a callable could also stand in: its contract is unsupported.
	{name: "callable_union.ts", lowering: "unsupported representation conversion contract"},
	// A static declared wider than Inner's count, holding a string, stops at lowering today (class_static.go).
	// Pinned for its own sake: when wide statics get native slots, the view's check must still stop it.
	{name: "wide_static.ts", lowering: "a static field without a native slot"},
	// The check lowering puts on a narrowed local (localRead, internal/lower/locals.go) asks typeof === 'object'
	// of what the read put back, and a class answers 'function'. It awaits a follow-up there.
	{name: "narrowed_read.ts", stderr: viewStaticClassNarrowed},
	{name: "write.ts", stdout: "before\n", stderr: viewStaticClassWrite},
	// The value is evaluated before the guard on both backends, as native's writeProperty orders it.
	{name: "write_value_order.ts", stdout: "before\nvalue\n", stderr: viewStaticClassWrite},
	{name: "logical.ts", stdout: "before\nvalue\n", stderr: viewStaticClassWrite},
	// Each lowers to a SetProperty (class.go's compound assignment, assignments.go's increment).
	{name: "compound.ts", stdout: "before\n", stderr: viewStaticClassWrite},
	{name: "increment.ts", stdout: "before\n", stderr: viewStaticClassWrite},
}

// A class constructor held where a checked view expects an object reads as Node reads it (#s03ndgw), when the
// object's contract admits one (ir.ViewContract's StaticsAdmitted: no field Object names, no optional field
// Function names) and every field it requires is a data static on the class's chain, its own or a base's, as
// Node walks the constructor's prototype chain: Sub's inherited count, a count two classes up, an optional
// extra only a base declares, a count the subclass shadows, and call and name, Function's names, where a class
// declares them. It is never converted: it stays the class, so typeof says function, and a field read through
// the view reads the static. Each program agrees with Node natively, sanitized and on the JavaScript backend,
// and leaks nothing (generatorAgrees). union_typeof holds it as a member of number | Inner, which the union's
// selection admits; the others hold it at a field typed as the object itself, which the read's own check
// admits (ir.Program's ObjectReadContract), so nothing narrows between the read and its use. Mutants
// (scratch-s03p1/proof.sh, then scratch-k24/proof.sh): the is_static refusal put back in plain_matches's
// object branch (internal/native_runtime/c/view_unions_untagged.c), and union_typeof goes red natively;
// adamic_view_static_read refusing every class constructor, and the others go red natively; adamicViewField's
// statics dropped (internal/javascript/view_fields.go), and the others go red on the JavaScript backend;
// static_own_slot or adamicViewStaticField (internal/javascript/readiness.go) not following the parent, and
// the inherited ones go red on that backend.
func TestViewStaticClassAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range viewStaticClassAgreeing {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, viewStaticClassDirectory+name)
		})
	}
}

// Where the class isn't what the view says, or the view would do more than read its statics, the program
// keeps a stop, though Node runs it: a static getter, the class's own or a base's, a static declared wider than
// its field, a union a callable could also stand in, a field every constructor has that the class doesn't
// declare (name, which a base's static doesn't shadow either, and call, required or optional), a read lowering
// can't yet follow past a typeof narrowing (narrowed_read, and inherited_count and call_field, which the
// selection admits), and a write through the view by each road a write takes: =, a logical assignment, a
// compound assignment and an increment. Each is held to where it stops: lowering, in lowering's words, or a
// run-time panic on every backend, in today's words. Mutants (scratch-s03p1/proof.sh, then
// scratch-k24/proof.sh): each backend's write guard dropped, and the write pins go red; each backend's read
// check admitting every class constructor, and name_inherited goes red; the builtin rule dropped natively, and
// name_inherited reads B; the JavaScript walk's enumerable test dropped, and name_field reads the emitted
// function's own name; lowering's optional test dropped, and call_optional_undeclared runs on natively.
func TestViewStaticClassStops(t *testing.T) {
	t.Parallel()
	for _, stop := range viewStaticClassStops {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, viewStaticClassDirectory+stop.name))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); node.exitCode != 0 {
				t.Fatalf("Node: %#v", node)
			}
			program, err := lowered(t, path)
			if stop.lowering != "" {
				if err == nil || !strings.Contains(err.Error(), stop.lowering) {
					t.Fatalf("lowering: %v, want a stop saying %q", err, stop.lowering)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte(stop.stderr)}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stdout %q, stderr %q", backend, diff, got.stdout, got.stderr)
				}
			}
		})
	}
}

// Object.keys through such a view reads the class's own statics, which agrees with Node's own enumerable keys,
// or stops: never other keys. Which of the two it is, each backend logs.
func TestViewStaticClassKeysAgreeOrStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, viewStaticClassDirectory+viewStaticClassKeys))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "before\ntag,count\n" {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Logf("stops at lowering: %v", err)
		return
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Logf("%s: agrees with Node", backend)
			continue
		}
		if got.exitCode != 70 || string(got.stdout) != "before\n" || !strings.HasPrefix(string(got.stderr), "adamic: panic: ") {
			t.Errorf("%s: neither Node's keys nor a stop: exit %d, stdout %q, stderr %q", backend, got.exitCode, got.stdout, got.stderr)
			continue
		}
		t.Logf("%s: stops: %q", backend, got.stderr)
	}
}

// Every program in the directory is one of the tests' above, and each is named once: a program added or
// removed needs its table changed with it.
func TestViewStaticClassProgramsAreAllRun(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, viewStaticClassDirectory))
	if err != nil {
		t.Fatal(err)
	}
	named := map[string]int{viewStaticClassKeys: 1}
	for _, name := range viewStaticClassAgreeing {
		named[name]++
	}
	for _, stop := range viewStaticClassStops {
		named[stop.name]++
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		if named[entry.Name()] != 1 {
			t.Errorf("%s is named %d times by the tests, want once", entry.Name(), named[entry.Name()])
		}
	}
	if programs != 28 || len(named) != 28 {
		t.Fatalf("%d programs in %s and %d named, want 28: a program added or removed needs this count changed with it", programs, viewStaticClassDirectory, len(named))
	}
}
