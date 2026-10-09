package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferForOf flags a counted `for` loop whose index is only ever used to index one array.
//
//	valid:   for (let i = 0; i < arr.length; i++) { doMath(i); }      the index itself is used
//	valid:   for (let i = 0; i < arr.length; i++) { arr[i] = 0; }     the element is assigned
//	valid:   for (let i = 0, len = arr.length; i < len; i++) {}       a second declarator
//	valid:   for (let i = 1; i < arr.length; i++) {}                  does not start at zero
//	invalid: for (let a = 0; a < arr.length; a++) { console.log(arr[a]); }
//	invalid: for (let i = 0; i < this.item.length; ++i) { this.item[i]; }
//
// A counted loop over an array states the mechanism where `for-of` states the intent, and the
// mechanism is where the off-by-one lives. The rule fires only when the index is provably doing
// nothing but subscripting, which is the case where the two forms mean the same thing.
//
// # What upstream checks, in order, and why every part is load bearing
//
//	the initializer is one non-const declarator, initialized to the literal 0
//	the test is `index < <something>.length`
//	the update increments the index by one, in any of four spellings
//	every reference to the index inside the body is `<that same something>[index]`,
//	  the receiver is not `this` on its own, and the access is not being assigned to
//
// The last clause is the one that makes the rule sound rather than merely plausible, and it carries
// most of upstream's corpus: thirteen of its fifty-two passing cases are an indexed element being
// written rather than read, where a for-of rewrite would change what the code does rather than how
// it reads.
//
// # The receiver is compared as TEXT, not as a symbol
//
// Upstream compares `getText(node.object)` against `getText(arrayExpression)`. That is a string
// comparison and it is reproduced as one, deliberately. Comparing symbols would be a different rule:
// it would report `for (let i = 0; i < a.length; i++) { b[i]; }` where `a` and `b` are aliases of one
// array, which upstream leaves alone, and it would decline where two spellings of the same object
// differ by whitespace. The rule is a readability judgment about what the source SAYS, so the source
// text is the right thing to compare.
//
// # Where our tree spells things differently
//
// Upstream reads an ESTree, which is flatter than ours in three places that matter here:
//
//	upstream                       ours
//	UpdateExpression `i++`         KindPostfixUnaryExpression
//	UpdateExpression `++i`         KindPrefixUnaryExpression
//	AssignmentExpression `i += 1`  KindBinaryExpression, operator token
//	VariableDeclaration.kind       a NodeFlags bit on the declaration list
//	MemberExpression, computed     KindElementAccessExpression, its own kind
//	MemberExpression, dotted       KindPropertyAccessExpression, its own kind
//
// The last pair is the one that changes the shape of the code rather than only its spelling.
// Upstream writes one condition testing `node.property === id`, which distinguishes `arr[i]` from
// `obj.i` because both are the same node type there. Ours cannot confuse them, so the equivalent
// question is simply whether the parent is an element access whose argument is this identifier.
//
// # This rule needs the checker, and it needs it for exactly one question
//
// Upstream asks scope analysis for the index variable's references. We have no reference index, so
// the body is walked and every identifier resolved with `GetSymbolAtLocation`, keeping the ones whose
// symbol carries the loop's own declarator among its declarations. That is the technique in
// `no_class_assign.go`, and it is what makes the shadowing case work: upstream's own corpus has a
// loop nested inside a loop where the inner index has the SAME NAME as the outer, and a port matching
// on the name would attribute the inner references to the outer loop and go silent on both.
//
// The loop over `symbol.Declarations` is deliberate rather than an index into the first. The question
// here is "does any declaration of this symbol belong to this loop", so more declarations only mean
// more chances to match, and a `var` index legitimately merges across declarations.
//
// # Cost
//
// One listener on a rare anchor, and the body walk runs only after the initializer, test and update
// have all matched, which is a syntactic filter that declines almost every for loop in a real tree
// before a single type question is asked.
var PreferForOf = rule.Rule{
	Name: "@typescript-eslint/prefer-for-of",

	// Resolving a body identifier back to the loop's own declarator is the only way to tell a
	// reference to this index from a reference to a shadowing one with the same name.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindForStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				statement := node.AsForStatement()

				declarator := singleNonConstDeclarator(statement.Initializer)
				if declarator == nil {
					return
				}
				declaration := declarator.AsVariableDeclaration()
				if !isZeroInitialized(declaration) {
					return
				}
				indexName := declaration.Name()
				if indexName == nil || indexName.Kind != ast.KindIdentifier {
					return
				}

				arrayExpression := arrayOfLessThanLengthTest(statement.Condition,
					indexName.AsIdentifier().Text)
				if arrayExpression == nil {
					return
				}

				if !isIncrementOfIndex(statement.Incrementor, indexName.AsIdentifier().Text) {
					return
				}

				if !isIndexOnlyUsedToSubscript(ctx, statement.Statement, declarator, arrayExpression) {
					return
				}

				ctx.ReportNode(node, preferForOfMessage)
			},
		}
	},
}

