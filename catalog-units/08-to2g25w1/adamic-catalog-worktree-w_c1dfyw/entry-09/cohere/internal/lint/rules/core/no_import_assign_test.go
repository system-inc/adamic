package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// importAssignFile is where the fixtures pretend to live.
//
// A real path matters because this rule reads the checker, so the harness builds an actual program
// and the file has to sit somewhere a tsconfig can reach. The corpus imports from 'mod', which
// resolves to nothing; that is fine and deliberate. An unresolved module still binds its local
// names, which is the only thing this rule asks about.
const importAssignFile = "/repository/source/ImportAssign.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_import_assign.rs`:
// one Tester block, 57 pass and 59 fail. The snapshot records 59 diagnostics from those 59 inputs,
// so exactly one finding per failing input, and the extractor reported no discrepancy before any
// code was written.
//
// The second column of each row is the source text the finding must cover, read out of the oxc
// snapshot's underline rather than guessed. That column is the whole reason this file is longer than
// a list of message ids: upstream points at three different nodes depending on the shape, and a rule
// that pointed at the identifier every time would pass every message-id assertion while being wrong
// on twelve of these.
func TestNoImportAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpans  []string
	}{
		{"import mod1 from 'mod'; mod1 = 0", []string{"mod1"}},
		{"import mod2 from 'mod'; mod2 += 0", []string{"mod2"}},
		{"import mod3 from 'mod'; mod3++", []string{"mod3"}},
		{"import mod4 from 'mod'; for (mod4 in foo);", []string{"mod4"}},
		{"import mod5 from 'mod'; for (mod5 of foo);", []string{"mod5"}},
		{"import mod6 from 'mod'; [mod6] = foo", []string{"mod6"}},
		{"import mod7 from 'mod'; [mod7 = 0] = foo", []string{"mod7"}},
		{"import mod8 from 'mod'; [...mod8] = foo", []string{"mod8"}},
		{"import mod9 from 'mod'; ({ bar: mod9 } = foo)", []string{"mod9"}},
		{"import mod10 from 'mod'; ({ bar: mod10 = 0 } = foo)", []string{"mod10"}},
		{"import mod11 from 'mod'; ({ ...mod11 } = foo)", []string{"mod11"}},
		{"import {named1} from 'mod'; named1 = 0", []string{"named1"}},
		{"import {named2} from 'mod'; named2 += 0", []string{"named2"}},
		{"import {named3} from 'mod'; named3++", []string{"named3"}},
		{"import {named4} from 'mod'; for (named4 in foo);", []string{"named4"}},
		{"import {named5} from 'mod'; for (named5 of foo);", []string{"named5"}},
		{"import {named6} from 'mod'; [named6] = foo", []string{"named6"}},
		{"import {named7} from 'mod'; [named7 = 0] = foo", []string{"named7"}},
		{"import {named8} from 'mod'; [...named8] = foo", []string{"named8"}},
		{"import {named9} from 'mod'; ({ bar: named9 } = foo)", []string{"named9"}},
		{"import {named10} from 'mod'; ({ bar: named10 = 0 } = foo)", []string{"named10"}},
		{"import {named11} from 'mod'; ({ ...named11 } = foo)", []string{"named11"}},
		{"import {named12 as foo} from 'mod'; foo = 0; named12 = 0", []string{"foo"}},
		{"import * as mod1 from 'mod'; mod1 = 0", []string{"mod1"}},
		{"import * as mod2 from 'mod'; mod2 += 0", []string{"mod2"}},
		{"import * as mod3 from 'mod'; mod3++", []string{"mod3"}},
		{"import * as mod4 from 'mod'; for (mod4 in foo);", []string{"mod4"}},
		{"import * as mod5 from 'mod'; for (mod5 of foo);", []string{"mod5"}},
		{"import * as mod6 from 'mod'; [mod6] = foo", []string{"mod6"}},
		{"import * as mod7 from 'mod'; [mod7 = 0] = foo", []string{"mod7"}},
		{"import * as mod8 from 'mod'; [...mod8] = foo", []string{"mod8"}},
		{"import * as mod9 from 'mod'; ({ bar: mod9 } = foo)", []string{"mod9"}},
		{"import * as mod10 from 'mod'; ({ bar: mod10 = 0 } = foo)", []string{"mod10"}},
		{"import * as mod11 from 'mod'; ({ ...mod11 } = foo)", []string{"mod11"}},
		{"import * as mod1 from 'mod'; mod1.named = 0", []string{"mod1.named"}},
		{"import * as mod2 from 'mod'; mod2.named += 0", []string{"mod2.named"}},
		{"import * as mod3 from 'mod'; mod3.named++", []string{"mod3.named"}},
		{"import * as mod4 from 'mod'; for (mod4.named in foo);", []string{"mod4.named"}},
		{"import * as mod5 from 'mod'; for (mod5.named of foo);", []string{"mod5.named"}},
		{"import * as mod6 from 'mod'; [mod6.named] = foo", []string{"mod6.named"}},
		{"import * as mod7 from 'mod'; [mod7.named = 0] = foo", []string{"mod7.named"}},
		{"import * as mod8 from 'mod'; [...mod8.named] = foo", []string{"mod8.named"}},
		{"import * as mod9 from 'mod'; ({ bar: mod9.named } = foo)", []string{"mod9.named"}},
		{"import * as mod10 from 'mod'; ({ bar: mod10.named = 0 } = foo)", []string{"mod10.named"}},
		{"import * as mod11 from 'mod'; ({ ...mod11.named } = foo)", []string{"mod11.named"}},
		{"import * as mod12 from 'mod'; delete mod12.named", []string{"mod12.named"}},
		{"import * as mod from 'mod'; Object.assign(mod, obj)", []string{"mod"}},
		{"import * as mod from 'mod'; Object.defineProperty(mod, key, d)", []string{"mod"}},
		{"import * as mod from 'mod'; Object.defineProperties(mod, d)", []string{"mod"}},
		{"import * as mod from 'mod'; Object.setPrototypeOf(mod, proto)", []string{"mod"}},
		{"import * as mod from 'mod'; Object.freeze(mod)", []string{"mod"}},
		{"import * as mod from 'mod'; Reflect.defineProperty(mod, key, d)", []string{"mod"}},
		{"import * as mod from 'mod'; Reflect.deleteProperty(mod, key)", []string{"mod"}},
		{"import * as mod from 'mod'; Reflect.set(mod, key, value)", []string{"mod"}},
		{"import * as mod from 'mod'; Reflect.setPrototypeOf(mod, proto)", []string{"mod"}},
		{"import mod, * as mod_ns from 'mod'; mod.prop = 0; mod_ns.prop = 0", []string{"mod_ns.prop"}},
		{"import * as mod from 'mod'; Object?.defineProperty(mod, key, d)", []string{"mod"}},
		{"import * as mod from 'mod'; (Object?.defineProperty)(mod, key, d)", []string{"mod"}},
		{"import * as mod from 'mod'; delete mod?.prop", []string{"mod?.prop"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoImportAssign, importAssignFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantSpans))
			for index := range wantIds {
				wantIds[index] = "noImportAssign"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			if len(result.Diagnostics) != len(testCase.wantSpans) {
				return
			}
			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}
		})
	}
}

