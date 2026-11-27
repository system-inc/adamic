package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A spread copies own enumerable data only, so { ...new Sized() } has no weight where Sized's weight is a
// getter or a setter on its prototype, while the checker types the copy with weight whenever the source is
// seen through a type that declares it (#vjvcnnj). In .ts the spread stays, and a read of the name from the
// copy stops at compile time (spreadCopyRead for a getter, checkGenericSpreadCopyReads for a setter), where
// Node answers undefined. Head 3 lowers the reads Node answers honestly, by Compiler's acceptance on
// #vjvcnnj (Nov 25): a read that acceptsUndefined says is observed as it is, rather than relied on as
// present (typeof, a template span, === undefined, the left of ??, a ?. receiver, a conditional arm in such
// a place, a slot whose contextual type includes undefined), is a read by name that answers what Node does
// for any receiver. That is the run-time key read Reflect.get lowers to (ir.DynamicKeyedProperty, Reflect):
// the accessor's getter value when the receiver holds the accessor, undefined when it holds one with no
// getter, an own slot read by the kind the slot holds, and undefined otherwise. Its union is then narrowed
// to the field's type or undefined, a number's or a boolean's maybe, by a check of the tag it holds, so a
// slot of another kind, which only an unsound road can bring, is a loud stop and never a number read from
// the wrong kind (Compiler's hold 1). Every other read is lowered as before and stops as before, a narrowed
// second read among them (hold 3), since it is no observer. The read is no ir.Property, so neither stop
// judges it.
//
// The names are every accessor of a program class that some spread in the program may pass over, a
// setter-only one included, found by the spreads' source types; a library class's accessor is none, since
// every spread of a library instance stops at the spread today (hold 2, probe J-jtaqj20). A name some class
// or object literal holds as a method, or an object literal holds as an accessor, is left out: the run-time
// key read stops on either, where the plain read lowers. The read is .ts's alone, and its receiver isn't a
// class instance, so a program with no such spread, or a read through a class's own type, lowers as on main.

// spreadAccessorReadWords names the read in the stops checkDynamicKeyedGetters shares with Reflect.get.
const spreadAccessorReadWords = "a read of an accessor's name a spread's copy may lack, answered by name"

// spreadAccessorReadMismatch is the stop of a read whose slot holds another kind than the field's type says.
const spreadAccessorReadMismatch = "a read by name found a value of another kind than its field's type says"

// spreadAccessorNames is every name spreadAccessorRead may read by name, found once a lowering.
func (l *lowering) spreadAccessorNames() map[string]bool {
	if l.spreadAccessorNamesFound != nil {
		return l.spreadAccessorNamesFound
	}
	passed := map[string]bool{}
	excluded := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindMethodDeclaration:
			if name, named := memberKeyText(node); named {
				excluded[name] = true
			}
		case ast.KindGetAccessor, ast.KindSetAccessor:
			if name, named := memberKeyText(node); named && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
				excluded[name] = true
			}
		case ast.KindSpreadAssignment:
			members := map[string]prototypeMember{}
			l.spreadPassedMembers(l.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression), members)
			for name, member := range members {
				if member.kind == "getter" || member.kind == "setter" {
					passed[name] = true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range l.programFiles() {
		file.AsNode().ForEachChild(visit)
	}
	l.spreadAccessorNamesFound = map[string]bool{}
	for name := range passed {
		if !excluded[name] && !strings.HasPrefix(name, "#") && !strings.HasPrefix(name, "__#") {
			l.spreadAccessorNamesFound[name] = true
		}
	}
	return l.spreadAccessorNamesFound
}

// spreadPassedMembers adds to members the prototype members of each program class a spread of a value of type
// written may copy while its copy's type still claims them: every class assignable to a type parameter's
// constraint, whose copy is typed as what the type parameter stands for, and every class assignable to an
// object type that doesn't name it (hiddenClassMembers). A class instance's own type names none, since the
// checker types its copy without the class's accessors and methods (probe J-qy3z2fy).
func (l *lowering) spreadPassedMembers(written *checker.Type, members map[string]prototypeMember) {
	if l.mentionsTypeParameter(written, map[*checker.Type]bool{}) {
		constraint := l.checker.GetBaseConstraintOfType(written)
		for _, class := range l.declaredClasses() {
			if constraint == nil || l.checker.IsTypeAssignableTo(class, constraint) {
				l.prototypeMembers(class, members)
			}
		}
		return
	}
	if written.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range written.Types() {
			l.spreadPassedMembers(member, members)
		}
		return
	}
	if !isClassInstance(written) {
		l.hiddenClassMembers(written, members)
	}
}

// spreadAccessorRead is the read by name of the comment above, when node, a read of name from object, is
// one: handled says so, and every other read is propertyRead's as before.
func (l *lowering) spreadAccessorRead(node *ast.Node, object ir.Expression, name string) (ir.Expression, bool, error) {
	access := node.AsPropertyAccessExpression()
	receiver := ast.SkipParentheses(access.Expression)
	if object.Type() != ir.Object || l.questionDot(node) || receiver.Kind == ast.KindThisKeyword || receiver.Kind == ast.KindSuperKeyword {
		return nil, false, nil
	}
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") || !l.spreadAccessorNames()[name] || !l.acceptsUndefined(node) {
		return nil, false, nil
	}
	field := l.checker.GetSymbolAtLocation(node.Name())
	if field == nil || field.Flags&ast.SymbolFlagsOptional != 0 {
		return nil, false, nil
	}
	// A number's or a boolean's maybe is a type of its own, which every consumer reads. A string's is the string,
	// undefined a null reference, and a consumer keyed on the checker's string, a template span's, reads it as
	// present (read_title.ts ended in UBSan on a null string, J-65ewf3z), so a string read stays as it was.
	of, err := l.typeOf(node)
	if err != nil || (of != ir.Number && of != ir.Boolean) {
		return nil, false, nil
	}
	if stored, known := l.representation(l.checker.GetTypeOfSymbol(field)); !known || stored != of {
		return nil, false, nil
	}
	held := l.concrete(l.checker.GetTypeAtLocation(access.Expression))
	if isClassInstance(held) || !l.mayHideClassInstance(held) {
		return nil, false, nil
	}
	if err := l.checkDynamicKeyedGetterSyntax(node); err != nil {
		return nil, true, err
	}
	if l.reflectKeyedSite == nil {
		l.reflectKeyedSite = node
	}
	l.graphDynamicSite = node
	read := ir.DynamicKeyedProperty{Object: fit(object, ir.Union), Key: ir.StringConstant{Index: l.constant(name)}, Reflect: true}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	value := b.read(b.parameters[0])
	kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean"}[of]
	matches := ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(kind)}}, Right: ir.IsUndefined{Value: value}}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(spreadAccessorReadMismatch)}}}})
	result := b.finish("spread_accessor_read", ir.Narrow{Value: value, To: ir.Maybe(of)})
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return result, true, nil
}
