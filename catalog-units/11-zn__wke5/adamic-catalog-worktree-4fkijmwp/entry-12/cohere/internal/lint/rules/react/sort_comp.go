package react

import (
	"fmt"
	"math"
	"regexp"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// SortCompOptions configures the expected order of a component's members.
//
// Upstream's schema is a single object with `order` (an array of strings) and `groups` (an object
// whose every value is an array of strings). Our config layer unwraps the severity tuple before
// dispatch, so the decoder receives that object directly rather than upstream's one-element array.
//
// Both fields are pointers to slices/maps rather than plain ones, because upstream distinguishes
// ABSENT from EMPTY on `order`: `userConfig.order || defaultConfig.order` falls back only when the
// field is missing or null, so an explicitly empty `order: []` is a real configuration meaning "no
// group matches anything". A plain `[]string` cannot tell those apart and would silently restore
// the default for a user who asked for nothing.
type SortCompOptions struct {
	// Order is the sequence of group names and member names to enforce. An entry naming a key of
	// Groups expands to that group's contents; anything else is matched directly, as a literal
	// member name, as one of the built-in shape predicates, or as a `/regex/flags` pattern.
	Order *[]string `json:"order"`

	// Groups maps a group name to the member names it contains. Upstream merges this OVER the
	// default groups rather than replacing them, so a user who defines `customGroup` still has
	// `lifecycle` available, and a user who redefines `lifecycle` replaces only that one.
	Groups *map[string][]string `json:"groups"`
}

// sortCompDefaultOrder is upstream's `defaultConfig.order`, copied in order.
var sortCompDefaultOrder = []string{
	"static-methods",
	"lifecycle",
	"everything-else",
	"render",
}

// sortCompDefaultLifecycle is upstream's `defaultConfig.groups.lifecycle`, copied in order.
//
// The order inside this list is load-bearing rather than decorative: it is what says `propTypes`
// belongs before `constructor` and `componentDidMount` before `componentWillUnmount`. Sorting it
// or deduplicating it would change the rule's verdict on real components.
var sortCompDefaultLifecycle = []string{
	"displayName",
	"propTypes",
	"contextTypes",
	"childContextTypes",
	"mixins",
	"statics",
	"defaultProps",
	"constructor",
	"getDefaultProps",
	"state",
	"getInitialState",
	"getChildContext",
	"getDerivedStateFromProps",
	"componentWillMount",
	"UNSAFE_componentWillMount",
	"componentDidMount",
	"componentWillReceiveProps",
	"UNSAFE_componentWillReceiveProps",
	"shouldComponentUpdate",
	"componentWillUpdate",
	"UNSAFE_componentWillUpdate",
	"getSnapshotBeforeUpdate",
	"componentDidUpdate",
	"componentDidCatch",
	"componentWillUnmount",
}

// sortCompKnownLifecycleNames is upstream's inline `arrayIncludes([...])` list in
// `getRefPropIndexes`, which is the default lifecycle list PLUS `render`.
//
// It is a separate list from `sortCompDefaultLifecycle` and the difference matters. This one
// decides whether an order entry is treated as a LITERAL MEMBER NAME rather than as a regex or a
// shape predicate; the other decides what the `lifecycle` group expands to. `render` is in this
// list and not in that one, because `render` is its own top-level entry in the default order.
//
// Membership here is checked before the regex branch, so a user cannot write an order entry named
// `render` and have it interpreted as a pattern.
var sortCompKnownLifecycleNames = func() map[string]bool {
	names := map[string]bool{"render": true}
	for _, name := range sortCompDefaultLifecycle {
		names[name] = true
	}
	return names
}()

// sortCompShapePredicates are the order entries that select members by SHAPE rather than by name.
//
// Listed here only so the set is greppable and so `sortCompRegexPattern` is never consulted for
// one of them; the dispatch itself is the switch in `sortCompGroupMatches`, which follows
// upstream's if/else chain in order.
var sortCompShapePredicates = map[string]bool{
	"getters":            true,
	"setters":            true,
	"type-annotations":   true,
	"static-variables":   true,
	"static-methods":     true,
	"instance-variables": true,
	"instance-methods":   true,
}

// sortCompRegexPattern is upstream's `regExpRegExp`, matching a `/body/flags` order entry.
//
// Upstream writes `/\/(.*)\/([gimsuy]*)/` and applies it with `String.prototype.match`, which is
// UNANCHORED: an entry like `foo/on.*/bar` matches, capturing `on.*`. That is reproduced with
// `FindStringSubmatch` rather than tightened to a full-string match, because tightening it would
// make a rule that silently disagrees with upstream on an input no corpus case writes. Go's regexp
// is RE2 and rejects the backtracking constructs Go does not support, which is why a pattern that
// fails to compile is treated as "matches nothing" rather than panicking; upstream's `new RegExp`
// would throw there and take the whole lint run down.
var sortCompRegexPattern = regexp.MustCompile(`/(.*)/([gimsuy]*)`)

// DefaultSortCompOptions is the unconfigured answer.
//
// Upstream's `getMethodsOrder(undefined)` produces exactly this: the default order, with
// `lifecycle` expanded and every other entry passed through.
func DefaultSortCompOptions() SortCompOptions {
	return SortCompOptions{}
}

// DecodeSortCompOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so a bare `"error"` configuration, which arrives
// as empty input, resolves to the documented default instead of erroring. The generic helper would
// also flatten the absent/empty distinction the pointer fields exist to preserve.
func DecodeSortCompOptions(raw []byte) (any, error) {
	options := DefaultSortCompOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// sortCompUnsortedProps builds the per-finding message naming both members and the direction.
//
// The Id is fixed and only the Description moves, so `ExpectFindings` can still count these while
// the rendered text names the two members the way upstream's `{{propA}} should be placed
// {{position}} {{propB}}` interpolation does. The first sentence is upstream's text verbatim,
// because it is the part a reader matches against upstream's own output; the rest is this
// codebase's house style of saying why rather than restating the rule name.
func sortCompUnsortedProps(propertyA string, position string, propertyB string) rule.Message {
	return rule.Message{
		Id: "unsortedProps",
		Description: fmt.Sprintf(
			"%s should be placed %s %s. A component read top to bottom tells a story in a fixed "+
				"order: what it is, what it takes, how it sets itself up, how it responds to the "+
				"outside, and last what it draws. When members arrive out of that order the reader "+
				"has to hold the whole class in their head to find the piece they came for, and "+
				"`render` in particular stops being the place the file ends. Move it.",
			propertyA, position, propertyB,
		),
	}
}

// SortComp enforces a consistent order for a React component's members.
//
//	valid:   createReactClass({displayName: 'H', render: function() { return <div/>; }})
//	valid:   class H extends React.Component { constructor() {} render() { return <div/>; } }
//	valid:   class H extends React.Component { static m() {} render() { return <div/>; } }
//	invalid: createReactClass({render: function() { return <div/>; }, displayName: 'H'})
//	invalid: class H extends React.Component { render() { return <div/>; } static displayName = 'H'; }
//
// Ported from `react/sort-comp` in `eslint-plugin-react`, read from the clone at
// `lib/rules/sort-comp.js` (7.37.5). One options object (`order` and `groups`), one message id, no
// fixer: `meta` carries docs, messages and schema, and the string `fixable` does not appear in the
// file.
//
// The whole 53-case corpus was extracted from upstream's own tester and replayed against the
// installed 7.37.5 build through the ESLint Linter API before any code was written. It agreed with
// the corpus on all 53, counts and message text alike.
//
// # The oracle needed a TypeScript parser, and that is a fact about the harness rather than the rule
//
// 24 of the 53 cases carry upstream's `class fields` or `types` feature flag, and under the plain
// JSX parser every one of them comes back as a fatal parse error, which reads exactly like a clean
// verdict. Driving the oracle with `@typescript-eslint/parser` instead moved it from 29 of 53 to
// 53 of 53 with zero parse failures. Our parser reads TypeScript unconditionally, so those 24 have
// real verdicts here and had to be measured rather than inherited.
//
// # What upstream's algorithm actually is, and why it is not a sort
//
// The name says sort and the implementation is a pairwise vote. For each ORDERED PAIR of members
// upstream asks whether their relative position is defensible, and records a violation against the
// FIRST member of any pair that is not. So a single misplaced member in a class of six can
// accumulate five votes against it, and every other member accumulates one.
//
// The votes are then reconciled twice over, and both reconciliations are load-bearing:
//
//	score    how many pairs a member lost. `storeError` increments it on every failing pair.
//	closest  the losing partner with the smallest index distance, which is what the message names.
//
// `dedupeErrors` then walks the recorded errors and, for each one, looks up the error recorded
// against its own closest partner. Whichever of the two has the higher score survives and the other
// is deleted. That is what turns "render and displayName are in the wrong order" from two findings
// into one, pointed at whichever member is more wrong.
//
// Reproducing this as an actual sort would be simpler and would disagree with upstream on any input
// where three or more members conflict, because a sort has no notion of one member being more at
// fault than another.
//
// # Map iteration order is the one place Go could disagree with itself
//
// Upstream's `errors` is a JavaScript object keyed by index, and both `dedupeErrors` and
// `reportErrors` walk it with `Object.entries`, which for integer-like keys iterates in ASCENDING
// NUMERIC ORDER. Go's map iteration is randomised, and `dedupeErrors` MUTATES the collection it
// walks, so an unordered walk makes the surviving finding depend on which entry was visited first.
//
// That is not a theoretical hazard: with two conflicting members of equal score, the two orders
// delete different entries, so the rule would report a different member run to run. A flaky rule
// reads as green most of the time and blames something unrelated on the run where it does not, so
// the determinism is put in the RULE rather than trusted to a fixture: `sortCompSortedIndexes`
// walks the keys in ascending numeric order at both sites, matching `Object.entries`.
//
// # Two shapes upstream's corpus cannot express, and how they are decided here
//
// `type-annotations` selects a member that carries a type annotation and NO value, which upstream
// spells `!!node.typeAnnotation && node.value === null`. In our AST a property declaration's `Type`
// is a sibling of the name rather than part of it, and `Initializer` is what upstream calls
// `value`, so the predicate is `declaration.Type != nil && declaration.Initializer == nil`. The
// corpus pins it: `props: { text: string };` matches and `state: Object = {};` does not, and the
// case that separates them is invalid[6], where reversing the second half of the test moves the
// finding.
//
// `getters` and `setters` read `node.kind === 'get'` / `'set'`, which are distinct node kinds here
// rather than a discriminant field, so they are matched on `ast.KindGetAccessor` and
// `ast.KindSetAccessor`. An accessor in an OBJECT LITERAL is the same two kinds, which is why the
// factory-component arm reaches them too.
var SortComp = rule.Rule{
	Name: "react/sort-comp",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		resolved, isSortCompOptions := rule.OptionsAs[SortCompOptions](options)
		if !isSortCompOptions {
			// A rule configured as a bare `"error"` is handed nil options rather than the decoded
			// default, and `options.(T)` on nil yields the zero value silently. Falling back
			// explicitly is what keeps that path from being a rule that registers on every file and
			// judges with an empty order, which would report nothing and look exactly like a clean
			// tree.
			resolved = DefaultSortCompOptions()
		}
		methodsOrder := sortCompMethodsOrder(resolved)

		return rule.Listeners{
			// The whole judgment happens once per file rather than per node, because upstream's is
			// a `Program:exit` handler over the component list it accumulated. `KindSourceFile`
			// fires before its children here, so the walk is done by hand inside it.
			ast.KindSourceFile: func(node *ast.Node) {
				for _, component := range collectDetectedComponents(ctx, node) {
					members := sortCompComponentMembers(component.node)
					if len(members) < 2 {
						// Fewer than two members cannot form a pair, so no vote can be cast. Not a
						// correctness guard, since the loops below would simply do nothing; it is
						// here because the common case in this tree is a one-member component and
						// the allocation below is not worth paying for it.
						continue
					}
					sortCompCheckOrder(ctx, methodsOrder, members)
				}
			},
		}
	},
}

// sortCompMember is one member of a component, with every shape fact the order entries can ask about.
//
// The facts are computed once per member rather than per comparison, because the pairwise loop is
// quadratic and upstream computes them once too (its `propertiesInfos.map`).
type sortCompMember struct {
	// node is the member itself, which is what a finding anchors on.
	node *ast.Node

	// name is upstream's `getPropertyName`, with the two accessor kinds replaced by the literal
	// strings upstream substitutes. Those strings reach the MESSAGE as well as the matching, which
	// is why they are stored rather than derived at report time.
	name string

	getter           bool
	setter           bool
	staticVariable   bool
	staticMethod     bool
	instanceVariable bool
	instanceMethod   bool
	typeAnnotation   bool
}

// sortCompMethodsOrder reproduces upstream's `getMethodsOrder`.
//
// Groups are merged over the defaults rather than replacing them, and each order entry is either
// expanded (if it names a group) or passed through. Note that the expansion is one level deep:
// upstream concatenates the group's contents without re-expanding them, so a group naming another
// group produces that group's NAME as a literal member name to match, not its contents.
func sortCompMethodsOrder(options SortCompOptions) []string {
	groups := map[string][]string{"lifecycle": sortCompDefaultLifecycle}
	if options.Groups != nil {
		for name, members := range *options.Groups {
			groups[name] = members
		}
	}

	order := sortCompDefaultOrder
	if options.Order != nil {
		order = *options.Order
	}

	resolved := []string{}
	for _, entry := range order {
		if members, isGroup := groups[entry]; isGroup {
			resolved = append(resolved, members...)
			continue
		}
		resolved = append(resolved, entry)
	}
	return resolved
}

// sortCompComponentMembers reproduces upstream's `astUtil.getComponentProperties`.
//
// A class contributes its body, an object literal contributes its properties, and everything else
// contributes nothing. That last arm is why a function component is silently skipped rather than
// judged: `collectDetectedComponents` finds arrow and function components too, and upstream's
// `default: return []` is what declines them. Reproduced rather than filtered upstream of here, so
// the decline stays visible at the line that makes it.
func sortCompComponentMembers(component *ast.Node) []sortCompMember {
	if component == nil {
		return nil
	}

	var nodes []*ast.Node
	switch component.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		declaration := component.ClassLikeData()
		if declaration == nil || declaration.Members == nil {
			return nil
		}
		nodes = declaration.Members.Nodes
	case ast.KindObjectLiteralExpression:
		literal := component.AsObjectLiteralExpression()
		if literal == nil || literal.Properties == nil {
			return nil
		}
		nodes = literal.Properties.Nodes
	default:
		return nil
	}

	members := make([]sortCompMember, 0, len(nodes))
	for _, node := range nodes {
		members = append(members, sortCompMemberFor(node))
	}
	return members
}

