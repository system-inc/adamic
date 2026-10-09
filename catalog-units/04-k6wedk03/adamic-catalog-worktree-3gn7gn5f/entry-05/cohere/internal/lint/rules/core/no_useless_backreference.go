package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUselessBackreference flags a backreference that can never match the group it names.
//
//	valid:   /(a)\1/
//	valid:   /(?<foo>a)\k<foo>/
//	valid:   /(?:a|(b))\1/
//	invalid: /\1(a)/
//	invalid: /(a|\1b)/
//	invalid: /(a\1)/
//	invalid: /\1(?!(a))/
//	invalid: /(?<=(a)\1)b/
//
// A backreference names a group and asks for the text that group captured. When the group cannot
// have captured anything at the moment the reference is evaluated, the reference matches the empty
// string instead of failing, so the pattern silently matches more than its author wrote it to
// match. That is the defect: not a syntax error the engine refuses, but a pattern that quietly
// means something else.
//
// # Five ways a group cannot have run, and they are five findings
//
// It reports five ids, ESLint's, since the kind is the diagnosis rather than decoration. A port collapsing
// them to one id passes a fixture that asserts only that something fired, while calling a
// different-alternative case a forward reference.
//
//	nested                  /(a\1)/       the group is still open, so it has captured nothing
//	disjunctive             /(a|\1b)/     only one branch runs, so the other never captured
//	forward                 /\1(a)/       matching is left to right and the group has not run
//	backward                /(?<=(a)\1)b/ a lookbehind matches right to left, so this is early
//	intoNegativeLookaround  /\1(?!(a))/   the group only runs on a path that fails
//
// oxc names the middle pair the other way round: its `Problem::Forward` is the lookbehind case. The
// ids here are ESLint's (#jjfa7qb), where forward is the ordinary left-to-right case.
//
// oxc's doc comment also lists `/(?<=\1(a))b/` as incorrect, and that exact pattern is a pass case in
// both corpora, correctly: a lookbehind matches right to left, so the group is reached before the
// reference and has captured.
//
// # Why this is not built on regexpattern.Walk
//
// `Walk` reports characters with a quantifier and class depth, and its own doc says a caller
// wanting alternation or lookaround should extend it rather than find it half-built. This rule is
// entirely about alternation and lookaround: every one of the five findings above is a question
// about which construct encloses the reference and which encloses the group. `Walk` also reports
// `ClassDepth` without saying which class, and does not recurse into v-flag nested classes, so the
// one thing it could have answered here is answered wrong for `new RegExp('[[]\\1](a)', 'v')`.
//
// So this scans with `regexsyntax.SkipPatternEscape` and `regexsyntax.ClassEnd`, the same two
// primitives `regexpattern.CountCapturingGroups` is built on, and keeps a path of enclosing nodes.
// `ClassEnd` handles v-flag nesting and `\q{...}`, which is what makes the two class cases in the
// corpus come out right without a second scanner.
//
// # Where a finding points, and which calls are read
//
// Each finding names the whole regex literal, or the whole `RegExp(...)` or `new RegExp(...)` call,
// as ESLint reports it (#jjfa7qb). The message names the backreference and the group, so pointing at
// the node loses nothing a reader needs, and it keeps one place per pattern however many references
// in it are useless.
//
// Every regex literal is checked under its own flags, including one passed to `RegExp`. The calls are
// found the way ESLint finds them, through the shelf's port of eslint-utils' ReferenceTracker: a call
// or construction of the global `RegExp`, directly or through anything the tracker follows, so a
// local named `RegExp` is not it and `const r = RegExp; new r(...)` is. The pattern and the flags are
// read as getStringIfConstant reads them, through reference.ConstantStringIn: a constant binding's
// value counts, and flags that are not constant read as none, as ESLint's `flags || ""` does. A
// pattern that is not constant is not checked.
var NoUselessBackreference = rule.Rule{
	Name: "no-useless-backreference",

	// The tracker tells the global `RegExp` from a local that shares its spelling by asking where its
	// symbol is declared, and a constant argument is followed to its binding the same way.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := regexsyntax.PatternAndFlags(node.Text())
				if pattern == "" {
					return
				}
				reportUselessBackreferences(ctx, node, pattern, flags)
			},
			ast.KindSourceFile: func(node *ast.Node) {
				// Every trace starts at a reference to `RegExp`, so a file that never spells it has
				// nothing to follow, and the tracker's index is never built.
				if ctx.TypeChecker == nil || !strings.Contains(ctx.SourceFile.Text(), "RegExp") {
					return
				}
				tracker := reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)
				for _, tracked := range tracker.GlobalReferences(regExpCallTraceMap) {
					checkBackreferencesInRegExpCall(ctx, tracked.Node)
				}
			},
		}
	},
}

