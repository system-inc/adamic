package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Ruling 17, road (a), lowered (#d1chbm7, commit 2b): `const fn = (table as R)[x.kind]`, used only as
// `fn === undefined` and in calls `fn(x, ...)` and `fn?.(x, ...)`, where dispatch_cast.go's scan admitted the lookup.
// fn never holds a value: R erases every slot to any, which nothing may lower, so fn has no local and its type is
// never asked for a representation. Instead:
//   - fn's declaration holds x's kind in one local, read once as the read x.kind lowers (dispatchDeclaration). That
//     is Language's condition 3: the lookup's comparisons and each arm's check all read the local, never x.kind
//     again, so a kind that could answer twice can't pick one entry and pass another's check. The scan's
//     dispatchPlainKind holds kind a plain data field on every member, with no getter the program writes named kind
//     (r03.ts keeps its stop);
//   - `fn === undefined` is the held kind equal to none of the table's keys, one comparison per key
//     (dispatchAbsent);
//   - a call is one arm per key, in the literal's order: where the held kind is the key, the arm calls the entry the
//     table holds at that key's field (commit 1's numeric field), read off the table's binding, after the ruling's
//     check, explicit, made on the held kind: if it isn't the key, panic with row 9's words, cast failed: this <x's
//     type> is not a <the key's members> (castFailure). By construction it never fires (dispatchCall, dispatchArm).
//     Her 20:46:47Z rule is what the arm hands x as: the key's members, the union of every member the mapped type
//     gives the key, read from the checker's mapped property (dispatchKey's members), never one member of it and
//     never the entry's parameter, held as that union's representation with no further check, since the arm proved
//     the tag. Every member under one key carries the key's one kind value, so the check is one comparison;
//   - what no key matches: a plain call is admitted only as the false arm of a conditional whose condition is one of
//     fn's presence tests (dispatchCallGuarded), so its tail is an unreachable arm that panics; an optional call
//     answers undefined there, tested before any argument is evaluated, as ?. skips them.
//
// What keeps everything else out is the scan's: the proof (dispatchBinding, dispatchProofReference) that nothing
// but admitted lookups reads the table, so each key still holds what tsc checked there; Language's condition 1
// (dispatchSubjectFixed), so x is one value at the lookup and at every call; and the escapes (dispatchEscapes and
// dispatchCallGuarded), so fn is only compared with undefined or called with x first, in the function that declares
// it: fn.call, .apply and .bind, a spread argument, storing, returning or passing fn, a read from a nested function,
// a comparison other than === undefined and an unguarded plain call each keep today's stop at the assertion.
// Each admitted lookup is one line of --explain-checks (Program.DispatchChecks).

// dispatchCallee is the admitted dispatch site whose fn declaration is declaration, or nil: every other declaration
// lowers as before.
func (l *lowering) dispatchCallee(declaration *ast.Node) *dispatchSite {
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) || declaration.Initializer() == nil || ast.SkipParentheses(declaration.Initializer()).Kind != ast.KindElementAccessExpression {
		return nil
	}
	site := l.dispatchCasts().callees[l.symbol(declaration.Name())]
	if site == nil || site.calleeDeclaration != declaration {
		return nil
	}
	return site
}

// dispatchDeclaration lowers fn's declaration: one local, the kind, initialized by one read of x's kind, lowered as
// that read lowers anywhere. Nothing else is made, and the site is recorded as checked.
func (l *lowering) dispatchDeclaration(site *dispatchSite) ([]ir.Statement, error) {
	kind, err := l.expression(site.subjectKind)
	if err != nil {
		return nil, err
	}
	if kind.Type() != ir.Number {
		return nil, l.notYet(site.subjectKind, "a dispatch whose "+site.subjectKind.Name().Text()+" isn't held as a number")
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "dispatch_kind", Type: ir.Number, Function: function})
	site.kind = local + 1
	l.result.DispatchChecks = append(l.result.DispatchChecks, ir.DispatchCheck{Where: l.program.Where(site.lookup), Lookup: sourceExpression(site.lookup), Keys: len(site.keys)})
	return []ir.Statement{ir.Declare{Local: local, Value: kind}}, nil
}

// dispatchHeldKind is the read of the kind site's declaration holds.
func dispatchHeldKind(site *dispatchSite) ir.Expression {
	return ir.Read{Local: site.kind - 1, Of: ir.Number}
}

