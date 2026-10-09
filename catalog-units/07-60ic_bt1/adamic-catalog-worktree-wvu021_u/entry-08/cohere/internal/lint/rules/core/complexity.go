package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ComplexityVariant selects which counting rule applies to a switch statement.
type ComplexityVariant string

const (
	// ComplexityClassic counts each `case` with a test, which is upstream's default.
	ComplexityClassic ComplexityVariant = "Classic"
	// ComplexityModified counts the whole switch as one, however many cases it has.
	ComplexityModified ComplexityVariant = "Modified"
)

// ComplexitySettings is the decoded option surface.
type ComplexitySettings struct {
	// Maximum is the highest complexity allowed before a function reports. Upstream's default is 20.
	Maximum int
	// Variant selects the switch-counting rule.
	Variant ComplexityVariant
}

// DefaultComplexitySettings is upstream's `defaultOptions: [20]` with the classic variant, spelled
// out because the zero value of the struct is not it: a Maximum of 0 reports every function.
func DefaultComplexitySettings() ComplexitySettings {
	return ComplexitySettings{Maximum: 20, Variant: ComplexityClassic}
}

// DecodeComplexityOptions reads the threshold and variant off the config.
//
// Hand rolled because upstream's schema is a `oneOf`: a bare integer, or an object carrying
// `maximum` or `max` plus an optional `variant`. The corpus configures the bare integer, both
// object spellings, and the modified variant.
//
// # The same JavaScript quirk max-depth has, and the same two opposite zero spellings
//
// Upstream's threshold handling is `hasOwn(option,"maximum") || hasOwn(option,"max")` gating
// `threshold = option.maximum || option.max`. The guard tests PRESENCE and the `||` tests
// TRUTHINESS, and they disagree exactly when a zero is written. See `DecodeMaxDepthOptions`, which
// carries the measured table for the identical expression; the outcome here is the same, and the
// `{"maximum": 0}` shape leaves the threshold `undefined` so every comparison against it is false
// and the rule goes silent.
//
// Empty input decodes to the defaults, which is what a rule configured as a bare "error" is handed.
func DecodeComplexityOptions(raw []byte) (any, error) {
	settings := DefaultComplexitySettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var maximum int
	if err := json.Unmarshal(raw, &maximum); err == nil {
		settings.Maximum = maximum
		return settings, nil
	}

	var object struct {
		Maximum *int    `json:"maximum"`
		Max     *int    `json:"max"`
		Variant *string `json:"variant"`
	}
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return settings, err
	}

	if object.Variant != nil && *object.Variant == "modified" {
		settings.Variant = ComplexityModified
	}

	// Upstream's `hasOwn` guard: neither key present leaves the default of 20.
	if object.Maximum == nil && object.Max == nil {
		return settings, nil
	}
	switch {
	case object.Maximum != nil && *object.Maximum != 0:
		settings.Maximum = *object.Maximum
	case object.Max != nil && *object.Max != 0:
		settings.Maximum = *object.Max
	case object.Max != nil:
		// `undefined || 0` and `0 || 0` both yield 0, so a `max` of zero wins.
		settings.Maximum = 0
	default:
		// Only `maximum` was written and it is zero: `0 || undefined` is undefined, and every
		// comparison against it is false. Modelled as an impossible threshold rather than a flag,
		// because unlike max-depth this rule has no other reason to decline a file.
		settings.Maximum = complexityNeverReports
	}
	return settings, nil
}

// complexityNeverReports is the threshold that reproduces upstream's `undefined`.
//
// A complexity is a count of syntactic constructs in one function, so it cannot approach this. Using
// a sentinel rather than a flag keeps the comparison in one place; see the decoder for why the
// `{"maximum": 0}` shape needs it at all.
const complexityNeverReports = int(^uint(0) >> 1)

