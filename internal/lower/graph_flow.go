package lower

import (
	"fmt"
	"reflect"

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

// graphFlows works backwards from graph views to the allocations that can supply
// them. Local assignments and function results are joined across all paths; this
// is a may-flow proof, not an execution or a choice of one conditional branch.
// Negative IDs name allocation sites in this IR, independently of fresh checker
// identities. They select the existing allocation emission without runtime work.
// follow classifies every IR expression by graphFlowRules (graph_flow_classes.go), and a graph
// value through one the plan doesn't follow is NotYet at the site that demanded it, never a leak.
// An argument passed to a function value whose targets are Unknown is followed too, quietly: it
// plans what it reaches and never stops (#41vb71c).
func (f *cycleFinder) graphFlows() error {
	program := f.l.result
	next := 0
	program.Main = graphAllocationSites(reflect.ValueOf(program.Main), &next).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = graphAllocationSites(reflect.ValueOf(program.Functions[i].Body), &next).Interface().([]ir.Statement)
	}

	var graphType func(*checker.Type) bool
	graphType = func(proven *checker.Type) bool {
		if proven == nil || f.weak(proven) {
			return false
		}
		if program.GraphTypes[int(proven.Id())] {
			return true
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				if graphType(member) {
					return true
				}
			}
		}
		return false
	}
	// Positive nodes are locals, then function results, then generator functions' yields.
	resultNode := func(function int) int { return len(program.Locals) + function + 1 }
	yieldNode := func(function int) int { return len(program.Locals) + len(program.Functions) + function + 1 }
	sources := map[int][]ir.Expression{}
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
		if graphType(proven) {
			demand(local + 1)
		}
	}
	for local, types := range f.l.localAlso {
		for _, proven := range types {
			if graphType(proven) {
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
	// reaches has it as a source, so it's demanded at the call, whatever its type (#41vb71c).
	quietly := []ir.Expression{}
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
		return graphType(slotType(site, name, argument))
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
		if graphType(proven) {
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
		case ir.Conditional:
			members(value.WhenTrue, proven, at)
			members(value.WhenNot, proven, at)
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
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
					// A spread argument is its array, the operand followed here like any other.
					quietly = append(quietly, value.Arguments...)
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
					quietly = append(quietly, value.Initial, value.Array)
				}
			// A slot write is seeded through members, as a declaration's value is, so a literal written
			// there seeds its own graph-typed parts too: new Map's entries among them (#dnb0yhj, s1).
			case ir.SetProperty:
				members(value.Value, slotType(value.Site, value.Name, 0), graphDemand{site: value.Site})
			case ir.SetIndex:
				members(value.Value, slotType(value.Site, "", 0), graphDemand{site: value.Site})
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
			// by a write this seeds where it was made.
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
			failure := &NotYet{Where: where(at), What: fmt.Sprintf("a graph value through %T the plan doesn't follow", expression)}
			if stopped == nil || failure.Where < stopped.Where || failure.Where == stopped.Where && failure.What < stopped.What {
				stopped = failure
			}
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