var preferForOfMessage = rule.Message{
	Id:          "preferForOf",
	Description: "Expected a `for-of` loop instead of a `for` loop with this simple iteration.",
}

// singleNonConstDeclarator is upstream's `isSingleVariableDeclaration` plus its declarator lookup.
//
// Upstream reads `node.kind !== 'const'` off the declaration itself. Our parser puts that on the
// declaration LIST as a flag rather than as a field, so the test is a flag test. `var` is flagless
// and passes, which matches upstream reading a kind that is neither const nor absent.
func singleNonConstDeclarator(initializer *ast.Node) *ast.Node {
	if initializer == nil || initializer.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	list := initializer.AsVariableDeclarationList()
	if list.Flags&ast.NodeFlagsConst != 0 {
		return nil
	}
	if list.Declarations == nil || len(list.Declarations.Nodes) != 1 {
		return nil
	}
	return list.Declarations.Nodes[0]
}

// isZeroInitialized is upstream's `isZeroInitialized`.
//
// It requires the literal `0` rather than any expression evaluating to zero, which is upstream's
// `node.value === 0` on a Literal. A `NumericLiteral`'s `.Text` here is the canonical rendering of the
// double, so `0x0` and `0.0` both read as "0" and are accepted the same way upstream accepts them.
func isZeroInitialized(declaration *ast.VariableDeclaration) bool {
	return declaration.Initializer != nil && isNumericLiteralWithText(declaration.Initializer, "0")
}

// isNumericLiteralWithText is upstream's `isLiteral(node, value)`.
//
// Upstream compares the parsed VALUE, so any spelling of the number matches. Our parser has already
// done that normalization: a numeric literal's `.Text` holds the canonical rendering of the double it
// denotes, so `0x0`, `0.0` and `0e0` all arrive as "0", and comparing that text is comparing the
// value rather than the spelling.
func isNumericLiteralWithText(node *ast.Node, text string) bool {
	return node != nil && node.Kind == ast.KindNumericLiteral &&
		node.AsNumericLiteral().Text == text
}

// isIdentifierWithText is upstream's `isMatchingIdentifier`.
func isIdentifierWithText(node *ast.Node, name string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == name
}

// arrayOfLessThanLengthTest is upstream's `isLessThanLengthExpression`.
//
// It answers the receiver of the `.length` when the test reads `index < <receiver>.length`, and nil
// otherwise. The receiver is what every body reference then has to match textually.
//
// An OPTIONAL access disqualifies, and that is not an accident of the node kind. `i < arr?.length` is
// silent upstream because its parser gives an optional member expression a flag this rule does not
// look for, so the property is reached through a shape the equality test rejects. Ours gives an
// optional chain the same kind with a question-dot token, so the token is what has to be checked.
// Measured: four of upstream's passing cases are exactly this, differing from a reporting case only
// by the `?.`.
func arrayOfLessThanLengthTest(condition *ast.Node, indexName string) *ast.Node {
	if condition == nil || condition.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := condition.AsBinaryExpression()
	if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindLessThanToken {
		return nil
	}
	if !isIdentifierWithText(binary.Left, indexName) {
		return nil
	}
	if binary.Right == nil || binary.Right.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	access := binary.Right.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil {
		return nil
	}
	if !isIdentifierWithText(access.Name(), "length") {
		return nil
	}
	return access.Expression
}

