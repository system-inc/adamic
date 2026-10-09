package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoUnusedVarsReportsAtTheBindingName asserts WHERE the finding points, which no message-id
// assertion can see. A rule naming the right message over the wrong span passes a complete fixture
// pair while being wrong, and this rule has several candidate spans to get wrong: the declaration
// statement, the declarator, and the name.
func TestNoUnusedVarsReportsAtTheBindingName(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   string
	}{
		{"variable", "const forgotten = 1;", "forgotten"},
		{"parameter", "function f(leftBehind) {} f();", "leftBehind"},
		{"import", "import { unusedThing } from './m';", "unusedThing"},
		{"function declaration", "function neverCalled() {}", "neverCalled"},
		{"interface", "interface Orphan { a: string }", "Orphan"},
		{"catch binding", "try { console.log(1); } catch (caught) {}", "caught"},
		{"destructured element", "const { plucked } = source;", "plucked"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly 1 finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			source := result.SourceFile.Text()
			if got := source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.want {
				t.Errorf("finding points at %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestNoUnusedVarsMessageIdAndDescription asserts the whole message against literals typed here
// rather than against the rule's own builder. Comparing a finding to the value it was reported with
// is equality that cannot fail: both sides move together under mutation.
//
// Three things the message used to get wrong, each pinned: it did not name the binding, it said
// "declared" whether or not a value was ever stored, and it told every finding to "prefix it with an
// underscore" even under the ahra config's `^$` patterns, where an underscore silences nothing. That
// advice reached `_nextContent` in `PensieveBootstrap.test.ts`, a binding already carrying the
// underscore and reported anyway. The escape is now offered only when the configured pattern for
// that kind of binding accepts `_` plus the name.
func TestNoUnusedVarsMessageIdAndDescription(t *testing.T) {
	t.Parallel()

	const body = " and nothing ever reads it. A name that is written but never read is almost always " +
		"the residue of an edit that moved on: an import whose call site was deleted, a parameter left " +
		"behind when a signature changed, a variable holding a value nobody asked for. It costs " +
		"nothing at runtime, which is what lets it accumulate, and it costs the next reader real time, " +
		"because an unused name reads exactly like a used one until you search the file and find " +
		"nothing. "
	const deleteOnly = "Delete it."
	const withEscape = "Delete it, or prefix it with an underscore, which the configured ignore pattern " +
		"accepts, to say the omission is deliberate."

	strict := `{"varsIgnorePattern":"^$","argsIgnorePattern":"^$","caughtErrorsIgnorePattern":"^$"}`
	underscore := `{"varsIgnorePattern":"^_","argsIgnorePattern":"^_","caughtErrorsIgnorePattern":"^_"}`

	for _, testCase := range []struct {
		name    string
		source  string
		options string
		want    string
	}{
		{"an initialized variable is assigned a value", "const forgotten = 1;", "",
			"'forgotten' is assigned a value" + body + deleteOnly},
		{"a parameter is declared", "function f(leftBehind: number) {} f(1);", `{"args":"all"}`,
			"'leftBehind' is declared" + body + deleteOnly},
		{"an uninitialized variable written later is assigned a value",
			"function f(): void {\n\tlet provenance: string;\n\tprovenance = 'a';\n}\nf();", "",
			"'provenance' is assigned a value" + body + deleteOnly},
		{"an import is declared", "import { unusedThing } from './m';", "",
			"'unusedThing' is declared" + body + deleteOnly},
		{"a strict config offers no underscore, even to a name that has one",
			"const files = [{ nextContent: 1, a: 2 }].map(({ nextContent: _nextContent, ...file }) => file);\nconsole.log(files);",
			strict, "'_nextContent' is declared" + body + deleteOnly},
		{"an underscore-accepting config offers the escape", "const forgotten = 1;", underscore,
			"'forgotten' is assigned a value" + body + withEscape},
		{"the escape follows the pattern for the binding's own kind",
			"function f(leftBehind: number) {} f(1);", `{"args":"all","varsIgnorePattern":"^_"}`,
			"'leftBehind' is declared" + body + deleteOnly},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var options any
			if testCase.options != "" {
				decoded, err := DecodeNoUnusedVarsOptions(json.RawMessage(testCase.options))
				if err != nil {
					t.Fatalf("decoding failed: %v", err)
				}
				options = decoded
			}
			result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", testCase.source, options)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Id; got != "noUnusedVars" {
				t.Errorf("message id is %q, want %q", got, "noUnusedVars")
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Errorf("description is\n%q\nwant\n%q", got, testCase.want)
			}
		})
	}
}

// TestNoUnusedVarsReportsAtTheLastWrite pins ESLint's position, which is not the declaration.
//
// Measured on the installed `@typescript-eslint/no-unused-vars` 8.67.0: a binding with writes is
// reported at the identifier of its last write in its own variable scope, so `provenance`, declared
// at 109 and written at 112, 115 and 119 in `FinancePositionCommandLineInterface.ts`, reported at
// 119:9 there and at 109:9 here. A write inside a nested function does not count, and a binding
// whose only write is its initializer still reports at the declaration.
func TestNoUnusedVarsReportsAtTheLastWrite(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   string
	}{
		{"the last of three branch writes",
			"function accountLine(valued: boolean): string {\n  let provenance: string;\n  if (valued) {\n    provenance = 'a';\n  } else {\n    provenance = 'b';\n  }\n  return 'line';\n}\naccountLine(true);",
			"provenance = 'b'"},
		{"a write in a nested function is skipped",
			"let v = 1;\nv = 2;\nfunction g() { v = 3; }\ng();",
			"v = 2"},
		{"an initializer alone reports at the declaration",
			"const forgotten = 1;",
			"forgotten = 1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			source := result.SourceFile.Text()
			start := int(result.Diagnostics[0].Range.Pos())
			end := int(result.Diagnostics[0].Range.End())
			name := source[start:end]
			if !strings.HasPrefix(source[start:], testCase.want) || !strings.HasPrefix(testCase.want, name) {
				t.Errorf("finding points at %q followed by %q, want the %q at the start of %q", name, source[end:min(end+12, len(source))], name, testCase.want)
			}
		})
	}
}