// regExpCallTraceMap asks the tracker for every call and construction of the global RegExp.
var regExpCallTraceMap = map[string]*reference.TraceMap{"RegExp": {Call: true, Construct: true}}

// checkBackreferencesInRegExpCall checks a call or construction of RegExp whose pattern is a constant.
func checkBackreferencesInRegExpCall(ctx rule.Context, call *ast.Node) {
	arguments := call.Arguments()
	if len(arguments) == 0 {
		return
	}
	pattern, isConstant := reference.ConstantStringIn(ctx, arguments[0])
	if !isConstant {
		return
	}
	flags := ""
	if len(arguments) > 1 {
		flags, _ = reference.ConstantStringIn(ctx, arguments[1])
	}
	reportUselessBackreferences(ctx, call, pattern, flags)
}

// regexNodeKind is what a node on a backreference's or group's enclosing path is.
type regexNodeKind int

const (
	// regexNodeAlternative is one branch of a `|` disjunction. Two paths diverging at an
	// alternative is exactly the different-branch case.
	regexNodeAlternative regexNodeKind = iota
	// regexNodeCapturingGroup is `(...)` or `(?<name>...)`.
	regexNodeCapturingGroup
	// regexNodeGroup is `(?:...)`, which nests but captures nothing.
	regexNodeGroup
	// regexNodeLookahead is `(?=...)` or `(?!...)`.
	regexNodeLookahead
	// regexNodeLookbehind is `(?<=...)` or `(?<!...)`, which matches right to left.
	regexNodeLookbehind
)

// regexNode is one enclosing construct in the pattern.
type regexNode struct {
	kind regexNodeKind
	// negative distinguishes `(?!` from `(?=` and `(?<!` from `(?<=`. A group inside a negative
	// lookaround only ever runs on a path that then fails, so it never keeps a capture.
	negative bool
}

// capturingGroup is a group the pattern declares, with where it sits and what encloses it.
type capturingGroup struct {
	name  string
	start int
	end   int
	// path is the chain of enclosing node indices, outermost first, ending at the group itself.
	path []int
}

// backreference is a `\1` or `\k<name>` in the pattern.
type backreference struct {
	// index is the group number for `\1`, or zero when the reference is by name.
	index int
	name  string
	start int
	end   int
	path  []int
}

// regexStructure is everything the rule needs about one pattern.
type regexStructure struct {
	nodes          []regexNode
	groups         []capturingGroup
	backreferences []backreference
}

// reportUselessBackreferences scans one pattern and reports every backreference that cannot match, at
// the node the pattern came from.
func reportUselessBackreferences(ctx rule.Context, node *ast.Node, pattern string, flagsText string) {
	flags := regexsyntax.ParseRegexFlags(flagsText)

	structure, ok := scanRegexStructure(pattern, flags)
	if !ok {
		return
	}
	if len(structure.backreferences) == 0 {
		return
	}

	// The group total decides whether `\1` is a backreference at all. `/\1/` has no groups, so the
	// escape is an octal, not a reference, and upstream is silent on it. This has to be the total
	// rather than a running count, because a pattern may reference a group declared later.
	groupCount := len(structure.groups)

	// A pattern the engine would refuse is not this rule's business. Upstream parses first and
	// reports nothing when the parse fails, so `new RegExp('\\1(a){', 'u')` is silent while
	// `RegExp('\\1(a){', flags)` reports: the brace is a syntax error under `u` and a literal
	// without it, and the second call's flags are unknown so no `u` is assumed. Four of upstream's
	// clean cases are this and nothing else, and a scanner permissive enough to walk the pattern
	// anyway reports a real backreference sitting next to the broken part.
	if !patternIsWellFormed(pattern, structure, groupCount, flags) {
		return
	}

	for _, reference := range structure.backreferences {
		groups := resolveBackreference(reference, structure.groups, groupCount)
		if len(groups) == 0 {
			continue
		}

		// Every group the reference names is judged, and one the reference can reach clears it. Of
		// the problems left, those in the reference's own disjunction are reported if there are any,
		// else all of them: the first is named and the rest are counted, as ESLint words it.
		type judged struct {
			problem backreferenceProblem
			group   capturingGroup
		}
		var all, sameDisjunction []judged
		reachable := false
		for _, group := range groups {
			problem, ok := classifyBackreference(reference, group, structure.nodes)
			if !ok {
				reachable = true
				break
			}
			all = append(all, judged{problem, group})
			if problem.id != problemAnotherAlternative.id {
				sameDisjunction = append(sameDisjunction, judged{problem, group})
			}
		}
		if reachable {
			continue
		}
		toReport := all
		if len(sameDisjunction) > 0 {
			toReport = sameDisjunction
		}
		first := toReport[0]
		ctx.ReportNode(node, rule.Message{
			Id: first.problem.id,
			Description: first.problem.describe(pattern[reference.start:reference.end],
				pattern[first.group.start:first.group.end], len(toReport)-1),
		})
	}
}