// isIncrementOfIndex is upstream's `isIncrement`.
//
// Four spellings, which upstream reaches through two node types and ours through three:
//
//	i++ and ++i    postfix and prefix unary here, one UpdateExpression upstream
//	i += 1         a binary expression with a compound assignment token
//	i = i + 1      and i = 1 + i, an assignment whose right side adds one either way
func isIncrementOfIndex(update *ast.Node, indexName string) bool {
	if update == nil {
		return false
	}

	switch update.Kind {
	case ast.KindPostfixUnaryExpression:
		unary := update.AsPostfixUnaryExpression()
		return unary.Operator == ast.KindPlusPlusToken &&
			isIdentifierWithText(unary.Operand, indexName)

	case ast.KindPrefixUnaryExpression:
		unary := update.AsPrefixUnaryExpression()
		return unary.Operator == ast.KindPlusPlusToken &&
			isIdentifierWithText(unary.Operand, indexName)

	case ast.KindBinaryExpression:
		binary := update.AsBinaryExpression()
		if binary.OperatorToken == nil || !isIdentifierWithText(binary.Left, indexName) {
			return false
		}

		if binary.OperatorToken.Kind == ast.KindPlusEqualsToken {
			return isNumericLiteralWithText(binary.Right, "1")
		}

		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			if binary.Right == nil || binary.Right.Kind != ast.KindBinaryExpression {
				return false
			}
			sum := binary.Right.AsBinaryExpression()
			if sum.OperatorToken == nil || sum.OperatorToken.Kind != ast.KindPlusToken {
				return false
			}
			return (isIdentifierWithText(sum.Left, indexName) &&
				isNumericLiteralWithText(sum.Right, "1")) ||
				(isNumericLiteralWithText(sum.Left, "1") &&
					isIdentifierWithText(sum.Right, indexName))
		}
	}
	return false
}

