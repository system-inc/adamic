package lower

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// graphFlowClass is what graphFlows' follow does at an IR expression it reaches walking back from a
// graph-typed local, slot or result. graphFlowRules gives every expression type internal/ir declares
// one, and graph_flow_classified_test.go holds the map to the IR's source, so a type added there fails
// the guard until someone says which, and a type no rule names stops the program, never leaks. A
// value whose representation is a number, a boolean, a string or void is no graph value whatever its
// class (graphFlowScalar), so follow goes no further there.
type graphFlowClass int

const (
	// graphFlowPlanned allocates here and carries GraphTypes: follow plans its site, and enters what
	// its sources name where a form of it hands back an operand instead (an in-place fill).
	graphFlowPlanned graphFlowClass = iota + 1
	// graphFlowWrapper is one of its operands: follow enters each operand its value can be and plans
	// only the allocations inside, never the wrapper, since adamic_graph_adopt panics unless it's
	// handed a new allocation (runtime graph_regions.c:104).
	graphFlowWrapper
	// graphFlowForwards is a local's value, a function's result or a generator's yield: follow
	// demands it, as it demands a graph-typed local.
	graphFlowForwards
	// graphFlowElsewhere is a value planned where it was made or stored, not here: a slot's read,
	// whose stores collect demands, or a closure, which Function.GraphClosure plans.
	graphFlowElsewhere
	// graphFlowCannotHold is a value no graph value is: by its type, it holds nothing a graph member
	// can be reached through, or a cycle through it is refused before graph regions.
	graphFlowCannotHold
	// graphFlowStops allocates outside the plan, or is refused in a graph program: a graph value
	// through it is NotYet at the demanding site.
	graphFlowStops
)

// graphFlowSources is what a value can come from past the expression follow reached.
type graphFlowSources struct {
	// operands are entered in turn.
	operands []ir.Expression
	// locals are demanded, by their index.
	locals []int
	// results are the functions whose results are demanded.
	results []int
	// yields are the generator functions whose yielded values are demanded.
	yields []int
}

// graphFlowRule is one IR expression type's class, why it has that class, and, for a type whose value
// can be something else, what. sources' false is a form the rule doesn't name, which follow stops at.
type graphFlowRule struct {
	class   graphFlowClass
	why     string
	sources func(expression ir.Expression, program *ir.Program) (graphFlowSources, bool)
}

// graphFlowStep is what follow does at one expression: plan a site, enter its sources, or stop.
type graphFlowStep struct {
	planned bool
	site    int
	sources graphFlowSources
	stop    bool
}

// graphFlowStepOf is follow's whole decision at one expression, read off graphFlowRules alone.
func graphFlowStepOf(expression ir.Expression, program *ir.Program) graphFlowStep {
	rule, classified := graphFlowRules[reflect.TypeOf(expression)]
	if !classified {
		// The default arm: a type no rule names is a stop, never a silent return.
		return graphFlowStep{stop: true}
	}
	step := graphFlowStep{}
	// A form the rule doesn't name stops before its type is asked, since a method no rule knows
	// may type itself as void (ArrayVisit's Type).
	if rule.sources != nil {
		sources, known := rule.sources(expression, program)
		if !known {
			return graphFlowStep{stop: true}
		}
		step.sources = sources
	}
	if graphFlowScalar(expression.Type()) {
		return graphFlowStep{}
	}
	switch rule.class {
	case graphFlowPlanned:
		ids := reflect.ValueOf(expression).FieldByName("GraphTypes").Interface().([]int)
		if len(ids) == 0 {
			// graphAllocationSites gives every allocation a site before follow runs; one without is
			// an allocation this plan can't name.
			return graphFlowStep{stop: true}
		}
		step.planned, step.site = true, ids[len(ids)-1]
	case graphFlowStops:
		return graphFlowStep{stop: true}
	}
	return step
}

// graphFlowScalar is a representation no graph value has: void, a number, a boolean, a string, a
// Maybe pair of a number or a boolean, or a typed array's numbers.
func graphFlowScalar(of ir.Type) bool {
	switch of {
	case 0, ir.Number, ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean:
		return true
	}
	return of.IsTypedArray()
}

