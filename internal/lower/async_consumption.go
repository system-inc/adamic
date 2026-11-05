package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
)

// A rejecting promise needs a consumer before control can leave its scope. A
// direct synchronous return transfers that obligation to the caller. Aliases
// remain the same obligation; a fulfilled library promise needs no rejection
// observer. This deliberately does not infer an observer in another closure.
//
// A string past V8's longest throws (ir.StringLengthMayThrow), so a template or a concatenation
// outside a try is an escaping completion like any throw, and the proof keeps every such edge
// (Adamic's ruling on #c3bk0fp). A stop the proof makes only because of those edges is its own
// NotYet, named for its cause and reported at the template or concatenation: the proof is run again
// with them set aside, and when that run goes through, the first run's stop is the limit's.
func (l *lowering) checkPromiseConsumption() error {
	graphs := map[int]*flow.Function{}
	for index, function := range l.result.Functions {
		if function.Async {
			graphs[index] = flow.Build(l.result, index)
		}
	}
	stop := l.promiseConsumption(graphs, true)
	if stop == nil {
		return nil
	}
	if l.promiseConsumption(graphs, false) != nil {
		return l.notYet(stop.node, "a promise not awaited in the same function before an escaping completion")
	}
	node := stop.node
	if stop.site > 0 {
		node = l.stringSites[stop.site-1]
	}
	return l.notYet(node, stringLengthConsumption)
}

// stringLengthConsumption is the NotYet of a promise the consumption proof can't see awaited only
// because a string past V8's longest could throw first (#c3bk0fp). It lifts once the runtime reports
// an unobserved rejection as Node does, so the proof no longer needs every promise fulfilled.
const stringLengthConsumption = "an await in an async function whose completion may throw at the string-length limit before every pending promise is awaited"

// consumptionStop is where the consumption proof stopped: the node it reports at, and the first
// template or concatenation (an ir.Concat Site, 0 for none) whose string-length edge it leaned on.
type consumptionStop struct {
	node *ast.Node
	site int
}

// firstSite is the lowest positive site of a and b, or 0.
func firstSite(a, b int) int {
	if a == 0 || (b != 0 && b < a) {
		return b
	}
	return a
}

// lengthSite is the first site of a concatenation in value that can throw past V8's longest, or 0.
func lengthSite(value any) int {
	site := 0
	walk(value, func(node any) bool {
		if concat, ok := node.(ir.Concat); ok && ir.StringLengthMayThrow(concat) {
			site = firstSite(site, concat.Site)
		}
		return true
	})
	return site
}

// statementValue is the expression a statement evaluates first and whole: a declaration's or an
// assignment's value, an evaluated expression, or a written line's string, or nil.
func statementValue(statement ir.Statement) ir.Expression {
	switch statement := statement.(type) {
	case ir.Declare:
		return statement.Value
	case ir.Assign:
		return statement.Value
	case ir.Evaluate:
		return statement.Value
	case ir.WriteLine:
		return statement.Value
	}
	return nil
}

// leadingAwait is the await value evaluates before anything else: value itself, or the first part of
// a concatenation, or what a conversion to a string writes, followed down.
func leadingAwait(value ir.Expression) (ir.Await, bool) {
	for {
		switch node := value.(type) {
		case ir.Await:
			return node, true
		case ir.Concat:
			if len(node.Parts) == 0 {
				return ir.Await{}, false
			}
			value = node.Parts[0]
		case ir.NumberToString:
			value = node.Value
		case ir.BooleanToString:
			value = node.Value
		case ir.MaybeToString:
			value = node.Value
		case ir.UnionToString:
			value = node.Value
		default:
			return ir.Await{}, false
		}
	}
}

// asyncCompletion is how an async function's graph can end by a throw: safe when it can't at all,
// lengthOnly when every way out by a throw is an instruction that can throw only past V8's longest,
// and site the first template or concatenation among those.
type asyncCompletion struct {
	safe, lengthOnly bool
	site             int
}