// sortCompMemberFor computes one member's name and shape facts.
//
// The four static/instance variable-or-method flags follow upstream's definitions exactly, and the
// pairing is easy to get subtly wrong, so it is worth stating what each one means here:
//
//	staticVariable    static, a property declaration, and NOT initialized to a function
//	staticMethod      static, and either a method OR a property initialized to a function
//	instanceVariable  not static, a property declaration, and NOT initialized to a function
//	instanceMethod    not static, a property declaration, and initialized to a function
//
// The asymmetry in the third column is upstream's, not a transcription slip: `staticMethod` accepts
// `MethodDefinition` while `instanceMethod` does not, so a plain `foo() {}` in a class is neither a
// static method nor an instance method by these flags and falls through to `everything-else`. A
// `static foo() {}` IS matched by `static-methods`. Verified against the installed build rather
// than read: a class whose only members are `render` and a non-static `onClick() {}` under an order
// naming `instance-methods` does not place `onClick` in that group.
func sortCompMemberFor(node *ast.Node) sortCompMember {
	member := sortCompMember{node: node}

	switch node.Kind {
	case ast.KindGetAccessor:
		member.getter = true
		member.name = "getter functions"
	case ast.KindSetAccessor:
		member.setter = true
		member.name = "setter functions"
	default:
		member.name = sortCompMemberName(node)
	}

	isStatic := ast.HasStaticModifier(node)

	// A property declaration is upstream's `ClassProperty` / `PropertyDefinition`, and an object
	// literal's `PropertyAssignment` is upstream's `Property`. Both carry a value that may or may
	// not be a function, which is the distinction every flag below turns on.
	initializer, isPropertyLike := sortCompPropertyInitializer(node)
	initializedToFunction := isPropertyLike && initializer != nil && isFunctionLikeExpression(initializer)

	// `isMethodDeclaration` names only the method kind, and NOT the constructor kind, which reads as
	// an omission and is not. A constructor is ESTree's `MethodDefinition` too, so including it here
	// looks right; but this variable is read on exactly one line below, guarded by `isStatic`, and a
	// constructor can never be static. Adding the constructor kind is therefore unreachable, which a
	// mutation confirmed: including it and excluding it both survive the whole suite.
	//
	// The constructor is still handled, in the two places where it is reachable. Its NAME comes from
	// `sortCompMemberName`, which special-cases the kind because `Name()` is nil for it, and it is
	// kept out of the variable flags below by `isPropertyLike`, which a constructor is not.
	isMethodDeclaration := node.Kind == ast.KindMethodDeclaration

	// Upstream writes the method arms as `(isPropertyLike || isMethodDeclaration) && node.value &&
	// isFunctionLikeExpression(node.value)`, and transcribing that shape here produces a DEAD
	// clause, because our `initializedToFunction` already requires `isPropertyLike` -- an initializer
	// only exists on a property. Two mutation runs established it rather than reading: removing
	// `|| isMethodDeclaration` from the static arm SURVIVED the whole suite, while replacing the
	// instance arm's predicate with a genuinely different one was CAUGHT by 7 assertions.
	//
	// The difference upstream is expressing is real, so it is spelled out below rather than dropped:
	// a `MethodDefinition`'s `node.value` is the FunctionExpression holding the body, which is always
	// present and always function-like, so the conjunction is simply TRUE for every method. Here a
	// method has no initializer at all, so the same fact has to be stated as its own arm.
	//
	// The asymmetry between the two arms is upstream's and is load-bearing: `static foo() {}` IS a
	// `static-methods` member, while a non-static `foo() {}` is NOT an `instance-methods` member and
	// falls through to `everything-else`. Measured on the installed build with a control that fires:
	// a plain `classMethod()` placed after `render` under an `instance-methods`-first order is clean,
	// while an arrow property `foo = () => {}` in the same position reports.
	member.staticVariable = isStatic && isPropertyLike && !initializedToFunction
	member.staticMethod = isStatic && (isPropertyLike && initializedToFunction || isMethodDeclaration)
	member.instanceVariable = !isStatic && isPropertyLike && !initializedToFunction
	member.instanceMethod = !isStatic && isPropertyLike && initializedToFunction

	member.typeAnnotation = sortCompHasBareTypeAnnotation(node)

	return member
}

