package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"strings"
)

// Mutability belongs to the array type, independently of whether its containing
// property is readonly. Optional and union fields retain every array arm's rule.
func mutableViewArrayContract(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) || seen[id] {
		return false
	}
	seen[id] = true
	contract := program.ViewContracts[id-1]
	if contract.Kind == ir.ViewArray {
		return !contract.ArrayReadonly
	}
	for _, member := range contract.Members {
		if mutableViewArrayContract(program, member, seen) {
			return true
		}
	}
	return false
}

// One inventory of Array* nodes. Only named handouts and operations reading no
// element are admitted. New nodes and library members fail closed.
type viewArrayRoad struct {
	member             string
	handout, noElement bool
}

var viewArrayRoads = map[reflect.Type]viewArrayRoad{
	reflect.TypeFor[ir.Length]():         {member: "length", noElement: true},
	reflect.TypeFor[ir.ArrayIndex]():     {member: "index", handout: true},
	reflect.TypeFor[ir.ViewElement]():    {member: "index", handout: true},
	reflect.TypeFor[ir.ArrayMap]():       {member: "map", handout: true},
	reflect.TypeFor[ir.ArrayVisit]():     {member: "visit", handout: true},
	reflect.TypeFor[ir.ArrayReduce]():    {member: "reduce", handout: true},
	reflect.TypeFor[ir.ForOf]():          {member: "for-of", handout: true},
	reflect.TypeFor[ir.ArrayIsArray]():   {member: "Array.isArray", noElement: true},
	reflect.TypeFor[ir.ArrayLiteral]():   {member: "spread", noElement: true},
	reflect.TypeFor[ir.ParallelMap]():    {member: "parallelMap"},
	reflect.TypeFor[ir.ArrayConcat]():    {member: "concat"},
	reflect.TypeFor[ir.ArrayFill]():      {member: "fill"},
	reflect.TypeFor[ir.ArrayFrom]():      {member: "Array.from"},
	reflect.TypeFor[ir.ArrayJoin]():      {member: "join"},
	reflect.TypeFor[ir.ArrayOfScalars](): {member: "whole-array scalar test"},
	reflect.TypeFor[ir.ArrayPop]():       {member: "pop"},
	reflect.TypeFor[ir.ArrayPush]():      {member: "push"},
	reflect.TypeFor[ir.ArrayReverse]():   {member: "reverse"},
	reflect.TypeFor[ir.ArraySearch]():    {member: "indexOf/includes"},
	reflect.TypeFor[ir.ArraySlice]():     {member: "slice"},
	reflect.TypeFor[ir.ArraySort]():      {member: "sort"},
	reflect.TypeFor[ir.ArraySplice]():    {member: "splice"},
}

func viewArrayRoadOf(node any) viewArrayRoad {
	road, known := viewArrayRoads[reflect.TypeOf(node)]
	if !known {
		return viewArrayRoad{member: reflect.TypeOf(node).Name()}
	}
	if visit, ok := node.(ir.ArrayVisit); ok {
		road.member = visit.Method
		switch visit.Method {
		case "forEach", "some", "every", "filter", "find", "findIndex", "findLast", "findLastIndex":
		default:
			road.handout = false
		}
	}
	return road
}

// viewElementKey is which element reads get a checked view's element check (Lower's ruling on
// #6wfvfhm): a read of E is checked when E is comparable to the element type of an array some view
// in the program holds, or, where any view holds an array, when E is or holds a type parameter, any
// or unknown. A program whose views hold no array checks no read. Viewed is every such element
// type, collected before any view is made or any read lowers (collectViewedElements). Every checks
// each read, where the pre-pass can't decide. Misses are the array contracts lowering made whose
// element viewed lacks: any one means a read may have gone unchecked, and the program is lowered
// again with every set (lowerCheckedWithCycleTypes).
type viewElementKey struct {
	every   bool
	viewed  map[*checker.Type]bool
	order   []*checker.Type
	related map[*checker.Type]bool
	misses  []string
}

