package lower

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// The guarded-array certificate (Lower's design (b) on #ytagqs5, with Language's two musts and Adamic's
// rulings there): an overload promising a NodeArray of its predicate's type, from an implementation typed
// NodeArray<Node>, served per call because the body checks every element of what it returns with the
// overload's own predicate in the statement right before the return. tsc's visitNodes overload 1 is the
// shape, its body as 99-apple-visit-nodes-guard adapts it (stage3/adapt):
//
//	function visitNodes<TIn extends Node, TInArray extends NodeArray<TIn> | undefined, TOut extends Node>(nodes: TInArray, visitor: Visitor | undefined, test: (node: Node) => node is TOut, start?: number, count?: number): NodeArray<TOut> | (TInArray & undefined);
//	function visitNodes(nodes: NodeArray<Node> | undefined, visitor: Visitor | undefined, test?: (node: Node) => boolean, start?: number, count?: number): NodeArray<Node> | undefined {
//	    if (nodes === undefined) {
//	        return nodes;
//	    }
//	    ...
//	        const updatedArray = factory.createNodeArray(updated, hasTrailingComma);
//	        setTextRangePosEnd(updatedArray, pos, end);
//	        if (test !== undefined) {
//	            for (const testedNode of updatedArray) {
//	                if (!test(testedNode)) Debug.fail("False expression: ...");
//	            }
//	        }
//	        return updatedArray;
//	    ...
//	    if (test !== undefined) {
//	        for (const testedNode of nodes) {
//	            if (!test(testedNode)) Debug.fail("False expression: ...");
//	        }
//	    }
//	    return nodes;
//	}
//
// Every return is one of two arms (overloadGuardedArrayProof):
//
//  1. the array parameter itself, narrowed to undefined by the checker, as the first statement of the
//     branch `P === undefined` takes, which is the promise's `P & undefined` and nothing else (Lower's
//     condition 6);
//  2. R, a const or a parameter nothing writes, right after a for-of over R whose one statement is
//     `if (!test(x)) F(...)`, F a namespace member that never returns, so R reaches the return only when
//     the overload's own predicate said yes to each of its elements (condition 3).
//
// Each yes is the claim of the value the caller passed as the test, which predicateArguments makes prove
// itself or check itself: a named predicate proven by its body, a checked one's forwarder, or a predicate
// parameter forwarded as #jqf1drp's checked closure, so a lie stops at its own check (condition 2). That
// reading takes the test as exactly what the caller passed and present, so like the node-test proof
// (overload_parse_tree_node.go) the certificate is judged per call, at a direct entry only, and a call
// is served only where nothing the test runs between a check and the return can make an element stop being
// a TOut (overloadGuardedArrayCall):
//
//   - nothing writes the array: the program holds no writable view of a NodeArray, no (iv-a) value (an
//     array literal cast to MutableNodeArray and kept raw) and no MutableNodeArray or Array view of one
//     (condition 1), and a delayed Array mutator on a readonly receiver is refused program-wide by
//     #3cee4gb's refusal (delayedReadonlyReceiver, library_method_values.go), which this certificate reads as
//     its support and doesn't scan again (Language's must 1, Adamic's ruling);
//   - nothing changes what the test's yes read: every field the membership of TOut reads is readonly in
//     every view the program holds (Language's must 2, condition 8), and no getter of that name is in the
//     program (condition 4);
//
// and storage: the implementation's NodeArray<Node> is held as the call's NodeArray<TOut> is, read through
// nodeArrayLayoutOf (overloadCertificateSameHeld, condition 5). Any other entry is NotYet at certificate A's
// gate (overloadCertificateCall), which this certificate shares.

// overloadGuardedArrayWords opens every stop this certificate says at a call.
const overloadGuardedArrayWords = "a call to an overload whose NodeArray result is certified by its own element checks, "

// overloadGuardedArray is what the certificate reads from an overload: its predicate parameter, as
// predicateOverloadCallback names one, and the NodeArray member of the result it promises.
type overloadGuardedArray struct {
	test  predicateOverloadCallback
	array *checker.Type
}

