package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// In .ts, a record seen as a record that drops named contracts, tsc's CompilerOptions and WatchOptions seen as
// OptionsBase and CompilerOptions seen as MapLike, is the same object read through the view, its named reads held
// by the run-time check wherever they're read (lower's droppedContractRead, #6c4swvt G3a). These hold it natively
// and on the JavaScript backend.

// optionsBasePath is a fixture of records_options_base, absolute, as a stop's Where names it.
func optionsBasePath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_options_base", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// optionsBaseShown is a run as a failure message shows it.
func optionsBaseShown(got run) string {
	return fmt.Sprintf("exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
}

// tsc's three read shapes through the view, serializeOptionBaseObject's, getOwnKeys' and
// handleOptionConfigDirTemplateSubstitution's, which hands back the same object when it substitutes nothing, and
// Language's narrower field, strict: boolean read through OptionsBase, each lower and agree with Node, whose
// output is pinned so a fixture can't pass by printing nothing.
func TestOptionsBaseReadsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string]string{
		"serialize.ts": "outDir=string out strict=boolean true lib=object target=number 9 paths=object noEmit=boolean false\n" +
			"watchFile=number 4 excludeFiles=object\n\ntrue:out:4\n",
		"own_keys.ts": "target,strict,outDir,lib,configFilePath,configFile\nconfigFilePath=string tsconfig.json lib=object outDir=string out strict=boolean false target=number 9\n" +
			"false:9:tsconfig.json\n0\n",
		"substitution.ts": "false:/project/out:src:true:tsconfig.json:outDir,rootDir,strict,lib,configFilePath\n${configDir}/out:src:true:tsconfig.json\n" +
			"true:out:undefined:undefined:kept.json:outDir,target,configFilePath\nout:9:undefined:kept.json\ntrue\n",
		"narrower_field.ts": "strict boolean true\ntarget number 9\nstrict on\ntrue:9:true\n" +
			"strict boolean false\ntarget undefined\nstrict off\nfalse:undefined:false\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a record seen as OptionsBase in .ts must lower: %v", err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, optionsBaseShown(node), backend, optionsBaseShown(got))
				}
			}
		})
	}
}

// Language's condition: the check sits on every named read of a record's value, so it follows the value. A write
// through OptionsBase puts "yes" under strict, declared boolean, in the object the program holds as CompilerOptions,
// and each read form there checks it: Node prints it, and both backends stop on it with the read check's own words,
// after the line before it. The forms lowering can't read off a record yet stop not yet where they read, as they do
// on main whatever wrote the value: a destructuring, a ??= into a slot boxed as a union, and an optional chain.
func TestOptionsBaseWrittenReadsStop(t *testing.T) {
	t.Parallel()
	const found = " is not a boolean; expected boolean, found string, which a computed-key or index-signature write may have stored"
	const written = "written outDir,strict\n"
	for name, stop := range map[string]struct{ read, notYet, at string }{
		"written_plain.ts":       {read: "options.strict"},
		"written_destructure.ts": {notYet: "destructuring a value", at: ":23:7"},
		"written_coalesce.ts":    {notYet: "record ??= with a tagged optional value", at: ":23:1"},
		"written_chain.ts":       {notYet: "optional record property access", at: ":24:20"},
		"written_parameter.ts":   {read: "compilerOptions.strict"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != written+"yes\n" {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if stop.notYet != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != stop.notYet || notYet.Where != path+stop.at {
					t.Fatalf("want NotYet %q at %s, got %v", stop.notYet, stop.at, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, _ := natively(t, program)
			want := run{stdout: []byte(written), stderr: []byte("adamic: panic: field read failed: " + stop.read + found + "\n"), exitCode: 70}
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
				}
			}
		})
	}
}

// The back-cast at tsc's commandLineParser.ts:3098, a value seen as OptionsBase cast back as CompilerOptions, puts back
// named contracts the view never held, and the cast rule refuses it ahead of the record view. The day it's admitted,
// this goes red, and the reads it admits want pins of their own.
func TestOptionsBaseBackCastRefused(t *testing.T) {
	t.Parallel()
	path := optionsBasePath(t, "back_cast.ts")
	_, err := lowered(t, path)
	const want = ":22:17: Adamic 0.1 refuses an unproven relation from OptionsBase to CompilerOptions: optional field strict has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
	var refused *lower.Refused
	if !errors.As(err, &refused) || err.Error() != path+want {
		t.Fatalf("want the cast refused, %q, got %v", want, err)
	}
}

