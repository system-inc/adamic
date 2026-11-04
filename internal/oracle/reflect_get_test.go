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
// immortal one, so prototype_names.a's identities print false; json_decode.c's canonical shapes
// marked host, so prototype_decoded.a stops natively at its first constructor; map_set.c's iterator
// result shapes marked host, so prototype_runtime_plain.a stops natively at its map entry (a plain
// shape's flag is what answers now, #ey7aysy, not its registration); dynamic_slot's representation
// bytes read only for an object with a write order, as before #rany24c, so decoded_scalar_reads.a's
// decoded number panics natively; directory.c's byte for a status's size dropped, so that row
// panics.
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
		{"prototype_decoded.a", "1 function function true\n3 function\nfunction\nfunction true\n"},
		{"prototype_runtime_plain.a", prototypeRuntimePlainAnswers},
		{"decoded_scalar_reads.a", decodedScalarReadsAnswers},
		{"spread_plain.a", "constructor function\ngiven x number constructor function\nwidth number constructor function\n"},
		{"error_fields.a", "code: \"ENOENT\"\nmessage: \"ENOENT: no such file or directory, open '/nonexistent-adamic-reflect-get/file'\"\ncaught foo: undefined\ncaught in: code yes message yes foo no\nerrorCode: \"ENOENT\"\nthrown: \"thrown\"\nname: \"Error\"\nthrown foo: undefined\nthrown in: message yes name yes foo no\n"},
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

// prototypeRuntimePlainAnswers is what prototype_runtime_plain.a prints on Node.
const prototypeRuntimePlainAnswers = `decoded object: function true
decoded number: function true
decoded boolean: function true
decode error: function true
read: function true
read error: function true
written: function true
write error: function true
listing: function true
listing error: function true
status: function true
status error: function true
map entry: function true
map number: function true
set boolean: function true
match: function true
`

// decodedScalarReadsAnswers is what decoded_scalar_reads.a prints on Node.
const decodedScalarReadsAnswers = `n number 1
b boolean true
s string text
count number 2
on boolean false
decoded result number 5
size number
symbolicLink boolean false
done boolean false
map value number 1
set value boolean true
present number 3
absent undefined
written number 4
cleared undefined
`

// reflectGetStops are the stop fixtures: what each prints before it stops, which stop and key, and
// what Node prints running the source.
var reflectGetStops = []struct{ file, stdout, stop, key, node string }{
	{"stop_primitive.a", "before\n", "Primitive", "length", "before\nnumber\n"},
	{"stop_private.a", "hidden\n", "Private", "#x", "hidden\nundefined\n"},
	{"stop_method.a", "thing\n", "Method", "describe", "thing\nfunction\n"},
	{"stop_prototype_class.a", "thing\n", "ClassConstructor", "constructor", "thing\nfunction\n"},
	{"stop_proto.a", "before\n", "Proto", "__proto__", "before\nobject\n"},
	{"stop_host_iterator.a", "iterator\n", "RuntimeValue", "constructor", "iterator\nfunction\n"},
	{"stop_host_object.a", "directory\n", "RuntimeValue", "constructor", "directory\nfunction\n"},
	{"stop_host_stats_method.a", "directory\n", "RuntimeValue", "isFile", "directory\nfunction\n"},
	{"stop_host_stats_field.a", "directory\n", "RuntimeValue", "dev", "directory\nnumber\n"},
	{"stop_host_stats_size.a", "directory\n", "RuntimeValue", "size", "directory\nnumber\n"},
	{"stop_host_matchall_input.a", "iterator\n", "RuntimeValue", "input", "iterator\nundefined\n"},
	{"stop_host_matchall_next.a", "iterator\n", "RuntimeValue", "next", "iterator\nfunction\n"},
	{"stop_host_iterator_next.a", "iterator\n", "RuntimeValue", "next", "iterator\nfunction\n"},
	{"stop_host_memory_field.a", "memory\n", "RuntimeValue", "rss", "memory\nnumber\n"},
	{"stop_host_memory_heap.a", "memory number\n", "RuntimeValue", "heapUsed", "memory number\nnumber\n"},
	{"stop_host_dirent.a", "entries true\n", "RuntimeValue", "isFile", "entries true\nfunction\n"},
	{"stop_host_performance_mark.a", "mark adamic\n", "RuntimeValue", "name", "mark adamic\nstring\n"},
	{"stop_host_handle.a", "handle true\n", "RuntimeValue", "setBlocking", "handle true\nfunction\n"},
	{"stop_host_groups.a", "groups true\n", "RuntimeValue", "letter", "groups true\nstring\n"},
	{"stop_host_error_prototype.a", "error made\n", "RuntimeValue", "valueOf", "error made\nfunction\n"},
	{"stop_host_fs_error_prototype.a", "caught\n", "RuntimeValue", "hasOwnProperty", "caught\nfunction\n"},
	{"stop_in_host_internal.a", "stats\n", "RuntimeValue", "_fsFileMode", "stats\ninternal absent\nsize number\n"},
	{"stop_in_host_method.a", "stats\n", "RuntimeValue", "isFile", "stats\nisFile yes\n"},
	{"stop_in_host_iterator.a", "iterator\n", "RuntimeValue", "input", "iterator\ninput no\n"},
	{"stop_in_host_literal.a", "", "RuntimeValue", "isFile", "isFile yes\ndev yes\nmtime yes\ninput no\nnext yes\n"},
	{"stop_in_host_error_code.a", "code ENOENT\n", "RuntimeValue", "code", "code ENOENT\nstats code none\n"},
	{"stop_hasownproperty_host.a", "stats\n", "RuntimeValue", "_fsFileMode", "stats\ninternal false\n"},
	{"stop_string_host.a", "iterator\n", "RuntimeValue", "toString", "iterator\nstring [object RegExp String Iterator]\n"},
	{"stop_keys_host.a", "stats\n", "OwnKeys", "", "stats\nkeys true false\n"},
	{"stop_keys_iterator.a", "iterator\n", "OwnKeys", "", "iterator\nkeys []\n"},
	{"stop_keys_memory.a", "memory\n", "OwnKeys", "", "memory\nkeys rss,heapTotal,heapUsed,external,arrayBuffers\n"},
	{"stop_error_stack.a", "before\n", "ErrorProperty", "stack", "before\nstring\n"},
	{"stop_array.a", "before\n", "Array", "length", "before\nnumber\n"},
	{"stop_literal_getter.a", "got\n", "ObjectLiteralGetter", "value", "got\nstring\n"},
}