// patternIsWellFormed reports whether the engine would accept the pattern.
//
// This is not a full validator and does not try to be. It answers the three ways a pattern in this
// corpus is refused, all of which only bite under `u` or `v`, where the grammar stops tolerating
// what annex B allows:
//
//	a numeric reference past the last group   `\1(a)\2` has one group, so `\2` is not an escape
//	a named reference naming no group         `\k<bar>` with no `bar` group
//	a lone `{` that opens no quantifier       `\1(a){`
//
// A named reference with no matching group is refused under every flag once the pattern declares
// any named group at all, which is why `new RegExp('\\k<foo>(?<foo>a)\\k<bar>')` is clean with no
// flags while `RegExp('\\1(a)\\k<foo>', 'u')` needs the `u` to be refused.
func patternIsWellFormed(pattern string, structure regexStructure, groupCount int, flags regexsyntax.RegexFlags) bool {
	hasNamedGroup := false
	for _, group := range structure.groups {
		if group.name != "" {
			hasNamedGroup = true
			break
		}
	}

	for _, reference := range structure.backreferences {
		if reference.name != "" {
			// A `\k<name>` naming nothing is a syntax error whenever the pattern has any named
			// group, and under `u` or `v` even when it has none.
			named := false
			for _, group := range structure.groups {
				if group.name == reference.name {
					named = true
					break
				}
			}
			if !named && (hasNamedGroup || flags.UV()) {
				return false
			}
			continue
		}
		// Under `u` or `v` a numeric escape with no such group is not a legacy octal, it is an
		// error. Without the flag it is an octal and the rule is right to ignore it.
		if flags.UV() && reference.index > groupCount {
			return false
		}
	}

	if flags.UV() && hasUnmatchedQuantifierBrace(pattern, flags) {
		return false
	}
	return true
}

// hasUnmatchedQuantifierBrace reports whether the pattern holds a `{` that opens no valid
// quantifier, which annex B reads as a literal brace and `u` mode refuses.
func hasUnmatchedQuantifierBrace(pattern string, flags regexsyntax.RegexFlags) bool {
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '\\':
			step, ok := regexsyntax.SkipPatternEscape(pattern, index, flags)
			if !ok {
				return true
			}
			index += step
		case '[':
			end, ok := regexsyntax.ClassEnd(pattern, index, flags)
			if !ok {
				return true
			}
			index = end
		case '{':
			if !opensQuantifier(pattern, index) {
				return true
			}
			index++
		default:
			index++
		}
	}
	return false
}

// opensQuantifier reports whether the `{` at index begins `{n}`, `{n,}` or `{n,m}`.
func opensQuantifier(pattern string, index int) bool {
	cursor := index + 1
	digits := 0
	for cursor < len(pattern) && pattern[cursor] >= '0' && pattern[cursor] <= '9' {
		cursor++
		digits++
	}
	if digits == 0 {
		return false
	}
	if cursor < len(pattern) && pattern[cursor] == ',' {
		cursor++
		for cursor < len(pattern) && pattern[cursor] >= '0' && pattern[cursor] <= '9' {
			cursor++
		}
	}
	return cursor < len(pattern) && pattern[cursor] == '}'
}

// backreferenceProblem is one of the five ways a group cannot have run.
type backreferenceProblem struct {
	id     string
	reason string
}