// overloadGuardedArrayOf reads overload's guarded-array shape over implementation: a function declaration
// with a body, no binders of its own and only plain parameters; a promised result whose members are one
// NodeArray over readonly bases alone (overloadGuardedNodeArray) and undefined arms (overloadGuardedUndefinedArm);
// and exactly one required parameter typed as a function claiming `x is U` of one of its own parameters,
// U relating to that NodeArray's element, written as a function type so predicateArguments judges every
// argument passed for it (predicateContract), and whose implementation parameter the body never assigns.
func (l *lowering) overloadGuardedArrayOf(implementation, overload *ast.Node) (overloadGuardedArray, bool) {
	guarded := overloadGuardedArray{}
	if overload == nil || !overloadCertificateFunction(implementation) || len(implementation.TypeParameters()) != 0 {
		return guarded, false
	}
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	// The overload's own types, whoever asks, as overloadNodeTest reads them.
	l.typeMapper = nil
	served := implementation.Parameters()
	for _, parameter := range served {
		declaration := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || declaration.Initializer != nil || declaration.DotDotDotToken != nil {
			return guarded, false // A default replaces what the caller passed, and a rest moves positions.
		}
	}
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	members := []*checker.Type{promised}
	if promised.Flags()&checker.TypeFlagsUnion != 0 {
		members = promised.Types()
	}
	for _, member := range members {
		if _, undefinedArm := l.overloadGuardedUndefinedArm(member); undefinedArm {
			continue
		}
		if guarded.array != nil || !l.overloadGuardedNodeArray(member) {
			return guarded, false
		}
		guarded.array = member
	}
	if guarded.array == nil {
		return guarded, false
	}
	element := l.checker.GetIndexTypeOfType(guarded.array, l.checker.GetNumberType())
	if element == nil {
		return guarded, false
	}
	found := false
	for index, parameter := range overload.Parameters() {
		if index >= len(served) || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			break // A rest takes any count of arguments, so no position past it says which is the test.
		}
		if !ast.IsIdentifier(parameter.Name()) || ast.HasQuestionToken(parameter) {
			continue
		}
		symbol := l.symbol(parameter.Name())
		signatures := l.checker.GetSignaturesOfType(l.censusCallableParameterType(symbol), checker.SignatureKindCall)
		if len(signatures) != 1 {
			continue
		}
		claim := predicateOfSignature(l.checker, signatures[0])
		if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || claim.Type() == nil || claim.ParameterIndex() < 0 || !l.censusRelated(claim.Type(), element) {
			continue
		}
		// The checked road (condition 2): a claim written as a function type is the one predicateArguments
		// holds every argument to, proven where it's passed or checked by its own forwarder or closure. A
		// claim reached through an alias is one predicateContract doesn't read, so nothing would judge what
		// the caller passed.
		if predicateContract(symbol) == nil {
			return overloadGuardedArray{}, false
		}
		actual := served[index]
		if l.predicateOverloadAssigned(implementation, l.symbol(actual.Name())) {
			continue
		}
		if found {
			return overloadGuardedArray{}, false // Two tests leave each call unsure which one it must pass.
		}
		guarded.test, found = predicateOverloadCallback{parameter: index, claimed: int(claim.ParameterIndex())}, true
	}
	return guarded, found
}

// overloadGuardedUndefinedArm is member's parts beside undefined when member is a promise's undefined arm:
// undefined itself, with no parts, or an intersection of undefined with other types, `TInArray & undefined`.
func (l *lowering) overloadGuardedUndefinedArm(member *checker.Type) ([]*checker.Type, bool) {
	if member.Flags()&checker.TypeFlagsUndefined != 0 {
		return nil, true
	}
	if member.Flags()&checker.TypeFlagsIntersection == 0 {
		return nil, false
	}
	parts, undefined := []*checker.Type{}, false
	for _, part := range member.Types() {
		if part.Flags()&checker.TypeFlagsUndefined != 0 {
			undefined = true
			continue
		}
		parts = append(parts, part)
	}
	return parts, undefined
}

// overloadGuardedNodeArray is whether proven is a NodeArray-shaped interface whose every library array base
// is ReadonlyArray: an interface reaching Array would be a writable view of its own elements, which the
// element checks can't hold to (condition 1).
func (l *lowering) overloadGuardedNodeArray(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(proven) || checker.IsTupleType(proven) {
		return false
	}
	readonly, writable := l.overloadGuardedArrayBases(proven, map[*checker.Type]bool{})
	return readonly && !writable
}

// overloadGuardedArrayBases is whether proven's base-type chain, as nodeArrayBase walks it, reaches the
// library's ReadonlyArray, and whether it reaches its Array, walking every base rather than the first.
func (l *lowering) overloadGuardedArrayBases(proven *checker.Type, seen map[*checker.Type]bool) (readonly, writable bool) {
	if proven == nil || seen[proven] || proven.Flags()&checker.TypeFlagsObject == 0 {
		return false, false
	}
	seen[proven] = true
	if l.isLibraryType(proven, "ReadonlyArray") {
		return true, false
	}
	if l.isLibraryType(proven, "Array") {
		return false, true
	}
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	if target.ObjectFlags()&checker.ObjectFlagsInterface == 0 {
		return false, false
	}
	for _, base := range l.checker.GetBaseTypes(target) {
		baseReadonly, baseWritable := l.overloadGuardedArrayBases(base, seen)
		readonly, writable = readonly || baseReadonly, writable || baseWritable
	}
	return readonly, writable
}