// TestNoUnusedVarsRequiresTheTypedHarness pins that this rule answers nothing without a checker.
//
// A rule declaring NeedsTypeChecker and reading it without a nil guard does not crash on our shim,
// it goes SILENT, and every StaysSilent fixture then passes vacuously. That failure is invisible
// from a green suite, so it gets its own assertion with a control on the other side.
func TestNoUnusedVarsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "const forgotten = 1;"
	if result := rule_testing.Run(t, NoUnusedVars, "a.ts", source); len(result.Diagnostics) != 0 {
		t.Errorf("untyped harness: want silence, got %d findings", len(result.Diagnostics))
	}
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", source); len(result.Diagnostics) != 1 {
		t.Errorf("typed harness control: want 1 finding, got %d", len(result.Diagnostics))
	}
}

// TestDecodeNoUnusedVarsOptions routes through the exported decoder rather than building the
// struct, because the defaults are what the decoder applies and a hand-built struct bypasses every
// one of them. Three of this rule's defaults are not the zero value, and the inventory records the
// rule as having no options at all, so each default is pinned here individually.
func TestDecodeNoUnusedVarsOptions(t *testing.T) {
	t.Parallel()

	t.Run("bare error supplies every default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUnusedVarsOptions(nil)
		if err != nil {
			t.Fatalf("decoding empty options failed: %v", err)
		}
		if decoded.Vars != "all" {
			t.Errorf("vars default is %q, want %q", decoded.Vars, "all")
		}
		if decoded.Args != "after-used" {
			t.Errorf("args default is %q, want %q — NOT \"all\"", decoded.Args, "after-used")
		}
		if decoded.CaughtErrors != "all" {
			t.Errorf("caughtErrors default is %q, want %q", decoded.CaughtErrors, "all")
		}
		if decoded.VarsIgnorePattern != "" || decoded.ArgsIgnorePattern != "" || decoded.CaughtErrorsIgnorePattern != "" {
			t.Error("an absent ignore pattern must stay absent, which ignores nothing, ESLint's default")
		}
	})

	t.Run("configured values override", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUnusedVarsOptions(json.RawMessage(
			`{"args":"none","varsIgnorePattern":"^ignore"}`))
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}
		if decoded.Args != "none" {
			t.Errorf("args is %q, want %q", decoded.Args, "none")
		}
		if decoded.VarsIgnorePattern != "^ignore" {
			t.Errorf("varsIgnorePattern is %q, want %q", decoded.VarsIgnorePattern, "^ignore")
		}
	})
}

// TestNoUnusedVarsIgnoresNoNameByDefault pins ESLint's default through the rule rather than through
// the decoder: with no pattern configured, a leading underscore exempts nothing.
//
// This rule used to take oxc's leading-underscore default. Every case below reports under the
// installed `@typescript-eslint/no-unused-vars` 8.67.0 with no options, measured, and every one was
// silent here before the default was reversed under the parity doctrine.
func TestNoUnusedVarsIgnoresNoNameByDefault(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"const _deliberate = 1;",
		"const _ = 1;",
		"function f(_) {} f();",
		"try {} catch(_) { }",
	} {
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnusedVars, "a.ts", source), "noUnusedVars")
	}

	// The control on the other side: a configured pattern still exempts, so the reports above come
	// from the missing default rather than from a rule that stopped reading patterns.
	options, err := DecodeNoUnusedVarsOptions(json.RawMessage(`{"varsIgnorePattern":"^_","argsIgnorePattern":"^_"}`))
	if err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", "const _deliberate = 1;", options))
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", "function f(_unused: number) {} f(1);", options))
}

