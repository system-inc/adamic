package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Reflect.get(Object(value), key, value), 94's propertyOf, lowered as a read of an unknown by a key
// known only at run time (#zhgxzjg: Adamic's ruling on #axochk, the shape Compiler, Runtime and Emit
// accepted). The fixtures sit in reflect_get, since the flow and fresh corpora lower every program at
// the top of testdata. Each answer is held to Node natively, sanitized and released, and on the
// JavaScript backend; each stop is held to the same bytes on all three, since Node, running the
// source, has no stop. The receiver can't be pinned: Object(value) is value for an object, and a
// primitive, the one receiver that would differ, stops before the read.
func reflectGetFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reflect_get", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Mutants, each its own run with the file restored by cmp: dynamic_slot's retain dropped (union.c), so
// read_twice.a's string is freed under its holder and ASan reads freed memory; keyed_getter skipped
// on an accessor, so getter.a's getter reads undefined; the Error message remap dropped, so
// property_of.a's empty message reads undefined; record.c's number and boolean shapes swapped, so
// records.a's booleans read as numbers; one closure dropped from union.c's prototype_methods, so
// prototype_names.a reads that name as undefined; a closure made for each read in place of the
// immortal one, so prototype_names.a's identities print false.
func TestReflectGetAnswersAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"property_of.a", propertyOfAnswers},
		{"read_twice.a", "namename\nname\n"},
		{"getter.a", "getter: \"ENOENT\"\nabsent: undefined\n"},
		{"records.a", "count: 3\nmissing: undefined\non: true\noff: false\nname: \"ada\"\nhere: 4\ngone: undefined\n"},
		{"literal_method.a", "described\nfunction\n"},
		{"prototype_object.a", "before\nfunction\n"},
		{"prototype_record.a", "value\nfunction\n"},
		{"prototype_names.a", prototypeNamesAnswers},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := reflectGetFixture(t, probe.file)
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.stdout)}, truth); difference != "" {
				t.Fatalf("source Node: %s: exit %d, stdout %q, stderr %q", difference, truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("propertyOf must lower: %v", err)
			}
			native, binary := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Errorf("leaked: %s", report)
			}
		})
	}
}

// propertyOfAnswers is what property_of.a prints on Node.
const propertyOfAnswers = `code: "ENOENT"
errorCode: "ENOENT"
absent: undefined
message: ""
given message: "given"
name: "TypeError"
concatenated: "first"
substring: "second"
null: "Cannot read properties of null (reading 'code')"
undefined: "Cannot read properties of undefined (reading 'code')"
`

// prototypeNamesAnswers is what prototype_names.a prints on Node.
const prototypeNamesAnswers = `toString: function function
toLocaleString: function function
valueOf: function function
hasOwnProperty: function function
isPrototypeOf: function function
propertyIsEnumerable: function function
__defineGetter__: function function
__defineSetter__: function function
__lookupGetter__: function function
__lookupSetter__: function function
constructor: function function
true
true
true
`

// reflectGetStops are the stop fixtures: what each prints before it stops, which stop and key, and
// what Node prints running the source.
var reflectGetStops = []struct{ file, stdout, stop, key, node string }{
	{"stop_primitive.a", "before\n", "Primitive", "length", "before\nnumber\n"},
	{"stop_private.a", "hidden\n", "Private", "#x", "hidden\nundefined\n"},
	{"stop_method.a", "thing\n", "Method", "describe", "thing\nfunction\n"},
	{"stop_prototype_class.a", "thing\n", "ClassConstructor", "constructor", "thing\nfunction\n"},
	{"stop_proto.a", "before\n", "Proto", "__proto__", "before\nobject\n"},
	{"stop_error_stack.a", "before\n", "ErrorProperty", "stack", "before\nstring\n"},
	{"stop_array.a", "before\n", "Array", "length", "before\nnumber\n"},
	{"stop_literal_getter.a", "got\n", "ObjectLiteralGetter", "value", "got\nstring\n"},
}

// reflectGetStop is the run a stop fixture makes on each backend: what it printed, and the stop's
// words from internal/ir's one table with the key it was reading.
func reflectGetStop(t *testing.T, stdout, stop, key string) run {
	t.Helper()
	for _, entry := range ir.DynamicKeyedStops {
		if entry.Name == stop {
			return run{stdout: []byte(stdout), stderr: []byte("adamic: panic: " + entry.Message + " (reading '" + key + "')\n"), exitCode: 70}
		}
	}
	t.Fatalf("no stop named %s", stop)
	return run{}
}

// Each stop, on both backends, in the same words byte for byte. The private stop's mutant is
// TestReflectGetPrivateMutantJavaScript's. stop_prototype_class.a's is the JavaScript helper finding a
// class's accessors through adamicFindAccessor, which takes Object's constructor for an accessor
// without a getter, so that backend stops in the setter-only stop's words.
func TestReflectGetStopsAlikeOnBothBackends(t *testing.T) {
	t.Parallel()
	for _, probe := range reflectGetStops {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := reflectGetFixture(t, probe.file)
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("the read must lower and stop at run time: %v", err)
			}
			want := reflectGetStop(t, probe.stdout, probe.stop, probe.key)
			native, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": backend} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if difference := disagreement(native, backend); difference != "" {
				t.Errorf("the backends differ: %s", difference)
			}
		})
	}
}