// overloadGuardedArrayProof is whether implementation serves overload's NodeArray result under the
// premise that a call passes its test: the body reads no arguments, holds no try, can't fall off its end,
// and every return in it, nested functions aside, is one of the two arms (overloadGuardedUndefinedReturn,
// overloadGuardedCheckedReturn). Anything else fails, and the overload stays refused for its result.
func (l *lowering) overloadGuardedArrayProof(implementation, overload *ast.Node) bool {
	guarded, found := l.overloadGuardedArrayOf(implementation, overload)
	if !found {
		return false
	}
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	l.typeMapper = nil
	statements := implementation.Body().AsBlock().Statements.Nodes
	if len(statements) == 0 || statements[len(statements)-1].Kind != ast.KindReturnStatement && statements[len(statements)-1].Kind != ast.KindThrowStatement {
		return false // Falling off the end returns undefined, which no call whose array is present was promised.
	}
	// A body reading arguments can write a parameter by its position, and a finally runs its own code after
	// a return's value is read, between the checks and the caller.
	sloppy := false
	var scan func(node *ast.Node, nested bool)
	scan = func(node *ast.Node, nested bool) {
		if ast.IsIdentifier(node) && node.Text() == "arguments" || !nested && node.Kind == ast.KindTryStatement {
			sloppy = true
		}
		inner := nested || ast.IsFunctionLike(node)
		node.ForEachChild(func(child *ast.Node) bool {
			scan(child, inner)
			return false
		})
	}
	implementation.Body().ForEachChild(func(child *ast.Node) bool {
		scan(child, false)
		return false
	})
	if sloppy {
		return false
	}
	test := l.symbol(implementation.Parameters()[guarded.test.parameter].Name())
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	proven := true
	var returns func(node *ast.Node)
	returns = func(node *ast.Node) {
		if !proven || ast.IsFunctionLike(node) {
			return // A nested function's return is its own result, not this call's.
		}
		if node.Kind == ast.KindReturnStatement && !l.overloadGuardedReturn(implementation, overload, guarded, test, promised, node) {
			proven = false
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			returns(child)
			return false
		})
	}
	implementation.Body().ForEachChild(func(child *ast.Node) bool {
		returns(child)
		return false
	})
	return proven
}

// overloadGuardedReturn is whether statement, a return in implementation's body, is one of the
// certificate's two arms: a name, declared once and never assigned anywhere in the body, nested functions
// included (predicateOverloadAssigned), returned either as the undefined arm or as a checked array.
func (l *lowering) overloadGuardedReturn(implementation, overload *ast.Node, guarded overloadGuardedArray, test *ast.Symbol, promised *checker.Type, statement *ast.Node) bool {
	expression := statement.AsReturnStatement().Expression
	if expression == nil {
		return false
	}
	returned := ast.SkipParentheses(expression)
	if !ast.IsIdentifier(returned) {
		return false
	}
	local := l.valueSymbol(returned)
	if local == nil || len(local.Declarations) != 1 || l.predicateOverloadAssigned(implementation, local) {
		return false
	}
	if l.overloadGuardedUndefinedReturn(implementation, overload, promised, statement, returned) {
		return true
	}
	return l.overloadGuardedCheckedReturn(implementation, guarded, test, statement, returned, local)
}

// overloadGuardedUndefinedReturn is the undefined arm (condition 6): returned is an implementation parameter,
// its type at the return narrowed by the checker to undefined alone, and the return is the first statement
// of the branch an if takes on `P === undefined` of that very parameter, so nothing runs between the test and
// the return. promised must hold `P & undefined` for the overload's own type of that parameter, or undefined
// itself: the value is the caller's argument, undefined, and only a call whose argument may be undefined is
// promised it. A return of anything else typed undefined, `return undefined` among them, isn't this arm.
func (l *lowering) overloadGuardedUndefinedReturn(implementation, overload *ast.Node, promised *checker.Type, statement, returned *ast.Node) bool {
	parameters := implementation.Parameters()
	index := l.overloadParameterIndex(parameters, returned)
	if index < 0 || index >= len(overload.Parameters()) {
		return false
	}
	narrowed := l.checker.GetTypeAtLocation(returned)
	if narrowed == nil || narrowed.Flags()&checker.TypeFlagsUndefined == 0 || narrowed.Flags()&checker.TypeFlagsUnion != 0 {
		return false
	}
	then := statement
	if statement.Parent != nil && statement.Parent.Kind == ast.KindBlock {
		if block := statement.Parent.AsBlock(); len(block.Statements.Nodes) == 0 || block.Statements.Nodes[0] != statement {
			return false
		}
		then = statement.Parent
	}
	branch := then.Parent
	if branch == nil || branch.Kind != ast.KindIfStatement || branch.AsIfStatement().ThenStatement != then {
		return false
	}
	condition := ast.SkipParentheses(branch.AsIfStatement().Expression)
	if condition.Kind != ast.KindBinaryExpression || condition.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
		return false
	}
	left, right := condition.AsBinaryExpression().Left, condition.AsBinaryExpression().Right
	if !(l.overloadParameterIndex(parameters, left) == index && l.overloadGuardedUndefined(right) || l.overloadParameterIndex(parameters, right) == index && l.overloadGuardedUndefined(left)) {
		return false
	}
	given, _, rest, err := l.censusOverloadParameter(overload.Parameters(), index)
	if err != nil || rest || given == nil {
		return false
	}
	members := []*checker.Type{promised}
	if promised.Flags()&checker.TypeFlagsUnion != 0 {
		members = promised.Types()
	}
	for _, member := range members {
		parts, undefinedArm := l.overloadGuardedUndefinedArm(member)
		if !undefinedArm {
			continue
		}
		if len(parts) == 0 {
			return true
		}
		if len(parts) == 1 && checker.Checker_isTypeIdenticalTo(l.checker, parts[0], given) {
			return true
		}
	}
	return false
}

