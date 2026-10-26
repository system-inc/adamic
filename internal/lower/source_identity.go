package lower

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// moduleStatements is the sole orchestration hook: retain the boundary without
// changing the statements or their execution order.
func (l *lowering) moduleStatements(module *ast.SourceFile) ([]ir.Statement, error) {
	body, err := l.statements(module.Statements.Nodes)
	if err == nil {
		l.result.MainModules = append(l.result.MainModules, ir.MainModule{Module: l.modulePath(module), Statements: len(body)})
	}
	return body, err
}

func (l *lowering) modulePath(module *ast.SourceFile) string {
	if load.IsPrelude(module) {
		return "adamic-prelude:///adamic.d.ts"
	}
	filename := l.program.FileName(module)
	if strings.HasPrefix(filename, "bundled://") {
		return filename
	}
	if l.sourceModules == nil {
		l.indexSourceModules()
	}
	if l.sourceModules[module] == "" {
		panic("lower: source module has no import-declaration route: " + filename)
	}
	return l.sourceModules[module]
}

// Module identities follow source import declarations, not file placement. The
// shortest route, with lexical tie-breaking, is independent of traversal order.
// Binding clauses exclude the module specifier: moving a source cluster may
// rewrite relative specifiers to absolute paths without changing its bindings.
// Each outgoing label is unique within its source module, so distinct modules
// never acquire the same route, even with identical basenames and contents.
func (l *lowering) indexSourceModules() {
	entry := l.program.Files()[0]
	l.sourceModules = map[*ast.SourceFile]string{entry: filepath.Base(l.program.FileName(entry))}
	type pending struct {
		module *ast.SourceFile
		route  []string
	}
	type edge struct {
		module *ast.SourceFile
		label  string
	}
	queue := []pending{{module: entry}}
	for next := 0; next < len(queue); next++ {
		current := queue[next]
		edges := []edge{}
		occurrences := map[string]int{}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			var specifier *ast.Node
			label := ""
			switch node.Kind {
			case ast.KindImportDeclaration:
				specifier = node.ModuleSpecifier()
				label = "import"
				if clause := node.AsImportDeclaration().ImportClause; clause != nil {
					label += fmt.Sprintf(":%t:%s", clause.IsTypeOnly(), strings.TrimSpace(current.module.Text()[clause.Pos():clause.End()]))
				}
			case ast.KindExportDeclaration:
				specifier = node.ModuleSpecifier()
				exported := node.AsExportDeclaration()
				label = fmt.Sprintf("export:%t", exported.IsTypeOnly)
				if clause := exported.ExportClause; clause != nil {
					label += ":" + strings.TrimSpace(current.module.Text()[clause.Pos():clause.End()])
				}
			case ast.KindImportType:
				argument := node.AsImportTypeNode().Argument
				if argument.Kind == ast.KindLiteralType {
					literal := argument.AsLiteralTypeNode().Literal
					if literal.Kind == ast.KindStringLiteral {
						specifier = literal
					}
				}
				// Import types have no binding clause. Their containing named
				// declaration owns an anonymous sequence, like sourceIdentity.
				parents := []string{}
				for parent := node.Parent; parent != nil && parent.Kind != ast.KindSourceFile; parent = parent.Parent {
					if name := parent.Name(); name != nil && (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) {
						parents = append(parents, fmt.Sprintf("%d:%s", parent.Kind, name.Text()))
					}
				}
				label = fmt.Sprintf("import-type:%q", parents)
			}
			if specifier != nil {
				if symbol := l.checker.GetSymbolAtLocation(specifier); symbol != nil {
					for _, declaration := range symbol.Declarations {
						if declaration.Kind != ast.KindSourceFile {
							continue
						}
						target := declaration.AsSourceFile()
						if load.IsPrelude(target) || load.IsLibrary(target) || strings.HasPrefix(l.program.FileName(target), "bundled://") {
							break
						}
						// Only indistinguishable anonymous declarations (side
						// effects, export-stars, repeated import types) use a
						// source-local occurrence, never a lowering ordinal.
						occurrence := occurrences[label]
						occurrences[label]++
						edges = append(edges, edge{module: target, label: fmt.Sprintf("%q/%d", label, occurrence)})
						break
					}
				}
			}
			return node.ForEachChild(visit)
		}
		current.module.AsNode().ForEachChild(visit)
		sort.Slice(edges, func(i, j int) bool { return edges[i].label < edges[j].label })
		for _, imported := range edges {
			if l.sourceModules[imported.module] != "" {
				continue
			}
			route := append(append([]string{}, current.route...), imported.label)
			l.sourceModules[imported.module] = fmt.Sprintf("imported:///%q/%s", route, filepath.Base(l.program.FileName(imported.module)))
			queue = append(queue, pending{module: imported.module, route: route})
		}
	}
}

