package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noUnusedVarsMessage renders the finding for one binding.
//
// It names the binding and says whether it was ever given a value, the two facts ESLint's
// "'x' is assigned a value but never used" carries, so a reader of a list of findings can tell them
// apart without opening every file. It offers the underscore escape only when the resolved config
// honours it, so the advice can never name a rename that leaves the finding standing: under a `^$`
// ignore pattern, or with no pattern at all, the only repair offered is deletion.
func noUnusedVarsMessage(name string, assigned bool, underscoreSilences bool) rule.Message {
	action := "declared"
	if assigned {
		action = "assigned a value"
	}
	repair := "Delete it."
	if underscoreSilences {
		repair = "Delete it, or prefix it with an underscore, which the configured ignore pattern " +
			"accepts, to say the omission is deliberate."
	}
	return rule.Message{
		Id: "noUnusedVars",
		Description: "'" + name + "' is " + action + " and nothing ever reads it. A name that is " +
			"written but never read is almost always the residue of an edit that moved on: an import " +
			"whose call site was deleted, a parameter left behind when a signature changed, a variable " +
			"holding a value nobody asked for. It costs nothing at runtime, which is what lets it " +
			"accumulate, and it costs the next reader real time, because an unused name reads exactly " +
			"like a used one until you search the file and find nothing. " + repair,
	}
}

// NoUnusedVars flags a variable, import, parameter, or type that nothing reads.
//
//	valid:   const x = 10; alert(x);
//	valid:   function getY([, y]) { return y; }
//	valid:   export const x = 1;
//	valid:   import type { T } from './t'; const a: T = read();
//	invalid: const x = 10;
//	invalid: function add(a, b) { return a; }
//	invalid: import { unusedThing } from './module';
//	invalid: let z = 0; z = z + 1;
//	invalid: const _unused = compute();   // no name is exempt unless a pattern is configured
//
// # What this rule decides, and at what scope
//
// The question here is deliberately narrow: is this binding read anywhere within the file that
// declares it. That is not the same question as "is this symbol used anywhere in the program", and
// the difference is the whole reason this rule can be a per-file gate at all.
//
// A per-file rule cannot see who imports the file it is handed, so every exported binding is
// treated as used. That is upstream's answer too, for the same reason, and it is the conservative
// direction: the alternative is reporting a component's only export as dead because this file
// happens not to call it. The program-wide claim is a real and different analysis and it already
// exists here as `cohere --unused`, which is opt-in precisely so that reading every file in the
// program does not poison the per-file findings cache. See `rule.ProgramReads` for why that
// separation is structural rather than stylistic: a rule reading beyond its own file, cached on
// that file's hash, serves a stale answer forever and nothing notices.
//
// So this rule declines the whole-program question rather than approximating it. Two analyses, two
// scopes, neither pretending to be the other.
//
// # Why this does not use the liveness dataflow, and what that costs
//
// `no_useless_assignment_analysis.go` in this package is a real backward-liveness dataflow over the
// control-flow graph, and the temptation to route this rule through it is strong: oxc spends
// roughly 990 lines in `usage.rs` hand-matching patterns — a self-reference inside a function's own
// body, a variable only ever reassigned to itself, a write nothing reads — that a dataflow answers
// in one query. oxc has an SSA crate in the same repository and this rule does not touch it, which
// reads like an omission.
//
// It is not an omission this port should correct, and the reason is the harness rather than the
// analysis. The differential compares our findings against oxlint's. A rule that is *more* correct
// than upstream produces a difference, and a difference costs the harness its meaning whether the
// extra finding is right or wrong. Measured on upstream's own corpus, the two approaches genuinely
// disagree: `test_vars_self_use` ships `chain = chain.extend()` inside a `for` loop as a PASS and
// the same statement outside a loop as a FAIL. A liveness pass sees a store whose value no later
// read observes in both cases and would report both. Upstream's answer is the loop-shape test at
// `usage.rs:612` (`is_in_for_loop_test_or_update`) and `:634` (`is_in_loop_body`), which is a
// syntactic proxy for the flow fact that a loop's back edge makes the store observable on the next
// iteration.
//
// The dataflow would get that particular case right and would still be a divergence, because it
// would also change the answer on shapes upstream never wrote a case for and where nobody has
// established which answer is correct. So the pattern catalogue is reproduced, upstream's loop
// exemptions included, and the flow analysis is left where it already ships a measured win.
//
// The cost is stated rather than hidden: this rule is exactly as blind as upstream is, on exactly
// the shapes upstream is blind on. That is the subset this port can show correct.
//
// # What this reports against upstream, measured
//
// 407 of 407 clean cases stay clean, 301 of 303 reporting cases report, and all 51 JSX cases agree.
// oxc's corpus ships two more clean cases that are clean only under its leading-underscore default;
// ESLint reports both, and so does this rule now (see the next section).
// The two that do not are named with their causes in `noUnusedVarsKnownGaps`, and one of them is a
// case upstream cannot report either and documents as such in its own test file.
//
// Zero false positives is the property that matters most here and it is the one held fixed: every
// change that closed a gap was measured against all the clean cases first, and two changes that
// closed gaps at the cost of a clean case were reworked rather than kept. A rule that misses
// something costs a finding nobody had; a rule that accuses wrongly costs trust in every finding it
// makes.
//
// # No name is ignored by default, which is ESLint's answer and no longer oxc's
//
// oxc ignores any binding whose name begins with `_` under default options; ESLint does not. This
// is measured on both: ESLint's own Linter API reports `_a` in `const _a = 1;` with no options
// configured, and oxc's `ignored.rs:455` resolves `IgnorePattern::Default` to
// `haystack.starts_with('_')`.
//
// This rule took oxc's answer while oxc was what the differential compared against, and reversed it
// on 2026-10-01 under the parity doctrine: never worse than ESLint, rule by rule. Measured on the
// installed `@typescript-eslint/no-unused-vars` 8.67.0 with no options, `const _deliberate = 1`,
// `const _ = 1`, `function f(_) {}` and `catch (_) {}` all report, so oxc's default was a set of
// findings ESLint makes and this rule did not. It was also an allowance built into the rule rather
// than written in a config: the ahra config carried `varsIgnorePattern`, `argsIgnorePattern` and
// `caughtErrorsIgnorePattern` at `^$` for no reason but to switch it off, and those three lines are
// now redundant. An exemption a project wants is still available by configuring a pattern, where it
// is visible.
//
// History worth keeping, because it explains the `^$` lines a reader will still find in configs: on
// 2026-09-05 the one finding ESLint made on ahra and this rule did not was `_nextContent` at
// `modules/pensieve/PensieveBootstrap.test.ts`, missing only because of the oxc default, and the
// `^$` patterns were added to close it.
//
// The message offers the underscore escape only when the resolved config honours it, which is to
// say only when `_` plus the name matches the pattern configured for that kind of binding. Before,
// it told every finding to prefix an underscore, including `_nextContent` under a `^$` config.
//
// Worth stating plainly because it has been assumed twice in the other direction: the missing
// finding is NOT evidence that this rule is blind to type declarations. It is not. A seeded probe
// carrying an unused `type` alias, an unused `interface` and an unused `const` reports all three,
// and the two type findings above are ones this rule already makes on the real tree. The
// `bindingType` arm in `collectCandidateBindings` is what does it. A port of the namespaced rule
// undertaken to recover type-declaration coverage would be recovering coverage that is already
// here.
var NoUnusedVars = rule.Rule{
	Name: "@typescript-eslint/no-unused-vars",

	// Symbol identity, not name matching. Two bindings can share a name across scopes and a read of
	// one must not keep the other alive: `let x; { let x; log(x); }` reads the inner binding only,
	// and a name-matching rule calls both used. That is the quiet direction of the mistake, and for
	// this rule the quiet direction is where every missed finding lives.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := resolveNoUnusedVarsOptions(options)

		return rule.Listeners{
			// One pre-pass over the file rather than a listener per declaration kind. The rule has
			// to know every read in the file before it can judge any declaration, and reads can sit
			// above a declaration inside a hoisted function as easily as below it. `KindSourceFile`
			// fires before its children, so the gather and the report both happen here.
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				sourceFile := node.AsSourceFile()
				if sourceFile == nil {
					return
				}

				// Upstream skips these entirely, and the reasons are not stylistic. A `.d.ts`
				// declares rather than defines and its bindings are consumed by other files by
				// construction, so every one of them would report. Vue, Svelte and Astro put reads
				// in a template the linter never parses, so a binding used in the markup looks
				// unused in the script.
				//
				// The `.d.ts` arm is SUBSUMED rather than load-bearing, and it is kept anyway.
				// Measured: with this suffix test disabled, `interface Unreferenced {}` in a
				// `a.d.ts` still reports nothing, because typescript-go marks every declaration in
				// a declaration file ambient and `isInsideSignatureOrAmbientDeclaration` already
				// declines it. A mutant disabling this line therefore survives every fixture, and
				// the fixture asserting it passes for a reason one layer down from what it names.
				//
				// It stays because the two guards answer different questions and only coincide
				// today: this one is about the FILE and would keep holding if a future `.d.ts`
				// carried a non-ambient binding, which is the shape the ambient test would miss.
				// Recorded here so the next reader does not delete it believing it does nothing,
				// and does not add a fixture believing it can be measured.
				fileName := sourceFile.FileName()
				if strings.HasSuffix(fileName.AsString(), ".d.ts") ||
					strings.HasSuffix(fileName.AsString(), ".vue") ||
					strings.HasSuffix(fileName.AsString(), ".svelte") ||
					strings.HasSuffix(fileName.AsString(), ".astro") {
					return
				}

				analyzeUnusedBindings(ctx, node, settings)
			},
		}
	},
}

// unusedBindingKind separates the declaration forms, because the exemptions differ per form and so
// does the wording of the finding.
//
// This mirrors upstream's match on `symbol.declaration().kind()` at `mod.rs:281`. Collapsing them
// would be wrong rather than merely coarse: a parameter is exempt under `args: "after-used"` when a
// later parameter is used, a variable never is, and a caught error has its own ignore pattern.
type unusedBindingKind uint8

const (
	// bindingVariable is a `var`, `let`, `const`, or destructured element.
	bindingVariable unusedBindingKind = iota
	// bindingParameter is a function or method parameter.
	bindingParameter
	// bindingImport is an imported name, default or named or namespace.
	bindingImport
	// bindingCaughtError is a `catch (error)` binding.
	bindingCaughtError
	// bindingType is an interface, type alias, or type parameter.
	bindingType
)

// candidateBinding is one declared name this rule may judge, with everything needed to judge it.
//
// The declaring node is carried alongside the name because the exemptions ask about the shape
// around the declaration — whether a parameter is followed by a used one, whether a variable sits
// in a `for` head — and the name alone cannot answer those.
type candidateBinding struct {
	name        *ast.Node
	declaration *ast.Node
	kind        unusedBindingKind
}

// analyzeUnusedBindings gathers every candidate binding in one file, then reports the ones nothing
// reads.
//
// Two passes rather than one, and the order is forced. A read can precede its declaration in source
// order — a hoisted function body referring to a `const` declared below it is legal and common — so
// no single walk can decide a binding at the moment it sees it.
// noUnusedVarsUsedOnlyAsTypeMessage renders the finding for a binding whose only references are
// `typeof` type queries, typescript-eslint's `usedOnlyAsType`.
func noUnusedVarsUsedOnlyAsTypeMessage(name string, assigned bool) rule.Message {
	action := "declared"
	if assigned {
		action = "assigned a value"
	}
	return rule.Message{
		Id: "usedOnlyAsType",
		Description: "'" + name + "' is " + action + " but only used as a type. Every reference is a " +
			"`typeof`, which reads the binding's type while compiling and nothing at runtime, so the " +
			"value itself is built for no reader. Write the type out and delete the value, or keep " +
			"the value if something outside this file is meant to read it and export it.",
	}
}