// sortCompPropertyInitializer reads the value of a property-shaped member, and says whether the
// member is property-shaped at all.
//
// The second return separates "this is a property with no value" from "this is not a property",
// which the flags above need to keep apart: a bare `foo;` in a class is an instance variable, while
// a `foo() {}` is not an instance variable at all.
func sortCompPropertyInitializer(node *ast.Node) (*ast.Node, bool) {
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		declaration := node.AsPropertyDeclaration()
		if declaration == nil {
			return nil, false
		}
		return declaration.Initializer, true
	case ast.KindPropertyAssignment:
		assignment := node.AsPropertyAssignment()
		if assignment == nil {
			return nil, false
		}
		return assignment.Initializer, true
	}
	return nil, false
}

// sortCompHasBareTypeAnnotation answers upstream's `!!node.typeAnnotation && node.value === null`.
//
// In ESTree a TypeScript annotation hangs off the property node as `typeAnnotation`; here `Type` is
// a sibling of the name on the property declaration, and `Initializer` is upstream's `value`. The
// second half is what keeps `state: Object = {};` out of the `type-annotations` group while
// `props: { text: string };` is in it, and invalid[6] in the corpus is the case that pins it.
func sortCompHasBareTypeAnnotation(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyDeclaration {
		return false
	}
	declaration := node.AsPropertyDeclaration()
	if declaration == nil {
		return false
	}
	return declaration.Type != nil && declaration.Initializer == nil
}

