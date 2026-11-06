package lower

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A .ts file's ambient value declaration is a host binding where internal/load/host_bindings.json names
// it, and lowers by its row (#00qn5az, Adamic's ruling; docs/rulings.md's entry "A `.ts` file's ambient
// value declaration is a host binding"). A declarator matches a row only by its name, its file's
// basename, its keyword and its declared type's exact text, in a top-level `declare` statement of a .ts
// file that isn't exported; anything else stays refused as an ambient value declaration (ambient.go),
// and so does every .a file's. A read is matched to the row by the symbol it resolves to, never by its
// name, so a local that shadows the declared name is judged as the local.
//
// One IR feeds every target, so a row lowers only when its three columns are equal and none is pending
// its owner's clearance; any other row is not yet, named (hostBindingStop), at its declaration and at
// every use. Of the rows today only onProfilerEvent lowers, Absent on every target.
//
// Absent: typeof answers the constant "undefined", which every backend compares as it compares any
// string, so a guard's test is false at run time with no folding. The name itself is never read. A read
// is admitted only under one of three dominating shapes whose guard is `typeof X === "t"` with t not
// "undefined", or `typeof X !== "undefined"` (hostBindingGuard): the then-branch of if, the right of &&
// and the true arm of ?:. Everything else is not yet (hostBindingRefusal): ||, an early return, a negated
// guard, a guard bound to a local, a guard in a called function. The guarded branch, arm or operand is
// never lowered: it becomes the unreachable placeholder (hostBindingDeadRead), a loud compiler-bug panic,
// with neverValue's typed placeholder after it where a value is owed, never a fabricated initialized
// local. The placeholder rather than the else branch, because the test still runs as written, so the
// placeholder is reached only if the test itself was wrong, and then it stops loudly rather than taking
// a branch on a guess. Nothing in the program may define an Absent name: a write to it, and a
// write naming any row through the global object, `globalThis.<name>` or `globalThis[name]` by any name
// for it, is refused.
//
// Provided rows (process, performance) are pending today, so no read routes through a host model yet;
// their model is named in the row for when their columns clear.

// hostBindingNamed is the row for a name, or nil; a table that doesn't parse names nothing here, and
// ambientRefusal reports its error at the first ambient value declaration.
func hostBindingNamed(name string) *load.HostBinding {
	rows, err := load.HostBindings()
	if err != nil {
		return nil
	}
	for index := range rows {
		if rows[index].Name == name {
			return &rows[index]
		}
	}
	return nil
}

// hostBindingRow is the row a declarator is the declaration of, or nil.
func (l *lowering) hostBindingRow(declaration *ast.Node) *load.HostBinding {
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) || declaration.Initializer() != nil {
		return nil
	}
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Parent == nil || list.Parent.Kind != ast.KindVariableStatement {
		return nil
	}
	statement := list.Parent
	if statement.Parent == nil || statement.Parent.Kind != ast.KindSourceFile || !ast.HasSyntacticModifier(statement, ast.ModifierFlagsAmbient) || ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
		return nil
	}
	file := ast.GetSourceFileOfNode(declaration)
	name := file.FileName().AsString()
	if path.Ext(name) != ".ts" || strings.HasSuffix(name, ".d.ts") {
		return nil
	}
	row := hostBindingNamed(declaration.Name().Text())
	if row == nil || row.Declaration.File != path.Base(name) {
		return nil
	}
	keyword := ""
	switch list.Flags & ast.NodeFlagsBlockScoped {
	case ast.NodeFlagsConst:
		keyword = "const"
	case ast.NodeFlagsLet:
		keyword = "let"
	}
	written := declaration.AsVariableDeclaration().Type
	if keyword != row.Declaration.Keyword || written == nil || file.Text()[scanner.GetTokenPosOfNode(written, file, false):written.End()] != row.Declaration.Type {
		return nil
	}
	return row
}

// hostBindingStatement reports whether every declarator of an ambient variable statement is a host
// binding's declaration, and if so, the stop for the first whose row doesn't lower, or nil.
func (l *lowering) hostBindingStatement(statement *ast.Node) (bool, error) {
	if _, err := load.HostBindings(); err != nil {
		return true, err
	}
	var stop error
	for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
		row := l.hostBindingRow(declaration)
		if row == nil {
			return false, nil
		}
		if reason := hostBindingStop(*row); stop == nil && reason != "" {
			stop = l.notYet(declaration, reason)
		}
	}
	return true, stop
}

// hostBindingDeclared reports whether a statement declares only host bindings that lower: it declares
// nothing, since the host's names are the host's.
func (l *lowering) hostBindingDeclared(statement *ast.Node) bool {
	if statement.Kind != ast.KindVariableStatement || !ambientDeclaration(statement) {
		return false
	}
	matched, err := l.hostBindingStatement(statement)
	return matched && err == nil
}

