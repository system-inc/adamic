package native

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A runtime shape is a host's when its methods are adamic_host_methods (adamic.h), and nothing makes
// that so: a shape whose methods are NULL reads as a plain object by a run-time key, so
// Object.prototype's closures stand in for a Stats' members and an iterator's internal slots go out
// as values. Nothing at run time catches that, so this test
// is the guard (#ey7aysy, Adamic's and Language's conditions): it finds every adamic_shape the runtime
// defines or builds, and every one a Go generator writes into C, and holds each to one row of
// hostShapeTable, which says what the shape's object is on Node and what pins that.
//
// The scanner reads C as tokens (cTokens, with line continuations joined), never by pattern. It finds
// a static or function-local initializer, each element of an array initializer as its own row, a
// declaration with no initializer (storage a builder fills, such as json_decode.c's
// decode_shape_entry), and a builder: a function that assigns a shape's names, writes an
// (adamic_shape){...} compound literal, sizes or aligns an adamic_shape, or holds a mutable
// adamic_shape pointer it got from a call. Anything else that names adamic_shape, a directive that
// does or a declaration it can't read, fails naming its file and line rather than being skipped.
//
// A row is one of three verdicts:
//   - Host: the object's prototype on Node isn't Object's, so the shape's methods must be
//     adamic_host_methods, or, for an Error, adamic_error_methods or adamic_coded_error_methods. Its pin is a
//     reflect_get stop fixture that stops on it on both backends where Node answers; or, for an
//     Error, the Error rule (its initializer's methods are one of the two Error sentinels, and
//     stop_host_error_prototype.a pins an Object.prototype name stopping on one); or, for a shape
//     no program holds, why not. That last one is prose, and so host only: if a program ever does
//     hold one, it stops rather than answering wrong.
//   - Plain: the object is a plain one on Node, so the shape mustn't name the sentinel. Its pin is a
//     reflect_get answer fixture and the text of the row in it that reads the object by a run-time
//     key, agreeing with Node on both backends.
//   - Unreachable: no program can read the object by a run-time key, because lowering refuses what
//     would make it. Its pin is a program in testdata/host_shapes that lowering must refuse in exactly
//     the refusal's words, so lifting the refusal fails this test. It names the host sentinel,
//     failing closed.
//
// A row whose shape no longer exists fails too, so the table can't rot. The JavaScript backend's
// wrappers for a native host object are held here as well (hostShapeWrappers), and so are the
// runtime's entrances by which a name reaches an unknown value (hostShapeEntrances), and every runtime
// function that reads a shape's names must call the host guard or be in hostShapeReaders with why it
// needn't. Mutants:
// TestHostShapesClosureMutants, run every time; and, by hand, stat_shape's sentinel dropped
// (node_fs_file.c), so the C disagrees with its row here and stop_host_stats_method.a answers
// natively, a row removed, so this test names its shape, and tsgo_query.a changed to lower, or to
// meet another refusal first, so this test says so.

type hostShapeVerdict string

const (
	hostShapeHost        hostShapeVerdict = "Host"
	hostShapePlain       hostShapeVerdict = "Plain"
	hostShapeUnreachable hostShapeVerdict = "Unreachable"
)

// hostShapeRow classifies one shape: its file under the repository, the function or struct it's in
// ("" at file scope), its name ("" for a builder or a Go generator), and which element of an array
// initializer or which of a generator's initializers it is.
type hostShapeRow struct {
	file, scope, name string
	index             int
	verdict           hostShapeVerdict
	// stop is a reflect_get stop fixture, for a host shape.
	stop string
	// errorShape is an Error's shape, host by the Error rule.
	errorShape bool
	// internal says why no program holds the object, for a host shape.
	internal string
	// answer is a reflect_get answer fixture, and row the text in it that reads the object by a
	// run-time key, for a plain shape. An answer with a directory is a program elsewhere in the
	// repository that runner runs against Node, which runner shows by holding runs.
	answer, row, runner, runs string
	// refusal is a testdata/host_shapes program and refused the refusal's exact words, for an
	// unreachable shape.
	refusal, refused string
	// replacedBy is the row of the builder that replaces this shape before any object takes it, for
	// an unreachable shape that needn't name the host sentinel itself.
	replacedBy string
}

func (row hostShapeRow) key() string {
	return fmt.Sprintf("%s %s %s[%d]", row.file, row.scope, row.name, row.index)
}

const (
	hostShapeRuntime = "internal/native/runtime/"
	plainFixture     = "prototype_runtime_plain.a"
	// A tuple stops as an array before the host test, so its flag is never read; it's set so that
	// nothing reads one as a plain object if that order ever changes.
	tupleReason = "a tuple, an array on Node, which stops as one before the host test"
	// The words of the refusals that keep an unreachable shape from a run-time key read, lower's
	// NotYets for an opaque host value and an object with a nullable field viewed as unknown, and
	// for a regular expression whose pattern isn't a constant.
	opaqueRefusal      = "an opaque host or collection value viewed as unknown or object (dynamic property metadata)"
	nullableRefusal    = "an object with a nullable field viewed as unknown or object (null and undefined slot tags)"
	nonconstantRefusal = "RegExp with a nonconstant pattern"
)

