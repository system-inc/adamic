package lower

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// dictionaryView is whether a cast castProof would stop as castIndexSignature is the checked dictionary view
// instead (#2v3xtgm, re-ruled Nov 23): tsc's 22 `object as MapLike<unknown>`, a value typed object read by its own
// keys. Natively a record's values are stored unboxed by the element its static type names (records.go's
// recordSlot), and only a read out of a slot held as a Union is checked (recordReadDeclared), so the view is admitted
// only where every read through it is a Union and the record holds Union-readable values:
//   - the target's non-nullish part is one dictionary type, held as ir.Record, whose one index signature is a string's
//     of exactly unknown, with no named property, call or construct signature beside it: reads through it are Unions,
//     and every use of one narrows it by its run-time tag. Any other value type keeps castIndexSignature's NotYet, since
//     its reads are unboxed by a storage the value may not have;
//   - the source's non-nullish part is object, held as ir.Union. A source with fields of its own keeps #mft3zc6's
//     storage rule's stop, unknown keeps its refusal, and a dictionary already held as a record is an upcast here;
//   - undefined beside the object passes through only where the target keeps it, and null keeps the NotYet, since a
//     record's reference has no null of its own;
//   - .a keeps main's refusal (castStop);
//   - the cast's result is only ever read (dictionaryViewReadOnly): the record may be another alias's
//     Record<string, string>, and a number written through the view would be read there as a string. The test is by
//     use, on the cast's own value, never by type over the program, which no tsc unit passes (J-psxf2gy): a value
//     read out of the view flows on freely, and only the view itself is judged.
//
// The test at the cast (dictionaryViewValue) is that the value is a record whose values are references, which a
// Union reads as they are (union.c's keyed_record_value and its note at 433).
func (l *lowering) dictionaryView(node *ast.Node, source, target *checker.Type) bool {
	if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return false
	}
	if l.includesNull(source) || l.includesNull(target) || l.includesUndefined(source) && !l.includesUndefined(target) {
		return false
	}
	presentSource, present := l.checker.GetNonNullableType(source), l.checker.GetNonNullableType(target)
	if presentSource.Flags()&checker.TypeFlagsNonPrimitive == 0 {
		return false
	}
	for _, member := range castMembers(source) {
		if member.Flags() != checker.TypeFlagsUndefined && member != presentSource {
			return false
		}
	}
	for _, member := range castMembers(target) {
		if member.Flags() != checker.TypeFlagsUndefined && member != present {
			return false
		}
	}
	if present.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(present) || len(l.checker.GetPropertiesOfType(present)) != 0 {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(present)
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString || infos[0].ValueType().Flags() != checker.TypeFlagsUnknown {
		return false
	}
	if len(l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(present, checker.SignatureKindConstruct)) != 0 || l.recordElement(present) == nil {
		return false
	}
	if from, known := l.representation(source); !known || from != ir.Union {
		return false
	}
	if to, known := l.representation(target); !known || to != ir.Record {
		return false
	}
	return l.dictionaryViewReadOnly(node)
}

// dictionaryViewMemo is the use test's answers for one lowering, made the first time it's asked for: each cast's by
// its node, since the census lowers each unit many times and tsc has 22 casts, and each const's and followed
// parameter's by its declaration (dictionaryViewHeld).
type dictionaryViewMemo struct {
	casts        map[*ast.Node]bool
	declarations map[*ast.Node]bool
}

// dictionaryViewReadOnly is whether every use of the cast's result only reads the view (dictionaryViewUse), the
// answer kept by the cast's node. It reads the program as declared, never the instantiation being lowered when it's
// first asked, as neverRestOnly does, so the answer kept is the same wherever it's asked again.
func (l *lowering) dictionaryViewReadOnly(cast *ast.Node) bool {
	if l.dictionaryViewUses == nil {
		l.dictionaryViewUses = &dictionaryViewMemo{casts: map[*ast.Node]bool{}, declarations: map[*ast.Node]bool{}}
	}
	if answer, known := l.dictionaryViewUses.casts[cast]; known {
		return answer
	}
	outerMapper := l.typeMapper
	l.typeMapper = nil
	defer func() { l.typeMapper = outerMapper }()
	answer := l.dictionaryViewUse(cast)
	l.dictionaryViewUses.casts[cast] = answer
	return answer
}