// sortCompMemberName answers upstream's `astUtil.getPropertyName` for the non-accessor kinds.
//
// Upstream returns the empty string when there is no readable name node, and an empty name matches
// no order entry, so a computed or otherwise unreadable member falls through to `everything-else`.
// That is reproduced rather than skipped, because skipping the member would change the INDEXES the
// pairwise comparison runs over and move findings on unrelated members.
//
// The kind guard before the text read is load-bearing rather than defensive: `Node.Text()` panics
// on several kinds rather than returning empty, and a computed member name like `[key]` reaches
// here. The walk recovers per FILE rather than per rule, so one such member would cost every rule
// in this package its verdict on that file.
func sortCompMemberName(node *ast.Node) string {
	// A constructor is a distinct node kind here with a nil `Name()`, while in ESTree it is a
	// `MethodDefinition` whose key is the identifier `constructor`. Without this arm the name reads
	// as empty, which is a real verdict change rather than a cosmetic one: `constructor` is an entry
	// in the default lifecycle group, so an empty name silently demotes every constructor to
	// `everything-else`. That moved 12 of this corpus's 53 cases, in both directions, and the
	// rendered message named the member as the empty string.
	if node.Kind == ast.KindConstructor {
		return "constructor"
	}

	name := node.Name()
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindPrivateIdentifier:
		return name.Text()
	}
	return ""
}

