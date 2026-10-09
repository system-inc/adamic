package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageAliasNotAssignedToThis = rule.Message{
	Id: "aliasNotAssignedToThis",
	Description: "This name is designated as the alias for `this`, and it is holding something " +
		"else. A reader who knows the convention will read it as the captured context and be wrong, " +
		"which is worse than an ordinary misleading name because the convention is what made it " +
		"trustworthy.",
}

var messageUnexpectedAlias = rule.Message{
	Id: "unexpectedAlias",
	Description: "This captures `this` under a name that is not the designated alias. Every " +
		"capture in the codebase should read the same way, so a reader recognises the pattern " +
		"instead of working out what each new name means.",
}

// ConsistentThisSettings is the decoded option surface.
//
// Upstream's schema is a bare array of strings and its default is `["that"]`, so an absent option
// is one designated alias rather than none. The distinction is load bearing in both directions: an
// empty list makes the aliasNotAssignedToThis half unreachable, and it makes the unexpectedAlias
// half report every capture of `this` under any name at all.
type ConsistentThisSettings struct {
	// Aliases are the names designated to hold `this`.
	Aliases []string
}

// DefaultConsistentThisSettings is upstream's `defaultOptions: ["that"]`.
func DefaultConsistentThisSettings() ConsistentThisSettings {
	return ConsistentThisSettings{Aliases: []string{"that"}}
}

// DecodeConsistentThisOptions reads the alias list off the config.
//
// Hand rolled rather than routed through the generic helper for two reasons. The default is
// `["that"]` rather than the zero value, and a generic decoder would turn an absent option into an
// empty list, which silently inverts the rule rather than disabling it. And upstream's option
// surface is VARIADIC: each alias is its own option element, `["error", "self", "vm"]`, so the rule
// registers with `DecodeOptionList` and is handed that list, which is upstream's `context.options`
// exactly.
//
// The config layer used to keep only the first element, so that spelling arrived here as `"self"`
// alone with `"vm"` dropped, and this decoder accepted a nested `["error", ["self", "vm"]]` to work
// around it. The workaround is refused now: an element that is not a string is an error, because
// upstream's schema says every element is a non-empty string.
func DecodeConsistentThisOptions(list []byte) (any, error) {
	if len(list) == 0 {
		return DefaultConsistentThisSettings(), nil
	}

	var aliases []string
	if err := json.Unmarshal(list, &aliases); err != nil {
		return DefaultConsistentThisSettings(), fmt.Errorf(
			"consistent-this takes each alias as its own string element, as in "+
				"[\"error\", \"self\", \"vm\"]: %w", err)
	}
	// Upstream's schema says `minLength: 1` on each item, so an empty name is refused rather than
	// quietly matching every unnamed thing.
	for _, alias := range aliases {
		if alias == "" {
			return DefaultConsistentThisSettings(), fmt.Errorf(
				"consistent-this takes non-empty aliases, got an empty string")
		}
	}
	if len(aliases) == 0 {
		return DefaultConsistentThisSettings(), nil
	}
	return ConsistentThisSettings{Aliases: aliases}, nil
}