// collectViewedElements is the key's pre-pass. A view is made only at a cast or a type predicate
// (a speculation's result is a cast too), so it reads every one in every module the program loads,
// .a and .ts alike, and collects the element type of every array its target can hold, through
// the edges viewContract builds contracts along. A target that names a type parameter is
// instantiated where it lowers, past what the pre-pass can see, so it sets every.
func (l *lowering) collectViewedElements(modules []*ast.SourceFile) {
	key := l.elementKey
	if key == nil || key.every {
		return
	}
	key.viewed = map[*checker.Type]bool{}
	key.related = map[*checker.Type]bool{}
	seen := map[*checker.Type]bool{}
	var hold func(target *checker.Type, arm bool)
	hold = func(target *checker.Type, arm bool) {
		if target == nil || seen[target] {
			return
		}
		seen[target] = true
		switch {
		case target.Flags()&checker.TypeFlagsUnionOrIntersection != 0:
			for _, member := range target.Types() {
				hold(member, true)
			}
			for _, property := range l.checker.GetPropertiesOfType(target) {
				hold(l.checker.GetTypeOfSymbol(property), false)
			}
		case l.isArrayType(target):
			element := l.elementTypeOfArray(target)
			if element == nil {
				key.every = true
				return
			}
			if !key.viewed[element] {
				key.viewed[element] = true
				key.order = append(key.order, element)
			}
			if element.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || l.mentionsTypeParameter(element, map[*checker.Type]bool{}) {
				key.every = true
			}
			hold(element, false)
		case checker.IsTupleType(target):
			for _, element := range l.checker.GetTypeArguments(target) {
				hold(element, false)
			}
		case target.Flags()&checker.TypeFlagsObject != 0:
			signatures := append(append([]*checker.Signature{}, l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)...), l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)...)
			if len(signatures) != 0 {
				// A callable's contract is a leaf, but a callable union arm's holds its parameters
				// and result (view_unions_callable_members.go).
				if arm {
					for _, signature := range signatures {
						for _, parameter := range signature.Parameters() {
							hold(l.checker.GetTypeOfSymbol(parameter), false)
						}
						hold(l.checker.GetReturnTypeOfSignature(signature), false)
					}
				}
				return
			}
			for _, property := range l.checker.GetPropertiesOfType(target) {
				hold(l.checker.GetTypeOfSymbol(property), false)
			}
		}
	}
	view := func(target *checker.Type) {
		if target == nil {
			return
		}
		if l.mentionsTypeParameter(target, map[*checker.Type]bool{}) {
			key.every = true
			return
		}
		hold(target, false)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			view(l.checker.GetTypeAtLocation(node))
		case ast.KindTypePredicate:
			if written := node.AsTypePredicateNode().Type; written != nil {
				view(l.checker.GetTypeAtLocation(written))
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
}

// elementKeyed is whether a read of element is checked. Comparable is assignable either way: a
// wider read (object, a Node supertype) takes a viewed Node, and so does a narrower one, since
// filter with a type predicate, or a narrowing, hands a viewed Node out as an Identifier, and Node
// isn't assignable to Identifier.
func (l *lowering) elementKeyed(element *checker.Type) bool {
	key := l.elementKey
	if key == nil || key.every || element.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || l.mentionsTypeParameter(element, map[*checker.Type]bool{}) {
		return true
	}
	if related, known := key.related[element]; known {
		return related
	}
	related := false
	for _, viewed := range key.order {
		if l.checker.IsTypeAssignableTo(viewed, element) || l.checker.IsTypeAssignableTo(element, viewed) {
			related = true
			break
		}
	}
	key.related[element] = related
	return related
}

func (l *lowering) viewArrayElementCheck(node, source *ast.Node) (*ir.ViewElementCheck, error) {
	target := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(source)))
	element := l.viewArrayElementType(target)
	if element == nil {
		return nil, nil
	}
	// The key is the element type, mutable array or readonly: filter, map and slice hand a viewed readonly
	// array's own elements to a new mutable array, so a mutable array can hold viewed elements too (Lower's seam 1
	// on #6wfvfhm). It's scoped by what the program views (viewElementKey), so a program that views no array has
	// no check at all. The check reads its array lent, as ArrayIndex does, so it adds no ownership calls.
	if key := l.elementKey; key != nil && !key.every && len(key.order) == 0 {
		// No view holds an array, so nothing can demand a check, whatever the element type.
		return nil, nil
	}
	element = l.concrete(element)
	check := &ir.ViewElementCheck{Where: l.program.Where(node), View: elementReadView(node, source), ViewType: l.checker.TypeToString(element)}
	if !l.elementKeyed(element) {
		// An unrelated read is marked, so checkViewArrays reads it raw rather than stopping it as unchecked.
		check.Unrelated = true
		return check, nil
	}
	if element.Flags()&checker.TypeFlagsUnknown != 0 {
		return check, nil
	}
	start, getters := len(l.result.ViewContracts), len(l.contractGetters)
	id, err := l.viewContract(node, element)
	// Recursive builders can discard an unavailable union arm. Do not retain
	// producer targets whose reserved rows were discarded or reused by that build.
	for candidate, target := range l.untaggedCallableTargets {
		if int(candidate) > start && (int(candidate) > len(l.result.ViewContracts) || l.result.ViewContractTypes[int(target.Id())] != candidate || l.result.ViewContracts[candidate-1].Kind != ir.ViewCallable) {
			delete(l.untaggedCallableTargets, candidate)
		}
	}
	if err != nil {
		// A tentative annotation must not stop an ordinary, unviewed array read.
		// Missing annotations fail closed later if the operand reaches a view.
		l.result.ViewContracts = l.result.ViewContracts[:start]
		l.contractGetters = l.contractGetters[:getters]
		for key, value := range l.result.ViewContractTypes {
			if int(value) > start {
				delete(l.result.ViewContractTypes, key)
			}
		}
		for id := range l.untaggedCallableTargets {
			if int(id) > start {
				delete(l.untaggedCallableTargets, id)
			}
		}
		switch err.(type) {
		case *NotYet, *Refused:
			return nil, nil
		}
		return nil, err
	}
	check.Contract = id
	return check, nil
}

