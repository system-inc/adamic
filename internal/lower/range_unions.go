package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// rangeUnion is what an IR expression does with an object operand it holds, for a closed range type held as
// a Union where a NodeArray may stand for it (range_types.go, #hae423v). Such a Union may hold an adamic_array
// where its type promises an object, so every road that takes an object operand either asks the kind before
// it reads a layout or stops (Lower's Must 2 on #hae423v): a classification guard, not a sample.
// range_unions_test.go finds every expression type in internal/ir's source with an operand and fails on one
// rangeUnionOf doesn't class, so a new one can't be forgotten, as view_reads_test.go does for checked views.
//
// Lower's roads, by the IR each lowers to: a spread is an ObjectLiteral's or a RecordLiteral's Spread;
// Object.keys, values, entries and assign are an ObjectCall or an ObjectKeys, and so is a for-in, whose keys
// are an ObjectKeys or a record's forInKeys; JSON.stringify is a JSONStringify and encodeJson a JSONEncode; `in`
// is a HasProperty; === is a Binary; ?. is a Property's Optional or a CallClosure's; a compound write is a
// Property read and a SetProperty, each classed where it is (checkRangeUnionReads walks SetProperty, a
// statement, beside the table); a record value is a RecordLiteral's entry or a RecordCall's argument, and a Map
// value a MapSet's; a generic is a Call. A view or a cast to a wider type, satisfies included, makes no IR of
// its own: the operand goes where the wider type is held, which castValue's sameKeeping stops when the
// representation changes, and rangeWhereObjectGoes stops at any other value site.
type rangeUnion uint8

const (
	// rangeUnionDispatches asks the kind at run time on both backends before it reads any layout of its
	// operand: a dynamic read, typeof, a checked narrow, a brand test.
	rangeUnionDispatches rangeUnion = iota + 1
	// rangeUnionMoves reads no object's layout: it moves its operand, compares it by reference, tests its
	// presence or truth, or takes no object operand at all (a string, a number, an array, a map), each held
	// as its own kind, which a range type's Union never is without a narrow this table classes.
	rangeUnionMoves
	// rangeUnionStops reads its object operand's layout raw, so it's NotYet where that operand is held as a
	// range type's Union (checkRangeUnionReads). A read or a write of pos or end through a range type never
	// hands a Property or a SetProperty the Union: lowering makes it a dispatch (range_fields.go), whose arms
	// read and write through checked narrows to an object and to an array, so the field read and write roads
	// dispatch while the IR's Property and SetProperty, handed such a Union by any other road, still stop.
	rangeUnionStops
)