// overloadGuardedUndefined is whether node, past parentheses, names the global undefined.
func (l *lowering) overloadGuardedUndefined(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return ast.IsIdentifier(node) && l.valueSymbol(node) == l.checker.GetUndefinedSymbol()
}

// overloadGuardedCheckedReturn is the checked arm (condition 3): local is a const declared in the body or one
// of its parameters, held as a NodeArray of the promised one's own interface, and the statement right before
// the return, in the same block, is the guard over it (overloadGuardedLoop). A label, a second statement or a
// write anywhere fails closed: the guard has to be the one statement between the last element check and the
// value the caller gets.
func (l *lowering) overloadGuardedCheckedReturn(implementation *ast.Node, guarded overloadGuardedArray, test *ast.Symbol, statement, returned *ast.Node, local *ast.Symbol) bool {
	declaration := local.Declarations[0]
	switch declaration.Kind {
	case ast.KindParameter:
		if declaration.Parent != implementation {
			return false
		}
	case ast.KindVariableDeclaration:
		if declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst || ast.GetContainingFunction(declaration) != implementation {
			return false
		}
	default:
		return false
	}
	held := l.definedMembers(l.checker.GetTypeAtLocation(returned))
	if len(held) != 1 || !l.overloadGuardedNodeArray(held[0]) || held[0].Symbol() == nil || held[0].Symbol() != guarded.array.Symbol() {
		return false
	}
	block := statement.Parent
	if block == nil || block.Kind != ast.KindBlock {
		return false
	}
	statements := block.AsBlock().Statements.Nodes
	position := slices.Index(statements, statement)
	if position < 1 {
		return false
	}
	return l.overloadGuardedLoop(implementation, guarded, test, statements[position-1], local)
}

// overloadGuardedLoop is whether guard is the element check over array: a for-of over array, no await, a
// const binding x, whose one statement is `if (!test(x)) F(...)` with no else, test the overload's own
// predicate parameter called with x where its claim reads (predicateOverloadCallTests), and F a call that
// never returns (overloadGuardedFails). tsc's form wraps the loop in `if (test !== undefined)` with no else,
// the loop its one statement; under the premise every served call keeps, that its test is passed and
// present, that branch is taken.
func (l *lowering) overloadGuardedLoop(implementation *ast.Node, guarded overloadGuardedArray, test *ast.Symbol, guard *ast.Node, array *ast.Symbol) bool {
	loop := guard
	if guard.Kind == ast.KindIfStatement {
		branch := guard.AsIfStatement()
		if branch.ElseStatement != nil || !l.overloadGuardedPresent(branch.Expression, test) {
			return false
		}
		loop = overloadGuardedOnly(branch.ThenStatement)
		if loop == nil {
			return false
		}
	}
	if loop.Kind != ast.KindForOfStatement {
		return false
	}
	iteration := loop.AsForInOrOfStatement()
	iterated := ast.SkipParentheses(iteration.Expression)
	if iteration.AwaitModifier != nil || !ast.IsIdentifier(iterated) || l.valueSymbol(iterated) != array || iteration.Initializer == nil || iteration.Initializer.Kind != ast.KindVariableDeclarationList {
		return false
	}
	bindings := iteration.Initializer.AsVariableDeclarationList().Declarations.Nodes
	if iteration.Initializer.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst || len(bindings) != 1 || !ast.IsIdentifier(bindings[0].Name()) {
		return false
	}
	element := l.symbol(bindings[0].Name())
	check := overloadGuardedOnly(iteration.Statement)
	if check == nil || check.Kind != ast.KindIfStatement || check.AsIfStatement().ElseStatement != nil {
		return false
	}
	condition := ast.SkipParentheses(check.AsIfStatement().Expression)
	if condition.Kind != ast.KindPrefixUnaryExpression || condition.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
		return false
	}
	call := ast.SkipParentheses(condition.AsPrefixUnaryExpression().Operand)
	if call.Kind != ast.KindCallExpression || call.AsCallExpression().QuestionDotToken != nil || call.AsCallExpression().TypeArguments != nil {
		return false
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if !ast.IsIdentifier(callee) || l.valueSymbol(callee) != test || !l.predicateOverloadCallTests(implementation, call, []predicateOverloadCallback{guarded.test}, element) {
		return false
	}
	failure := overloadGuardedOnly(check.AsIfStatement().ThenStatement)
	return failure != nil && l.overloadGuardedFails(failure)
}

// overloadGuardedOnly is statement itself, or the one statement of a block, or nil for a block of any other
// count.
func overloadGuardedOnly(statement *ast.Node) *ast.Node {
	if statement == nil || statement.Kind != ast.KindBlock {
		return statement
	}
	if statements := statement.AsBlock().Statements.Nodes; len(statements) == 1 {
		return statements[0]
	}
	return nil
}