// The clean cases carry the whole discrimination, and there are 57 of them against 59 failures.
//
// That ratio is the rule. Nearly every failing shape has a clean twin whose text differs by one
// character: `mod1 = 0` fails and `mod.prop = 0` passes, `mod1.named = 0` fails and
// `mod.named.prop = 0` passes, `Object.assign(mod, obj)` fails and `Object.assign(mod.prop, obj)`
// passes. Each pair was added upstream when somebody hit the bug, and each is a place this port
// could be silently wide.
func TestNoImportAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"import mod from 'mod'; mod.prop = 0",
		"import mod from 'mod'; mod.prop += 0",
		"import mod from 'mod'; mod.prop++",
		"import mod from 'mod'; delete mod.prop",
		"import mod from 'mod'; for (mod.prop in foo);",
		"import mod from 'mod'; for (mod.prop of foo);",
		"import mod from 'mod'; [mod.prop] = foo;",
		"import mod from 'mod'; [...mod.prop] = foo;",
		"import mod from 'mod'; ({ bar: mod.prop } = foo);",
		"import mod from 'mod'; ({ ...mod.prop } = foo);",
		"import {named} from 'mod'; named.prop = 0",
		"import {named} from 'mod'; named.prop += 0",
		"import {named} from 'mod'; named.prop++",
		"import {named} from 'mod'; delete named.prop",
		"import {named} from 'mod'; for (named.prop in foo);",
		"import {named} from 'mod'; for (named.prop of foo);",
		"import {named} from 'mod'; [named.prop] = foo;",
		"import {named} from 'mod'; [...named.prop] = foo;",
		"import {named} from 'mod'; ({ bar: named.prop } = foo);",
		"import {named} from 'mod'; ({ ...named.prop } = foo);",
		"import * as mod from 'mod'; mod.named.prop = 0",
		"import * as mod from 'mod'; mod.named.prop += 0",
		"import * as mod from 'mod'; mod.named.prop++",
		"import * as mod from 'mod'; delete mod.named.prop",
		"import * as mod from 'mod'; for (mod.named.prop in foo);",
		"import * as mod from 'mod'; for (mod.named.prop of foo);",
		"import * as mod from 'mod'; [mod.named.prop] = foo;",
		"import * as mod from 'mod'; [...mod.named.prop] = foo;",
		"import * as mod from 'mod'; ({ bar: mod.named.prop } = foo);",
		"import * as mod from 'mod'; ({ ...mod.named.prop } = foo);",
		"import * as mod from 'mod'; obj[mod] = 0",
		"import * as mod from 'mod'; obj[mod.named] = 0",
		"import * as mod from 'mod'; for (var foo in mod.named);",
		"import * as mod from 'mod'; for (var foo of mod.named);",
		"import * as mod from 'mod'; [bar = mod.named] = foo;",
		"import * as mod from 'mod'; ({ bar = mod.named } = foo);",
		"import * as mod from 'mod'; ({ bar: baz = mod.named } = foo);",
		"import * as mod from 'mod'; ({ [mod.named]: bar } = foo);",
		"import * as mod from 'mod'; var obj = { ...mod.named };",
		"import * as mod from 'mod'; var obj = { foo: mod.named };",
		"import mod from 'mod'; { let mod = 0; mod = 1 }",
		"import * as mod from 'mod'; { let mod = 0; mod = 1 }",
		"import * as mod from 'mod'; { let mod = 0; mod.named = 1 }",
		"import {} from 'mod'",
		"import 'mod'",
		"import mod from 'mod'; Object.assign(mod, obj);",
		"import {named} from 'mod'; Object.assign(named, obj);",
		"import * as mod from 'mod'; Object.assign(mod.prop, obj);",
		"import * as mod from 'mod'; Object.assign(obj, mod, other);",
		"import * as mod from 'mod'; Object[assign](mod, obj);",
		"import * as mod from 'mod'; Object.getPrototypeOf(mod);",
		"import * as mod from 'mod'; Reflect.set(obj, key, mod);",
		"import * as mod from 'mod'; { var Object; Object.assign(mod, obj); }",
		"import * as mod from 'mod'; var Object; Object.assign(mod, obj);",
		"import * as mod from 'mod'; Object.seal(mod, obj)",
		"import * as mod from 'mod'; Object.preventExtensions(mod)",
		"import * as mod from 'mod'; Reflect.preventExtensions(mod)",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoImportAssign, importAssignFile, sourceText))
		})
	}
}