func (l *lowering) viewArrayIndex(node, source *ast.Node, read ir.ArrayIndex) (ir.Expression, error) {
	check, err := l.viewArrayElementCheck(node, source)
	if err == nil && check == nil {
		l.recordViewArrayUse(node, read.Array, "index")
	}
	if err != nil || check == nil {
		return read, err
	}
	return ir.ViewElement{Array: read.Array, Index: read.Index, Relative: read.Relative, Contract: check.Contract, Where: check.Where, View: check.View, ViewType: check.ViewType, Of: read.Type(), Element: read.Element, Unrelated: check.Unrelated, Optional: read.Optional}, nil
}

func elementContractCheckable(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	if seen[id] {
		return true
	}
	seen[id] = true
	contract := program.ViewContracts[id-1]
	if contract.Kind == ir.ViewUnknown && contract.Unsupported == "unknown" {
		return true
	}
	if contract.Unsupported != "" && !(contract.Kind == ir.ViewArray && contract.ArrayReadonly) && contract.Unsupported != "untagged object union" {
		return false
	}
	switch contract.Kind {
	case ir.ViewNull:
		return contract.Of == ir.Union
	case ir.ViewScalar, ir.ViewUndefined:
		return true
	case ir.ViewCallable:
		return contract.ProducerCertified
	case ir.ViewArray:
		return contract.ArrayReadonly && elementContractCheckable(program, contract.Element, seen)
	case ir.ViewUnion:
		// Legacy nullable pointers do not store a separate null tag. The array's
		// reference-storage byte cannot prove which nullish value NULL meant.
		if contract.Of.IsReference() && contract.Of != ir.Union {
			for _, member := range contract.Members {
				if program.ViewContracts[member-1].Kind == ir.ViewNull {
					return false
				}
			}
		}
		if len(contract.Members) == 0 {
			return false
		}
		for _, member := range contract.Members {
			if !elementContractCheckable(program, member, seen) {
				return false
			}
		}
		return true
	case ir.ViewObject:
		for _, field := range contract.Fields {
			child := program.ViewContracts[field.Contract-1]
			if !field.Optional && child.Kind == ir.ViewScalar && (field.Name == "kind" || len(child.Allowed) > 0) {
				return true
			}
		}
		for _, field := range contract.Fields {
			child := program.ViewContracts[field.Contract-1]
			// Unknown is raw only as the element itself. This matcher cannot
			// yet certify a plain object by an untagged unknown payload.
			if child.Kind == ir.ViewUnknown {
				return false
			}
			if !elementContractCheckable(program, field.Contract, seen) {
				return false
			}
		}
		return true
	}
	return false
}