func analyzeUnusedBindings(ctx rule.Context, sourceFile *ast.Node, settings NoUnusedVarsOptions) {
	candidates := collectCandidateBindings(sourceFile)
	if len(candidates) == 0 {
		return
	}

	// The name set is a pre-filter, not a discrimination. Symbol identity still decides every
	// match; this only keeps the walk from paying for a checker call on identifiers that could not
	// possibly resolve to a candidate, and this walk visits every identifier in the file.
	interesting := map[string]bool{}
	for _, candidate := range candidates {
		interesting[candidate.name.Text()] = true
	}

	reads, writes, typeOnlyReads := collectReadSymbols(ctx, sourceFile, interesting)

	// Compiled once per file. An unparseable pattern ignores nothing, as the other patterns do.
	var destructuredArrayIgnore *regexp.Regexp
	if settings.DestructuredArrayIgnorePattern != "" {
		destructuredArrayIgnore, _ = regexp.Compile(settings.DestructuredArrayIgnorePattern)
	}

	// The bindings reported unused so far, in report order, which an import's removal reads. See
	// importRemoval.
	reportedUnused := map[*ast.Node]bool{}

	for _, candidate := range candidates {
		// The gates before the exemptions run in upstream's order, collectUnusedVariables', because
		// reportUsedIgnorePattern makes the order visible: a name an earlier gate sets aside is never
		// asked whether it is used. `vars: "local"` first, then an ignored array element, then a class
		// with a static block, then the two kinds a mode skips whole, then each kind's ignore pattern,
		// then `using`. See isExemptFromUnusedReport for what follows.
		if settings.Vars == "local" && isGlobalScopeBinding(ctx.SourceFile, candidate) {
			continue
		}
		name := candidate.name.Text()
		var candidateWrites []*ast.Node
		usedOnlyAsType := false
		symbol := ctx.TypeChecker.GetSymbolAtLocation(candidate.name)
		if symbol != nil {
			candidateWrites = writes[symbol]
			usedOnlyAsType = typeOnlyReads[symbol]
		}
		// Upstream asks this of every kind before any other: a parameter or a caught error
		// destructured from an array is an ignored array element first.
		arrayElement := destructuredArrayIgnore != nil && isDestructuredFromAnArray(candidate, candidateWrites)
		if arrayElement && destructuredArrayIgnore.MatchString(name) {
			reportUsedIgnoredName(ctx, candidate, symbol, candidateWrites, settings, reads,
				"destructuredArrayIgnorePattern", settings.DestructuredArrayIgnorePattern, "an element of array destructuring")
			continue
		}
		if settings.IgnoreClassWithStaticInitBlock && candidate.declaration.Kind == ast.KindClassDeclaration &&
			hasStaticBlock(candidate.declaration) {
			continue
		}
		// `caughtErrors: "none"` skips every caught error, including each name destructured out of
		// one, and `args: "none"` every parameter, before their patterns are asked.
		if candidate.kind == bindingCaughtError && settings.CaughtErrors == "none" {
			continue
		}
		if candidate.kind == bindingParameter && settings.Args == "none" {
			continue
		}
		if matchesIgnorePattern(candidate, settings) {
			option, pattern, noun := ignorePatternFor(candidate.kind, settings)
			reportUsedIgnoredName(ctx, candidate, symbol, candidateWrites, settings, reads, option, pattern, noun)
			continue
		}
		if settings.IgnoreUsingDeclarations && isUsingDeclaration(candidate) {
			continue
		}
		if isExemptFromUnusedReport(ctx, candidate, symbol, candidates, settings, reads) {
			continue
		}
		// Before either report, as upstream's gate is: a rest sibling read only through `typeof` is
		// silent too.
		if settings.IgnoreRestSiblings && hasRestSiblingDeclarationOrWrite(candidate, candidateWrites) {
			continue
		}
		if usedOnlyAsType {
			// A value import read only through `typeof` is consistent-type-imports' finding, so
			// upstream withholds this one rather than report the import twice.
			if candidate.kind == bindingImport {
				continue
			}
			ctx.ReportNode(
				unusedBindingReportNode(candidate, candidateWrites),
				noUnusedVarsUsedOnlyAsTypeMessage(name, len(candidateWrites) != 0 || declaringNameIsInitialized(candidate.name)),
			)
			continue
		}
		reportNode := unusedBindingReportNode(candidate, candidateWrites)
		message := noUnusedVarsMessage(
			name,
			len(candidateWrites) != 0 || declaringNameIsInitialized(candidate.name),
			nameMatchesIgnorePattern("_"+name, candidate.kind, settings) ||
				(arrayElement && destructuredArrayIgnore.MatchString("_"+name)),
		)
		// Marked before the removal is computed, as upstream's report does: the removal asks whether
		// every binding of the declaration has been reported yet, this one included.
		reportedUnused[candidate.name] = true
		if candidate.kind == bindingImport {
			if removal, removalId, removable := importRemoval(ctx, candidate, symbol, reportedUnused); removable {
				if settings.EnableAutofixRemoval.Imports {
					ctx.ReportNodeWithFixes(reportNode, message, removal)
				} else {
					ctx.ReportNodeWithSuggestions(reportNode, message, rule.Suggestion{
						Message: noUnusedVarsRemovalMessage(removalId, name),
						Fixes:   []rule.Fix{removal},
					})
				}
				continue
			}
		}
		ctx.ReportNode(reportNode, message)
	}
}

// unusedBindingReportNode is where a finding points: the last write in the binding's own variable
// scope, or the declaring name when there is none.
//
// ESLint's position, measured on the installed `@typescript-eslint/no-unused-vars` 8.67.0: it reports
// the identifier of the last write reference whose variable scope is the binding's, and falls back
// to the declaration. `let provenance: string;` written on three branches and never read reports at
// the third write, which is the line a reader needs, since that is where the value was last
// produced for nobody. A write inside a nested function is skipped, so
// `let v = 1; v = 2; function g() { v = 3; }` reports at `v = 2`.
//
// A declaring name that initializes, `let v = 1`, is itself a write in ESLint's model, so a write
// that precedes it in source order (only possible for a hoisted `var`) does not win over it.
func unusedBindingReportNode(candidate candidateBinding, writes []*ast.Node) *ast.Node {
	var last *ast.Node
	for _, write := range writes {
		if !sameVariableScope(write, candidate.declaration) {
			continue
		}
		if last == nil || write.Pos() > last.Pos() {
			last = write
		}
	}
	if last == nil {
		return candidate.name
	}
	if last.Pos() < candidate.name.Pos() && declaringNameIsInitialized(candidate.name) {
		return candidate.name
	}
	return last
}

// declaringNameIsInitialized reports whether the declaring name receives a value where it is
// declared: a variable or a destructured element whose declarator has an initializer, a binding
// element or parameter with a default, or a `for...in`/`for...of` head, which assigns on every
// iteration.
func declaringNameIsInitialized(name *ast.Node) bool {
	for current := name.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBindingElement:
			if current.Initializer() != nil {
				return true
			}
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			// Climb to the declarator or parameter that owns the pattern.
		case ast.KindVariableDeclaration:
			if current.Initializer() != nil {
				return true
			}
			list := current.Parent
			if list != nil && list.Parent != nil &&
				(list.Parent.Kind == ast.KindForInStatement || list.Parent.Kind == ast.KindForOfStatement) {
				return true
			}
			return false
		case ast.KindParameter:
			return current.Initializer() != nil
		default:
			return false
		}
	}
	return false
}

// collectReadSymbols resolves every identifier that could be a read of a candidate, and returns the
// symbols those reads name.
//
// "Read" here is the rule's own notion and it is narrower than "appears". An identifier in a
// declaring position is not a read of itself, and a pure write is not a read: `y = 5` does not make
// `y` used, which upstream states outright in its documentation ("Write-only variables are not
// considered as used"). Both exclusions are what make the rule find anything at all.
//
// It also returns every write to those symbols, which the report needs and the read test does not:
// where the finding points, and whether the binding was ever given a value. A write here is any
// non-declaring occurrence that stores, update forms included, which is ESLint's `isWrite()`.
//
// The third map holds the symbols read through a `typeof` type query or named by a type predicate.
// typescript-eslint counts such a read as type-only, not a use, so a value whose every reference is
// one is unused and reports as "only used as a type". Measured on the installed plugin, 8.67.0:
// `const d = {}; export type D = typeof d;` reports, and so do `keyof typeof d` and
// `ReturnType<typeof f>`. A symbol here that is also read is exempt before this map is asked, and an
// import is never reported from it at all, so neither needs excluding here: upstream's exception for
// a type-only import reaches the same silence.
func collectReadSymbols(ctx rule.Context, sourceFile *ast.Node, interesting map[string]bool) (map[*ast.Symbol]bool, map[*ast.Symbol][]*ast.Node, map[*ast.Symbol]bool) {
	reads := map[*ast.Symbol]bool{}
	writes := map[*ast.Symbol][]*ast.Node{}
	typeOnlyReads := map[*ast.Symbol]bool{}

	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil {
			return
		}

		if current.Kind == ast.KindIdentifier && interesting[current.Text()] &&
			current.Parent != nil && current.Parent.Kind == ast.KindTypePredicate &&
			current.Parent.AsTypePredicateNode().ParameterName == current {
			// `data is string` names the parameter without reading it. Still no read, as before, but
			// typescript-eslint words the finding "only used as a type", so it is recorded as one.
			for _, symbol := range resolveIdentifierSymbols(ctx, current) {
				typeOnlyReads[symbol] = true
			}
		}

		if current.Kind == ast.KindIdentifier && interesting[current.Text()] {
			isWrite := !isDeclaringName(current) && reference.WritesToBinding(current)
			for _, symbol := range resolveIdentifierSymbols(ctx, current) {
				if isWrite {
					writes[symbol] = append(writes[symbol], current)
				}
				if !countsAsRead(current, declaringScopeOf(ctx, symbol)) {
					continue
				}
				if isSelfReferenceWithinOwnDeclaration(ctx, current, symbol) {
					continue
				}
				if isInsideTypeQuery(current) {
					typeOnlyReads[symbol] = true
					continue
				}
				reads[symbol] = true
			}
		}

		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)

	return reads, writes, typeOnlyReads
}

// isInsideTypeQuery is typescript-eslint's `referenceContainsTypeQuery`: the identifier sits in a
// `typeof` type query, reached through nothing but qualified names, so `typeof a.b` counts for `a`.
func isInsideTypeQuery(identifier *ast.Node) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindTypeQuery:
			return true
		case ast.KindQualifiedName:
			continue
		}
		return false
	}
	return false
}

// declaringScopeOf returns the variable scope a symbol is declared in, or nil when that cannot be
// established.
//
// The FIRST declaration is deliberate rather than arbitrary here, and the question it answers is
// "where does this name live", which every declaration of a merged symbol answers identically: a
// symbol's declarations all bind the same name in the same scope, which is what makes them merge.
// A rule asking a different question of a merged symbol would need a different choice.
func declaringScopeOf(ctx rule.Context, symbol *ast.Symbol) *ast.Node {
	// This file's declarations only: the bindings this rule reports are the ones this file declares.
	declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
	if len(declarations) == 0 {
		return nil
	}
	return enclosingVariableScope(declarations[0])
}