// ConsistentThis flags a capture of `this` under a name other than the designated alias, and a
// designated alias holding anything other than `this`.
//
//	valid:   var that = this
//	valid:   var self; self = this            (with the alias "self")
//	valid:   var {foo, bar} = this
//	invalid: var context = this
//	invalid: var self = 42                    (with the alias "self")
//	invalid: var self                         (with the alias "self")
//
// # Two judgments, and one input can trip both
//
// The rule asks two questions that happen to share an option. Is a designated alias holding
// something other than `this`, and is `this` being captured under a name that is not designated.
// They are separate messages and a single input can produce both: with the alias `self`,
// `var self; self = 42` reports twice, once for the declaration that never received `this` and once
// for the assignment that gave it something else. A fixture asserting one finding per input is
// wrong for it, and upstream's corpus states the pair explicitly.
//
// # Destructuring is exempt, and the exemption is the whole reason four clean cases exist
//
// `var {foo, bar} = this` and `[foo, bar] = this` capture properties of `this` rather than `this`
// itself, so no name here is an alias for the context. Upstream tests the binding form and declines
// before reading a name. A port matching on the initializer alone reports all four.
//
// # A compound assignment to the alias reports even when the value is `this`
//
// `self += this` is a finding, because the alias ends up holding a concatenation rather than the
// context. Upstream reads `node.operator !== "="` for this, and it is a real discrimination rather
// than defensive coding: the assignment's value IS a `this` expression, so a port testing only the
// value reports nothing. Measured on the installed rule, `self ||= this` reports TWICE, since the
// declaration also never received a plain assignment.
//
// # The declared-and-never-assigned half needs a scope model, and ours is measured rather than
// derived
//
// `var self` with the alias `self` reports, and `var self; self = this` does not, so the rule has to
// find whether a later plain assignment of `this` reached that declaration. Upstream asks
// eslint-scope for the variable and walks `variable.references`, requiring `reference.from === scope`
// so that a write from an inner scope does not count.
//
// We have no scope table, so this asks the same question of the tree, and what `reference.from`
// actually means here was measured against the installed rule rather than read off the source. It is
// the innermost ESLINT scope, and eslint opens one for a block, a switch, a loop body, a catch and a
// `with`, but NOT for an unbraced `if` consequent, a labeled statement, or any expression nesting.
// So all of these report, with the alias `self`:
//
//	var self; { self = this; }                  a bare block
//	var self; for(;;) { self = this; }          a loop BODY, not the loop
//	var self; while(a) { self = this; }
//	var self; switch(a){case 1: self = this;}
//	var self; try { self = this; } catch(e) {}
//	var self; with(o) self = this;
//
// while these are clean:
//
//	var self; if(a) self = this;                no block
//	var self; label: self = this;
//	var self; for(var i=0;;) self = this;       a loop with no block body
//	var self; (self = this);                    expression nesting is invisible
//	var self; a, self = this;
//	var self; a ? self = this : 0;
//
// Reproduced rather than corrected. In every reporting case above the write reaches the same VAR
// declaration and the alias does end up holding `this`, so upstream is reporting correct code, and
// the intuitive reading of the rule would call all of them clean. That intuition is the thing most
// likely to be wrong here, so it is written down beside the measurement rather than acted on.
//
// # This runs per scope, and the walk owns its own descent
//
// The declared-and-never-assigned judgment is about a whole scope rather than about one node, since
// the assignment that rescues a declaration may be written anywhere below it. The walk here is
// pre-order, so a listener on the declaration cannot see what follows. Everything therefore happens
// inside one source-file listener, which fires before its children.
var ConsistentThis = rule.Rule{
	Name: "consistent-this",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[ConsistentThisSettings](options)
		if !ok || len(settings.Aliases) == 0 {
			// A rule configured as a bare severity is handed nil options, and the zero value of the
			// struct is an empty alias list, which would invert the rule rather than disable it.
			settings = DefaultConsistentThisSettings()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkConsistentThisScope(ctx, node, settings.Aliases)
			},
		}
	},
}