func (l *lowering) checkViewArrays() error {
	program := l.result
	// Callable producers become known after descriptor construction. Retire an
	// array's provisional stop once its element row is ready for checked loads.
	for i, contract := range program.ViewContracts {
		if contract.Kind == ir.ViewArray && contract.ArrayReadonly && contract.Unsupported == "array element contract" && elementContractCheckable(program, contract.Element, map[ir.ViewContractID]bool{}) {
			program.ViewContracts[i].Unsupported = ""
		}
	}
	graph, viewed, unknown := viewedAllocations(program, true)
	hasArrayView := false
	for _, contract := range program.ViewContracts {
		hasArrayView = hasArrayView || contract.Kind == ir.ViewArray && contract.ArrayReadonly
	}

	demanded := func(array ir.Expression) bool {
		if len(program.ViewOrigins) == 0 || !hasArrayView {
			return false
		}
		reaches := graph.ReachingAllocations(array)
		if unknown || reaches.Unknown {
			return true
		}
		for _, site := range reaches.Sites {
			if viewed[site] {
				return true
			}
		}
		return false
	}
	// A fresh container inherits viewed reach when it stores a viewed value.
	// This includes stores after construction and collection-to-array handoffs.
	for changed := true; changed; {
		changed = false
		mark := func(source, destination ir.Expression) {
			if !viewArrayAggregate(source) || !demanded(source) || destination == nil {
				return
			}
			reaches := graph.ReachingAllocations(destination)
			if reaches.Unknown && !unknown {
				unknown = true
				changed = true
			}
			for _, site := range reaches.Sites {
				if !viewed[site] {
					viewed[site] = true
					changed = true
				}
			}
		}
		visit := func(node any) bool {
			switch value := node.(type) {
			case ir.ArrayLiteral:
				for _, item := range value.Elements {
					mark(item, value)
				}
			case ir.ObjectLiteral:
				for _, field := range value.Fields {
					mark(field.Value, value)
				}
				mark(value.Spread, value)
			case ir.ArrayPush:
				mark(value.Value, value.Array)
			case ir.SetIndex:
				mark(value.Value, value.Array)
			case ir.SetProperty:
				mark(value.Value, value.Object)
			case ir.SetAdd:
				mark(value.Value, value.Set)
			case ir.MapSet:
				mark(value.Key, value.Map)
				mark(value.Value, value.Map)
			case ir.ArrayMap:
				mark(value.Array, value)
			case ir.ArrayVisit:
				if value.Method == "filter" {
					mark(value.Array, value)
				}
			case ir.SetValues:
				mark(value.Set, value)
			case ir.MapKeys:
				mark(value.Map, value)
			case ir.MapValues:
				mark(value.Map, value)
			case ir.MapEntries:
				mark(value.Map, value)
			case ir.ObjectCall:
				if value.Method == "values" || value.Method == "entries" {
					for _, argument := range value.Arguments {
						mark(argument, value)
					}
				}
			}
			return true
		}
		walk(program.Main, visit)
		for _, fn := range program.Functions {
			walk(fn.Body, visit)
		}
	}

	var refused error
	stop := func(where, member string) {
		if refused == nil {
			refused = &NotYet{Where: where, What: "a checked view's array read through " + member + " (not checked yet)"}
		}
	}
	for _, cast := range program.ViewArrayCasts {
		origins := append([]ir.Expression{}, program.ViewOrigins[:cast.OriginStart]...)
		origins = append(origins, program.ViewOrigins[cast.OriginEnd:]...)
		if len(origins) == 0 {
			continue
		}
		writable, opaque := graph.viewedFrom(origins, true)
		reaches := graph.ReachingAllocations(cast.Value)
		unsafe := opaque || reaches.Unknown
		for _, site := range reaches.Sites {
			unsafe = unsafe || writable[site]
		}
		if unsafe {
			return &NotYet{Where: cast.Where, What: "a cast that makes a viewed readonly array writable"}
		}
	}
	wholeDemanded := func(array ir.Expression) bool {
		return viewArrayAggregate(array) && demanded(array)
	}
	for _, use := range program.ViewArrayUses {
		if (!use.Whole && demanded(use.Array)) || (use.Whole && wholeDemanded(use.Array)) {
			stop(use.Where, use.Member)
		}
	}
	rewrite := func(node any) any {
		var array ir.Expression
		var check *ir.ViewElementCheck
		var element ir.Expression
		whole := false
		road := viewArrayRoadOf(node)
		switch value := node.(type) {
		case ir.ViewElement:
			element = value
			array = value.Array
			check = &ir.ViewElementCheck{Contract: value.Contract, Where: value.Where, View: value.View, ViewType: value.ViewType, Unrelated: value.Unrelated}
		case ir.ArrayPush:
			array = value.Array
		case ir.ArrayIndex:
			element = value
			array = value.Array
		case ir.ArrayVisit:
			element = ir.ArrayIndex{Array: value.Array, Index: ir.NumberConstant{Value: -1}, Element: value.Element}
			array = value.Array
			check = value.ElementCheck
		case ir.ArrayMap:
			element = ir.ArrayIndex{Array: value.Array, Index: ir.NumberConstant{Value: -1}, Element: value.Element}
			array = value.Array
			check = value.ElementCheck
		case ir.ArrayReduce:
			element = ir.ArrayIndex{Array: value.Array, Index: ir.NumberConstant{Value: -1}, Element: value.Element}
			array = value.Array
			check = value.ElementCheck
		case ir.ForOf:
			if value.Iterable.Type() == ir.Array {
				array = value.Iterable
				element = ir.ArrayIndex{Array: value.Iterable, Index: ir.NumberConstant{Value: -1}, Element: value.Element}
				check = value.ElementCheck
				if value.Pattern != nil {
					road.handout = false
					road.member = "destructuring"
				}
			}
		case ir.ParallelMap:
			array = value.Items
		case ir.ArrayFrom:
			array = value
			element = ir.ArrayIndex{Array: value, Index: ir.NumberConstant{Value: -1}, Element: value.Element}
		case ir.ArrayConcat:
			array = value.Array
			for _, other := range value.Others {
				if demanded(other) {
					stop(program.Source, "concat")
				}
			}
		case ir.JSONEncode:
			whole = true
			array = value.Value
			road.member = "JSON.stringify"
		case ir.JSONStringify:
			whole = true
			array = value.Value
			road.member = "JSON.stringify"
		case ir.CollectionIterator:
			array = value.Collection
			road.member = value.Part
		case ir.SetNew:
			array = value.Values
			road.member = "Set constructor"
		case ir.MapNew:
			array = value.Pairs
			road.member = "Map constructor"
		case ir.ObjectCall:
			if value.Method == "values" || value.Method == "entries" {
				for _, arg := range value.Arguments {
					if arg.Type() == ir.Array && demanded(arg) {
						// Only an array let through the earlier Object.values/entries refusal reaches this stop; today none is.
						stop(program.Source, "Object."+value.Method)
					}
				}
			}
		case ir.ArrayLiteral:
			for i, item := range value.Elements {
				if i < len(value.Spread) && value.Spread[i] && demanded(item) {
					stop(program.Source, "spread")
				}
			}
			return node
		default:
			reflected := reflect.ValueOf(node)
			if reflected.Kind() == reflect.Struct {
				field := reflected.FieldByName("Array")
				if field.IsValid() && field.CanInterface() {
					array, _ = field.Interface().(ir.Expression)
				}
			}
		}
		if whole && wholeDemanded(array) {
			stop(program.Source, road.member)
			return node
		}
		if array == nil || array.Type() != ir.Array {
			return node
		}
		needed := demanded(array)
		if viewArrayAggregate(element) {
			needed = needed || demanded(element)
		}
		// A read the key relates to no viewed element type has no check of its own (viewElementKey), but
		// unrelated isn't unreachable: an any or unknown hop can hand a viewed allocation to an Other[]
		// that relates to Node neither way (Lower's must on #6wfvfhm). So it reads raw only where its
		// array can't reach a viewed allocation, as every read here does, and stops where it can.
		if check != nil && check.Unrelated {
			if needed {
				stop(check.Where, road.member)
				return node
			}
			switch value := node.(type) {
			case ir.ViewElement:
				return ir.ArrayIndex{Array: value.Array, Index: value.Index, Element: value.Element, Relative: value.Relative, Optional: value.Optional}
			case ir.ArrayVisit:
				value.ElementCheck = nil
				return value
			case ir.ArrayMap:
				value.ElementCheck = nil
				return value
			case ir.ArrayReduce:
				value.ElementCheck = nil
				return value
			case ir.ForOf:
				value.ElementCheck = nil
				return value
			}
		}
		if check != nil && program.ViewLegacyNullArrays && (needed || check.Contract == 0) {
			if refused == nil {
				refused = &NotYet{Where: check.Where, What: "an array element read in a program with legacy nullable-pointer storage"}
			}
		}
		if needed && !road.handout && !road.noElement {
			stop(program.Source, road.member)
		}
		if needed && road.handout && check == nil {
			stop(program.Source, road.member)
		}
		if check != nil && needed && check.Contract != 0 && !elementContractCheckable(program, check.Contract, map[ir.ViewContractID]bool{}) {
			stop(check.Where, road.member)
		}
		// Unknown elements admit anything, but still need a storage-aware load
		// when a scalar array has been widened to readonly unknown[].
		if !needed && (check == nil || check.Contract != 0) {
			switch value := node.(type) {
			case ir.ViewElement:
				// The read put back as it was lowered, with the array's own element representation.
				return ir.ArrayIndex{Array: value.Array, Index: value.Index, Element: value.Element, Relative: value.Relative, Optional: value.Optional}
			case ir.ArrayVisit:
				value.ElementCheck = nil
				return value
			case ir.ArrayMap:
				value.ElementCheck = nil
				return value
			case ir.ArrayReduce:
				value.ElementCheck = nil
				return value
			case ir.ForOf:
				value.ElementCheck = nil
				return value
			}
		}
		return node
	}
	program.Main = rewriteShapeStatements(program.Main, rewrite)
	for i := range program.Functions {
		program.Functions[i].Body = rewriteShapeStatements(program.Functions[i].Body, rewrite)
	}
	// Pure annotations must not leave checker-private symbol names in programs
	// that have no view reads. They are dead after the ordinary reads are restored.
	if len(program.ViewOrigins) == 0 {
		live := false
		inspect := func(node any) bool {
			switch value := node.(type) {
			case ir.Property:
				live = live || program.CheckedRead(value)
			case ir.ViewElement:
				live = live || value.Contract > 0
			case ir.ObjectCall:
				live = live || value.Checked
			}
			return !live
		}
		walk(program.Main, inspect)
		for _, fn := range program.Functions {
			walk(fn.Body, inspect)
		}
		if !live {
			program.ViewContracts = nil
			program.ViewContractTypes = nil
			l.untaggedCallableTargets = nil
		}
	}
	return refused
}