// countsAsRead reports whether an identifier occurrence loads the binding's value.
//
// The two exclusions are the rule's core judgment and they run in opposite directions, which is why
// they are separated here rather than folded into one predicate.
//
// A declaring name is not a use of itself. Without this every binding would be trivially used by
// its own declaration and the rule would report nothing — the failure mode that looks exactly like
// a clean tree.
//
// A pure write is not a use. `let y = 10; y = 5;` leaves `y` unused, because nothing ever observes
// either value. An update (`y++`, `y += 1`) is deliberately NOT excluded: it loads before it
// stores, so it is a genuine read of the previous value. That asymmetry is upstream's and it is the
// difference between `y = 5` and `y++` in the documentation's own examples.
func countsAsRead(identifier *ast.Node, declaringScope *ast.Node) bool {
	if isDeclaringName(identifier) {
		return false
	}

	// A pure assignment target reads nothing. `reference.WritesToBinding` already handles the
	// parenthesized target, the destructuring target, and the shorthand-property positions, and
	// declines a property name on the left of a member access, which is not a binding at all.
	if reference.WritesToBinding(identifier) && !isUpdateWrite(identifier) {
		return false
	}

	// A read whose value is thrown away by a comma sequence is not a use. `return (a, 0)` and
	// `(a, 0) + 1` both evaluate `a` and then discard it: the sequence yields its right operand, so
	// nothing downstream can observe the left one. Upstream calls this a discarded read
	// (`usage.rs:709`) and reports the binding.
	//
	// Only the LEFT operand qualifies, and that is the whole discrimination: `(0, a) + 1` is a
	// genuine read because `a` is the value the sequence produces. Both spellings sit in the corpus
	// and a fix that ignored the side would trade two findings for a false positive.
	if isDiscardedSequenceOperandRead(identifier) {
		return false
	}

	// A read from inside a function that is itself being assigned to the same binding is not a
	// use. `function foo(cb) { cb = function(a) { cb(1 + a); }; ... }` stores a closure into `cb`
	// whose only reader is `cb` itself, so nothing outside the cycle observes the binding and
	// upstream reports the parameter.
	//
	// This is deliberately narrow, and the narrowness is upstream's rather than caution of ours.
	// Two neighbouring shapes are upstream PASS cases and must stay clean: an immediately-invoked
	// function expression (`cb = function(a){ return cb(1+a); }()`, which runs now and so really
	// does read `cb`) and a sequence whose value is the binding itself (`cb = (function(a){...},
	// cb)`). Upstream ships both spellings of the sequence form in its fail list too, commented
	// out, at `tests/eslint.rs:743` and `:745` — it knows they should report and cannot make them.
	// Those two stay silent here for the same reason, which is a reproduced gap and not an
	// oversight.
	if isReadInsideAClosureAssignedToItself(identifier) {
		return false
	}

	// An update or a compound assignment whose only purpose is to feed the binding back into
	// itself is not a use. `var a = 0; a = a + 1;` and `a++` both load the binding, but nothing
	// ever observes the result, so upstream calls them unused and documents it outright: "A read
	// for a modification of itself is not considered as used."
	//
	// This is `usage.rs:445`'s `is_self_reassignment`, and it is the single largest source of
	// findings this rule would otherwise miss: it accounted for 129 of the corpus's failing cases
	// going silent before it existed.
	if isSelfReassignmentRead(identifier, declaringScope) {
		return false
	}

	return true
}

// isUpdateWrite reports whether a write to this identifier also loads it first.
//
// `x++`, `x--`, and every compound assignment (`x += 1`, `x ||= y`) read the current value before
// storing the new one, so they keep an otherwise-unused binding alive in a way a plain `x = 1` does
// not.
func isUpdateWrite(identifier *ast.Node) bool {
	for current := identifier; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			continue

		case ast.KindPrefixUnaryExpression:
			return reference.IsUpdateOperator(parent.AsPrefixUnaryExpression().Operator)

		case ast.KindPostfixUnaryExpression:
			return reference.IsUpdateOperator(parent.AsPostfixUnaryExpression().Operator)

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.Left != current {
				return false
			}
			// A plain `=` stores without loading; every other assignment operator loads first.
			return binary.OperatorToken != nil &&
				binary.OperatorToken.Kind != ast.KindEqualsToken &&
				ast.IsAssignmentOperator(binary.OperatorToken.Kind)

		default:
			return false
		}
	}
	return false
}

// isDeclaringName reports whether an identifier is the name being declared rather than a use of it.
//
// Anchored on the parent's own name slot rather than on a kind list, because the same node kind
// appears in both positions: the `x` in `import { x }` declares, the `x` in `f(x)` uses, and both
// are plain identifiers whose parent chain differs only in which field points at them.
func isDeclaringName(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration,
		ast.KindParameter,
		ast.KindBindingElement,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindEnumMember,
		ast.KindModuleDeclaration,
		ast.KindTypeParameter,
		ast.KindImportClause,
		ast.KindImportEqualsDeclaration,
		ast.KindNamespaceImport,
		ast.KindPropertyDeclaration,
		ast.KindPropertySignature,
		ast.KindMethodDeclaration,
		ast.KindMethodSignature,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return parent.Name() == identifier

	// An import specifier declares BOTH of its names, and the export specifier below declares
	// neither of its own, which is why the two cannot share the simple `parent.Name()` test above.
	//
	// `import { a as b } from './m'` puts `a` in the `propertyName` slot and `b` in `name`. Neither
	// is a reference to anything in THIS file: `b` is the binding being created, and `a` names an
	// export of the other module. The plain `parent.Name() == identifier` test answered true only
	// for `b`, so the source name `a` fell through and was walked as an ordinary reference.
	//
	// That is invisible until a module cycle makes the checker resolve it back into this file, and
	// then it is a false positive with no obvious cause. Measured on the ahra tree:
	//
	//	useRouter.ts:11   import { useRouter as useNextRouter } from '.../router/Navigation'
	//	Navigation.ts:31  export { useRouter } from '.../router/hooks/useRouter'
	//	useRouter.ts:16   export function useRouter() { ... }
	//
	// The re-export sends the symbol straight back, so `useRouter` at line 11 resolved to the
	// function declared at line 16 and `no-use-before-define` reported a temporal-dead-zone error
	// on an import binding, which is hoisted and cannot have one. ESLint reports zero on that file
	// because eslint-scope never creates a reference for either half of an import specifier.
	//
	// No imported fixture can see this: it needs a real module graph with a cycle, and upstream's
	// corpus is single-file.
	case ast.KindImportSpecifier:
		return true

	// An export specifier's local name is a genuine reference to the binding it exports, which is
	// why `const y = 1; export { y };` is clean. Its `propertyName` slot, when present, is the
	// local side and the `name` slot is the exported alias, so `export { y as z }` reads `y`.
	case ast.KindExportSpecifier:
		// An export specifier inside `export { x } from './m'` names something in the OTHER
		// module, not the local binding. So `import { resolve } from "path"; export { resolve }
		// from "path";` leaves the import genuinely unused, and upstream reports it, while the
		// same file without the `from` clause is clean because there the specifier does read the
		// local binding.
		//
		// The two spellings are one node kind differing only in whether the enclosing export
		// declaration carries a module specifier, which is why this is asked of the grandparent.
		if exportDeclarationHasModuleSpecifier(parent) {
			return true
		}
		specifier := parent.AsExportSpecifier()
		if specifier.PropertyName != nil {
			return specifier.PropertyName != identifier
		}
		return false

	// A property in an object literal names a key, not a binding. `{ foo: bar }` declares nothing
	// and reads `bar`; the `foo` is a name in the object's own namespace.
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() == identifier

	case ast.KindPropertyAccessExpression:
		// The right side of a member access is a property name, not a binding reference. Reading
		// `a.b` is a read of `a` and says nothing about any binding named `b`.
		return parent.AsPropertyAccessExpression().Name() == identifier

	case ast.KindQualifiedName:
		// The type-position equivalent of a member access: `N.T` reads `N`, not `T`.
		return parent.AsQualifiedName().Right == identifier

	// A `typeof x` inside x's OWN declaration is not a use of it. `function foo(...args: typeof
	// args) {}` names the parameter in its own type annotation, which is self-reference rather than
	// consumption, and upstream reports the parameter. Same judgment as a recursive function or a
	// self-referential type alias, arriving through a type query.
	case ast.KindTypeQuery:
		return typeQueryIsInsideItsOwnDeclaration(identifier)

	// The subject of a return type predicate names the parameter without reading it.
	// `function f(a: unknown): a is string { return true }` never touches `a` at runtime, and
	// upstream reports the parameter as unused while noting the predicate in its message. Counting
	// the predicate as a use makes every type guard's parameter permanently alive.
	//
	// Both spellings land here: `a is string` and `asserts a is string` are one node kind, with
	// `AssertsModifier` set on the second.
	case ast.KindTypePredicate:
		return parent.AsTypePredicateNode().ParameterName == identifier
	}

	return false
}

// exportDeclarationHasModuleSpecifier reports whether an export specifier's declaration re-exports
// from another module rather than exporting a local binding.
func exportDeclarationHasModuleSpecifier(specifier *ast.Node) bool {
	for current := specifier; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindExportDeclaration:
			return current.AsExportDeclaration().ModuleSpecifier != nil
		case ast.KindSourceFile:
			return false
		}
	}
	return false
}

// typeQueryIsInsideItsOwnDeclaration reports whether a `typeof x` sits inside the declaration of x.
//
// See the call site: this is the self-reference judgment applied to a type query.
func typeQueryIsInsideItsOwnDeclaration(identifier *ast.Node) bool {
	name := identifier.Text()
	for current := identifier.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindParameter, ast.KindVariableDeclaration:
			if declared := current.Name(); declared != nil &&
				declared.Kind == ast.KindIdentifier && declared.Text() == name {
				return true
			}
		case ast.KindSourceFile:
			return false
		}
	}
	return false
}

// resolveIdentifierSymbols resolves an identifier to every binding a read of it could keep alive.
//
// Usually one symbol. An export specifier is the exception and it needs both: `export { a }`
// resolves through `GetSymbolAtLocation` to the *export's* own symbol, which is not the symbol any
// local declaration carries, so recording only that leaves `const a = 1; export { a }` reporting a
// binding that is plainly exported. Following the alias reaches the local `const`, which fixes that
// case and breaks `import { a } from 'a'; export { a }`, because there the alias chain runs past the
// local import binding all the way to the foreign declaration.
//
// Both are upstream clean cases, so neither symbol alone is the answer and the union is. Recording
// an extra symbol is safe in the direction that matters: a symbol no candidate holds marks nothing,
// while a missing symbol is a false positive on correct code.
//
// A shorthand destructuring target resolves through the plain accessor to the *property's* symbol
// rather than to the value binding, which reads exactly like a correctly-declined shadow and is
// therefore silent in the expensive direction. The dedicated accessor is what gets `({a} = o)`
// right, and it lives in `services.go` rather than `checker.go`.
func resolveIdentifierSymbols(ctx rule.Context, identifier *ast.Node) []*ast.Symbol {
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if symbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent); symbol != nil {
			return []*ast.Symbol{symbol}
		}
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return nil
	}

	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindExportSpecifier &&
		symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbols := []*ast.Symbol{symbol}
		if aliased := resolvedAlias(ctx, symbol); aliased != nil && aliased != symbol {
			symbols = append(symbols, aliased)
		}
		// The local binding the specifier names, looked up in its own scope without following the alias:
		// `import { a } from 'missing'; export { a }` reads the import even when its module does not
		// resolve and resolvedAlias has nothing to give.
		if local := exportedLocalBinding(ctx, identifier); local != nil && !slices.Contains(symbols, local) {
			symbols = append(symbols, local)
		}
		return symbols
	}

	return []*ast.Symbol{symbol}
}

// exportedLocalBinding is the binding in this file an `export { name }` with no module specifier names,
// or nil for a re-export from another module, whose names are that module's.
func exportedLocalBinding(ctx rule.Context, identifier *ast.Node) *ast.Symbol {
	exportDeclaration := identifier.Parent.Parent.Parent
	if exportDeclaration == nil || exportDeclaration.Kind != ast.KindExportDeclaration ||
		exportDeclaration.AsExportDeclaration().ModuleSpecifier != nil {
		return nil
	}
	return ctx.TypeChecker.ResolveName(identifier.Text(), identifier,
		ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias, false)
}

// resolvedAlias is the declaration an import or export alias reaches, or nil when it reaches none.
//
// An alias whose module does not resolve resolves to the checker's one shared unknown symbol. Followed
// as if it were a binding, every unresolved import in a file looks like the same symbol, so one
// `export { Used }` of an unresolved import marked every other unresolved import as read, and
// `import { Unused, Used } from 'missing'; export { Used }` reported nothing (#e1zk9s0, found by
// typescript-eslint's fixer rows, which import from a module that does not exist). Asked through
// IsUnknownSymbol rather than by reading the symbol's declarations, which would be a read of an
// imported declaration and cost the rule its Shapes claim.
func resolvedAlias(ctx rule.Context, symbol *ast.Symbol) *ast.Symbol {
	aliased := ctx.TypeChecker.GetAliasedSymbol(symbol)
	if aliased == nil || ctx.TypeChecker.IsUnknownSymbol(aliased) {
		return nil
	}
	return aliased
}