// sortCompGroupMatches answers whether one order entry selects one member.
//
// The if/else chain follows upstream's order exactly, and the order matters at one place: the
// literal-name test against `sortCompKnownLifecycleNames` comes BEFORE the regex branch, so an
// order entry spelled `render` is a member name rather than a pattern. Reordering these two changes
// the answer for any project whose order names a lifecycle method directly.
func sortCompGroupMatches(group string, member sortCompMember) bool {
	switch group {
	case "getters":
		return member.getter
	case "setters":
		return member.setter
	case "type-annotations":
		return member.typeAnnotation
	case "static-variables":
		return member.staticVariable
	case "static-methods":
		return member.staticMethod
	case "instance-variables":
		return member.instanceVariable
	case "instance-methods":
		return member.instanceMethod
	}

	if sortCompKnownLifecycleNames[group] {
		return group == member.name
	}

	// Upstream's `currentGroup.match(regExpRegExp)` is unanchored, so this reproduces the same
	// permissiveness. A pattern Go's RE2 cannot compile is treated as matching nothing rather than
	// panicking: upstream's `new RegExp` would throw and take the run down, and a rule that crashes
	// costs every OTHER rule its verdict on that file because the walk recovers per file.
	//
	// # The RE2 gap, measured rather than assumed
	//
	// Go's regexp is RE2 and has no lookaround or backreferences, so a JavaScript order pattern
	// using either compiles upstream and is dropped here. `core/no-param-reassign` takes the same
	// drop for the same reason and states it the same way, and the direction is deliberate: a
	// dropped pattern demotes its members to `everything-else`, which reports MORE rather than
	// less, and over-reporting is visible where under-reporting hides.
	//
	// Measured across this rule's whole corpus: 13 order entries are regex patterns and exactly ONE
	// uses lookaround, `^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$` in the case
	// recorded as `sortCompRe2LookaroundCase`. On that input the rule reports the same COUNT and the
	// same offending member as upstream and names a different partner in the message, because the
	// members the dropped pattern would have grouped fall through to `everything-else`. That case is
	// pinned in the test with both verdicts written out, rather than removed to make the suite
	// green.
	if submatches := sortCompRegexPattern.FindStringSubmatch(group); submatches != nil {
		pattern, compiled := sortCompCompilePattern(submatches[1], submatches[2])
		if compiled && pattern.MatchString(member.name) {
			return true
		}
		return false
	}

	return group == member.name
}