// overloadGuardedPresent is whether condition, past parentheses, is `test !== undefined` or its mirror.
func (l *lowering) overloadGuardedPresent(condition *ast.Node, test *ast.Symbol) bool {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindBinaryExpression || condition.AsBinaryExpression().OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	left, right := ast.SkipParentheses(condition.AsBinaryExpression().Left), ast.SkipParentheses(condition.AsBinaryExpression().Right)
	names := func(node *ast.Node) bool { return ast.IsIdentifier(node) && l.valueSymbol(node) == test }
	return names(left) && l.overloadGuardedUndefined(right) || names(right) && l.overloadGuardedUndefined(left)
}

// overloadGuardedFails is whether statement is a call that never returns: an expression statement calling a
// namespace's member, tsc's Debug.fail, with no ?., whose resolved signature's result is never, so the loop
// goes on, and the return is reached, only past a yes.
func (l *lowering) overloadGuardedFails(statement *ast.Node) bool {
	if statement.Kind != ast.KindExpressionStatement {
		return false
	}
	call := ast.SkipParentheses(statement.AsExpressionStatement().Expression)
	if call.Kind != ast.KindCallExpression || call.AsCallExpression().QuestionDotToken != nil {
		return false
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.AsPropertyAccessExpression().QuestionDotToken != nil || !l.namespaceMember(callee) {
		return false
	}
	signature := l.checker.GetResolvedSignature(call)
	return signature != nil && l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsNever != 0
}

// overloadGuardedArrayCertified is whether censusOverload accepted overload by the guarded-array
// certificate, judged once per overload as overloadNodeTestCertified judges the node-test proof: an
// overload not yet judged, when its body passes the proof, is judged by running its set's census, which
// records the acceptance.
func (l *lowering) overloadGuardedArrayCertified(overload *ast.Node) (bool, error) {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil || overload.Name() == nil {
		return false, nil
	}
	if certified, judged := l.overloadGuardedArrays[overload]; judged {
		return certified, nil
	}
	certified := false
	if implementation := l.censusImplementation(overload); implementation != nil && l.overloadGuardedArrayProof(implementation, overload) {
		// Its own verdict, never its set's, as overloadCertified reads it (#4czdn4q).
		if err := l.overloadVerdict(overload); err != nil {
			return false, err
		}
		certified = l.overloadGuardedArrays[overload]
	}
	l.noteOverloadGuardedArray(overload, certified)
	return certified, nil
}

// noteOverloadGuardedArray records whether census accepted overload by the guarded-array certificate.
func (l *lowering) noteOverloadGuardedArray(overload *ast.Node, certified bool) {
	if l.overloadGuardedArrays == nil {
		l.overloadGuardedArrays = map[*ast.Node]bool{}
	}
	l.overloadGuardedArrays[overload] = certified
}

// overloadGuardedArrayCall is nil when node, a direct call resolved to an overload the guarded-array
// certificate accepted, keeps the premises the certificate reads its body under, and NotYet at the call
// naming the first it doesn't keep. overloadCertificateCall asks it once the call's entry is direct, before
// the storage condition and before the callee or an argument is lowered.
func (l *lowering) overloadGuardedArrayCall(node, overload *ast.Node, resolved *checker.Signature) error {
	guarded, found := l.overloadGuardedArrayOf(l.censusImplementation(overload), overload)
	if !found {
		return l.notYet(node, overloadGuardedArrayWords+"whose test isn't found")
	}
	// The test is passed and present, so predicateArguments holds it to its claim at this call, and the
	// body's `test !== undefined` takes the branch that checks (condition 2).
	arguments := node.AsCallExpression().Arguments.Nodes
	if guarded.test.parameter >= len(arguments) || arguments[guarded.test.parameter].Kind == ast.KindSpreadElement {
		return l.notYet(node, overloadGuardedArrayWords+"made without its test")
	}
	passed := l.concrete(l.checker.GetTypeAtLocation(arguments[guarded.test.parameter]))
	if l.censusHasUndefined(passed) || passed.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return l.notYet(node, overloadGuardedArrayWords+"made with a test that may be absent")
	}
	// The claim at this call: its instance of TOut, and the element its predicate takes.
	parameters := resolved.Parameters()
	if guarded.test.parameter >= len(parameters) {
		return l.notYet(node, overloadGuardedArrayWords+"whose test isn't found")
	}
	signatures := l.checker.GetSignaturesOfType(l.withoutUndefined(l.checker.GetTypeOfSymbol(parameters[guarded.test.parameter])), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return l.notYet(node, overloadGuardedArrayWords+"whose test isn't one function")
	}
	claim := predicateOfSignature(l.checker, signatures[0])
	takes := signatures[0].Parameters()
	if claim == nil || claim.Type() == nil || guarded.test.claimed >= len(takes) {
		return l.notYet(node, overloadGuardedArrayWords+"whose test claims nothing")
	}
	target := l.concrete(claim.Type())
	element := l.concrete(l.checker.GetTypeOfSymbol(takes[guarded.test.claimed]))
	fields, named := l.overloadGuardedMembership(target)
	if !named {
		return l.notYet(node, overloadGuardedArrayWords+"whose test's claim "+l.checker.TypeToString(target)+" reads fields the certificate can't name")
	}
	views := l.overloadGuardedViewsOf()
	if views.writable != nil {
		// Condition 1: a writable view can put a non-TOut element in an array after its check, during the
		// later checks themselves, since each runs program code. Delayed mutators on a readonly receiver are
		// #3cee4gb's refusal, program-wide, not this scan's (Language's must 1, Adamic's ruling).
		return l.notYet(node, overloadGuardedArrayWords+"in a program holding a writable view of a NodeArray at "+l.program.Where(views.writable))
	}
	// Condition 4: a getter answers each read afresh, so the yes a check read is no fact about the next read.
	// Judged by name across the program, which is wider than accessorReach and fails closed: a getter whose
	// name isn't known may have any name, so it stops every claim that reads a field.
	if getter := views.getters[""]; getter != nil && len(fields) != 0 {
		return l.notYet(node, overloadGuardedArrayWords+"in a program with a getter whose name isn't known, at "+l.program.Where(getter))
	}
	for _, field := range fields {
		if getter := views.getters[field]; getter != nil {
			return l.notYet(node, overloadGuardedArrayWords+"whose test's claim reads "+field+", a getter's name at "+l.program.Where(getter))
		}
		// Condition 8 (Language's must 2): membership read from a field some view can write is a moment's
		// answer, and a later check can write an earlier element's field.
		if view, writable := l.overloadGuardedWritableField(views, field, element, target); writable {
			where := l.program.Where(node)
			if view != nil {
				where = l.program.Where(view)
			}
			return l.notYet(node, overloadGuardedArrayWords+"whose test's claim reads "+field+", writable in a view the program holds at "+where)
		}
	}
	return nil
}