// hostBindingStop is why a row doesn't lower, said as what stage 0 can't lower yet, or "" when it does.
func hostBindingStop(row load.HostBinding) string {
	var waits []string
	for _, target := range row.Targets() {
		if target.Column.Answer == load.HostBindingPending {
			waits = append(waits, possessive(target.Column.Owner)+" clearance of its "+target.Target+" column")
		}
	}
	if row.Model.Path == "" && row.Model.Task != "" {
		waits = append(waits, possessive(row.Model.Task)+" host model")
	}
	if len(waits) > 0 {
		said := waits[len(waits)-1]
		if len(waits) > 1 {
			said = strings.Join(waits[:len(waits)-1], ", ") + " and " + said
		}
		return "the host binding " + row.Name + ", pending " + said + ","
	}
	if row.Native.Answer != row.JavaScript.Answer || row.Native.Answer != row.WASM32.Answer {
		return "the host binding " + row.Name + ", whose targets answer differently (native " + string(row.Native.Answer) + ", javascript " + string(row.JavaScript.Answer) + ", wasm32 " + string(row.WASM32.Answer) + ") where one IR feeds every target, until an IR presence node is ruled,"
	}
	if row.Native.Answer == load.HostBindingProvided {
		return "the host binding " + row.Name + "'s reads through its host model, " + row.Model.Path + ","
	}
	return ""
}

// possessive is an owner's name with its possessive ending: Runtime's, Platforms'.
func possessive(owner string) string {
	if strings.HasSuffix(owner, "s") {
		return owner + "'"
	}
	return owner + "'s"
}

// hostBindingOf is the row an identifier reads, by the symbol it resolves to, and that symbol; nil when
// it reads anything else.
func (l *lowering) hostBindingOf(node *ast.Node) (*load.HostBinding, *ast.Symbol) {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return nil, nil
	}
	named := hostBindingNamed(node.Text())
	if named == nil {
		return nil, nil
	}
	var symbol *ast.Symbol
	switch parent := node.Parent; {
	case parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node:
		symbol = l.checker.GetShorthandAssignmentValueSymbol(parent)
	case parent != nil && parent.Kind == ast.KindExportSpecifier:
		symbol = l.checker.GetExportSpecifierLocalTargetSymbol(parent)
	default:
		symbol = l.checker.GetSymbolAtLocation(node)
	}
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil
	}
	row := l.hostBindingRow(symbol.Declarations[0])
	if row == nil {
		return nil, nil
	}
	return row, symbol
}

// hostBindingAbsent reports whether a row lowers as Absent on every target.
func hostBindingAbsent(row load.HostBinding) bool {
	return hostBindingStop(row) == "" && row.Native.Answer == load.HostBindingAbsent
}

// hostBindingGuard is the symbol of the Absent host binding a condition's typeof test proves missing:
// `typeof X === "t"` with t not "undefined", or `typeof X !== "undefined"`, through parentheses. Any
// other condition, a negation, a local holding the test or a call that makes it, proves nothing.
func (l *lowering) hostBindingGuard(condition *ast.Node) *ast.Symbol {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := condition.AsBinaryExpression()
	test, answer := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	if test.Kind != ast.KindTypeOfExpression || answer.Kind != ast.KindStringLiteral && answer.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return nil
	}
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken:
		if answer.Text() == "undefined" {
			return nil
		}
	case ast.KindExclamationEqualsEqualsToken:
		if answer.Text() != "undefined" {
			return nil
		}
	default:
		return nil
	}
	row, symbol := l.hostBindingOf(test.AsTypeOfExpression().Expression)
	if row == nil || !hostBindingAbsent(*row) {
		return nil
	}
	return symbol
}

// hostBindingDominated reports whether a read sits in the then-branch of an if, the right of an && or
// the true arm of a ?: whose guard proves its binding missing, within the read's own function.
func (l *lowering) hostBindingDominated(read *ast.Node, symbol *ast.Symbol) bool {
	for child, parent := read, read.Parent; parent != nil && !ast.IsFunctionLike(parent) && parent.Kind != ast.KindClassStaticBlockDeclaration; child, parent = parent, parent.Parent {
		var guard *ast.Node
		switch parent.Kind {
		case ast.KindIfStatement:
			if parent.AsIfStatement().ThenStatement == child {
				guard = parent.AsIfStatement().Expression
			}
		case ast.KindBinaryExpression:
			if parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Right == child {
				guard = parent.AsBinaryExpression().Left
			}
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().WhenTrue == child {
				guard = parent.AsConditionalExpression().Condition
			}
		}
		if guard != nil && l.hostBindingGuard(guard) == symbol {
			return true
		}
	}
	return false
}

