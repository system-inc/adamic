package checker

// SymbolGraph exposes generic checker identities and declaration/module links.
// It does not classify callees, hooks, effects or compiler IR. Edges are numeric.
import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"sort"
	"strings"
	"unicode/utf16"
)

type graphExpr struct {
	kind                   string
	symbol, inner          int
	propertyKind, property string
}
type graphModule struct {
	name   string
	symbol int
}
type graphDecl struct {
	kind, name, property string
	module, initializer  int
	constant, blocked    bool
}
type graphExport struct {
	name   string
	symbol int
}
type graphSymbol struct {
	immediate, resolved int
	declarations        []graphDecl
	exports             []graphExport
	stars               []int
}
type symbolGraphBuilder struct {
	c           *checker.Checker
	home        *ast.SourceFile
	ids         map[*ast.Symbol]int
	symbols     []*ast.Symbol
	modules     []graphModule
	moduleIDs   map[*ast.Node]int
	expressions []graphExpr
}

func (g *symbolGraphBuilder) symbol(s *ast.Symbol) int {
	if s == nil {
		return 0
	}
	if id := g.ids[s]; id != 0 {
		return id
	}
	id := len(g.symbols) + 1
	g.ids[s] = id
	g.symbols = append(g.symbols, s)
	return id
}
func graphName(n *ast.Node) string {
	if n != nil && (n.Kind == ast.KindIdentifier || n.Kind == ast.KindPrivateIdentifier || ast.IsStringLiteralLike(n) || n.Kind == ast.KindNumericLiteral) {
		return n.Text()
	}
	return ""
}
func (g *symbolGraphBuilder) module(n *ast.Node) int {
	if n == nil || !ast.IsStringLiteralLike(n) {
		return 0
	}
	if id := g.moduleIDs[n]; id != 0 {
		return id
	}
	id := len(g.modules) + 1
	g.moduleIDs[n] = id
	s := g.c.GetSymbolAtLocation(n)
	if s != nil && s.Flags&ast.SymbolFlagsAlias != 0 {
		s = checker.Checker_resolveAlias(g.c, s)
	}
	if s != nil && s.Flags&ast.SymbolFlagsModule == 0 {
		s = nil
	}
	g.modules = append(g.modules, graphModule{n.Text(), g.symbol(s)})
	return id
}
func (g *symbolGraphBuilder) expression(n *ast.Node) int {
	if n == nil {
		return 0
	}
	value := graphExpr{kind: strings.TrimPrefix(n.Kind.String(), "Kind")}
	switch n.Kind {
	case ast.KindIdentifier:
		value.symbol = g.symbol(g.c.GetSymbolAtLocation(n))
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		value.inner = g.expression(n.Expression())
	case ast.KindPropertyAccessExpression:
		value.inner = g.expression(n.Expression())
		value.propertyKind = "Identifier"
		value.property = graphName(n.Name())
	case ast.KindElementAccessExpression:
		value.inner = g.expression(n.Expression())
		property := n.AsElementAccessExpression().ArgumentExpression
		if property != nil {
			value.propertyKind = strings.TrimPrefix(property.Kind.String(), "Kind")
			value.property = graphName(property)
		}
	}
	g.expressions = append(g.expressions, value)
	return len(g.expressions)
}
func (g *symbolGraphBuilder) declaration(n *ast.Node) graphDecl {
	d := graphDecl{kind: strings.TrimPrefix(n.Kind.String(), "Kind"), name: graphName(n.Name()), property: graphName(n.PropertyName())}
	if ast.GetSourceFileOfNode(n) != g.home {
		for p := n.Parent; p != nil; p = p.Parent {
			if ast.IsFunctionLikeDeclaration(p) || ast.IsClassStaticBlockDeclaration(p) {
				if body := p.Body(); body != nil && n.Pos() >= body.Pos() && n.End() <= body.End() {
					d.blocked = true
					return d
				}
			}
		}
	}
	switch n.Kind {
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindExportSpecifier:
		for p := n.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindImportDeclaration {
				d.module = g.module(p.AsImportDeclaration().ModuleSpecifier)
				break
			}
			if p.Kind == ast.KindExportDeclaration {
				d.module = g.module(p.AsExportDeclaration().ModuleSpecifier)
				break
			}
		}
	case ast.KindVariableDeclaration:
		d.constant = n.Parent != nil && n.Parent.Kind == ast.KindVariableDeclarationList && n.Parent.Flags&ast.NodeFlagsConst != 0
		if d.constant {
			d.initializer = g.expression(n.AsVariableDeclaration().Initializer)
		}
	}
	return d
}
func SymbolGraph(c *checker.Checker, home *ast.SourceFile, ids map[*ast.Symbol]int) string {
	g := symbolGraphBuilder{c: c, home: home, ids: ids, moduleIDs: map[*ast.Node]int{}}
	g.symbols = make([]*ast.Symbol, len(ids))
	for s, id := range ids {
		g.symbols[id-1] = s
	}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier {
			var s *ast.Symbol
			if p := n.Parent; p != nil && p.Kind == ast.KindShorthandPropertyAssignment && p.Name() == n {
				s = c.GetShorthandAssignmentValueSymbol(p)
			}
			if s == nil {
				s = c.GetSymbolAtLocation(n)
			}
			if s == nil {
				s = n.Symbol()
			}
			g.symbol(s)
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(home.AsNode())
	var nodes []graphSymbol
	for index := 0; index < len(g.symbols); index++ {
		s := g.symbols[index]
		node := graphSymbol{resolved: index + 1}
		if s.Flags&ast.SymbolFlagsAlias != 0 {
			node.immediate = g.symbol(checker.Checker_getImmediateAliasedSymbol(c, s))
			node.resolved = g.symbol(checker.Checker_resolveAlias(c, s))
		}
		for _, d := range s.Declarations {
			node.declarations = append(node.declarations, g.declaration(d))
			if d.Kind == ast.KindSourceFile {
				for _, statement := range d.AsSourceFile().Statements.Nodes {
					if statement.Kind == ast.KindExportDeclaration {
						x := statement.AsExportDeclaration()
						if x.ExportClause == nil && x.ModuleSpecifier != nil {
							node.stars = append(node.stars, g.module(x.ModuleSpecifier))
						}
					}
				}
			}
		}
		if s.Flags&ast.SymbolFlagsModule != 0 {
			exports := append([]*ast.Symbol(nil), c.GetExportsOfModule(s)...)
			sort.Slice(exports, func(i, j int) bool { return exports[i].Name < exports[j].Name })
			for _, e := range exports {
				node.exports = append(node.exports, graphExport{e.Name, g.symbol(e)})
			}
		}
		nodes = append(nodes, node)
	}
	var out strings.Builder
	field := func(s string) { fmt.Fprintf(&out, "%d\n%s", len(utf16.Encode([]rune(s))), s) }
	number := func(n int) { field(fmt.Sprint(n)) }
	yes := func(b bool) {
		if b {
			number(1)
		} else {
			number(0)
		}
	}
	number(1)
	field("symbol-graph")
	number(len(nodes))
	for _, n := range nodes {
		number(n.immediate)
		number(n.resolved)
		number(len(n.declarations))
		for _, d := range n.declarations {
			field(d.kind)
			field(d.name)
			field(d.property)
			number(d.module)
			yes(d.constant)
			number(d.initializer)
			yes(d.blocked)
		}
		number(len(n.exports))
		for _, e := range n.exports {
			field(e.name)
			number(e.symbol)
		}
		number(len(n.stars))
		for _, m := range n.stars {
			number(m)
		}
	}
	number(len(g.modules))
	for _, m := range g.modules {
		field(m.name)
		number(m.symbol)
	}
	number(len(g.expressions))
	for _, e := range g.expressions {
		field(e.kind)
		number(e.symbol)
		number(e.inner)
		field(e.propertyKind)
		field(e.property)
	}
	return out.String()
}