// A function over Options, boxed as a union, seen as one over a record of strings: the view drops strict, but the
// index value representations differ, so it stops at the existing storage stop. It's reached through a parameter,
// where no contract is dropped (parameterStorage); with that lifted, the first condition, one representation and
// sameKeeping, stops it alone. Mutant: lift both, and it's admitted there, to be refused later as another rule's
// widening, or read wrong. Outside a parameter, tsc's relation and a union slot leave no mismatch to reach.
func TestOptionsBaseRepresentationMismatchStops(t *testing.T) {
	t.Parallel()
	path := optionsBasePath(t, "representation.ts")
	_, err := lowered(t, path)
	const want = "a (options: Options) => string seen as (options: { [option: string]: string; }) => string (fixed objects and records have different storage; copy explicitly)"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != want || notYet.Where != path+":8:7" {
		t.Fatalf("want NotYet %q at :8:7, got %v", want, err)
	}
}

// A named field held as a union, kind?: string | number, is read back unchecked, so the view keeps its contract and
// stops at the existing storage stop. Mutant: drop the third condition's scalar test, and it lowers.
func TestOptionsBaseUncheckedFieldStops(t *testing.T) {
	t.Parallel()
	path := optionsBasePath(t, "unchecked_union.ts")
	_, err := lowered(t, path)
	const want = "a Options seen as OptionsBase (fixed objects and records have different storage; copy explicitly)"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != want || notYet.Where != path+":19:22" {
		t.Fatalf("want NotYet %q at :19:22, got %v", want, err)
	}
}

// #6c4swvt's dictionary and fixed-object head: paths?: MapLike<string[]> and readonly configFile?:
// TsConfigSourceFile are dropped by the view, since each one's named read out of the slot is a checked heap narrow
// to the slot union's only member of its runtime class (recordReadHeapMember). Language's must (08:38Z), an honest
// MapLike written under paths through the view read natively as another member, was the literal: { k: ["a"] } in a
// union slot was built fixed. It's built as the union's one dictionary member the program declares now
// (recordUnionDictionaryMember), so paths_honest.ts agrees with Node, and an honest TsConfigSourceFile, of its own
// shape or a wider one, does too. Mutants: build the literal fixed again, and paths_honest.ts stops natively on its
// honest value; take ir.Record and ir.Object out of droppedContractChecked, and every program here stops not yet at
// the view.
func TestOptionsBaseObjectFieldsDropped(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string]string{
		"paths_honest.ts":  "a\n",
		"config_honest.ts": "tsconfig.json:2\n",
		"config_wider.ts":  "tsconfig.json:2\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a dictionary or fixed-object member dropped by OptionsBase in .ts must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, optionsBaseShown(node), backend, optionsBaseShown(got))
				}
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// The lies those drops admit, each written through the view under a name its member declares, each caught at the
// named read's narrow on both backends alike, before the line after it: Language's g5 p01 to p05 under paths (a
// string, a string array, a number, a string read for its keys, a mixed array), and a dictionary and a string under
// configFile. Node reads each one.
func TestOptionsBaseObjectFieldLiesStop(t *testing.T) {
	t.Parallel()
	const paths = "adamic: panic: field read failed: options.paths is not a MapLike<string[]>; expected MapLike<string[]>, found another member of the index signature's union, which a computed-key or index-signature write may have stored\n"
	const config = "adamic: panic: field read failed: options.configFile is not a TsConfigSourceFile; expected TsConfigSourceFile, found another member of the index signature's union, which a computed-key or index-signature write may have stored\n"
	for name, stop := range map[string]struct{ node, stdout, stderr string }{
		"paths_string.ts":   {node: "undefined\n", stderr: paths},
		"paths_array.ts":    {node: "string\n1\n", stderr: paths},
		"paths_number.ts":   {node: "after\n", stderr: paths},
		"paths_keys.ts":     {node: "0,1\n", stderr: paths},
		"paths_mixed.ts":    {node: "number\n", stderr: paths},
		"config_maplike.ts": {node: "before\nundefined\n", stdout: "before\n", stderr: config},
		"config_string.ts":  {node: "before\nundefined\n", stdout: "before\n", stderr: config},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != stop.node {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, _ := natively(t, program)
			want := run{stdout: []byte(stop.stdout), stderr: []byte(stop.stderr), exitCode: 70}
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
				}
			}
		})
	}
}