var hostShapeTable = []hostShapeRow{
	{file: hostShapeRuntime + "async_host_impl.h", scope: "host_settle", name: "shape", verdict: hostShapeHost, internal: "an embedder's host promise's value: only the C API's adamic_host_promise_new makes one, never a lowered program"},
	{file: hostShapeRuntime + "directory.c", name: "listing_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('listing', "},
	{file: hostShapeRuntime + "directory.c", name: "status_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('status', "},
	{file: hostShapeRuntime + "directory.c", name: "error_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('listing error', "},
	{file: hostShapeRuntime + "directory.c", scope: "real_path", name: "shape", verdict: hostShapeHost, internal: "realpath's result: fs.realpathSync, its one caller a program reaches, hands over the path alone"},
	{file: hostShapeRuntime + "exceptions.c", name: "error_shape", verdict: hostShapeHost, errorShape: true, stop: "stop_host_error_prototype.a"},
	{file: hostShapeRuntime + "exceptions.c", name: "absent_message_shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "input.c", name: "ok_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('read', "},
	{file: hostShapeRuntime + "input.c", name: "error_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('read error', "},
	{file: hostShapeRuntime + "input.c", name: "written_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('written', "},
	{file: hostShapeRuntime + "json_decode.c", scope: "make_object", verdict: hostShapePlain, answer: "prototype_decoded.a", row: "propertyOf(point.value, 'constructor')"},
	{file: hostShapeRuntime + "json_decode.c", scope: "struct decode_shape_entry", name: "shape", verdict: hostShapePlain, answer: "prototype_decoded.a", row: "propertyOf(point.value, 'constructor')"},
	{file: hostShapeRuntime + "json_decode.c", scope: "canonical_shape", verdict: hostShapePlain, answer: "prototype_decoded.a", row: "propertyOf(point.value, 'constructor')"},
	{file: hostShapeRuntime + "json_decode.c", name: "ok_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('decoded object', "},
	{file: hostShapeRuntime + "json_decode.c", name: "number_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('decoded number', "},
	{file: hostShapeRuntime + "json_decode.c", name: "boolean_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('decoded boolean', "},
	{file: hostShapeRuntime + "json_decode.c", name: "error_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('decode error', "},
	{file: hostShapeRuntime + "library_object.c", scope: "adamic_object_values", name: "number_pair", verdict: hostShapeHost, internal: "an Object.entries pair: " + tupleReason},
	{file: hostShapeRuntime + "library_object.c", scope: "adamic_object_values", name: "string_pair", verdict: hostShapeHost, internal: "an Object.entries pair: " + tupleReason},
	{file: hostShapeRuntime + "library_object.c", scope: "adamic_object_values", name: "boolean_pair", verdict: hostShapeHost, internal: "an Object.entries pair: " + tupleReason},
	{file: hostShapeRuntime + "library_object.c", scope: "adamic_object_values_checked", name: "number_pair", verdict: hostShapeHost, internal: "a checked Object.entries pair: " + tupleReason},
	{file: hostShapeRuntime + "library_object.c", scope: "adamic_object_values_checked", name: "boolean_pair", verdict: hostShapeHost, internal: "a checked Object.entries pair: " + tupleReason},
	{file: hostShapeRuntime + "library_object.c", scope: "adamic_object_values_checked", name: "string_pair", verdict: hostShapeHost, internal: "a checked Object.entries pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "iterator_shape", verdict: hostShapeHost, stop: "stop_host_iterator_next.a"},
	{file: hostShapeRuntime + "map_set.c", name: "state_shape", verdict: hostShapeHost, internal: "a collection iterator's state, held in its next closure's cell, which no program reads"},
	{file: hostShapeRuntime + "map_set.c", name: "result_shapes", index: 0, verdict: hostShapePlain, answer: plainFixture, row: "show('map number', "},
	{file: hostShapeRuntime + "map_set.c", name: "result_shapes", index: 1, verdict: hostShapePlain, answer: plainFixture, row: "show('set boolean', "},
	{file: hostShapeRuntime + "map_set.c", name: "result_shapes", index: 2, verdict: hostShapePlain, answer: plainFixture, row: "show('map entry', "},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 0, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 1, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 2, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 3, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 4, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 5, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 6, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 7, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "pair_shapes", index: 8, verdict: hostShapeHost, internal: "a Map's entries() pair: " + tupleReason},
	{file: hostShapeRuntime + "map_set.c", name: "string_state_shape", verdict: hostShapeHost, internal: "a string iterator's state, held in its next closure's cell, which no program reads"},
	{file: hostShapeRuntime + "node_crypto.c", name: "shape", verdict: hostShapeUnreachable, refusal: "hash.a", refused: nullableRefusal},
	{file: hostShapeRuntime + "node_fs_directory.c", name: "error_shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "node_fs_directory.c", name: "coded_error_shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "node_fs_directory.c", name: "dirent_shape", verdict: hostShapeHost, stop: "stop_host_dirent.a"},
	{file: hostShapeRuntime + "node_fs_file.c", name: "error_shape", verdict: hostShapeHost, errorShape: true, stop: "stop_host_fs_error_prototype.a"},
	{file: hostShapeRuntime + "node_fs_file.c", name: "coded_error_shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "node_fs_file.c", name: "date_shape", verdict: hostShapeUnreachable, refusal: "date.a", refused: opaqueRefusal},
	{file: hostShapeRuntime + "node_fs_file.c", scope: "realpath_system_error", name: "shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "node_fs_file.c", name: "stat_shape", verdict: hostShapeHost, stop: "stop_host_stats_method.a"},
	{file: hostShapeRuntime + "node_process.c", scope: "host_directory_error", name: "shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "node_process.c", scope: "adamic_node_memory_usage", name: "shape", verdict: hostShapeHost, stop: "stop_host_memory_field.a"},
	{file: hostShapeRuntime + "node_process.c", scope: "entry", name: "shape", verdict: hostShapeHost, stop: "stop_host_performance_mark.a"},
	{file: hostShapeRuntime + "node_process.c", scope: "adamic_node_stdout_handle", name: "shape", verdict: hostShapeHost, stop: "stop_host_handle.a"},
	{file: hostShapeRuntime + "node_process.c", scope: "adamic_node_performance", name: "shape", verdict: hostShapeUnreachable, refusal: "performance.a", refused: nullableRefusal},
	{file: hostShapeRuntime + "process.c", scope: "valid_code", name: "shape", verdict: hostShapeHost, errorShape: true},
	{file: hostShapeRuntime + "record.c", name: "record_number_shape", verdict: hostShapePlain, answer: "records.a", row: "show('count', "},
	{file: hostShapeRuntime + "record.c", name: "record_boolean_shape", verdict: hostShapePlain, answer: "records.a", row: "show('on', "},
	{file: hostShapeRuntime + "record.c", name: "record_reference_shape", verdict: hostShapePlain, answer: "records.a", row: "show('name', "},
	{file: hostShapeRuntime + "record.c", name: "iteration_shape", verdict: hostShapeHost, internal: "a record iterator's state, which only the emitter's loops over a record read"},
	{file: hostShapeRuntime + "regexp.c", name: "regex_shape", verdict: hostShapeUnreachable, refusal: "regexp.a", refused: opaqueRefusal},
	{file: hostShapeRuntime + "regexp.c", name: "match_shape", verdict: hostShapeHost, internal: "a match array's own properties, held behind the array, which stops as an array"},
	{file: hostShapeRuntime + "regexp.c", name: "pair_shape", verdict: hostShapeHost, internal: "a match's indices pair: " + tupleReason},
	{file: hostShapeRuntime + "regexp.c", name: "iterator_shape", verdict: hostShapeHost, stop: "stop_host_matchall_input.a"},
	{file: hostShapeRuntime + "regexp.c", name: "next_shape", verdict: hostShapePlain, answer: plainFixture, row: "show('match', "},
	// The run-time compiler's groups shape is built twice: regex_compile_names' first, which
	// adamic_regex_compile_new replaces with one holding the compiler's storage before any match
	// is made, so only the second names the sentinel (and regexp_compile_bytecode_test.go links the
	// first without the rest of the runtime). No program reaches either, or a run-time regular
	// expression, while a pattern must be a constant.
	{file: hostShapeRuntime + "regexp_compile_bytecode.c", scope: "regex_compile_names", verdict: hostShapeUnreachable, refusal: "regexp_runtime.a", refused: nonconstantRefusal, replacedBy: hostShapeRuntime + "regexp_compile_runtime.c adamic_regex_compile_new [0]"},
	{file: hostShapeRuntime + "regexp_compile_ownership.c", name: "regex_owned_shape", verdict: hostShapeUnreachable, refusal: "regexp_runtime.a", refused: nonconstantRefusal},
	{file: hostShapeRuntime + "regexp_compile_runtime.c", scope: "adamic_regex_compile_new", verdict: hostShapeUnreachable, refusal: "regexp_runtime.a", refused: nonconstantRefusal},
	{file: hostShapeRuntime + "tsgo.c", scope: "adamic_tsgo_query_in", name: "shape", verdict: hostShapeUnreachable, refusal: "tsgo_query.a", refused: "an unlinked typescript-go library call"},
	// Go generators of shapes the runtime's objects take: a regular expression literal's groups
	// shape, written when the program is compiled; a wasm export's record argument, which the host
	// passes as a plain object; and the emitter's own, a literal's or a class's layout, which the
	// program makes and Node calls plain.
	{file: "internal/regexp/native.go", scope: "NativeDeclarations", verdict: hostShapeHost, stop: "stop_host_groups.a"},
	{file: "internal/native/wasm_exports.go", scope: "WASIExports", verdict: hostShapePlain, answer: "internal/native/wasm/exports/types.a", row: "export function recordKey(", runner: "internal/native/wasm/exports/oracle.mjs", runs: "'recordKey'"},
	{file: "internal/native/emit_objects.go", scope: "shapeWith", index: 0, verdict: hostShapePlain, answer: "prototype_names.a", row: "propertyOf({}, 'toString')"},
	{file: "internal/native/emit_objects.go", scope: "shapeWith", index: 1, verdict: hostShapePlain, answer: "spread_plain.a", row: "propertyOf(made, 'constructor')"},
}

// hostShapeWrappers are the JavaScript backend's stand-ins for a native host object: each must still
// give its value a prototype that isn't Object's, as Node's own does, so adamicDynamicGet stops on it
// (dynamic_keyed.go) where native's host_object does.
var hostShapeWrappers = []struct {
	file, declaration, construction, native string
}{
	{"internal/javascript/library_map_set.go", "class AdamicIterator {", "new AdamicIterator(", hostShapeRuntime + "map_set.c  iterator_shape[0]"},
	{"internal/javascript/library_node_process.go", "class AdamicHostObject {}", "Object.setPrototypeOf(usage, AdamicHostObject.prototype)", hostShapeRuntime + "node_process.c adamic_node_memory_usage shape[0]"},
}

// hostShapeEntrances are the runtime's entrances by which a name or a key reaches a value viewed as
// unknown: an in test, the read it narrows and a run-time key read (Adamic's ruling on #ey7aysy at
// 23:03Z). Each must test the host predicate before a slot is consulted, every adamic_has_ or
// adamic_dynamic_ function a runtime header declares must be one of them, so a fourth reader of a
// value's slots by name fails here until it's guarded and listed, and the emitter must call each.
var hostShapeEntrances = []string{"adamic_has_property", "adamic_dynamic_property", "adamic_dynamic_property_keyed"}

var entranceDeclaration = regexp.MustCompile(`\b(adamic_(?:has|dynamic)_\w*)\s*\(`)

// entranceProblems holds union.c, the runtime's headers and the emitter to hostShapeEntrances.
func entranceProblems(union string, headers map[string]string, emitter string) []string {
	var problems []string
	tokens, err := cTokens(strings.ReplaceAll(union, "\\\n", "  "))
	if err != nil {
		return []string{"union.c: " + err.Error()}
	}
	guarded := map[string]bool{}
	depth, scope := 0, ""
	for index, current := range tokens {
		switch current.text {
		case "{":
			if depth == 0 {
				scope = braceScope(tokens, index)
			}
			depth++
		case "}":
			depth--
			if depth == 0 {
				scope = ""
			}
		case "host_stop", "host_object":
			if depth > 0 && index+1 < len(tokens) && tokens[index+1].text == "(" {
				guarded[scope] = true
			}
		}
	}
	for _, entrance := range hostShapeEntrances {
		if !guarded[entrance] {
			problems = append(problems, "union.c's "+entrance+" doesn't test host_stop or host_object, so a host object's slots answer through it")
		}
		if !strings.Contains(emitter, entrance+"(") {
			problems = append(problems, "the emitter no longer calls "+entrance+"; hostShapeEntrances names what it calls")
		}
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		for _, match := range entranceDeclaration.FindAllStringSubmatch(headers[name], -1) {
			if !slices.Contains(hostShapeEntrances, match[1]) {
				problems = append(problems, fmt.Sprintf("%s declares %s, a reader by name hostShapeEntrances doesn't hold: guard it with host_stop and list it", name, match[1]))
			}
		}
	}
	return problems
}

// entranceSources reads union.c, the runtime's headers and the emitter's expressions.
func entranceSources(t *testing.T) (string, map[string]string, string) {
	t.Helper()
	union, err := os.ReadFile(filepath.Join("runtime", "union.c"))
	if err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join("runtime", "*.h"))
	if err != nil {
		t.Fatal(err)
	}
	headers := map[string]string{}
	for _, path := range paths {
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		headers[filepath.Base(path)] = string(text)
	}
	emitter, err := os.ReadFile("emit_expressions.go")
	if err != nil {
		t.Fatal(err)
	}
	return string(union), headers, string(emitter)
}

// hostShapeReaders is every runtime function that reads a shape's names without the host guard
// (adamic_host_stop, host_stop or host_object), with why it needn't call it. Any other function that
// reads ->names or .names, other than to assign them, must call the guard first, or this test names
// it: a new reader of a value's names by a key fails here until it's guarded or listed. A reader
// whose host door is closed by lowering names the program that pins the refusal, which must be
// refused in exactly those words.
var hostShapeReaders = []struct {
	file, function, reason string
	refusal, refused       string
}{
	{"adamic.h", "adamic_shape_check", "checks a shape's own kinds and names no key", "", ""},
	{"class_features.c", "adamic_public_index", "a class instance's public shape, never a host's", "", ""},
	{"json_decode.c", "make_object", "lays out the decoder's own plain shape", "", ""},
	{"json_decode.c", "canonical_shape", "copies the decoder's own plain shape", "", ""},
	{"json_decode.c", "validate", "fills the decoder's own plain object", "", ""},
	{"json_encode.c", "encode_field", "reads only the fields an encoder schema names from the value's static type", "", ""},
	{"library_language.c", "adamic_plain_object_keys", "nothing the emitter writes calls it, and a write-ordered object goes through adamic_object_keys", "", ""},
	{"library_object.c", "adamic_object_has_own", "Object.hasOwn takes only a literal naming a declared public field", "door_has_own.a", "Object.hasOwn"},
	{"library_object.c", "adamic_object_ordered", "orders the slots of an object its callers have already admitted", "", ""},
	{"library_object.c", "adamic_object_values", "Object.values and Object.entries of a host are refused", "door_values.a", "Object.values"},
	{"library_object.c", "adamic_object_assign", "Object.assign from a host source is refused", "door_assign.a", "Object.assign"},
	{"object.c", "adamic_object_copy_checked", "a spread of a host is refused", "door_spread.a", "an object spread through a view that may hold library class Stats (its internal fields are not own enumerable properties)"},
	{"object.c", "adamic_object_find", "a static read by a name the program's type declares, which lowering maps to the shape's own field", "", ""},
	{"object.c", "adamic_object_callee", "a method call by a name the program's type declares", "", ""},
	{"object.c", "adamic_object_optional_find", "a static read by a name the program's type declares", "", ""},
	{"object.c", "adamic_object_view_write", "a static write by a name the program's type declares", "", ""},
	{"regexp.c", "regex_storage", "the regular expression runtime reading its own objects", "", ""},
	{"regexp.c", "adamic_regex_done", "the regular expression runtime reading its own objects", "", ""},
	{"regexp.c", "adamic_regex_group_lookup", "the regular expression runtime reading its own objects", "", ""},
	{"regexp.c", "adamic_regex_property", "the regular expression runtime reading its own objects", "", ""},
	{"regexp_compile_runtime.c", "adamic_regex_compile_new", "copies the run-time compiler's own groups shape", "", ""},
	{"union.c", "has_property", "reached past host_stop at adamic_has_property, or for an array's own properties", "", ""},
	{"union.c", "dynamic_property", "reached past host_stop at adamic_dynamic_property, or for an array's own properties", "", ""},
	{"union.c", "keyed_slot", "called by adamic_dynamic_property_keyed after its host_object test", "", ""},
}

// hostShapeDoors are the closed doors with no runtime reader of their own: lowering refuses the
// program before any function reads the host's names, so the refusal is the pin.
var hostShapeDoors = []struct{ refusal, refused string }{
	{"door_keys_unknown.a", "Object.keys on a non-object"},
	{"door_has_own_property_unknown.a", "hasOwnProperty on a union of differently held members (mixed value descriptors are not represented)"},
	{"door_string_typed.a", "String conversion of an object, array, map or function (ToPrimitive is not lowered)"},
	{"door_template_typed.a", "a template interpolating an object, an array, a map, a function or undefined"},
}

// nameReaders finds, in one C file, each function that reads a shape's names other than to assign
// them, and whether it calls the host guard.
func nameReaders(source string) (map[string]bool, error) {
	tokens, err := cTokens(strings.ReplaceAll(source, "\\\n", "  "))
	if err != nil {
		return nil, err
	}
	readers := map[string]bool{}
	guarded := map[string]bool{}
	depth, scope := 0, ""
	for index, current := range tokens {
		switch current.text {
		case "{":
			if depth == 0 {
				scope = braceScope(tokens, index)
			}
			depth++
			continue
		case "}":
			depth--
			if depth == 0 {
				scope = ""
			}
			continue
		}
		if depth == 0 || scope == "" || strings.HasPrefix(scope, "struct") {
			continue
		}
		next := func(offset int) string {
			if index+offset < len(tokens) {
				return tokens[index+offset].text
			}
			return ""
		}
		member := index > 0 && tokens[index-1].text == "." || index > 1 && tokens[index-1].text == ">" && tokens[index-2].text == "-"
		if current.text == "names" && member && !(next(1) == "=" && next(2) != "=") {
			readers[scope] = true
		}
		if (current.text == "host_stop" || current.text == "adamic_host_stop" || current.text == "adamic_host_stop_keys" || current.text == "host_object") && next(1) == "(" {
			guarded[scope] = true
		}
	}
	unguarded := map[string]bool{}
	for function := range readers {
		unguarded[function] = !guarded[function]
	}
	return unguarded, nil
}

// readerProblems holds every runtime reader of a shape's names to the guard or to hostShapeReaders.
func readerProblems(sources map[string]string) []string {
	var problems []string
	allowed := map[string]bool{}
	for _, reader := range hostShapeReaders {
		allowed[reader.file+" "+reader.function] = true
	}
	seen := map[string]bool{}
	files := make([]string, 0, len(sources))
	for file := range sources {
		files = append(files, file)
	}
	slices.Sort(files)
	for _, file := range files {
		readers, err := nameReaders(sources[file])
		if err != nil {
			problems = append(problems, file+": "+err.Error())
			continue
		}
		functions := make([]string, 0, len(readers))
		for function := range readers {
			functions = append(functions, function)
		}
		slices.Sort(functions)
		for _, function := range functions {
			key := file + " " + function
			seen[key] = true
			switch {
			case readers[function] && !allowed[key]:
				problems = append(problems, fmt.Sprintf("%s's %s reads a shape's names without adamic_host_stop: guard it, or list it in hostShapeReaders with why it needn't be", file, function))
			case !readers[function] && allowed[key]:
				problems = append(problems, fmt.Sprintf("%s's %s calls the host guard now; remove it from hostShapeReaders", file, function))
			}
		}
	}
	for _, reader := range hostShapeReaders {
		if !seen[reader.file+" "+reader.function] {
			problems = append(problems, fmt.Sprintf("hostShapeReaders names %s's %s, which no longer reads a shape's names; remove it", reader.file, reader.function))
		}
		if reader.refusal != "" {
			if problem := refusalProblem(filepath.Join("testdata", "host_shapes", reader.refusal), reader.refused); problem != "" {
				problems = append(problems, fmt.Sprintf("%s's %s: %s", reader.file, reader.function, problem))
			}
		}
	}
	for _, door := range hostShapeDoors {
		if problem := refusalProblem(filepath.Join("testdata", "host_shapes", door.refusal), door.refused); problem != "" {
			problems = append(problems, "hostShapeDoors: "+problem)
		}
	}
	return problems
}

// readerSources reads every runtime C file and header by its name.
func readerSources(t *testing.T) map[string]string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("runtime", "*.[ch]"))
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, path := range paths {
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sources[filepath.Base(path)] = string(text)
	}
	return sources
}

// foundShape is a shape the scanner found, under hostShapeRow's key, with what its C says.
type foundShape struct {
	file, scope, name string
	index, line       int
	host, errorShape  bool
}

func (shape foundShape) key() string {
	return fmt.Sprintf("%s %s %s[%d]", shape.file, shape.scope, shape.name, shape.index)
}

var hostShapeScan struct {
	once         sync.Once
	shapes       []foundShape
	unclassified []string
	err          error
}

// scannedHostShapes scans the runtime and the generators once for every test that reads them.
func scannedHostShapes(t *testing.T) ([]foundShape, []string) {
	t.Helper()
	hostShapeScan.once.Do(func() {
		shapes, unclassified, err := runtimeShapes(filepath.Join("..", "..", hostShapeRuntime))
		if err != nil {
			hostShapeScan.err = err
			return
		}
		generated, generatedUnclassified, err := generatedShapes(filepath.Join("..", ".."))
		if err != nil {
			hostShapeScan.err = err
			return
		}
		hostShapeScan.shapes = append(shapes, generated...)
		hostShapeScan.unclassified = append(unclassified, generatedUnclassified...)
	})
	if hostShapeScan.err != nil {
		t.Fatal(hostShapeScan.err)
	}
	return hostShapeScan.shapes, hostShapeScan.unclassified
}

var shapeWord = regexp.MustCompile(`\badamic_shape\b`)

// runtimeShapes finds every adamic_shape the runtime's C defines or builds.
func runtimeShapes(directory string) ([]foundShape, []string, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "*.[ch]"))
	if err != nil {
		return nil, nil, err
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no runtime sources in %s", directory)
	}
	slices.Sort(paths)
	var shapes []foundShape
	var unclassified []string
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		found, problems := fileShapes(hostShapeRuntime+filepath.Base(path), string(source))
		shapes = append(shapes, found...)
		unclassified = append(unclassified, problems...)
	}
	return shapes, unclassified, nil
}