// dispatchValue lowers node where it's one of an admitted fn's uses: a presence test, or a call. handled is false
// anywhere else.
func (l *lowering) dispatchValue(node *ast.Node) (ir.Expression, bool, error) {
	switch node.Kind {
	case ast.KindBinaryExpression:
		if node.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
			return nil, false, nil
		}
	case ast.KindCallExpression:
		if !ast.IsIdentifier(ast.SkipParentheses(node.AsCallExpression().Expression)) {
			return nil, false, nil
		}
	default:
		return nil, false, nil
	}
	site, used := l.dispatchCasts().uses[node]
	if !used {
		return nil, false, nil
	}
	if site.kind == 0 {
		// fn's uses are in the function that declares it, after it, as tsc's use-before-declaration check holds them.
		return nil, true, l.notYet(node, "a use of a dispatch's fn lowered before its declaration")
	}
	if node.Kind == ast.KindBinaryExpression {
		return l.dispatchAbsent(site), true, nil
	}
	value, err := l.dispatchCall(site, node)
	return value, true, err
}

// dispatchAbsent is `fn === undefined`: the held kind equals none of the table's keys, one comparison per key,
// joined. A table with no entries finds nothing.
func (l *lowering) dispatchAbsent(site *dispatchSite) ir.Expression {
	var absent ir.Expression = ir.BooleanConstant{Value: true}
	for index, key := range site.keys {
		differs := ir.Binary{Operator: ir.NotEqual, Left: dispatchHeldKind(site), Right: ir.NumberConstant{Value: key.value}}
		if index == 0 {
			absent = differs
			continue
		}
		absent = ir.Binary{Operator: ir.And, Left: absent, Right: differs}
	}
	return absent
}

// dispatchCall lowers a call of fn: x and the other arguments once, in order, then one arm per key (dispatchArm),
// each taken where the held kind is its key, and a tail no admitted call reaches. A plain call stands where a
// presence test proved some key matches (dispatchCallGuarded), so its tail panics; an optional call tests for no
// key first and answers undefined, evaluating no argument, as ?. does. An argument the arms can't share as a plain
// read is held in a local of its own first, so it's evaluated once whichever arm runs.
func (l *lowering) dispatchCall(site *dispatchSite, node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	written := call.Arguments.Nodes
	// x is a const or a never-assigned parameter, by its name (dispatchSubjectFixed, dispatchCalleeUse), so its
	// value is its variable's, read as a variable is; the slot the view gives it is any, which it never goes into.
	subject := ast.SkipParentheses(written[0])
	subjectLocal, subjectIsLocal := l.local(subject)
	if !subjectIsLocal {
		return nil, l.notYet(subject, "a dispatch whose x is no variable of the program")
	}
	read, err := l.localRead(subject, subjectLocal)
	if err != nil {
		return nil, err
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	held := []ir.Statement{}
	hold := func(value ir.Expression) ir.Expression {
		if _, plain := value.(ir.Read); plain {
			return value
		}
		kept := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "dispatch_argument", Type: value.Type(), Function: function, ExpressionAssigned: true})
		held = append(held, ir.Declare{Local: kept, Value: value})
		return ir.Read{Local: kept, Of: value.Type()}
	}
	arguments := []ir.Expression{hold(read)}
	for _, argument := range written[1:] {
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, hold(value))
	}
	operand := ast.SkipParentheses(site.assertion.AsAsExpression().Expression)
	tableLocal, tableIsLocal := l.local(operand)
	if !tableIsLocal {
		return nil, l.notYet(operand, "a dispatch whose table is no variable of the program")
	}
	table, err := l.localRead(operand, tableLocal)
	if err != nil {
		return nil, err
	}
	// The tail: no admitted call gets here, so it stops loudly if one ever does.
	unreachable := ir.Panic{Message: ir.StringConstant{Index: l.constant("compiler bug: a dispatch call found no entry where its presence test found one")}}
	chain := ir.Expression(ir.Effects{Body: []ir.Statement{unreachable}, Result: unreachablePlaceholder(of)})
	if chain.Type() != of {
		return nil, l.notYet(node, "a dispatch call returning "+typeName(of))
	}
	source := l.concrete(l.checker.GetTypeAtLocation(written[0]))
	for index := len(site.keys) - 1; index >= 0; index-- {
		arm, err := l.dispatchArm(site, node, site.keys[index], table, arguments, source, of)
		if err != nil {
			return nil, err
		}
		chain = ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: dispatchHeldKind(site), Right: ir.NumberConstant{Value: site.keys[index].value}}, WhenTrue: arm, WhenNot: chain, Of: of}
	}
	if len(held) != 0 {
		chain = ir.Effects{Body: held, Result: chain}
	}
	if call.QuestionDotToken == nil {
		return chain, nil
	}
	absent := fit(ir.Undefined{}, of)
	if absent.Type() != of {
		return nil, l.notYet(node, "an optional dispatch call whose result can't hold undefined")
	}
	return ir.Conditional{Condition: l.dispatchAbsent(site), WhenTrue: absent, WhenNot: chain, Of: of}, nil
}

