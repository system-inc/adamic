package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Result publication currently shares values. A graph result needs an exclusive
// join proof instead, including region members disconnected by past overwrites.
// Type reachability only detects this demand; it cannot certify that ownership.
func (f *cycleFinder) parallelJoinResults(modules []*ast.SourceFile) error {
	// Lowering records every emitted parallelMap. Programs without a task
	// boundary need no additional source walk.
	if len(f.l.moveSites) == 0 {
		return nil
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil || ast.IsPartOfTypeNode(node) {
			return false
		}
		if node.Kind == ast.KindCallExpression && f.l.isPreludeFunction(node.AsCallExpression().Expression, "parallelMap") {
			result := f.l.concrete(f.l.checker.GetTypeAtLocation(node))
			pending := []cycleNode{}
			for _, element := range f.l.checker.GetTypeArguments(result) {
				pending = append(pending, cycleNode{proven: element})
			}
			seen := map[cycleNode]bool{}
			for len(pending) != 0 {
				current := pending[len(pending)-1]
				pending = pending[:len(pending)-1]
				if seen[current] {
					continue
				}
				seen[current] = true
				graph := current.cell != 0 && f.l.result.Locals[current.cell-1].GraphCell
				if current.proven != nil && !f.weak(current.proven) {
					graph = graph || f.template(current.proven) || f.l.result.GraphTypes[int(current.proven.Id())]
				}
				if graph {
					found = &Refused{
						Where: f.l.program.Where(node),
						What:  "cannot move work.result: whole graph ownership at join is not proven",
						Fix:   "return an acyclic result until the whole-graph join proof is implemented",
					}
					return false
				}
				pending = append(pending, f.graphLinks(current)...)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}