// overloadGuardedMembership is the fields a check of target reads, as predicateMembership reads them: none
// for a primitive or undefined, tested by typeof, kind for a member tagged by a kind literal, and none for a
// branded supertype, which is tested as an object while its views check each field where it's read. Any
// other member, a branded refinement or an untagged object among them, answers false, so the call is NotYet.
func (l *lowering) overloadGuardedMembership(target *checker.Type) ([]string, bool) {
	members := []*checker.Type{target}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		members = target.Types()
	}
	fields := []string{}
	for _, member := range members {
		switch {
		case member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0:
		case member.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(member) && l.fieldLiteral(member, "kind") != nil:
			if !slices.Contains(fields, "kind") {
				fields = append(fields, "kind")
			}
		case member.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(member) && l.brandedSupertype(member):
		default:
			return nil, false
		}
	}
	return fields, true
}

// overloadGuardedView is one type the program holds a value as, and where.
type overloadGuardedView struct {
	node *ast.Node
	held *checker.Type
}

// overloadGuardedViews is what the certificate reads of the whole program, once: the first place it holds a
// writable view of a NodeArray, every type a binding, a field, a cast or a declaration holds a value as, and
// the first getter of each name, "" for one whose name isn't known.
type overloadGuardedViews struct {
	writable *ast.Node
	held     []overloadGuardedView
	getters  map[string]*ast.Node
}