// TestNoUnusedVarsJsxFactoryImport pins the exemption that removed 510 false positives from our own
// tree. `import React` in a `.tsx` looks untouched and is used by the JSX transform.
func TestNoUnusedVarsJsxFactoryImport(t *testing.T) {
	t.Parallel()

	const source = "import React from 'react';\nexport const A = 1;\n"
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.tsx", source); len(result.Diagnostics) != 0 {
		t.Errorf("a React import in a .tsx file is exempt; got %d findings", len(result.Diagnostics))
	}
	// The control that makes the case above mean something: the same import in a `.ts` file, where
	// no JSX transform can reach it, still reports.
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", source); len(result.Diagnostics) != 1 {
		t.Errorf("control: the same import in a .ts file should report; got %d",
			len(result.Diagnostics))
	}
}

// TestNoUnusedVarsSeparatesShadowedBindingsBySymbol is the identity test.
//
// Name matching cannot separate these, and it fails in the quiet direction: it sees a read spelled
// like the binding and declines, missing the finding. Both orderings are pinned because a symbol
// can carry more than one declaration and the ordering is not something to assume.
func TestNoUnusedVarsSeparatesShadowedBindingsBySymbol(t *testing.T) {
	t.Parallel()

	// The inner binding is read; the outer one is not. A name-matching rule reports neither.
	result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts",
		"const outer = 1; { const outer = 2; console.log(outer); }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want exactly 1 finding for the unread outer binding, got %d", len(result.Diagnostics))
	}
	source := result.SourceFile.Text()
	if got := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]; got != "outer" {
		t.Errorf("finding points at %q, want the outer binding's name", got)
	}
	if position := result.Diagnostics[0].Range.Pos(); position > 20 {
		t.Errorf("finding is at offset %d, which is the INNER binding; identity was not used", position)
	}
}

// TestNoUnusedVarsDeclarationMergingExportsFromEitherSide pins that the export check asks every
// declaration of a merged symbol rather than the first.
//
// Both orderings, because an index-zero test passes one and fails the other, and which one it fails
// depends on which declaration the checker happens to list first.
func TestNoUnusedVarsDeclarationMergingExportsFromEitherSide(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"interface Merged {\n  bar: string;\n}\nexport const Merged = 'bar';\n",
		"export const Merged = 'bar';\ninterface Merged {\n  bar: string;\n}\n",
	} {
		if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", source); len(result.Diagnostics) != 0 {
			t.Errorf("a merged symbol exported from either side is exempt; got %d findings for %q",
				len(result.Diagnostics), source)
		}
	}
}

// TestNoUnusedVarsSkipsDeclarationFiles pins the `.d.ts` behavior, and says what it does NOT prove.
//
// It asserts the observable answer, which is real: a declaration file reports nothing while the same
// source as `.ts` reports both. It does NOT prove the suffix test in the rule is what produces that,
// and a mutant disabling that test survives this fixture. The cause is recorded at the rule's own
// line: typescript-go marks declarations in a `.d.ts` ambient, so the ambient guard already declines
// them and the suffix test is subsumed today.
//
// Kept as a behavior fixture rather than deleted, and labelled rather than left to imply more than
// it establishes, because a fixture whose name promises to guard a branch it cannot see is worse
// than no fixture: it stops the next reader from checking.
func TestNoUnusedVarsSkipsDeclarationFiles(t *testing.T) {
	t.Parallel()

	const source = "interface Unreferenced {}\ntype AlsoUnreferenced = {};\n"
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.d.ts", source); len(result.Diagnostics) != 0 {
		t.Errorf("a .d.ts declares rather than defines and is skipped whole; got %d findings",
			len(result.Diagnostics))
	}
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", source); len(result.Diagnostics) != 2 {
		t.Errorf("control: the same source as .ts must report both; got %d",
			len(result.Diagnostics))
	}
}