// rangeUnionOf is the one table: each IR expression with an operand, by its class, which
// checkRangeUnionReads reads. Zero is an expression nobody has classed. A class can turn on the value: an
// object literal reads a layout only through a spread, a narrow asks the kind only when checked, and
// Object.is only compares.
func rangeUnionOf(expression ir.Expression) rangeUnion {
	switch value := expression.(type) {
	case ir.Narrow:
		if value.Checked {
			return rangeUnionDispatches
		}
		if value.To == ir.Object || value.To == ir.Array {
			return rangeUnionStops
		}
		return rangeUnionMoves
	case ir.ObjectLiteral:
		if value.Spread != nil {
			return rangeUnionStops
		}
		return rangeUnionMoves
	case ir.RecordLiteral:
		if value.Spread != nil {
			return rangeUnionStops
		}
		return rangeUnionMoves
	case ir.ObjectCall:
		if value.Method == "is" {
			return rangeUnionMoves
		}
		return rangeUnionStops
	case ir.HasProperty:
		// An in test asks the kind on both backends: native's has_property (union.c) switches on the heap kind
		// and answers an array's own names from its index, its side record and its methods, and JavaScript's in
		// handles an array. So a NodeArray's in test (node-array-keys.a's 'pos' in nodes) lowers because its
		// reader asks, never because its boxed operand was followed.
		return rangeUnionDispatches
	}
	switch expression.(type) {
	case ir.ArrayIsArray, ir.CheckedCast, ir.DynamicKeyedProperty, ir.DynamicProperty, ir.ErrorToString,
		ir.IsCollectionIterator, ir.IsGenerator, ir.IsRegExpGroups, ir.JSONStringify, ir.ObjectTag, ir.OwnFunction,
		ir.TypeOf, ir.UnionToString, ir.UnknownKind, ir.ViewElement:
		return rangeUnionDispatches
	case ir.BoundMethod, ir.HasAccessor, ir.HasOwn, ir.InstanceOf, ir.JSONEncode, ir.ObjectKeys,
		ir.Property, ir.RecordStorage:
		return rangeUnionStops
	case ir.ArrayConcat, ir.ArrayFill, ir.ArrayFrom, ir.ArrayIndex, ir.ArrayJoin, ir.ArrayLiteral, ir.ArrayMap,
		ir.ArrayOfScalars, ir.ArrayPop, ir.ArrayPush, ir.ArrayReduce, ir.ArrayReverse, ir.ArraySearch, ir.ArraySlice,
		ir.ArraySort, ir.ArraySplice, ir.ArrayVisit, ir.Await, ir.Binary, ir.BooleanToString, ir.Box, ir.Call,
		ir.CallClosure, ir.CharCodeAt, ir.Coalesce, ir.CodePoints, ir.CollectionIterator, ir.Comma, ir.Concat,
		ir.Conditional, ir.Defined, ir.Effects, ir.FileStatus, ir.FunctionLength, ir.GeneratorDone, ir.GeneratorNext,
		ir.IntlCollatorCompareValue, ir.IntlCollatorNew, ir.IsNull, ir.IsUndefined, ir.IterableOf, ir.IteratorNext,
		ir.JSONDecode, ir.JoinEnter, ir.Length, ir.Logical, ir.LogicalAssignment, ir.MakeError, ir.MapClear,
		ir.MapDelete, ir.MapEntries, ir.MapForEach, ir.MapGet, ir.MapHas, ir.MapKeys, ir.MapNew, ir.MapSet,
		ir.MapSize, ir.MapValues, ir.MathCall, ir.MaybeOf, ir.MaybeToString, ir.NodeBufferCall, ir.NodeFSFile,
		ir.NodeHostCall, ir.NumberCall, ir.NumberFormat, ir.NumberToString, ir.ParallelMap, ir.PhantomMember,
		ir.PlaceholderUse, ir.ProcessCall, ir.PromiseValue, ir.ReadDirectory, ir.ReadTextFile, ir.RecordCall,
		ir.RecordCoalesce, ir.RegExpCall, ir.RegExpGroup, ir.RegExpNew, ir.RegExpProperty, ir.SetAdd, ir.SetNew,
		ir.SetValues, ir.StringCall, ir.StringFromCodes, ir.StringIndex, ir.StringIterator, ir.StringLength,
		ir.ToFixed, ir.Trim, ir.Truthy, ir.TypedArrayFill, ir.TypedArrayNew, ir.TypedArraySet, ir.TypedArraySort,
		ir.TypedArraySubarray, ir.Unary, ir.Unwrap, ir.Utf8At, ir.Utf8Length, ir.Void, ir.WeakCollectionDelete,
		ir.WeakCollectionHas, ir.WeakMapGet, ir.WeakMapSet, ir.WeakOf, ir.WeakSetAdd, ir.WeakTarget, ir.WriteTextFile,
		ir.Yield:
		return rangeUnionMoves
	}
	return 0
}

// rangeUnionStopped is the operands a stopping expression reads the layout of, which checkRangeUnionReads
// asks about: nil for any other expression.
func rangeUnionStopped(expression ir.Expression) []ir.Expression {
	if rangeUnionOf(expression) != rangeUnionStops {
		return nil
	}
	switch value := expression.(type) {
	case ir.Narrow:
		return []ir.Expression{value.Value}
	case ir.ObjectLiteral:
		return []ir.Expression{value.Spread}
	case ir.RecordLiteral:
		return []ir.Expression{value.Spread}
	case ir.ObjectCall:
		return value.Arguments
	case ir.BoundMethod:
		return []ir.Expression{value.Object}
	case ir.HasAccessor:
		return []ir.Expression{value.Object}
	case ir.HasOwn:
		return []ir.Expression{value.Object}
	case ir.HasProperty:
		return []ir.Expression{value.Object}
	case ir.InstanceOf:
		return []ir.Expression{value.Value}
	case ir.JSONEncode:
		return []ir.Expression{value.Value}
	case ir.ObjectKeys:
		return []ir.Expression{value.Object}
	case ir.Property:
		return []ir.Expression{value.Object}
	case ir.RecordStorage:
		return []ir.Expression{value.Value}
	}
	return nil
}