// describe builds the finding's sentence from the reference and the group it names.
func (problem backreferenceProblem) describe(reference string, group string, otherGroups int) string {
	also := ""
	switch {
	case otherGroups == 1:
		also = " (and so does one other group of that name)"
	case otherGroups > 1:
		also = fmt.Sprintf(" (and so do %d other groups of that name)", otherGroups)
	}
	return fmt.Sprintf("The backreference %s can never match, because %s %s%s. It matches the empty "+
		"string instead of failing, so the pattern quietly matches more than it looks like it does.",
		reference, group, problem.reason, also)
}

var (
	problemInsideItsOwnGroup = backreferenceProblem{
		id:     "nested",
		reason: "encloses the reference and so has captured nothing yet when it is reached",
	}
	problemAnotherAlternative = backreferenceProblem{
		id:     "disjunctive",
		reason: "sits in a different branch of the same disjunction, and only one branch ever runs",
	}
	problemBeforeItsGroup = backreferenceProblem{
		id:     "forward",
		reason: "appears later in the pattern, and matching runs left to right",
	}
	problemAfterItsGroupInLookbehind = backreferenceProblem{
		id:     "backward",
		reason: "appears earlier in a lookbehind, which matches right to left",
	}
	problemIntoNegativeLookaround = backreferenceProblem{
		id:     "intoNegativeLookaround",
		reason: "sits in a negative lookaround, so it only runs on a path that then fails",
	}
)

// resolveBackreference finds the group a reference names.
//
// A numeric reference past the group count is not a reference at all but a legacy octal escape,
// which is why `/\2(a)/` and `/\11(a)/` are both silent. A named reference with no matching group
// is a syntax error the engine refuses, so it is skipped rather than reported.
//
// A name can be declared more than once under ES2025 duplicate named groups, where
// `(?<foo>a)|(?<foo>b)` is legal because the two live in different branches. Upstream reports such
// a reference only when **every** group carrying the name is unreachable from it, which is why
// `/((?<foo>bar)\k<foo>|(?<foo>baz))/` is clean: one of the two `foo` groups does run first. So a
// named reference resolves to every group carrying its name, in pattern order.
func resolveBackreference(reference backreference, groups []capturingGroup, groupCount int) []capturingGroup {
	if reference.name == "" {
		if reference.index < 1 || reference.index > groupCount {
			return nil
		}
		return groups[reference.index-1 : reference.index]
	}
	var named []capturingGroup
	for _, group := range groups {
		if group.name == reference.name {
			named = append(named, group)
		}
	}
	return named
}

// classifyBackreference decides which of the five problems a reference has, or that it has none.
//
// The order matters and matches upstream. Containment is checked first because a reference inside
// its own group is the one case where the two paths do not diverge at all. Then the paths are cut
// at their lowest common ancestor, and what the group's side of the cut begins with answers the
// disjunction question. Direction decides between the two ordering problems, and the negative
// lookaround check comes last because a group inside one can still be reachable in every other
// respect.
func classifyBackreference(reference backreference, group capturingGroup, nodes []regexNode) (backreferenceProblem, bool) {
	if group.start <= reference.start && reference.end <= group.end {
		return problemInsideItsOwnGroup, true
	}

	// Walk both paths together while they agree. Index 0 is the pattern root and is shared by
	// everything, so the comparison starts at 1 and the cut sits one past the last agreement.
	common := 1
	for common < len(reference.path) && common < len(group.path) &&
		reference.path[common] == group.path[common] {
		common++
	}

	groupCut := group.path[common:]

	// Direction is decided by the innermost lookaround the two share. A lookbehind *below* the cut
	// on one side only does not reverse the other side, which is what keeps `/(?<=(a))b\1/` clean:
	// the reference sits outside the lookbehind and matches forwards.
	matchesBackwards := false
	for index := common - 1; index >= 0; index-- {
		node := nodes[reference.path[index]]
		if node.kind == regexNodeLookahead || node.kind == regexNodeLookbehind {
			matchesBackwards = node.kind == regexNodeLookbehind
			break
		}
	}

	// Two paths that diverge at an alternative are in different branches of one disjunction, and
	// only one branch ever runs.
	if len(groupCut) > 0 && nodes[groupCut[0]].kind == regexNodeAlternative {
		return problemAnotherAlternative, true
	}

	if !matchesBackwards && reference.end <= group.start {
		return problemBeforeItsGroup, true
	}

	if matchesBackwards && group.end <= reference.start {
		return problemAfterItsGroupInLookbehind, true
	}

	// A group under a negative lookaround the reference is not itself under can never keep a
	// capture, because the assertion succeeds only when the group's path failed.
	for _, nodeIndex := range groupCut {
		node := nodes[nodeIndex]
		if node.negative && (node.kind == regexNodeLookahead || node.kind == regexNodeLookbehind) {
			return problemIntoNegativeLookaround, true
		}
	}

	return backreferenceProblem{}, false
}