// sortCompCompilePattern builds a Go regexp from upstream's captured body and flags.
//
// Only `i`, `s` and `m` have Go equivalents; `g` and `y` are stateful match modifiers that mean
// nothing for a single `test()` call, and `u` is Unicode mode, which Go's regexp is always in. So
// the three that change a match are translated and the three that do not are dropped, which is the
// same set of behaviours upstream's `new RegExp(body, flags).test(name)` has.
func sortCompCompilePattern(body string, flags string) (*regexp.Regexp, bool) {
	inline := ""
	for _, flag := range flags {
		switch flag {
		case 'i', 's', 'm':
			inline += string(flag)
		}
	}
	if inline != "" {
		body = "(?" + inline + ")" + body
	}
	pattern, err := regexp.Compile(body)
	if err != nil {
		return nil, false
	}
	return pattern, true
}

// sortCompRefIndexes reproduces upstream's `getRefPropIndexes`: which positions in the resolved
// order a member is allowed to occupy.
//
// A member can match several entries, which is why this returns a list rather than an index. When
// nothing matches, the `everything-else` position is used, and when there is no `everything-else`
// entry either, upstream pushes `Infinity`, which sorts after everything. That is reproduced with
// `math.MaxInt` rather than a sentinel like -1, because the value is COMPARED with `<` and `>`
// against real indexes and a negative sentinel would invert every comparison it took part in.
func sortCompRefIndexes(methodsOrder []string, member sortCompMember) []int {
	indexes := []int{}
	for index, group := range methodsOrder {
		if sortCompGroupMatches(group, member) {
			indexes = append(indexes, index)
		}
	}
	if len(indexes) > 0 {
		return indexes
	}

	for index, group := range methodsOrder {
		if group == "everything-else" {
			return []int{index}
		}
	}
	return []int{math.MaxInt}
}

