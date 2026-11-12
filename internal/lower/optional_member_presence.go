package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A presence test of an optional member a program's interface or type literal declares, `if (host.m)`
// or `typeof host.m`, whether the member is written as a method (m?(): T) or as a field holding a
// function (m?: () => T), is read off the receiver's slot as it runs (#p2n17c4 shape 1): an
// ir.Property with Absent, of a Closure, under its consumer, never folded. The consumers are
// presenceConsumer's, the ones that only test the value; || keeps it, and stays refused for a method.
//
// Natively the read looks at the object's own slots alone (adamic_object_optional_field), while
// JavaScript's o["m"] walks the prototype chain and runs a getter. So the two agree exactly where every
// object the receiver may hold is a plain object literal and its prototype, Object.prototype, has no
// member of that name. A literal's method shorthand, { m() {} }, is an own data slot holding its closure
// (objectLiteral's KindMethodDeclaration arm), so it reads as an arrow field does (Compiler's ruling 1
// and Lower's read on shape 1). checkOptionalMemberPresence holds every recorded read to that once the
// program is lowered. It's NotYet where:
//
//  1. the member's name is one Object.prototype holds (objectPrototypeName), which a literal without
//     its own inherits on Node (Lower's must on shape 1);
//  2. lowering didn't leave the read as exactly one such field read: finishAccessors made it a
//     dispatcher's call because a getter of that name may own the receiver, or the read was lowered
//     twice, as a generic instance's is, or never kept;
//  3. the receiver's reach is unknown: a closure's or an exported function's parameter, a for-of or
//     catch binding, a Map's value, a host value, anything the allocation graph doesn't follow;
//  4. some allocation it may reach isn't an object literal, or is a class instance (whose method is on
//     its prototype, and which also covers a class getter and an accessor literal, which accessorLiteral
//     registers as a class), a spread's copy, or record storage.
//
// A literal can't change its prototype to put a member there: a __proto__ key is refused in
// objectLiteral ("__proto__ in an object literal"), and Object.setPrototypeOf and Object.create in
// objectCallArguments (library_object.go). A store into a plain literal writes an own data slot, which
// reads the same on both backends, so stores aren't weighed; Object.defineProperty is refused but for a
// record entry, and record storage stops above.

// optionalPresenceRead is a presence test lowering read off its slot: the read, where it is, which is
// its ir.Property's ViewWhere, and the field's name.
type optionalPresenceRead struct {
	node  *ast.Node
	where string
	name  string
}

// optionalPresenceKey finds a recorded read's ir.Property again: its ViewWhere and its Name.
type optionalPresenceKey struct {
	where string
	name  string
}

// optionalMemberPresenceTest reports whether node reads an optional member only to test for it: every
// declaration of the member is a method or property signature with a question token in a file the
// program wrote, never TypeScript's library, Node's types or the prelude, and the member's type, once
// undefined is taken out, can be called. A class's optional method is a method declaration, so it and any
// merge holding one stay refused.
func (l *lowering) optionalMemberPresenceTest(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression || !presenceConsumer(node) {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodSignature && declaration.Kind != ast.KindPropertySignature || !ast.HasQuestionToken(declaration) {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if load.IsLibrary(file) || load.IsNodeLibrary(file) || load.IsPrelude(file) {
			return false
		}
	}
	return len(l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(symbol)), checker.SignatureKindCall)) > 0
}

// objectPrototypeName reports whether a plain object literal inherits a member named name. Which names
// the global Object interface declares is asked of the checker by symbol, so the bundled library says
// (constructor, toString, valueOf and the rest), and a program's own augmentation of Object counts too.
// Node's Object.prototype also holds __proto__ and the __defineGetter__ family, which the library
// doesn't declare; each begins with two underscores, so any such name counts, with no list here to keep.
// Where the checker has no Object interface to ask, every name counts.
func (l *lowering) objectPrototypeName(name string) bool {
	if strings.HasPrefix(name, "__") {
		return true
	}
	object := l.checker.GetGlobalSymbol("Object", ast.SymbolFlagsInterface, nil)
	if object == nil {
		return true
	}
	return l.checker.GetPropertyOfType(l.checker.GetDeclaredTypeOfSymbol(object), name) != nil
}

// checkOptionalMemberPresence holds each presence test lowering read off its slot to what the receiver
// may hold, as the comment above says.
func (l *lowering) checkOptionalMemberPresence() error {
	if len(l.optionalPresenceReads) == 0 {
		return nil
	}
	// A copy of the program numbers its allocation sites, as newAccessorReach's does, so the program
	// itself is left as lowering made it.
	numbered := *l.result
	numbered.Functions = append([]ir.Function(nil), l.result.Functions...)
	assignAllocationSites(&numbered)
	recorded := map[optionalPresenceKey]bool{}
	for _, read := range l.optionalPresenceReads {
		recorded[optionalPresenceKey{where: read.where, name: read.name}] = true
	}
	found := map[optionalPresenceKey][]ir.Property{}
	inspect := func(node any) bool {
		if read, isProperty := node.(ir.Property); isProperty && read.Absent && read.Of == ir.Closure {
			if key := (optionalPresenceKey{where: read.ViewWhere, name: read.Name}); recorded[key] {
				found[key] = append(found[key], read)
			}
		}
		return true
	}
	walk(numbered.Main, inspect)
	for _, function := range numbered.Functions {
		walk(function.Body, inspect)
	}
	graph := newAllocationFlowGraph(&numbered)
	judged := map[optionalPresenceKey]bool{}
	for _, read := range l.optionalPresenceReads {
		key := optionalPresenceKey{where: read.where, name: read.name}
		if judged[key] {
			continue
		}
		judged[key] = true
		tested := "a presence test of optional member " + read.name
		if l.objectPrototypeName(read.name) {
			return l.notYet(read.node, tested+", a name Object.prototype holds too, which Node finds on an object without its own where native reads own slots alone")
		}
		if len(found[key]) != 1 {
			return l.notYet(read.node, tested+" that lowering didn't leave as one field read")
		}
		reach := graph.ReachingAllocations(found[key][0].Object)
		if reach.Unknown {
			return l.notYet(read.node, tested+" whose receiver may hold a value the program didn't build as a plain object literal ("+strings.Join(reach.Reasons, "; ")+")")
		}
		records := graph.projectionIndex().records
		for _, site := range reach.Sites {
			literal, isLiteral := records[site]
			switch {
			case !isLiteral:
				return l.notYet(read.node, tested+" whose receiver may be an allocation that isn't an object literal")
			case literal.Class != 0:
				return l.notYet(read.node, tested+" whose receiver may be a class instance, whose method lives on its prototype, not an own slot")
			case literal.Spread != nil:
				return l.notYet(read.node, tested+" whose receiver may be an object a spread made")
			case literal.Record:
				return l.notYet(read.node, tested+" whose receiver may be record storage")
			}
		}
	}
	return nil
}