// Cases upstream does not cover, each with the reason it exists.
//
// Upstream's corpus is generated against `get_resolved_references`, an index that already answered
// the resolution question, so it never had to prove that resolution was being asked correctly. This
// port asks it directly and these are the shapes where asking it wrongly still passes all 116 cases
// above.
func TestNoImportAssignFiresOnCasesUpstreamOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		// The shorthand destructuring target. `GetSymbolAtLocation` on this identifier answers with
		// the object property's own symbol rather than the value being written, which reads exactly
		// like a correctly declined shadow, so a port using only the plain accessor stays silent
		// here and looks right. Upstream's corpus has `({ bar: named9 } = foo)` but never the bare
		// shorthand, so nothing above catches it.
		// An import inside a `declare module` block binds there too, and the rule has always anchored
		// it: the listener used to fire on every import declaration in the tree. Anchoring reads
		// statements and module blocks rather than the whole tree (#9gcv6bf), and this pins that a
		// module block is still read. A statement in an ambient block is a compile error, which is why
		// neither corpus has one, and the rule reports the write all the same.
		{"an import inside a declare module block",
			"declare module 'x' { import named from 'mod'; named = 0; }", []string{"named"}},

		// An import inside a function or a bare block is a compile error (TS1232), and the parser
		// recovers it as an ordinary import declaration. Upstream reports the write: measured on the
		// installed ESLint 10 with @typescript-eslint/parser, both of these give "'x' is read-only."
		// The rule judged only imports among a file's or a module block's statements until #fcac58b
		// moved it onto the shared walk, which reaches these too, and that matches upstream (#q488ss2).
		{"an import recovered inside a function",
			"function f() { import x from 'y'; x = 1; }", []string{"x"}},
		{"an import recovered inside a bare block",
			"{ import x from 'y'; x = 1; }", []string{"x"}},

		{"a shorthand destructuring target on a named import",
			"import {named} from 'mod'; ({named} = foo)", []string{"named"}},
		{"a shorthand destructuring target on a default import",
			"import mod from 'mod'; ({mod} = foo)", []string{"mod"}},

		// A parenthesis directly around a namespace member write. The climb passes through the
		// paren and then compared a wrapped child against an unwrapped left, so every one of these
		// read as a read. All four throw a TypeError at runtime, which is the whole subject of the
		// rule: a module namespace object is sealed and its properties are non-writable.
		//
		// Nothing in either corpus covers them, and it is structural rather than an oversight:
		// parentheses are a real node in typescript-go and absent from ESTree, so no imported
		// fixture can contain one. `no-class-assign` hit the identical defect on a plain assignment
		// target and fixed it the same way; this rule inherited that climb before the fix and kept
		// it after that rule moved onto the shelf.
		//
		// The span is the member expression rather than the parenthesis, matching the unwrapped
		// spelling, which is why these assert spans rather than counts.
		{"a parenthesized namespace member write",
			"import * as mod from 'mod'; (mod.named) = 0", []string{"mod.named"}},
		{"a doubly parenthesized namespace member write",
			"import * as mod from 'mod'; ((mod.named)) = 0", []string{"mod.named"}},
		{"a parenthesized computed namespace member write",
			"import * as mod from 'mod'; (mod['named']) = 0", []string{"mod['named']"}},
		{"a parenthesized delete of a namespace member",
			"import * as mod from 'mod'; delete (mod.named)", []string{"mod.named"}},

		// Two imports in one file. Cheap coverage that the anchor set is per declaration.
		{"one write with two named imports in the file",
			"import {a} from 'x'; import {b} from 'y'; b = 0", []string{"b"}},
		{"one namespace member write with two namespace imports in the file",
			"import * as a from 'x'; import * as b from 'y'; b.p = 0", []string{"b.p"}},

		// The input that separates node identity from declaration kind, and the only shape that
		// does once the name pre-filter is in place.
		//
		// A kind comparison needs one listener to see an identifier that passes its own name filter
		// while resolving to a different import declaration of the same kind. That needs one local
		// name bound by two import declarations, which the two-imports cases above cannot produce
		// because their names differ. A type-only import and a value import may both bind `T`, since
		// they occupy different declaration spaces, and both declare at `KindImportSpecifier`.
		//
		// Measured: a rule comparing kinds reports this twice, once per listener. Identity reports it
		// once. Without this row that mutation survives the entire corpus and every other case here,
		// which is how it was found.
		{"a value import and a type-only import binding the same local name",
			"import type {T} from 'x'; import {T} from 'y'; T = 0", []string{"T"}},

		// Element access on a namespace. Upstream's own code routes this through its computed-member
		// arm and its corpus never exercises it, so the shape is deliberate upstream and untested
		// upstream.
		{"a bracketed property write on a namespace import",
			"import * as mod from 'mod'; mod['named'] = 0", []string{"mod['named']"}},

		// A write that precedes the import textually. Imports hoist, so this is the same binding and
		// upstream would report it; every one of its 59 failing inputs happens to write after the
		// declaration, so nothing above proves the walk covers the whole file.
		{"a write above the import declaration", "mod = 0; import mod from 'mod';",
			[]string{"mod"}},

		// Two writes to one import. The corpus has exactly one finding per input, so nothing above
		// proves the walk does not stop at the first.
		{"two writes to the same import", "import mod from 'mod'; mod = 0; mod = 1",
			[]string{"mod", "mod"}},

		// A logical assignment. `reference.WritesToBinding` accepts any assignment operator and the
		// corpus only ever uses `=` and `+=`.
		{"a logical assignment to an import", "import mod from 'mod'; mod ||= 0",
			[]string{"mod"}},

		// A prefix update. The corpus has `mod3++` and never `--mod`.
		{"a prefix decrement of an import", "import mod from 'mod'; --mod", []string{"mod"}},

		// A write from inside a nested function. Every corpus write sits at the top level, so
		// nothing above proves the walk descends into bodies.
		{"a write from inside a function body",
			"import mod from 'mod'; function f() { mod = 0; }", []string{"mod"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoImportAssign, importAssignFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantSpans))
			for index := range wantIds {
				wantIds[index] = "noImportAssign"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			if len(result.Diagnostics) != len(testCase.wantSpans) {
				return
			}
			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}
		})
	}
}