// reflectGetStop is the run a stop fixture makes on each backend: what it printed, and the stop's
// words from internal/ir's one table with the key it was reading.
func reflectGetStop(t *testing.T, stdout, stop, key string) run {
	t.Helper()
	// Object.keys of a host value stops in a whole message that names no key (ir.DynamicKeysStop).
	if stop == "OwnKeys" {
		return run{stdout: []byte(stdout), stderr: []byte("adamic: panic: " + ir.DynamicKeysStop + "\n"), exitCode: 70}
	}
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
// without a getter, so that backend stops in the setter-only stop's words. stop_host_object.a's are
// union.c's host_object test dropped, so a Stats' constructor reads as Object's closure natively,
// and the helper's prototype test dropped, so the JavaScript backend reads Node's Stats.
// stop_host_iterator.a's is the JavaScript backend's AdamicIterator made a plain object again
// (library_map_set.go), so that backend hands out Object as a Map iterator's constructor.
// The stop_host_ rows after stop_host_object.a are #ey7aysy's: a host object stops for every key.
// Their mutants are union.c's guard put back on Object.prototype's names only, so
// stop_host_stats_method.a's isFile reads undefined natively, and
// TestReflectGetHostMutantJavaScript's. The rows from stop_host_dirent.a on pin each runtime shape
// marked host to Node (host_shapes_test.go's table names each one's fixture); a shape's host sentinel
// dropped, stat_shape's in node_fs_file.c or the groups shape's in internal/regexp/native.go, and its
// fixture answers natively. The stop_in_host_ rows are an in test on a host object (Adamic's ruling on
// #ey7aysy at 23:03Z). Their mutant is union.c's host_stop dropped from adamic_has_property, so
// stop_in_host_method.a's isFile answers no natively (stop_in_host_internal.a still stops, at the
// read the in test narrows, which keeps its own guard), and TestReflectGetInMutantJavaScript's.
// stop_hasownproperty_host.a and stop_string_host.a are the same ruling's other doors, hasOwnProperty
// (adamic_object_has) and ToPrimitive's own-toString test (adamic_object_own_function); their
// mutants are each guard dropped, so hasOwnProperty answers true and String answers [object Object]
// natively. The stop_keys_ rows are Object.keys of a host value, stopping in the keyless message
// (OwnKeys, Adamic's ruling at 08:28Z); their mutants are the keys guard dropped from
// adamic_object_keys and from adamic_class_object_keys, and the fixture each path serves answers from
// native's slots.
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

// The JavaScript backend's host test runs before an own property is read, as native's host_object
// does (#ey7aysy). Mutant: that early test dropped, so a host value is judged only after hasOwn, as
// before, and stop_host_stats_method.a's isFile reads undefined there instead of stopping.
func TestReflectGetHostMutantJavaScript(t *testing.T) {
	t.Parallel()
	path := reflectGetFixture(t, "stop_host_stats_method.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(program)
	const check = "\tif (host && !error) return adamicDynamicStop('RuntimeValue', key);\n"
	if strings.Count(code, check) != 1 {
		t.Fatal("the host check is not in the helper once")
	}
	mutated := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(mutated, []byte(strings.Replace(code, check, "", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, mutated)
	want := reflectGetStop(t, "directory\n", "RuntimeValue", "isFile")
	if disagreement(want, got) == "" {
		t.Fatalf("the host check's mutant survived: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
	}
}

// The JavaScript backend's in test stops on a host value as native's adamic_has_property does
// (#ey7aysy, Adamic's ruling at 23:03Z). Mutant: adamicDynamicHas answering key in value whatever
// the value, so stop_in_host_internal.a prints internal absent there and stops later, at the read of
// size its next in test narrows, which adamicDynamicRead guards.
func TestReflectGetInMutantJavaScript(t *testing.T) {
	t.Parallel()
	program, err := lowered(t, reflectGetFixture(t, "stop_in_host_internal.a"))
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(program)
	const check = "const adamicDynamicHas = (value, key) => adamicHostValue(value) ? adamicHostStop(key) : key in value;\n"
	if strings.Count(code, check) != 1 {
		t.Fatal("the in test's host check is not in the runtime once")
	}
	mutated := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(mutated, []byte(strings.Replace(code, check, "const adamicDynamicHas = (value, key) => key in value;\n", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, mutated)
	if disagreement(reflectGetStop(t, "stats\n", "RuntimeValue", "_fsFileMode"), got) == "" {
		t.Fatalf("the in test's mutant survived: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
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