// The JavaScript backend keeps a class's private fields as string keys, so without its own stop a
// run-time key "#x" reads one there, where native stops. Mutant: the helper's private check dropped,
// and stop_private.a no longer stops on the JavaScript backend.
func TestReflectGetPrivateMutantJavaScript(t *testing.T) {
	t.Parallel()
	path := reflectGetFixture(t, "stop_private.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(program)
	const check = "\tif (key.startsWith('#')) return adamicDynamicStop('Private', key);\n"
	if strings.Count(code, check) != 1 {
		t.Fatal("the private check is not in the helper once")
	}
	mutated := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(mutated, []byte(strings.Replace(code, check, "", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, mutated)
	want := reflectGetStop(t, "hidden\n", "Private", "#x")
	if disagreement(want, got) == "" {
		t.Fatalf("the private check's mutant survived: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
	}
	t.Logf("private check dropped: JavaScript exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
}

// Every other Reflect.get stays NotYet, as it was on main, and so does a whole program with one class
// getter the run-time read could call unseen, judged by its syntax: one that assigns, on a class the
// read never reaches; one that calls; one that reads another getter; one that writes an object field
// into a template, whose toString would run. Each is NotYet at the read, judged there before any
// getter's body is lowered. An
// Error subclass is NotYet at its class, so no emitted shape is an Error the read could miss
// (Runtime's must on #zhgxzjg): its stack, cause and errno reads never run on either backend, and
// when subclasses lower it goes red and asks for the mark. An object literal with a getter holding a
// method is NotYet at the method, so the read never meets that pair.
func TestReflectGetOtherShapesNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"notyet_two_arguments.a", "notyet_two_arguments.a:3:9", "Reflect.get other than Reflect.get(Object(value), key, value)"},
		{"notyet_receiver.a", "notyet_receiver.a:4:9", "Reflect.get other than Reflect.get(Object(value), key, value)"},
		{"notyet_unguarded.a", "notyet_unguarded.a:4:28", "Reflect.get of a value that may be null or undefined"},
		{"notyet_getter_effect.a", "hostErrors.ts:16:12", "a class getter that does more than return its fields and literals joined by operators"},
		{"notyet_getter_call.a", "hostErrors.ts:16:12", "a class getter that does more than return its fields and literals joined by operators"},
		{"notyet_getter_on_getter.a", "hostErrors.ts:16:12", "a class getter that does more than return its fields and literals joined by operators"},
		{"notyet_getter_to_primitive.a", "hostErrors.ts:16:12", "a class getter that converts an object field to a primitive"},
		{"notyet_literal_getter_method.a", "notyet_literal_getter_method.a:10:2", "a non-accessor method in an accessor literal"},
		{"error_subclass.a", "error_subclass.a:5:17", "a computed class base; name the base class directly"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, reflectGetFixture(t, probe.file))
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), probe.what) {
				t.Fatalf("want NotYet at %s naming %q, got %v", probe.at, probe.what, err)
			}
		})
	}
}

// hostErrors.ts as stage3 will carry it, 47's two helpers and propertyOf with Performance's message,
// lowered as the census lowers a root, lowers whole.
func TestPropertyOfHostErrorsLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowered(t, reflectGetFixture(t, "hostErrors.ts")); err != nil {
		t.Fatalf("hostErrors.ts must lower with propertyOf: %v", err)
	}
}

// Natively, a class's literal flag and method names come from the table the emitter writes for a
// program with the read (adamic_keyed_class). Mutant: the literal's row written false, so
// stop_literal_getter.a's getter is called and answers where both backends stop.
func TestReflectGetLiteralRowMutant(t *testing.T) {
	t.Parallel()
	program, err := lowered(t, reflectGetFixture(t, "stop_literal_getter.a"))
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	row := regexp.MustCompile(`(\t\{&\w+, )true(, )`)
	if len(row.FindAllString(code, -1)) != 1 {
		t.Fatal("the table has no single literal row")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(row.ReplaceAllString(code, "${1}false${2}"), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if want := reflectGetStop(t, "got\n", "ObjectLiteralGetter", "value"); disagreement(want, got) == "" {
		t.Fatalf("the literal row's mutant survived: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
	}
	t.Logf("literal row written false: native exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
}

// A program without a run-time key read emits no class table and no keyed read, so its C is what it
// was: the commit's proof compares this fixture's C with its base's byte for byte.
func TestReflectGetLeavesOtherProgramsAlone(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/arrow_this_super.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Classes) == 0 {
		t.Fatal("the fixture holds no class")
	}
	if code := native.C(program); strings.Contains(code, "adamic_keyed_class") || strings.Contains(code, "adamic_dynamic_property_keyed") {
		t.Fatal("a program without the read carries its table")
	}
}
