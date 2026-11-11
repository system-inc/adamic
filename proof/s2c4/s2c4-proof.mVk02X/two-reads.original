package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A generator stored in a local (#qk8rztp S2c-4-1), the form tsc's checker.ts:21755-21759 writes:
// `const children = generateJsxChildren(...)`, then `children` handed on in the next statement. A
// `const` declared in a block of a function that doesn't suspend, with no type written and a
// generator's direct call as its initializer, one the roads already admit (top-level, nested or
// written in place), lowers when the checker finds exactly one other reference to it and that
// reference is a consumer the roads admit: a for-of over it, or the one-read Iterable<T> argument
// of a one-body callee (iterableArgument). No closure may name it, an unused capture included, and
// the read may not sit inside a loop, a function or a class that doesn't also hold the declaration
// (Compiler's must), so it runs at most once per run of the declaration. Making a generator runs none
// of its body, only its arguments, which run at the declaration either way, so the only new thing
// is the frame waiting in the local between its declaration and its read (Adamic's ruling).
//
// The local owns the frame from its declaration, and the one read is a move, never a retain (Memory's
// ruling): natively it takes the frame out of the local and leaves `NULL` there, at the read
// (native's moveOut), so a path that never reaches the read, an untaken branch, a throw, a return
// or a break, still finds the frame in the local, and the local's release at its scope's end drops
// it unstepped, running none of the body and so no finally, as Node collects one. A read moved into
// an argument and then left by a later argument's throw is the statement's temporary, released
// unstepped on that throw.
//
// No reachLinks edge is owed for this head (Memory's ruling): the local can't be captured, and its one
// read hands the frame to a consumer that steps it. That holds while the callee road refuses a callee
// that keeps its Iterable parameter, which generatorParameter does: the parameter's one read in the
// lowered program must be the view's for-of over it (or, stepped by hand, only its nexts and its
// undefined tests), no closure may capture it and nothing may write it, so it reaches no field, no
// captured cell and no returned object.

// storedGeneratorNotYet stops a generator stored in a local that this head doesn't lower, in Lower's
// words for every such use: two reads, typeof, a test, `===`, a let, a capture, a field, a return,
// destructuring or a spread of the local, and a read inside a loop or a function the declaration
// isn't in.
func (l *lowering) storedGeneratorNotYet(node *ast.Node) error {
	return l.notYet(node, "a generator stored in a local, used other than once by a consumer that steps it (#qk8rztp)")
}

// storedGeneratorCallee is the generator function a declaration's initializer makes, when it is a
// generator's direct call by name or one written in place, the two callees generatorCall takes. Any
// other callee is left to the initializer's own lowering, which stops a generator's call there, so
// the checker is never asked for the signature of a call this head can't take.
func (l *lowering) storedGeneratorCallee(initializer *ast.Node) (int, bool, error) {
	call := ast.SkipParentheses(initializer)
	if call.Kind != ast.KindCallExpression {
		return 0, false, nil
	}
	if callee := ast.SkipParentheses(call.AsCallExpression().Expression); !ast.IsIdentifier(callee) && !generatorInPlace(callee) {
		return 0, false, nil
	}
	return l.generatorCallee(call)
}

// storedGenerator lowers a declaration whose initializer is a generator's direct call
// (storedGeneratorCallee said function makes it), inside a function that doesn't suspend: the frame made into a local of its
// own, held as a Generator, when the declaration and its one read are S2c-4-1's (storedGeneratorRead),
// and a stop in Lower's words otherwise. The local has no checker type for the cycle finder, as a
// for-of's own generator local has none (TestNoSlotTheIteratorArmWalksHoldsAGenerator): a type can't
// tell a generator declared SetIterator from a Set's own iterator.
func (l *lowering) storedGenerator(list, declaration *ast.Node, function int) ([]ir.Statement, error) {
	name := declaration.Name()
	statement := list.Parent
	// A const, one plain name, no type written that could make it a view, and declared as a statement
	// of a block: never a for's own declaration, never a case clause's, whose binding the switch's
	// shared environment holds (switchBindings).
	if list.Flags&ast.NodeFlagsConst == 0 || declaration.Type() != nil || declaration.AsVariableDeclaration().ExclamationToken != nil ||
		statement == nil || statement.Kind != ast.KindVariableStatement || statement.Parent == nil || statement.Parent.Kind != ast.KindBlock {
		return nil, l.storedGeneratorNotYet(name)
	}
	if err := l.storedGeneratorRead(declaration, statement.Parent); err != nil {
		return nil, err
	}
	made, err := l.generatorCall(declaration.AsVariableDeclaration().Initializer, function)
	if err != nil {
		return nil, err
	}
	// The local a parent with nested functions preallocated for its hoisted closures, or a new one.
	local, err := l.declareLocal(name)
	if err != nil {
		return nil, err
	}
	if declared := l.result.Locals[local]; declared.Global || declared.Captured || declared.Ready != 0 || declared.Function != l.functionIndex {
		return nil, l.storedGeneratorNotYet(name)
	}
	l.result.Locals[local].Type = ir.Generator
	delete(l.localTypes, local)
	delete(l.localAlso, local)
	if l.instance != nil {
		l.instance.templates = slices.DeleteFunc(l.instance.templates, func(kept template) bool { return !kept.isClosure && kept.local == local })
	}
	if l.storedGenerators == nil {
		l.storedGenerators = map[int]int{}
	}
	l.storedGenerators[local] = function
	return []ir.Statement{ir.Declare{Local: local, Value: made}}, nil
}

// storedGeneratorRead holds a stored generator's declaration to its one read: every reference the
// checker finds to the name in scope, the block the declaration is a statement of, counts, a type's
// `typeof` and a shorthand property among them, so there must be exactly one, after the declaration,
// and it must be a for-of over the local or the one-read Iterable<T> argument of a one-body callee.
// Between it and that block it may cross no function, class or static block, which would capture the
// local, and no loop but the for-of whose expression it is, which reads it once: a read in a loop the
// declaration isn't in would run on a second pass after the first moved the frame out (Compiler's
// must, the two-pass program).
func (l *lowering) storedGeneratorRead(declaration, scope *ast.Node) error {
	name := declaration.Name()
	symbol := l.symbol(name)
	if symbol == nil {
		return l.storedGeneratorNotYet(name)
	}
	var references []*ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node != name {
			target := l.symbol(node)
			if node.Parent != nil && ast.IsShorthandPropertyAssignment(node.Parent) {
				target = l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
			}
			if target == symbol {
				references = append(references, node)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	scope.ForEachChild(visit)
	if len(references) != 1 {
		return l.storedGeneratorNotYet(name)
	}
	read := references[0]
	if read.Pos() < declaration.End() {
		return l.storedGeneratorNotYet(read)
	}
	site := read
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	consumer := site.Parent
	steps := consumer != nil && consumer.Kind == ast.KindForOfStatement && consumer.AsForInOrOfStatement().Expression == site && consumer.AsForInOrOfStatement().AwaitModifier == nil
	if !steps && !l.iterableArgument(read) {
		return l.storedGeneratorNotYet(read)
	}
	for at := site.Parent; at != scope; at = at.Parent {
		switch {
		case at == nil:
			// The read lies outside the block that declares it, which a const's scope can't be.
			return l.storedGeneratorNotYet(read)
		case steps && at == consumer:
			// The for-of evaluates the expression it steps once, before its first pass.
		case ast.IsFunctionLike(at), ast.IsClassLike(at), at.Kind == ast.KindClassStaticBlockDeclaration, ast.IsIterationStatement(at, false):
			return l.storedGeneratorNotYet(read)
		}
	}
	return nil
}

// storedGeneratorLocal is the stored generator's local a read names, parentheses aside, and the
// generator function that made it, when node is the read of one (storedGenerator).
func (l *lowering) storedGeneratorLocal(node *ast.Node) (int, int, bool) {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) || l.storedGenerators == nil {
		return 0, 0, false
	}
	local, isLocal := l.locals[l.symbol(node)]
	if !isLocal {
		return 0, 0, false
	}
	function, isStored := l.storedGenerators[local]
	return local, function, isStored
}

// generatorMade is what a consumer steps: the generator's call, which makes its frame there, or since
// S2c-4-1 the one read of the local a stored generator waits in, which moves the frame out of it
// (storedGeneratorMove).
func (l *lowering) generatorMade(node *ast.Node, function int) (ir.Expression, error) {
	if local, _, isStored := l.storedGeneratorLocal(node); isStored {
		// Touched as any read is, so a read from another function would be marked captured, which
		// checkGeneratorRoads says out loud; storedGeneratorRead admits none.
		l.local(ast.SkipParentheses(node))
		return ir.Read{Local: local, Of: ir.Generator}, nil
	}
	return l.generatorCall(node, function)
}

// storedGeneratorMove is the one read of a stored generator's local (#qk8rztp S2c-4-1), a read of a
// Generator local in the only two places lowering puts one besides a step's or a done test's operand:
// the value a for-of's own generator local is declared from, and IterableOf's value. Each is a move,
// which the native backend makes at the read (moveOut), so it needs no IR of its own.
func storedGeneratorMove(value ir.Expression) (ir.Read, bool) {
	read, isRead := value.(ir.Read)
	return read, isRead && read.Of == ir.Generator
}