// dictionaryViewUse is whether the view's value at node, the cast itself, a read of a const it initializes or a read
// of a parameter it's handed to, is used only where nothing writes through it, a parenthesized or non-null asserted
// use being the same use:
//   - an element read v[k] that is no assignment's, compound assignment's, increment's, destructuring's or loop's
//     target (ast.IsAssignmentTarget), no delete's operand and not called, which would hand the view on as this;
//   - the expression of a for...in;
//   - the receiver of an own-key test: hasOwnProperty.call(v, k), Object.prototype's own or a const alias of it
//     (detachedOwnCallReceiver), Object.hasOwn(v, k), or Object.keys(v) (dictionaryViewOwnKeyCall);
//   - an operand of === or !==, or only tested for truthiness (dictionaryViewCondition);
//   - the initializer of a const whose every read passes (dictionaryViewConst);
//   - an argument to a parameter whose every use in its callee passes (dictionaryViewParameter).
//
// Anything else answers no and keeps castIndexSignature's NotYet at the cast: returned, stored in a field or an
// element, assigned to a let or a var, captured, spread, handed to a callee that doesn't resolve or to a library
// function not named here.
//
// The mutant (M7, cast_stop.mutants.md): this answers true for every use, and TestDictionaryViewCasts' rows that
// return the view, store it in a const then written by key or hand it to a helper that writes go red, admitted where
// a write through MapLike<unknown> could put a number in a Record<string, string>.
func (l *lowering) dictionaryViewUse(node *ast.Node) bool {
	if dictionaryViewCondition(node) {
		return true
	}
	outer := dictionaryViewOuter(node)
	parent := outer.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindElementAccessExpression:
		if parent.AsElementAccessExpression().Expression != outer || ast.IsAssignmentTarget(parent) {
			return false
		}
		read := dictionaryViewOuter(parent)
		if read.Parent == nil {
			return false
		}
		return read.Parent.Kind != ast.KindDeleteExpression && (read.Parent.Kind != ast.KindCallExpression || read.Parent.AsCallExpression().Expression != read)
	case ast.KindForInStatement:
		return parent.AsForInOrOfStatement().Expression == outer
	case ast.KindBinaryExpression:
		operator := parent.AsBinaryExpression().OperatorToken.Kind
		return operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == outer && l.dictionaryViewConst(parent)
	case ast.KindCallExpression:
		if parent.AsCallExpression().Arguments == nil {
			return false
		}
		arguments := parent.AsCallExpression().Arguments.Nodes
		position := slices.Index(arguments, outer)
		if position < 0 {
			return false // The callee, v(...), which a dictionary of unknown values never is.
		}
		for _, argument := range arguments[:position] {
			if argument.Kind == ast.KindSpreadElement {
				return false // A spread before it moves it off the parameter its place names.
			}
		}
		if position == 0 && (l.detachedOwnCallReceiver(parent) != nil || l.dictionaryViewOwnKeyCall(parent)) {
			return true
		}
		if parent.Flags&ast.NodeFlagsOptionalChain != 0 {
			return false
		}
		return l.dictionaryViewParameter(parent, position)
	}
	return false
}

// dictionaryViewOuter is node with the parentheses and non-null assertions around it, each the same use as node.
func dictionaryViewOuter(node *ast.Node) *ast.Node {
	for node.Parent != nil && (node.Parent.Kind == ast.KindParenthesizedExpression || node.Parent.Kind == ast.KindNonNullExpression) {
		node = node.Parent
	}
	return node
}

// dictionaryViewCondition is whether node's value is only tested for truthiness: the operand of !, an if's, a
// while's or a do's condition, a conditional's, or either operand of && or || whose whole is one of these. An && or
// || used as a value may be the view itself and answers no.
func dictionaryViewCondition(node *ast.Node) bool {
	outer := dictionaryViewOuter(node)
	parent := outer.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == outer
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == outer
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == outer
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == outer
	case ast.KindBinaryExpression:
		operator := parent.AsBinaryExpression().OperatorToken.Kind
		return (operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken) && dictionaryViewCondition(parent)
	}
	return false
}

// dictionaryViewOwnKeyCall is whether a call is Object.hasOwn's or Object.keys's, through the library's Object, each
// of which reads its first argument's own keys and writes nothing.
func (l *lowering) dictionaryViewOwnKeyCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return false
	}
	return callee.Name().Text() == "hasOwn" || callee.Name().Text() == "keys"
}

// dictionaryViewConst is whether a const the view initializes is only read where dictionaryViewUse admits: a plain
// name, not exported, whose every read in the function or file declaring it passes, none from a function or class
// nested in it, which captures the view (dictionaryViewReads). An export specifier's name resolves to the const, as
// l.symbol resolves an alias, and is a use dictionaryViewUse doesn't name.
func (l *lowering) dictionaryViewConst(declaration *ast.Node) bool {
	if !ast.IsIdentifier(declaration.Name()) || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return false
	}
	own := l.symbol(declaration.Name())
	if own == nil {
		return false
	}
	scope := declaration.Parent
	for scope.Parent != nil && scope.Kind != ast.KindSourceFile && !ast.IsFunctionLike(scope) {
		scope = scope.Parent
	}
	return l.dictionaryViewHeld(declaration, func() bool { return l.dictionaryViewReads(scope, declaration.Name(), own, false) })
}