// checkConsistentThisScope judges one scope and descends into the scopes below it.
//
// A scope here is what upstream hooks: the program, a function declaration and a function
// expression. Arrow functions are deliberately absent, matching upstream's listener list, and that
// is not an oversight on either side: an arrow has no `this` of its own, so a capture inside one
// belongs to the enclosing scope.
func checkConsistentThisScope(ctx rule.Context, scope *ast.Node, aliases []string) {
	// Declarations of an alias in THIS scope, and whether a plain assignment of `this` rescued each.
	declaredAliases := []*ast.Node{}
	rescuedAliases := map[string]bool{}
	nestedScopes := []*ast.Node{}

	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		// A nested scope is judged on its own pass, so the descent stops here and resumes below.
		if isConsistentThisScope(current) {
			nestedScopes = append(nestedScopes, current)
			return false
		}
		// An arrow is NOT a scope this rule judges, and it is not transparent either.
		//
		// Upstream hooks only Program, FunctionExpression and FunctionDeclaration exits, so an
		// arrow body's scope is never visited and the declared-and-never-assigned judgment simply
		// does not run inside one. The assignment judgment still does, because it hangs off the
		// declarator and the assignment listeners rather than off a scope.
		//
		// Measured against the installed rule, with the alias `self`:
		//
		//	var f = () => { var self; };           clean       the scope is never visited
		//	var f = () => { var self = 42; };      reports     the declarator is still checked
		//	var f = () => { var context = this; }; reports     so is the capture
		//
		// Treating an arrow as transparent reports the first of those, and treating it as a scope
		// of its own reports it too, so both intuitive readings are wrong in the same direction.
		if current.Kind == ast.KindArrowFunction {
			checkConsistentThisAssignmentsOnly(ctx, current, aliases)
			return false
		}

		switch current.Kind {
		case ast.KindVariableDeclaration:
			declaration := current.AsVariableDeclaration()
			if declaration == nil || declaration.Name() == nil {
				break
			}
			name := declaration.Name()
			// A destructuring pattern captures properties of `this` rather than `this`, so no name
			// in it is an alias and upstream declines before reading one.
			if name.Kind != ast.KindIdentifier {
				break
			}
			if declaration.Initializer != nil {
				checkConsistentThisAssignment(ctx, current, name.Text(),
					declaration.Initializer, ast.KindEqualsToken, aliases)
			}
			if isDesignatedAlias(name.Text(), aliases) {
				declaredAliases = append(declaredAliases, current)
				// ANY initializer settles this declaration, not only a `this` one.
				//
				// Upstream's checkWasAssigned returns as soon as a def is a declarator with a
				// non-null init, without looking at what the init holds, so an initialized
				// declaration is never judged a second time here. That matters because a WRONG
				// initializer has already been reported by checkAssignment above: with the alias
				// `self`, `var self = 42` is one finding upstream and two under the intuitive
				// reading, which is what this port shipped until the corpus said otherwise.
				if declaration.Initializer != nil {
					rescuedAliases[name.Text()] = true
				}
			}

		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil || binary.Left == nil {
				break
			}
			if !isAssignmentOperator(binary.OperatorToken.Kind) {
				break
			}
			// Upstream reads `node.left.type === "Identifier"`, so a member or a pattern target is
			// not this rule's business.
			if binary.Left.Kind != ast.KindIdentifier {
				break
			}
			name := binary.Left.Text()
			checkConsistentThisAssignment(ctx, current, name, binary.Right,
				binary.OperatorToken.Kind, aliases)

			// A plain assignment of `this` written directly in this scope rescues a declaration.
			// "Directly" is upstream's `reference.from === scope`, measured to mean that no block,
			// switch, loop body, catch or `with` sits between the write and the scope.
			if isDesignatedAlias(name, aliases) &&
				binary.OperatorToken.Kind == ast.KindEqualsToken &&
				binary.Right != nil && binary.Right.Kind == ast.KindThisKeyword &&
				!liesInsideANestedEslintScope(current, scope) {
				rescuedAliases[name] = true
			}
		}

		current.ForEachChild(visit)
		return false
	}
	scope.ForEachChild(visit)

	for _, declaration := range declaredAliases {
		name := declaration.AsVariableDeclaration().Name().Text()
		if rescuedAliases[name] {
			continue
		}
		ctx.ReportNode(declaration, messageAliasNotAssignedToThis)
	}

	for _, nested := range nestedScopes {
		checkConsistentThisScope(ctx, nested, aliases)
	}
}

// isConsistentThisScope says whether a node opens a scope this rule judges separately.
//
// Upstream hooks `Program:exit`, `FunctionExpression:exit` and `FunctionDeclaration:exit`, and a
// method or accessor body is a function expression in that grammar, so those are here too. An arrow
// is deliberately absent: it has no `this` of its own, so a capture inside one belongs to the
// enclosing scope, and upstream does not hook it either.
func isConsistentThisScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor:
		return true
	}
	return false
}

