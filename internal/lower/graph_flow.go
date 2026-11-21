package lower

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// graphAllocationSites copies the IR tree before assigning unique allocation
// IDs. Pointer and map children must be copied too, so their sites do not depend
// solely on a checker identity. Nil containers preserve their original shape.
func graphAllocationSites(value reflect.Value, next *int) reflect.Value {
	var sites func(reflect.Value) reflect.Value
	sites = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := sites(value.Elem())
			result := reflect.New(value.Type()).Elem()
			result.Set(mapped)
			return result
		case reflect.Ptr:
			if value.IsNil() {
				return value
			}
			result := reflect.New(value.Type().Elem())
			result.Elem().Set(sites(value.Elem()))
			return result
		case reflect.Map:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeMapWithSize(value.Type(), value.Len())
			entries := value.MapRange()
			for entries.Next() {
				result.SetMapIndex(sites(entries.Key()), sites(entries.Value()))
			}
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(sites(value.Field(i)))
			}
			if field := result.FieldByName("GraphTypes"); field.IsValid() && field.Type() == reflect.TypeOf([]int{}) {
				*next--
				field.Set(reflect.ValueOf(append(append([]int{}, field.Interface().([]int)...), *next)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(sites(value.Index(i)))
			}
			return result
		case reflect.Array:
			// A fixed array of expressions, as MapNew's [key, value] entries are, holds allocations too:
			// left uncopied, they had no site and follow stopped on them (#dnb0yhj, s1).
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(sites(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	return sites(value)
}

// graphLocalWrites is how graphFlows' collect takes every IR field that writes a local, keyed as
// view_flow.go's flowLocalWrites is: empty for a source, a node whose Value collect seeds as the
// local's (Declare and Assign), and otherwise how collect takes it, and why. Its rows are exactly
// local_uses.go's localFields' write rows (TestGraphLocalWritesAreLocalFieldsWrites), so an IR node
// that comes to write a local is classified here before a graph value can reach a demanded local
// through a writer collect doesn't follow, a leak or worse (#hny64km).
var graphLocalWrites = map[string]string{
	"Declare.Local": "",
	"Assign.Local":  "",
	// collect's ForOf case.
	"ForOf.Local": "over an array literal, its elements are the binding's sources; over anything else, an element was planned where it was stored",
	// graphUnknownWriters, which a demand of the local stops at.
	"Binding.Local":  "a pattern over an array literal is unknown, its elements' parts never seeded; over anything else, an element was planned where it was stored",
	"Try.CatchLocal": "unknown: the binding takes what was thrown",
}

// graphUnknownWriters records, by local plus one, each local node writes with a value collect doesn't
// follow, with the stop a graph-typed demand of it is (graphLocalWrites' unknown rows). A thrown value
// is always an Error today (exceptions.go) and none is graph-typed, so a catch's stop is defense; a
// pattern over an array literal binds its elements' parts, which no source names (#hny64km).
func graphUnknownWriters(node any, unknown map[int]string) {
	switch value := node.(type) {
	case ir.Try:
		if value.HasCatch && value.CatchLocal >= 0 {
			unknown[value.CatchLocal+1] = "a graph value a catch binding takes, which the plan doesn't follow"
		}
	case ir.ForOf:
		if _, isLiteral := value.Iterable.(ir.ArrayLiteral); isLiteral {
			for _, binding := range value.Pattern {
				unknown[binding.Local+1] = "a graph value a for-of pattern takes from an array literal, which the plan doesn't follow"
			}
		}
	}
}

// graphFlows works backwards from graph views to the allocations that can supply
// them. Local assignments and function results are joined across all paths; this
// is a may-flow proof, not an execution or a choice of one conditional branch.
// Negative IDs name allocation sites in this IR, independently of fresh checker
// identities. They select the existing allocation emission without runtime work.
// follow classifies every IR expression by graphFlowRules (graph_flow_classes.go), and a graph
// value through one the plan doesn't follow is NotYet at the site that demanded it, never a leak.
// collect takes every local-writing field by graphLocalWrites, and a demanded local some node writes
// with a value collect doesn't follow is NotYet at the local the same way (#hny64km). An argument
// passed to a function value whose targets are Unknown is followed too, quietly: it plans what it
// reaches and never stops (#41vb71c). One whose every type can't reach a graph type (mayReachGraph)
// holds no graph value and isn't followed (#x490w4a).
func (f *cycleFinder) graphFlows() error {
	program := f.l.result
	next := 0
	program.Main = graphAllocationSites(reflect.ValueOf(program.Main), &next).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = graphAllocationSites(reflect.ValueOf(program.Functions[i].Body), &next).Interface().([]ir.Statement)
	}

	// Positive nodes are locals, then function results, then generator functions' yields.
	resultNode := func(function int) int { return len(program.Locals) + function + 1 }
	yieldNode := func(function int) int { return len(program.Locals) + len(program.Functions) + function + 1 }
	sources := map[int][]ir.Expression{}
	// unknownWriter is each local some node writes with a value collect doesn't follow
	// (graphUnknownWriters), by node: a demand of one stops rather than plans only its other writers.
	unknownWriter := map[int]string{}
	wanted := map[int]bool{}
	queue := []int{}
	demand := func(node int) {
		if !wanted[node] {
			wanted[node] = true
			queue = append(queue, node)
		}
	}
	// A quiet demand plans and never stops. It runs once every graph-typed demand has, so a node
	// already followed for its type isn't followed again.
	quietlyWanted := map[int]bool{}
	quietQueue := []int{}
	demandQuietly := func(node int) {
		if !wanted[node] && !quietlyWanted[node] {
			quietlyWanted[node] = true
			quietQueue = append(quietQueue, node)
		}
	}
	for local, proven := range f.l.localTypes {
		if f.graphType(proven) {
			demand(local + 1)
		}
	}
	for local, types := range f.l.localAlso {
		for _, proven := range types {
			if f.graphType(proven) {
				demand(local + 1)
			}
		}
	}
	// A graphDemand is what asked for a graph value: a positive node, or a write site's slot. A stop
	// is reported where it is.
	type graphDemand struct{ node, site int }
	type graphPending struct {
		expression ir.Expression
		at         graphDemand
	}
	pending := []graphPending{}
	// quietly is every argument passed to a function value whose targets are Unknown: no parameter it
	// reaches has it as a source, so it's demanded at the call (#41vb71c), unless it holds no graph
	// value. mayHold says whether an argument lowering numbered id can: it named no types for it, or one
	// of the types argumentTypes keeps for it can reach a graph type (#x490w4a).
	quietly := []ir.Expression{}
	judged := map[*checker.Type]bool{}
	mayHold := func(id int) bool {
		if id <= 0 || id > len(f.l.argumentTypes) || len(f.l.argumentTypes[id-1]) == 0 {
			return true
		}
		for _, proven := range f.l.argumentTypes[id-1] {
			if f.mayReachGraph(proven, judged) {
				return true
			}
		}
		return false
	}
	// Site retains the checker view of the actual slot being written. slotType is that slot's type,
	// or nil, and slot says whether it's a graph type.
	slotType := func(site int, name string, argument int) *checker.Type {
		if site <= 0 || site > len(f.l.writeSites) {
			return nil
		}
		holder := f.l.writeSites[site-1].holder
		if holder == nil {
			return nil
		}
		if name != "" {
			for _, field := range f.fields(holder) {
				if f.l.cycleFieldMatches(holder, field.Name, name) {
					return f.l.checker.GetTypeOfSymbol(field)
				}
			}
			return nil
		}
		args := f.graphArguments(holder)
		if argument < len(args) {
			return args[argument]
		}
		return nil
	}
	slot := func(site int, name string, argument int) bool {
		return f.graphType(slotType(site, name, argument))
	}
	// recordSlot is the value type of the record a write site writes into, or nil when its holder isn't
	// a record (#1vb0x5s).
	recordSlot := func(site int) *checker.Type {
		if site <= 0 || site > len(f.l.writeSites) {
			return nil
		}
		return f.l.recordElement(f.l.writeSites[site-1].holder)
	}
	// A counted aggregate can have graph-typed fields/elements. Seed those
	// initializer slots individually, rather than promoting its whole payload.
	returns := map[int]*checker.Type{}
	for symbol, index := range f.l.functions {
		signatures := f.l.checker.GetSignaturesOfType(f.l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[index] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	for _, closure := range f.l.closureRecords {
		signatures := f.l.checker.GetSignaturesOfType(closure.proven, checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[closure.function] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	var members func(ir.Expression, *checker.Type, graphDemand)
	members = func(expression ir.Expression, proven *checker.Type, at graphDemand) {
		if expression == nil || proven == nil || f.weak(proven) {
			return
		}
		if f.graphType(proven) {
			pending = append(pending, graphPending{expression, at})
		}
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			for _, field := range f.fields(proven) {
				for _, initializer := range value.Fields {
					if f.l.cycleFieldMatches(proven, field.Name, initializer.Name) {
						members(initializer.Value, f.l.checker.GetTypeOfSymbol(field), at)
					}
				}
			}
		case ir.ArrayLiteral:
			arguments := f.graphArguments(proven)
			if len(arguments) > 0 {
				for _, element := range value.Elements {
					members(element, arguments[0], at)
				}
			}
		case ir.MapNew:
			// A new Map's entries are written into it as it's made (#dnb0yhj, s1).
			arguments := f.graphArguments(proven)
			if len(arguments) == 2 {
				for _, entry := range value.Entries {
					members(entry[0], arguments[0], at)
					members(entry[1], arguments[1], at)
				}
			}
		case ir.RecordLiteral:
			// A new record's entries are written into it as it's made, each by the value type of the record
			// it's held as, every record arm's of a union; its spread's values were planned where they were
			// stored (#1vb0x5s).
			for _, element := range f.recordElements(proven) {
				for _, entry := range value.Entries {
					members(entry.Value, element, at)
				}
			}
		case ir.Conditional:
			members(value.WhenTrue, proven, at)
			members(value.WhenNot, proven, at)
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			graphUnknownWriters(node, unknownWriter)
			switch value := node.(type) {
			case ir.Declare:
				members(value.Value, f.l.localTypes[value.Local], graphDemand{node: value.Local + 1})
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Assign:
				members(value.Value, f.l.localTypes[value.Local], graphDemand{node: value.Local + 1})
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Return:
				at := graphDemand{}
				if function >= 0 {
					at.node = resultNode(function)
					sources[at.node] = append(sources[at.node], value.Value)
				}
				members(value.Value, returns[function], at)
			case ir.Yield:
				// A generator's step is the value its body yields (GeneratorNext's rule).
				if function >= 0 {
					sources[yieldNode(function)] = append(sources[yieldNode(function)], value.Value)
				}
			case ir.Call:
				for _, target := range program.CallTargets(value) {
					if target < 0 || target >= len(program.Functions) {
						continue
					}
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if targets.Unknown {
					// A spread argument is its array, the operand followed here like any other, and its id is 0.
					for i, argument := range value.Arguments {
						if i >= len(value.ArgumentTypeIDs) || mayHold(value.ArgumentTypeIDs[i]) {
							quietly = append(quietly, argument)
						}
					}
					break
				}
				for _, target := range targets.Functions {
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.ArrayReduce:
				// The callback is handed Initial as its first accumulator and the array itself; its
				// elements were planned where they were stored.
				if program.ClosureTargets(value).Unknown {
					if mayHold(value.InitialTypeID) {
						quietly = append(quietly, value.Initial)
					}
					if mayHold(value.ArrayTypeID) {
						quietly = append(quietly, value.Array)
					}
				}
			// A slot write is seeded through members, as a declaration's value is, so a literal written
			// there seeds its own graph-typed parts too: new Map's entries among them (#dnb0yhj, s1).
			// A SetProperty or a SetIndex whose holder is a record writes its value into the record's table,
			// by the record's value type. Neither does today: a record's writes are RecordCall's and
			// RecordCoalesce's below, and the entries form's SetProperty{Record} is refused in a graph
			// program with its ObjectLiteral{Record} (graphAdoptionBoundary).
			case ir.SetProperty:
				proven := slotType(value.Site, value.Name, 0)
				if proven == nil {
					proven = recordSlot(value.Site)
				}
				members(value.Value, proven, graphDemand{site: value.Site})
			case ir.SetIndex:
				proven := slotType(value.Site, "", 0)
				if proven == nil {
					proven = recordSlot(value.Site)
				}
				members(value.Value, proven, graphDemand{site: value.Site})
			// A record's writes store into its table, the runtime taking what they store into a graph
			// record's region (runtime record.c's take_in): set's and defineHidden's value, Arguments[2],
			// and ??='s Value, each demanded at its write site by the record's value type, as a Map's value
			// is (#1vb0x5s). A record's delete stores nothing. Object.assign never takes a record target (it
			// stops, records.go's recordObjectCallArguments), and no other library call writes one.
			case ir.RecordCall:
				if (value.Method == "set" || value.Method == "defineHidden") && len(value.Arguments) == 3 {
					members(value.Arguments[2], recordSlot(value.Site), graphDemand{site: value.Site})
				}
			case ir.RecordCoalesce:
				members(value.Value, recordSlot(value.Site), graphDemand{site: value.Site})
			case ir.ArrayPush:
				members(value.Value, slotType(value.Site, "", 0), graphDemand{site: value.Site})
			// fill in place and splice's inserted items write into the array as a push does (#dnb0yhj,
			// s2 and s3). A new array's fill is the array, which follow enters when it's demanded.
			case ir.ArrayFill:
				if value.Array != nil {
					members(value.Value, slotType(value.Site, "", 0), graphDemand{site: value.Site})
				}
			case ir.ArraySplice:
				for _, item := range value.Items {
					members(item, slotType(value.Site, "", 0), graphDemand{site: value.Site})
				}
			// A for-of binding over an array literal is each of its elements, the binding's sources as
			// an assignment's value would be (#dnb0yhj, s5). Over anything else, an element was stored
			// by a write this seeds where it was made. A pattern's locals aren't sources here:
			// graphUnknownWriters takes them (graphLocalWrites).
			case ir.ForOf:
				if literal, isLiteral := value.Iterable.(ir.ArrayLiteral); isLiteral && value.Pattern == nil {
					sources[value.Local+1] = append(sources[value.Local+1], literal.Elements...)
				}
			case ir.MapSet:
				if slot(value.Site, "", 0) {
					pending = append(pending, graphPending{value.Key, graphDemand{site: value.Site}})
				}
				if slot(value.Site, "", 1) {
					pending = append(pending, graphPending{value.Value, graphDemand{site: value.Site}})
				}
			case ir.SetAdd:
				if slot(value.Site, "", 0) {
					pending = append(pending, graphPending{value.Value, graphDemand{site: value.Site}})
				}
			}
			return true
		})
	}
	collect(program.Main, -1)
	for i, function := range program.Functions {
		collect(function.Body, i)
	}
	// where is a demand's place in the source: the write, the local's declaration, or the function's.
	where := func(at graphDemand) string {
		if at.site > 0 {
			return f.l.writeSiteWhere(at.site)
		}
		var node *ast.Node
		switch locals, functions := len(program.Locals), len(program.Functions); {
		case at.node > 0 && at.node <= locals:
			node = f.l.localNodes[at.node-1]
		case at.node > locals && at.node <= locals+functions:
			node = f.l.functionNodes[at.node-locals-1]
		case at.node > locals+functions && at.node <= locals+2*functions:
			node = f.l.functionNodes[at.node-locals-functions-1]
		}
		if node == nil {
			return program.Source
		}
		return f.l.program.Where(node)
	}
	// The stop reported is the least by place and words, so it doesn't hang on map order.
	var stopped *NotYet
	stop := func(at graphDemand, what string) {
		failure := &NotYet{Where: where(at), What: what}
		if stopped == nil || failure.Where < stopped.Where || failure.Where == stopped.Where && failure.What < stopped.What {
			stopped = failure
		}
	}
	var follow func(ir.Expression, graphDemand, bool)
	follow = func(expression ir.Expression, at graphDemand, quiet bool) {
		if expression == nil {
			return
		}
		step := graphFlowStepOf(expression, program)
		if step.stop && quiet {
			return
		}
		if step.stop {
			stop(at, fmt.Sprintf("a graph value through %T the plan doesn't follow", expression))
			return
		}
		if step.planned {
			program.GraphTypes[step.site] = true
		}
		ask := demand
		if quiet {
			ask = demandQuietly
		}
		for _, operand := range step.sources.operands {
			follow(operand, at, quiet)
		}
		for _, local := range step.sources.locals {
			ask(local + 1)
		}
		for _, function := range step.sources.results {
			ask(resultNode(function))
		}
		for _, function := range step.sources.yields {
			ask(yieldNode(function))
		}
	}
	for _, each := range pending {
		follow(each.expression, each.at, false)
	}
	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		// A graph value may reach this local through a writer collect doesn't follow, so planning its
		// sources alone would leave that value unplanned. A quiet demand never stops, as follow's doesn't.
		if what, unknown := unknownWriter[node]; unknown {
			stop(graphDemand{node: node}, what)
		}
		for _, expression := range sources[node] {
			follow(expression, graphDemand{node: node}, false)
		}
	}
	for _, argument := range quietly {
		follow(argument, graphDemand{}, true)
	}
	for len(quietQueue) > 0 {
		node := quietQueue[len(quietQueue)-1]
		quietQueue = quietQueue[:len(quietQueue)-1]
		for _, expression := range sources[node] {
			follow(expression, graphDemand{node: node}, true)
		}
	}
	if stopped != nil {
		return stopped
	}
	return nil
}

// graphType is whether graphTypes selected a type, or a member of its union or intersection. A
// Weak<Target> holds nothing, so it's never one.
func (f *cycleFinder) graphType(proven *checker.Type) bool {
	if proven == nil || f.weak(proven) {
		return false
	}
	if f.l.result.GraphTypes[int(proven.Id())] {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if f.graphType(member) {
				return true
			}
		}
	}
	return false
}

// graphReachScalar is every flag of a type no value of which holds anything: a string, a number, a
// bigint, a boolean, an enum's member, their literals, undefined, null, void and never.
const graphReachScalar = checker.TypeFlagsStringLike | checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike | checker.TypeFlagsBooleanLike | checker.TypeFlagsEnumLike | checker.TypeFlagsUndefined | checker.TypeFlagsNull | checker.TypeFlagsVoid | checker.TypeFlagsNever

// mayReachGraph is whether a value of a type can reach a value of a graph type, by what the cycle
// finder says a type holds (reachLinks): its fields and a value it may really be (shapeCandidates), an
// array's, a tuple's, a Map's or a Set's type arguments, a record's values, and an index signature's
// values besides. It's false only for a type every part of which is a scalar (graphReachScalar), or a
// union or an intersection of such. Anything it can't see to the end of reaches one, failing closed:
// a nil type, any, unknown, a type parameter or another instantiable type, object, a symbol, a type
// over any or a type parameter (template), a function value, which holds what it captured, a Promise,
// which holds its reactions, a method an interface or a literal declares, which may be a closure, and
// a captured cell or an Apple class reached through one. A class's own methods are its prototype's,
// holding nothing of the instance, so its fields alone are its parts (cycles.go's fields). judged keeps
// each answer: a type asked about, and every type a type that reaches none reached, since a part of it
// reaches none either (#x490w4a).
func (f *cycleFinder) mayReachGraph(proven *checker.Type, judged map[*checker.Type]bool) bool {
	if proven == nil {
		return true
	}
	if reaches, known := judged[proven]; known {
		return reaches
	}
	visited := map[cycleNode]bool{}
	queue := []cycleNode{{proven: proven}}
	reaches := false
	for len(queue) != 0 && !reaches {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if visited[node] {
			continue
		}
		visited[node] = true
		if node.proven == nil {
			// A captured cell or an Apple class holds what this doesn't judge.
			reaches = true
			break
		}
		if prior, known := judged[node.proven]; known {
			reaches = prior
			continue
		}
		flags := node.proven.Flags()
		switch {
		case f.weak(node.proven):
			// A Weak holds nothing.
		case f.graphType(node.proven):
			reaches = true
		case flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
			for _, member := range node.proven.Types() {
				queue = append(queue, cycleNode{proven: member})
			}
		case flags&graphReachScalar != 0:
		case flags&checker.TypeFlagsObject == 0 || f.template(node.proven) || f.isFunction(node.proven) || f.l.isLibraryType(node.proven, "Promise"):
			reaches = true
		default:
			queue = append(queue, f.reachLinks(node)...)
			for _, info := range f.l.checker.GetIndexInfosOfType(node.proven) {
				queue = append(queue, cycleNode{proven: info.ValueType()})
			}
			if f.l.isArrayType(node.proven) || checker.IsTupleType(node.proven) || f.l.isLibraryType(node.proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "MapIterator", "SetIterator") || f.l.recordElement(node.proven) != nil {
				continue
			}
			for _, property := range f.l.checker.GetPropertiesOfType(node.proven) {
				if graphReachMethod(property) {
					reaches = true
					break
				}
				// An accessor's value is what reachLinks' fields leave out; a class method holds nothing.
				if property.Flags&ast.SymbolFlagsMethod == 0 {
					queue = append(queue, cycleNode{proven: f.l.checker.GetTypeOfSymbol(property)})
				}
			}
		}
	}
	judged[proven] = reaches
	if !reaches {
		for node := range visited {
			if node.proven != nil {
				judged[node.proven] = false
			}
		}
	}
	return reaches
}

// graphReachMethod is whether a property is a method a value can hold as a closure of its own: one an
// interface, a type literal or an object literal declares. A class's is its prototype's.
func graphReachMethod(property *ast.Symbol) bool {
	if property.Flags&ast.SymbolFlagsMethod == 0 {
		return false
	}
	for _, declaration := range property.Declarations {
		if declaration.Kind != ast.KindMethodDeclaration || declaration.Parent == nil || !ast.IsClassLike(declaration.Parent) {
			return true
		}
	}
	return false
}

// argumentTypeIDs is argumentTypeID of each argument of a call, in order.
func (l *lowering) argumentTypeIDs(arguments []*ast.Node) []int {
	ids := make([]int, len(arguments))
	for index, argument := range arguments {
		ids[index] = l.argumentTypeID(argument)
	}
	return ids
}

// argumentTypeID numbers an argument's types in argumentTypes, from 1: its own, the slot it's passed into
// and any view a caller names. It's 0 for a spread, which graphFlows follows whatever it holds. The slot
// can hold more than the value's own type says: a { n: undefined } passed as a { n: Node | undefined } can
// be written a graph pointer. The number is lowering's own, never the checker's id, which the checker hands
// out lazily, so the same program numbers the same however its cycle finder asked (#x490w4a).
func (l *lowering) argumentTypeID(argument *ast.Node, views ...*checker.Type) int {
	if argument == nil || argument.Kind == ast.KindSpreadElement {
		return 0
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(argument))
	if proven == nil {
		return 0
	}
	kept := []*checker.Type{proven}
	keep := func(each *checker.Type) {
		if each != nil && !slices.Contains(kept, each) {
			kept = append(kept, each)
		}
	}
	if slot := l.slotType(argument); slot != nil {
		keep(l.concrete(slot))
	}
	for _, view := range views {
		if view != nil {
			keep(l.concrete(view))
		}
	}
	l.argumentTypes = append(l.argumentTypes, kept)
	return len(l.argumentTypes)
}
