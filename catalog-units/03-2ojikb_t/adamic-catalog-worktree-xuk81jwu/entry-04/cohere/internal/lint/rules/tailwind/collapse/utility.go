// The `@utility` evaluator: the component that answers the readings the descriptor table declines.
//
// Ported from `createCssUtility` and `resolveValueFunction` in `src/utilities.ts` at Tailwind
// 4.3.3, read at the pinned tag rather than from the minified bundle, so a disagreement with the
// engine is a finding rather than version skew.
//
// # What this closes, and why it could not be another table axis
//
// The descriptor table (descriptor.go) answers a functional root's `{order, count}` as a function
// of its value's inferred data type, and Phase 0 measured that model holding on 37,625 of 37,643
// registry classes. The 18 it declines are all repository-defined `@utility` blocks of the form
// `<animation-root>-translate-full`, and they are not a defect in the model, they are its boundary:
//
//	@utility slide-in-from-top-* {
//	    --enter-translate-y: calc(--value(integer) * var(--spacing) * -1);
//	    --enter-translate-y: calc(--value(--percentage-*, --percentage-translate-*) * -100%);
//	    --enter-translate-y: calc(--value(ratio) * -100%);
//	    --enter-translate-y: calc(--value(--translate-*, [percentage], [length]) * -1);
//	}
//
// Each declaration resolves independently, and one whose `--value()` fails is dropped rather than
// defaulted. So arity is the count of survivors, and a value can satisfy several resolution paths
// at once: `50` is both an `integer` and a `--percentage` key, so two declarations survive and the
// class reads `[]#2`; `4` is only an integer and `translate-full` only a `--percentage` key, so one
// survives each and both read `[]#1`. No row of a table keyed on a single data type can express a
// count that depends on how many *different* types a value simultaneously satisfies, which is why
// this is a loop with a resolve-or-drop test rather than another axis.
//
// # The drop is the mechanism, and defaulting is the plausible wrong answer
//
// The single rule this file exists to implement is that an unresolvable `--value()` removes its
// declaration from the body. A port that instead left the declaration in place with an empty or
// literal value would agree with the engine on every class where every declaration resolves, which
// is most of them, and would report `[]#4` for every one of the 18. That is the shape of failure
// this component is tested against: see TestUtilityDropIsNotDefault, which mutates exactly this
// rule and asserts the differential catches it.
//
// # Four post-conditions, none of which is the resolve-or-drop rule
//
// Upstream applies four gates after the walk, and they decide validity rather than arity. Each is
// carried because each rejects a class the loop alone would accept:
//
//  1. `--value()` must have been used, and at least one use must have resolved. A body where every
//     declaration dropped compiles to nothing, which is a null reading rather than `[]#0`.
//  2. A used `--modifier()` that resolved nothing, with a modifier present, is invalid.
//  3. `--value(ratio)` resolving alongside a resolved `--modifier()` is invalid, because
//     `foo-2/3` cannot be read as both a fraction and a value-plus-modifier.
//  4. A candidate carrying a modifier is invalid unless either the modifier resolved or a ratio
//     did, for the same reason read from the other side.
//
// And then the splice: when `--value(ratio)` resolved, every declaration that resolved a
// *non-ratio* `--value()` is removed. `slide-in-from-top-1/2` reads `[]#1` and not `[]#2` because
// of it, since `1` alone would otherwise satisfy the `integer` declaration. This is a removal after
// the fact rather than a condition during the walk, and it is ordered that way here because the
// engine cannot know a ratio resolved until it has walked past declarations that resolved without
// one.
//
// # What is deliberately not ported
//
// `createCssUtility`'s pre-processing walk normalizes `--value(…)` argument spellings
// (`--value(--spacing)` to `--value(--spacing-*)`, whitespace removal, escaped `\*`) and collects
// suggestion metadata. The normalization is ported, in normalizeValueFunctionArguments, because it
// changes which theme namespace an argument names and therefore changes readings. The suggestion
// half is not: it feeds editor autocomplete, has no reader here, and an untested surface that reads
// as supported is worse than an absent one.
//
// The `--spacing(…)` tracking in the same walk exists only to serve suggestions and is likewise
// absent. `console.warn` on an unsupported bare data type is absent too: it is diagnostic output,
// and the `continue` that follows it, which is what actually decides the reading, is ported.
package tailwind