func (l *lowering) asyncCompletion(graph *flow.Function) asyncCompletion {
	var thrown flow.BlockId
	found := false
	for _, block := range graph.Blocks {
		if _, throws := block.Terminal.(*flow.Throw); throws {
			thrown, found = block.Id, true
		}
	}
	if !found {
		return asyncCompletion{safe: true}
	}
	completion := asyncCompletion{lengthOnly: true}
	for _, block := range graph.Blocks {
		switch terminal := block.Terminal.(type) {
		case *flow.MayThrow:
			if terminal.Handler != thrown {
				continue
			}
			instruction := graph.Instructions[block.Instructions[len(block.Instructions)-1]]
			if flow.CanThrowBeyondStringLength(l.result, instruction) {
				completion.lengthOnly = false
			} else {
				completion.site = firstSite(completion.site, lengthSite(instruction.Expression))
			}
		case *flow.Goto:
			completion.lengthOnly = completion.lengthOnly && terminal.Block != thrown
		case *flow.Choose:
			for _, next := range terminal.Blocks {
				completion.lengthOnly = completion.lengthOnly && next != thrown
			}
		case *flow.Suspend:
			completion.lengthOnly = completion.lengthOnly && terminal.Rejected != thrown && terminal.Fulfilled != thrown
		case *flow.If:
			completion.lengthOnly = completion.lengthOnly && terminal.Consequent != thrown && terminal.Alternate != thrown
		}
	}
	return completion
}

