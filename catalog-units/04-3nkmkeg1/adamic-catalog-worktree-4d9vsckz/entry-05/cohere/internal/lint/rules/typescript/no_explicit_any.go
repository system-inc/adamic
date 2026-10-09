package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/sourcename"
)

var messageNoExplicitAny = rule.Message{
	Id: "unexpectedAny",
	Description: "This annotation is `any`, which switches the type system off for every value " +
		"that flows through it: no property is checked, no argument is checked, and no assignment " +
		"is refused. The compiler stops answering questions about this value and reports nothing " +
		"when it is used wrongly. Write the type it actually holds, or `unknown` when the shape is " +
		"genuinely not known yet, which keeps the value opaque until it is narrowed rather than " +
		"treating it as anything at all.",
}

// NoExplicitAnyOptions configures which `any` annotations report and whether a repair is offered.
//
// The authoritative surface is the `NoExplicitAny` struct oxc derives `JsonSchema` from, which
// carries exactly these two fields under `serde(rename_all = "camelCase", deny_unknown_fields)`.
// Read from the struct rather than from an inventory column, because that column has been wrong.
type NoExplicitAnyOptions struct {
	// FixToUnknown offers to rewrite the `any` token as `unknown`.
	//
	// Off by default, and deliberately opt-in upstream, because the rewrite is not meaning
	// preserving in the way a repair normally is: `any` accepts every use of the value while
	// `unknown` accepts none until it is narrowed, so turning one into the other converts silence
	// into a wave of new compiler errors at every use site. That is the point of the rewrite and
	// also why nobody should get it without asking.
	//
	// Measured on the release binary rather than reasoned about, because the fix/suggestion split
	// matters here and the brief's own framing pointed the other way. With `fixToUnknown` set,
	// `oxlint --fix` rewrote `let a: any = 1` to `let a: unknown = 1` unattended, with no human
	// choosing it; with the option unset the same `--fix` run left the file byte-identical. So
	// upstream ships a *conditional fix*, not a suggestion, and this port reproduces that: a fix
	// when the flag is on, and no repair at all when it is off.
	FixToUnknown bool `json:"fixToUnknown"`

	// IgnoreRestArgs exempts every `any` written anywhere inside a rest parameter.
	//
	// The exemption is by containment rather than by shape, which is wider than it first reads and
	// was measured rather than inferred. Upstream walks *every* ancestor of the `any` token looking
	// for a rest parameter, so `...args: Map<any, any>` and `...args: any[][]` are exempt as
	// wholly as `...args: any[]` is. See `isInsideRestParameter` for the four inputs that pin it.
	IgnoreRestArgs bool `json:"ignoreRestArgs"`
}

// NoExplicitAny flags an explicit `any` type annotation.
//
//	valid:   const age: number = 17;
//	valid:   function greet(): Array<string> {}
//	valid:   function foo(...args: any[]) {}     // only under ignoreRestArgs
//	invalid: const age: any = 'seventeen';
//	invalid: function greet(): any[] {}
//	invalid: function greet(param: Array<any>): Array<any> {}
//	invalid: type Any = any;
//	invalid: const x = y as any;
//
// Ported from `typescript/no-explicit-any`, which oxc in turn ports from
// `@typescript-eslint/no-explicit-any`.
//
// # The rule is one node kind, and everything interesting is around it
//
// There is no traversal and no classification to reproduce: `any` is a keyword with its own node
// kind, so every position it can appear in reports, and the corpus's forty failing inputs are
// forty different syntactic homes for one token rather than forty judgments. A variable annotation,
// a return type, a type argument, an array element type, a union member, an intersection member, a
// type parameter default, a heritage type argument, a constraint, an index signature value, an
// `as` assertion, a bare type alias — all of them are the same node reached by a different route,
// and none of them needs its own arm.
//
// That makes the two option flags and the file gate the whole of the port, which is why they carry
// the reasoning below and the detection carries none.
//
// # The file gate is load-bearing here, and it is not an optimization
//
// Upstream's `should_run` reads `ctx.source_type().is_typescript()`, which reads as a cheap skip of
// files that cannot contain the syntax. In this tree it is a correctness guard, because our parser
// produces `any` nodes in files oxc's never would.
//
// typescript-go parses JavaScript with the TypeScript grammar and resolves JSDoc type comments into
// real type nodes. Probed on our own harness: `/** @type {any} */ let x = 1;` in a `.js` file
// yields a live `KindAnyKeyword` node, and so does `/** @param {any} p */` in a `.jsx` file. Both
// are silent upstream, confirmed on the release binary with a `.ts` control firing in the same run
// so the zero was a measurement rather than a misconfigured probe. Without this gate the rule would
// report every JSDoc `any` in every JavaScript file in the tree, which upstream structurally cannot
// see.
//
// The extension set is oxc's own, from `VALID_EXTENSIONS` and `Language` in `oxc_span/source_type.rs`:
// `is_typescript` is true for `TypeScript` and `TypeScriptDefinition`, which covers `.ts`, `.mts`,
// `.cts`, `.tsx` and `.d.ts`. All five were run against the release binary and all five report;
// `.js`, `.jsx`, `.mjs` and `.cjs` were run in the same batch and all four are silent. Declaration
// files are explicitly *not* exempt, which is worth stating because a rule about type annotations
// exempting the files that are nothing but type annotations would be the plausible guess.
//
// # Where this port reports and upstream cannot
//
// Two of upstream's own failing inputs produce no rule finding there, and both are recorded here as
// reporting rather than as clean:
//
//	interface Greeter { constructor(param: Array<any>) {} }
//	type obj = { constructor(param: Array<any>) {} }
//
// A method signature in an interface or a type literal may not carry a body, so both are syntax
// errors. oxc's parser refuses them and emits `Unexpected token` in place of any rule finding,
// which is why the snapshot holds forty-one rule findings across thirty-nine failing inputs rather
// than the forty-three a naive count suggests. typescript-go's parser recovers instead, producing a
// `MethodSignature` with a body it then reports separately, and the `any` inside the parameter list
// is a perfectly ordinary node by the time this rule sees it.
//
// This is a difference in parser tolerance rather than in what the rule decides, so it is left
// alone: making the rule silent on recovered syntax would mean teaching it to detect parse errors,
// which is not its job and would misfire on every other tolerated shape. Both cases are pinned as
// fixtures asserting a finding, with the upstream behavior noted at the line, so the next reader
// does not discover the mismatch by grepping the snapshot and quietly "fix" it.
var NoExplicitAny = rule.Rule{
	Name: "@typescript-eslint/no-explicit-any",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Declining the file rather than the node, matching upstream's `should_run`. It is also the
		// cheapest possible decline in a tree where one walk serves every rule.
		if !isTypeScriptSourceFile(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		parsed, _ := rule.OptionsAs[NoExplicitAnyOptions](options)

		return rule.Listeners{
			ast.KindAnyKeyword: func(node *ast.Node) {
				if parsed.IgnoreRestArgs && isInsideRestParameter(node) {
					return
				}

				if parsed.FixToUnknown {
					ctx.ReportNodeWithFixes(node, messageNoExplicitAny,
						rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), "unknown"))
					return
				}

				ctx.ReportNode(node, messageNoExplicitAny)
			},
		}
	},
}