// graphFlowOperands is a rule's sources for a type whose value is some of its operands, as pick names
// them. A nil operand is skipped where follow enters it.
func graphFlowOperands[T ir.Expression](pick func(T) []ir.Expression) func(ir.Expression, *ir.Program) (graphFlowSources, bool) {
	return func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		return graphFlowSources{operands: pick(expression.(T))}, true
	}
}

// graphFlowRules classifies every IR expression type for graphFlows' follow, each with why. Lines cited
// are where the IR or its emitter says what the value is.
var graphFlowRules = map[reflect.Type]graphFlowRule{
	// Planned: an allocation carrying GraphTypes, whose site follow marks in program.GraphTypes.
	reflect.TypeFor[ir.ObjectLiteral](): {class: graphFlowPlanned, why: "a new object (ir.go:509); a class's is planned through Classes[].Graph too (native graph_regions.go:22)"},
	reflect.TypeFor[ir.ArrayLiteral]():  {class: graphFlowPlanned, why: "a new array (ir.go:590)"},
	reflect.TypeFor[ir.MapNew]():        {class: graphFlowPlanned, why: "a new Map (ir.go:979)"},
	reflect.TypeFor[ir.SetNew]():        {class: graphFlowPlanned, why: "a new Set (ir.go:1046)"},
	reflect.TypeFor[ir.ArrayMap]():      {class: graphFlowPlanned, why: "a new array of the callback's results (ir.go:913)"},
	reflect.TypeFor[ir.ArrayFrom]():     {class: graphFlowPlanned, why: "a new array of the callback's results (ir.go:830)"},
	reflect.TypeFor[ir.ArrayFill](): {class: graphFlowPlanned, why: "new Array(n).fill(v), a new array, when Array is nil; otherwise array.fill(v), which fills in place and is Array (ir.go:818, native emit_expressions.go:478), so follow enters it", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		fill := expression.(ir.ArrayFill)
		if fill.Array == nil {
			return graphFlowSources{}, true
		}
		return graphFlowSources{operands: []ir.Expression{fill.Array}}, true
	}},
	reflect.TypeFor[ir.ArraySplice](): {class: graphFlowPlanned, why: "a new array of what's removed (ir.go:805)"},
	reflect.TypeFor[ir.ArrayConcat](): {class: graphFlowPlanned, why: "a new array of every operand's elements (ir.go:845)"},
	reflect.TypeFor[ir.MapEntries]():  {class: graphFlowPlanned, why: "a new array of [key, value] pairs (ir.go:950)"},
	reflect.TypeFor[ir.ArraySlice]():  {class: graphFlowPlanned, why: "a new array of some of the elements (ir.go:959)"},
	reflect.TypeFor[ir.MapKeys]():     {class: graphFlowPlanned, why: "a new array of a map's keys (ir.go:991)"},
	reflect.TypeFor[ir.MapValues]():   {class: graphFlowPlanned, why: "a new array of a map's values (ir.go:998)"},
	reflect.TypeFor[ir.SetValues]():   {class: graphFlowPlanned, why: "a new array of a set's elements (ir.go:1064)"},
	reflect.TypeFor[ir.ArrayVisit](): {class: graphFlowPlanned, why: "filter's new array (ir.go:926); find and findLast are an element, planned where it was stored; the other methods are a boolean, a number or void", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		switch expression.(ir.ArrayVisit).Method {
		case "filter", "find", "findLast", "forEach", "some", "every", "findIndex", "findLastIndex":
			return graphFlowSources{}, true
		}
		return graphFlowSources{}, false
	}},
	reflect.TypeFor[ir.BoundClosure](): {class: graphFlowPlanned, why: "a new closure over captured cells (ir.go:882)"},
	reflect.TypeFor[ir.RecordCall](): {class: graphFlowPlanned, why: "values and entries make a new array (native emit_records.go:47); get is an entry, planned where it was stored; set is its value argument (native emit_records.go:25); keys is strings, and delete, has, hasOwn and defineHidden a boolean", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		call := expression.(ir.RecordCall)
		switch call.Method {
		case "values", "entries", "get", "keys", "delete", "has", "hasOwn", "defineHidden":
			return graphFlowSources{}, true
		case "set":
			if len(call.Arguments) == 3 {
				return graphFlowSources{operands: []ir.Expression{call.Arguments[2]}}, true
			}
		}
		return graphFlowSources{}, false
	}},

	// Wrappers: the value is one of the operands, and only an allocation inside is planned.
	reflect.TypeFor[ir.ArraySort]():    {class: graphFlowWrapper, why: "sorts in place and is its Array (ir.go:969, native emit_expressions.go:581); toSorted's copy is the ArraySlice inside", sources: graphFlowOperands(func(sort ir.ArraySort) []ir.Expression { return []ir.Expression{sort.Array} })},
	reflect.TypeFor[ir.ArrayReverse](): {class: graphFlowWrapper, why: "reverses in place and is its Array (ir.go:842, native emit_expressions.go:486); toReversed's copy is the ArraySlice inside", sources: graphFlowOperands(func(reverse ir.ArrayReverse) []ir.Expression { return []ir.Expression{reverse.Array} })},
	reflect.TypeFor[ir.Coalesce](): {class: graphFlowWrapper, why: "Value when it's present, or Fallback (ir.go:735); Panic is no value. Each arm is followed, never the join: groupBy's bucket ?? [] plans the literal, never an existing bucket", sources: graphFlowOperands(func(coalesce ir.Coalesce) []ir.Expression {
		return []ir.Expression{coalesce.Value, coalesce.Fallback}
	})},
	reflect.TypeFor[ir.Conditional](): {class: graphFlowWrapper, why: "one of its branches (ir.go:499); each arm is followed, never the join", sources: graphFlowOperands(func(conditional ir.Conditional) []ir.Expression {
		return []ir.Expression{conditional.WhenTrue, conditional.WhenNot}
	})},
	reflect.TypeFor[ir.Logical]():        {class: graphFlowWrapper, why: "returns an operand (taste.go:8); each arm is followed, never the join", sources: graphFlowOperands(func(logical ir.Logical) []ir.Expression { return []ir.Expression{logical.Left, logical.Right} })},
	reflect.TypeFor[ir.Defined]():        {class: graphFlowWrapper, why: "its Value, checked for undefined (ir.go:675)", sources: graphFlowOperands(func(defined ir.Defined) []ir.Expression { return []ir.Expression{defined.Value} })},
	reflect.TypeFor[ir.Comma]():          {class: graphFlowWrapper, why: "returns Right (taste.go:30)", sources: graphFlowOperands(func(comma ir.Comma) []ir.Expression { return []ir.Expression{comma.Right} })},
	reflect.TypeFor[ir.Effects]():        {class: graphFlowWrapper, why: "returns Result (taste.go:35); its Body's statements are collected as every statement is", sources: graphFlowOperands(func(effects ir.Effects) []ir.Expression { return []ir.Expression{effects.Result} })},
	reflect.TypeFor[ir.Box]():            {class: graphFlowWrapper, why: "a reference as itself where a Union goes (ir.go:695)", sources: graphFlowOperands(func(box ir.Box) []ir.Expression { return []ir.Expression{box.Value} })},
	reflect.TypeFor[ir.Narrow]():         {class: graphFlowWrapper, why: "a Union as one member (ir.go:698)", sources: graphFlowOperands(func(narrow ir.Narrow) []ir.Expression { return []ir.Expression{narrow.Value} })},
	reflect.TypeFor[ir.Unwrap]():         {class: graphFlowWrapper, why: "a Maybe pair proven present, as what it holds (ir.go:667)", sources: graphFlowOperands(func(unwrap ir.Unwrap) []ir.Expression { return []ir.Expression{unwrap.Value} })},
	reflect.TypeFor[ir.CheckedCast]():    {class: graphFlowWrapper, why: "its Value, its discriminant checked (ir.go:771)", sources: graphFlowOperands(func(cast ir.CheckedCast) []ir.Expression { return []ir.Expression{cast.Value} })},
	reflect.TypeFor[ir.PlaceholderUse](): {class: graphFlowWrapper, why: "its Value, checked set before its use (ir.go:122)", sources: graphFlowOperands(func(use ir.PlaceholderUse) []ir.Expression { return []ir.Expression{use.Value} })},
	reflect.TypeFor[ir.IterableOf]():     {class: graphFlowWrapper, why: "the generator its Value made, a pointer cast natively (generator.go:47)", sources: graphFlowOperands(func(iterable ir.IterableOf) []ir.Expression { return []ir.Expression{iterable.Value} })},
	reflect.TypeFor[ir.MapSet]():         {class: graphFlowWrapper, why: "map.set(k, v) is the map (ir.go:1026); the value it stores is a slot write collect demands", sources: graphFlowOperands(func(set ir.MapSet) []ir.Expression { return []ir.Expression{set.Map} })},
	reflect.TypeFor[ir.SetAdd]():         {class: graphFlowWrapper, why: "set.add(v) is the set (ir.go:1054); the value it stores is a slot write collect demands", sources: graphFlowOperands(func(add ir.SetAdd) []ir.Expression { return []ir.Expression{add.Set} })},
	reflect.TypeFor[ir.Await]():          {class: graphFlowWrapper, why: "the value its promise settles with (async.go:5): a call's result, or a settled PromiseValue's Value", sources: graphFlowOperands(func(await ir.Await) []ir.Expression { return []ir.Expression{await.Value} })},
	reflect.TypeFor[ir.PromiseValue]():   {class: graphFlowWrapper, why: "a settled promise of Value (async.go:14), reached only through an Await, which is Value; a Promise itself never joins a graph (graph_types.go:234)", sources: graphFlowOperands(func(promise ir.PromiseValue) []ir.Expression { return []ir.Expression{promise.Value} })},
	reflect.TypeFor[ir.RecordCoalesce](): {class: graphFlowWrapper, why: "r[k] ??= value: the entry, planned where it was stored, or Value, stored and given back (records.go:33)", sources: graphFlowOperands(func(coalesce ir.RecordCoalesce) []ir.Expression { return []ir.Expression{coalesce.Value} })},
	reflect.TypeFor[ir.LogicalAssignment](): {class: graphFlowWrapper, why: "Read's value, or the value Write stores (taste.go:17, native logical_assignment.go:14)", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		assignment := expression.(ir.LogicalAssignment)
		switch write := assignment.Write.(type) {
		case ir.Assign:
			return graphFlowSources{operands: []ir.Expression{assignment.Read, write.Value}}, true
		case ir.SetProperty:
			return graphFlowSources{operands: []ir.Expression{assignment.Read, write.Value}}, true
		case ir.SetIndex:
			return graphFlowSources{operands: []ir.Expression{assignment.Read, write.Value}}, true
		}
		return graphFlowSources{}, false
	}},
	reflect.TypeFor[ir.ObjectCall](): {class: graphFlowWrapper, why: "assign and freeze are their first argument (native library_object.go:20, :44); keys, values and entries are arrays of numbers, strings or booleans (lower library_object.go:159), and is, isFrozen and hasOwn a boolean", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		call := expression.(ir.ObjectCall)
		switch call.Method {
		case "assign", "freeze":
			if len(call.Arguments) > 0 {
				return graphFlowSources{operands: []ir.Expression{call.Arguments[0]}}, true
			}
		case "keys", "values", "entries", "is", "isFrozen", "hasOwn":
			return graphFlowSources{}, true
		}
		return graphFlowSources{}, false
	}},

	// Forwards: a local's value, a function's result or a generator's yield, demanded.
	reflect.TypeFor[ir.Read](): {class: graphFlowForwards, why: "a local's value (ir.go:430)", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		return graphFlowSources{locals: []int{expression.(ir.Read).Local}}, true
	}},
	reflect.TypeFor[ir.Call](): {class: graphFlowForwards, why: "its targets' results (ir.go:442)", sources: func(expression ir.Expression, program *ir.Program) (graphFlowSources, bool) {
		return graphFlowSources{results: program.CallTargets(expression.(ir.Call))}, true
	}},
	reflect.TypeFor[ir.CallClosure](): {class: graphFlowForwards, why: "its targets' results (ir.go:900); a function value whose targets are Unknown plans its result at each target's Return, and collect demands each argument at the call (#41vb71c)", sources: func(expression ir.Expression, program *ir.Program) (graphFlowSources, bool) {
		return graphFlowSources{results: program.ClosureTargets(expression.(ir.CallClosure)).Functions}, true
	}},
	reflect.TypeFor[ir.BoundMethod](): {class: graphFlowForwards, why: "the closure its Factory function makes and returns (native emit_expressions.go:976)", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		return graphFlowSources{results: []int{expression.(ir.BoundMethod).Factory}}, true
	}},
	reflect.TypeFor[ir.ArrayReduce](): {class: graphFlowForwards, why: "Initial or what the callback last returned (ir.go:855); a callback whose targets are Unknown plans its result at each target's Return, and collect demands Initial and the array it's handed at the reduce (#41vb71c)", sources: func(expression ir.Expression, program *ir.Program) (graphFlowSources, bool) {
		reduce := expression.(ir.ArrayReduce)
		return graphFlowSources{operands: []ir.Expression{reduce.Initial}, results: program.ClosureTargets(reduce).Functions}, true
	}},
	reflect.TypeFor[ir.GeneratorNext](): {class: graphFlowForwards, why: "the value Function's body yields (generator.go:13)", sources: func(expression ir.Expression, _ *ir.Program) (graphFlowSources, bool) {
		return graphFlowSources{yields: []int{expression.(ir.GeneratorNext).Function}}, true
	}},

	// Elsewhere: planned where it was made or stored, not here.
	reflect.TypeFor[ir.Property]():    {class: graphFlowElsewhere, why: "reads a field (ir.go:551), whose stores collect demands; a getter is a Call"},
	reflect.TypeFor[ir.ArrayIndex]():  {class: graphFlowElsewhere, why: "reads an element (ir.go:785), whose stores collect demands"},
	reflect.TypeFor[ir.ArrayPop]():    {class: graphFlowElsewhere, why: "an element taken out of its array (ir.go:869), planned where it was stored"},
	reflect.TypeFor[ir.MapGet]():      {class: graphFlowElsewhere, why: "reads a map's value (ir.go:1019), whose stores collect demands"},
	reflect.TypeFor[ir.ViewElement](): {class: graphFlowElsewhere, why: "the same element after its contract check (view_elements.go:3)"},
	reflect.TypeFor[ir.WeakTarget]():  {class: graphFlowElsewhere, why: "the target a Weak names (ir.go:722), held strongly where it was stored or already freed"},
	reflect.TypeFor[ir.MakeClosure](): {class: graphFlowElsewhere, why: "a new closure planned by its Function's GraphClosure (graph_types.go:289, native emit_expressions.go:948)"},
	reflect.TypeFor[ir.ClosureSelf](): {class: graphFlowElsewhere, why: "the running closure itself (library_language.go:9), made by a MakeClosure"},

	// Cannot hold: no graph value is one, by its type.
	reflect.TypeFor[ir.NumberConstant]():           {class: graphFlowCannotHold, why: "a number (ir.go:1159)"},
	reflect.TypeFor[ir.BooleanConstant]():          {class: graphFlowCannotHold, why: "a boolean (ir.go:1160)"},
	reflect.TypeFor[ir.StringConstant]():           {class: graphFlowCannotHold, why: "a string (ir.go:1161)"},
	reflect.TypeFor[ir.NumberToString]():           {class: graphFlowCannotHold, why: "a string (ir.go:1162)"},
	reflect.TypeFor[ir.BooleanToString]():          {class: graphFlowCannotHold, why: "a string (ir.go:1163)"},
	reflect.TypeFor[ir.Concat]():                   {class: graphFlowCannotHold, why: "a string (ir.go:1164)"},
	reflect.TypeFor[ir.HasAccessor]():              {class: graphFlowCannotHold, why: "a boolean (ir.go:1157)"},
	reflect.TypeFor[ir.InstanceOf]():               {class: graphFlowCannotHold, why: "a boolean (ir.go:1156)"},
	reflect.TypeFor[ir.Unary]():                    {class: graphFlowCannotHold, why: "a number or a boolean (ir.go:1307)"},
	reflect.TypeFor[ir.Binary]():                   {class: graphFlowCannotHold, why: "a number, a boolean or a maybe boolean (ir.go:1314)"},
	reflect.TypeFor[ir.Length]():                   {class: graphFlowCannotHold, why: "a number or a maybe number (ir.go:1179)"},
	reflect.TypeFor[ir.MathCall]():                 {class: graphFlowCannotHold, why: "a number (ir.go:1185)"},
	reflect.TypeFor[ir.StringFromCodes]():          {class: graphFlowCannotHold, why: "a string (ir.go:1186)"},
	reflect.TypeFor[ir.NumberCall]():               {class: graphFlowCannotHold, why: "a number or a boolean (ir.go:1188)"},
	reflect.TypeFor[ir.ToFixed]():                  {class: graphFlowCannotHold, why: "a string (ir.go:1194)"},
	reflect.TypeFor[ir.NumberFormat]():             {class: graphFlowCannotHold, why: "a string (ir.go:1195)"},
	reflect.TypeFor[ir.Undefined]():                {class: graphFlowCannotHold, why: "undefined, a null reference, no allocation (ir.go:1196)"},
	reflect.TypeFor[ir.Null]():                     {class: graphFlowCannotHold, why: "null, a null reference, no allocation (regexp.go:57)"},
	reflect.TypeFor[ir.IsUndefined]():              {class: graphFlowCannotHold, why: "a boolean (ir.go:1202)"},
	reflect.TypeFor[ir.ArrayPush]():                {class: graphFlowCannotHold, why: "the new length, a number (ir.go:1203); the value it stores is a slot write collect demands"},
	reflect.TypeFor[ir.ArrayJoin]():                {class: graphFlowCannotHold, why: "a string (ir.go:1204)"},
	reflect.TypeFor[ir.MaybeOf]():                  {class: graphFlowCannotHold, why: "a number or a boolean that may be missing (ir.go:1207)"},
	reflect.TypeFor[ir.MaybeToString]():            {class: graphFlowCannotHold, why: "a string (ir.go:1208)"},
	reflect.TypeFor[ir.TypeOf]():                   {class: graphFlowCannotHold, why: "a string (ir.go:1211)"},
	reflect.TypeFor[ir.UnionToString]():            {class: graphFlowCannotHold, why: "a string (ir.go:1212)"},
	reflect.TypeFor[ir.WeakOf]():                   {class: graphFlowCannotHold, why: "a Weak handle (ir.go:1213), never a strong edge (graph_types.go:172), and never graph-typed (graph_flow.go's graphType)"},
	reflect.TypeFor[ir.MakeError]():                {class: graphFlowCannotHold, why: "an Error whose message and name are strings (ir.go:1214)"},
	reflect.TypeFor[ir.StringLength]():             {class: graphFlowCannotHold, why: "a number or a maybe number (ir.go:1217)"},
	reflect.TypeFor[ir.CharCodeAt]():               {class: graphFlowCannotHold, why: "a number (ir.go:1223)"},
	reflect.TypeFor[ir.Trim]():                     {class: graphFlowCannotHold, why: "a string (ir.go:1224)"},
	reflect.TypeFor[ir.CodePoints]():               {class: graphFlowCannotHold, why: "a new array of strings (ir.go:1225)"},
	reflect.TypeFor[ir.StringIndex]():              {class: graphFlowCannotHold, why: "a string (ir.go:1226)"},
	reflect.TypeFor[ir.StringCall]():               {class: graphFlowCannotHold, why: "a string, a number, a boolean, or split's array of strings (ir.go:1267)"},
	reflect.TypeFor[ir.ArraySearch]():              {class: graphFlowCannotHold, why: "a number or a boolean (ir.go:1253)"},
	reflect.TypeFor[ir.ReadTextFile]():             {class: graphFlowCannotHold, why: "{ kind, text } or { kind, message }, strings (ir.go:1280)"},
	reflect.TypeFor[ir.ProgramArguments]():         {class: graphFlowCannotHold, why: "a new array of strings (ir.go:1281)"},
	reflect.TypeFor[ir.Utf8Length]():               {class: graphFlowCannotHold, why: "a number (ir.go:1282)"},
	reflect.TypeFor[ir.Utf8At]():                   {class: graphFlowCannotHold, why: "a number (ir.go:1283)"},
	reflect.TypeFor[ir.WriteTextFile]():            {class: graphFlowCannotHold, why: "{ kind } or { kind, message }, strings (ir.go:1284)"},
	reflect.TypeFor[ir.ReadDirectory]():            {class: graphFlowCannotHold, why: "{ kind, names } or { kind, message }, strings and an array of strings (ir.go:1285)"},
	reflect.TypeFor[ir.FileStatus]():               {class: graphFlowCannotHold, why: "{ kind, type, size, symbolicLink } or { kind, message }, strings, a number and a boolean (ir.go:1286)"},
	reflect.TypeFor[ir.MapClear]():                 {class: graphFlowCannotHold, why: "void (ir.go:1291)"},
	reflect.TypeFor[ir.MapForEach]():               {class: graphFlowCannotHold, why: "void (ir.go:1292)"},
	reflect.TypeFor[ir.MapHas]():                   {class: graphFlowCannotHold, why: "a boolean (ir.go:1297)"},
	reflect.TypeFor[ir.MapDelete]():                {class: graphFlowCannotHold, why: "a boolean (ir.go:1298)"},
	reflect.TypeFor[ir.MapSize]():                  {class: graphFlowCannotHold, why: "a number (ir.go:1299)"},
	reflect.TypeFor[ir.HasOwn]():                   {class: graphFlowCannotHold, why: "a boolean (ir.go:1300)"},
	reflect.TypeFor[ir.Truthy]():                   {class: graphFlowCannotHold, why: "a boolean (taste.go:6)"},
	reflect.TypeFor[ir.Void]():                     {class: graphFlowCannotHold, why: "undefined, its operand dropped (taste.go:57)"},
	reflect.TypeFor[ir.Yield]():                    {class: graphFlowCannotHold, why: "void, a statement only (generator.go:11); what it hands the consumer is a GeneratorNext's, which follow demands"},
	reflect.TypeFor[ir.GeneratorDone]():            {class: graphFlowCannotHold, why: "a boolean (generator.go:78)"},
	reflect.TypeFor[ir.IsGenerator]():              {class: graphFlowCannotHold, why: "a boolean (generator.go:64)"},
	reflect.TypeFor[ir.IsCollectionIterator]():     {class: graphFlowCannotHold, why: "a boolean (generator.go:71)"},
	reflect.TypeFor[ir.ArrayIsArray]():             {class: graphFlowCannotHold, why: "a boolean (array_is_array.go:6)"},
	reflect.TypeFor[ir.HasProperty]():              {class: graphFlowCannotHold, why: "a boolean (unknown.go:9)"},
	reflect.TypeFor[ir.UnknownKind]():              {class: graphFlowCannotHold, why: "a number (unknown.go:36)"},
	reflect.TypeFor[ir.OwnFunction]():              {class: graphFlowCannotHold, why: "a boolean (unknown.go:66)"},
	reflect.TypeFor[ir.ErrorToString]():            {class: graphFlowCannotHold, why: "a string (unknown.go:73)"},
	reflect.TypeFor[ir.ArrayOfScalars]():           {class: graphFlowCannotHold, why: "a boolean (unknown.go:83)"},
	reflect.TypeFor[ir.JoinEnter]():                {class: graphFlowCannotHold, why: "a boolean (unknown.go:90)"},
	reflect.TypeFor[ir.JoinLeave]():                {class: graphFlowCannotHold, why: "void (unknown.go:95)"},
	reflect.TypeFor[ir.RecordStorage]():            {class: graphFlowCannotHold, why: "a boolean (records.go:51)"},
	reflect.TypeFor[ir.RecordLiteral]():            {class: graphFlowCannotHold, why: "a new record (records.go:31), never reached here: graphType selects no record type (only arrays, tuples, maps and sets join as caches, graph_types.go:262), so no demand follow walks names one. A record held on a cycle is the cycle rule's, not this pass's: #a9e21hc measured it leaking on main and adds its stop"},
	reflect.TypeFor[ir.RegExpNew]():                {class: graphFlowCannotHold, why: "a pattern of its source and flags (regexp.go:12)"},
	reflect.TypeFor[ir.RegExpCall]():               {class: graphFlowCannotHold, why: "a string, a number, a boolean, or match arrays of strings and their iterator (regexp.go:23)"},
	reflect.TypeFor[ir.RegExpProperty]():           {class: graphFlowCannotHold, why: "a match's index, input or groups, numbers and strings (regexp.go:47)"},
	reflect.TypeFor[ir.RegExpGroup]():              {class: graphFlowCannotHold, why: "a named group's string or undefined (regexp.go:79)"},
	reflect.TypeFor[ir.IsNull]():                   {class: graphFlowCannotHold, why: "a boolean (regexp.go:69)"},
	reflect.TypeFor[ir.StringIterator]():           {class: graphFlowCannotHold, why: "an iterator holding a string (library_map_set.go:17)"},
	reflect.TypeFor[ir.ObjectKeys]():               {class: graphFlowCannotHold, why: "a new array of strings (library_language.go:7)"},
	reflect.TypeFor[ir.LibraryGlobal]():            {class: graphFlowCannotHold, why: "a built-in's identity (library_language.go:19), which holds no program value"},
	reflect.TypeFor[ir.IntlCollatorNew]():          {class: graphFlowCannotHold, why: "a host collator holding its locale and its own compare (library_intl.go:10)"},
	reflect.TypeFor[ir.IntlCollatorCompareValue](): {class: graphFlowCannotHold, why: "the collator's own compare, which captures its locale and never the program's (library_intl.go:20)"},
	reflect.TypeFor[ir.JSONEncode]():               {class: graphFlowCannotHold, why: "a string (library_json_encode.go:9)"},
	reflect.TypeFor[ir.JSONStringify]():            {class: graphFlowCannotHold, why: "a string (library_json_stringify.go:24)"},
	reflect.TypeFor[ir.JSONNull]():                 {class: graphFlowCannotHold, why: "a null constant inside stringify's input (library_json_stringify.go:29)"},
	reflect.TypeFor[ir.PhantomMember]():            {class: graphFlowCannotHold, why: "undefined, or a throw; no brand object is allocated (phantom_brands.go:13)"},
	reflect.TypeFor[ir.TypedArrayNew]():            {class: graphFlowCannotHold, why: "a typed array, numbers in one buffer (typed_arrays.go:38)"},
	reflect.TypeFor[ir.TypedArrayFill]():           {class: graphFlowCannotHold, why: "its typed array, numbers (typed_arrays.go:39)"},
	reflect.TypeFor[ir.TypedArraySort]():           {class: graphFlowCannotHold, why: "its typed array, numbers (typed_arrays.go:40)"},
	reflect.TypeFor[ir.TypedArraySet]():            {class: graphFlowCannotHold, why: "void (typed_arrays.go:41)"},
	reflect.TypeFor[ir.TypedArraySubarray]():       {class: graphFlowCannotHold, why: "a view of a typed array's numbers (typed_arrays.go:42)"},

	// Stops: made outside the plan, or refused in a graph program.
	reflect.TypeFor[ir.IteratorNext]():         {class: graphFlowStops, why: "a fresh { value, done } (generator.go:27) with no GraphTypes, which holds the yielded value: a graph type when the program holds results (Memory's row 12)"},
	reflect.TypeFor[ir.CollectionIterator]():   {class: graphFlowElsewhere, why: "a new iterator holding its collection (library_map_set.go:3), linked to it in the cycle proof (graph_types.go:401), which the runtime plans itself: adamic_collection_iterator adopts the iterator, its state and its closure whenever its collection is a graph member (runtime map_set.c:178-199), as adamic_map_entries adopts a graph map's pairs; fresh_refused's map_iterator and set_iterator are its witnesses"},
	reflect.TypeFor[ir.JSONDecode]():           {class: graphFlowStops, why: "a decoded tree the runtime makes outside every region (library_json_decode.go:42), whose type can be a graph type"},
	reflect.TypeFor[ir.ParallelMap]():          {class: graphFlowStops, why: "a new array of the callbacks' results (ir.go:1091) with no GraphTypes"},
	reflect.TypeFor[ir.DynamicProperty]():      {class: graphFlowStops, why: "a dynamic-shape read, which graphAdoptionBoundary refuses in a graph program (graph_types.go:475) before this stop is read"},
	reflect.TypeFor[ir.DynamicKeyedProperty](): {class: graphFlowStops, why: "a dynamic-shape read, which graphAdoptionBoundary refuses in a graph program (graph_types.go:475) before this stop is read"},
	reflect.TypeFor[ir.NodeFSFile]():           {class: graphFlowStops, why: "a host result made outside the plan (node_fs_file.go:3); one that's a string, a number or a boolean is no graph value"},
	reflect.TypeFor[ir.NodeHostCall]():         {class: graphFlowStops, why: "a host result made outside the plan (node_fs_directory.go:3); one that's a string, a number or a boolean is no graph value"},
	reflect.TypeFor[ir.ProcessCall]():          {class: graphFlowStops, why: "a host result made outside the plan (process.go:3); one that's a string, a number or a boolean is no graph value"},
	reflect.TypeFor[ir.NodeBufferCall]():       {class: graphFlowStops, why: "a host result made outside the plan (node_buffer.go:3); one that's a string, a number or a boolean is no graph value"},
}