// overloadGuardedViewsOf walks the program's own files, as overloadCallbackOwnSlice does, once per lowering.
// A writable view of a NodeArray is made only where a type is asserted, since a NodeArray is readonly to
// every relation the checker runs and the invariance pass refuses a readonly view widened to a mutable one:
// a cast or assertion to an interface over Array, construction (iv-a)'s MutableNodeArray, unless it's an
// array literal written in place and handed straight on as a holder that can't write it
// (overloadGuardedHandedOn), as tsc's createNodeArray hands its cast to its NodeArray result; from a
// NodeArray, or from a source that may hold one without saying so (overloadGuardedMayHoldNodeArray), to
// anything but a readonly holder; or to a NodeArray from a mutable array that outlives it, which an array
// literal written in place doesn't; and where a narrowing makes one, a predicate claiming a mutable array of
// either (Array.isArray), or instanceof Array of either. Each is the program's, whether or not it lowers, so
// the answer never leans on the order stops come in.
func (l *lowering) overloadGuardedViewsOf() *overloadGuardedViews {
	if l.overloadGuardedViews != nil {
		return l.overloadGuardedViews
	}
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	l.typeMapper = nil
	views := &overloadGuardedViews{getters: map[string]*ast.Node{}}
	hold := func(node *ast.Node, held *checker.Type) {
		if held != nil {
			views.held = append(views.held, overloadGuardedView{node: node, held: held})
		}
	}
	writable := func(node *ast.Node) {
		if views.writable == nil {
			views.writable = node
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			var expression *ast.Node
			if node.Kind == ast.KindAsExpression {
				expression = node.AsAsExpression().Expression
			} else {
				expression = node.AsTypeAssertion().Expression
			}
			target := l.checker.GetTypeAtLocation(node)
			hold(node, target)
			literal := ast.SkipParentheses(expression).Kind == ast.KindArrayLiteralExpression
			if l.overloadGuardedWritableCast(l.checker.GetTypeAtLocation(expression), target, literal) && !(literal && l.overloadGuardedHandedOn(node)) {
				writable(node)
			}
		case ast.KindCallExpression:
			if signature := l.checker.GetResolvedSignature(node); signature != nil {
				if claim := predicateOfSignature(l.checker, signature); claim != nil && claim.Type() != nil && claim.Kind() == checker.TypePredicateKindIdentifier && claim.ParameterIndex() >= 0 {
					arguments := node.AsCallExpression().Arguments.Nodes
					if index := int(claim.ParameterIndex()); index < len(arguments) {
						argument := l.checker.GetTypeAtLocation(arguments[index])
						nodeArray, _, _ := l.overloadGuardedArrayKinds(argument)
						_, _, mutable := l.overloadGuardedArrayKinds(claim.Type())
						if (nodeArray || l.overloadGuardedMayHoldNodeArray(argument, map[*checker.Type]bool{})) && mutable {
							writable(node)
						}
					}
				}
			}
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindInstanceOfKeyword && l.isLibraryGlobal(binary.Right, "Array") {
				left := l.checker.GetTypeAtLocation(binary.Left)
				if nodeArray, _, _ := l.overloadGuardedArrayKinds(left); nodeArray || l.overloadGuardedMayHoldNodeArray(left, map[*checker.Type]bool{}) {
					writable(node)
				}
			}
		case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindBindingElement:
			if name := node.Name(); name != nil && ast.IsIdentifier(name) {
				hold(node, l.checker.GetTypeAtLocation(name))
			} else if node.Type() != nil {
				hold(node, l.checker.GetTypeAtLocation(node.Type()))
			}
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration:
			if node.Name() != nil {
				hold(node, l.checker.GetTypeAtLocation(node.Name()))
			}
		case ast.KindTypeAliasDeclaration:
			if node.Type() != nil {
				hold(node, l.checker.GetTypeAtLocation(node.Type()))
			}
		case ast.KindGetAccessor:
			name, known := memberKeyText(node)
			if !known {
				name = ""
			}
			if views.getters[name] == nil {
				views.getters[name] = node
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range l.program.SourceFiles() {
		if load.IsNodeLibrary(file) || strings.HasPrefix(file.FileName().AsString(), "bundled:///node/node_modules/undici-types/") {
			continue
		}
		file.AsNode().ForEachChild(visit)
	}
	l.overloadGuardedViews = views
	return views
}

// overloadGuardedHandedOn is whether node, a cast, is handed straight to a holder that can't write an array:
// through parentheses alone, the operand of a return whose function, neither async nor a generator, declares
// a result no member of which can, or the initializer of a declaration whose annotation is such a type.
// It's sound because the literal is fresh: written in place, it's an array nothing else has ever held, and the
// cast's own view, the one writable view of it, is never bound, since the one holder it reaches is that result
// or that declaration, neither of which can write it. Any later writable view of the NodeArray is another cast,
// from a NodeArray or from a source that may hold one, which the scan flags. So an array literal cast there is a
// construction that keeps no writable alias, not (iv-a)'s value kept raw. Anything else, an arrow's expression
// body, satisfies, an async function or a generator, an unannotated declaration, answers false, as kept.
func (l *lowering) overloadGuardedHandedOn(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	var holder *checker.Type
	switch {
	case parent == nil:
		return false
	case parent.Kind == ast.KindReturnStatement:
		function := ast.GetContainingFunction(parent)
		if function == nil || function.Type() == nil || ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) || generatorDeclaration(function) {
			return false
		}
		holder = l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function))
	case parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == outer && parent.Type() != nil:
		holder = l.checker.GetTypeAtLocation(parent.Type())
	default:
		return false
	}
	_, writable, mutable := l.overloadGuardedArrayKinds(holder)
	return !writable && !mutable
}

// overloadGuardedArrayKinds reads proven's members, nullish aside and through unions and intersections:
// whether one is a NodeArray-shaped interface, whether one is such an interface over Array, writable, and
// whether one can write an array's elements or its length: a mutable array or tuple, such an interface, any,
// or an object type that isn't a readonly holder (overloadGuardedReadonlyHolder).
func (l *lowering) overloadGuardedArrayKinds(proven *checker.Type) (nodeArray, writable, mutable bool) {
	if proven == nil {
		return false, false, false
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			memberNodeArray, memberWritable, memberMutable := l.overloadGuardedArrayKinds(member)
			nodeArray, writable, mutable = nodeArray || memberNodeArray, writable || memberWritable, mutable || memberMutable
		}
		return nodeArray, writable, mutable
	}
	if proven.Flags()&checker.TypeFlagsAny != 0 {
		return false, false, true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false, false, false
	}
	if l.isArrayType(proven) || checker.IsTupleType(proven) {
		return false, false, !l.overloadGuardedReadonlyHolder(proven)
	}
	readonly, overArray := l.overloadGuardedArrayBases(proven, map[*checker.Type]bool{})
	if readonly || overArray {
		return true, overArray, overArray
	}
	return false, false, !l.overloadGuardedReadonlyHolder(proven)
}