// sortCompComparison is the answer for one ordered pair.
type sortCompComparison struct {
	correct bool
	indexA  int
	indexB  int
}

// sortCompComparePropsOrder reproduces upstream's `comparePropsOrder`.
//
// A pair is correct when ANY combination of their allowed positions is defensible: the two share a
// position, or their reference order agrees with their source order. Only if no combination works
// is the pair a violation.
//
// The indexes it returns differ by branch, and that asymmetry is upstream's and is load-bearing.
// On success it returns the members' SOURCE positions; on failure it returns the last REFERENCE
// positions the loops considered. Those failure indexes are what `storeError` keys the error map on
// and what `dedupeErrors` looks entries up by, so returning source positions there would change
// which findings survive deduplication.
func sortCompComparePropsOrder(methodsOrder []string, members []sortCompMember, indexA int, indexB int) sortCompComparison {
	refIndexesA := sortCompRefIndexes(methodsOrder, members[indexA])
	refIndexesB := sortCompRefIndexes(methodsOrder, members[indexB])

	// Upstream computes these with `propertiesInfos.indexOf(propA)`, an identity lookup that returns
	// the FIRST structurally-equal element in some engines' terms but is reference equality here.
	// We are handed the positions directly, which is the same answer without the lookup.
	classIndexA := indexA
	classIndexB := indexB

	lastRefIndexA := 0
	lastRefIndexB := 0
	for _, refIndexA := range refIndexesA {
		lastRefIndexA = refIndexA
		for _, refIndexB := range refIndexesB {
			lastRefIndexB = refIndexB
			if refIndexA == refIndexB ||
				(refIndexA < refIndexB && classIndexA < classIndexB) ||
				(refIndexA > refIndexB && classIndexA > classIndexB) {
				return sortCompComparison{correct: true, indexA: classIndexA, indexB: classIndexB}
			}
		}
	}

	return sortCompComparison{correct: false, indexA: lastRefIndexA, indexB: lastRefIndexB}
}

// sortCompError is one member's accumulated case against it.
type sortCompError struct {
	member sortCompMember

	// score is how many pairs this member lost, which is what `dedupeErrors` compares.
	score int

	// closestDistance is the smallest index gap to a losing partner seen so far, and
	// closestMember / closestIndex are that partner. The message names this partner rather than the
	// most recent one.
	closestDistance int
	closestMember   sortCompMember
	closestHasRef   bool
	closestIndex    int
}