// #6c4swvt (a), Runtime's pin: a record literal's computed entry puts a dictionary under configFile, which
// readGuardedInUnionSlot leaves to the read's checked narrow to a fixed object (recordReadHeapMember). Native's narrow
// tells a record by its shape; the JavaScript backend's read straight through it, as Node does, until every record it
// makes carried a mark (javascript's adamicRecords). Both stop alike now, after the line before the read. Mutant: let
// the JavaScript narrow to a fixed object ignore the mark, and the javascript row prints Node's undefined.
func TestOptionsBaseComputedEntryDictionaryStops(t *testing.T) {
	t.Parallel()
	path := optionsBasePath(t, "computed_entry_dictionary.ts")
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "before\nundefined\n" {
		t.Fatalf("Node: %s", optionsBaseShown(node))
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("must lower, as on main: %v", err)
	}
	native, _ := natively(t, program)
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: field read failed: options.configFile is not a TsConfigSourceFile; expected TsConfigSourceFile, found another member of the index signature's union, which a computed-key or index-signature write may have stored\n"), exitCode: 70}
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
		}
	}
}

// #6c4swvt's string[] head: an array member, lib?: string[], is dropped by the view, since its named read out of the
// slot is #y852c3c's element-checked view origin. An honest string[] written under lib through OptionsBase, and a
// (string | number)[] holding only strings, read back and agree with Node on both backends, with nothing leaked.
// Mutant: take the array out of droppedContractChecked, and each stops not yet at the view, :23:12.
func TestOptionsBaseArrayFieldDropped(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string]string{
		"lib_honest.ts": "DOM\nES2022\n2:es2022\n",
		"lib_wide.ts":   "DOM\nES2022\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("an array member dropped by OptionsBase in .ts must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, optionsBaseShown(node), backend, optionsBaseShown(got))
				}
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// The lies the dropped array member admits through the view, each caught on both backends alike: a 1 in a
// (string | number)[] stops at the element read that hands it out, a string under lib at the named read's narrow, and
// a push into the written array through the writer's own alias at compile time, since the read can't check what
// arrives after it (V3's demanded reach), pinned by its words, since V3's road stop names no position. Mutant:
// register no view origin for the slot read (recordReadSlotView), and lib_number.ts hands out the 1 as Node does, on
// both backends.
func TestOptionsBaseArrayFieldLiesStop(t *testing.T) {
	t.Parallel()
	for name, stop := range map[string]struct{ node, stdout, stderr, notYet, at string }{
		"lib_number.ts": {node: "before\nnumber\n", stdout: "before\n", stderr: "adamic: panic: element read failed: lib[1] is not a string; expected string, found number\n"},
		"lib_string.ts": {node: "before\nd\no\nm\n", stdout: "before\n", stderr: "adamic: panic: field read failed: options.lib is not a string[]; expected string[], found another member of the index signature's union, which a computed-key or index-signature write may have stored\n"},
		"lib_pushed.ts": {node: "number\n", notYet: "a checked view's array read through push (not checked yet)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != stop.node {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if stop.notYet != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != stop.notYet || stop.at != "" && notYet.Where != path+stop.at {
					t.Fatalf("want NotYet %q at %s, got %v", stop.notYet, stop.at, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, _ := natively(t, program)
			want := run{stdout: []byte(stop.stdout), stderr: []byte(stop.stderr), exitCode: 70}
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
				}
			}
		})
	}
}