// checkRangeUnionReads is the table's pass, once lowering is done: every stopping expression, and every field
// write (a SetProperty, a statement), whose operand is held as a Union that may be a range type's value, is
// NotYet, at the function or the statement of Main holding it (stopWhere), since the IR keeps no node for it. A
// program with no NodeArray layout, or none of whose types was held as a range type's Union, has no such operand,
// and the pass reads nothing.
func (l *lowering) checkRangeUnionReads() error {
	if !l.rangeUnionsHeld || !l.rangeArraysPresent() {
		return nil
	}
	operands := &rangeOperands{l: l}
	var refused error
	inspect := func(where string, site stopSite) func(node any) bool {
		return func(node any) bool {
			if refused != nil {
				return false
			}
			var stopped []ir.Expression
			road := ""
			switch node := node.(type) {
			case ir.SetProperty:
				stopped, road = []ir.Expression{node.Object}, "a write of field "+node.Name
			case ir.Expression:
				stopped, road = rangeUnionStopped(node), rangeUnionRoad(node)
			}
			for _, operand := range stopped {
				if operand != nil && operand.Type() == ir.Union && operands.held(operand) {
					// Probe only, never for main (#nxb3xh1): a stop inside a helper is traced (range_unions_trace.go).
					if l.firstLinkTraps && site.function >= 0 && site.function < len(l.functionNodes) && l.functionNodes[site.function] == nil {
						l.rangeTrace(operands, site.function, operand, road, site)
					}
					refused = &NotYet{Where: l.stopWhere(site), What: road + " in " + where + ", whose operand may be a range type's value held as a union, where a NodeArray may stand for it"}
					return false
				}
			}
			return true
		}
	}
	for index, statement := range l.result.Main {
		walk(statement, inspect("the program's body", stopSite{function: -1, statement: index}))
	}
	for index, function := range l.result.Functions {
		walk(function.Body, inspect("function "+function.Name, stopSite{function: index, statement: -1}))
	}
	return refused
}

// rangeUnionRoad names a stopping expression for its NotYet.
func rangeUnionRoad(expression ir.Expression) string {
	switch value := expression.(type) {
	case ir.Property:
		return "a read of field " + value.Name
	case ir.Narrow:
		return "a narrow of a union taken unasked"
	case ir.ObjectLiteral, ir.RecordLiteral:
		return "a spread"
	case ir.ObjectCall:
		return "Object." + value.Method
	case ir.ObjectKeys:
		return "Object.keys or a for-in"
	case ir.HasProperty:
		return "an in test"
	case ir.HasOwn:
		return "an own-field test"
	case ir.HasAccessor:
		return "an accessor test"
	case ir.InstanceOf:
		return "an instanceof test"
	case ir.JSONEncode:
		return "encodeJson"
	case ir.BoundMethod:
		return "a bound method"
	case ir.RecordStorage:
		return "a record storage test"
	}
	return "an expression"
}

// rangeOperands answers whether a Union operand may be a range type's value: by the checker's type where the
// IR keeps one, a local's (localTypes), and otherwise by where the value came from, through the shared flow
// graph's sources (a parameter's arguments, a local's writes, a helper's returns), failing closed on any it
// can't place. A call of a function lowered from a declaration is judged by the return type it declares, and a
// field read by its name (rangeFieldNamed). A box and a literal made in place are their values: a slot's range
// type is answered where the slot is read, never where a value enters it. The graph is built at the first
// question that needs it.
type rangeOperands struct {
	l     *lowering
	graph *allocationFlowGraph
}

func (r *rangeOperands) flow() *allocationFlowGraph {
	if r.graph == nil {
		r.graph = r.l.flowGraph(r.l.result)
	}
	return r.graph
}