// hostBindingRefusal judges every use of a host binding before lowering: a write is refused, a row that
// doesn't lower is not yet at each use, and an Absent name's read is admitted only at typeof or where a
// guard proves it dead (hostBindingDominated).
func (l *lowering) hostBindingRefusal(node *ast.Node) error {
	if !ast.IsIdentifier(node) {
		return nil
	}
	row, symbol := l.hostBindingOf(node)
	if row == nil || symbol.Declarations[0].Name() == node || ast.IsPartOfTypeNode(node) {
		return nil
	}
	if ast.IsAssignmentTarget(node) {
		return &Refused{Where: l.program.Where(node), What: "a write to the host binding " + row.Name, Fix: hostBindingWriteFix}
	}
	if stop := hostBindingStop(*row); stop != "" {
		return l.notYet(node, stop)
	}
	operand := node
	for operand.Parent != nil && operand.Parent.Kind == ast.KindParenthesizedExpression {
		operand = operand.Parent
	}
	if operand.Parent != nil && operand.Parent.Kind == ast.KindTypeOfExpression {
		return nil
	}
	if !l.hostBindingDominated(node, symbol) {
		return l.notYet(node, "a read of "+row.Name+", a host binding absent on every target, outside the then-branch of if, the right of && or the true arm of ?: under a typeof guard that proves it dead,")
	}
	return nil
}

// hostBindingWriteFix is what a refused write to a host binding is told.
const hostBindingWriteFix = "a host binding is the host's to define, and nothing in the program may; keep the value in a variable of your own"

// hostBindingGlobalThisRefusal refuses, at the write, an assignment, update or delete through the global
// object whose key names a host binding or may: globalThis.onProfilerEvent = f would define an Absent
// name behind its folded typeof. The receiver is judged by its type, typeof globalThis, through whatever
// name reads it: globalThis, an alias of it, Node's global, globalThis.globalThis, and through a union or
// intersection holding it, a type parameter constrained to it (globalObject) or an operand that is it
// where the checker's type for the whole forgot it (globalObjectOperand). It runs before every
// other rule, so the stop is this ruling's, not a rule's that happens to meet the receiver or delete
// first. A receiver widened to another type (object, a record view) is no longer typeof globalThis; the
// stops that judge those today are pinned in TestHostBindingFoldRestsOnTheseStops.
func (l *lowering) hostBindingGlobalThisRefusal(node *ast.Node) error {
	var target *ast.Node
	switch node.Kind {
	case ast.KindBinaryExpression:
		if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			target = node.AsBinaryExpression().Left
		}
	case ast.KindPrefixUnaryExpression:
		if operator := node.AsPrefixUnaryExpression().Operator; operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			target = node.AsPrefixUnaryExpression().Operand
		}
	case ast.KindPostfixUnaryExpression:
		target = node.AsPostfixUnaryExpression().Operand
	case ast.KindDeleteExpression:
		target = node.AsDeleteExpression().Expression
	}
	if target == nil {
		return nil
	}
	target = ast.SkipParentheses(target)
	var receiver *ast.Node
	name, known := "", false
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		receiver, name, known = target.AsPropertyAccessExpression().Expression, target.AsPropertyAccessExpression().Name().Text(), true
	case ast.KindElementAccessExpression:
		receiver = target.AsElementAccessExpression().Expression
		if key := l.checker.GetTypeAtLocation(target.AsElementAccessExpression().ArgumentExpression); key.Flags()&checker.TypeFlagsStringLiteral != 0 {
			name, known = key.AsLiteralType().Value().(string)
		}
	default:
		return nil
	}
	receiver = ast.SkipParentheses(receiver)
	if !l.globalObject(receiver) && !l.globalObjectOperand(receiver) {
		return nil
	}
	if known && hostBindingNamed(name) == nil {
		return nil
	}
	what := "a write through globalThis to the host binding " + name
	if !known {
		what = "a write through globalThis whose key may name a host binding"
	}
	return &Refused{Where: l.program.Where(target), What: what, Fix: hostBindingWriteFix}
}

// globalObject reports whether an expression's type may be typeof globalThis, whatever name the
// expression reads it by: the checker's own globalThis symbol is that type's symbol. It fails closed
// (globalObjectType): a union or intersection holding the global object, and a type parameter whose
// constraint is it, may be the global object too, though the type itself has no such symbol (Compiler's
// read of b5a2f3f3).
func (l *lowering) globalObject(node *ast.Node) bool {
	global := l.checker.GetGlobalSymbol("globalThis", ast.SymbolFlagsModule, nil)
	return global != nil && l.globalObjectType(l.checker.GetTypeAtLocation(node), global, map[*checker.Type]bool{})
}