// isIndexOnlyUsedToSubscript is upstream's `isIndexOnlyUsedWithArray`.
//
// Upstream iterates the index variable's resolved references and requires every one inside the body
// to be the argument of an access on the same array. We have no reference index, so the body is
// walked and each identifier resolved back to the loop's own declarator, which is the technique
// `no_class_assign.go` uses and which is what makes the shadowing case work.
//
// Upstream's `!contains(body, id)` clause has no counterpart here, and its absence is not a
// simplification. That clause exists because upstream is handed EVERY reference to the variable,
// including the ones in the loop header, and has to skip them. Walking the body reaches only the
// references inside it, so there is nothing to skip.
func isIndexOnlyUsedToSubscript(
	ctx rule.Context,
	body *ast.Node,
	declarator *ast.Node,
	arrayExpression *ast.Node,
) bool {
	if body == nil {
		return true
	}

	arrayRange := rule.TokenRange(ctx.SourceFile, arrayExpression)
	arrayText := ctx.SourceFile.Text()[arrayRange.Pos():arrayRange.End()]

	onlySubscripts := true
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || !onlySubscripts {
			return
		}
		if node.Kind == ast.KindIdentifier && resolvesToDeclarator(ctx, node, declarator) {
			if !isSubscriptOfArray(ctx, node, arrayText) {
				onlySubscripts = false
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return onlySubscripts
}

// resolvesToDeclarator answers whether an identifier binds to the loop's own index declaration.
//
// The loop over `symbol.Declarations` rather than an index into the first is the question this rule
// is asking: "does ANY declaration of this symbol belong to this loop". Node identity rather than
// kind comparison, because a kind comparison passes every corpus case and fails on exactly the
// shadowing shape upstream tests, where both declarations are the same kind.
func resolvesToDeclarator(ctx rule.Context, identifier *ast.Node, declarator *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == declarator {
			return true
		}
	}
	return false
}

// isSubscriptOfArray answers upstream's condition on one reference.
//
// The identifier has to be the SUBSCRIPT of an element access, the receiver has to read exactly like
// the array in the loop test, the receiver must not be a bare `this`, and the access must not be
// being assigned to.
//
// The bare-`this` exclusion looks arbitrary and is upstream's: `for (let i = 0; i < test.length; ++i)
// { this[i]; }` is a passing case there. `this.item[i]` against `this.item.length` reports, so what is
// excluded is `this` standing alone as the whole receiver rather than `this` appearing in it.
func isSubscriptOfArray(ctx rule.Context, identifier *ast.Node, arrayText string) bool {
	parent := identifier.Parent
	if parent == nil || parent.Kind != ast.KindElementAccessExpression {
		return false
	}
	access := parent.AsElementAccessExpression()

	// Upstream's `node.property === id`. Ours cannot confuse a subscript with a dotted property,
	// because the two are different node kinds, so this is the whole of that test.
	//
	// It is SUBSUMED by the receiver comparison below, and kept anyway. A mutation removing it
	// survives every fixture, including one written specifically to kill it, which is what sent this
	// to an argument rather than to a third fixture. For this line to decide anything the identifier
	// would have to be the RECEIVER of the access while the receiver's text still equalled the
	// array's, which means the index and the array are spelled the same; and a loop whose test reads
	// `i < i.length` cannot have `i` resolve to its own zero-initialized declarator and also be the
	// array. So no input separates the two versions, and `i[arr]` is declined below by text rather
	// than here by position.
	//
	// Written out because it is upstream's, and because it declines in one pointer comparison what
	// otherwise costs two range lookups and a string compare on every body reference.
	if access.ArgumentExpression != identifier {
		return false
	}
	if access.Expression == nil || access.Expression.Kind == ast.KindThisKeyword {
		return false
	}

	receiverRange := rule.TokenRange(ctx.SourceFile, access.Expression)
	if ctx.SourceFile.Text()[receiverRange.Pos():receiverRange.End()] != arrayText {
		return false
	}

	return !isAssignedTo(parent)
}

// isAssignedTo is upstream's `isAssignee`.
//
// Every shape where the indexed element is being WRITTEN rather than read, which is what makes a
// for-of rewrite wrong rather than merely different. Recursive through the assertion wrappers,
// because `(arr[i] as number)++` writes through the assertion.
func isAssignedTo(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	// `arr[i] = 1`, `arr[i] += 1`, and every other compound assignment. Our parser files all of
	// them as one binary expression, so the operator has to be recognized as an assignment rather
	// than read off a distinct node type the way upstream reads AssignmentExpression.
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken != nil &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
			binary.Left == node

	// `delete arr[i]`. Its own node kind here rather than a unary expression with an operator.
	case ast.KindDeleteExpression:
		return parent.AsDeleteExpression().Expression == node

	// `arr[i]++` and `--arr[i]`.
	case ast.KindPostfixUnaryExpression:
		return parent.AsPostfixUnaryExpression().Operand == node
	case ast.KindPrefixUnaryExpression:
		operator := parent.AsPrefixUnaryExpression().Operator
		if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
			return false
		}
		return parent.AsPrefixUnaryExpression().Operand == node

	// `[arr[i]] = [0]` and `[...arr[i]] = [0]`. Upstream reads an ArrayPattern and a RestElement,
	// which its parser produces because it knows the array literal is a destructuring target. Ours
	// parses the same source as an array LITERAL with a spread element until the assignment says
	// otherwise, so the kinds differ while the question is the same: is this element position part
	// of something being assigned to.
	case ast.KindArrayLiteralExpression:
		return isAssignedTo(parent)
	case ast.KindSpreadElement:
		return isAssignedTo(parent)

	// `({ foo: arr[i] }) = { foo: 0 }`, and NOT `({ foo: arr[i] } = { foo: 0 })`.
	//
	// The asymmetry is upstream's and it looks like a defect, so it is written down rather than
	// corrected. Upstream requires the enclosing object be an `ObjectExpression`, and its parser
	// gives a real destructuring target the distinct type `ObjectPattern` instead, so this arm fires
	// only on the form where the parentheses sit INSIDE the assignment and the object is therefore
	// not a pattern at all. Measured against the installed build: with the parentheses around the
	// whole assignment the loop REPORTS, and with them around the object alone it is silent, which is
	// the opposite of what the two shapes do at runtime.
	//
	// Our parser has no ObjectPattern, so both spellings arrive as an object literal and this arm
	// would fire on both. The intervening parenthesis is what separates them here, and that is what
	// the check below reads. Upstream's own corpus contains only the parenthesized-object form, so
	// nothing imported can see the difference; it was found by measuring the other spelling.
	case ast.KindPropertyAssignment:
		// Upstream's `parent.value === node`, and it is SUBSUMED here rather than deciding anything.
		//
		// A mutation removing it survives every fixture, including a computed-key case written to
		// kill it. Four parses answer why: a property assignment has exactly two children, its name
		// and its initializer, and an element access can only ever be the second. A computed key
		// wraps it in a KindComputedPropertyName first, so the recursion stops there and never
		// reaches this arm, and a shorthand property is a different node kind altogether.
		//
		// Kept because it is upstream's and because it is the line that would matter if a future
		// parse shape ever put an expression somewhere else in this node.
		if parent.AsPropertyAssignment().Initializer != node {
			return false
		}
		// The object-literal kind test is upstream's and it is REDUNDANT here, kept for the reason
		// below rather than because it decides anything.
		//
		// A mutation removing it survives every fixture, including one written to kill it. Fifteen
		// parses answer why: a KindPropertyAssignment is produced only inside an object literal, and
		// the five malformed inputs among them, a truncated literal, a missing value, a missing key,
		// and two shapes that look like one, either keep that parent or produce no property
		// assignment at all. A type literal, a class body and a destructuring pattern all use other
		// node kinds, so nothing else can reach this arm.
		//
		// It stays because a nil parent still has to be declined before the recursion, and because
		// the invariant is the parser's rather than this rule's: if error recovery ever synthesizes a
		// property assignment somewhere else, this is the line that keeps the rule from recursing
		// into it.
		if parent.Parent == nil || parent.Parent.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		// Upstream's ObjectExpression requirement, expressed through the parenthesis our parser keeps
		// and its parser drops. An object literal sitting DIRECTLY on the left of an assignment is a
		// destructuring pattern, which upstream types as ObjectPattern and this arm therefore
		// declines; one reached through a parenthesis is an ordinary object expression and is where
		// upstream fires.
		if objectLiteralIsADestructuringTarget(parent.Parent) {
			return false
		}
		return isAssignedTo(parent.Parent)

	// `(arr[i] as number)++`, `[...arr[i]!] = [0]`, and the rest. The write happens through the
	// wrapper, so the question passes through it.
	case ast.KindNonNullExpression,
		ast.KindAsExpression,
		ast.KindTypeAssertionExpression,
		ast.KindSatisfiesExpression:
		return isAssignedTo(parent)

	// A parenthesis is a real node here and is folded away by upstream's parser, so upstream never
	// sees one and its recursion has no arm for it. Without this, `(arr[i] as number)++` reaches the
	// parenthesis and stops, and four of upstream's own passing cases report.
	case ast.KindParenthesizedExpression:
		return isAssignedTo(parent)
	}

	return false
}