// held walks back from operand to what made it, and says whether any of it may be a range type's value.
func (r *rangeOperands) held(operand ir.Expression) bool {
	queue := []ir.Expression{operand}
	visited := map[int]bool{}
	// sources queues a graph node's sources, and says true where the node may hold a value the graph can't
	// place, or has no source at all.
	sources := func(node int) bool {
		if visited[node] {
			return false
		}
		visited[node] = true
		graph := r.flow()
		if _, unknown := graph.unknownAt(node); unknown || len(graph.sources[node]) == 0 {
			return true
		}
		queue = append(queue, graph.sources[node]...)
		return false
	}
	for len(queue) != 0 {
		value := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		switch value := value.(type) {
		case nil:
			return true
		case ir.Read:
			if proven, typed := r.l.localTypes[value.Local]; typed {
				if r.l.holdsRangeUnion(proven) {
					return true
				}
				for _, also := range r.l.localAlso[value.Local] {
					if r.l.holdsRangeUnion(also) {
						return true
					}
				}
				continue
			}
			if sources(value.Local + 1) {
				return true
			}
		case ir.Call:
			targets := r.l.result.CallTargets(value)
			if len(targets) == 0 {
				return true
			}
			for _, target := range targets {
				// A function lowered from a declaration gives back a value of the type it declares, whatever it
				// returns was held as before (a Box makes no type of its own), so that type answers; a helper
				// lowering made is followed to what it returns.
				if declaration := r.l.functionNodes[target]; declaration != nil {
					if r.l.holdsRangeUnion(r.l.checker.GetReturnTypeOfSignature(r.l.checker.GetSignatureFromDeclaration(declaration))) {
						return true
					}
					continue
				}
				if sources(r.flow().resultNode(target)) {
					return true
				}
			}
		case ir.CallClosure:
			targets := r.l.result.ClosureTargets(value)
			if targets.Unknown || len(targets.Functions) == 0 {
				return true
			}
			for _, target := range targets.Functions {
				if sources(r.flow().resultNode(target)) {
					return true
				}
			}
		case ir.Box:
			// A box of an array may be the NodeArray a range type takes, so it counts, failing closed: a reader
			// that reads a layout raw and is handed a boxed array stops. Lower's condition on #hae423v: following
			// the box to its value would let a NodeArray-typed local's box past every such reader with no
			// witness. A reader that asks the kind (an in test, a dynamic read) never asks this. Any other box
			// holds its value as it was before a union slot took it, so it's followed.
			if value.Value == nil || value.Value.Type() == ir.Array {
				return true
			}
			queue = append(queue, value.Value)
		case ir.Conditional:
			queue = append(queue, value.WhenTrue, value.WhenNot)
		case ir.Coalesce:
			queue = append(queue, value.Value)
			if value.Panic == nil {
				queue = append(queue, value.Fallback)
			}
		case ir.Logical:
			queue = append(queue, value.Left, value.Right)
		case ir.Comma:
			queue = append(queue, value.Right)
		case ir.Effects:
			queue = append(queue, value.Result)
		case ir.Narrow:
			queue = append(queue, value.Value)
		case ir.Unwrap:
			queue = append(queue, value.Value)
		case ir.Defined:
			queue = append(queue, value.Value)
		case ir.CheckedCast:
			queue = append(queue, value.Value)
		case ir.Property:
			if r.l.rangeFieldNamed(value.Name) {
				return true
			}
		case ir.DynamicProperty:
			if r.l.rangeFieldNamed(value.Name) {
				return true
			}
		case ir.Undefined, ir.Null, ir.NumberConstant, ir.StringConstant, ir.BooleanConstant:
		case ir.ObjectLiteral, ir.ArrayLiteral:
			// A value made right here, an object or an array of its own type; a slot that takes it as a range
			// type is answered where the slot is read, as a box's is.
		default:
			// A value read out of storage the IR keeps no type for: an element, a map's or a record's value, a
			// dynamic key's read, a settled promise.
			return true
		}
	}
	return false
}

// noteRangeField adds a field name a range type's value is read from.
func (l *lowering) noteRangeField(name string) {
	if l.rangeFields == nil {
		l.rangeFields = map[string]bool{}
	}
	l.rangeFields[name] = true
}

// rangeFieldNamed is whether a field of this name may hold a range type's value: one some declaration in the
// program's files gives a type holding a range type, a type parameter's included (holdsRangeUnion), or one
// property read as one (noteRangeField). Every file the program holds is walked once, never the bundled
// library or the prelude (load.Program.SourceFiles): a property declared, an object literal's property, a
// constructor's parameter property, a getter, and each property a concrete mapped type gives. A computed
// name holding one may be any name, and so may every name in a lowering with no program to walk.
func (l *lowering) rangeFieldNamed(name string) bool {
	if !l.rangeFieldsKnown && l.program == nil {
		l.rangeFieldsKnown, l.rangeFieldsAny = true, true
	}
	if !l.rangeFieldsKnown {
		l.rangeFieldsKnown = true
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindPropertySignature, ast.KindPropertyDeclaration, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindGetAccessor:
				l.noteRangeFieldDeclaration(node)
			case ast.KindParameter:
				if parameterProperty(node) {
					l.noteRangeFieldDeclaration(node)
				}
			case ast.KindMappedType:
				// Its names come from its keys, which no declaration spells. A generic one gives none until
				// it's instantiated, and a read of an instance's field is noted where property lowers it.
				for _, property := range l.checker.GetPropertiesOfType(l.checker.GetTypeFromTypeNode(node)) {
					if l.holdsRangeUnion(l.checker.GetTypeOfSymbol(property)) {
						l.noteRangeField(property.Name)
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		for _, file := range l.program.SourceFiles() {
			file.AsNode().ForEachChild(visit)
		}
	}
	return l.rangeFieldsAny || l.rangeFields[name]
}

// noteRangeFieldDeclaration notes a property declaration's name when its type holds a range type.
func (l *lowering) noteRangeFieldDeclaration(node *ast.Node) {
	symbol := l.checker.GetSymbolOfDeclaration(node)
	if symbol == nil || !l.holdsRangeUnion(l.checker.GetTypeOfSymbol(symbol)) {
		return
	}
	name := node.Name()
	switch {
	case name == nil || name.Kind == ast.KindComputedPropertyName:
		l.rangeFieldsAny = true
	case name.Kind == ast.KindIdentifier || name.Kind == ast.KindPrivateIdentifier:
		l.noteRangeField(l.fieldName(name))
	default:
		l.noteRangeField(name.Text())
	}
}