// overloadGuardedReadonlyHolder is whether an object type proven can't write an array's elements or length
// held as it: lib's ReadonlyArray, a readonly tuple, or a type with no index signature and no writable
// length. Anything else, a mutable array or a writable number index among them, can.
func (l *lowering) overloadGuardedReadonlyHolder(proven *checker.Type) bool {
	if l.isLibraryType(proven, "ReadonlyArray") {
		return true
	}
	if checker.IsTupleType(proven) {
		return proven.TargetTupleType().IsReadonly()
	}
	if l.isArrayType(proven) || len(l.checker.GetIndexInfosOfType(proven)) != 0 {
		return false
	}
	length := l.checker.GetPropertyOfType(proven, "length")
	return length == nil || l.checker.IsReadonlySymbol(length)
}

// overloadGuardedWritableCast is whether a cast from source to target makes a writable view of a NodeArray:
// to an interface over Array, from a NodeArray, or from a source that may hold one without saying so
// (overloadGuardedMayHoldNodeArray), to anything that can write its elements or length, or to a NodeArray
// from a mutable array that isn't an array literal written in place, which nothing else holds.
func (l *lowering) overloadGuardedWritableCast(source, target *checker.Type, literal bool) bool {
	sourceNodeArray, _, sourceMutable := l.overloadGuardedArrayKinds(source)
	sourceNodeArray = sourceNodeArray || l.overloadGuardedMayHoldNodeArray(source, map[*checker.Type]bool{})
	targetNodeArray, targetWritable, targetMutable := l.overloadGuardedArrayKinds(target)
	return targetWritable || sourceNodeArray && targetMutable || targetNodeArray && sourceMutable && !literal
}

// overloadGuardedMayHoldNodeArray is whether a value of type proven may be a NodeArray without its type saying so
// (Lower's must candidate on #ytagqs5): a top type hop, `nodes as unknown` and then `as Node[]`, makes a writable
// view the cast from a NodeArray alone never shows. So any, unknown, object, an object type with no property
// and no call or construct signature ({}, or an index signature alone), lib's ReadonlyArray and a readonly tuple,
// either of which a NodeArray is assigned to as it is, and a type parameter whose constraint, absent or read the
// same way, doesn't exclude one, each may. A union may when a member may, an intersection when every part may,
// and a type met again on the walk may, failing closed.
func (l *lowering) overloadGuardedMayHoldNodeArray(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if proven == nil {
		return false
	}
	if seen[proven] {
		return true
	}
	seen[proven] = true
	switch {
	case proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0:
		return true
	case proven.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range proven.Types() {
			if l.overloadGuardedMayHoldNodeArray(member, seen) {
				return true
			}
		}
		return false
	case proven.Flags()&checker.TypeFlagsIntersection != 0:
		for _, part := range proven.Types() {
			if !l.overloadGuardedMayHoldNodeArray(part, seen) {
				return false
			}
		}
		return true
	case proven.Flags()&checker.TypeFlagsTypeParameter != 0:
		constraint := l.checker.GetBaseConstraintOfType(proven)
		return constraint == nil || constraint == proven || l.overloadGuardedMayHoldNodeArray(constraint, seen)
	case proven.Flags()&checker.TypeFlagsObject == 0:
		return false
	case l.isLibraryType(proven, "ReadonlyArray"), checker.IsTupleType(proven) && proven.TargetTupleType().IsReadonly():
		return true
	}
	return len(l.checker.GetPropertiesOfType(proven)) == 0 && !l.castSignatures(proven)
}

// overloadGuardedWritableField is whether some view the program holds can write field of a value the check
// read it on: element, the type the test takes, target's members, and every held type (overloadGuardedViewsOf)
// one of whose members, nullish aside, declares field writable and relates to element or to one of target's
// members by assignment either way. It answers where, nil when it's element or target itself.
func (l *lowering) overloadGuardedWritableField(views *overloadGuardedViews, field string, element, target *checker.Type) (*ast.Node, bool) {
	related := []*checker.Type{element}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		related = append(related, target.Types()...)
	} else {
		related = append(related, target)
	}
	writes := func(held *checker.Type) bool {
		members := []*checker.Type{held}
		if held.Flags()&checker.TypeFlagsUnion != 0 {
			members = held.Types()
		}
		for _, member := range members {
			if member.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
				continue
			}
			property := l.checker.GetPropertyOfType(member, field)
			if property == nil || l.checker.IsReadonlySymbol(property) {
				continue
			}
			for _, other := range related {
				if l.checker.IsTypeAssignableTo(member, other) || l.checker.IsTypeAssignableTo(other, member) {
					return true
				}
			}
		}
		return false
	}
	for _, own := range related {
		if writes(own) {
			return nil, true
		}
	}
	for _, view := range views.held {
		if writes(view.held) {
			return view.node, true
		}
	}
	return nil, false
}