// objectLiteralIsADestructuringTarget answers whether an object literal is upstream's ObjectPattern
// rather than its ObjectExpression.
//
// Upstream's parser makes this a node type; ours does not, so it is a position question: an object
// literal written directly on the left of an assignment is a pattern, and one reached through a
// parenthesis is an expression. The distinction only matters because upstream's isAssignee tests for
// ObjectExpression, which makes it fire on the parenthesized form and not on the real destructuring.
//
// ANY assignment operator makes it a pattern, not only a plain equals. That was wrong here first: a
// mutation widening this test survived, the fixture written to kill it disagreed with upstream, and
// running upstream's own parser on `({ foo: a[i] } += x)` settled it, which reports ObjectPattern for
// a compound assignment exactly as it does for a plain one. The narrow version made this rule go
// silent on a loop upstream reports.
func objectLiteralIsADestructuringTarget(objectLiteral *ast.Node) bool {
	parent := objectLiteral.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	// The left-side test is SUBSUMED: this function is called only from the property-assignment arm,
	// which is itself reached only while walking UP from an indexed access, and the very next thing
	// that arm does with a non-target object is recurse into isAssignedTo, which applies its own
	// left-side test to the same binary. So an object on the right of an assignment is declined
	// either way, and a mutation removing this clause produces identical output on every shape tried,
	// including the two where an indexed read sits inside the right-hand object. Written out because
	// it states what a destructuring target IS rather than relying on a caller two frames away to
	// notice.
	return binary.OperatorToken != nil &&
		ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
		binary.Left == objectLiteral
}