// collectCandidateBindings gathers every binding this rule may judge in one file.
//
// Anchored on declaration nodes rather than on a symbol table, because our substrate has no
// per-file symbol enumeration the way oxc's `scoping().symbol_ids()` does. The walk is the
// replacement for that enumeration, and it is why the kind list here has to be explicit: a
// declaration form nobody adds is a finding nobody reports, silently.
func collectCandidateBindings(sourceFile *ast.Node) []candidateBinding {
	var candidates []candidateBinding

	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil {
			return
		}

		switch current.Kind {
		case ast.KindVariableDeclaration:
			// A catch binding is spelled `KindVariableDeclaration` in this AST and is ALSO reached
			// through its catch clause below, where it gets its own kind and its own ignore
			// pattern. Collecting it here as well reported it twice, which a span fixture caught
			// and a message-id fixture would not have: two findings with the right id at the right
			// place still read as correct until you count them.
			if current.Parent != nil && current.Parent.Kind == ast.KindCatchClause {
				break
			}
			// A binding pattern declares several names and each is judged separately, so the
			// pattern's own elements are reached by the walk rather than handled here.
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				candidates = append(candidates, candidateBinding{name, current, bindingVariable})
			}

		case ast.KindBindingElement:
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				kind := bindingVariable
				if owner := patternOwner(current); owner.Kind == ast.KindParameter {
					kind = bindingParameter
				} else if owner.Kind == ast.KindVariableDeclaration && owner.Parent != nil && owner.Parent.Kind == ast.KindCatchClause {
					kind = bindingCaughtError
				}
				candidates = append(candidates, candidateBinding{name, current, kind})
			}

		case ast.KindParameter:
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				candidates = append(candidates, candidateBinding{name, current, bindingParameter})
			}

		// `import X = Y` binds a local alias exactly like an import specifier does, and an unused
		// one is upstream's finding. It is a separate node kind from the module-import family, so
		// it needs its own arm rather than falling out of the one below.
		case ast.KindImportEqualsDeclaration:
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				candidates = append(candidates, candidateBinding{name, current, bindingImport})
			}

		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				candidates = append(candidates, candidateBinding{name, current, bindingImport})
			}

		// A named function or class declaration is a binding like any other, and upstream reports
		// an unused one: `function foo() {}` at file scope with nothing calling it is
		// `test_functions`' first failing case.
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				candidates = append(candidates, candidateBinding{name, current, bindingVariable})
			}

		// Types, interfaces, enums and namespaces are checked exactly like values. Upstream states
		// this outright: "Enums and namespaces are treated the same as variables, classes,
		// functions, etc." An enum's MEMBERS are deliberately not candidates, because only the
		// enum itself is checked.
		case ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration,
			ast.KindModuleDeclaration,
			ast.KindTypeParameter:
			if name := current.Name(); name != nil && name.Kind == ast.KindIdentifier {
				candidates = append(candidates, candidateBinding{name, current, bindingType})
			}

		case ast.KindCatchClause:
			if variable := current.AsCatchClause().VariableDeclaration; variable != nil {
				if name := variable.Name(); name != nil && name.Kind == ast.KindIdentifier {
					candidates = append(candidates, candidateBinding{name, variable, bindingCaughtError})
				}
			}
		}

		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)

	return candidates
}

// patternOwner is the node a binding element's pattern belongs to: the parameter, the variable
// declaration or the catch binding it destructures, found by climbing patterns only. A pattern in a
// default's arrow function therefore belongs to that function's parameter, not to the outer one.
//
// The owner decides the element's kind, because upstream's definition type for every name in a
// pattern is its owner's. A name destructured out of a catch is a caught error: `caughtErrors:
// "none"` skips it and `caughtErrorsIgnorePattern` decides it, which cohere had the other way round
// until #6esg2nx. A name destructured out of a parameter is a parameter: `args: "none"` skips it and
// `argsIgnorePattern` decides it, while `varsIgnorePattern` applies to neither, which cohere had as
// a variable until #s48y8eg. Measured on the installed 8.71.0: `function f({ _a }) {}` is silent
// under `argsIgnorePattern: "^_"` and reports under `varsIgnorePattern: "^_"`.
func patternOwner(element *ast.Node) *ast.Node {
	current := element.Parent
	for current != nil && (current.Kind == ast.KindObjectBindingPattern ||
		current.Kind == ast.KindArrayBindingPattern || current.Kind == ast.KindBindingElement) {
		current = current.Parent
	}
	return current
}

// declaringParameter is the parameter a parameter candidate is declared by: the candidate's own
// declaration, or, for a name destructured out of one, the parameter that pattern belongs to.
func declaringParameter(candidate candidateBinding) *ast.Node {
	if candidate.declaration.Kind == ast.KindParameter {
		return candidate.declaration
	}
	return patternOwner(candidate.declaration)
}

// hasDefaultOnTheWayToItsParameter reports whether a parameter binding is given a default anywhere
// between its name and its parameter: `b = 1`, `{ b = 1 }`, `{ x: { b } = {} }` or `{ b } = {}`.
//
// Each of those is a write reference to the binding in ESLint's scope model, so upstream's
// isAfterLastUsedArg counts it as a later parameter that is used, and the ones before it are kept.
// Measured on the installed 8.71.0: `(a, { b = 1 })`, `(a, [b = 1])` and `(a, { b } = {})` with
// nothing read each report `b` and not `a`.
func hasDefaultOnTheWayToItsParameter(declaration *ast.Node) bool {
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBindingElement:
			if current.Initializer() != nil {
				return true
			}
		case ast.KindParameter:
			return current.AsParameterDeclaration().Initializer != nil
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		default:
			return false
		}
	}
	return false
}

// isGlobalScopeBinding reports whether a binding lives in the global scope, which `vars: "local"`
// leaves alone because another script can read it.
//
// Only a script has a global scope. A module's top level is its own scope, so in a module nothing is
// global and the option changes nothing; cohere reads a file with no top-level import or export as a
// script, as TypeScript does and as no-implicit-globals and strict here do. ESLint's flat config
// parses every `.ts` file as a module whatever its shape, so on a TypeScript file with no import or
// export the two engines disagree under this option, and cohere's reading is the one the compiler
// holds: that file's top-level names are visible to every other file in the program.
//
// In a script, measured on the installed 8.71.0 under `sourceType: "script"`:
//
//	var a; let b; const c; function f() {}; class C {}     global, at the top level
//	interface I {}; type T = 1; enum E {}; namespace N {}  global too, as the scope manager places them
//	{ var hoisted; }  if (x) { var v; }  for (var i ...)   global, because var climbs out of a block
//	{ let b; function inBlock() {} }  for (let j ...)      a block's own, so reported
//	function g(p) { var inner; }  namespace N { var m; }   a function's or a namespace's own
//
// Parameters, caught errors and imports are never global: a script cannot import, and the other two
// belong to the function or the catch that binds them.
func isGlobalScopeBinding(sourceFile *ast.SourceFile, candidate candidateBinding) bool {
	if ast.IsExternalModule(sourceFile) {
		return false
	}
	switch candidate.kind {
	case bindingParameter, bindingCaughtError, bindingImport:
		return false
	}
	declaration := candidate.declaration
	if declaration.Kind == ast.KindVariableDeclaration || declaration.Kind == ast.KindBindingElement {
		list := declaration.Parent
		for list != nil && list.Kind != ast.KindVariableDeclarationList {
			list = list.Parent
		}
		if list == nil {
			return false
		}
		if list.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst|ast.NodeFlagsUsing|ast.NodeFlagsAwaitUsing) == 0 {
			return enclosingVariableScope(list).Kind == ast.KindSourceFile
		}
		// A block-scoped declaration is global only when its statement sits directly in the file.
		statement := list.Parent
		return statement != nil && statement.Kind == ast.KindVariableStatement &&
			statement.Parent != nil && statement.Parent.Kind == ast.KindSourceFile
	}
	return declaration.Parent != nil && declaration.Parent.Kind == ast.KindSourceFile
}

// isDestructuredFromAnArray reports whether a binding is declared, or written, directly as an element
// of an array pattern, which is what `destructuredArrayIgnorePattern` applies to.
//
// Upstream asks it of the definition's ESTree parent and of every reference's, so the identifier has
// to BE the element. Measured on the installed 8.71.0, each of these follows:
//
//	const [_a, b] = list            skipped       so is a parameter `([_a, b])`, a catch `([_e])`,
//	const { x: [_a] } = n           skipped       and a for-of head `for (const [_k, v] of entries)`
//	const [_a = 1] = list           reports       a default makes the parent an AssignmentPattern
//	const [..._rest] = list         reports       so does rest, a RestElement
//	const { _a } = o                reports       an object pattern is not covered
//	let _a; [_a] = list             skipped       through the write, for a binding declared apart
//	let _a; [(_a)] = list           skipped       ESTree drops the parentheses
//	let _a; [_a = 1] = list         reports       as the declaration's default does
//	let _a; [[_a]] = nested         skipped       the inner array is the parent
//
// A read never sits in a pattern, so only writes are asked on the reference side. The check gates
// both reports, the unused one and "only used as a type", as upstream's `continue` does.
func isDestructuredFromAnArray(candidate candidateBinding, writes []*ast.Node) bool {
	declaration := candidate.declaration
	if declaration.Kind == ast.KindBindingElement && declaration.Name() == candidate.name &&
		declaration.Initializer() == nil && declaration.AsBindingElement().DotDotDotToken == nil &&
		declaration.Parent != nil && declaration.Parent.Kind == ast.KindArrayBindingPattern {
		return true
	}
	for _, write := range writes {
		element := ast.WalkUpParenthesizedExpressions(write.Parent)
		if element == nil {
			continue
		}
		if element.Kind == ast.KindArrayLiteralExpression {
			return true
		}
	}
	return false
}

