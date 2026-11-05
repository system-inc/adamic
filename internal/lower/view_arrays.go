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

func (l *lowering) viewArrayElementCheck(node, source *ast.Node) (*ir.ViewElementCheck, error) {
	target := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(source)))
	element := l.viewArrayElementType(target)
	if element == nil {
		return nil, nil
	}
	element = l.concrete(element)
	check := &ir.ViewElementCheck{Where: l.program.Where(node), ViewType: l.checker.TypeToString(element)}
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
	if err != nil || check == nil {
		return read, err
	}
	return ir.ViewElement{Array: read.Array, Index: read.Index, Relative: read.Relative, Contract: check.Contract, Where: check.Where, ViewType: check.ViewType, Of: read.Type()}, nil
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
	// Copies returned by map/filter and arrays storing viewed values preserve reach.
	for changed := true; changed; {
		changed = false
		visit := func(node any) bool {
			var source ir.Expression
			var sites []int
			switch value := node.(type) {
			case ir.ArrayMap:
				source = value.Array
				sites = value.GraphTypes
			case ir.ArrayVisit:
				if value.Method == "filter" {
					source = value.Array
					sites = value.GraphTypes
				}
			case ir.ArrayLiteral:
				for _, item := range value.Elements {
					if demanded(item) {
						source = item
						break
					}
				}
				sites = value.GraphTypes
			}
			if source != nil && demanded(source) && len(sites) > 0 {
				for _, site := range sites[len(sites)-1:] {
					if !viewed[site] {
						viewed[site] = true
						changed = true
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
	for _, use := range program.ViewArrayUses {
		if demanded(use.Array) {
			stop(use.Where, use.Member)
		}
	}
	rewrite := func(node any) any {
		var array ir.Expression
		var check *ir.ViewElementCheck
		road := viewArrayRoadOf(node)
		switch value := node.(type) {
		case ir.ViewElement:
			array = value.Array
			check = &ir.ViewElementCheck{Contract: value.Contract, Where: value.Where, ViewType: value.ViewType}
		case ir.ArrayPush:
			if value.UnknownElements {
				return node
			}
			array = value.Array
		case ir.ArrayIndex:
			array = value.Array
		case ir.ArrayVisit:
			array = value.Array
			check = value.ElementCheck
		case ir.ArrayMap:
			array = value.Array
			check = value.ElementCheck
		case ir.ArrayReduce:
			array = value.Array
			check = value.ElementCheck
		case ir.ForOf:
			if value.Iterable.Type() == ir.Array {
				array = value.Iterable
				check = value.ElementCheck
				if value.Pattern != nil {
					road.handout = false
					road.member = "destructuring"
				}
			}
		case ir.ArrayConcat:
			array = value.Array
			for _, other := range value.Others {
				if demanded(other) {
					stop(program.Source, "concat")
				}
			}
		case ir.JSONEncode:
			array = value.Value
			road.member = "JSON.stringify"
		case ir.JSONStringify:
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
		if array == nil || array.Type() != ir.Array {
			return node
		}
		needed := demanded(array)
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
				element := value.Of
				if value.Contract != 0 && program.ViewContracts[value.Contract-1].Of != 0 {
					element = program.ViewContracts[value.Contract-1].Of
				}
				return ir.ArrayIndex{Array: value.Array, Index: value.Index, Element: element, Relative: value.Relative}
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
	return refused
}

func (l *lowering) recordViewArrayUse(node *ast.Node, array ir.Expression, member string) {
	l.result.ViewArrayUses = append(l.result.ViewArrayUses, ir.ViewArrayUse{Array: array, Member: member, Where: l.program.Where(node)})
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
	if l.checker.IsArrayType(target) {
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
	source := l.concrete(l.checker.GetTypeAtLocation(node.AsAsExpression().Expression))
	if source.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0 || l.isLibraryType(source, "ReadonlyArray") {
		return l.notYet(node, "a cast that makes a viewed readonly array writable")
	}
	return nil
}

// Implicit handouts have the same failure summary as an explicit ViewElement.
func viewArrayHandoutChecks(node any) bool {
	switch value := node.(type) {
	case ir.ForOf:
		return value.ElementCheck != nil
	case ir.ArrayMap:
		return value.ElementCheck != nil
	case ir.ArrayVisit:
		return value.ElementCheck != nil
	case ir.ArrayReduce:
		return value.ElementCheck != nil
	}
	return false
}