// globalObjectOperand reports whether an expression may evaluate to an operand of it that is the global
// object, where the checker's type for the whole no longer says so: (flag ? globalThis : other) is other's
// type alone when other's is a supertype of typeof globalThis (both carry process), since a conditional's
// union is reduced by subtype. So the arms of ?:, the operands of ||, && and ??, the right of a comma or an
// assignment, and what a cast or a non-null assertion holds are each judged as the global object too.
func (l *lowering) globalObjectOperand(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	var operands []*ast.Node
	switch node.Kind {
	case ast.KindConditionalExpression:
		operands = []*ast.Node{node.AsConditionalExpression().WhenTrue, node.AsConditionalExpression().WhenFalse}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			operands = []*ast.Node{binary.Left, binary.Right}
		case ast.KindCommaToken, ast.KindEqualsToken:
			operands = []*ast.Node{binary.Right}
		}
	case ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
		operands = []*ast.Node{node.Expression()}
	}
	for _, operand := range operands {
		if l.globalObject(operand) || l.globalObjectOperand(operand) {
			return true
		}
	}
	return false
}

// globalObjectType reports whether a type is the global object's, a union or intersection any member
// of which is, or an instantiable type (a type parameter, an indexed access, a conditional) whose base
// constraint is.
func (l *lowering) globalObjectType(proven *checker.Type, global *ast.Symbol, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	seen[proven] = true
	if proven.Symbol() == global {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range proven.Types() {
			if l.globalObjectType(member, global, seen) {
				return true
			}
		}
	}
	if proven.Flags()&checker.TypeFlagsInstantiable != 0 {
		return l.globalObjectType(l.checker.GetBaseConstraintOfType(proven), global, seen)
	}
	return false
}

func init() {
	refusalRules.register(50, "hostBindingGlobalThisRefusal", (*lowering).hostBindingGlobalThisRefusal)
	refusalRules.register(2250, "hostBindingRefusal", (*lowering).hostBindingRefusal)
}

// hostBindingDeadRead is the unreachable placeholder for what a guard proves dead: the guard's test
// compares typeof's constant "undefined" at run time, so this runs only if that was wrong.
func (l *lowering) hostBindingDeadRead(symbol *ast.Symbol) ir.Statement {
	return ir.Panic{Message: ir.StringConstant{Index: l.constant("compiler bug: the typeof guard of " + symbol.Name + ", a host binding absent on every target, was taken")}}
}

// hostBindingTypeOf lowers typeof of a host binding: "undefined" for an Absent one.
func (l *lowering) hostBindingTypeOf(node *ast.Node) (ir.Expression, bool, error) {
	row, _ := l.hostBindingOf(node.AsTypeOfExpression().Expression)
	if row == nil {
		return nil, false, nil
	}
	if stop := hostBindingStop(*row); stop != "" {
		return nil, true, l.notYet(node, stop)
	}
	return ir.StringConstant{Index: l.constant("undefined")}, true, nil
}

// hostBindingAnd lowers guard && right, whose right a guard proves dead, as a value: the guard, and the
// placeholder where right would be. Only a boolean result is held; another stops as not yet.
func (l *lowering) hostBindingAnd(node *ast.Node) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		return nil, false, nil
	}
	symbol := l.hostBindingGuard(binary.Left)
	if symbol == nil {
		return nil, false, nil
	}
	left, err := l.expression(binary.Left)
	if err != nil {
		return nil, true, err
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	if of != ir.Boolean {
		return nil, true, l.notYet(node, "the value of an && whose right a host binding's typeof guard proves dead, held as a "+typeName(of))
	}
	return ir.Binary{Operator: ir.And, Left: left, Right: ir.Effects{Body: []ir.Statement{l.hostBindingDeadRead(symbol)}, Result: unreachablePlaceholder(of)}}, true, nil
}

// hostBindingAndStatement lowers guard && right as a statement, as if (guard) right.
func (l *lowering) hostBindingAndStatement(node *ast.Node) ([]ir.Statement, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		return nil, false, nil
	}
	symbol := l.hostBindingGuard(binary.Left)
	if symbol == nil {
		return nil, false, nil
	}
	condition, err := l.condition(binary.Left)
	if err != nil {
		return nil, true, err
	}
	return []ir.Statement{ir.If{Condition: condition, Then: []ir.Statement{l.hostBindingDeadRead(symbol)}}}, true, nil
}

// hostBindingConditional lowers guard ? whenTrue : whenFalse, whose true arm the guard proves dead: the
// placeholder there, and the false arm as it is.
func (l *lowering) hostBindingConditional(node *ast.Node, condition ir.Expression, symbol *ast.Symbol) (ir.Expression, error) {
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	whenNot, err := l.expression(node.AsConditionalExpression().WhenFalse)
	if err != nil {
		return nil, err
	}
	whenTrue := ir.Effects{Body: []ir.Statement{l.hostBindingDeadRead(symbol)}, Result: unreachablePlaceholder(of)}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: fit(whenNot, of), Of: of}, nil
}