// sourceIdentity indexes a source tree once, before lowering order can influence
// anonymous numbering. A nested declaration owns a separate anonymous sequence.
func (l *lowering) sourceIdentity(node *ast.Node) ir.SourceIdentity {
	if l.sourceIdentities == nil {
		l.sourceIdentities = map[*ast.Node]ir.SourceIdentity{}
	}
	if identity, found := l.sourceIdentities[node]; found {
		return identity
	}
	module := ast.GetSourceFileOfNode(node)
	duplicates := map[string]int{}
	var walk func(*ast.Node, []string, *int)
	walk = func(node *ast.Node, parents []string, anonymous *int) {
		path := parents
		own := (ast.IsFunctionLike(node) && (node.Body() != nil || node.Kind == ast.KindMethodDeclaration)) || node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindInterfaceDeclaration || node.Kind == ast.KindTypeAliasDeclaration || node.Kind == ast.KindVariableDeclaration || node.Kind == ast.KindPropertyDeclaration || node.Kind == ast.KindPropertyAssignment
		if own {
			name := ""
			if node.Kind == ast.KindConstructor {
				name = "constructor"
			} else if node.Name() != nil {
				named := node.Name()
				if named.Kind == ast.KindIdentifier || named.Kind == ast.KindPrivateIdentifier || named.Kind == ast.KindStringLiteral || named.Kind == ast.KindNumericLiteral {
					name = named.Text()
				} else {
					name = strings.TrimSpace(module.Text()[named.Pos():named.End()])
				}
				if name == "" {
					name = "computed_method"
				}
			}
			if name == "" {
				*anonymous++
				name = fmt.Sprintf("anonymous_%d", *anonymous)
			}
			if node.Kind == ast.KindGetAccessor {
				name = "get_" + name
			}
			if node.Kind == ast.KindSetAccessor {
				name = "set_" + name
			}
			path = append(append([]string{}, parents...), name)
			if node.Name() != nil {
				key := fmt.Sprintf("%q", path)
				duplicates[key]++
				if duplicates[key] > 1 {
					path = append(path, fmt.Sprintf("duplicate_%d", duplicates[key]))
				}
			}
		}
		l.sourceIdentities[node] = ir.SourceIdentity{Module: l.modulePath(module), Declaration: path}
		counter := anonymous
		if own && node.Kind != ast.KindPropertyAssignment {
			counter = new(int)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child, path, counter); return false })
	}
	walk(module.AsNode(), nil, new(int))
	return l.sourceIdentities[node]
}

// functionSource includes concrete type spellings, never checker IDs or lowering IDs.
func (l *lowering) functionSource(node *ast.Node) ir.SourceIdentity {
	identity := l.sourceIdentity(node)
	parts := []string{}
	if l.function != nil && l.function.Source.Specialization != "" {
		parent := l.function.Source
		nested := parent.Module == identity.Module && len(parent.Declaration) < len(identity.Declaration)
		for i := 0; nested && i < len(parent.Declaration); i++ {
			nested = parent.Declaration[i] == identity.Declaration[i]
		}
		if nested {
			parts = append(parts, parent.Specialization)
		}
	}
	if l.classType != nil && l.declaredInCurrentClass(node) {
		parts = append(parts, l.sourceTypeKey(l.classType, map[*checker.Type]bool{}))
	}
	for _, parameter := range node.TypeParameters() {
		concrete := l.concrete(l.checker.GetTypeAtLocation(parameter.Name()))
		// The runtime representation merges distinct object layouts into ir.Object.
		// Specializations need the concrete checker type, as generic.go does.
		parts = append(parts, parameter.Name().Text()+"="+l.sourceTypeKey(concrete, map[*checker.Type]bool{}))
	}
	identity.Specialization = strings.Join(parts, ";")
	return identity
}

// Class lowering remains active while it calls module generics. Only lexical
// descendants of that class own its specialization; callers cannot name callees.
func (l *lowering) declaredInCurrentClass(node *ast.Node) bool {
	if l.classNode == nil {
		return false
	}
	// Instance and static class lowering store declaration.Name() in classNode.
	// Its parent is the class declaration itself, not the containing source file.
	declaration := l.classNode.Parent
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor == declaration {
			return true
		}
	}
	return false
}

// sourceTypeKey qualifies identically spelled imported types without checker IDs.
func (l *lowering) sourceTypeKey(proven *checker.Type, visiting map[*checker.Type]bool) string {
	if proven == nil {
		return "unread"
	}
	key := l.checker.TypeToString(proven)
	if symbol := proven.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		source := l.sourceIdentity(symbol.Declarations[0])
		key += fmt.Sprintf("@%q/%q", source.Module, source.Declaration)
	}
	if visiting[proven] {
		return key
	}
	visiting[proven] = true
	defer delete(visiting, proven)
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		for _, argument := range l.checker.GetTypeArguments(proven) {
			key += "[" + l.sourceTypeKey(argument, visiting) + "]"
		}
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, part := range proven.Types() {
			key += "[" + l.sourceTypeKey(part, visiting) + "]"
		}
	}
	return key
}