// isExemptFromUnusedReport reports whether a binding nothing reads is nonetheless not a finding.
//
// Every arm here is an upstream exemption rather than a judgment of ours, and each one is the
// difference between a rule that is usable and one that reports every well-written file. The gates
// upstream asks first (`vars: "local"`, the ignore patterns, `caughtErrors` and `args` at "none",
// static-block classes and `using`) run before this, in analyzeUnusedBindings, in upstream's order.
func isExemptFromUnusedReport(
	ctx rule.Context,
	candidate candidateBinding,
	symbol *ast.Symbol,
	all []candidateBinding,
	settings NoUnusedVarsOptions,
	reads map[*ast.Symbol]bool,
) bool {
	// An import named `React` or `h` in a file containing JSX is used by the JSX transform, which
	// no source-level read expresses. Under the classic `"jsx": "react"` transform every element
	// compiles to `React.createElement`, so the import is load-bearing while looking untouched.
	//
	// Upstream carries this as an unconditional exemption at `mod.rs`'s `should_skip_symbol`, with
	// the comment that it cannot detect `jsxPragma` or tell `react` from `react-jsx`, so it allows
	// all cases rather than guessing. This reproduces that decision including its breadth.
	//
	// Measured, and it is the reason this guard is here rather than in a later pass: without it
	// this rule reported 512 findings on our tree and roughly 470 of them were `import React from
	// 'react'` at the top of a `.tsx` file. Every one was wrong, and every one looked exactly like
	// a genuine unused import in the output.
	if candidate.kind == bindingImport && isJsxFactoryImportInJsxFile(candidate) {
		return true
	}

	// A class declaration or a parameter that carries a decorator is used, whatever reads its name:
	// the decorator runs when the class is defined, with the class (or the parameter's index) as its
	// argument, so deleting the declaration changes what runs. A test that declares a class only so its
	// decorators register, or a parameter only so a parameter decorator records it, is the shape.
	//
	// A deliberate divergence from ESLint, by @system_cohere's ruling of 2026-10-03 (#ynneze5, Base's
	// 18 test sites). typescript-eslint 8.67's no-unused-vars has no decorator handling, and its scope
	// manager marks only the decorator function as referenced, so ESLint reports both shapes. The
	// divergence is in cohere's favor and is exact: no decorated declaration is dead. It is the
	// decorator, not a naming convention, that exempts, so `argsIgnorePattern` stays unset.
	//
	// It is the declaration the decorator sits on that is used, so a name destructured out of a
	// decorated parameter is still judged: deleting it from the pattern leaves the parameter, and its
	// position, where the decorator recorded them. ESLint reports it too (#s48y8eg).
	if (candidate.declaration.Kind == ast.KindClassDeclaration || candidate.declaration.Kind == ast.KindParameter) &&
		carriesDecorator(candidate.declaration) {
		return true
	}

	// Under the default `after-used`, only parameters following the last used one are reported
	// (`args: "none"` was settled before this). A name destructured out of a parameter is a
	// parameter for all of this except the last: upstream's isAfterLastUsedArg applies only when the
	// name's parent is the function itself, so `({ a, b }) => b` reports `a` under `after-used`.
	if candidate.kind == bindingParameter {
		// A setter's parameter cannot be removed: `set foo() {}` is a syntax error. Upstream
		// exempts it for exactly that reason at `allowed.rs`'s
		// `is_allowed_param_because_of_method`, and so does a constructor parameter carrying an
		// accessibility modifier, which declares a class member rather than a mere argument. A name
		// destructured out of a setter's parameter is silent upstream too, measured on 8.71.0.
		if isStructurallyRequiredParameter(declaringParameter(candidate)) {
			return true
		}
		if settings.Args != "all" && candidate.declaration.Kind == ast.KindParameter &&
			isParameterBeforeAUsedOne(ctx, candidate, all, reads) {
			return true
		}
	}

	// A mapped type's key is always used by the type it builds, and it has no other spelling.
	// `{ [K in Slots]?: string }` binds `K` and the value position may legitimately ignore it, as
	// in `{ [K in Slots]?: React.ReactNode }`. Upstream states this as a bare no-op arm at
	// `mod.rs:388` with the comment "Mapped type keys are always used within the type definition".
	//
	// This is a deliberate divergence from ESLint, which DOES report it: measured with ESLint's own
	// Linter API, `type S<T extends string> = { [K in T]?: string }` reports `K` as unused. oxc is
	// the differential authority, so oxc's answer ships, and the tree agrees with oxc: the one
	// occurrence in our own source is a correct mapped type that ESLint would have flagged.
	if candidate.kind == bindingType && candidate.declaration.Parent != nil &&
		candidate.declaration.Parent.Kind == ast.KindMappedType {
		return true
	}

	// A `this` parameter declares the receiver's type and is not a binding anyone can delete: it
	// occupies no argument position, so removing it changes the signature's meaning rather than
	// tidying it. ESLint reports it, measured; oxc never collects it as a symbol at all because its
	// parser models `this` separately from formal parameters. Same answer, reached differently.
	if candidate.kind == bindingParameter && candidate.name.Text() == "this" {
		return true
	}

	// A declaration with no body declares a signature rather than defining anything, so its
	// parameter names are documentation and removing them changes no behavior. Overload
	// signatures, ambient declarations, interface methods and abstract members all land here.
	if isInsideSignatureOrAmbientDeclaration(candidate.declaration) {
		return true
	}

	// A namespace written with a dotted name (`namespace foo.bar {}`) parses as nested module
	// declarations, and only the innermost one carries the body. The outer segments are path, not
	// bindings anybody could use, so reporting them accuses the author of not using syntax that
	// has no other spelling.
	if candidate.declaration.Kind == ast.KindModuleDeclaration &&
		candidate.declaration.Body() != nil &&
		candidate.declaration.Body().Kind == ast.KindModuleDeclaration {
		return true
	}

	if symbol == nil {
		// A name the checker cannot resolve is not provably unused. Reporting it would be a guess
		// in the direction that costs a false positive on code that is very likely fine.
		return true
	}

	return isUsedBinding(ctx, candidate, symbol, reads)
}

// isUsedBinding reports whether something keeps a resolved binding alive: a read, through its own
// symbol or the one an import aliases, the for-in leading-return idiom, or an export. It is the half
// of upstream's usedVariables this rule computes, asked both by isExemptFromUnusedReport and, for a
// name an ignore pattern set aside, by reportUsedIgnorePattern.
func isUsedBinding(ctx rule.Context, candidate candidateBinding, symbol *ast.Symbol, reads map[*ast.Symbol]bool) bool {
	if reads[symbol] {
		return true
	}

	// An imported name is an alias, and a read reached through a different route can land on the
	// declaration it aliases rather than on the local binding. `import { a } from 'a'; export { a }`
	// is upstream's clean case and both halves resolve to a symbol spelled `a` that is not the same
	// symbol, so the equality above misses it. Asking the alias question from the candidate's side
	// closes that gap without widening what a read means.
	//
	// An `import X = Y` is excluded, and this is the one place the two alias forms must be told
	// apart. There the aliased symbol is a LOCAL binding in the same file, so a read of `Y`
	// anywhere — including the `= Y` in the alias's own declaration — would count as a read of `X`
	// and keep every such alias permanently alive. A module import has no such hazard because the
	// symbol it aliases lives in another file that this rule never reads.
	if symbol.Flags&ast.SymbolFlagsAlias != 0 &&
		candidate.declaration.Kind != ast.KindImportEqualsDeclaration {
		if aliased := resolvedAlias(ctx, symbol); aliased != nil && reads[aliased] {
			return true
		}
	}

	// A `for (x in o)` or `for (x of it)` whose body's first statement is `return` marks the loop
	// variable used. The loop exists for its side effect, proving the collection is non-empty, and
	// the binding is the syntax that expresses it, so upstream declines to report it.
	//
	// This looks arbitrary and it is deliberately narrow: only a leading `return` counts, so
	// `for (const k in o) { log(k) }` is decided by the ordinary read test and
	// `for (const k in o) { sideEffect() }` still reports. Upstream's `is_used_in_for_of_loop` at
	// `usage.rs:203` is the whole specification and this reproduces its shape rather than
	// generalizing it.
	if isForInOrOfBindingWithLeadingReturn(candidate.declaration) ||
		hasForInOrOfWriteWithLeadingReturn(ctx, candidate, symbol) {
		return true
	}

	// An exported binding is read by a file this rule cannot see. See the rule's doc comment for
	// why that stays an assumption rather than becoming a program-wide lookup.
	//
	// EVERY declaration of the symbol is asked, not just the one this candidate came from, because
	// declaration merging can put the `export` on a sibling: `interface Foo {}` beside
	// `export const Foo = 'bar'` is one exported symbol wearing two declarations, and upstream
	// treats it as exported. Both orderings appear in the corpus as clean cases, which is what
	// forced the loop; asking only `candidate.declaration` reported the interface in both.
	//
	// A loop rather than an index, and the reason is the question being asked: "does ANY
	// declaration export this", where more declarations only mean more chances to match. An
	// index-zero test goes silent whenever the exporting declaration is not written first, and the
	// corpus pins exactly that ordering.
	for _, declaration := range rule.DeclarationsIn(ctx.SourceFile, symbol) {
		if isExportedBinding(declaration) {
			return true
		}
	}
	if isExportedBinding(candidate.declaration) {
		return true
	}

	return false
}

// hasRestSiblingDeclarationOrWrite is upstream's `hasRestSpreadSibling`, which `ignoreRestSiblings`
// turns on: the binding is declared, or written, as a property of an object pattern whose last
// element is a rest element.
//
// `const { removed, ...rest } = o` is the idiom for copying an object without one property, and
// `removed` exists only to be left out of `rest`. Without the option it reports, and the only other
// way to say "everything but this key" is a helper that cannot be typed without a cast.
//
// Upstream asks `hasRestSibling` of the identifier's ESTree parent, so the binding has to sit
// DIRECTLY in a property of that pattern. Measured on the installed 8.71.0, each of these follows:
//
//	const { a, ...rest } = o           silent        so is `{ a: b, ...rest }`, `{ [k]: a, ...rest }`
//	const { a = 1, ...rest } = o       reports `a`   a default makes the parent an AssignmentPattern
//	const { b: { c }, ...rest } = n    reports `c`   the inner pattern has no rest of its own
//	const [a, ...rest] = list          reports `a`   an array pattern is not covered
//	({ a, ...rest } = o)               silent        through the write, for a binding declared apart
//	({ x: (a), ...rest } = o)          silent        ESTree drops the parentheses
//
// The rest element itself is not its own sibling, so an unused `rest` still reports. Only writes are
// asked on the reference side: an identifier can sit in an object pattern only as a target, and
// every read has already made the binding used.
func hasRestSiblingDeclarationOrWrite(candidate candidateBinding, writes []*ast.Node) bool {
	declaration := candidate.declaration
	if declaration.Kind == ast.KindBindingElement && declaration.Name() == candidate.name &&
		declaration.Initializer() == nil && declaration.AsBindingElement().DotDotDotToken == nil &&
		declaration.Parent != nil && declaration.Parent.Kind == ast.KindObjectBindingPattern {
		elements := declaration.Parent.AsBindingPattern().Elements.Nodes
		if last := elements[len(elements)-1]; last.AsBindingElement().DotDotDotToken != nil {
			return true
		}
	}
	for _, write := range writes {
		if isWrittenBesideARestProperty(write) {
			return true
		}
	}
	return false
}

// isWrittenBesideARestProperty is the reference half: the identifier is a property's value in a
// destructuring object literal whose last property is a spread.
func isWrittenBesideARestProperty(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	var object *ast.Node
	if parent.Kind == ast.KindShorthandPropertyAssignment {
		// `{ a = 1 }` is an AssignmentPattern in ESTree, so a shorthand with a default is not a
		// property's direct value.
		if parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
			return false
		}
		object = parent.Parent
	} else {
		// `{ x: (a) }` is a direct value too, since ESTree drops the parentheses.
		property := ast.WalkUpParenthesizedExpressions(parent)
		if property == nil || property.Kind != ast.KindPropertyAssignment ||
			ast.SkipParentheses(property.AsPropertyAssignment().Initializer) != identifier {
			return false
		}
		object = property.Parent
	}
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	properties := object.AsObjectLiteralExpression().Properties.Nodes
	return len(properties) > 0 && properties[len(properties)-1].Kind == ast.KindSpreadAssignment
}

// isReadInsideAClosureAssignedToItself reports whether a read sits inside a function expression
// that is being assigned to the very binding being read.
//
// See the call site for the two neighbouring shapes that must NOT match, and why.
func isReadInsideAClosureAssignedToItself(identifier *ast.Node) bool {
	name := identifier.Text()

	for current := identifier.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		// A read inside a parameter's own default belongs to the closure's parameter scope, where
		// the parameter shadows anything outside spelled the same. `let a; a = function(a = a) {};`
		// is upstream's clean case and the inner `a` is the parameter, not the outer binding, so
		// climbing past this reported a variable that is genuinely used.
		case ast.KindParameter:
			return false

		case ast.KindFunctionExpression, ast.KindArrowFunction:
			// The closure must be the DIRECT right-hand side of the assignment. A call wrapping it
			// means it runs immediately, and a sequence means the assigned value is something
			// else; both are upstream passing cases.
			parent := current.Parent
			if parent == nil || parent.Kind != ast.KindBinaryExpression {
				return false
			}
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil ||
				binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Right != current {
				return false
			}
			target := unwrapAssignmentTarget(binary.Left)
			return target != nil && target.Kind == ast.KindIdentifier && target.Text() == name

		case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindSourceFile:
			return false
		}
	}

	return false
}

// isDiscardedSequenceOperandRead reports whether a read sits in a comma sequence position whose
// value nothing can observe.
//
// See the call site for why only the left operand counts.
func isDiscardedSequenceOperandRead(identifier *ast.Node) bool {
	current := identifier
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindCommaToken {
				return false
			}
			if binary.Left == current {
				return true
			}
			// The right operand is the sequence's value, so it inherits the sequence's own fate
			// and the walk continues rather than answering here.
			current = parent

		default:
			return false
		}
	}
	return false
}