// scanRegexStructure reads a pattern into the groups, backreferences and enclosing paths the rule
// needs. It returns false for a pattern it cannot make sense of, which is a syntax error the engine
// has already refused and where a second opinion would be noise.
func scanRegexStructure(pattern string, flags regexsyntax.RegexFlags) (regexStructure, bool) {
	scanner := &regexStructureScanner{pattern: pattern, flags: flags}
	if !scanner.run() {
		return regexStructure{}, false
	}
	return scanner.structure, true
}

// regexStructureScanner walks a pattern once, keeping a stack of what encloses the current byte.
type regexStructureScanner struct {
	pattern   string
	flags     regexsyntax.RegexFlags
	structure regexStructure
	// stack holds the indices of the currently open nodes, outermost first.
	stack []int
	// groupStack records, for each open `(`, the group index it opened or -1 for a non-capturing
	// one, so the closing `)` can finish the right group's span.
	groupStack []int
	// alternativeStack records the alternative node index open inside each enclosing group, so a
	// `|` can close one branch and open the next at the right level.
	alternativeStack []int
}

// push adds a node and returns its index.
func (scanner *regexStructureScanner) push(kind regexNodeKind, negative bool) int {
	scanner.structure.nodes = append(scanner.structure.nodes, regexNode{kind: kind, negative: negative})
	index := len(scanner.structure.nodes) - 1
	scanner.stack = append(scanner.stack, index)
	return index
}

// pop removes the innermost open node.
func (scanner *regexStructureScanner) pop() {
	if len(scanner.stack) > 0 {
		scanner.stack = scanner.stack[:len(scanner.stack)-1]
	}
}

// currentPath copies the open stack, which is what a group or reference records about itself.
func (scanner *regexStructureScanner) currentPath() []int {
	path := make([]int, len(scanner.stack))
	copy(path, scanner.stack)
	return path
}

func (scanner *regexStructureScanner) run() bool {
	pattern := scanner.pattern

	// The root node stands for the whole pattern, so every path shares index 0 and the walk that
	// finds the lowest common ancestor always has somewhere to start.
	scanner.push(regexNodeGroup, false)
	scanner.openAlternative()

	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '\\':
			step, ok := scanner.readEscape(index)
			if !ok {
				return false
			}
			index += step

		case '[':
			// Everything inside a character class is a literal, which is the whole of
			// `/^[\1](a)$/` being clean: the `\1` there matches a control character, not a group.
			end, ok := regexsyntax.ClassEnd(pattern, index, scanner.flags)
			if !ok {
				return false
			}
			index = end

		case '(':
			index = scanner.openGroup(index)

		case ')':
			if !scanner.closeGroup(index) {
				return false
			}
			index++

		case '|':
			scanner.closeAlternative()
			scanner.openAlternative()
			index++

		default:
			index++
		}
	}

	// A pattern whose parentheses do not balance is a syntax error. The root group and its
	// alternative are the only two nodes that should still be open.
	return len(scanner.groupStack) == 0
}

// openAlternative starts a new branch at the current level.
func (scanner *regexStructureScanner) openAlternative() {
	index := scanner.push(regexNodeAlternative, false)
	scanner.alternativeStack = append(scanner.alternativeStack, index)
}

// closeAlternative ends the branch at the current level.
func (scanner *regexStructureScanner) closeAlternative() {
	if len(scanner.alternativeStack) > 0 {
		scanner.alternativeStack = scanner.alternativeStack[:len(scanner.alternativeStack)-1]
	}
	scanner.pop()
}

