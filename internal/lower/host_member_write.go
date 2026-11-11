package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// hostMemberWrite is a write recv.m = value whose m a program declares as a method of an interface or a
// type literal, which lowering let through as a field store (hostMemberWritable) and checkHostMemberWrites
// judges once lowering is done: target is the write, value what it writes, and site its store's
// ir.SetProperty.Site, the mark that finds the store again after assignAllocationSites renumbers every
// body.
type hostMemberWrite struct {
	target *ast.Node
	value  *ast.Node
	site   int
}

// hostMemberWritable says whether target, a property access, is the whole left side of a plain `=`
// statement and names a method member every declaration of which is a method signature of an interface
// or a type literal the program wrote, never TypeScript's library, Node's types or the prelude, and
// never optional. tsc patches its host interfaces' members so (program.ts:547, `host.readFile =
// fileName => {...}`), shape 2 of #p2n17c4. Such a write is a field store on a program object literal,
// whose method-syntax member and arrow property are both a field holding a closure, so the unbound-method
// rule (methodValueRefusal) and lowering's method-replacement stop (setProperty) let it through, and
// checkHostMemberWrites holds it to what's known about it once the program is lowered (#fj0vqtf, wave 1).
// A class's own method, a library's or a host's, a logical or compound assignment, and a write used as a
// value are untouched by this and stay refused where they were.
func (l *lowering) hostMemberWritable(target *ast.Node) bool {
	if target.Kind != ast.KindPropertyAccessExpression || target.AsPropertyAccessExpression().QuestionDotToken != nil || !ast.IsIdentifier(target.Name()) {
		return false
	}
	held := target
	for held.Parent != nil && held.Parent.Kind == ast.KindParenthesizedExpression {
		held = held.Parent
	}
	assignment := held.Parent
	if assignment == nil || assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken || assignment.AsBinaryExpression().Left != held {
		return false
	}
	if assignment.Parent == nil || assignment.Parent.Kind != ast.KindExpressionStatement {
		return false
	}
	member := l.checker.GetSymbolAtLocation(target)
	if member == nil || member.Flags&ast.SymbolFlagsMethod == 0 || member.Flags&ast.SymbolFlagsOptional != 0 {
		return false
	}
	roots := l.checker.GetRootSymbols(member)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			if declaration.Kind != ast.KindMethodSignature || declaration.Parent == nil || declaration.Parent.Kind != ast.KindInterfaceDeclaration && declaration.Parent.Kind != ast.KindTypeLiteral {
				return false
			}
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) || load.IsNodeLibrary(file) || load.IsPrelude(file) {
				return false
			}
		}
	}
	return true
}

