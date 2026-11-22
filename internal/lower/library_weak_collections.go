package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A WeakMap or a WeakSet is held as ir.WeakCollection (representation), its own type, never a Map's:
// a table keyed by each key's weak handle (runtime/weak_collection.c), so an entry holds its key
// without keeping it alive. Only the new site and the methods below make or reach one. A WeakSet is
// whole; a WeakMap stops at weakMapNotYet until its get and set are built (#hcgy751 W2), and the
// admission of its values (which programs may hold one at all, the ephemeron condition) is Lower's
// (#q2ybjr1).
//
// A key is an object, never a primitive: Node throws a TypeError for a primitive key ("Invalid value
// used as weak map key", "Invalid value used in weak set"). tsc lets a primitive through a key type it
// fits, lib's Object or String interface or any shape a primitive satisfies (WeakSet<String> takes
// 'abc' without a cast), so a key type a primitive fits stops by name here, at the type, and no
// primitive ever reaches the runtime's adamic_weak_key (Runtime's ruling 1 on #hcgy751). Every other
// road to a primitive key stops before it: a cast the runtime can't check, any, and a symbol, which
// has no representation.

// weakCollectionKey is the representation a WeakMap's or a WeakSet's keys are held as, or the stop.
// A key is held as itself: an object, an array, a Map or a Set, a function or a typed array, each a
// counted heap value whose address the runtime makes a weak handle for, or a union of object types
// (the object keyword), whose reference is the object's own pointer, never a box. It's admitted only
// where no primitive fits the type (noPrimitiveFits), Map's identity rule for a key held as a union
// (keyableType), asked of every key type here, since a key type a primitive fits structurally is held
// as an object or a union all the same. Anything else, a type parameter, unknown, a dictionary or a
// promise, fails closed.
func (l *lowering) weakCollectionKey(node *ast.Node, collection string, key *checker.Type) (ir.Type, error) {
	if !l.noPrimitiveFits(key) {
		// An object type a primitive fits stops in its own words, whatever it's held as: lib's String has an
		// index signature and no representation at all, and Object is held as a union.
		if l.onlyObjects(l.concrete(key), 0) {
			return 0, l.notYet(node, "a "+collection+" keyed by "+l.checker.TypeToString(key)+", which a primitive fits (Node throws a TypeError for a primitive key; type the keys as objects)")
		}
		return 0, l.notYet(node, "a "+collection+" keyed by "+l.checker.TypeToString(key)+" (keys are objects, arrays, maps, sets, functions or typed arrays so far)")
	}
	held, known := l.representation(key)
	switch {
	case !known:
	case held == ir.Object, held == ir.Array, held == ir.Map, held == ir.Closure, held == ir.Union, held.IsTypedArray():
		return held, nil
	}
	return 0, l.notYet(node, "a "+collection+" keyed by "+l.checker.TypeToString(key)+" (keys are objects, arrays, maps, sets, functions or typed arrays so far)")
}

// weakMapNotYet is where a WeakMap stops until its get and set are built: its new and every method.
func (l *lowering) weakMapNotYet(node *ast.Node) error {
	return l.notYet(node, "a WeakMap (its get and set aren't built yet, #hcgy751)")
}

// newWeakCollection lowers new WeakSet(), with no argument; new WeakMap() stops at weakMapNotYet.
func (l *lowering) newWeakCollection(node *ast.Node) (ir.Expression, error) {
	created := node.AsNewExpression()
	if l.isLibraryGlobal(created.Expression, "WeakMap") {
		return nil, l.weakMapNotYet(node)
	}
	arguments := l.typeArguments(l.intrinsicReceiverType(node))
	if len(arguments) != 1 {
		return nil, l.notYet(node, "a WeakSet whose key type isn't known")
	}
	key, err := l.weakCollectionKey(node, "WeakSet", arguments[0])
	if err != nil {
		return nil, err
	}
	if created.Arguments != nil && len(created.Arguments.Nodes) > 0 {
		return nil, l.notYet(created.Arguments.Nodes[0], "new WeakSet from an iterable")
	}
	return ir.WeakCollectionNew{Key: key, Set: true}, nil
}

// weakCollectionMethod lowers a WeakSet's add, has and delete. handled is false for any other name,
// which is left to be refused as a call stage 0 doesn't know; a WeakMap's every method stops at
// weakMapNotYet. The receiver's key type is judged again here, as at its new: a WeakSet seen as one of
// another key type is invariance's to stop (invariance.go's containers), and this keeps a key type a
// primitive fits out of every road that reaches a method without a new.
func (l *lowering) weakCollectionMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	proven := l.intrinsicReceiverType(receiver)
	if !l.isLibraryType(proven, "WeakSet") {
		return nil, true, l.weakMapNotYet(node)
	}
	if name != "add" && name != "has" && name != "delete" {
		return nil, false, nil
	}
	arguments := l.typeArguments(proven)
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, "a WeakSet whose key type isn't known")
	}
	key, err := l.weakCollectionKey(node, "WeakSet", arguments[0])
	if err != nil {
		return nil, true, err
	}
	set, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 {
		return nil, true, l.notYet(node, "weakSet."+name+" with other than one key")
	}
	value, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	if value = fit(value, key); value.Type() != key {
		return nil, true, l.notYet(written[0], "weakSet."+name+" with a key held otherwise than the set's keys")
	}
	switch name {
	case "add":
		return ir.WeakSetAdd{Set: set, Key: value, KeyType: key}, true, nil
	case "has":
		return ir.WeakCollectionHas{Collection: set, Key: value, KeyType: key}, true, nil
	}
	return ir.WeakCollectionDelete{Collection: set, Key: value, KeyType: key}, true, nil
}