// dictionaryViewParameter is whether the view given at position to call goes into a parameter whose every use
// inside its callee passes dictionaryViewUse. The callee is resolved as neverRestParameter resolves it, with its
// rejections kept: the one function declaration with a body that an identifier or a property access's name names,
// never a closure, a method or a new, of the program's own files and taking no this; and a rest, destructured or
// defaulted parameter at position answers no. Inside it, one assigned, one read from a nested function or class and
// any read of arguments answer no (dictionaryViewReads).
func (l *lowering) dictionaryViewParameter(call *ast.Node, position int) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	var named *ast.Node
	switch callee.Kind {
	case ast.KindIdentifier:
		named = callee
	case ast.KindPropertyAccessExpression:
		named = callee.AsPropertyAccessExpression().Name()
	default:
		return false
	}
	symbol := l.symbol(named)
	if symbol == nil {
		return false
	}
	// An overload's argument lands in its implementation's parameter, the one body there is.
	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil {
			if implementation != nil {
				return false
			}
			implementation = declaration
		}
	}
	if implementation == nil || !neverRestProgramFile(ast.GetSourceFileOfNode(implementation)) || neverRestThisParameter(implementation) {
		return false
	}
	parameters := implementation.Parameters()
	if position >= len(parameters) {
		return false
	}
	parameter := parameters[position]
	if parameter.AsParameterDeclaration().DotDotDotToken != nil || parameter.AsParameterDeclaration().Initializer != nil || !ast.IsIdentifier(parameter.Name()) {
		return false
	}
	own := l.symbol(parameter.Name())
	if own == nil {
		return false
	}
	return l.dictionaryViewHeld(parameter, func() bool { return l.dictionaryViewReads(implementation, parameter.Name(), own, true) })
}

// dictionaryViewHeld answers a const's or a parameter's question once, by its declaration. It holds false while it's
// being answered, so a walk that meets the declaration again, a helper handing the view to itself or two handing it
// to each other, answers no: recursion fails closed. Every answer is kept, a no reached through a declaration still
// being answered included, since a yes never leans on one and a no only ever keeps the NotYet.
func (l *lowering) dictionaryViewHeld(declaration *ast.Node, answer func() bool) bool {
	if held, known := l.dictionaryViewUses.declarations[declaration]; known {
		return held
	}
	l.dictionaryViewUses.declarations[declaration] = false
	held := answer()
	l.dictionaryViewUses.declarations[declaration] = held
	return held
}

// dictionaryViewReads is whether every read of own, the binding name declares, under scope passes
// dictionaryViewUse, as neverRestParameterUses walks a followed parameter: a read from a function or class nested
// under scope captures the view and answers no, and so does one written. A parameter's walk asks arguments too,
// since a read of arguments reads every parameter without naming one.
func (l *lowering) dictionaryViewReads(scope, name *ast.Node, own *ast.Symbol, parameter bool) bool {
	var holds func(node *ast.Node, captured bool) bool
	holds = func(node *ast.Node, captured bool) bool {
		if ast.IsPartOfTypeNode(node) {
			return true
		}
		if ast.IsIdentifier(node) {
			if parameter && node.Text() == "arguments" && (node.Parent == nil || node.Parent.Kind != ast.KindPropertyAccessExpression || node.Parent.AsPropertyAccessExpression().Name() != node) {
				return false
			}
			if node == name || l.neverRestReferent(node) != own {
				return true
			}
			return !captured && !ast.IsAssignmentTarget(node) && l.dictionaryViewUse(node)
		}
		inner := captured || ast.IsFunctionLike(node) || ast.IsClassLike(node)
		every := true
		node.ForEachChild(func(child *ast.Node) bool {
			every = holds(child, inner)
			return !every
		})
		return every
	}
	every := true
	scope.ForEachChild(func(child *ast.Node) bool {
		every = holds(child, false)
		return !every
	})
	return every
}

// dictionaryViewValue lowers the cast dictionaryView proved, in a helper that takes the value once, as
// checkedClassCast's does: undefined passes through where the target keeps it, and anything else is a record whose
// values are references, as one checked narrow tests it on both backends alike, or a run-time NotYet in the words
// that say which test it failed (ir's DictionaryViewRecordStop and DictionaryViewReferencesStop). A fixed object, a
// class instance, an array, a function and a host value fail the first; a record of numbers or booleans the second.
func (l *lowering) dictionaryViewValue(node *ast.Node) (ir.Expression, error) {
	as := node.AsAsExpression()
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	value, err := l.expression(as.Expression)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Union {
		return nil, l.notYet(node, "a checked dictionary view of "+l.checker.TypeToString(l.concrete(l.checker.GetTypeAtLocation(node)))+" over a value of "+l.checker.TypeToString(source)+" held as other than a union")
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	if l.includesUndefined(source) {
		b.body = append(b.body, ir.If{Condition: ir.IsUndefined{Value: held}, Then: []ir.Statement{ir.Return{Value: ir.Undefined{Of: ir.Record}}}})
	}
	// The mutants (cast_stop.mutants.md): drop the two tests, an unchecked narrow, and oracle's
	// TestDictionaryViewStopsAtTheCast goes red on its array, literal and class rows; drop ReferencesMessage alone,
	// and on its number and boolean rows.
	return b.finish("dictionary_view", ir.Narrow{Value: held, To: ir.Record, Checked: true, Message: ir.DictionaryViewRecordStop, ReferencesMessage: ir.DictionaryViewReferencesStop}), nil
}