// TestNoUnusedVarsExemptsMappedTypeKeys pins a deliberate divergence from ESLint.
//
// Also a surviving mutant, and for a reason worth recording: the only occurrence of this shape
// anywhere in reach was in our own tree rather than in the corpus, so the case that proves the
// guard matters had to be written from the real-tree finding it removed. Upstream oxc never reports
// it (`mod.rs:388`); ESLint does, measured with its own Linter API.
func TestNoUnusedVarsExemptsMappedTypeKeys(t *testing.T) {
	t.Parallel()

	// NOT exported, and that is the whole point of the case. The first fixture written for this
	// survivor used an exported type and the mutant survived it again: an exported declaration is
	// already exempt for a different reason, so both the guarded and the unguarded rule reach
	// silence by different routes and no fixture over that shape can separate them. Measured on the
	// mutated rule, this input reports `K` and the exported one does not.
	const source = "type Slots<T extends string> = { [K in T]?: string };\n" +
		"const value: Slots<'a'> = {};\nconsole.log(value);\n"
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", source); len(result.Diagnostics) != 0 {
		t.Errorf("a mapped type key is always used by the type it builds; got %d findings",
			len(result.Diagnostics))
	}
	// The control: an ordinary type parameter that genuinely goes unused still reports, so the
	// case above cannot be passing because type parameters are exempt as a class. Not exported,
	// because an exported declaration is exempt for a different reason and would mask this.
	const control = "type Ignores<T extends string> = { a: string };\nconst v: Ignores<'x'> = { a: '' };\nconsole.log(v);\n"
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", control); len(result.Diagnostics) != 1 {
		t.Errorf("control: an unused type parameter outside a mapped type reports; got %d",
			len(result.Diagnostics))
	}
}

// TestNoUnusedVarsExportedContainerDoesNotExemptItsContents pins the change that closed the largest
// group of gaps.
//
// An exported function leaves the file; its parameter list does not, and no importer can read a
// parameter name. The same holds for a type parameter and for anything inside an exported
// namespace's body. Before this, the export climb walked from the binding all the way to the
// enclosing `export` and exempted everything under it.
//
// The controls matter as much as the cases: the exported binding ITSELF must still be exempt, or
// this test would pass on a rule that had simply stopped believing in exports.
func TestNoUnusedVarsExportedContainerDoesNotExemptItsContents(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{"parameter of an exported function", "export function f(x, y) { return x; }", 1},
		{"type parameter of an exported interface", "export interface M<T> { a: string }", 1},
		{"binding inside an exported namespace", "export namespace N { function inner() {} }", 1},
		{"import equals inside an exported namespace",
			"namespace Foo { export const foo = 1; }\nexport namespace Bar { import TheFoo = Foo; }", 1},

		// Controls: the exported thing itself stays exempt, and a local inside an exported arrow
		// is still judged rather than being swept up by the same change.
		{"control, the exported function itself", "export function used() {}", 0},
		{"control, the exported interface itself", "export interface Used { a: string }", 0},
		{"control, exported const", "export const value = 1;", 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			if len(result.Diagnostics) != testCase.want {
				t.Errorf("want %d findings, got %d", testCase.want, len(result.Diagnostics))
			}
		})
	}
}

// TestNoUnusedVarsDiscardedReads pins the three positions where a read's value goes nowhere.
//
// Each has a control differing in exactly the one thing that decides it, because every one of these
// is a left-versus-right or a same-name-versus-different-name distinction where a rule that ignored
// the side would pass a one-sided fixture.
func TestNoUnusedVarsDiscardedReads(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		// A comma sequence yields its RIGHT operand, so only the left one is discarded.
		{"sequence left operand", "let a = 0; let b = (a, 0) + 1; f(b);", 1},
		{"control, sequence right operand", "let a = 0; let b = (0, a) + 1; f(b);", 0},

		// A self-update assigned back into its own binding observes nothing.
		{"update fed back to itself", "let a = 0; a = ++a;", 1},
		{"control, update fed to another binding", "let a = 0; let b = ++a; f(b);", 0},
		{"update through a cast", "let a = 0; a = a++ as any;", 1},

		// A closure assigned to the binding it reads.
		{"closure assigned to itself",
			"function foo(cb) { cb = function(a) { cb(1 + a); }; bar(not_cb); } foo();", 1},
		{"control, immediately invoked so it really runs",
			"function foo(cb) { cb = function(a) { return cb(1 + a); }(); } foo();", 0},
		{"control, parameter default shadows the outer binding",
			"let a; a = function(a = a) {}; a();", 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			if len(result.Diagnostics) != testCase.want {
				t.Errorf("want %d findings, got %d", testCase.want, len(result.Diagnostics))
			}
		})
	}
}

// TestNoUnusedVarsTypePositionsThatNameWithoutReading pins the type-level shapes where a binding is
// mentioned without being used.
func TestNoUnusedVarsTypePositionsThatNameWithoutReading(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{"return type predicate", "export function f(a: unknown): a is string { return true; }", 1},
		{"asserts predicate", "function g(a: unknown): asserts a is string {} g('');", 1},
		{"rest parameter typed by itself", "function h(...args: typeof args) {} h();", 1},
		// Only `R` reports. `T` is genuinely read by the `T extends` on its left, which is worth
		// pinning: the first version of this fixture expected two and the rule was right.
		{"infer binding nothing uses",
			"export type F<T> = T extends infer R ? string : never;", 1},

		// Control: a parameter the body genuinely reads is not reported just because a predicate
		// also names it.
		{"control, predicate over a parameter the body reads",
			"export function f(a: unknown): a is string { return typeof a === 'string'; }", 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			if len(result.Diagnostics) != testCase.want {
				t.Errorf("want %d findings, got %d", testCase.want, len(result.Diagnostics))
			}
		})
	}
}