// complexitySettingsFrom recovers the settings from whatever the config layer handed over.
func complexitySettingsFrom(options any) ComplexitySettings {
	if settings, ok := rule.OptionsAs[ComplexitySettings](options); ok {
		return settings
	}
	return DefaultComplexitySettings()
}

var messageComplexityComplex = rule.Message{
	Id: "complex",
	Description: "Every branch here is another path a reader has to hold open at once, and another " +
		"path a test has to cover. Split the decision out into named functions, or replace the " +
		"branching with a lookup, so each piece can be understood on its own.",
}

// Complexity reports a function whose cyclomatic complexity exceeds the configured threshold.
//
//	valid:   function a() { if (x) {} }                  (complexity 2, threshold 20)
//	invalid: function a() { if (x) {} if (y) {} }        (complexity 3, threshold 2)
//
// # This rule is audited No and stays unenabled, and the argument against it is worth keeping
//
// The audit rates it No at 326 violations. The standard also quotes the Tricorder paper naming
// cyclomatic complexity as failing the bar for a useful diagnostic, and that judgment is about
// ENABLING rather than about porting: a metric that fires on 326 sites with no fixer is a cleanup
// nobody has agreed to, and a number that correlates poorly with what makes code hard to read is a
// poor thing to gate a build on. The port exists so the decision can be revisited without redoing
// the work, and the rule is registered unenabled.
//
// # The metric is well defined against our AST, which was the open question
//
// Two things could have made agreement unreachable and neither does.
//
// The counting set is a fixed list of node kinds and every one exists here. There is no dataflow in
// it: cyclomatic complexity counts decision points syntactically, so a walk is the right tool.
//
// `onCodePathStart` and `onCodePathEnd` sound like they need ESLint's code path analysis and do
// not. The rule uses a code path only as a STACK of function-like scopes, and reads `codePath.origin`
// to filter to three of the four kinds ESLint produces -- `function`, `class-field-initializer` and
// `class-static-block`, excluding `program`. That is the same frame stack `max-depth` maintains, and
// it is maintained the same way here: one `KindSourceFile` listener recursing itself, since this
// tree has no exit hook.
//
// # What counts, and the two arms that are variant-dependent
//
// Upstream's set, reproduced arm for arm:
//
//	catch clause, conditional (?:), logical && || ??, for/for-in/for-of,
//	if, while, do-while, a parameter default, a logical assignment (&&= ||= ??=),
//	an optional member access (?.) and an optional call
//
// A `switch` is counted differently by variant, and the two arms are exclusive: classic counts each
// `case` that carries a test, so `default` never counts; modified counts the switch itself as one
// regardless of how many cases it has. Both are in the corpus.
//
// # The message interpolates a NAME, and the name is most of the port
//
// Upstream builds it with `getFunctionNameWithKind` -- 77 lines assembling `static`, `private`,
// `async`, `generator`, a kind word, and a quoted name -- then upper-cases the first letter. The
// corpus exercises 14 distinct renderings across 103 findings, from `Function 'a'` through
// `Static method 'static'` and `Class field initializer`, so the name is what a message fixture
// actually tests rather than an incidental detail.
//
// It is SIMPLER here than upstream, because of a parser difference worth stating: in ESTree a
// method is a `MethodDefinition` wrapping a `FunctionExpression`, so upstream has to read modifiers
// off the parent and the kind word off `parent.kind`. Measured, our parser has no wrapper at all --
// `KindMethodDeclaration` is both the member and the function -- so the whole parent dispatch
// collapses into the node's own kind and its modifier list.
//
// # The span is the function HEAD, not the function
//
// Upstream reports `loc: getFunctionHeadLoc(node, sourceCode)`: from the declaration's start to the
// opening parenthesis of its parameters, or the `=>` token for an arrow, or the `static` keyword for
// a static block. A class field initializer reports on the initializer expression instead. Measured
// from the oracle, the spans are `function a`, `=>`, `static static`, `c || d || e` -- never the
// whole function body.
//
// # No fix
//
// Upstream offers none and none is possible: reducing complexity means restructuring, and only the
// author knows how.
var Complexity = rule.Rule{
	Name: "complexity",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := complexitySettingsFrom(options)

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// One walk maintaining a frame stack, for the same reason max-depth does it: the
				// rule needs a counter that belongs to the innermost enclosing function, and the
				// listener walk is pre-order with no exit hook to pop on.
				walker := &complexityWalker{ctx: ctx, settings: settings}
				walker.walk(node)
			},
		}
	},
}

