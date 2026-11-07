package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Each function body is inspected once. Tarjan's active stack identifies cycles;
// only a complete component receives a cached transitive reach set. Sets contain
// all namespace and regular enum declarations, independent of the current module evaluation point.
type namespaceCallGraph struct {
	lowering  *lowering
	functions map[*ast.Node]*namespaceCallNode
	stack     []*namespaceCallNode
	next      int
	walks     int
}

type namespaceCallNode struct {
	index, low int
	active     bool
	edges      []*namespaceCallNode
	reads      map[*ast.Node]*ast.Node
	reaches    map[*ast.Node]*ast.Node
}

func (graph *namespaceCallGraph) reach(function *ast.Node) map[*ast.Node]*ast.Node {
	return graph.discover(function).reaches
}

func (graph *namespaceCallGraph) discover(function *ast.Node) *namespaceCallNode {
	if cached := graph.functions[function]; cached != nil {
		return cached
	}
	graph.walks++
	entry := &namespaceCallNode{index: graph.next, low: graph.next, active: true, reads: map[*ast.Node]*ast.Node{}}
	graph.next++
	graph.functions[function] = entry
	graph.stack = append(graph.stack, entry)
	connect := func(target *ast.Node) {
		previous := graph.functions[target]
		edge := graph.discover(target)
		entry.edges = append(entry.edges, edge)
		if edge.active {
			if previous == nil {
				entry.low = min(entry.low, edge.low)
			} else {
				entry.low = min(entry.low, edge.index)
			}
		}
	}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || ast.IsTypeNode(node) || node != function && ast.IsFunctionLike(node) {
			return
		}
		if branch := graph.lowering.literalCallableBranch(node); branch != nil {
			walk(branch)
			return
		}
		if ast.IsClassLike(node) {
			if node == function {
				for _, clause := range nodesOf(node.ClassLikeData().HeritageClauses) {
					if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
						for _, element := range clause.AsHeritageClause().Types.Nodes {
							if base := graph.lowering.namespaceCallable(
								element.AsExpressionWithTypeArguments().Expression,
							); base != nil {
								connect(base)
							}
						}
					}
				}
				// Construction executes instance initializers, parameter defaults and
				// the constructor body. Static members ran at the class declaration.
				for _, member := range node.ClassLikeData().Members.Nodes {
					if member.Kind == ast.KindPropertyDeclaration && !ast.HasStaticModifier(member) {
						walk(member.Initializer())
					}
					if member.Kind == ast.KindConstructor {
						for _, parameter := range member.Parameters() {
							walk(parameter.Initializer())
						}
						walk(member.Body())
					}
				}
				return
			}
			for _, clause := range nodesOf(node.ClassLikeData().HeritageClauses) {
				if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
					for _, element := range clause.AsHeritageClause().Types.Nodes {
						walk(element.AsExpressionWithTypeArguments().Expression)
					}
				}
			}
			for _, member := range node.ClassLikeData().Members.Nodes {
				if ast.HasStaticModifier(member) || member.Kind == ast.KindClassStaticBlockDeclaration {
					walk(member)
				}
			}
			return
		}
		if graph.lowering.namespaceValueNode(node) {
			if declaration := graph.lowering.namespaceRuntimeEnum(node); declaration != nil {
				entry.reads[declaration] = node
			}
			if declaration := graph.lowering.namespaceDeclaration(
				node,
			); declaration != nil && namespaceRuntime(declaration) {
				entry.reads[declaration] = node
			}
		}
		if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {
			callees := append([]*ast.Node{node.Expression()}, node.Arguments()...)
			for _, callee := range callees {
				if target := graph.lowering.namespaceCallable(callee); target != nil {
					connect(target)
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(function)
	if entry.low == entry.index {
		component := []*namespaceCallNode{}
		for {
			last := graph.stack[len(graph.stack)-1]
			graph.stack = graph.stack[:len(graph.stack)-1]
			last.active = false
			component = append(component, last)
			if last == entry {
				break
			}
		}
		reaches := map[*ast.Node]*ast.Node{}
		for _, member := range component {
			for declaration, read := range member.reads {
				reaches[declaration] = read
			}
			for _, edge := range member.edges {
				for declaration, read := range edge.reaches {
					reaches[declaration] = read
				}
			}
		}
		for _, member := range component {
			member.reaches = reaches
		}
	}
	return entry
}

// Both module and namespace enums participate in the same initialization graph.
func (lowering *lowering) namespaceRuntimeEnum(node *ast.Node) *ast.Node {
	declaration := lowering.enumObject(node)
	if declaration == nil || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) {
		return nil
	}
	return declaration
}
