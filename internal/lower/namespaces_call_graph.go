package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Each function body is inspected once. Tarjan's active stack identifies cycles;
// only a complete component receives a cached transitive reach set. Sets contain
// all namespace declarations, independent of the current module evaluation point.
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

func (g *namespaceCallGraph) reach(function *ast.Node) map[*ast.Node]*ast.Node {
	return g.discover(function).reaches
}

func (g *namespaceCallGraph) discover(function *ast.Node) *namespaceCallNode {
	if cached := g.functions[function]; cached != nil {
		return cached
	}
	g.walks++
	entry := &namespaceCallNode{index: g.next, low: g.next, active: true, reads: map[*ast.Node]*ast.Node{}}
	g.next++
	g.functions[function] = entry
	g.stack = append(g.stack, entry)
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || ast.IsTypeNode(node) || node != function && ast.IsFunctionLike(node) {
			return
		}
		if ast.IsClassLike(node) {
			for _, clause := range nodesOf(node.AsClassDeclaration().HeritageClauses) {
				if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
					for _, element := range clause.AsHeritageClause().Types.Nodes {
						walk(element.AsExpressionWithTypeArguments().Expression)
					}
				}
			}
			for _, member := range node.AsClassDeclaration().Members.Nodes {
				if ast.HasStaticModifier(member) || member.Kind == ast.KindClassStaticBlockDeclaration {
					walk(member)
				}
			}
			return
		}
		if g.lowering.namespaceValueNode(node) {
			if declaration := g.lowering.namespaceRuntimeEnum(node); declaration != nil {
				entry.reads[declaration] = node
			}
			if declaration := g.lowering.namespaceDeclaration(node); declaration != nil && namespaceRuntime(declaration) {
				entry.reads[declaration] = node
			}
		}
		if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {
			callees := append([]*ast.Node{node.Expression()}, node.Arguments()...)
			for _, callee := range callees {
				if target := g.lowering.namespaceCallable(callee); target != nil {
					previous := g.functions[target]
					edge := g.discover(target)
					entry.edges = append(entry.edges, edge)
					if edge.active {
						if previous == nil {
							entry.low = min(entry.low, edge.low)
						} else {
							entry.low = min(entry.low, edge.index)
						}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(function)
	if entry.low == entry.index {
		component := []*namespaceCallNode{}
		for {
			last := g.stack[len(g.stack)-1]
			g.stack = g.stack[:len(g.stack)-1]
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

// Namespace-scoped regular enums retain their own initialization boundary even
// when integration replaces the module-level enum analysis with ready checks.
func (l *lowering) namespaceRuntimeEnum(node *ast.Node) *ast.Node {
	declaration := l.enumObject(node)
	if declaration == nil || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) || declaration.Parent == nil || declaration.Parent.Kind != ast.KindModuleBlock {
		return nil
	}
	return declaration
}