// dispatchArm is one key's arm of a call: the ruling's check on the held kind, then the call of the entry the table
// holds at the key's field, with x handed as the key's members (her 20:46:47Z rule) and the other arguments as the
// call passes them, each fitted to the entry's own parameter as callClosure fits a call's, and the result fitted to
// the call's own type, of.
func (l *lowering) dispatchArm(site *dispatchSite, node *ast.Node, key dispatchKey, table ir.Expression, arguments []ir.Expression, source *checker.Type, of ir.Type) (ir.Expression, error) {
	call := node.AsCallExpression()
	property := l.checker.GetPropertyOfType(l.checker.GetTypeOfSymbol(site.table), key.name)
	if property == nil || key.signature == nil || key.members == nil {
		return nil, l.notYet(node, "a dispatch whose key "+key.name+" has no slot that takes x")
	}
	stored, known := l.representation(l.checker.GetTypeOfSymbol(property))
	if !known || stored != ir.Closure {
		return nil, l.notYet(node, "a dispatch whose key "+key.name+" isn't held as a function value")
	}
	members, known := l.representation(key.members)
	if !known {
		return nil, l.notYet(node, "a dispatch whose key "+key.name+" takes "+l.checker.TypeToString(key.members)+", held no way yet")
	}
	// The arm proved x's tag is the key's, so x is the key's members: a value held as they're held passes as it is,
	// and a Union narrows to their representation with no test of its own, as a proven narrowed read does.
	subject := arguments[0]
	switch {
	case subject.Type() == members:
	case subject.Type() == ir.Union && members != ir.Union:
		subject = ir.Narrow{Value: subject, To: members}
	default:
		subject = fit(subject, members)
		if subject.Type() != members {
			return nil, l.notYet(node, "a dispatch whose x is held otherwise than key "+key.name+"'s "+l.checker.TypeToString(key.members))
		}
	}
	handed := append([]ir.Expression{subject}, arguments[1:]...)
	parameters := l.checker.GetSignatureFromDeclaration(key.function).Parameters()
	for index := range handed {
		if index >= len(parameters) {
			break
		}
		parameter := parameters[index]
		takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known {
			continue
		}
		// As callClosure: a defaulted parameter's incoming slot keeps undefined until the body runs.
		if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().Initializer != nil {
			takes = ir.Maybe(takes)
		}
		handed[index] = fit(handed[index], takes)
	}
	for _, argument := range handed {
		if censusCallableSlotless(argument.Type()) {
			return nil, l.notYet(node, "passing "+typeName(argument.Type())+" to a dispatch's entry")
		}
	}
	var returns ir.Type
	if result := l.concrete(l.checker.GetReturnTypeOfSignature(l.dispatchSlotAtCall(node, key.signature, call.Arguments.Nodes))); result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		if returns, known = l.representation(result); !known {
			return nil, l.notYet(node, "a dispatch call returning "+l.checker.TypeToString(result))
		}
	}
	if returns == 0 || censusCallableSlotless(returns) {
		return nil, l.notYet(node, "a dispatch call whose entries return "+typeName(returns))
	}
	// FunctionType is the function type the entry was stored as, the key's slot, whose targets the entries are, as a
	// call of table[K] would name it; never fn's own, the view's Fn<any>, whose targets aren't the table's.
	called := fit(ir.CallClosure{
		Closure:         ir.Property{Object: table, Name: key.name, Of: stored},
		Arguments:       handed,
		FunctionType:    int(l.concrete(l.checker.GetTypeOfSymbol(property)).Id()),
		Returns:         returns,
		ArgumentTypeIDs: l.argumentTypeIDs(call.Arguments.Nodes),
	}, of)
	if called.Type() != of {
		return nil, l.notYet(node, "a dispatch call whose entry's result is held otherwise than the call's "+typeName(of))
	}
	check := ir.If{
		Condition: ir.Binary{Operator: ir.NotEqual, Left: dispatchHeldKind(site), Right: ir.NumberConstant{Value: key.value}},
		Then:      []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(l.castFailure(source, key.members))}}},
	}
	return ir.Effects{Body: []ir.Statement{check}, Result: called}, nil
}