// TestNoUnusedVarsAmbientModuleExplicitExports pins the split between an ambient block that states
// an interface and one that does not, and the interface-versus-alias split inside an ambient block.
//
// Both are upstream distinctions visible only in its snapshot rather than in its source, and each
// one has a near-identical neighbour falling the other way.
func TestNoUnusedVarsAmbientModuleExplicitExports(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{"ambient block with no exports is left alone",
			"declare module 'm' { type Unused = any; }", 0},
		{"ambient block WITH an export has its contents judged",
			"declare module 'm' { type Unused = any; const x = 1; export = x; }", 1},
		{"ambient interface type parameter is left alone",
			"declare module 'vitest' { interface Matchers<T> { toBeFoo(v: unknown): unknown; } }", 0},
		{"ambient type alias type parameter is judged",
			"declare module 'bun:test' { type Matchers2<T> = {} }", 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			if len(result.Diagnostics) != testCase.want {
				t.Errorf("want %d findings, got %d", testCase.want, len(result.Diagnostics))
			}
		})
	}
}

// TestNoUnusedVarsReExportFromModuleIsNotALocalRead pins the one-word difference between a specifier
// that reads a local binding and one that names something in another module.
func TestNoUnusedVarsReExportFromModuleIsNotALocalRead(t *testing.T) {
	t.Parallel()

	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts",
		"import { resolve } from \"path\";\nexport { resolve } from \"path\";"); len(result.Diagnostics) != 1 {
		t.Errorf("`export { x } from './m'` names the other module, so the import is unused; got %d",
			len(result.Diagnostics))
	}
	// The control is the same file with the `from` clause removed, where the specifier really does
	// read the local binding.
	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts",
		"import { resolve } from \"path\";\nexport { resolve };"); len(result.Diagnostics) != 0 {
		t.Errorf("control: `export { x }` without a module specifier reads the local binding; got %d",
			len(result.Diagnostics))
	}
}

// TestNoUnusedVarsCaughtErrorsHaveNoDefaultIgnorePattern pins that a caught error is reported by
// default and silenced only by its own pattern, not by the variables pattern.
func TestNoUnusedVarsCaughtErrorsHaveNoDefaultIgnorePattern(t *testing.T) {
	t.Parallel()

	if result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", "try {} catch(_) { }"); len(result.Diagnostics) != 1 {
		t.Errorf("caughtErrorsIgnorePattern has no default, so `catch(_)` reports; got %d",
			len(result.Diagnostics))
	}
	// The variables pattern does not reach a caught error.
	varsOnly, err := DecodeNoUnusedVarsOptions(json.RawMessage(`{"varsIgnorePattern":"^_"}`))
	if err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", "try {} catch(_) { }", varsOnly); len(result.Diagnostics) != 1 {
		t.Errorf("varsIgnorePattern does not cover a caught error; got %d", len(result.Diagnostics))
	}
	// And configuring its own pattern turns it off.
	options, err := DecodeNoUnusedVarsOptions(json.RawMessage(`{"caughtErrorsIgnorePattern":"^_"}`))
	if err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", "try {} catch(_) { }", options); len(result.Diagnostics) != 0 {
		t.Errorf("an explicit caughtErrorsIgnorePattern ignores it; got %d", len(result.Diagnostics))
	}
}

