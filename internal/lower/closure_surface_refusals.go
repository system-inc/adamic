package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func boundSurfaceRefusal(where, operation string) error {
	if operation == "bind" {
		return &Refused{Where: where, What: "binding a bound function again", Fix: "bind once"}
	}
	return &Refused{Where: where, What: operation + " on a bound function", Fix: "call the bound function directly"}
}

// Catch inferred tuple-rest bind results before ordinary signature lowering.
// Typed aliases and helper parameters are checked against value flow below.
func (l *lowering) boundSurfaceExpression(node *ast.Node, active map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindAsExpression:
		return l.boundSurfaceExpression(node.AsAsExpression().Expression, active)
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			access := callee.AsPropertyAccessExpression()
			of, known := l.representation(l.concrete(l.checker.GetTypeAtLocation(access.Expression)))
			return access.Name().Text() == "bind" && known && of == ir.Closure
		}
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || active[symbol] {
			return false
		}
		active[symbol] = true
		defer delete(active, symbol)
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Initializer() != nil && l.boundSurfaceExpression(declaration.Initializer(), active) {
				return true
			}
		}
	}
	return false
}

// The factory result can cross typed aliases and helper arguments. Refuse at
// the source operation rather than relying on a backend dispatch failure.
func (graph *allocationFlowGraph) checkBoundSurfaceCalls() error {
	program := graph.program
	// ClosureTargets deliberately leaves returned/local function values opaque.
	// Use the existing closed-world type targets only to add conservative edges
	// for this refusal, without changing the shared allocation-flow graph.
	values := map[int][]ir.Expression{}
	for node, sources := range graph.sources {
		values[node] = append([]ir.Expression{}, sources...)
	}
	connect := func(node any) bool {
		call, ok := node.(ir.CallClosure)
		if !ok || call.CheckBound {
			return true
		}
		targets := program.ClosureTargets(call)
		if !targets.Unknown {
			return true
		}
		candidates := program.FunctionTypeTargets[call.FunctionType]
		for _, target := range candidates {
			function := program.Functions[target]
			parameters := function.Parameters
			if function.Receiver && len(parameters) != 0 {
				parameters = parameters[1:]
			}
			for i, local := range parameters {
				if i < len(call.Arguments) {
					values[local+1] = append(values[local+1], call.Arguments[i])
				}
			}
		}
		return true
	}
	walk(program.Main, connect)
	for _, function := range program.Functions {
		walk(function.Body, connect)
	}
	var refused error
	inspect := func(node any) bool {
		call, ok := node.(ir.CallClosure)
		if !ok || call.SurfaceOperation == "" || refused != nil {
			return refused == nil
		}
		seen := map[int]bool{}
		fields := map[struct {
			site int
			name string
		}]bool{}
		var bound func(ir.Expression) bool
		var sources func(int) bool
		sources = func(node int) bool {
			if seen[node] {
				return false
			}
			seen[node] = true
			for _, value := range values[node] {
				if bound(value) {
					return true
				}
			}
			return false
		}
		bound = func(value ir.Expression) bool {
			switch value := value.(type) {
			case ir.MakeClosure:
				return program.Functions[value.Function].Bound != nil
			case ir.Read:
				return sources(value.Local + 1)
			case ir.Call:
				for _, function := range program.CallTargets(value) {
					if sources(graph.resultNode(function)) {
						return true
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				candidates := targets.Functions
				if targets.Unknown {
					candidates = program.FunctionTypeTargets[value.FunctionType]
				}
				for _, function := range candidates {
					if sources(graph.resultNode(function)) {
						return true
					}
				}
			case ir.Conditional:
				return bound(value.WhenTrue) || bound(value.WhenNot)
			case ir.Defined:
				return bound(value.Value)
			case ir.Narrow:
				return bound(value.Value)
			case ir.Box:
				return bound(value.Value)
			case ir.Property:
				index := graph.projectionIndex()
				for _, site := range graph.ReachingAllocations(value.Object).Sites {
					key := struct {
						site int
						name string
					}{site, value.Name}
					if fields[key] {
						continue
					}
					fields[key] = true
					for _, field := range index.records[site].Fields {
						if field.Name == value.Name && bound(field.Value) {
							return true
						}
					}
					for _, stored := range index.stores[site][value.Name] {
						if bound(stored) {
							return true
						}
					}
				}
			}
			return false
		}
		if bound(program.ClosureTargets(call).Value) {
			refused = boundSurfaceRefusal(call.CallWhere, call.SurfaceOperation)
		}
		return refused == nil
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return refused
}