// Runtime's road proof (#6c4swvt, 17:11Z): the head rests on every road that reads a named array member back out of
// the record reaching recordReadSlotView's origin, or stopping. Each fixture writes a (string | number)[] holding a
// 1 at index 1 under lib through OptionsBase, and reads it back another way. Node reads the 1, and no backend may:
//   - a literal key and a spread copy reach the origin, and stop at the element read on both backends alike;
//   - a destructuring, a destructured parameter and an optional chain stop not yet where they read, as they do on main
//     whatever wrote the value (TestOptionsBaseWrittenReadsStop's destructure and chain);
//   - for...in and Object.entries read the value through the index, as the slot union, and seeing it as string[] is a
//     cast the runtime can't check, refused;
//   - Language's library roads (o03, o04, o07 on #6c4swvt): join and slice read elements inside the runtime and hand
//     none out, so they stop not yet, pinned by their words (V3's road stop); every hands each element to its
//     callback, checked as it goes.
//
// The day one of the compile-time stops lifts, its row goes red here, and the road it opens must reach the origin.
func TestOptionsBaseArrayMemberRoadsStop(t *testing.T) {
	t.Parallel()
	const elementStop = "adamic: panic: element read failed: lib[1] is not a string; expected string, found number\n"
	const castRefused = "Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
	for road, stop := range map[string]struct{ node, notYet, refused, at, stderr string }{
		"literal_key":           {stderr: elementStop},
		"spread":                {stderr: elementStop},
		"destructure":           {notYet: "destructuring a value", at: ":26:7"},
		"destructure_parameter": {notYet: "destructuring a value", at: ":26:17"},
		"optional_chain":        {notYet: "optional record property access", at: ":27:13"},
		"for_in":                {refused: castRefused, at: ":29:21"},
		"entries":               {refused: castRefused, at: ":28:21"},
		"join":                  {node: "before\ndom,1\n", notYet: "a checked view's array read through join (not checked yet)"},
		"slice":                 {notYet: "a checked view's array read through slice (not checked yet)"},
		"every":                 {node: "before\nfalse\n", stderr: "adamic: panic: element read failed: lib's element, in every is not a string; expected string, found number\n"},
	} {
		t.Run(road, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, "lib_road_"+road+".ts")
			node := onNode(t, path)
			printed := "before\nnumber\n"
			if stop.node != "" {
				printed = stop.node
			}
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			switch {
			case stop.notYet != "":
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != stop.notYet || stop.at != "" && notYet.Where != path+stop.at {
					t.Fatalf("want NotYet %q at %s, got %v", stop.notYet, stop.at, err)
				}
				return
			case stop.refused != "":
				var refused *lower.Refused
				if !errors.As(err, &refused) || err.Error() != path+stop.at+": "+stop.refused {
					t.Fatalf("want refused %q at %s, got %v", stop.refused, stop.at, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, _ := natively(t, program)
			want := run{stdout: []byte("before\n"), stderr: []byte(stop.stderr), exitCode: 70}
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
				}
			}
		})
	}
}

// Language's should (08:38Z): a function over CompilerOptions seen as one over OptionsBase is handed an OptionsBase,
// so the view through its parameter adds strict. A view drops no contract through a parameter (parameterStorage),
// so each of Language's g4 p01 to p04 stops not yet at the view, as on main, where 419579d6 let the view through to
// the optional-widening refusal. Mutant: leave droppedContractsRead set through parameters, and each is refused
// there again.
func TestOptionsBaseParameterViewStops(t *testing.T) {
	t.Parallel()
	const function = "a (o: CompilerOptions) => string seen as (o: OptionsBase) => string (fixed objects and records have different storage; copy explicitly)"
	for name, stop := range map[string]struct{ what, at string }{
		"parameter_function.ts":   {function, ":20:7"},
		"parameter_bivariance.ts": {function, ":23:27"},
		"parameter_argument.ts":   {function, ":26:25"},
		"parameter_implements.ts": {"a StrictReader seen as Reader (fixed objects and records have different storage; copy explicitly)", ":28:7"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || notYet.What != stop.what || notYet.Where != path+stop.at {
				t.Fatalf("want NotYet %q at %s, got %v", stop.what, stop.at, err)
			}
		})
	}
}