// TestNoUnusedVarsReportsAValueUsedOnlyAsAType covers typescript-eslint's `usedOnlyAsType`: a value
// whose only references are `typeof` type queries is unused, because a type query reads nothing at
// runtime. The reporting cases are typescript-eslint's own invalid cases for that message, and the
// silent ones are the controls that keep the change from reaching too far. All measured on the
// installed plugin, 8.67.0.
//
// Not covered, recorded on #a2vq6d3: upstream's three cases where a value merges with an interface of
// the same name and is referenced only as that type.
func TestNoUnusedVarsReportsAValueUsedOnlyAsAType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		ids    []string
	}{
		{"typeof", "const foo: number = 1;\n\nexport type Foo = typeof foo;\n", []string{"usedOnlyAsType"}},
		{"typeof inside a union", "const foo: number = 1;\n\nexport type Foo = typeof foo | string;\n", []string{"usedOnlyAsType"}},
		{"typeof inside an intersection", "const foo: number = 1;\n\nexport type Foo = (typeof foo | string) & { __brand: 'foo' };\n", []string{"usedOnlyAsType"}},
		{"typeof of a member", "const foo = {\n  bar: {\n    baz: 123,\n  },\n};\n\nexport type Bar = typeof foo.bar;\n", []string{"usedOnlyAsType"}},
		{"an indexed typeof", "const foo = {\n  bar: {\n    baz: 123,\n  },\n};\n\nexport type Bar = (typeof foo)['bar'];\n", []string{"usedOnlyAsType"}},
		{"a parameter read only by its own return type", "export const myTypeGuard2 = (data2: unknown): typeof data2 => {\n  return true;\n};\n", []string{"usedOnlyAsType"}},
		{"a type predicate's parameter", "export const myTypeGuard = (data: unknown): data is string => {\n  return true;\n};\n", []string{"usedOnlyAsType"}},
		{"keyof typeof", "const defaults = { a: 1 };\nexport type Key = keyof typeof defaults;\n", []string{"usedOnlyAsType"}},
		{"a function behind ReturnType", "function make() { return 1; }\nexport type Made = ReturnType<typeof make>;\n", []string{"usedOnlyAsType"}},

		// A value read anywhere is read, and an export is a use.
		{"read as a value too", "const defaults = { a: 1 };\nexport type Defaults = typeof defaults;\nconsole.log(defaults);\n", nil},
		{"exported", "export const defaults = { a: 1 };\nexport type Defaults = typeof defaults;\n", nil},
		// A type-only import can only ever be read by a type query, so the query is its use.
		{"a type-only import read by typeof", "import type { foo } from 'foo';\nexport type Foo = typeof foo;\n", nil},
		// A value import read only by typeof is consistent-type-imports' finding, so this rule stays
		// out of it, as upstream does.
		{"a value import read only by typeof", "import { foo } from 'foo';\nexport type Foo = typeof foo;\n", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "/repository/source/TypeOnly.ts", testCase.source)
			if len(testCase.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestNoUnusedVarsJudgesOverrideParametersAsTypeScriptEslintDoes holds #c6jhg93: an override method's
// unused trailing parameter reports, and a later defaulted parameter shields the ones before it.
// Each row's expected names were measured with lintText on the installed typescript-eslint.
func TestNoUnusedVarsJudgesOverrideParametersAsTypeScriptEslintDoes(t *testing.T) {
	t.Parallel()

	base := "export class Base {\n    shouldRun(previous: number | null): boolean { return previous === null; }\n" +
		"    create(items: string[], discounts: number[], options: object = {}): number { return items.length + discounts.length + Object.keys(options).length; }\n}\n"
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		// DailyMetricsScheduledExecutable.ts:47 and SystemLogReportScheduledExecutable.ts:74.
		{"an override's unused only parameter", base + "export class Child extends Base {\n    override shouldRun(_previousRun: number | null): boolean { return true; }\n}\n", []string{"_previousRun"}},
		// FakeStripePaymentProcessor.ts:122: the defaulted _options reports and shields _appliedDiscount.
		{"an override's trailing defaulted parameter", base + "export class Child extends Base {\n    override create(items: string[], _appliedDiscount: number[], _options: object = {}): number { return items.length; }\n}\n", []string{"_options"}},
		// oxc's clean case, withheld from the corpus because typescript-eslint reports it.
		{"oxc's override case", "class Foo {\n    public method(a: number, b: number): number { return a + b; }\n}\nclass Bar extends Foo {\n    public override method(a: number, b: number): number { return a; }\n}\nnew Bar();\n", []string{"b"}},
		// The shield, both ways: a default after b shields b; without the default, both report.
		{"a defaulted parameter shields the one before it", "export function run(a: number, b: string, c: object = {}): number { return a; }\n", []string{"c"}},
		{"without the default, both trailing parameters report", "export function run(a: number, b: string, c: object): number { return a; }\n", []string{"b", "c"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedVars, "a.ts", testCase.source)
			var got []string
			for _, diagnostic := range result.Diagnostics {
				got = append(got, testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()])
			}
			if strings.Join(got, ",") != strings.Join(testCase.want, ",") {
				t.Fatalf("reported %v, want %v", got, testCase.want)
			}
		})
	}
}