// promiseConsumption is the proof, keeping string-length edges when lengths is true and setting
// them aside when it's false, and where it stops, or nil.
func (l *lowering) promiseConsumption(graphs map[int]*flow.Function, lengths bool) *consumptionStop {
	initializers := map[int]ir.Expression{}
	mutable := map[int]bool{}
	collect := func(statements []ir.Statement) {
		walk(statements, func(node any) bool {
			switch value := node.(type) {
			case ir.Declare:
				if value.Value != nil && value.Value.Type() == ir.Promise {
					initializers[value.Local] = value.Value
				}
			case ir.Assign:
				if value.Value.Type() == ir.Promise {
					mutable[value.Local] = true
				}
			}
			return true
		})
	}
	collect(l.result.Main)
	for _, function := range l.result.Functions {
		collect(function.Body)
	}
	var origin func(int, map[int]bool) int
	origin = func(local int, seen map[int]bool) int {
		if seen[local] || mutable[local] {
			return local
		}
		seen[local] = true
		if read, ok := initializers[local].(ir.Read); ok {
			return origin(read.Local, seen)
		}
		return local
	}
	// A graph with no escaping Throw has handled every rejection and synchronous
	// exception. Unknown targets and ordinary Promise forwarders remain unproved.
	// Set aside, a graph whose only escaping throws are strings past V8's longest is safe.
	safeFunctions := map[int]bool{}
	lengthSites := map[int]int{}
	for index, graph := range graphs {
		completion := l.asyncCompletion(graph)
		safeFunctions[index] = completion.safe || !lengths && completion.lengthOnly
		if !completion.safe && completion.lengthOnly {
			lengthSites[index] = completion.site
		}
	}
	// site is the first template or concatenation whose string-length edge the decision being made
	// leaned on, in a function it awaits, and near the first in the statement where the promise could
	// be bypassed, the one the stop is reported at when there is one.
	site, near := 0, 0
	var fulfilled func(ir.Expression, map[int]bool) bool
	fulfilled = func(value ir.Expression, seen map[int]bool) bool {
		switch value := value.(type) {
		case ir.PromiseValue:
			return !value.Reject
		case ir.Read:
			if seen[value.Local] || mutable[value.Local] {
				return false
			}
			seen[value.Local] = true
			return fulfilled(initializers[value.Local], seen)
		case ir.Call:
			targets := l.result.CallTargets(value)
			if len(targets) == 0 {
				return false
			}
			for _, target := range targets {
				if !safeFunctions[target] {
					site = firstSite(site, lengthSites[target])
					return false
				}
			}
			return true
		}
		return false
	}
	var consumedAfter func([]ir.Statement, int) bool
	consumedAfter = func(statements []ir.Statement, local int) bool {
		for _, statement := range statements {
			var expression ir.Expression
			switch value := statement.(type) {
			case ir.Declare:
				expression = value.Value
			case ir.Evaluate:
				expression = value.Value
			case ir.Return:
				read, ok := value.Value.(ir.Read)
				return ok && read.Of == ir.Promise && origin(read.Local, map[int]bool{}) == origin(local, map[int]bool{})
			case ir.Block:
				if consumedAfter(value.Body, local) {
					return true
				}
			case ir.If:
				if consumedAfter(value.Then, local) && consumedAfter(value.Else, local) {
					return true
				}
			case ir.Try:
				if consumedAfter(value.Body, local) {
					return true
				}
			}
			if awaited, ok := expression.(ir.Await); ok {
				if read, ok := awaited.Value.(ir.Read); ok && origin(read.Local, map[int]bool{}) == origin(local, map[int]bool{}) {
					return true
				}
			}
			// An await of it that the statement evaluates before anything else, a template's first
			// value among them, observes its rejection before anything in the statement can throw.
			if awaited, ok := leadingAwait(statementValue(statement)); ok {
				if read, ok := awaited.Value.(ir.Read); ok && origin(read.Local, map[int]bool{}) == origin(local, map[int]bool{}) {
					return true
				}
			}
			// A prior rejection or abrupt completion could bypass the later consumer.
			barrier := false
			walk(statement, func(node any) bool {
				switch value := node.(type) {
				case ir.Return, ir.Throw, ir.Panic, ir.Break, ir.Continue:
					barrier = true
				case ir.Await:
					barrier = barrier || !fulfilled(value.Value, map[int]bool{})
				}
				return !barrier
			})
			at := statement
			instruction := &flow.Instruction{At: &at, Expression: expression}
			if barrier {
				near = firstSite(near, lengthSite(statement))
				return false
			}
			if flow.CanThrow(l.result, instruction) {
				if !flow.CanThrowBeyondStringLength(l.result, instruction) {
					near = firstSite(near, lengthSite(expression))
					if !lengths {
						continue
					}
				}
				return false
			}
		}
		return false
	}
	var checkScope func([]ir.Statement, *ast.Node) *consumptionStop
	checkScope = func(statements []ir.Statement, where *ast.Node) *consumptionStop {
		for index, statement := range statements {
			if declared, ok := statement.(ir.Declare); ok && declared.Value != nil && declared.Value.Type() == ir.Promise {
				node := where
				if original := l.localNodes[declared.Local]; original != nil {
					node = original
				}
				site, near = 0, 0
				if !fulfilled(declared.Value, map[int]bool{}) && (mutable[declared.Local] || l.result.Locals[declared.Local].Captured || !consumedAfter(statements[index+1:], declared.Local)) {
					if near > 0 {
						site = near
					}
					return &consumptionStop{node: node, site: site}
				}
			}
			var failure *consumptionStop
			var inspect func(reflect.Value, bool)
			inspect = func(value reflect.Value, consumed bool) {
				if !value.IsValid() || failure != nil {
					return
				}
				if value.CanInterface() {
					if body, ok := value.Interface().([]ir.Statement); ok {
						failure = checkScope(body, where)
						return
					}
					if expression, ok := value.Interface().(ir.Expression); ok {
						site, near = 0, 0
						if expression.Type() == ir.Promise && !consumed && !fulfilled(expression, map[int]bool{}) {
							if _, read := expression.(ir.Read); !read {
								failure = &consumptionStop{node: where, site: site}
								return
							}
						}
						if awaited, ok := expression.(ir.Await); ok {
							inspect(reflect.ValueOf(awaited.Value), true)
							return
						}
					}
					switch node := value.Interface().(type) {
					case ir.Declare:
						inspect(reflect.ValueOf(node.Value), true)
						return
					case ir.Return:
						inspect(reflect.ValueOf(node.Value), true)
						return
					}
				}
				switch value.Kind() {
				case reflect.Interface, reflect.Pointer:
					if !value.IsNil() {
						inspect(value.Elem(), consumed)
					}
				case reflect.Struct:
					for field := 0; field < value.NumField(); field++ {
						inspect(value.Field(field), false)
					}
				case reflect.Slice, reflect.Array:
					for item := 0; item < value.Len(); item++ {
						inspect(value.Index(item), false)
					}
				}
			}
			inspect(reflect.ValueOf(statement), false)
			if failure != nil {
				return failure
			}
		}
		return nil
	}
	if stop := checkScope(l.result.Main, l.entry.AsNode()); stop != nil {
		return stop
	}
	for _, function := range l.result.Functions {
		where := l.entry.AsNode()
		if function.AsyncSite > 0 {
			where = l.asyncSites[function.AsyncSite-1]
		}
		if stop := checkScope(function.Body, where); stop != nil {
			return stop
		}
	}
	return nil
}