// Language's g4 p06: a method over CompilerOptions written in a literal where Reader's read takes OptionsBase. A
// literal is made as its slot, so recordStorageView never sees the method, here as on main, and it lowers. Its body's
// named read stops on the "yes" under strict on both backends, so it rests on the read check alone. Stopping it at
// the method is a rule for methods written in literals, which this head doesn't make.
func TestOptionsBaseLiteralMethodStopsAtRead(t *testing.T) {
	t.Parallel()
	path := optionsBasePath(t, "method_parameter.ts")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("must lower, as on main: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "on\n" {
		t.Fatalf("Node: %s", optionsBaseShown(node))
	}
	native, _ := natively(t, program)
	want := run{stderr: []byte("adamic: panic: field read failed: o.strict is not a boolean; expected boolean, found string, which a computed-key or index-signature write may have stored\n"), exitCode: 70}
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
		}
	}
}

// The same views in .a stay refused, as #bq3fvdt rules, OptionsBase's and a library Record's alike:
// recordDeclaredKeyRefusal refuses them at 1975, before recordStorageView, whose gate keeps droppedContractRead to .ts
// behind it.
func TestOptionsBaseRefusedInAdamic(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"refused.a":        ":19:22: Adamic 0.1 refuses a value with declared properties seen through a wider index signature, which can write under a declared property what it does not hold (CompilerOptions seen as OptionsBase: strict is declared boolean | undefined, and the index signature holds string | number | boolean | undefined); give the record no declared keys, or write the declared property by its name",
		"refused_record.a": ":16:22: Adamic 0.1 refuses a value with declared properties seen through a wider index signature, which can write under a declared property what it does not hold (CompilerOptions seen as Record<string, string | number | boolean | undefined>: strict is declared boolean | undefined, and the index signature holds string | number | boolean | undefined); give the record no declared keys, or write the declared property by its name",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			_, err := lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || err.Error() != path+want {
				t.Fatalf("want the view refused, %q, got %v", want, err)
			}
		})
	}
}

// #6c4swvt's record-to-fixed direction, piece 2 (Runtime's ruling): a record seen in .ts as Options or
// WatchOptions, which declare optional names the record's own values don't promise, hides nothing under them,
// since each named read through the view is checked by its declared type (the optional-widening rule's
// recordAddedNameChecked, on addedContractRead's conditions). The honest views agree with Node on both backends,
// with nothing leaked, at tsc's three shapes: the view itself, a parse result whose options field is seen so
// (commandLineParser.ts:2129) and a MapLike in a conditional's branch seen as WatchOptions (3100). A lie under
// an added name stops at its read on both backends alike, and a view through a parameter still stops at the
// view. Mutant: recordAddedNameChecked returning false, and every program here but the parameter's is refused
// as one that can hide fields.
func TestOptionsBaseAddedNames(t *testing.T) {
	t.Parallel()
	const strict = "adamic: panic: field read failed: %s is not a boolean; expected boolean, found string, which a computed-key or index-signature write may have stored\n"
	for name, view := range map[string]struct{ node, stdout, stderr, notYet string }{
		"added_honest.ts":    {node: "true:out:undefined\n9:strict,outDir,target\n"},
		"added_watch.ts":     {node: "true:x.ts\n"},
		"added_lie.ts":       {node: "before\nyes\n", stdout: "before\n", stderr: fmt.Sprintf(strict, "options.strict")},
		"added_nested.ts":    {node: "true:out:a.ts\nbefore\nyes\n", stdout: "true:out:a.ts\nbefore\n", stderr: fmt.Sprintf(strict, "lie.options.strict")},
		"added_parameter.ts": {node: "yes\n", notYet: "(fixed objects and records have different storage; copy explicitly)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != view.node {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if view.notYet != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.What, view.notYet) {
					t.Fatalf("want NotYet ending %q, got %v", view.notYet, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("a record seen as a view adding checked names in .ts must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			want := node
			if view.stderr != "" {
				want = run{stdout: []byte(view.stdout), stderr: []byte(view.stderr), exitCode: 70}
			}
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
				}
			}
			if view.stderr == "" {
				if report := leaks(t, program, sanitized); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
