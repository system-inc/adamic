package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// In .ts, a record seen as a record that drops named contracts, tsc's CompilerOptions and WatchOptions seen as
// OptionsBase and CompilerOptions seen as MapLike, is the same object read through the view, its named reads held
// by the run-time check wherever they're read (lower's droppedContractRead, #6c4swvt G3a). These hold it natively
// and on the JavaScript backend.

// optionsBasePath is a fixture of records_options_base, absolute, as a stop's Where names it.
func optionsBasePath(test *testing.T, name string) string {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_options_base", name),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
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
func TestOptionsBaseReadsAgreeWithNode(test *testing.T) {
	test.Parallel()
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
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := optionsBasePath(test, name)
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("a record seen as OptionsBase in .ts must lower: %v", loweredError)
			}
			node := onNode(test, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				test.Fatalf("Node: %s", optionsBaseShown(node))
			}
			native, _ := natively(test, program)
			for backend, got := range map[string]run{
				"native":     native,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(node, got); difference != "" {
					test.Fatalf(
						"%s disagrees with Node (%s): Node %s; %s %s",
						backend,
						difference,
						optionsBaseShown(node),
						backend,
						optionsBaseShown(got),
					)
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
func TestOptionsBaseWrittenReadsStop(test *testing.T) {
	test.Parallel()
	const found = " is not a boolean; expected boolean, found string, which a computed-key or index-signature write may have stored"
	const written = "written outDir,strict\n"
	for name, stop := range map[string]struct{ read, notYet, at string }{
		"written_plain.ts":       {read: "options.strict"},
		"written_destructure.ts": {notYet: "destructuring a value", at: ":23:7"},
		"written_coalesce.ts":    {notYet: "record ??= with a tagged optional value", at: ":23:1"},
		"written_chain.ts":       {notYet: "optional record property access", at: ":24:20"},
		"written_parameter.ts":   {read: "compilerOptions.strict"},
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := optionsBasePath(test, name)
			node := onNode(test, path)
			if node.exitCode != 0 || string(node.stdout) != written+"yes\n" {
				test.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, loweredError := lowered(test, path)
			if stop.notYet != "" {
				var notYet *lower.NotYet
				if !errors.As(loweredError, &notYet) || notYet.What != stop.notYet || notYet.Where != path+stop.at {
					test.Fatalf("want NotYet %q at %s, got %v", stop.notYet, stop.at, loweredError)
				}
				return
			}
			if loweredError != nil {
				test.Fatalf("must lower: %v", loweredError)
			}
			native, _ := natively(test, program)
			want := run{
				stdout:   []byte(written),
				stderr:   []byte("adamic: panic: field read failed: " + stop.read + found + "\n"),
				exitCode: 70,
			}
			for backend, got := range map[string]run{
				"native":     native,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf(
						"%s: %s: want %s; got %s",
						backend,
						difference,
						optionsBaseShown(want),
						optionsBaseShown(got),
					)
				}
			}
		})
	}
}

// The back-cast at tsc's commandLineParser.ts:3098, a value seen as OptionsBase cast back as CompilerOptions, puts back
// named contracts the view never held, and the cast rule refuses it ahead of the record view. The day it's admitted,
// this goes red, and the reads it admits want pins of their own.
func TestOptionsBaseBackCastRefused(test *testing.T) {
	test.Parallel()
	path := optionsBasePath(test, "back_cast.ts")
	_, loweredError := lowered(test, path)
	const want = ":22:17: Adamic 0.1 refuses an unproven relation from OptionsBase to CompilerOptions: optional field strict has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) || loweredError.Error() != path+want {
		test.Fatalf("want the cast refused, %q, got %v", want, loweredError)
	}
}

// A function over Options, boxed as a union, seen as one over a record of strings: the view drops strict, but the
// index value representations differ, so it stops at the existing storage stop. It's reached through a parameter,
// where no contract is dropped (parameterStorage); with that lifted, the first condition, one representation and
// sameKeeping, stops it alone. Mutant: lift both, and it's admitted there, to be refused later as another rule's
// widening, or read wrong. Outside a parameter, tsc's relation and a union slot leave no mismatch to reach.
func TestOptionsBaseRepresentationMismatchStops(test *testing.T) {
	test.Parallel()
	path := optionsBasePath(test, "representation.ts")
	_, loweredError := lowered(test, path)
	const want = "a (options: Options) => string seen as (options: { [option: string]: string; }) => string (fixed objects and records have different storage; copy explicitly)"
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) || notYet.What != want || notYet.Where != path+":8:7" {
		test.Fatalf("want NotYet %q at :8:7, got %v", want, loweredError)
	}
}

// A named field held as a union, kind?: string | number, is read back unchecked, so the view keeps its contract and
// stops at the existing storage stop. Mutant: drop the third condition's scalar test, and it lowers.
func TestOptionsBaseUncheckedFieldStops(test *testing.T) {
	test.Parallel()
	path := optionsBasePath(test, "unchecked_union.ts")
	_, loweredError := lowered(test, path)
	const want = "a Options seen as OptionsBase (fixed objects and records have different storage; copy explicitly)"
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) || notYet.What != want || notYet.Where != path+":19:22" {
		test.Fatalf("want NotYet %q at :19:22, got %v", want, loweredError)
	}
}

