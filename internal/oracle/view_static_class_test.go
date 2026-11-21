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
var viewStaticClassAgreeing = []string{"union_typeof.ts", "tagged_count.ts", "string_static.ts", "optional_absent.ts"}

const viewStaticClassKeys = "object_keys.ts"

// viewStaticClassStop is a program that keeps a stop on every backend, though Node runs it.
type viewStaticClassStop struct {
	name, stdout, stderr string
	// prefix and suffix hold the words where the declared type's spelling isn't pinned.
	prefix, suffix string
	// lowering admits a lowering refusal as the stop.
	lowering bool
}

const viewStaticClassNoMember = "adamic: panic: field read failed: viewed.value matches no member of number | Inner; expected number | Inner, found function\n"

const viewStaticClassWrite = "adamic: panic: field write failed: a write to count through a checked view that holds a class constructor is not yet lowered\n"

var viewStaticClassStops = []viewStaticClassStop{
	{name: "static_getter.ts", stdout: "before\n", stderr: viewStaticClassNoMember},
	{name: "inherited_count.ts", stdout: "before\n", stderr: viewStaticClassNoMember},
	{name: "name_field.ts", stdout: "before\n", stderr: viewStaticClassNoMember},
	// Klass owns call, so only StaticsAdmitted stops it: without the flag it would read Node's 5.
	{name: "call_field.ts", stdout: "before\n", stderr: viewStaticClassNoMember},
	{name: "callable_union.ts", stdout: "before\n", prefix: "adamic: panic: field read failed: viewed.value matches no member of ", suffix: ", found function\n", lowering: true},
	{name: "write.ts", stdout: "before\n", stderr: viewStaticClassWrite},
	// The value is evaluated before the guard on both backends, as native's writeProperty orders it.
	{name: "write_value_order.ts", stdout: "before\nvalue\n", stderr: viewStaticClassWrite},
	{name: "logical.ts", stdout: "before\nvalue\n", stderr: viewStaticClassWrite},
	// Each lowers to a SetProperty (class.go's compound assignment, assignments.go's increment), unless lowering
	// refuses it first; the log says which.
	{name: "compound.ts", stdout: "before\n", stderr: viewStaticClassWrite, lowering: true},
	{name: "increment.ts", stdout: "before\n", stderr: viewStaticClassWrite, lowering: true},
}

// A class constructor held where a checked view expects an object reads as Node reads it (#s03ndgw), when the
// object's contract admits one (ir.ViewContract's StaticsAdmitted: no field Function or Object names) and every
// field it requires is a data static the class owns. It is never converted: it stays the class, so typeof says
// function, and a field read through the view reads the static. Each program agrees with Node natively,
// sanitized and on the JavaScript backend, and leaks nothing (generatorAgrees). Before, each stopped at
// viewed.value, found function. Mutant: the is_static refusal put back in plain_matches's object branch
// (internal/native_runtime/c/view_unions_untagged.c), and these go red natively (scratch-s03p1/proof.sh).
func TestViewStaticClassAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range viewStaticClassAgreeing {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, viewStaticClassDirectory+name)
		})
	}
}

// Where the class isn't what the view says, or the view would do more than read its own statics, the program
// keeps a stop on every backend, though Node runs it: a static getter, a count only a base declares, a union a
// callable could also stand in, a field every constructor has (name, and call, which the class here owns too),
// and a write through the view by each road a write takes: =, a logical assignment, a compound assignment and
// an increment. A stop at the read says today's words. The callable union's declared type is printed by the
// checker, whose member order isn't pinned here, so its words are held by their start and end. Where lowering
// may refuse first, that is a stop too, and the log says so. Mutants: each backend's write guard dropped
// (scratch-s03p1/proof.sh), and the write pins go red.
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
			if err != nil {
				if !stop.lowering {
					t.Fatal(err)
				}
				t.Logf("stops at lowering: %v", err)
				return
			}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if stop.stderr != "" {
					want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte(stop.stderr)}
					if diff := disagreement(want, got); diff != "" {
						t.Errorf("%s: %s, stdout %q, stderr %q", backend, diff, got.stdout, got.stderr)
					}
					continue
				}
				stderr := string(got.stderr)
				if got.exitCode != 70 || string(got.stdout) != stop.stdout || !strings.HasPrefix(stderr, stop.prefix) || !strings.HasSuffix(stderr, stop.suffix) {
					t.Errorf("%s: exit %d, stdout %q, stderr %q; want exit 70, %q, and %q...%q", backend, got.exitCode, got.stdout, stderr, stop.stdout, stop.prefix, stop.suffix)
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
	if programs != 15 || len(named) != 15 {
		t.Fatalf("%d programs in %s and %d named, want 15: a program added or removed needs this count changed with it", programs, viewStaticClassDirectory, len(named))
	}
}