// liesInsideANestedEslintScope says whether a write sits inside a scope eslint opens below `scope`.
//
// This reproduces `reference.from === scope` and it is a measurement rather than a derivation: the
// list below is exactly what eslint opens a scope for, established by driving the installed rule
// over each construct with and without a block. A loop appears here through its BODY rather than
// through the loop itself, which is why `for(var i=0;;) self = this;` is clean and
// `for(;;) { self = this; }` is not.
//
// The walk stops at `scope` because anything above it is a different question, and a nested function
// never reaches here at all, since the descent stops at one.
func liesInsideANestedEslintScope(write *ast.Node, scope *ast.Node) bool {
	for ancestor := write.Parent; ancestor != nil && ancestor != scope; ancestor = ancestor.Parent {
		switch ancestor.Kind {
		case ast.KindBlock:
			// A function's OWN body block is that function's scope rather than a nested one, and
			// skipping this exemption costs a real finding: with the alias `self`,
			// `var self; function f() { var self; self = this; }` reports the outer declaration
			// upstream and reported both here, because the inner write never rescued the inner
			// declaration it sits directly inside. Measured against the installed rule, which
			// reports at column 5 only.
			if ancestor.Parent == scope && isConsistentThisScope(scope) {
				continue
			}
			return true
		case ast.KindSwitchStatement, ast.KindCaseBlock,
			ast.KindCatchClause, ast.KindWithStatement:
			return true
		}
	}
	return false
}

// checkConsistentThisAssignment is upstream's `checkAssignment`, and it holds both judgments.
//
// A designated alias must receive `this` through a plain `=`, and anything else receiving `this` is
// a capture under the wrong name. The operator test is the half a port is most likely to drop:
// `self += this` assigns a `this` expression and is still a finding, because the alias ends up
// holding a concatenation.
func checkConsistentThisAssignment(ctx rule.Context, node *ast.Node, name string,
	value *ast.Node, operator ast.Kind, aliases []string) {
	valueIsThis := value != nil && value.Kind == ast.KindThisKeyword

	if isDesignatedAlias(name, aliases) {
		if !valueIsThis || operator != ast.KindEqualsToken {
			ctx.ReportNode(node, messageAliasNotAssignedToThis)
		}
		return
	}
	if valueIsThis {
		ctx.ReportNode(node, messageUnexpectedAlias)
	}
}

// isDesignatedAlias says whether a name is one of the configured aliases.
func isDesignatedAlias(name string, aliases []string) bool {
	for _, alias := range aliases {
		if name == alias {
			return true
		}
	}
	return false
}

// isAssignmentOperator says whether a binary operator writes to its left operand.
//
// Every compound form is here rather than only `=`, because upstream reaches the rule on any
// AssignmentExpression and then distinguishes the operator inside. Listing only `=` would make
// `self += this` silent, which is one of upstream's failing cases.
func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken, ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken, ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken, ast.KindCaretEqualsToken,
		ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// checkConsistentThisAssignmentsOnly runs the assignment half of the rule and not the scope half.
//
// This is what an arrow body gets. Upstream reaches every declarator and every assignment inside an
// arrow through listeners that do not depend on a scope, while its scope-exit listeners never fire
// for one, so the declared-and-never-assigned judgment is skipped there and only there. A nested
// function INSIDE an arrow is a scope again, which is why the descent resumes at one.
func checkConsistentThisAssignmentsOnly(ctx rule.Context, node *ast.Node, aliases []string) {
	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		if isConsistentThisScope(current) {
			checkConsistentThisScope(ctx, current, aliases)
			return false
		}

		switch current.Kind {
		case ast.KindVariableDeclaration:
			declaration := current.AsVariableDeclaration()
			if declaration == nil || declaration.Name() == nil ||
				declaration.Name().Kind != ast.KindIdentifier || declaration.Initializer == nil {
				break
			}
			checkConsistentThisAssignment(ctx, current, declaration.Name().Text(),
				declaration.Initializer, ast.KindEqualsToken, aliases)

		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil || binary.Left == nil ||
				!isAssignmentOperator(binary.OperatorToken.Kind) ||
				binary.Left.Kind != ast.KindIdentifier {
				break
			}
			checkConsistentThisAssignment(ctx, current, binary.Left.Text(), binary.Right,
				binary.OperatorToken.Kind, aliases)
		}

		current.ForEachChild(visit)
		return false
	}
	node.ForEachChild(visit)
}