// fileShapes scans one C file. A line continuation is two bytes, replaced by two spaces, so every
// token keeps its offset and a directive stays one token.
func fileShapes(file, source string) ([]foundShape, []string) {
	tokens, err := cTokens(strings.ReplaceAll(source, "\\\n", "  "))
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", file, err)}
	}
	line := func(index int) int { return strings.Count(source[:tokens[index].start], "\n") + 1 }
	at := func(index int) string {
		if index < 0 || index >= len(tokens) {
			return ""
		}
		return tokens[index].text
	}
	var unclassified []string
	refuse := func(index int, what string) {
		unclassified = append(unclassified, fmt.Sprintf("%s:%d: %s", file, line(index), what))
	}
	// scopes[index] is the function or struct whose braces hold the token, "" at file scope.
	scopes := make([]string, len(tokens))
	depth, scope := 0, ""
	for index, current := range tokens {
		switch current.text {
		case "{":
			if depth == 0 {
				scope = braceScope(tokens, index)
			}
			depth++
		case "}":
			depth--
			if depth < 0 {
				return nil, []string{fmt.Sprintf("%s:%d: a closing brace with none open; the scanner can't follow this file's scopes", file, line(index))}
			}
		}
		scopes[index] = scope
		if depth == 0 && current.text == "}" {
			scope = ""
		}
	}
	if depth != 0 {
		return nil, []string{fmt.Sprintf("%s: braces don't balance, as an #if branch that opens a function twice would leave them; the scanner can't follow this file's scopes", file)}
	}
	var shapes []foundShape
	// builders holds each function that builds a shape at run time, by name, with whether it marks it.
	builders := map[string]*foundShape{}
	builder := func(index int) *foundShape {
		name := scopes[index]
		if name == "" || strings.HasPrefix(name, "struct") {
			refuse(index, "an adamic_shape built outside a function")
			return nil
		}
		if builders[name] == nil {
			builders[name] = &foundShape{file: file, scope: name, line: line(index)}
		}
		return builders[name]
	}
	for index, current := range tokens {
		switch {
		case strings.HasPrefix(current.text, "#"):
			if shapeWord.MatchString(current.text) && !strings.HasPrefix(current.text, "#include") {
				refuse(index, "a directive that names adamic_shape")
			}
			continue
		case current.text == "names" && (at(index-1) == "." || at(index-1) == ">" && at(index-2) == "-") && at(index+1) == "=" && at(index+2) != "=":
			builder(index)
			continue
		case current.text == "methods" && (at(index-1) == "." || at(index-1) == ">" && at(index-2) == "-") && at(index+1) == "=" && at(index+2) != "=" && (at(index+3) == "adamic_host_methods" || errorSentinel(at(index+3))):
			if found := builder(index); found != nil {
				found.host = found.host || at(index+3) == "adamic_host_methods"
				found.errorShape = found.errorShape || errorSentinel(at(index+3))
			}
			continue
		case current.text != "adamic_shape":
			continue
		}
		before, after := at(index-1), at(index+1)
		switch {
		case before == "struct" && after == "{":
			if file != hostShapeRuntime+"adamic.h" {
				refuse(index, "a second definition of struct adamic_shape")
			}
		case before == "}" && after == ";":
			// The typedef's own name.
		case after == "*":
			// A pointer to a shape someone else defines, unless it's a mutable one the function got
			// from a call, to fill.
			if before != "const" && isIdentifier(at(index+2)) && at(index+3) == "=" && isIdentifier(at(index+4)) && at(index+5) == "(" {
				builder(index)
			}
		case before == "(" && after == ")" && (at(index-2) == "sizeof" || at(index-2) == "_Alignof"):
			builder(index)
		case before == "(" && after == ")" && at(index+2) == "{":
			found := builder(index)
			end := closingBrace(tokens, index+2)
			if found == nil || end < 0 {
				refuse(index, "an (adamic_shape){...} compound literal the scanner can't read")
				continue
			}
			host, errorShape, problem := shapeInitializer(tokens[index+3 : end])
			if problem != "" {
				refuse(index, problem)
				continue
			}
			found.host = found.host || host
			found.errorShape = found.errorShape || errorShape
		case isIdentifier(after) && at(index+2) == "=" && at(index+3) == "{":
			end := closingBrace(tokens, index+3)
			if end < 0 {
				refuse(index, "an initializer whose braces don't close")
				continue
			}
			host, errorShape, problem := shapeInitializer(tokens[index+4 : end])
			if problem != "" {
				refuse(index, problem)
				continue
			}
			shapes = append(shapes, foundShape{file: file, scope: scopes[index], name: after, line: line(index), host: host, errorShape: errorShape})
		case isIdentifier(after) && at(index+2) == "[":
			equals := index + 3
			for equals < len(tokens) && at(equals) != "]" {
				equals++
			}
			if at(equals+1) != "=" || at(equals+2) != "{" {
				refuse(index, "an array of adamic_shape with no initializer")
				continue
			}
			end := closingBrace(tokens, equals+2)
			if end < 0 {
				refuse(index, "an array initializer whose braces don't close")
				continue
			}
			for element, part := range initializerElements(tokens[equals+3 : end]) {
				if len(part) < 2 || part[0].text != "{" || part[len(part)-1].text != "}" {
					refuse(index, fmt.Sprintf("element %d of %s isn't a braced shape", element, after))
					continue
				}
				host, errorShape, problem := shapeInitializer(part[1 : len(part)-1])
				if problem != "" {
					refuse(index, fmt.Sprintf("element %d of %s: %s", element, after, problem))
					continue
				}
				shapes = append(shapes, foundShape{file: file, scope: scopes[index], name: after, index: element, line: line(index), host: host, errorShape: errorShape})
			}
		case isIdentifier(after) && at(index+2) == ";":
			// Storage some builder fills: it reads as plain until one marks it.
			shapes = append(shapes, foundShape{file: file, scope: scopes[index], name: after, line: line(index)})
		default:
			refuse(index, fmt.Sprintf("adamic_shape between %q and %q, which the scanner can't classify", before, after))
		}
	}
	names := make([]string, 0, len(builders))
	for name := range builders {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		shapes = append(shapes, *builders[name])
	}
	return shapes, unclassified
}