// isTypeScriptSourceFile is oxc's `source_type().is_typescript()`, by file name, with an Adamic `.a`
// file read as the `.ts` it is (sourcename.TreatedAs, #kwt1htp).
//
// The `.d.ts` case needs no arm of its own: it ends in `.ts` and upstream treats a
// `TypeScriptDefinition` as TypeScript for this predicate, so the suffix test already answers it.
// Kept as a named function rather than inlined so the extension list has one home and the reasoning
// above has something to point at.
func isTypeScriptSourceFile(fileName string) bool {
	return strings.HasSuffix(sourcename.TreatedAs(fileName), ".ts") ||
		strings.HasSuffix(fileName, ".tsx") ||
		strings.HasSuffix(fileName, ".mts") ||
		strings.HasSuffix(fileName, ".cts")
}

// isInsideRestParameter answers whether any ancestor of this node is a rest parameter.
//
// Upstream is `ctx.nodes().ancestors(node.id()).any(|parent| matches!(parent.kind(),
// AstKind::FormalParameterRest(_)))`, an unbounded walk to the root rather than a look at the
// immediate parent, and reproducing the *unbounded* part is the whole content of this function.
//
// oxc has a distinct node kind for a rest parameter; typescript-go has one `KindParameter` whose
// `DotDotDotToken` is non-nil for the rest case, so the test is on the field rather than the kind.
//
// # Why containment and not shape, measured
//
// The natural reading of "ignore rest args" is that it exempts `...args: any[]`, and a port that
// implemented that shape would pass every one of upstream's forty-five option-carrying pass cases,
// because all forty-five are `any[]`, `readonly any[]`, `Array<any>` or `ReadonlyArray<any>`. The
// corpus cannot distinguish the two readings. The release binary can, and does:
//
//	function r04(...args: Map<any, any>) {}   silent   — not an array shape at all
//	function r06(...args: any[][]) {}          silent   — nested two deep
//	function r01(...args: any) {}              silent   — no array anywhere
//	function r05(cb: (...a: any[]) => any) {}  reports once, on the RETURN type only
//
// The last one is what pins the walk as unbounded-but-scoped: the `any` in `...a: any[]` is two
// levels below the rest parameter and is exempt, while the `=> any` beside it shares a function
// type but no rest-parameter ancestor and reports. A parent-only check would report both; a
// whole-subtree check would report neither.
//
// # The stale TODO
//
// Upstream's corpus carries `function foo(...args: any) {}` commented out under a `// todo`, which
// reads as a known gap where the bare form is not yet exempt. It is stale: measured on the release
// binary that input is silent, with a control `any` on the next line of the same file firing, so
// the ancestor walk already covers it. This port matches the measured behavior rather than the
// comment, and the case is pinned as a fixture so a later reader who finds the TODO does not
// "restore" a gap that never shipped.
func isInsideRestParameter(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindParameter && parent.AsParameterDeclaration().DotDotDotToken != nil {
			return true
		}
	}
	return false
}