// Language's must (08:38Z): an object-kind field, paths?: MapLike<string[]>, is never dropped, since the native
// union-slot check read an honest MapLike written under it through the view as another member of the slot's union,
// where Node and the JavaScript backend read it (paths_honest.ts, g5 p06). So each of Language's paths programs stops
// not yet at the view, before its write: the honest one and the lies (g5 p01 to p05), which stopped at their reads.
// Mutant: let object kinds be dropped again, and they lower; paths_honest.ts then disagrees with Node natively, which
// the failure shows.
func TestOptionsBaseObjectFieldStops(test *testing.T) {
	test.Parallel()
	const want = "a CompilerOptions seen as OptionsBase (fixed objects and records have different storage; copy explicitly)"
	for name, printed := range map[string]string{
		"paths_honest.ts": "a\n",
		"paths_string.ts": "undefined\n",
		"paths_array.ts":  "string\n1\n",
		"paths_number.ts": "after\n",
		"paths_keys.ts":   "0,1\n",
		"paths_mixed.ts":  "number\n",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := optionsBasePath(test, name)
			node := onNode(test, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				test.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, loweredError := lowered(test, path)
			if loweredError == nil {
				native, _ := natively(test, program)
				test.Fatalf(
					"lowered, want NotYet %q at :19:12; Node %s; native %s; javascript %s",
					want,
					optionsBaseShown(node),
					optionsBaseShown(native),
					optionsBaseShown(onJavaScriptBackend(test, program)),
				)
			}
			var notYet *lower.NotYet
			if !errors.As(loweredError, &notYet) || notYet.What != want || notYet.Where != path+":19:12" {
				test.Fatalf("want NotYet %q at :19:12, got %v", want, loweredError)
			}
		})
	}
}

// Language's should (08:38Z): a function over CompilerOptions seen as one over OptionsBase is handed an OptionsBase,
// so the view through its parameter adds strict. A view drops no contract through a parameter (parameterStorage),
// so each of Language's g4 p01 to p04 stops not yet at the view, as on main, where 419579d6 let the view through to
// the optional-widening refusal. Mutant: leave droppedContractsRead set through parameters, and each is refused
// there again.
func TestOptionsBaseParameterViewStops(test *testing.T) {
	test.Parallel()
	const function = "a (o: CompilerOptions) => string seen as (o: OptionsBase) => string (fixed objects and records have different storage; copy explicitly)"
	for name, stop := range map[string]struct{ what, at string }{
		"parameter_function.ts":   {function, ":20:7"},
		"parameter_bivariance.ts": {function, ":23:27"},
		"parameter_argument.ts":   {function, ":26:25"},
		"parameter_implements.ts": {
			"a StrictReader seen as Reader (fixed objects and records have different storage; copy explicitly)",
			":28:7",
		},
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := optionsBasePath(test, name)
			_, loweredError := lowered(test, path)
			var notYet *lower.NotYet
			if !errors.As(loweredError, &notYet) || notYet.What != stop.what || notYet.Where != path+stop.at {
				test.Fatalf("want NotYet %q at %s, got %v", stop.what, stop.at, loweredError)
			}
		})
	}
}

// Language's g4 p06: a method over CompilerOptions written in a literal where Reader's read takes OptionsBase. A
// literal is made as its slot, so recordStorageView never sees the method, here as on main, and it lowers. Its body's
// named read stops on the "yes" under strict on both backends, so it rests on the read check alone. Stopping it at
// the method is a rule for methods written in literals, which this head doesn't make.
func TestOptionsBaseLiteralMethodStopsAtRead(test *testing.T) {
	test.Parallel()
	path := optionsBasePath(test, "method_parameter.ts")
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatalf("must lower, as on main: %v", loweredError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != "on\n" {
		test.Fatalf("Node: %s", optionsBaseShown(node))
	}
	native, _ := natively(test, program)
	want := run{stderr: []byte(
		"adamic: panic: field read failed: o.strict is not a boolean; expected boolean, found string, which a computed-key or index-signature write may have stored\n",
	), exitCode: 70}
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
		}
	}
}

// The same views in .a stay refused, as #bq3fvdt rules, OptionsBase's and a library Record's alike:
// recordDeclaredKeyRefusal refuses them at 1975, before recordStorageView, whose gate keeps droppedContractRead to .ts
// behind it.
func TestOptionsBaseRefusedInAdamic(test *testing.T) {
	test.Parallel()
	for name, want := range map[string]string{
		"refused.a":        ":19:22: Adamic 0.1 refuses a value with declared properties seen through a wider index signature, which can write under a declared property what it does not hold (CompilerOptions seen as OptionsBase: strict is declared boolean | undefined, and the index signature holds string | number | boolean | undefined); give the record no declared keys, or write the declared property by its name",
		"refused_record.a": ":16:22: Adamic 0.1 refuses a value with declared properties seen through a wider index signature, which can write under a declared property what it does not hold (CompilerOptions seen as Record<string, string | number | boolean | undefined>: strict is declared boolean | undefined, and the index signature holds string | number | boolean | undefined); give the record no declared keys, or write the declared property by its name",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := optionsBasePath(test, name)
			_, loweredError := lowered(test, path)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) || loweredError.Error() != path+want {
				test.Fatalf("want the view refused, %q, got %v", want, loweredError)
			}
		})
	}
}