// braceScope names what a brace opened at file scope holds: a function by its name, a struct as
// "struct name", and "" for anything else, an initializer's braces among them.
func braceScope(tokens []cToken, open int) string {
	close := open - 1
	for close >= 0 && tokens[close].text == ")" {
		nesting, at := 0, close
		for ; at >= 0; at-- {
			if tokens[at].text == ")" {
				nesting++
			} else if tokens[at].text == "(" {
				nesting--
				if nesting == 0 {
					break
				}
			}
		}
		if at < 1 {
			return ""
		}
		// An attribute after the parameters: look past it for the function's own.
		if tokens[at-1].text == "__attribute__" {
			close = at - 2
			continue
		}
		return tokens[at-1].text
	}
	if close >= 1 && isIdentifier(tokens[close].text) && tokens[close-1].text == "struct" {
		return "struct " + tokens[close].text
	}
	return ""
}

func isIdentifier(text string) bool {
	if text == "" || text[0] >= '0' && text[0] <= '9' {
		return false
	}
	for index := 0; index < len(text); index++ {
		if !identifierByte(text[index]) {
			return false
		}
	}
	return true
}

// closingBrace is the index of the brace that closes the one at open, or -1.
func closingBrace(tokens []cToken, open int) int {
	nesting := 0
	for index := open; index < len(tokens); index++ {
		switch tokens[index].text {
		case "{":
			nesting++
		case "}":
			nesting--
			if nesting == 0 {
				return index
			}
		}
	}
	return -1
}

