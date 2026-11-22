package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A WeakMap or a WeakSet is held as ir.WeakCollection (representation), its own type, never a Map's:
// a table keyed by each key's weak handle (runtime/weak_collection.c), so an entry holds its key
// without keeping it alive. Only the new site and the methods below make or reach one. A WeakSet is
// whole. A WeakMap holds numbers, booleans and strings so far (weakCollectionAdmitted, Runtime's ruling 2
// on #hcgy751): the admission of every other value (which programs may hold one at all, the ephemeron
// condition) is Lower's (#q2ybjr1).
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

// weakCollectionAdmitted is the representation a WeakMap's values are held as, or the stop. A value held
// as a plain number, boolean or string is a scalar or a string, which holds no reference back to any key,
// so no entry can keep its own key alive through its value (Runtime's ruling 2 on #hcgy751: scalar values
// only, first cut). A value that may be missing (string | null, number | undefined) stops in the words a
// Map whose values it can't hold stops in (mapTypes), and every other value, an object, an array, a
// closure, a union, or anything the representation can't name, waits on Lower's admission (#q2ybjr1),
// which replaces this one function. Unknown fails closed.
func (l *lowering) weakCollectionAdmitted(node *ast.Node, value *checker.Type) (ir.Type, error) {
	if l.includesNull(value) || l.includesUndefined(value) {
		return 0, l.notYet(node, "a WeakMap of "+l.checker.TypeToString(value))
	}
	if held, known := l.representation(value); known && (held == ir.Number || held == ir.Boolean || held == ir.String) {
		return held, nil
	}
	return 0, l.notYet(node, "a WeakMap holding "+l.checker.TypeToString(value)+" (values other than numbers, booleans and strings wait on what a WeakMap may hold, #q2ybjr1)")
}

// weakCollectionTypes is the representations a WeakMap's or a WeakSet's keys and values are held as,
// read from its type, or the stop: value is 0 for a WeakSet, which keeps none.
func (l *lowering) weakCollectionTypes(node *ast.Node, proven *checker.Type, collection string) (key ir.Type, value ir.Type, err error) {
	arguments := l.typeArguments(proven)
	if collection == "WeakSet" {
		if len(arguments) != 1 {
			return 0, 0, l.notYet(node, "a WeakSet whose key type isn't known")
		}
	} else if len(arguments) != 2 {
		return 0, 0, l.notYet(node, "a WeakMap whose key and value types aren't known")
	}
	if key, err = l.weakCollectionKey(node, collection, arguments[0]); err != nil {
		return 0, 0, err
	}
	if collection == "WeakMap" {
		if value, err = l.weakCollectionAdmitted(node, arguments[1]); err != nil {
			return 0, 0, err
		}
	}
	return key, value, nil
}

// newWeakCollection lowers new WeakMap() and new WeakSet(), with no argument.
func (l *lowering) newWeakCollection(node *ast.Node) (ir.Expression, error) {
	created := node.AsNewExpression()
	collection := "WeakSet"
	if l.isLibraryGlobal(created.Expression, "WeakMap") {
		collection = "WeakMap"
	}
	key, value, err := l.weakCollectionTypes(node, l.intrinsicReceiverType(node), collection)
	if err != nil {
		return nil, err
	}
	if created.Arguments != nil && len(created.Arguments.Nodes) > 0 {
		return nil, l.notYet(created.Arguments.Nodes[0], "new "+collection+" from an iterable")
	}
	return ir.WeakCollectionNew{Key: key, Value: value, Set: collection == "WeakSet"}, nil
}

// weakCollectionMethod lowers a WeakSet's add, has and delete, and a WeakMap's get, set, has and delete.
// handled is false for any other name, which is left to be refused as a call stage 0 doesn't know. The
// receiver's key and value types are judged again here, as at its new: a WeakSet or a WeakMap seen as one
// of other type arguments is invariance's to stop (invariance.go's containers), and this keeps a key type
// a primitive fits, and a value not yet admitted, out of every road that reaches a method without a new.
func (l *lowering) weakCollectionMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	proven := l.intrinsicReceiverType(receiver)
	collection, written := "", "weakSet."
	switch {
	case l.isLibraryType(proven, "WeakSet"):
		if name != "add" && name != "has" && name != "delete" {
			return nil, false, nil
		}
		collection = "WeakSet"
	case l.isLibraryType(proven, "WeakMap"):
		if name != "get" && name != "set" && name != "has" && name != "delete" {
			return nil, false, nil
		}
		collection, written = "WeakMap", "weakMap."
	default:
		return nil, true, l.notYet(node, "a WeakMap or WeakSet whose type isn't known")
	}
	key, value, err := l.weakCollectionTypes(node, proven, collection)
	if err != nil {
		return nil, true, err
	}
	held, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	want, holder := 1, "the set's"
	if collection == "WeakMap" {
		holder = "the map's"
	}
	if name == "set" {
		want = 2
	}
	if len(arguments) != want {
		if name == "set" {
			return nil, true, l.notYet(node, "weakMap.set with other than a key and a value")
		}
		return nil, true, l.notYet(node, written+name+" with other than one key")
	}
	keyed, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if keyed = fit(keyed, key); keyed.Type() != key {
		return nil, true, l.notYet(arguments[0], written+name+" with a key held otherwise than "+holder+" keys")
	}
	switch name {
	case "add":
		return ir.WeakSetAdd{Set: held, Key: keyed, KeyType: key}, true, nil
	case "get":
		return ir.WeakMapGet{Map: held, Key: keyed, KeyType: key, ValueType: value}, true, nil
	case "set":
		stored, err := l.expression(arguments[1])
		if err != nil {
			return nil, true, err
		}
		if stored = fit(stored, value); stored.Type() != value {
			return nil, true, l.notYet(arguments[1], "weakMap.set with a value held otherwise than the map's values")
		}
		return ir.WeakMapSet{Map: held, Key: keyed, Value: stored, KeyType: key, ValueType: value, Site: l.writeSite(receiver)}, true, nil
	case "has":
		return ir.WeakCollectionHas{Collection: held, Key: keyed, KeyType: key}, true, nil
	}
	return ir.WeakCollectionDelete{Collection: held, Key: keyed, KeyType: key}, true, nil
}