// complexityFrame is one function-like scope and the complexity accumulated inside it.
type complexityFrame struct {
	// node is the scope's own node, which the report anchors and names itself from.
	node *ast.Node
	// count starts at 1, which is upstream's `complexities.push(1)`: a function with no branches
	// has one path through it.
	count int
}

// complexityWalker carries the frame stack across the recursive walk.
type complexityWalker struct {
	ctx      rule.Context
	settings ComplexitySettings
	frames   []complexityFrame
}

func (w *complexityWalker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	// A class field's code path covers its INITIALIZER only, not the whole member, and the
	// difference is visible: a computed key is evaluated where the class is defined, so its
	// branches belong to the enclosing function rather than to the field. Measured against the
	// installed rule at max 1, `function foo() { class C { [x || y] = a; } }` reports on `foo`
	// with a complexity of 2, which is only true if the key counted outside the field's frame.
	//
	// So a property declaration is walked in two pieces rather than one: everything but the
	// initializer stays in the enclosing frame, and the initializer gets its own.
	if node.Kind == ast.KindPropertyDeclaration {
		declaration := node.AsPropertyDeclaration()
		if declaration != nil && declaration.Initializer != nil {
			node.ForEachChild(func(child *ast.Node) bool {
				if child == declaration.Initializer {
					return false
				}
				w.walk(child)
				return false
			})
			w.frames = append(w.frames, complexityFrame{node: node, count: 1})
			w.walk(declaration.Initializer)
			frame := w.frames[len(w.frames)-1]
			w.frames = w.frames[:len(w.frames)-1]
			w.reportIfOverThreshold(frame)
			return
		}
	}

	pushed := false
	if complexityStartsAFrame(node) {
		w.frames = append(w.frames, complexityFrame{node: node, count: 1})
		pushed = true
	} else {
		w.countIfDecisionPoint(node)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child)
		return false
	})

	if pushed {
		frame := w.frames[len(w.frames)-1]
		w.frames = w.frames[:len(w.frames)-1]
		w.reportIfOverThreshold(frame)
	}
}

// countIfDecisionPoint is upstream's `increaseComplexity` listener set.
//
// A node that starts a frame is never also counted, matching upstream: a nested function's own
// complexity belongs to its own code path rather than to its parent's.
func (w *complexityWalker) countIfDecisionPoint(node *ast.Node) {
	if len(w.frames) == 0 {
		// Upstream still counts into the `program` code path here; it simply never reports it,
		// because `onCodePathEnd` filters origins. Nothing to do.
		return
	}

	counts := false
	switch node.Kind {
	case ast.KindCatchClause, ast.KindConditionalExpression,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
		ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement:
		counts = true

	// Upstream's LogicalExpression and its logical-assignment arm are one kind here, so both are
	// answered by looking at the operator.
	case ast.KindBinaryExpression:
		counts = complexityIsLogicalOperator(node)

	// Upstream's AssignmentPattern: a parameter default is a branch, because the parameter may or
	// may not be supplied. Our parser models it as an Initializer on the parameter rather than as
	// its own node, so the arm is on the parameter and gated on having one.
	case ast.KindParameter:
		declaration := node.AsParameterDeclaration()
		counts = declaration != nil && declaration.Initializer != nil

	// A destructuring default, which upstream also reaches through AssignmentPattern.
	case ast.KindBindingElement:
		element := node.AsBindingElement()
		counts = element != nil && element.Initializer != nil

	// Upstream's optional MemberExpression and CallExpression arms. Our parser marks the optional
	// chain with a token on the access rather than a boolean on the node.
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		counts = access != nil && access.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		counts = access != nil && access.QuestionDotToken != nil
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		counts = call != nil && call.QuestionDotToken != nil

	// The two variant-dependent arms, which are exclusive.
	case ast.KindCaseClause:
		// Upstream's `SwitchCase[test]`: only a case carrying a test counts, so `default` never
		// does -- and our parser gives `default` its own kind, so this arm cannot see one.
		counts = w.settings.Variant != ComplexityModified
	case ast.KindSwitchStatement:
		counts = w.settings.Variant == ComplexityModified
	}

	if counts {
		w.frames[len(w.frames)-1].count++
	}
}

