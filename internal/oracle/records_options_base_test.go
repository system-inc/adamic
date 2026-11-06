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
		"own_keys.ts": "target,strict,outDir,lib,configFile\nlib=object outDir=string out strict=boolean false target=number 9\n" +
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
// index value representations differ, so it stops at the existing storage stop. Mutant: drop the first condition,
// one representation and sameKeeping, and it's admitted there, to be refused later as another rule's widening, or
// read wrong.
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
// stops at the existing storage stop. Mutant: drop the third condition's readGuardedInUnionSlot test, and it lowers.
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