// isSelfReassignmentRead reports whether a read of a binding exists only to write that same binding.
//
// The shape is `a = a + 1`, `a += 1`, `a++`. Each loads the binding, so the load is a genuine read
// in the dataflow sense, and upstream nonetheless declines to count it: the value goes nowhere but
// back into the same name, so no observer of the program can tell whether the binding exists.
//
// Two escapes make the read count again, and both matter on real code:
//
// A different variable scope. `let cancel = () => {}; function f(){ cancel = cancel?.(); }` is
// clean upstream, because the assignment sits in a new function scope and the value can be observed
// on a later call. The same statement in a plain block is a finding. That is `usage.rs:503`'s
// variable-scope comparison, and the corpus pins both halves in `test_vars_self_use`.
//
// A loop. `for (let i = 0; i < 10; i++) { chain = chain.extend(); }` is clean, because the back
// edge makes each store observable by the next iteration's load. Outside a loop the same statement
// reports. This is the case a liveness dataflow would answer differently, and the rule's doc
// comment records why upstream's syntactic answer is reproduced instead.
func isSelfReassignmentRead(identifier *ast.Node, declaringScope *ast.Node) bool {
	name := identifier.Text()

	for current := identifier; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent

		switch parent.Kind {
		// Transparent wrappers on the way out to the operation that decides.
		case ast.KindParenthesizedExpression,
			ast.KindPropertyAccessExpression,
			ast.KindElementAccessExpression,
			ast.KindNonNullExpression,
			ast.KindAsExpression,
			ast.KindCallExpression:
			// A call whose ARGUMENTS contain the reference is a genuine use: the value escapes
			// into the callee. Only the callee position stays transparent, which is what keeps
			// `cancel = cancel?.()` self-referential while `f(a)` is not.
			if parent.Kind == ast.KindCallExpression &&
				parent.AsCallExpression().Expression != current {
				return false
			}
			continue

		// A read feeding another binding's declaration or a property definition escapes, so the
		// value is observable and this is a real use.
		case ast.KindVariableDeclaration, ast.KindPropertyDeclaration, ast.KindJsxExpression:
			return false

		case ast.KindPrefixUnaryExpression:
			if !reference.IsUpdateOperator(parent.AsPrefixUnaryExpression().Operator) {
				return false
			}
			if current != identifier {
				// The update targets a property reached through this binding: `acc[k]++` writes
				// to the object, so `acc` itself is genuinely read. Upstream notes this at
				// `usage.rs:734` and separates it from `(a++, 0)` by exactly this test.
				return false
			}
			return isDiscardedResult(parent) && !isInsideLoop(parent)

		case ast.KindPostfixUnaryExpression:
			if !reference.IsUpdateOperator(parent.AsPostfixUnaryExpression().Operator) {
				return false
			}
			if current != identifier {
				return false
			}
			return isDiscardedResult(parent) && !isInsideLoop(parent)

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				// An ordinary operator such as `+` is transparent: `a = a + 1` reaches the
				// assignment through it, and that is the shape this whole function is about.
				continue
			}
			// The reference sits on the right of an assignment. It is self-referential only when
			// the target is the same name, in the same variable scope, outside a loop.
			target := unwrapAssignmentTarget(binary.Left)
			if target == nil || target.Kind != ast.KindIdentifier || target.Text() != name {
				return false
			}
			// Upstream compares the REFERENCE's variable scope against the BINDING's declaring
			// scope, not the two identifiers against each other: those are always in the same
			// scope as one another, so that comparison is vacuously true and the exemption never
			// fires. `let cancel = () => {}; export function close() { cancel = cancel?.() }` is
			// upstream's clean case and it is clean precisely because the assignment sits in a new
			// function scope where the stored value can be observed on a later call. The same
			// statement in a plain block reports.
			if enclosingVariableScope(identifier) == declaringScope && isDiscardedResult(parent) &&
				!isInsideLoop(parent) {
				return true
			}
			return false

		default:
			return false
		}
	}

	return false
}

// isDiscardedResult reports whether an expression's value goes nowhere.
//
// This is what separates `a++` from `b = a++`. Both load `a`, but only the second lets anything
// observe the loaded value, and upstream counts only the second as a use. Its name for this is
// `is_discarded_read` at `usage.rs:709`.
//
// An expression's result is discarded when it sits directly as a statement, or as a non-final
// element of a comma sequence whose own result is discarded. Everything else — an initializer, an
// argument, an operand, a return value, a condition — consumes it.
func isDiscardedResult(expression *ast.Node) bool {
	current := expression
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent

		case ast.KindExpressionStatement:
			return true

		// A cast is transparent to where the value goes. `a = a++ as any` discards through the
		// `as` exactly as `a = a++` does, and upstream reports both.
		case ast.KindAsExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
			current = parent

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			// A comma sequence discards every operand but the last, and the last inherits the
			// sequence's own fate. `(a++, 0)` discards `a++`; `(0, x++)` does not, because the
			// update is the value the sequence yields.
			if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
				if binary.Left == current {
					return true
				}
				current = parent
				continue
			}
			// An assignment feeding the value straight back into the SAME binding discards it as
			// surely as a bare statement does: `a = ++a` and `a = (0, ++a)` leave nothing any
			// reader can observe, and upstream reports both. The target has to be the same name,
			// otherwise `b = a++` is a genuine escape into `b`.
			if binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
				binary.Right == current {
				target := unwrapAssignmentTarget(binary.Left)
				if target != nil && target.Kind == ast.KindIdentifier &&
					target.Text() == updatedBindingName(expression) {
					current = parent
					continue
				}
			}
			return false

		// A `for` head's initializer and incrementor are evaluated for effect, so a value there is
		// discarded. The test is not: the loop reads it.
		case ast.KindForStatement:
			forStatement := parent.AsForStatement()
			return forStatement.Initializer == current || forStatement.Incrementor == current

		default:
			return false
		}
	}
	return false
}

// updatedBindingName returns the name an update expression writes to, or the empty string.
//
// Used to tell `a = ++a` from `b = ++a`: the first discards the value back into the binding it came
// from and the second lets `b` observe it.
func updatedBindingName(expression *ast.Node) string {
	var operand *ast.Node
	switch expression.Kind {
	case ast.KindPrefixUnaryExpression:
		operand = expression.AsPrefixUnaryExpression().Operand
	case ast.KindPostfixUnaryExpression:
		operand = expression.AsPostfixUnaryExpression().Operand
	case ast.KindBinaryExpression:
		operand = expression.AsBinaryExpression().Left
	default:
		return ""
	}
	if operand = unwrapAssignmentTarget(operand); operand != nil &&
		operand.Kind == ast.KindIdentifier {
		return operand.Text()
	}
	return ""
}

// unwrapAssignmentTarget strips the wrappers an assignment target can wear.
func unwrapAssignmentTarget(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		default:
			return node
		}
	}
	return nil
}

// sameVariableScope reports whether two nodes sit in the same function-level scope.
//
// Function-level rather than block-level, deliberately: upstream compares *variable* scopes, so a
// plain block does not separate a binding from a write to it while a function does. That is the
// difference between `{ cancel = cancel?.() }` reporting and `function f(){ cancel = cancel?.() }`
// staying clean, both of which the corpus pins.
func sameVariableScope(first *ast.Node, second *ast.Node) bool {
	return enclosingVariableScope(first) == enclosingVariableScope(second)
}

// enclosingVariableScope returns the nearest function, module, or class static block containing a
// node, or nil at file scope.
func enclosingVariableScope(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
			ast.KindClassStaticBlockDeclaration,
			ast.KindModuleDeclaration,
			ast.KindSourceFile:
			return current
		}
	}
	return nil
}

// isInsideLoop reports whether a node sits in a loop's body, test, or update.
//
// A loop's back edge is what makes a self-assignment observable: the store this iteration performs
// is loaded by the next one. Upstream reaches the same conclusion through two separate predicates,
// `is_in_for_loop_test_or_update` and `is_in_loop_body`, and both are collapsed here because the
// answer they feed is the same.
//
// The walk stops at a function boundary. A loop outside a closure does not make an assignment
// inside that closure observable, since the closure's body runs when it is called rather than on
// each iteration.
func isInsideLoop(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindForStatement,
			ast.KindForInStatement,
			ast.KindForOfStatement,
			ast.KindWhileStatement,
			ast.KindDoStatement:
			return true
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindSourceFile:
			return false
		}
	}
	return false
}

// isSelfReferenceWithinOwnDeclaration reports whether an identifier refers to the very function or
// class that lexically contains it.
//
// `function foox() { return foox(); }` is upstream's own failing case: a recursive function nothing
// else calls is unreachable, and counting the recursive call as a use is what would keep an entire
// abandoned call graph looking alive. This tree's `internal/unused` package reaches the identical
// verdict from an argument rather than from a corpus; this is the corpus confirming it.
//
// The comparison is on SYMBOL identity rather than on the spelled name, and that is load-bearing
// rather than tidy. `function foo() { var foo = 1; return foo }` returns the inner variable, not the
// function, and a name comparison calls it a self-reference and reports a variable upstream calls
// clean. The same shape arrives as a shadowing parameter in `function foo(foo) { return foo }`.
// Both are upstream passing cases and both were false positives here until identity replaced text.
func isSelfReferenceWithinOwnDeclaration(ctx rule.Context, identifier *ast.Node, symbol *ast.Symbol) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		// A type that only appears in its own definition is not used. `type Foo = Array<Foo>` and
		// `interface LinkedList<T> { next: LinkedList<T> }` are upstream's own failing cases, and
		// its documentation states the rule outright: "A type or interface is not considered to be
		// used if it is only ever used in its own definition." Same judgment as the recursive
		// function below, reached for the same reason.
		case ast.KindTypeAliasDeclaration,
			ast.KindInterfaceDeclaration:
			if declarationDeclaresSymbol(ctx, current.Name(), symbol) {
				return true
			}

		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindClassDeclaration,
			ast.KindClassExpression,
			ast.KindArrowFunction:
			if declarationDeclaresSymbol(ctx, current.Name(), symbol) {
				return true
			}
			// A variable declarator holding a function expression names it: `const foo = () => {
			// return foo }` is upstream's second self-use failure, and the arrow is anonymous, so
			// the name has to come from the declarator.
			if current.Parent != nil && current.Parent.Kind == ast.KindVariableDeclaration &&
				declarationDeclaresSymbol(ctx, current.Parent.Name(), symbol) {
				return true
			}

		case ast.KindSourceFile:
			return false
		}
	}

	return false
}

// declarationDeclaresSymbol reports whether a declaration's name node binds the given symbol.
func declarationDeclaresSymbol(ctx rule.Context, name *ast.Node, symbol *ast.Symbol) bool {
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	return ctx.TypeChecker.GetSymbolAtLocation(name) == symbol
}

// isForInOrOfBindingWithLeadingReturn reports whether a binding is the loop variable of a
// `for...in` or `for...of` whose body returns immediately.
//
// See the call site for why this exemption exists and why it is not generalized.
func isForInOrOfBindingWithLeadingReturn(declaration *ast.Node) bool {
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindForInStatement, ast.KindForOfStatement:
			// The typed accessor rather than `Node.Body()`. Measured: `Body()` returns nil on both
			// loop kinds, so the exemption below silently never fired and sixteen of upstream's
			// clean cases reported. A nil from a convenience accessor reads exactly like a loop
			// with no body, which is why this cost a full measurement cycle to find.
			body := current.AsForInOrOfStatement().Statement
			if body == nil {
				return false
			}
			if body.Kind == ast.KindReturnStatement {
				return true
			}
			if body.Kind == ast.KindBlock {
				statements := body.Statements()
				return len(statements) > 0 && statements[0].Kind == ast.KindReturnStatement
			}
			return false

		// A function boundary means the loop, if any, is outside the binding's own scope.
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindSourceFile:
			return false
		}
	}
	return false
}

// isJsxFactoryImportInJsxFile reports whether a binding is a JSX factory import in a file using JSX.
//
// See the call site for why the name test is `React` or `h` specifically and why the breadth is
// upstream's rather than ours.
func isJsxFactoryImportInJsxFile(candidate candidateBinding) bool {
	name := candidate.name.Text()
	if name != "React" && name != "h" {
		return false
	}

	sourceFile := ast.GetSourceFileOfNode(candidate.declaration)
	if sourceFile == nil {
		return false
	}
	return strings.HasSuffix(sourceFile.FileName().AsString(), ".tsx") ||
		strings.HasSuffix(sourceFile.FileName().AsString(), ".jsx")
}