// sortCompCheckOrder runs the pairwise vote over one component's members and reports what survives.
func sortCompCheckOrder(ctx rule.Context, methodsOrder []string, members []sortCompMember) {
	errors := map[int]*sortCompError{}

	storeError := func(memberA sortCompMember, indexA int, memberB sortCompMember, indexB int) {
		existing, alreadyRecorded := errors[indexA]
		if !alreadyRecorded {
			existing = &sortCompError{member: memberA, closestDistance: math.MaxInt}
			errors[indexA] = existing
		}
		existing.score++

		// Upstream compares the NAMES rather than the nodes here, and the comment above it says
		// "stop here if we already have pushed another node at this position". Two different
		// members can land on the same reference index, and only the first one's message is kept.
		//
		// # A REPORTED mutation survivor, with what was and was not established
		//
		// Replacing this condition with `false` survives the whole suite. It is kept, and the
		// survival is reported rather than argued away, because the two halves of the question came
		// out differently.
		//
		// Reachability: measured, not assumed. Counting both sides across every corpus case gives
		// REACHED=46, FIRED=1 -- the branch is live, and the one firing is `invalid-12`, where `bar`
		// and `classMethod` share a reference index. An earlier attempt to measure this with
		// `println` reported 0 for both the guard AND its control, which is the broken-instrument
		// zero this repository keeps finding; the counter version has a live control at 46.
		//
		// Effect: NOT established. On that one firing, and on five hand-built variations of its
		// shape (more plain methods, more statics, a reordered class), the guarded and unguarded
		// rules produce byte-identical findings, because `dedupeErrors` resolves the pair the same
		// way either side of this early return. So the guard is reachable and, on every input
		// available here, subsumed.
		//
		// It stays because it is upstream's, and because "I could not build the input that
		// separates them" is not the same claim as "no such input exists". Deleting a guard on the
		// strength of a survivor is how a port acquires a divergence nobody can find later.
		if existing.member.name != memberA.name {
			return
		}
		distance := indexA - indexB
		if distance < 0 {
			distance = -distance
		}
		if distance > existing.closestDistance {
			return
		}
		existing.closestDistance = distance
		existing.closestMember = memberB
		existing.closestHasRef = true
		existing.closestIndex = indexB
	}

	for indexA := range members {
		for indexB := range members {
			if indexA == indexB {
				continue
			}
			comparison := sortCompComparePropsOrder(methodsOrder, members, indexA, indexB)
			if comparison.correct {
				continue
			}
			storeError(members[indexA], comparison.indexA, members[indexB], comparison.indexB)
		}
	}

	// dedupeErrors, walked in ascending numeric key order to match `Object.entries` on an
	// integer-keyed JavaScript object. The walk MUTATES the map, so an unordered one would make the
	// surviving finding depend on Go's map randomisation; see the rule doc.
	for _, index := range sortCompSortedIndexes(errors) {
		current, stillPresent := errors[index]
		if !stillPresent {
			continue
		}
		if !current.closestHasRef {
			continue
		}
		partner, partnerRecorded := errors[current.closestIndex]
		if !partnerRecorded {
			continue
		}
		if current.score > partner.score {
			delete(errors, current.closestIndex)
			continue
		}
		delete(errors, index)
	}

	for _, index := range sortCompSortedIndexes(errors) {
		reported := errors[index]
		if !reported.closestHasRef {
			continue
		}
		position := "after"
		if index < reported.closestIndex {
			position = "before"
		}
		ctx.ReportNode(
			reported.member.node,
			sortCompUnsortedProps(reported.member.name, position, reported.closestMember.name),
		)
	}
}

// sortCompSortedIndexes returns the map's keys in ascending numeric order.
//
// This is the determinism the rule doc argues for, and it is a property of the RULE rather than of
// any fixture: `Object.entries` on an integer-keyed JavaScript object iterates ascending, both
// call sites walk while deleting, and Go's map order is randomised.
func sortCompSortedIndexes(errors map[int]*sortCompError) []int {
	indexes := make([]int, 0, len(errors))
	for index := range errors {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	return indexes
}