// TestNoUnusedVarsTreatsADecoratedClassOrParameterAsUsed pins a deliberate divergence from ESLint, by
// @system_cohere's ruling of 2026-10-03 (#ynneze5): a decorator runs when its class is defined, so a
// class or a parameter that carries one is used whatever reads its name. The shapes are Base's tests,
// which declare a class only so its decorators register, or a parameter only so a parameter decorator
// records it. ESLint reports both.
//
// Each exempt row has a neighbour that must still report, so a mutant widening the arm to every class,
// every parameter, or a decorator on a sibling fails here.
func TestNoUnusedVarsTreatsADecoratedClassOrParameterAsUsed(t *testing.T) {
	t.Parallel()

	decorators := "declare const registry: string[];\n" +
		"function Table(name: string): ClassDecorator { return () => { registry.push(name); }; }\n" +
		"function Column(): PropertyDecorator { return () => {}; }\n" +
		"function Argument(name: string): ParameterDecorator { return () => { registry.push(name); }; }\n" +
		"export { Table, Column, Argument };\n"
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{"a class declared only so its class decorator runs", decorators + "export function check(): void {\n    @Table('conflict')\n    class TwoFlags {}\n}\n", nil},
		{"the same class without the decorator still reports", decorators + "export function check(): void {\n    class TwoFlags {}\n}\n", []string{"TwoFlags"}},
		{"a decorator on a member does not make the class used", decorators + "export function check(): void {\n    class TwoFlags {\n        @Column() declare a: string;\n    }\n}\n", []string{"TwoFlags"}},
		{"a parameter declared only so its decorator records it", decorators + "export class Resolver {\n    tickets(@Argument('pagination') pagination: number): string { return 'ok'; }\n}\n", nil},
		{"the same parameter without the decorator still reports", decorators + "export class Resolver {\n    tickets(pagination: number): string { return 'ok'; }\n}\n", []string{"pagination"}},
		{"a decorated parameter does not exempt the undecorated one after it", decorators + "export class Resolver {\n    tickets(@Argument('pagination') pagination: number, extra: string): string { return 'ok'; }\n}\n", []string{"extra"}},
		{"a decorated parameter shields the undecorated ones before it", decorators + "export class Resolver {\n    tickets(input: string, first: number, @Argument('context') context: object): string { return 'ok'; }\n}\n", nil},
		{"under args all, the undecorated ones before it still report", decorators + "export class Resolver {\n    tickets(input: string, @Argument('context') context: object): string { return 'ok'; }\n}\n", []string{"input"}},
		{"a decorated parameter is used even under args all", decorators + "export class Resolver {\n    tickets(@Argument('pagination') pagination: number, used: string): string { return used; }\n}\n", nil},
		// The decorator records the parameter's position, which a name inside its pattern does not hold, so
		// that name is still judged, as ESLint judges it; the pattern still holds the position (#s48y8eg).
		{"a name destructured out of a decorated parameter is still judged", decorators + "export class Resolver {\n    tickets(@Argument('page') { size }: { size: number }): string { return 'ok'; }\n}\n", []string{"size"}},
		{"a decorated destructured parameter shields the ones before it", decorators + "export class Resolver {\n    tickets(input: string, @Argument('page') { size }: { size: number }): string { return 'ok'; }\n}\n", []string{"size"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := any(nil)
			if strings.Contains(testCase.name, "args all") {
				options = NoUnusedVarsOptions{Args: "all"}
			}
			result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", testCase.source, options)
			var got []string
			for _, diagnostic := range result.Diagnostics {
				got = append(got, testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()])
			}
			if strings.Join(got, ",") != strings.Join(testCase.want, ",") {
				t.Fatalf("reported %v, want %v", got, testCase.want)
			}
		})
	}
}