func (l *lowering) recordViewArrayUse(node *ast.Node, array ir.Expression, member string) {
	l.result.ViewArrayUses = append(l.result.ViewArrayUses, ir.ViewArrayUse{Array: array, Member: member, Where: l.program.Where(node)})
}

func (l *lowering) recordWholeViewArrayUse(node *ast.Node, array ir.Expression) {
	l.result.ViewArrayUses = append(l.result.ViewArrayUses, ir.ViewArrayUse{Array: array, Member: "JSON.stringify", Where: l.program.Where(node), Whole: true})
}

// The static checker type is instantiated before this boundary, never keyed by
// an interface's name or by an array allocation's first contents.

func (l *lowering) cast(node *ast.Node) (ir.Expression, error) {
	start := len(l.result.ViewOrigins)
	value, err := l.castValue(node)
	if err != nil {
		return nil, err
	}
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	if l.viewMakesArrayWritable(target, map[*checker.Type]bool{}) {
		l.result.ViewArrayCasts = append(l.result.ViewArrayCasts, ir.ViewArrayCast{Value: value, Where: l.program.Where(node), OriginStart: start, OriginEnd: len(l.result.ViewOrigins)})
	}
	return value, nil
}

func (l *lowering) viewMakesArrayWritable(target *checker.Type, seen map[*checker.Type]bool) bool {
	if target == nil || seen[target] {
		return false
	}
	seen[target] = true
	if l.isArrayType(target) {
		if !l.isLibraryType(target, "ReadonlyArray") {
			return true
		}
		return l.viewMakesArrayWritable(l.viewArrayElementType(target), seen)
	}
	if target.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range target.Types() {
			if l.viewMakesArrayWritable(member, seen) {
				return true
			}
		}
		return false
	}
	if checker.IsTupleType(target) {
		for _, element := range l.checker.GetTypeArguments(target) {
			if l.viewMakesArrayWritable(element, seen) {
				return true
			}
		}
		return false
	}
	if target.Flags()&checker.TypeFlagsObject != 0 {
		for _, field := range l.checker.GetPropertiesOfType(target) {
			if l.viewMakesArrayWritable(l.checker.GetTypeOfSymbol(field), seen) {
				return true
			}
		}
	}
	return false
}