import (
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// bareValueDataTypes is upstream's `BARE_VALUE_DATA_TYPES`, the only data types a `--value()` may
// name for a bare candidate value.
//
// The restriction is upstream's and it is deliberate rather than incidental: allowing
// `--value(color)` would make `text-#0088cc` a valid utility, which is syntax Tailwind does not
// want to exist. An argument naming any other type is skipped, so `--value(length)` never resolves
// a bare value and its declaration drops.
var bareValueDataTypes = map[DataType]bool{
	DataTypeNumber:     true,
	DataTypeInteger:    true,
	DataTypeRatio:      true,
	DataTypePercentage: true,
}

// UtilityDefinition is one `@utility` block, parsed and normalized, ready to evaluate.
//
// Name is the root without its `-*` suffix, so a `@utility slide-in-from-top-*` block has Name
// `slide-in-from-top`. Nodes is the block's body, already normalized: normalizeValueFunctionArguments
// has rewritten every `--value()` and `--modifier()` argument list into the spelling the resolver
// expects, so evaluation never has to re-derive it. That split matches upstream, where the
// normalization runs once at registration and the resolution runs per candidate.
type UtilityDefinition struct {
	// Name is the utility root, without the `-*` suffix.
	Name string
	// Nodes is the normalized body of the block.
	Nodes []*Node
}

// UtilityEvaluator evaluates `@utility` definitions against a theme.
//
// It holds the theme rather than taking it per call because the theme is what makes a repository's
// utilities repository-specific: `--value(--percentage-*)` resolves against this theme's
// `--percentage` namespace and no other, which is the whole reason the descriptor table is
// generated per repository and this evaluator is not a table at all.
type UtilityEvaluator struct {
	// Theme is the design system's resolved theme, consulted by every `--value(--namespace-*)`
	// argument.
	Theme *Theme
	// Definitions maps a utility root to its definition.
	Definitions map[string]*UtilityDefinition
}

// NewUtilityEvaluator builds an evaluator over a theme and a set of `@utility` blocks.
//
// The definitions are normalized here rather than at evaluation time, once per definition rather
// than once per candidate, matching upstream's split between `createCssUtility`'s registration walk
// and the per-candidate function it returns.
func NewUtilityEvaluator(theme *Theme, definitions []*UtilityDefinition) *UtilityEvaluator {
	byName := make(map[string]*UtilityDefinition, len(definitions))
	for _, definition := range definitions {
		normalizeUtilityDefinition(definition)
		byName[definition.Name] = definition
	}
	return &UtilityEvaluator{Theme: theme, Definitions: byName}
}

// Has reports whether a root is defined by an `@utility` block this evaluator holds.
func (evaluator *UtilityEvaluator) Has(root string) bool {
	_, found := evaluator.Definitions[root]
	return found
}

// Reading returns the `{order, count}` reading of a functional candidate against its `@utility`
// definition, and whether the candidate compiles at all.
//
// A false return is a candidate the engine rejects, which is a null reading rather than an empty
// one. The distinction decides sort position: an empty order sorts last, and a control in Phase 0
// that read a rejection as `[]#0` reported 365,174 spurious mismatches while the model was fine.
func (evaluator *UtilityEvaluator) Reading(candidate *ParsedCandidate) (Reading, bool) {
	nodes, ok := evaluator.Compile(candidate)
	if !ok {
		return Reading{}, false
	}
	sort := PropertySort(nodes)
	return Reading{Order: sort.Order, Count: sort.Count}, true
}

// Compile evaluates a candidate against its `@utility` definition and returns the surviving body.
//
// This is `createCssUtility`'s per-candidate function. The returned nodes are the utility's body
// before it is wrapped in its selector rule and before variants are applied, which is what
// PropertySort expects and what upstream passes to `getPropertySort`.
//
// The definition is not mutated. Upstream clones the at-rule per candidate for the same reason:
// resolution rewrites declaration values in place and removes declarations, and a definition
// consumed once would answer differently on the second class that used it.
func (evaluator *UtilityEvaluator) Compile(candidate *ParsedCandidate) ([]*Node, bool) {
	return evaluator.compile(candidate, utilityRemovalsAll)
}

// utilityRemovals selects which of the two post-walk removals `compile` applies.
//
// It exists so the tests can suppress one removal at a time and observe what changes, on the real
// evaluation path rather than on a reimplementation of it. A mutation test that reimplemented the
// walk would drift away from the code it claims to be mutating at the first refactor, and would then
// keep passing while proving nothing.
//
// Production always passes utilityRemovalsAll, which is what Compile is.
type utilityRemovals int

const (
	// utilityRemovalsAll applies both removals: the resolve-or-drop rule, and the ratio splice.
	utilityRemovalsAll utilityRemovals = iota
	// utilityRemovalsKeepDropped suppresses the resolve-or-drop rule, keeping declarations whose
	// `--value()` resolved nothing. This is the plausible wrong port: it defaults instead of
	// dropping. See TestUtilityDropIsNotDefault.
	utilityRemovalsKeepDropped
	// utilityRemovalsKeepNonRatio suppresses the ratio splice, keeping declarations that resolved a
	// non-ratio `--value()` alongside one that resolved a ratio. See
	// TestUtilityRatioSpliceIsLoadBearing.
	utilityRemovalsKeepNonRatio
)

// compileKeepingDroppedDeclarations is Compile with the resolve-or-drop rule suppressed.
//
// Test-only, and unexported for that reason. It is the mutant TestUtilityDropIsNotDefault asserts
// the differential catches.
func (evaluator *UtilityEvaluator) compileKeepingDroppedDeclarations(candidate *ParsedCandidate) ([]*Node, bool) {
	return evaluator.compile(candidate, utilityRemovalsKeepDropped)
}

// compileWithoutRatioSplice is Compile with the ratio splice suppressed. Test-only, as above.
func (evaluator *UtilityEvaluator) compileWithoutRatioSplice(candidate *ParsedCandidate) ([]*Node, bool) {
	return evaluator.compile(candidate, utilityRemovalsKeepNonRatio)
}

func (evaluator *UtilityEvaluator) compile(candidate *ParsedCandidate, removals utilityRemovals) ([]*Node, bool) {
	if candidate == nil || candidate.Kind != ParsedCandidateKindFunctional {
		return nil, false
	}
	definition, found := evaluator.Definitions[candidate.Root]
	if !found {
		return nil, false
	}

	state := &utilityEvaluation{
		evaluator: evaluator,
		value:     candidate.Value,
		modifier:  candidate.Modifier,
	}
	body := cloneNodes(definition.Nodes)
	state.walk(body)

	// Post-condition 1. A functional utility must use `--value()` and at least one use must have
	// resolved, or the candidate is not this utility at all. `--default(…)` inside `--value(…)` is
	// how a definition opts into the omitted-value case; without it, `foo` bare does not compile.
	if !state.usedValueFunction || !state.resolvedValueFunction {
		return nil, false
	}
	// Post-condition 2. A `--modifier()` that was reached and never resolved, with a modifier
	// actually present, is a modifier the utility could not consume.
	if state.usedModifierFunction && !state.resolvedModifierFunction && candidate.Modifier != nil {
		return nil, false
	}
	// Post-condition 3. A ratio and a modifier cannot both be right: `foo-2/3` is either a fraction
	// or a value with a modifier.
	if state.resolvedRatioValue && state.resolvedModifierFunction {
		return nil, false
	}
	// Post-condition 4. The same rule read from the candidate's side: a written modifier has to be
	// consumed by something.
	if candidate.Modifier != nil && !state.resolvedRatioValue && !state.resolvedModifierFunction {
		return nil, false
	}

	// The two removals, in this order because the second depends on the whole walk having finished.
	//
	// First the drops: a declaration whose `--value()` did not resolve is not in the compiled body
	// at all. This is the mechanism the component exists for, and it is what makes arity the count
	// of survivors rather than the count of declarations written.
	//
	// Then the splice: a resolved ratio invalidates every declaration that resolved a non-ratio
	// `--value()`, because `slide-in-from-top-1/2` would otherwise also satisfy the `integer`
	// declaration through the `1` and read `[]#2`.
	var removed map[*Node]bool
	if removals != utilityRemovalsKeepDropped {
		removed = state.droppedDeclarations
	}
	if state.resolvedRatioValue && removals != utilityRemovalsKeepNonRatio {
		removed = unionNodeSets(removed, state.nonRatioDeclarations)
	}
	body = removeNodes(body, removed)

	return body, true
}

// utilityEvaluation is the mutable state of one candidate's evaluation.
//
// Five booleans and a set, matching upstream's locals. They are a struct rather than closure
// captures so the walk can be a method and read as the same sequence of branches upstream is.
type utilityEvaluation struct {
	evaluator *UtilityEvaluator
	value     *ParsedValue
	modifier  *ParsedModifier

	// usedValueFunction records that a `--value()` was reached at all, which is what separates a
	// functional utility from a definition that never consumes its value.
	usedValueFunction bool
	// resolvedValueFunction records that some `--value()` resolved. One is enough: a body of four
	// declarations where three drop is a valid utility reading `[]#1`.
	resolvedValueFunction bool
	// usedModifierFunction and resolvedModifierFunction are the same pair for `--modifier()`.
	usedModifierFunction     bool
	resolvedModifierFunction bool
	// resolvedRatioValue records that a `--value(ratio)` resolved, which drives both post-condition
	// 3 and the splice.
	resolvedRatioValue bool
	// nonRatioDeclarations is every declaration that resolved a `--value()` without a ratio, which
	// is exactly the set the splice removes when a ratio did resolve. Identity, not equality: two
	// declarations in one body can be textually identical and only one of them resolved.
	nonRatioDeclarations map[*Node]bool
	// droppedDeclarations is every declaration whose `--value()` failed to resolve. This is the
	// mechanism the whole component exists for, and it is a set keyed on node identity rather than
	// a flag on Node: the body is cloned per candidate, so identity is unambiguous within one
	// evaluation, and ast.go stays the CSS tree the parser produces rather than gaining a field
	// only this evaluator writes.
	droppedDeclarations map[*Node]bool
}

// walk visits every declaration under a rule or at-rule and resolves its value functions.
//
// Depth-first through containers, matching upstream's `walk([atRule], …)` with its
// `parent?.kind !== 'rule' && parent?.kind !== 'at-rule'` guard. That guard is why a declaration at
// the top level of the body is still visited: upstream walks the cloned at-rule itself, so the
// body's own declarations have the at-rule as their parent. Here the body is the at-rule's Nodes,
// so the top level is visited directly and nested rules recurse.
//
// Descent is by container kind rather than by upstream's parent guard, and the two differ on
// KindContext and KindAtRoot: upstream would skip a declaration whose parent is either, and this
// resolves it. Neither kind can occur here, because ParseCSS produces only rules, at-rules,
// declarations and comments, and the two wrappers are built by the compiler rather than parsed. They
// are listed so the switch is exhaustive over container kinds rather than silently dropping a
// subtree if a future caller hands one in, and the divergence is stated because it is a divergence:
// if such a caller ever appears, this is the line to revisit.
func (state *utilityEvaluation) walk(nodes []*Node) {
	for _, node := range nodes {
		switch node.Kind {
		case KindDeclaration:
			state.resolveDeclaration(node)
		case KindRule, KindAtRule, KindContext, KindAtRoot:
			state.walk(node.Nodes)
		}
	}
}

// resolveDeclaration resolves every `--value()` and `--modifier()` in one declaration, and marks it
// for removal when any of them fails.
//
// The removal is marked rather than applied, because upstream's walk returns
// `WalkAction.ReplaceSkip([])` and the node is spliced out of its parent by the walker. Here the
// node is flagged and removeNodes does the splice, which keeps the traversal free of the
// index-shifting that in-place removal during iteration invites.
func (state *utilityEvaluation) resolveDeclaration(node *Node) {
	if !node.ValuePresent || node.Value == "" {
		// Upstream's `if (!node.value) return`, which skips a declaration with no value and,
		// because the empty string is falsy in JavaScript, one whose value is empty. Both are left
		// in the body untouched, so `--tw-foo:;` still counts toward arity.
		return
	}

	valueAst := ParseValue(node.Value)
	dropped := state.resolveValueFunctions(valueAst, node)
	if dropped {
		if state.droppedDeclarations == nil {
			state.droppedDeclarations = map[*Node]bool{}
		}
		state.droppedDeclarations[node] = true
		return
	}
	node.Value = ValueToCss(valueAst)
}

// resolveValueFunctions walks a parsed declaration value, replacing each `--value()` and
// `--modifier()` call with what it resolved to, and reports whether the declaration must be
// dropped.
//
// It returns on the first failure rather than continuing, matching upstream's `WalkAction.Stop`.
// The distinction is observable: a declaration holding two `--value()` calls where the first fails
// leaves the second unresolved, and the declaration is dropped either way, so the only difference
// is which side effects were recorded. `usedValueFunction` is one of those side effects and it
// gates post-condition 1.
//
// Replacement is `ReplaceSkip`: the substituted nodes are not themselves rewalked, so a theme value
// that happens to contain the text `--value(` is inserted literally rather than resolved again.
func (state *utilityEvaluation) resolveValueFunctions(nodes []ValueNode, declaration *Node) bool {
	for index := range nodes {
		node := &nodes[index]
		if node.Kind != ValueNodeKindFunction {
			continue
		}

		switch node.Value {
		case "--value":
			state.usedValueFunction = true
			// A nil value is upstream's `null`, which reaches the `--default(…)` path rather than
			// failing outright.
			resolved, isRatio, ok := state.resolveValueFunction(state.value, node)
			if !ok {
				return true
			}
			state.resolvedValueFunction = true
			if isRatio {
				state.resolvedRatioValue = true
			} else {
				if state.nonRatioDeclarations == nil {
					state.nonRatioDeclarations = map[*Node]bool{}
				}
				state.nonRatioDeclarations[declaration] = true
			}
			replaceValueNode(node, resolved)

		case "--modifier":
			state.usedModifierFunction = true
			resolved, _, ok := state.resolveValueFunction(state.modifierAsValue(), node)
			if !ok {
				return true
			}
			state.resolvedModifierFunction = true
			replaceValueNode(node, resolved)

		default:
			// Any other function is descended into, because `calc(--value(integer) * -1)` holds its
			// `--value()` one level down and every declaration in the corpus is written that way.
			if state.resolveValueFunctions(node.Nodes, declaration) {
				return true
			}
		}
	}
	return false
}

// modifierAsValue presents the candidate's modifier as a ParsedValue.
//
// Upstream passes the modifier to the same `resolveValueFunction` the value goes through, and the
// two types differ only in that a modifier has no `fraction`. Converting rather than writing a
// second resolver keeps one implementation of the argument grammar, which is what makes
// `--modifier(--percentage-*)` and `--value(--percentage-*)` provably the same lookup.
func (state *utilityEvaluation) modifierAsValue() *ParsedValue {
	if state.modifier == nil {
		return nil
	}
	kind := ParsedValueKindNamed
	if state.modifier.Kind == ParsedModifierKindArbitrary {
		kind = ParsedValueKindArbitrary
	}
	// No Fraction and no DataType: a modifier is never half of a fraction, and `/[0.5]` carries no
	// typehint because the parser does not read one there.
	return &ParsedValue{Kind: kind, Value: state.modifier.Value}
}

// resolveValueFunction is upstream's `resolveValueFunction`: it tries each argument of a
// `--value()` or `--modifier()` in order and returns the first that resolves.
//
// The three returns are the nodes to substitute, whether the resolution was a ratio, and whether
// anything resolved at all. The ratio flag is separate from the nodes because it drives the splice
// and post-condition 3, and a caller that inferred it from the substituted text would be reading
// the answer rather than being told it.
//
// Order is the argument order as written, and it is load-bearing: `--value(--percentage-*,
// --percentage-translate-*)` tries `--percentage` first, so `translate-full` resolves through
// `--percentage` and never reaches the second argument. Which one wins does not change this
// declaration's survival, but the general rule does decide which theme value is substituted.
func (state *utilityEvaluation) resolveValueFunction(value *ParsedValue, function *ValueNode) ([]ValueNode, bool, bool) {
	// No value at all: only `--default(…)` can answer, and it is the utility's own opt-in to the
	// bare form. A definition without one does not compile bare, which is why `slide-in-from-top`
	// is a separate static `@utility` block rather than a fallback of the functional one.
	if value == nil {
		for _, argument := range function.Nodes {
			if argument.Kind == ValueNodeKindFunction && argument.Value == "--default" {
				return argument.Nodes, false, true
			}
		}
		return nil, false, false
	}

	for _, argument := range function.Nodes {
		switch {
		// A literal, e.g. `--modifier('closest-side')`. Quoted, and the quotes must match at both
		// ends, so `'closest-side` is not a literal and falls through to the bare-value branch,
		// where it is not a known data type and is skipped.
		case value.Kind == ParsedValueKindNamed && argument.Kind == ValueNodeKindWord &&
			isQuotedLiteral(argument.Value) && argument.Value[1:len(argument.Value)-1] == value.Value:
			return ParseValue(value.Value), false, true

		// A theme namespace, e.g. `--value(--percentage-*)`.
		case value.Kind == ParsedValueKindNamed && argument.Kind == ValueNodeKindWord &&
			strings.HasPrefix(argument.Value, "--"):
			resolved, ok := state.resolveThemeArgument(argument.Value, value.Value)
			if ok {
				return ParseValue(resolved), false, true
			}

		// A bare value with a data type, e.g. `--value(integer)`.
		case value.Kind == ParsedValueKindNamed && argument.Kind == ValueNodeKindWord:
			resolved, isRatio, ok := state.resolveBareArgument(argument.Value, value)
			if ok {
				return resolved, isRatio, true
			}

		// An arbitrary value, e.g. `--value([length])`.
		case value.Kind == ParsedValueKindArbitrary && argument.Kind == ValueNodeKindWord &&
			isBracketed(argument.Value):
			resolved, ok := state.resolveArbitraryArgument(argument.Value, value)
			if ok {
				return resolved, false, true
			}
		}
	}

	return nil, false, false
}

// resolveThemeArgument resolves a `--value(--namespace-*)` or `--value(--text-*--line-height)`
// argument against the theme.
//
// The two forms are genuinely different lookups. A trailing `-*` is a namespace: the suffix is
// stripped and the candidate value is resolved under it. An embedded `-*` is the nested form, where
// the text before it is the namespace and the text after is a sub-variable suffix, so
// `--value(--text-*--line-height)` resolves `--text-sm` and then returns its `--line-height`
// sub-variable rather than the size.
//
// An argument with no `-*` at all resolves nothing and is skipped, because normalization appends
// `-*` to every bare `--foo` argument, so reaching here without one means the argument held a `(`
// and was left alone.
func (state *utilityEvaluation) resolveThemeArgument(argument, candidateValue string) (string, bool) {
	if strings.HasSuffix(argument, "-*") {
		namespace := strings.TrimSuffix(argument, "-*")
		// `resolve`, not `resolveValue`: upstream calls `designSystem.theme.resolve`, which returns
		// a `var()` reference for a normal entry and the literal only for an inlined one. The
		// substituted text is what the declaration's value becomes, and while the reading counts
		// declarations rather than reading their text, substituting the literal here would diverge
		// from the engine on any consumer that does read it.
		return state.evaluator.Theme.Resolve(candidateValue, true, []string{namespace}, ThemeOptionNone)
	}

	// The nested form. `--text-*--line-height` splits into the namespace `--text` and the nested
	// suffixes after it; a split yielding one part has no `-*` and cannot be nested.
	parts := strings.Split(argument, "-*")
	if len(parts) <= 1 {
		return "", false
	}
	namespace := parts[0]
	nested := parts[1:]
	_, extra, ok := state.evaluator.Theme.ResolveWith(candidateValue, []string{namespace}, nested)
	if !ok {
		return "", false
	}
	// Upstream pops the last nested key and reads it out of the options map, so
	// `--text-*--line-height` returns the line height rather than the size.
	resolved, present := extra[nested[len(nested)-1]]
	if !present {
		return "", false
	}
	return resolved, true
}

// resolveBareArgument resolves a `--value(integer)`, `--value(number)`, `--value(ratio)` or
// `--value(percentage)` argument against a bare candidate value.
//
// Four guards past the type check, and each rejects a value the type alone accepts:
//
//   - A `ratio` argument reads the candidate's Fraction rather than its Value, because `foo-1/2`
//     parses as value `1` with modifier `2` and only Fraction holds the whole `1/2`. A candidate
//     with no fraction cannot satisfy a ratio argument.
//   - A resolved ratio must be `<positive-integer>/<positive-integer>`, so `1.5/2` is rejected
//     after inferring as a ratio.
//   - A `number` must be a valid spacing multiplier, a non-negative multiple of 0.25 written
//     canonically, so `1.3` is rejected and `1.25` is not.
//   - A `percentage` must be an integer percentage, so `12.5%` is rejected.
func (state *utilityEvaluation) resolveBareArgument(argument string, value *ParsedValue) ([]ValueNode, bool, bool) {
	dataType := DataType(argument)
	if !bareValueDataTypes[dataType] {
		return nil, false, false
	}

	resolved := value.Value
	if dataType == DataTypeRatio {
		resolved = value.Fraction
	}
	if resolved == "" {
		return nil, false, false
	}

	inferred := InferDataType(resolved, []DataType{dataType})
	if inferred == "" {
		return nil, false, false
	}

	switch inferred {
	case DataTypeRatio:
		parts := segment(resolved, '/')
		if len(parts) != 2 || !IsPositiveInteger(parts[0]) || !IsPositiveInteger(parts[1]) {
			return nil, false, false
		}
		// Reprinted with spaces around the slash, matching upstream, so the substituted text is
		// `1 / 2` rather than `1/2`. A bare `1/2` inside a `calc()` would be a division the browser
		// reads differently from the ratio the engine intends.
		return ParseValue(text.TrimWhitespace(parts[0]) + " / " + text.TrimWhitespace(parts[1])), true, true
	case DataTypeNumber:
		if !isValidSpacingMultiplier(resolved) {
			return nil, false, false
		}
	case DataTypePercentage:
		if !IsPositiveInteger(strings.TrimSuffix(resolved, "%")) {
			return nil, false, false
		}
	}

	return ParseValue(resolved), false, true
}

// resolveArbitraryArgument resolves a `--value([length])` or `--value([*])` argument against a
// bracketed candidate value.
//
// `[*]` accepts anything. Otherwise the candidate's own typehint decides: `bg-[color:var(--x)]`
// carries `color`, and an argument naming a different type is skipped rather than inferred, because
// the author's annotation is a claim the engine trusts. With no typehint the value is inferred
// against the single named type.
func (state *utilityEvaluation) resolveArbitraryArgument(argument string, value *ParsedValue) ([]ValueNode, bool) {
	dataType := argument[1 : len(argument)-1]

	if dataType == "*" {
		return ParseValue(value.Value), true
	}
	if value.DataType != "" {
		if value.DataType != dataType {
			return nil, false
		}
		return ParseValue(value.Value), true
	}
	if InferDataType(value.Value, []DataType{DataType(dataType)}) != "" {
		return ParseValue(value.Value), true
	}
	return nil, false
}

// isQuotedLiteral reports whether a word is a quoted literal: at least two characters, opening with
// a quote, and closing with the same one.
//
// The length guard matters. Upstream's test compares `value[0]` against
// `value[value.length - 1]`, which for the one-character string `"` compares the quote against
// itself and passes, and then slices it to the empty string. A candidate value of `""` would
// therefore match a `--value(")` argument. The guard is written explicitly rather than left to
// coincide, because the coincidence is the kind that changes under a refactor.
func isQuotedLiteral(word string) bool {
	if len(word) < 2 {
		return false
	}
	quote := word[0]
	if quote != '\'' && quote != '"' {
		return false
	}
	return word[len(word)-1] == quote
}

// isBracketed reports whether a word is `[...]`.
func isBracketed(word string) bool {
	return len(word) >= 2 && word[0] == '[' && word[len(word)-1] == ']'
}

// isValidSpacingMultiplier is upstream's `isValidSpacingMultiplier`: a non-negative multiple of
// 0.25 that reprints identically.
//
// The reprint test is what rejects `1.50` and `+1`, which are multiples of 0.25 and are not how
// JavaScript prints the number they parse to. `0.25` reaches this only through a `--value(number)`
// argument, so the whole guard exists to make `foo-1.3` drop its declaration while `foo-1.25` keeps
// it.
func isValidSpacingMultiplier(value string) bool {
	return isMultipleOf(value, 0.25)
}

// isValidOpacityValue is upstream's `isValidOpacityValue`: the same predicate as the spacing
// multiplier, against the same divisor.
//
// Identical to isValidSpacingMultiplier today, and kept separate rather than aliased because they
// answer different questions and upstream has changed one without the other before. Folding them
// would make a future divergence a silent behaviour change here rather than a compile error.
func isValidOpacityValue(value string) bool {
	return isMultipleOf(value, 0.25)
}

// isMultipleOf is upstream's `isMultipleOf`, the shared body of `isValidSpacingMultiplier` and
// `isValidOpacityValue`.
//
// `String(Number(value)) === String(value)` is reproduced by formatting the parsed float with Go's
// shortest round-trip formatting, which picks the same digits ECMAScript's Number::toString does
// for every value in range here. The modulo is done in float64 exactly as upstream does it, so a
// value where floating point makes `num % 0.25` non-zero is rejected in both.
func isMultipleOf(value string, divisor float64) bool {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	if parsed < 0 {
		return false
	}
	// Go's math.Mod and JavaScript's `%` agree: both are the IEEE 754 remainder truncated toward
	// zero, so the same inputs produce the same non-zero residues.
	if remainder := modFloat(parsed, divisor); remainder != 0 {
		return false
	}
	return formatJavaScriptNumber(parsed) == value
}