// openGroup handles a `(` and returns the index just past whatever prefix it consumed.
func (scanner *regexStructureScanner) openGroup(index int) int {
	pattern := scanner.pattern
	kind := regexNodeCapturingGroup
	negative := false
	name := ""
	next := index + 1

	if next < len(pattern) && pattern[next] == '?' {
		switch {
		case next+1 < len(pattern) && pattern[next+1] == ':':
			kind, next = regexNodeGroup, next+2
		case next+1 < len(pattern) && pattern[next+1] == '=':
			kind, next = regexNodeLookahead, next+2
		case next+1 < len(pattern) && pattern[next+1] == '!':
			kind, negative, next = regexNodeLookahead, true, next+2
		case next+2 < len(pattern) && pattern[next+1] == '<' && pattern[next+2] == '=':
			kind, next = regexNodeLookbehind, next+3
		case next+2 < len(pattern) && pattern[next+1] == '<' && pattern[next+2] == '!':
			kind, negative, next = regexNodeLookbehind, true, next+3
		case next+1 < len(pattern) && pattern[next+1] == '<':
			// `(?<name>` is the one `(?<` form that captures.
			closing := strings.IndexByte(pattern[next+2:], '>')
			if closing < 0 {
				kind, next = regexNodeGroup, next+2
			} else {
				name = pattern[next+2 : next+2+closing]
				next = next + 2 + closing + 1
			}
		default:
			// A `(?` this scanner does not know is treated as a plain nesting group rather than
			// guessed at, so the paths stay right even if the construct is exotic.
			kind, next = regexNodeGroup, next+2
		}
	}

	scanner.push(kind, negative)

	if kind == regexNodeCapturingGroup {
		scanner.structure.groups = append(scanner.structure.groups, capturingGroup{
			name:  name,
			start: index,
			path:  scanner.currentPath(),
		})
		scanner.groupStack = append(scanner.groupStack, len(scanner.structure.groups)-1)
	} else {
		scanner.groupStack = append(scanner.groupStack, -1)
	}

	scanner.openAlternative()
	return next
}

// closeGroup handles a `)`, finishing the group's span if it captured.
func (scanner *regexStructureScanner) closeGroup(index int) bool {
	if len(scanner.groupStack) == 0 {
		return false
	}
	scanner.closeAlternative()

	groupIndex := scanner.groupStack[len(scanner.groupStack)-1]
	scanner.groupStack = scanner.groupStack[:len(scanner.groupStack)-1]
	if groupIndex >= 0 {
		scanner.structure.groups[groupIndex].end = index + 1
	}
	scanner.pop()
	return true
}

// readEscape handles a `\` outside a character class and returns how many bytes it consumed.
//
// Two of them are backreferences and the rest are skipped by the shelf's own scanner, which already
// knows the widths of `\x`, `\u{...}`, `\p{...}` and `\q{...}` under each flag.
func (scanner *regexStructureScanner) readEscape(index int) (int, bool) {
	pattern := scanner.pattern
	if index+1 >= len(pattern) {
		return 0, false
	}

	// `\1` through `\9...`: a numeric backreference, or a legacy octal when no such group exists.
	// Which of the two it is cannot be decided here, because the group may be declared later in the
	// pattern; the total is known only after the scan, so the decision is deferred to resolution.
	if pattern[index+1] >= '1' && pattern[index+1] <= '9' {
		end := index + 1
		value := 0
		for end < len(pattern) && pattern[end] >= '0' && pattern[end] <= '9' {
			value = value*10 + int(pattern[end]-'0')
			end++
		}
		scanner.structure.backreferences = append(scanner.structure.backreferences, backreference{
			index: value,
			start: index,
			end:   end,
			path:  scanner.currentPath(),
		})
		return end - index, true
	}

	// `\k<name>`: a named backreference. Bare `\k` with no angle bracket is an identity escape
	// outside unicode mode, and the shelf's skip handles it.
	if pattern[index+1] == 'k' && index+2 < len(pattern) && pattern[index+2] == '<' {
		closing := strings.IndexByte(pattern[index+3:], '>')
		if closing >= 0 {
			end := index + 3 + closing + 1
			scanner.structure.backreferences = append(scanner.structure.backreferences, backreference{
				name:  pattern[index+3 : index+3+closing],
				start: index,
				end:   end,
				path:  scanner.currentPath(),
			})
			return end - index, true
		}
	}

	step, ok := regexsyntax.SkipPatternEscape(pattern, index, scanner.flags)
	if !ok {
		return 0, false
	}
	return step, true
}