// An erased operand is Unknown reach, including hiding behind object. Diagnose
// the writable target before the generic cast proof can give a less precise stop.
// Adamic files retain rule 1950's existing refusal for unproven assertions.
func (l *lowering) erasedWritableViewCast(node *ast.Node) error {
	if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.viewMakesArrayWritable(target, map[*checker.Type]bool{}) {
		return nil
	}
	// A cast to a phantom brand (tsc's SortedArray<T>, #gb0fqaa) is castProof's to judge: a mint, whose
	// operand under the unknown hop must relate to the brand's mutable base, so no readonly array is ever
	// minted writable, or else a refusal naming the fix. Seen as its base, SortedArray<T> is a writable
	// array, and this stop, which only words the cast earlier, would replace both. checkViewArrays still
	// judges every recorded cast's reach.
	if base, _ := l.phantomExtension(target); base != nil {
		return nil
	}
	source := l.concrete(l.checker.GetTypeAtLocation(node.AsAsExpression().Expression))
	if source.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0 || l.isLibraryType(source, "ReadonlyArray") {
		return l.notYet(node, "a cast that makes a viewed readonly array writable")
	}
	return nil
}

// Reference arrays currently share NULL for legacy null and undefined slots.
// Until storage preserves a distinct tag, an erased or checked handout cannot
// reconstruct that value from its heap header. Never guess from the static view.
func (l *lowering) noteViewArrayStorage(node *ast.Node, value ir.Expression) error {
	if value.Type() != ir.Array {
		return nil
	}
	target := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(node)))
	if !l.isArrayType(target) {
		return nil
	}
	element := l.concrete(l.elementTypeOfArray(target))
	if element.Flags()&checker.TypeFlagsUnknown != 0 {
		if err := l.unknownArrayMutation(node); err != nil {
			return err
		}
	}
	if literal, ok := value.(ir.ArrayLiteral); ok && literal.Element == ir.Union {
		return nil
	}
	if l.includesNull(element) {
		if of, known := l.representation(element); known && of != ir.Union {
			l.result.ViewLegacyNullArrays = true
		}
	}
	return nil
}