// hasForInOrOfWriteWithLeadingReturn covers the same exemption for a binding declared apart from
// the loop that assigns it.
//
// `var name; for (name in obj) return;` binds the loop variable by assignment rather than in the
// head, so the declaration has no `ForInStatement` above it and the climb in the sibling helper
// cannot see the loop. Both spellings appear in upstream's corpus, twelve cases apart, and they are
// the same judgment reached from two directions.
func hasForInOrOfWriteWithLeadingReturn(ctx rule.Context, candidate candidateBinding, symbol *ast.Symbol) bool {
	sourceFile := ast.GetSourceFileOfNode(candidate.declaration)
	if sourceFile == nil {
		return false
	}

	name := candidate.name.Text()
	found := false

	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		if current.Kind == ast.KindIdentifier && current.Text() == name &&
			isForInOrOfBindingWithLeadingReturn(current) &&
			ctx.TypeChecker.GetSymbolAtLocation(current) == symbol {
			found = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	return found
}

// isStructurallyRequiredParameter reports whether a parameter cannot be deleted without changing
// what the surrounding code means.
//
// Two shapes. A setter must take exactly one parameter, so an unused one is required by the
// grammar. And a constructor parameter carrying `public`, `private`, `protected` or `readonly`
// declares a class property, so it is doing work even when the constructor body never reads it.
// carriesDecorator reports whether a declaration's own modifiers include a decorator.
func carriesDecorator(declaration *ast.Node) bool {
	modifiers := declaration.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindDecorator {
			return true
		}
	}
	return false
}

func isStructurallyRequiredParameter(parameter *ast.Node) bool {
	owner := parameter.Parent
	if owner == nil {
		return false
	}

	if owner.Kind == ast.KindSetAccessor {
		return true
	}

	// An `override` method is deliberately not a third shape. oxc exempts its parameters (its
	// `is_allowed_param_because_of_method` family), and this port once did too; typescript-eslint does
	// not, and TypeScript lets an override declare fewer parameters than the method it overrides, so an
	// unused trailing one can simply go. The arm cost api-phi-health three findings ESLint reports
	// (#c6jhg93), each fixed upstream by deleting the parameter.

	if owner.Kind == ast.KindConstructor {
		if modifiers := parameter.Modifiers(); modifiers != nil {
			for _, modifier := range modifiers.Nodes {
				switch modifier.Kind {
				case ast.KindPublicKeyword,
					ast.KindPrivateKeyword,
					ast.KindProtectedKeyword,
					ast.KindReadonlyKeyword,
					ast.KindOverrideKeyword:
					return true
				}
			}
		}
	}

	return false
}

// isAmbientInterfaceMember reports whether a declaration belongs to an interface.
//
// See the ambient-block reasoning in isInsideSignatureOrAmbientDeclaration for why an interface is
// treated differently from a type alias or a class in the same position.
func isAmbientInterfaceMember(declaration *ast.Node) bool {
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindInterfaceDeclaration:
			return true
		case ast.KindModuleBlock, ast.KindSourceFile:
			return false
		}
	}
	return false
}

// enclosingAmbientModuleBlock returns the module block a node sits in, or nil.
func enclosingAmbientModuleBlock(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindModuleBlock:
			return current
		case ast.KindSourceFile:
			return nil
		}
	}
	return nil
}

// moduleBlockHasExplicitExports reports whether a module block states an interface of its own.
//
// An `export` modifier on a member, an `export {}` or `export ... from`, or an `export =` all
// count. See the call site for why the distinction decides whether the block's contents are judged.
func moduleBlockHasExplicitExports(block *ast.Node) bool {
	for _, statement := range block.Statements() {
		switch statement.Kind {
		case ast.KindExportDeclaration, ast.KindExportAssignment:
			return true
		}
		if hasExportModifier(statement) {
			return true
		}
	}
	return false
}

// isInsideSignatureOrAmbientDeclaration reports whether a binding belongs to a declaration that
// declares a shape rather than defining a value.
//
// An overload signature, an interface method, an abstract member, and anything under `declare` all
// name their parameters for the reader's benefit; nothing runs, so nothing can use them. Upstream
// reaches the same answer through `formal_parameters.kind.is_signature()` at `mod.rs`'s
// `should_skip_symbol`.
func isInsideSignatureOrAmbientDeclaration(declaration *ast.Node) bool {
	for current := declaration; current != nil; current = current.Parent {
		if ast.GetCombinedNodeFlags(current)&ast.NodeFlagsAmbient != 0 {
			// An ambient module exempts its contents only when it declares NO explicit exports.
			// `declare module 'm' { type T = any; }` has no export, so everything in it is part of
			// the ambient surface and nothing is judged. Add one export and the block becomes a
			// module with a stated interface, so upstream judges the rest of its contents against
			// that interface: `declare module 'foo' { type Test = any; const x = 1; export = x; }`
			// reports `Test`.
			//
			// Upstream states this as `is_ambient_external_module_without_explicit_exports` at
			// `allowed.rs:93`, and the same test governs ambient namespaces at `:105`. Reading the
			// ambient flag alone exempted both shapes and cost four of this port's missed cases.
			if block := enclosingAmbientModuleBlock(current); block != nil &&
				moduleBlockHasExplicitExports(block) {
				return false
			}
			// An ambient block still gets its type-alias and class type parameters judged, while
			// an interface's are left alone. That split is upstream's and it is visible only in
			// the snapshot: `declare module 'bun:test' { type Matchers2<T> = {} }` reports `T` and
			// `{ class MyClass<T> {} }` reports both the class and `T`, while
			// `declare module 'vitest' { interface Matchers<T> {...} }` is CLEAN.
			//
			// The reason is what an interface in an ambient module is for. It is the declaration
			// merging idiom: a library augments a type it does not own, and the type parameter is
			// dictated by the interface being merged into rather than chosen here. A type alias or
			// a class in the same position declares something new, so an unused parameter on it is
			// the author's own loose end.
			if isAmbientInterfaceMember(declaration) {
				return true
			}
			if declaration.Kind == ast.KindTypeParameter {
				return false
			}
			return true
		}

		switch current.Kind {
		case ast.KindMethodSignature,
			ast.KindCallSignature,
			ast.KindConstructSignature,
			ast.KindIndexSignature,
			ast.KindFunctionType,
			ast.KindConstructorType:
			return true

		case ast.KindFunctionDeclaration,
			ast.KindMethodDeclaration,
			ast.KindConstructor:
			// A body-less declaration is an overload signature or an abstract member, and its
			// parameter names are documentation. But that only holds when the declaration IS the
			// thing being judged or encloses it: `export namespace N { function foo() }` has a
			// body-less `foo` whose own NAME is the candidate, and upstream reports it. The
			// signature exemption is about what a signature's parameters mean, not about the
			// signature's own name.
			if current == declaration {
				return false
			}
			return current.Body() == nil

		case ast.KindSourceFile:
			return false
		}
	}
	return false
}

// isParameterBeforeAUsedOne implements the default `args: "after-used"` behavior.
//
// A parameter can only be removed if every parameter after it is also removable, because argument
// position is the calling convention: deleting the second of three parameters silently reassigns
// the third. So upstream reports only the trailing run of unused parameters, and this reproduces
// that by asking whether any later parameter in the same list is read.
func isParameterBeforeAUsedOne(
	ctx rule.Context,
	candidate candidateBinding,
	all []candidateBinding,
	reads map[*ast.Symbol]bool,
) bool {
	owner := declaringParameter(candidate).Parent
	if owner == nil {
		return false
	}

	seenSelf := false
	for _, other := range all {
		if other.declaration == candidate.declaration {
			seenSelf = true
			continue
		}
		// A later parameter is any name its function's parameters declare, destructured ones included,
		// as upstream's getDeclaredVariables lists them: `(a, { b }) => b` keeps `a`.
		if !seenSelf || other.kind != bindingParameter || declaringParameter(other).Parent != owner {
			continue
		}
		parameter := declaringParameter(other)
		// A later parameter carrying a modifier counts as used even when nothing reads it, because
		// it declares a class property and so cannot be removed. Upstream says this at the line:
		// "has_modifier() to handle: constructor(unused: number, public property: string) {}".
		// Without it, `constructor(baz: string, private logger: Logger)` reports `baz`, which is a
		// parameter that genuinely cannot be deleted without breaking the property behind it.
		if isStructurallyRequiredParameter(parameter) {
			return true
		}
		// A later parameter carrying a decorator is used (see isExemptFromUnusedReport), so it shields
		// the ones before it like any used parameter: `(_input, @Inject() context)` keeps `_input`,
		// which holds the decorated parameter's position. Missing at 92059a9, found by api's Base pass.
		if carriesDecorator(parameter) {
			return true
		}
		// A later parameter with a default value shields the ones before it, as upstream's
		// `isAfterLastUsedArg` does: it asks whether a later parameter has any reference, and a default
		// is a write reference. So `(a, b, c = {})` with only `a` read reports `c`, as assigned and
		// never used, and not `b`. Measured with lintText on the installed typescript-eslint, and
		// the shape of api-phi-health's FakeStripePaymentProcessor.ts:122 (#c6jhg93). A default
		// anywhere in a destructured parameter's pattern counts the same way (#s48y8eg).
		if hasDefaultOnTheWayToItsParameter(other.declaration) {
			return true
		}
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(other.name); symbol != nil && reads[symbol] {
			return true
		}
	}

	return false
}

// isExportedBinding reports whether a declaration leaves the file.
//
// Climbing rather than testing the declaration itself, because the export modifier sits on the
// statement while the binding sits several nodes down: `export const { a, b } = o;` puts `export`
// on the variable statement and `a` inside a binding pattern inside a declaration list.
func isExportedBinding(declaration *ast.Node) bool {
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		// A parameter of an exported function is NOT exported. The function leaves the file; its
		// parameter list does not, and no importer can read a parameter name. Climbing past these
		// exempted every parameter of every exported function, which was 18 of this port's 35
		// missed cases and the largest single group of them.
		//
		// The boundary is the declaration that OWNS a parameter list or a body, rather than a
		// scope boundary generally: `export const f = () => { const inner = 1; }` must still stop
		// at the arrow, because `inner` is not exported either.
		// A parameter is never exported, whatever encloses it. The function leaves the file; its
		// parameter list does not, and no importer can read a parameter name. This is checked on
		// the parameter ITSELF rather than only on the way past one, because a parameter is where
		// the climb starts and a `current != declaration` guard would let it walk straight up to
		// the exported function it belongs to. That was the bug in the first attempt at this, and
		// it read as the narrowing simply having no effect.
		case ast.KindParameter:
			return false

		// A TYPE PARAMETER is never exported either, for the same reason and with the same
		// mechanics: `export interface M<T> {}` exports `M`, and no importer can name `T`. It has
		// to answer on the node ITSELF rather than on the way past one, because a type parameter is
		// where the climb starts. Upstream reports `T` here and reports the `R` bound by an `infer`
		// in an exported conditional type, both of which stayed silent while the climb walked up to
		// the exported declaration.
		case ast.KindTypeParameter:
			return false

		// The declarations that own a body. Anything reached THROUGH one of these is inside a
		// scope rather than at the top of an exported statement, so it does not leave the file:
		// `export const f = () => { const inner = 1; }` exports `f` and not `inner`. The
		// `current != declaration` guard is what lets the declaration itself still be judged, since
		// `export function f() {}` is a genuine export of `f`.
		case ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindConstructor,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindBlock,
			// A namespace body is a scope, not a re-export. `export namespace N { function foo() }`
			// exports `N`; `foo` stays inside and upstream reports it. Only a declaration carrying
			// its own `export` inside the namespace leaves, and that is caught by the modifier test
			// below on the inner declaration itself rather than by the climb.
			ast.KindModuleBlock:
			if current != declaration {
				return false
			}

		// An `import X = Y` leaves the file only when it carries its own `export`. Sitting inside an
		// exported namespace does not export it: `export namespace Bar { import TheFoo = Foo; }`
		// exports `Bar`, and `TheFoo` is a local alias upstream reports. Answering here rather than
		// continuing the climb is what separates the two.
		case ast.KindImportEqualsDeclaration:
			return hasExportModifier(current)

		case ast.KindVariableStatement,
			ast.KindFunctionDeclaration,
			ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration,
			ast.KindModuleDeclaration:
			if hasExportModifier(current) {
				return true
			}
		case ast.KindSourceFile:
			return false
		}
	}
	return false
}

// hasExportModifier reports whether a statement carries `export` or `export default`.
func hasExportModifier(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindExportKeyword || modifier.Kind == ast.KindDefaultKeyword {
			return true
		}
	}
	return false
}

