package lower

import (
	"reflect"
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
// a place, a slot whose contextual type includes undefined), and that one of those stops would meet today,
// judged as they judge it by the receiver's reach, is a read by name that answers what Node does for any
// receiver. That is the run-time key read Reflect.get lowers to (ir.DynamicKeyedProperty, Reflect):
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
//
// The reach that says whether a stop would meet a read is the whole program's, which lowering hasn't built
// when it lowers the read, and by the time it has, the read's consumers are built around its type. So a read
// that passes the gates above is lowered as main lowers it, as the field's type or undefined, and recorded
// (spreadAccessorRead); once the program is lowered, spreadAccessorReads asks each stop's judgment of it and
// reads by name only those a stop would meet. Every other read, a plain literal's own weight under typeof
// among them, keeps main's ir.Property, held as a present number or boolean of the maybe type its consumers
// take, and never takes on the run-time key read's program-wide getter judgment (checkDynamicKeyedGetters)
// or a graph program's stop on a dynamic read.

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

// spreadAccessorRecord is a read spreadAccessorRead lowered as main does, wrapped so its consumers take the
// field's type or undefined, which spreadAccessorReads turns into the read by name where a stop would meet it.
type spreadAccessorRecord struct {
	node    *ast.Node
	object  ir.Expression
	name    string
	of      ir.Type
	lowered ir.Expression
}

// spreadAccessorRead lowers node, a read of name from object, when it's one the comment above may read by name:
// as main lowers it (propertyRead, its branch here skipped), wrapped as the field's type or undefined, and
// recorded for spreadAccessorReads. handled says so; every other read is propertyRead's as before.
func (l *lowering) spreadAccessorRead(node *ast.Node, object ir.Expression, name string) (ir.Expression, bool, error) {
	if l.spreadAccessorPlain == node || node.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
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
	outer := l.spreadAccessorPlain
	l.spreadAccessorPlain = node
	plain, err := l.propertyRead(node, object, name)
	l.spreadAccessorPlain = outer
	if err != nil || plain.Type() != of {
		return plain, true, err
	}
	lowered := fit(plain, ir.Maybe(of))
	l.spreadAccessorRecords = append(l.spreadAccessorRecords, spreadAccessorRecord{node: node, object: object, name: name, of: of, lowered: lowered})
	return lowered, true, nil
}

// spreadAccessorReads reads by name each recorded read that a stop would meet today, judged by its receiver's
// reach as the stops judge it (spreadAccessorStops), and leaves every other one as main lowers it, so a read
// that lowers on main, a plain literal's own weight under typeof among them, lowers as it did. It runs once
// finishClassCalls has recorded what a graph reads, and before certifySourceSlots and the stops, so each of
// them sees the read it would have seen: main's read, or the read by name, which is no ir.Property. A read
// whose getters the run-time key read can't take by their syntax, or whose lowered expression a pass since
// changed, stays main's and stops as it did.
func (l *lowering) spreadAccessorReads() {
	if len(l.spreadAccessorRecords) == 0 {
		return
	}
	reach := l.newAccessorReach(l.result)
	copied, failure := l.spreadCopiedGetters()
	var stopped []int
	for index, record := range l.spreadAccessorRecords {
		if failure == nil && l.spreadAccessorStops(reach, copied, record) && l.checkDynamicKeyedGetterSyntax(record.node) == nil {
			stopped = append(stopped, index)
		}
	}
	if len(stopped) == 0 {
		return
	}
	// The read by name is made where its read is found, so a record a rolled-back lowering left, whose read isn't in
	// the program, makes no helper and names no site. A read is found by its value: two records whose lowered
	// expressions are equal read the same name from the same receiver, so the stops judge them alike and either's
	// read by name is the other's. Only a node of the wrapper's Go type can be one (fit's ir.MaybeOf), so no other
	// node is compared.
	byName := map[int]ir.Expression{}
	wrappers := map[reflect.Type]bool{}
	for _, index := range stopped {
		wrappers[reflect.TypeOf(l.spreadAccessorRecords[index].lowered)] = true
	}
	rewrite := func(node any) any {
		if !wrappers[reflect.TypeOf(node)] {
			return node
		}
		for _, index := range stopped {
			record := l.spreadAccessorRecords[index]
			if !reflect.DeepEqual(node, record.lowered) {
				continue
			}
			if byName[index] == nil {
				byName[index] = l.spreadAccessorByName(record)
				if l.reflectKeyedSite == nil {
					l.reflectKeyedSite = record.node
				}
				l.graphDynamicSite = record.node
			}
			return byName[index]
		}
		return node
	}
	l.result.Main = rewriteShapeStatements(l.result.Main, rewrite)
	// A helper the rewrite makes is appended after the bodies it walks, and holds no recorded read.
	count := len(l.result.Functions)
	for index := 0; index < count; index++ {
		body := rewriteShapeStatements(l.result.Functions[index].Body, rewrite)
		l.result.Functions[index].Body = body
	}
}

// spreadAccessorStops is whether a stop would meet record's read as main lowers it: spreadCopyRead's, for a getter
// a spread's copy may lack (spread_copy_reads.go), or #r8mazrh's, for a member a copy whose source may hide a
// program class may lack, a setter's included, by checkGenericSpreadCopyReads' rule (generic_spread_copy_reads.go):
// a receiver whose allocations may hold such a copy, and one whose reach the graph can't close while any spread in
// the program may copy such a member.
func (l *lowering) spreadAccessorStops(reach *accessorReach, copied map[string]string, record spreadAccessorRecord) bool {
	if l.spreadCopyRead(reach, copied, ir.Property{Object: record.object, Name: record.name}, record.name) != nil {
		return true
	}
	if len(l.genericSpreadSites) == 0 {
		return false
	}
	records := reach.graph.projectionIndex().records
	// As checkGenericSpreadCopyReads' plainSource: a spread that isn't generic whose source the graph proves is only
	// plain literals the program built hides no class.
	plainSource := func(literal ir.ObjectLiteral) bool {
		if l.genericSpreadSites[literal.SpreadSite] == genericSpreadCopy {
			return false
		}
		sources := reach.graph.ReachingAllocations(literal.Spread)
		if sources.Unknown || len(sources.Sites) == 0 {
			return false
		}
		for _, site := range sources.Sites {
			source, built := records[site]
			if !built || source.Class != 0 || source.Spread != nil {
				return false
			}
		}
		return true
	}
	// lacks is whether literal's copy may lack the member, as that pass's judge says, a getter from a spread that
	// isn't generic left to spreadCopyRead (leftToSpreadCopyRead).
	lacks := func(literal ir.ObjectLiteral) bool {
		copy := l.genericSpreadSites[literal.SpreadSite]
		if literal.Spread == nil || copy == "" || plainSource(literal) {
			return false
		}
		member, left := l.genericSpreadMembers(literal)[record.name]
		return left && !(member.kind == "getter" && copy == hiddenClassSpreadCopy)
	}
	allocations := reach.graph.ReachingAllocations(record.object)
	if allocations.Unknown {
		found := false
		collect := func(node any) bool {
			if literal, ok := node.(ir.ObjectLiteral); ok && !found && lacks(literal) {
				found = true
			}
			return !found
		}
		walk(l.result.Main, collect)
		for _, function := range l.result.Functions {
			walk(function.Body, collect)
		}
		return found
	}
	for _, site := range allocations.Sites {
		if literal, ok := records[site]; ok && lacks(literal) {
			return true
		}
	}
	return false
}

// spreadAccessorByName is record's read by name: the run-time key read of its name, through a helper that checks
// the tag it gives before narrowing it to the field's type or undefined.
func (l *lowering) spreadAccessorByName(record spreadAccessorRecord) ir.Expression {
	read := ir.DynamicKeyedProperty{Object: fit(record.object, ir.Union), Key: ir.StringConstant{Index: l.constant(record.name)}, Reflect: true}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	value := b.read(b.parameters[0])
	kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean"}[record.of]
	matches := ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(kind)}}, Right: ir.IsUndefined{Value: value}}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(spreadAccessorReadMismatch)}}}})
	result := b.finish("spread_accessor_read", ir.Narrow{Value: value, To: ir.Maybe(record.of)})
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return result
}