// Unknown slots have no ownership schema. The compiler pin writes a scalar
// through an original unknown[] owner. That does not license reference cycles
// or ownership of a scalar-storage array widened to a mutable unknown[].
func (l *lowering) unknownArrayMutation(node *ast.Node) error {
	source := node
	parent := node.Parent
	for parent != nil && (parent.Kind == ast.KindParenthesizedExpression || parent.Kind == ast.KindNonNullExpression || parent.Kind == ast.KindSatisfiesExpression) {
		source, parent = parent, parent.Parent
	}
	if parent == nil {
		return nil
	}
	if parent.Kind == ast.KindSpreadElement {
		return l.notYet(parent, "a checked view's array read through spread (not checked yet)")
	}
	if parent.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") && callee.Name().Text() != "isArray" {
			return l.notYet(parent, "a checked view's array read through Array."+callee.Name().Text()+" (not checked yet)")
		}
	}
	if parent.Kind == ast.KindElementAccessExpression {
		access := parent
		outer := access.Parent
		for outer != nil && outer.Kind == ast.KindParenthesizedExpression {
			access, outer = outer, outer.Parent
		}
		if outer != nil && outer.Kind == ast.KindBinaryExpression {
			binary := outer.AsBinaryExpression()
			if binary.Left == access && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return l.notYet(outer, "an index write into an unknown array whose ownership is not proven")
			}
		}
	}
	if parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != source {
		return nil
	}
	member := parent.Name().Text()
	switch member {
	case "push":
		call := parent.Parent
		if call == nil || call.Kind != ast.KindCallExpression {
			return nil
		}
		for _, argument := range call.AsCallExpression().Arguments.Nodes {
			if !l.unknownArrayScalar(l.concrete(l.checker.GetTypeAtLocation(argument))) {
				return l.notYet(call, "a reference write into an unknown array whose ownership is not proven")
			}
		}
	case "pop", "shift", "unshift", "splice", "fill", "copyWithin", "reverse", "sort":
		return l.notYet(parent, "a mutation of an unknown array through "+member+" (not checked yet)")
	default:
		admitted := false
		if member == "at" {
			admitted = viewArrayRoads[reflect.TypeFor[ir.ArrayIndex]()].handout
		}
		if member == "values" {
			call := parent.Parent
			admitted = call != nil && call.Kind == ast.KindCallExpression && call.Parent != nil && call.Parent.Kind == ast.KindForOfStatement && viewArrayRoads[reflect.TypeFor[ir.ForOf]()].handout
		}
		if viewArrayRoadOf(ir.ArrayVisit{Method: member}).handout {
			admitted = true
		}
		for _, road := range viewArrayRoads {
			if road.member == member && (road.handout || road.noElement) {
				admitted = true
			}
		}
		if !admitted {
			return l.notYet(parent, "a checked view's array read through "+member+" (not checked yet)")
		}
	}
	return nil
}

func (l *lowering) unknownArrayScalar(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !l.unknownArrayScalar(member) {
				return false
			}
		}
		return true
	}
	return target.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0
}

// A known boxed primitive cannot carry a viewed allocation. Unknown references
// and mutable aggregates retain the existing fail-closed reach rule.
func viewArrayAggregate(value ir.Expression) bool {
	if boxed, ok := value.(ir.Box); ok {
		return viewArrayAggregate(boxed.Value)
	}
	return viewAggregate(value)
}

// elementReadView is an element read in the program's own words, for its stop, as a field read's is
// (sourceExpression): an index or at read is its own source text, view._children[1]; a handout, which
// has no expression of its own, is its receiver and its road, view._children's element, in map.
func elementReadView(node, source *ast.Node) string {
	switch node.Kind {
	case ast.KindElementAccessExpression:
		return sourceExpression(node)
	case ast.KindForOfStatement:
		return sourceExpression(source) + "'s element, in for...of"
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			name := callee.AsPropertyAccessExpression().Name().Text()
			if name == "at" || name == "values" {
				return sourceExpression(node)
			}
			return sourceExpression(source) + "'s element, in " + name
		}
	}
	return sourceExpression(node)
}