// noUnusedVarsUsedIgnoredMessage renders typescript-eslint's `usedIgnoredVar`: a name an ignore
// pattern marks as deliberately unused, which is used.
func noUnusedVarsUsedIgnoredMessage(name string, option string, pattern string, noun string) rule.Message {
	return rule.Message{
		Id: "usedIgnoredVar",
		Description: "'" + name + "' matches " + option + " " + pattern + ", which marks " + noun +
			" as deliberately unused, but it is used. The name now says the opposite of what the code " +
			"does, and the pattern that should catch a forgotten binding is hiding this one from every " +
			"later check. Rename it so it no longer matches.",
	}
}

// reportUsedIgnoredName reports `usedIgnoredVar` on a name its kind's ignore pattern set aside, when
// reportUsedIgnorePattern is on and the name is used, at the node an unused finding would take.
// A name the checker cannot resolve is not provably used, so it stays silent.
func reportUsedIgnoredName(
	ctx rule.Context,
	candidate candidateBinding,
	symbol *ast.Symbol,
	writes []*ast.Node,
	settings NoUnusedVarsOptions,
	reads map[*ast.Symbol]bool,
	option string,
	pattern string,
	noun string,
) {
	if !settings.ReportUsedIgnorePattern || symbol == nil || !isUsedBinding(ctx, candidate, symbol, reads) {
		return
	}
	ctx.ReportNode(unusedBindingReportNode(candidate, writes), noUnusedVarsUsedIgnoredMessage(candidate.name.Text(), option, pattern, noun))
}

// ignorePatternFor names the ignore pattern a kind of binding answers to, as nameMatchesIgnorePattern
// chooses it, and what the pattern calls that kind.
func ignorePatternFor(kind unusedBindingKind, settings NoUnusedVarsOptions) (string, string, string) {
	switch kind {
	case bindingParameter:
		return "argsIgnorePattern", settings.ArgsIgnorePattern, "an argument"
	case bindingCaughtError:
		return "caughtErrorsIgnorePattern", settings.CaughtErrorsIgnorePattern, "a caught error"
	}
	return "varsIgnorePattern", settings.VarsIgnorePattern, "a variable"
}

// hasStaticBlock reports whether a class's body holds a `static {}` block, which
// ignoreClassWithStaticInitBlock exempts: the block runs when the class is defined, so a class nothing
// reads can still be there for its side effect.
func hasStaticBlock(class *ast.Node) bool {
	for _, member := range class.Members() {
		if member.Kind == ast.KindClassStaticBlockDeclaration {
			return true
		}
	}
	return false
}

// isUsingDeclaration reports whether a variable is declared with `using` or `await using`, which
// ignoreUsingDeclarations exempts: the declaration exists to dispose of its value at the end of the
// scope, whether or not anything reads it.
func isUsingDeclaration(candidate candidateBinding) bool {
	if candidate.kind != bindingVariable || candidate.declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	list := candidate.declaration.Parent
	// The Using bit alone: `await using` is Const|Using, so masking with it would also match `const`.
	return list != nil && list.Kind == ast.KindVariableDeclarationList && list.Flags&ast.NodeFlagsUsing != 0
}

// matchesIgnorePattern reports whether a binding's name says the omission is deliberate.
func matchesIgnorePattern(candidate candidateBinding, settings NoUnusedVarsOptions) bool {
	return nameMatchesIgnorePattern(candidate.name.Text(), candidate.kind, settings)
}

// nameMatchesIgnorePattern reports whether a name, declared as this kind of binding, matches the
// ignore pattern configured for that kind.
//
// An unconfigured pattern ignores nothing, for every kind, which is ESLint's default. This rule used
// to default to a leading underscore, oxc's answer, back when oxc was the reference the differential
// compared against; the parity doctrine (2026-10-01) made ESLint the floor, and under it a default
// that exempts `_deliberate` is a finding ESLint makes and cohere does not. It was also an allowance
// built into the rule: the ahra config had to write `^$` three times to switch it off. See the rule's
// doc comment for the history.
//
// Asked by name rather than by candidate so the report can ask the same question of a name the
// binding does not have yet: whether prefixing an underscore would silence it, which decides
// whether the message may offer that escape.
func nameMatchesIgnorePattern(name string, kind unusedBindingKind, settings NoUnusedVarsOptions) bool {
	pattern := settings.VarsIgnorePattern
	switch kind {
	case bindingParameter:
		pattern = settings.ArgsIgnorePattern
	case bindingCaughtError:
		pattern = settings.CaughtErrorsIgnorePattern
	}
	if pattern == "" {
		return false
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// An unparseable pattern ignores nothing rather than everything. The other direction would
		// silence the whole rule on a typo in a config file, which is the failure that looks clean.
		return false
	}
	return compiled.MatchString(name)
}

// NoUnusedVarsOptions is the rule's configuration surface.
//
// Every field is taken from oxc's `options.rs` rather than from the inventory, which records this
// rule as having no options and is wrong: the option module alone is 891 lines. The defaults are
// ESLint's. Two are not the zero value and are pinned with fixtures: `args` defaults to
// `after-used` rather than to `all`, and `caughtErrors` defaults to `all`. Every ignore pattern
// defaults to nothing, so an unconfigured rule exempts no name.
//
// Every field is read. `vars`, `caughtErrors` and `destructuredArrayIgnorePattern` were decoded and
// never consulted until #6esg2nx, so a config setting them changed nothing; the registry guard
// TestEveryOptionFieldARuleDecodesIsRead now fails on a field like that.
//
// `ignoreClassWithStaticInitBlock`, `ignoreUsingDeclarations` and `reportUsedIgnorePattern` are
// typescript-eslint 8.71's and ESLint 10.8.1's core schema's too, and `enableAutofixRemoval` is
// typescript-eslint's alone, all ported against upstream's rows (#e1zk9s0). Each defaults to false,
// the zero value.
type NoUnusedVarsOptions struct {
	// Vars is `all` or `local`. Default `all`. `local` skips a binding in the global scope, which
	// only a script has; see isGlobalScopeBinding.
	Vars string `json:"vars"`
	// VarsIgnorePattern is a regular expression naming variables to skip.
	VarsIgnorePattern string `json:"varsIgnorePattern"`
	// Args is `all`, `after-used`, or `none`. Default `after-used`.
	Args string `json:"args"`
	// ArgsIgnorePattern is a regular expression naming parameters to skip.
	ArgsIgnorePattern string `json:"argsIgnorePattern"`
	// CaughtErrors is `all` or `none`. Default `all`, which is oxc's answer and differs from older
	// ESLint releases where it defaulted to `none`. `none` skips every caught error, including each
	// name destructured out of one.
	CaughtErrors string `json:"caughtErrors"`
	// CaughtErrorsIgnorePattern is a regular expression naming catch bindings to skip, and names
	// destructured out of a catch binding with them.
	CaughtErrorsIgnorePattern string `json:"caughtErrorsIgnorePattern"`
	// IgnoreRestSiblings keeps a binding alive when it sits beside a rest element, which is the
	// idiom for omitting a property: `const { removed, ...rest } = o`. Default false, as upstream's.
	// See hasRestSiblingDeclarationOrWrite.
	IgnoreRestSiblings bool `json:"ignoreRestSiblings"`
	// DestructuredArrayIgnorePattern is a regular expression naming array-destructured elements to
	// skip: `const [_first, second] = pair`. It is asked before the other patterns and applies to a
	// parameter or a caught error destructured from an array too. See isDestructuredFromAnArray.
	DestructuredArrayIgnorePattern string `json:"destructuredArrayIgnorePattern"`
	// IgnoreClassWithStaticInitBlock exempts a class whose body holds a `static {}` block, which runs
	// when the class is defined. See hasStaticBlock.
	IgnoreClassWithStaticInitBlock bool `json:"ignoreClassWithStaticInitBlock"`
	// IgnoreUsingDeclarations exempts a `using` or `await using` binding, which exists to dispose of
	// its value. See isUsingDeclaration.
	IgnoreUsingDeclarations bool `json:"ignoreUsingDeclarations"`
	// ReportUsedIgnorePattern reports `usedIgnoredVar` on a name an ignore pattern sets aside that is
	// used, so the pattern cannot hide a name that lies about itself. See reportUsedIgnoredName.
	ReportUsedIgnorePattern bool `json:"reportUsedIgnorePattern"`
	// EnableAutofixRemoval is typescript-eslint's alone (core ESLint has no fixer here). Its one key,
	// imports, makes the removal of an unused import a fix the edit engine applies; off, the default,
	// the same removal is offered as a suggestion. See importRemoval.
	EnableAutofixRemoval NoUnusedVarsAutofixRemoval `json:"enableAutofixRemoval"`
}

// NoUnusedVarsAutofixRemoval is enableAutofixRemoval's object. Decoded strictly like the rest, so a key
// upstream's schema does not name is refused.
type NoUnusedVarsAutofixRemoval struct {
	Imports bool `json:"imports"`
}

// noUnusedVarsOptionFields is NoUnusedVarsOptions without its UnmarshalJSON, the shape the object
// form decodes into.
type noUnusedVarsOptionFields NoUnusedVarsOptions

// noUnusedVarsModes are the spellings upstream's schema allows for each mode key. Any other value is
// refused, because reading it as the default runs a mode nobody wrote (#p9s1131, #rfbha44).
var noUnusedVarsModes = []struct {
	key     string
	allowed []string
	value   func(*NoUnusedVarsOptions) string
}{
	{"vars", []string{"all", "local"}, func(options *NoUnusedVarsOptions) string { return options.Vars }},
	{"args", []string{"all", "after-used", "none"}, func(options *NoUnusedVarsOptions) string { return options.Args }},
	{"caughtErrors", []string{"all", "none"}, func(options *NoUnusedVarsOptions) string { return options.CaughtErrors }},
}

// UnmarshalJSON reads upstream's two forms: the object, decoded strictly, and the bare string
// `"all"` or `"local"`, which is shorthand for `{"vars": ...}` (`["error", "local"]`). A mode value
// outside upstream's enum is refused by key and value.
func (options *NoUnusedVarsOptions) UnmarshalJSON(raw []byte) error {
	var shorthand string
	if err := json.Unmarshal(raw, &shorthand); err == nil {
		if shorthand != "all" && shorthand != "local" {
			return fmt.Errorf("the string form is \"all\" or \"local\", and the config gives %q", shorthand)
		}
		*options = NoUnusedVarsOptions{Vars: shorthand}
		return nil
	}
	var fields noUnusedVarsOptionFields
	if err := rule.UnmarshalOptions(raw, &fields); err != nil {
		return err
	}
	decoded := NoUnusedVarsOptions(fields)
	for _, mode := range noUnusedVarsModes {
		value := mode.value(&decoded)
		if value != "" && !slices.Contains(mode.allowed, value) {
			return fmt.Errorf("%s is one of %q, and the config gives %q", mode.key, mode.allowed, value)
		}
	}
	*options = decoded
	return nil
}

// resolveNoUnusedVarsOptions applies the defaults to whatever the config supplied.
//
// A rule configured as bare `"error"` is handed nil, and the zero value of this struct is not the
// rule's default state: it would set `args` and `caughtErrors` to the empty string, which is two
// wrong answers. So the nil case is written explicitly rather than relying on the zero value, which
// is the shape a shipped rule got wrong on 3,407 files.
func resolveNoUnusedVarsOptions(options any) NoUnusedVarsOptions {
	settings, _ := rule.OptionsAs[NoUnusedVarsOptions](options)

	if settings.Vars == "" {
		settings.Vars = "all"
	}
	if settings.Args == "" {
		settings.Args = "after-used"
	}
	if settings.CaughtErrors == "" {
		settings.CaughtErrors = "all"
	}

	return settings
}

// DecodeNoUnusedVarsOptions parses this rule's options from a config file.
//
// Exported so fixtures can route through the real decoder rather than building the struct, which is
// what puts the default resolution above under test instead of bypassing it.
func DecodeNoUnusedVarsOptions(raw json.RawMessage) (NoUnusedVarsOptions, error) {
	var decoded NoUnusedVarsOptions
	if len(raw) == 0 {
		return resolveNoUnusedVarsOptions(nil), nil
	}
	if err := rule.UnmarshalOptions(raw, &decoded); err != nil {
		return resolveNoUnusedVarsOptions(nil), err
	}
	return resolveNoUnusedVarsOptions(decoded), nil
}