// Clean cases upstream does not cover, each with the reason it exists.
func TestNoImportAssignStaysSilentOnCasesUpstreamOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The coordinator's distinguishing input for the kind-versus-identity question, in the
		// direction that costs a false positive. A local shadows an imported name inside a block;
		// both resolve to something and only node identity separates them.
		{"a block-local shadowing a named import",
			"import { A } from './M'; { let A = 1; A = 2; }"},
		{"a block-local shadowing a named import, written through a member",
			"import { A } from './M'; { let A = { p: 0 }; A.p = 2; }"},

		// A parameter shadowing an import. Upstream shadows with `let` in a block and never with a
		// parameter, and the two reach different checker paths.
		{"a parameter shadowing a namespace import",
			"import * as mod from 'mod'; function f(mod) { mod.named = 0; }"},

		// Reading an import in every position that is not a write. The corpus proves member writes
		// and mutation-function arguments; it never proves a plain read stays clean, so a rule that
		// reported on symbol identity alone would pass everything above and fail here.
		{"an import passed to a call", "import mod from 'mod'; foo(mod);"},
		{"an import read on the right of an assignment", "import mod from 'mod'; foo = mod;"},
		{"a namespace read on the right of an assignment", "import * as mod from 'mod'; foo = mod;"},
		{"a namespace property read", "import * as mod from 'mod'; foo = mod.named;"},

		// A named import shadowed by nothing but written through a member. `import {named}` is not a
		// namespace, so `named.prop = 0` is upstream's first clean case; the bracketed form is the
		// same judgment and upstream has no fixture for it.
		{"a bracketed property write on a named import",
			"import {named} from 'mod'; named['prop'] = 0"},

		// `Object.assign` on a non-namespace import. Upstream covers this for default and named
		// imports at the top level; nothing proves the namespace gate is read per binding when the
		// same declaration binds both, which `import mod, * as ns` does.
		{"a mutation function on the default half of a mixed import",
			"import mod, * as ns from 'mod'; Object.assign(mod, obj);"},

		// A shadowed `Reflect`. Upstream shadows `Object` twice and never `Reflect`, so nothing
		// above proves the global check is asked for both names.
		{"a shadowed Reflect", "import * as mod from 'mod'; var Reflect; Reflect.set(mod, k, v);"},

		// A member write whose object is a deeper access. `mod.a.b = 0` writes to `mod.a`'s object,
		// not to the namespace, and upstream keeps the two-level form clean.
		{"a three-level property write on a namespace import",
			"import * as mod from 'mod'; mod.a.b.c = 0"},

		// A call that is not on Object or Reflect.
		{"a namespace passed to an unrelated call",
			"import * as mod from 'mod'; Foo.assign(mod, obj);"},

		// A type-only reference. Our tree is TypeScript and upstream's corpus is not, so nothing
		// above proves an import used as a type stays clean.
		{"an import used as a type annotation",
			"import {Thing} from 'mod'; let x: Thing; x = 0;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoImportAssign, importAssignFile, testCase.sourceText))
		})
	}
}

// The rule declares NeedsTypeChecker, and the plain harness hands it a nil checker.
//
// This test exists so a revert of `RunTyped` to `Run` in the fixtures above fails loudly rather than
// turning every clean case green vacuously. Without it, that revert makes the rule silent on all 116
// corpus cases and the suite still passes.
func TestNoImportAssignNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoImportAssign.NeedsTypeChecker {
		t.Fatal("the rule no longer declares NeedsTypeChecker, so the fixtures above may be running against a nil checker")
	}
	result := rule_testing.Run(t, NoImportAssign, importAssignFile, "import mod from 'mod'; mod = 0")
	if len(result.Diagnostics) != 0 {
		t.Fatalf("the untyped harness produced %d findings, so this test no longer proves what it claims", len(result.Diagnostics))
	}
}