// initializerElements splits an initializer's contents at its top-level commas.
func initializerElements(tokens []cToken) [][]cToken {
	var elements [][]cToken
	nesting, start := 0, 0
	for index, current := range tokens {
		switch current.text {
		case "{", "(", "[":
			nesting++
		case "}", ")", "]":
			nesting--
		case ",":
			if nesting == 0 {
				elements = append(elements, tokens[start:index])
				start = index + 1
			}
		}
	}
	if start < len(tokens) {
		elements = append(elements, tokens[start:])
	}
	return elements
}

// shapeInitializer reads whether a shape's methods are the host sentinel or an Error's, from the
// contents of its braces, positional or designated.
func shapeInitializer(tokens []cToken) (host, errorShape bool, problem string) {
	elements := initializerElements(tokens)
	text := func(element []cToken) string {
		parts := make([]string, len(element))
		for index, part := range element {
			parts[index] = part.text
		}
		return strings.Join(parts, " ")
	}
	methods := ""
	for position, element := range elements {
		if len(element) > 0 && element[0].text == "." {
			if len(element) < 4 || element[2].text != "=" {
				return false, false, "a designated initializer the scanner can't read"
			}
			if element[1].text == "methods" {
				methods = text(element[3:])
			}
			continue
		}
		if position == 3 {
			methods = text(element)
		}
		if position > 4 {
			return false, false, "a shape initializer with more than five fields"
		}
	}
	return methods == "& adamic_host_methods", strings.HasPrefix(methods, "& ") && errorSentinel(strings.TrimPrefix(methods, "& ")), ""
}