// reportIfOverThreshold is upstream's `onCodePathEnd`.
func (w *complexityWalker) reportIfOverThreshold(frame complexityFrame) {
	if frame.count <= w.settings.Maximum {
		return
	}
	name := complexityNameOf(w.ctx, frame.node)
	w.ctx.Report(rule.Diagnostic{
		Range:      complexityHeadRange(w.ctx, frame.node),
		SourceFile: w.ctx.SourceFile,
		Message: rule.Message{
			Id: messageComplexityComplex.Id,
			Description: fmt.Sprintf("%s has a complexity of %d. Maximum allowed is %d. %s",
				name, frame.count, w.settings.Maximum, messageComplexityComplex.Description),
		},
	})
}

// complexityStartsAFrame answers which nodes open a code path this rule reports on.
//
// Upstream's `onCodePathEnd` filters `codePath.origin` to `function`,
// `class-field-initializer` and `class-static-block`, excluding `program`. So the file itself is
// deliberately NOT a frame here: a top-level `if` belongs to the program's path, which is never
// reported, and giving the file a frame would report the whole module as one function.
//
// A property declaration is a frame only when it HAS an initializer, because that is what makes it
// a class field initializer rather than a bare declaration.
func complexityStartsAFrame(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	// A property declaration is deliberately absent: it is handled by its own arm in `walk`, which
	// has to split the member into a key part and an initializer part. See the comment there.
	return false
}