// ignoreRestSiblings keeps a binding alive beside a rest element, and only there (#6qxs15e).
//
// cohere declared the option and never read it, so a config turning it on changed nothing and every
// `const { removed, ...rest } = o` reported. Every row was replayed through the installed
// `@typescript-eslint/no-unused-vars` 8.71.0 with the option true and false, and both columns are
// asserted, so neither direction can drift: the option has to silence exactly the rows it silences
// upstream, and silence nothing when off. Upstream's id `unusedVar` is this rule's `noUnusedVars`.
//
// Findings in the prelude are left out of the comparison. It is `declare` statements, which this
// rule exempts as ambient and ESLint's Linter reports, a difference that is not this option's.
func TestNoUnusedVarsIgnoreRestSiblings(t *testing.T) {
	t.Parallel()

	const prelude = "declare const o: Record<string, number>;\ndeclare const n: Record<string, Record<string, number>>;\ndeclare const list: number[];\ndeclare const list2: Record<string, number>[];\ndeclare const key: string;\ndeclare function use(...values: unknown[]): void;\n"
	cases := []struct {
		body            string
		wantWhenIgnored []string
		wantOtherwise   []string
	}{
		{"const { a, ...rest } = o;\nuse(rest);", nil, []string{"unusedVar a"}},
		{"const { a: b, ...rest } = o;\nuse(rest);", nil, []string{"unusedVar b"}},
		{"const { a = 1, ...rest } = o;\nuse(rest);", []string{"unusedVar a"}, []string{"unusedVar a"}},
		{"const { a: b = 1, ...rest } = o;\nuse(rest);", []string{"unusedVar b"}, []string{"unusedVar b"}},
		{"const { a, ...rest } = o;", []string{"unusedVar rest"}, []string{"unusedVar a", "unusedVar rest"}},
		{"const { a, b: { c }, ...rest } = n;\nuse(rest);", []string{"unusedVar c"}, []string{"unusedVar a", "unusedVar c"}},
		{"const { x: { a, ...inner }, ...outer } = n;\nuse(inner, outer);", nil, []string{"unusedVar a"}},
		{"const { a, b } = o;\nuse(a);", []string{"unusedVar b"}, []string{"unusedVar b"}},
		{"const [a, ...rest] = list;\nuse(rest);", []string{"unusedVar a"}, []string{"unusedVar a"}},
		{"let a;\nlet rest;\n({ a, ...rest } = o);\nuse(rest);", nil, []string{"unusedVar a"}},
		{"let a;\nlet rest;\n({ x: a, ...rest } = o);\nuse(rest);", nil, []string{"unusedVar a"}},
		{"let a;\nlet rest;\n({ a = 1, ...rest } = o);\nuse(rest);", []string{"unusedVar a"}, []string{"unusedVar a"}},
		{"let a;\nlet rest;\n({ x: (a), ...rest } = o);\nuse(rest);", nil, []string{"unusedVar a"}},
		{"let a;\nlet rest;\n({ a, rest } = o);\nuse(rest);", []string{"unusedVar a"}, []string{"unusedVar a"}},
		{"function f({ a, ...rest }: Record<string, number>) {\n  return rest;\n}\nuse(f);", nil, []string{"unusedVar a"}},
		{"for (const { a, ...rest } of list2) {\n  use(rest);\n}", nil, []string{"unusedVar a"}},
		{"const { a, ...rest } = o;\ntype A = typeof a;\nuse(rest);\nexport type { A };", nil, []string{"usedOnlyAsType a"}},
		{"const { [key]: a, ...rest } = o;\nuse(rest);", nil, []string{"unusedVar a"}},
		{"const { 'quoted': a, ...rest } = o;\nuse(rest);", nil, []string{"unusedVar a"}},
	}

	for _, testCase := range cases {
		for _, ignored := range []bool{true, false} {
			want := testCase.wantOtherwise
			if ignored {
				want = testCase.wantWhenIgnored
			}
			name := testCase.body
			if ignored {
				name = "ignored: " + name
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				options, err := DecodeNoUnusedVarsOptions(json.RawMessage(
					`{"ignoreRestSiblings":` + map[bool]string{true: "true", false: "false"}[ignored] + `}`))
				if err != nil {
					t.Fatalf("decoding: %v", err)
				}
				source := prelude + testCase.body
				result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", source, options)
				var got []string
				for _, diagnostic := range result.Diagnostics {
					if diagnostic.Range.Pos() < len(prelude) {
						continue
					}
					id := diagnostic.Message.Id
					if id == "noUnusedVars" {
						id = "unusedVar"
					}
					got = append(got, id+" "+source[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
				if strings.Join(got, ",") != strings.Join(want, ",") {
					t.Fatalf("reported %v, want %v", got, want)
				}
			})
		}
	}
}

// The fix through the harness's own check, which applies every fix and refuses an overlap, on two rows of
// the corpus whose fixes do not overlap: one specifier beside a used one, and a whole declaration with its
// line. The corpus above applies ESLint's one pass itself, which this does not trust it to.
func TestNoUnusedVarsAutofixRemovalWritesUpstreamsEdit(t *testing.T) {
	t.Parallel()

	options, err := DecodeNoUnusedVarsOptions(json.RawMessage(`{"enableAutofixRemoval": {"imports": true}}`))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{"one specifier beside a used one", "import { Unused, Used } from 'module';\nexport { Used };\n", "import {  Used } from 'module';\nexport { Used };\n"},
		{"the whole declaration and its line", "import { Unused } from 'module';\nexport {};\n", "export {};\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoUnusedVars, map[string]string{"file.ts": testCase.source}, "file.ts", options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(noUnusedVarsOptionsCorpusTsconfig), 0o644); err != nil {
						t.Fatal(err)
					}
				})
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}

	// The default proposes no fix, only the suggestion, so nothing is rewritten unattended.
	defaults, err := DecodeNoUnusedVarsOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoUnusedVars, map[string]string{"file.ts": "import { Unused } from 'module';\nexport {};\n"}, "file.ts", defaults,
		func(directory string) {
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(noUnusedVarsOptionsCorpusTsconfig), 0o644); err != nil {
				t.Fatal(err)
			}
		})
	if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 1 {
		t.Fatalf("the default gives %+v, want one finding with one suggestion and no fix", result.Diagnostics)
	}
}