// errorSentinel reports whether a methods table is one adamic_error_is reads as an Error's (exceptions.c):
// adamic_error_methods, or adamic_coded_error_methods for a Node-coded host error.
func errorSentinel(name string) bool {
	return name == "adamic_error_methods" || name == "adamic_coded_error_methods"
}

// generatorShape matches a shape initializer a Go generator writes into C, its fields captured.
var generatorShape = regexp.MustCompile(`adamic_shape\s+[%\w]+\s*=\s*\{([^{}]*)\}`)

// generatedShapes finds every shape initializer a Go generator writes, in internal, cmd and bridge,
// by the function whose string literal holds it. A literal that is just the name adamic_shape is a
// mention, as a stable name's prefix or a token compared; any other literal naming it is refused.
func generatedShapes(root string) ([]foundShape, []string, error) {
	var shapes []foundShape
	var unclassified []string
	for _, top := range []string{"internal", "cmd", "bridge"} {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !shapeWord.Match(source) {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			files := token.NewFileSet()
			parsed, err := parser.ParseFile(files, path, source, parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			for _, declaration := range parsed.Decls {
				function, isFunction := declaration.(*ast.FuncDecl)
				scope := ""
				if isFunction {
					scope = function.Name.Name
				}
				count := 0
				ast.Inspect(declaration, func(node ast.Node) bool {
					literal, isLiteral := node.(*ast.BasicLit)
					if !isLiteral || literal.Kind != token.STRING {
						return true
					}
					value, err := strconv.Unquote(literal.Value)
					if err != nil || !shapeWord.MatchString(value) {
						return true
					}
					where := fmt.Sprintf("%s:%d", relative, files.Position(literal.Pos()).Line)
					if value == "adamic_shape" {
						return true
					}
					matches := generatorShape.FindAllStringSubmatch(value, -1)
					if len(matches) == 0 || len(shapeWord.FindAllString(value, -1)) != len(matches) {
						unclassified = append(unclassified, where+": a string literal naming adamic_shape that isn't a shape initializer the scanner can read")
						return true
					}
					for _, match := range matches {
						fields := strings.Split(match[1], ",")
						if len(fields) != 5 {
							unclassified = append(unclassified, where+": a generated shape initializer without five fields")
							continue
						}
						host := strings.TrimSpace(fields[3]) == "&adamic_host_methods"
						errorShape := strings.HasPrefix(strings.TrimSpace(fields[3]), "&") && errorSentinel(strings.TrimPrefix(strings.TrimSpace(fields[3]), "&"))
						shapes = append(shapes, foundShape{file: relative, scope: scope, index: count, line: files.Position(literal.Pos()).Line, host: host, errorShape: errorShape})
						count++
					}
					return true
				})
				if !isFunction && count > 0 {
					unclassified = append(unclassified, relative+": a shape initializer outside a function")
				}
			}
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return shapes, unclassified, nil
}

// hostShapeProblems holds the found shapes to the table, and the table to its pins.
func hostShapeProblems(t *testing.T, table []hostShapeRow, shapes []foundShape, unclassified []string) []string {
	t.Helper()
	problems := append([]string{}, unclassified...)
	rows := map[string]hostShapeRow{}
	for _, row := range table {
		if _, twice := rows[row.key()]; twice {
			problems = append(problems, "hostShapeTable names "+row.key()+" twice")
		}
		rows[row.key()] = row
	}
	seen := map[string]bool{}
	for _, shape := range shapes {
		key := shape.key()
		if seen[key] {
			problems = append(problems, fmt.Sprintf("%s:%d: two shapes under one key, %s; the table can't tell them apart", shape.file, shape.line, key))
		}
		seen[key] = true
		row, classified := rows[key]
		if !classified {
			problems = append(problems, fmt.Sprintf("%s:%d: %s is in no row of hostShapeTable; classify it as host, plain or unreachable, with what pins that", shape.file, shape.line, key))
			continue
		}
		switch row.verdict {
		case hostShapeHost, hostShapeUnreachable:
			// An Error is a host by its own sentinel, which the checks below hold, and a replaced
			// shape by its replacement's.
			if replacement, replaced := rows[row.replacedBy]; row.replacedBy != "" && (!replaced || replacement.verdict == hostShapePlain) {
				problems = append(problems, fmt.Sprintf("%s:%d: %s is replaced by %s, which isn't a host row", shape.file, shape.line, key, row.replacedBy))
			}
			if !shape.host && !row.errorShape && row.replacedBy == "" {
				problems = append(problems, fmt.Sprintf("%s:%d: %s is %s in hostShapeTable, but its methods aren't adamic_host_methods", shape.file, shape.line, key, row.verdict))
			}
		case hostShapePlain:
			if shape.host || shape.errorShape {
				problems = append(problems, fmt.Sprintf("%s:%d: %s is plain in hostShapeTable, but its methods are a sentinel", shape.file, shape.line, key))
			}
		}
		if shape.errorShape && !row.errorShape {
			problems = append(problems, fmt.Sprintf("%s:%d: %s is an Error's shape, but its row doesn't say so", shape.file, shape.line, key))
		}
		if row.errorShape && !shape.errorShape {
			problems = append(problems, fmt.Sprintf("%s:%d: %s's row says it's an Error's shape, but its methods aren't an Error sentinel", shape.file, shape.line, key))
		}
	}
	for _, row := range table {
		if !seen[row.key()] {
			problems = append(problems, "hostShapeTable names "+row.key()+", which no longer exists; remove its row")
		}
	}
	for _, row := range table {
		problems = append(problems, hostShapePinProblems(t, row)...)
	}
	for _, wrapper := range hostShapeWrappers {
		source, err := os.ReadFile(filepath.Join("..", "..", wrapper.file))
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		if !strings.Contains(string(source), wrapper.declaration) || !strings.Contains(string(source), wrapper.construction) {
			problems = append(problems, fmt.Sprintf("%s no longer holds %q and %q, which give its value a prototype that isn't Object's", wrapper.file, wrapper.declaration, wrapper.construction))
		}
		if row, known := rows[wrapper.native]; !known || row.verdict != hostShapeHost {
			problems = append(problems, fmt.Sprintf("%s stands for %s, which isn't a host row", wrapper.file, wrapper.native))
		}
	}
	return problems
}

// reflectGetListed reports whether internal/oracle's reflect_get tests run the fixture.
func reflectGetListed(t *testing.T, fixture string) bool {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("..", "oracle", "reflect_get_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(source), `{"`+fixture+`", `)
}

func hostShapePinProblems(t *testing.T, row hostShapeRow) []string {
	t.Helper()
	fixtures := filepath.Join("..", "oracle", "testdata", "reflect_get")
	var problems []string
	switch row.verdict {
	case hostShapeHost:
		if row.stop == "" && !row.errorShape && row.internal == "" {
			problems = append(problems, row.key()+": a host row needs a stop fixture, the Error rule or why no program holds it")
		}
		if row.answer != "" || row.refusal != "" {
			problems = append(problems, row.key()+": a host row's pin is a stop fixture, the Error rule or a reason")
		}
		if row.stop != "" {
			if _, err := os.Stat(filepath.Join(fixtures, row.stop)); err != nil || !reflectGetListed(t, row.stop) {
				problems = append(problems, fmt.Sprintf("%s: its stop fixture %s isn't in reflect_get or isn't run by reflect_get_test.go", row.key(), row.stop))
			}
		}
		if row.errorShape && !reflectGetListed(t, "stop_host_error_prototype.a") {
			problems = append(problems, row.key()+": the Error rule's fixture, stop_host_error_prototype.a, isn't run by reflect_get_test.go")
		}
	case hostShapePlain:
		if row.answer == "" || row.row == "" || row.stop != "" || row.internal != "" || row.refusal != "" {
			problems = append(problems, row.key()+": a plain row's pin is an answer fixture and its row")
			break
		}
		path := filepath.Join(fixtures, row.answer)
		if strings.Contains(row.answer, "/") {
			path = filepath.Join("..", "..", row.answer)
			runner, err := os.ReadFile(filepath.Join("..", "..", row.runner))
			if err != nil || row.runs == "" || !strings.Contains(string(runner), row.runs) {
				problems = append(problems, fmt.Sprintf("%s: %s doesn't run %s's row (%q)", row.key(), row.runner, row.answer, row.runs))
			}
		} else if !reflectGetListed(t, row.answer) {
			problems = append(problems, fmt.Sprintf("%s: its answer fixture %s isn't run by reflect_get_test.go", row.key(), row.answer))
		}
		source, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(source), row.row) {
			problems = append(problems, fmt.Sprintf("%s: %s doesn't hold the row %q", row.key(), row.answer, row.row))
		}
	case hostShapeUnreachable:
		if row.refusal == "" || row.refused == "" || row.stop != "" || row.internal != "" || row.answer != "" {
			problems = append(problems, row.key()+": an unreachable row's pin is a refusal program and its words")
			break
		}
		if problem := refusalProblem(filepath.Join("testdata", "host_shapes", row.refusal), row.refused); problem != "" {
			problems = append(problems, row.key()+": "+problem)
		}
	default:
		problems = append(problems, row.key()+": no verdict")
	}
	return problems
}

// refusalProblem lowers the program and says what's wrong unless lowering refuses it in exactly
// the given words.
func refusalProblem(path, refused string) string {
	checked, err := load.Load([]string{path})
	if err != nil {
		return fmt.Sprintf("%s doesn't load: %v", path, err)
	}
	_, err = lower.Lower(context.Background(), checked)
	var refusal *lower.Refused
	var notYet *lower.NotYet
	switch {
	case errors.As(err, &refusal):
		if refusal.What == refused {
			return ""
		}
		return fmt.Sprintf("%s is refused as %q, not %q", path, refusal.What, refused)
	case errors.As(err, &notYet):
		if notYet.What == refused {
			return ""
		}
		return fmt.Sprintf("%s is NotYet as %q, not refused as %q", path, notYet.What, refused)
	case err == nil:
		return fmt.Sprintf("%s lowers, so the refusal that kept its shape from a run-time key read is lifted: classify the shape anew", path)
	}
	return fmt.Sprintf("%s fails to lower otherwise: %v", path, err)
}

// Every runtime shape, and every one a generator writes, is in hostShapeTable, its C agrees with its
// row, and each row's pin holds.
func TestHostShapesClosure(t *testing.T) {
	t.Parallel()
	shapes, unclassified := scannedHostShapes(t)
	if len(shapes) < len(hostShapeTable)/2 {
		t.Fatalf("the scanner found %d shapes for a table of %d: it has stopped seeing them", len(shapes), len(hostShapeTable))
	}
	for _, problem := range hostShapeProblems(t, hostShapeTable, shapes, unclassified) {
		t.Error(problem)
	}
	for _, problem := range entranceProblems(entranceSources(t)) {
		t.Error(problem)
	}
	for _, problem := range readerProblems(readerSources(t)) {
		t.Error(problem)
	}
	union, err := os.ReadFile(filepath.Join("runtime", "union.c"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(union), strconv.Quote(ir.DynamicKeysStop)) {
		t.Errorf("union.c's adamic_host_stop_keys doesn't say %q, ir.DynamicKeysStop's words the JavaScript backend stops with", ir.DynamicKeysStop)
	}
}

// The closure test's own mutants, run every time: a row dropped, and the test names its shape; a
// shape's C flag flipped, and the test names the disagreement; a builder unseen by the scanner's
// rules, and the test refuses it; a refusal program that lowers, and the test says the refusal is
// lifted.
func TestHostShapesClosureMutants(t *testing.T) {
	t.Parallel()
	shapes, unclassified := scannedHostShapes(t)
	dropped := hostShapeRuntime + "node_fs_file.c  stat_shape[0]"
	var fewer []hostShapeRow
	for _, row := range hostShapeTable {
		if row.key() != dropped {
			fewer = append(fewer, row)
		}
	}
	if len(fewer) != len(hostShapeTable)-1 {
		t.Fatalf("no row %s to drop", dropped)
	}
	if !mentions(hostShapeProblems(t, fewer, shapes, unclassified), dropped+" is in no row") {
		t.Errorf("a dropped row went unnamed")
	}
	flipped := append([]foundShape{}, shapes...)
	for index := range flipped {
		if flipped[index].key() == dropped {
			flipped[index].host = false
		}
	}
	if !mentions(hostShapeProblems(t, hostShapeTable, flipped, unclassified), dropped+" is Host in hostShapeTable, but its methods aren't adamic_host_methods") {
		t.Errorf("a host sentinel dropped went unnamed")
	}
	_, refused := fileShapes("mutant.c", "#include \"adamic.h\"\nvoid build(adamic_shape *into) { adamic_shape copy = *into; (void)copy; }\n")
	if len(refused) != 1 || !strings.Contains(refused[0], "mutant.c:2:") {
		t.Errorf("a shape copied by assignment wasn't refused: %q", refused)
	}
	_, refused = fileShapes("mutant.c", "#define SHAPE(n) static const adamic_shape n = {0};\n")
	if len(refused) != 1 {
		t.Errorf("a macro that writes a shape wasn't refused: %q", refused)
	}
	found, refused := fileShapes("mutant.c", "static adamic_shape *grow(size_t count) { adamic_shape *shape = malloc(sizeof *shape); return shape; }\n")
	if len(refused) != 0 || len(found) != 1 || found[0].scope != "grow" || found[0].host {
		t.Errorf("an allocating builder wasn't found as one: %+v %q", found, refused)
	}
	union, headers, emitter := entranceSources(t)
	unguarded := strings.Replace(union, "\thost_stop(value, name);\n\treturn has_property(value, name);", "\treturn has_property(value, name);", 1)
	if unguarded == union || !mentions(entranceProblems(unguarded, headers, emitter), "adamic_has_property doesn't test") {
		t.Errorf("an entrance unguarded went unnamed")
	}
	headers["mutant.h"] = "adamic_heap *adamic_dynamic_property_raw(adamic_heap *value, const char *name);\n"
	if !mentions(entranceProblems(union, headers, emitter), "declares adamic_dynamic_property_raw") {
		t.Errorf("a fourth reader by name went unnamed")
	}
	sources := readerSources(t)
	sources["mutant.c"] = "#include \"adamic.h\"\nbool mutant_has(const adamic_object *object, const char *name) {\n\tfor (size_t index = 0; index < object->shape->count; index++) if (strcmp(object->shape->names[index], name) == 0) return true;\n\treturn false;\n}\n"
	if !mentions(readerProblems(sources), "mutant.c's mutant_has reads a shape's names without adamic_host_stop") {
		t.Errorf("a new unguarded reader of a shape's names went unnamed")
	}
	delete(sources, "mutant.c")
	sources["object.c"] = strings.Replace(sources["object.c"], "\tadamic_host_stop(&object->heap, name);\n", "", 1)
	if !mentions(readerProblems(sources), "object.c's adamic_object_has reads a shape's names without adamic_host_stop") {
		t.Errorf("adamic_object_has's guard dropped went unnamed")
	}
	lowers := filepath.Join(t.TempDir(), "tsgo_query.a")
	if err := os.WriteFile(lowers, []byte("console.log('lowers');\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if problem := refusalProblem(lowers, "an unlinked typescript-go library call"); !strings.Contains(problem, "lowers, so the refusal") {
		t.Errorf("a refusal program that lowers went unnoticed: %q", problem)
	}
}

func mentions(problems []string, text string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, text) {
			return true
		}
	}
	return false
}