// complexityIsLogicalOperator answers upstream's LogicalExpression and its
// `isLogicalAssignmentOperator` arm at once, because our parser gives both one kind.
func complexityIsLogicalOperator(node *ast.Node) bool {
	binary := node.AsBinaryExpression()
	if binary == nil {
		return false
	}
	switch binary.OperatorToken.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// complexityNameOf is upstream's `getFunctionNameWithKind` followed by `upperCaseFirst`.
//
// Simpler than upstream's 77 lines because our parser has no MethodDefinition wrapper: the member
// and the function are one node, so the modifiers and the kind word come off the node itself rather
// than off a parent. The 14 renderings the corpus exercises are all produced here.
func complexityNameOf(ctx rule.Context, node *ast.Node) string {
	switch node.Kind {
	case ast.KindClassStaticBlockDeclaration:
		return "Class static block"
	case ast.KindPropertyDeclaration:
		return "Class field initializer"
	case ast.KindConstructor:
		// Upstream returns "constructor" before assembling any other token, so a static or async
		// modifier never appears on it.
		return "Constructor"
	}

	var tokens []string
	if complexityHasModifier(node, ast.KindStaticKeyword) {
		tokens = append(tokens, "static")
	}
	if name := node.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
		tokens = append(tokens, "private")
	}
	if complexityHasModifier(node, ast.KindAsyncKeyword) {
		tokens = append(tokens, "async")
	}
	if complexityIsGenerator(node) {
		tokens = append(tokens, "generator")
	}

	// The kind word. Upstream reads it off the PARENT when the function is a property's value or a
	// class field's value, so what the function is assigned to wins over what it is syntactically:
	// measured, `class C { x = () => a||b||c; }` renders "Method 'x'" rather than "Arrow function",
	// and `{ b: (a) => ... }` renders "Method 'b'". A count-only fixture cannot see either, because
	// this rule has exactly one message id.
	if complexityNamingOwnerOf(node) != nil {
		// Always "method". Upstream's Property and PropertyDefinition branch has no getter or
		// setter arm either, and there is nothing for one to match here: the owner is a property
		// assignment or a class field, never an accessor, because only an arrow or a function
		// expression borrows a name at all. A first draft carried accessor arms in this branch and
		// a mutation gutting them survived, correctly -- they were unreachable.
		tokens = append(tokens, "method")
	} else {
		switch node.Kind {
		case ast.KindGetAccessor:
			tokens = append(tokens, "getter")
		case ast.KindSetAccessor:
			tokens = append(tokens, "setter")
		case ast.KindMethodDeclaration:
			tokens = append(tokens, "method")
		case ast.KindArrowFunction:
			tokens = append(tokens, "arrow", "function")
		default:
			tokens = append(tokens, "function")
		}
	}

	// A private name is pushed UNQUOTED, where every other name is quoted. Upstream does this by
	// pushing `` `#${key.name}` `` directly rather than going through its quoted-name branch, so
	// `class C { #p() {} }` renders `Private method #p` and not `Private method '#p'`. Measured
	// against the installed rule; no corpus case reaches it, because upstream never writes a
	// private member whose complexity crosses a threshold.
	if name := node.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
		tokens = append(tokens, name.Text())
	} else if name := complexityReadableNameOf(node); name != "" {
		tokens = append(tokens, "'"+name+"'")
	}

	rendered := strings.Join(tokens, " ")
	if rendered == "" {
		return rendered
	}
	return strings.ToUpper(rendered[:1]) + rendered[1:]
}

// complexityReadableNameOf reads the name upstream quotes, or "" when the function is anonymous.
//
// A function expression assigned to something borrows that binding's name upstream, through
// `getStaticPropertyName` on the parent for a property and the variable's own name otherwise. The
// corpus exercises the property form: `var o = { c: function() {} }` renders `Method 'c'`.
func complexityReadableNameOf(node *ast.Node) string {
	// The OWNER's name wins when there is one, which is upstream's ordering: it reads
	// `getStaticPropertyName(parent)` first and consults `node.id` only as a fallback.
	if owner := complexityNamingOwnerOf(node); owner != nil && owner.Name() != nil {
		if owner.Name().Kind == ast.KindPrivateIdentifier {
			return owner.Name().Text()
		}
		if text, ok := property.Name(owner.Name(), property.Static); ok {
			return text
		}
		// A computed or otherwise non-static property name, where upstream falls through to the
		// function's own name.
	}

	if name := node.Name(); name != nil {
		if name.Kind == ast.KindPrivateIdentifier {
			return name.Text()
		}
		if text, ok := property.Name(name, property.Static); ok {
			return text
		}
		if name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral {
			return name.Text()
		}
	}
	return ""
}

// complexityNamingOwnerOf returns the property or class member whose name this function borrows,
// or nil when the function names itself.
//
// Upstream's `getFunctionNameWithKind` reads `parent.type === "Property"` and
// `"PropertyDefinition"` for both the kind word and the name, which is how an arrow assigned to a
// field becomes "Method 'x'". A function that has its OWN name, such as a method declaration or a
// named function expression, is not borrowing and returns nil here.
func complexityNamingOwnerOf(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression {
		return nil
	}
	// Deliberately NOT gated on the function being anonymous. Upstream tries the property name
	// FIRST and falls back to the function's own `id` only when the property name is not static, so
	// a named function expression in a property position is named for the PROPERTY:
	// `{ d: function named() {} }` renders `Method 'd'`, measured against the installed rule. An
	// earlier draft read the function's own name first and rendered `Method 'named'`.
	parent := node.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		if assignment := parent.AsPropertyAssignment(); assignment != nil &&
			assignment.Initializer == node {
			return parent
		}
	case ast.KindPropertyDeclaration:
		if declaration := parent.AsPropertyDeclaration(); declaration != nil &&
			declaration.Initializer == node {
			return parent
		}
	}
	return nil
}