// checkHostMemberWrites judges each write hostMemberWritable let through, on the shared allocation graph
// once lowering is done, since what reaches a receiver is known only then (Lower's design on #fj0vqtf,
// with Language's must and conditions, Adamic's ruling):
//
//  1. Reach: every allocation reaching the receiver is a program object literal, never a class's
//     instance, whose method is no own field, nor a record, nor something the graph can't follow
//     (ReachingAllocations' Unknown: a parameter of a closure, an opaque call, a host value), which fails
//     closed. A literal with an accessor is a class to lowering (accessorLiteral), so a member it
//     implements with a get or set pair, which Node runs on the write, stops here too (shape 1's ground).
//  2. Shape: each such literal holds m as a field of its own, and isn't made with a spread.
//  3. Holders: the value fits m in every type each literal is held as (entriesRecordHandles, #ny0c3gb's
//     handles: an annotated variable, a parameter, a property, an element, an argument, a return, an
//     `as` target), judged as #9dmh4em judges a method seen as a writable field. tsc's view rule relates a
//     method covariantly, so a literal held as `Narrow { read(): string }` can be seen as
//     `Wide { read(): string | undefined }`, and a write through Wide would put undefined where every
//     call through Narrow reads a string (Language's h1). A literal held nowhere more specific than
//     itself is held as its own type, as entriesRecordHandles holds it, and one whose holders can't be
//     named fails closed.
//
// A write to an object that may be frozen is an ir.SetProperty like any other, so it throws Node's
// TypeError where Object.freeze reached it (ir.Program's FrozenWriteMayThrow). Each stop is NotYet at
// the write, in its own words.
func (l *lowering) checkHostMemberWrites() error {
	if len(l.hostMemberWrites) == 0 {
		return nil
	}
	assignAllocationSites(l.result)
	stores := map[int]ir.SetProperty{}
	inspect := func(node any) bool {
		if store, isStore := node.(ir.SetProperty); isStore && store.Site > 0 {
			stores[store.Site] = store
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	graph := newAllocationFlowGraph(l.result)
	for _, write := range l.hostMemberWrites {
		// A store a later pass reshaped can't be judged, so it fails closed.
		store, lowered := stores[write.site]
		if !lowered {
			return l.notYet(write.target, "a write to method "+write.target.Name().Text()+" of "+sourceExpression(write.target.AsPropertyAccessExpression().Expression)+" whose store the lowered program no longer holds as written")
		}
		reaches := graph.ReachingAllocations(store.Object)
		if err := l.hostMemberReach(write, store, reaches, graph.projectionIndex()); err != nil {
			return err
		}
		if err := l.hostMemberLiterals(write, store, reaches, graph.projectionIndex()); err != nil {
			return err
		}
	}
	return nil
}

// hostMemberWriting is how a stop at a host member write begins: the member and the receiver it's
// written through, as the source spells it.
func (l *lowering) hostMemberWriting(write hostMemberWrite, store ir.SetProperty) string {
	return "a write to method " + store.Name + " of " + sourceExpression(write.target.AsPropertyAccessExpression().Expression)
}

// hostMemberReach is checkHostMemberWrites' first rule: every allocation reaching the receiver is a plain
// program object literal, and an unknown reach fails closed.
func (l *lowering) hostMemberReach(write hostMemberWrite, store ir.SetProperty, reaches ir.AllocationSet, index *allocationProjection) error {
	if reaches.Unknown {
		return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object not known to be one of the program's object literals")
	}
	for _, site := range reaches.Sites {
		literal, isLiteral := index.records[site]
		switch {
		case !isLiteral:
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold something other than an object literal")
		case literal.Class > 0 && literal.Class <= len(l.result.Classes) && l.result.Classes[literal.Class-1].Literal:
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object literal with an accessor, whose setter a write may run")
		case literal.Class != 0:
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an instance of a class, whose methods aren't fields of its own")
		case literal.Record:
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold a record")
		}
	}
	return nil
}

// hostMemberLiterals is checkHostMemberWrites' second and third rules, over each plain object literal
// reaching the receiver (hostMemberReach stopped every other site): the literal holds the member as a field
// of its own, and the value fits the member in every type the literal is held as.
func (l *lowering) hostMemberLiterals(write hostMemberWrite, store ir.SetProperty, reaches ir.AllocationSet, index *allocationProjection) error {
	value := l.checker.GetTypeAtLocation(write.value)
	for _, site := range reaches.Sites {
		literal, isLiteral := index.records[site]
		if !isLiteral || literal.Class != 0 || literal.Record {
			continue
		}
		if literal.Spread != nil {
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object literal made with a spread")
		}
		var held ir.Expression
		for _, field := range literal.Fields {
			if field.Name == store.Name && !field.Absent && !field.Uninitialized {
				held = field.Value
			}
		}
		if held == nil {
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object literal that doesn't hold "+store.Name+" as a field of its own")
		}
		node := l.hostMemberLiteralNode(held)
		if node == nil {
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object literal whose "+store.Name+" isn't a function written in it, so the types it's held as can't be named")
		}
		holders := l.entriesRecordHandles(node)
		if len(holders) == 0 {
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object literal whose holding types can't be named")
		}
		for _, holder := range holders {
			if err := l.hostMemberHolder(write, store, value, holder); err != nil {
				return err
			}
		}
	}
	return nil
}

// hostMemberLiteralNode is the object literal a literal's field value was written in, found through the
// closure it holds: a method-syntax member's function, or an arrow's or a function expression's written as
// a property's whole initializer. Any other value names no literal, and is nil.
func (l *lowering) hostMemberLiteralNode(held ir.Expression) *ast.Node {
	made, isMade := held.(ir.MakeClosure)
	if !isMade {
		return nil
	}
	declaration := l.functionNodes[made.Function]
	if declaration == nil {
		return nil
	}
	var literal *ast.Node
	switch declaration.Kind {
	case ast.KindMethodDeclaration:
		literal = declaration.Parent
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		initializer := declaration
		for initializer.Parent != nil && initializer.Parent.Kind == ast.KindParenthesizedExpression {
			initializer = initializer.Parent
		}
		if initializer.Parent == nil || initializer.Parent.Kind != ast.KindPropertyAssignment || initializer.Parent.AsPropertyAssignment().Initializer != initializer {
			return nil
		}
		literal = initializer.Parent.Parent
	}
	if literal == nil || literal.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	return literal
}

// hostMemberHolder judges the written value against the member as holder, one type a reaching literal is
// held as, sees it: each of holder's members that has the member, a union's included, must hold the value,
// by the field rule #9dmh4em uses for a method seen as a writable field. A member that doesn't have it
// never reads it. An any or unknown holder can't be judged, and fails closed.
func (l *lowering) hostMemberHolder(write hostMemberWrite, store ir.SetProperty, value, holder *checker.Type) error {
	members := []*checker.Type{holder}
	if holder.Flags()&checker.TypeFlagsUnion != 0 {
		members = holder.Types()
	}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return l.notYet(write.target, l.hostMemberWriting(write, store)+", which may hold an object literal held as "+l.checker.TypeToString(member)+", whose "+store.Name+" can't be judged")
		}
		property := l.checker.GetPropertyOfType(member, store.Name)
		if property == nil {
			continue
		}
		slot := l.checker.GetTypeOfSymbol(property)
		if !l.enumAssignable(value, slot) || !l.checker.IsTypeAssignableTo(value, slot) {
			return l.notYet(write.target, l.hostMemberWriting(write, store)+" of a value of type "+l.checker.TypeToString(value)+", which the object, also held as "+l.checker.TypeToString(member)+", can't hold: its "+store.Name+" there is "+l.checker.TypeToString(slot))
		}
	}
	return nil
}