// complexityHasModifier answers whether a declaration carries a modifier keyword.
func complexityHasModifier(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}

// complexityIsGenerator answers upstream's `node.generator`.
func complexityIsGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// complexityHeadRange is upstream's `getFunctionHeadLoc`.
//
// From the declaration's start to the opening parenthesis of its parameters, so the reported span
// is the function's HEAD rather than its body. Two shapes differ: an arrow reports on its `=>`
// token, and a class field initializer reports on the initializer expression.
func complexityHeadRange(ctx rule.Context, node *ast.Node) core.TextRange {
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		// Upstream reports a class field initializer on `node.loc`, where `node` is the
		// initializer expression the code path belongs to rather than the whole member.
		if declaration := node.AsPropertyDeclaration(); declaration != nil && declaration.Initializer != nil {
			return rule.TokenRange(ctx.SourceFile, declaration.Initializer)
		}
		return rule.TokenRange(ctx.SourceFile, node)

	case ast.KindClassStaticBlockDeclaration:
		// Upstream's `getFirstToken(node).loc`, which is the `static` keyword alone.
		return scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())

	case ast.KindArrowFunction:
		// An arrow that borrows a property's name also borrows its span, so this arm is only
		// reached by a free-standing arrow. Upstream's ordering is the same: its Property and
		// PropertyDefinition branch is tested BEFORE its ArrowFunctionExpression branch, so an
		// arrow in a property position never reaches the arrow arm at all.
		if complexityNamingOwnerOf(node) == nil {
			if arrow := node.AsArrowFunction(); arrow != nil && arrow.EqualsGreaterThanToken != nil {
				return scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, arrow.EqualsGreaterThanToken.Pos()).
					WithEnd(arrow.EqualsGreaterThanToken.End())
			}
			return rule.TokenRange(ctx.SourceFile, node)
		}
	}

	// Everything else: start of the declaration through to the opening paren of the parameters.
	//
	// The start is the OWNER's when the function borrows a name, which is upstream's
	// `start = parent.loc.start` for its Property and PropertyDefinition branch. Measured, that is
	// what makes `{ c: function (a) {...} }` span `c: function ` rather than `function ` -- the
	// property name is part of the head.
	anchor := node
	if owner := complexityNamingOwnerOf(node); owner != nil {
		anchor = owner
	}
	start := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, anchor.Pos()).Pos()
	end := complexityParameterListStart(ctx, node)
	if end <= start {
		return rule.TokenRange(ctx.SourceFile, node)
	}
	return core.NewTextRange(start, end)
}

// complexityParameterListStart finds the offset of the `(` that opens the parameter list, which is
// upstream's `getOpeningParenOfParams`.
//
// Derived from the source text rather than from a token list, because the parameters themselves may
// be absent and a type parameter list can sit between the name and the paren.
func complexityParameterListStart(ctx rule.Context, node *ast.Node) int {
	text := ctx.SourceFile.Text()
	body := complexityBodyOf(node)
	limit := node.End()
	if body != nil {
		limit = body.Pos()
	}
	if limit > len(text) {
		limit = len(text)
	}

	depth := 0
	for index := node.Pos(); index < limit; index++ {
		switch text[index] {
		case '<':
			depth++
		case '>':
			if depth > 0 {
				depth--
			}
		case '(':
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

// complexityBodyOf reads a function-like node's body, or nil when it has none.
func complexityBodyOf(node *ast.Node) *ast.Node {
	if body := node.Body(); body != nil {
		return body
	}
	return nil
}
