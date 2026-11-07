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
func (lowering *lowering) moduleStatements(module *ast.SourceFile) ([]ir.Statement, error) {
	// A module's own calls are judged as a body's are (predicates_branded.go).
	if judgeBrandedCallsError := lowering.judgeBrandedCalls(module.AsNode()); judgeBrandedCallsError != nil {
		return nil, judgeBrandedCallsError
	}
	body, statementsError := lowering.statements(module.Statements.Nodes)
	if statementsError == nil {
		lowering.result.MainModules = append(
			lowering.result.MainModules,
			ir.MainModule{Module: lowering.modulePath(module), Statements: len(body)},
		)
	}
	return body, statementsError
}

func (lowering *lowering) modulePath(module *ast.SourceFile) string {
	if load.IsPrelude(module) {
		return "adamic-prelude:///adamic.d.ts"
	}
	filename := lowering.program.FileName(module)
	if strings.HasPrefix(filename, "bundled://") {
		return filename
	}
	if lowering.sourceModules == nil {
		lowering.indexSourceModules()
	}
	if lowering.sourceModules[module] == "" {
		panic("lower: source module has no import-declaration route: " + filename)
	}
	return lowering.sourceModules[module]
}

// Module identities follow source import declarations, not file placement. The
// shortest route, with lexical tie-breaking, is independent of traversal order.
// Binding clauses exclude the module specifier: moving a source cluster may
// rewrite relative specifiers to absolute paths without changing its bindings.
// Each outgoing label is unique within its source module, so distinct modules
// never acquire the same route, even with identical basenames and contents.
func (lowering *lowering) indexSourceModules() {
	entry := lowering.program.Files()[0]
	lowering.sourceModules = map[*ast.SourceFile]string{entry: filepath.Base(lowering.program.FileName(entry))}
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
					label += fmt.Sprintf(
						":%t:%s",
						clause.IsTypeOnly(),
						strings.TrimSpace(current.module.Text()[clause.Pos():clause.End()]),
					)
				}
			case ast.KindImportEqualsDeclaration:
				imported := node.AsImportEqualsDeclaration()
				if reference := imported.ModuleReference; reference.Kind == ast.KindExternalModuleReference {
					specifier = reference.AsExternalModuleReference().Expression
					label = fmt.Sprintf("import-equals:%t:%s", imported.IsTypeOnly, node.Name().Text())
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
					if name := parent.Name(); name != nil &&
						(ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) {
						parents = append(parents, fmt.Sprintf("%d:%s", parent.Kind, name.Text()))
					}
				}
				label = fmt.Sprintf("import-type:%q", parents)
			}
			if specifier != nil {
				if symbol := lowering.checker.GetSymbolAtLocation(specifier); symbol != nil {
					for _, declaration := range symbol.Declarations {
						if declaration.Kind != ast.KindSourceFile {
							continue
						}
						target := declaration.AsSourceFile()
						if load.IsPrelude(target) ||
							load.IsLibrary(target) ||
							strings.HasPrefix(lowering.program.FileName(target), "bundled://") {
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
			if lowering.sourceModules[imported.module] != "" {
				continue
			}
			route := append(append([]string{}, current.route...), imported.label)
			lowering.sourceModules[imported.module] = fmt.Sprintf(
				"imported:///%q/%s",
				route,
				filepath.Base(lowering.program.FileName(imported.module)),
			)
			queue = append(queue, pending{module: imported.module, route: route})
		}
	}
}

// sourceIdentity indexes a source tree once, before lowering order can influence
// anonymous numbering. A nested declaration owns a separate anonymous sequence.
func (lowering *lowering) sourceIdentity(node *ast.Node) ir.SourceIdentity {
	if lowering.sourceIdentities == nil {
		lowering.sourceIdentities = map[*ast.Node]ir.SourceIdentity{}
	}
	if identity, found := lowering.sourceIdentities[node]; found {
		return identity
	}
	module := ast.GetSourceFileOfNode(node)
	duplicates := map[string]int{}
	var walk func(*ast.Node, []string, *int)
	walk = func(node *ast.Node, parents []string, anonymous *int) {
		path := parents
		own := (ast.IsFunctionLike(node) && (node.Body() != nil || node.Kind == ast.KindMethodDeclaration)) ||
			node.Kind == ast.KindClassDeclaration ||
			node.Kind == ast.KindInterfaceDeclaration ||
			node.Kind == ast.KindTypeAliasDeclaration ||
			node.Kind == ast.KindVariableDeclaration ||
			node.Kind == ast.KindPropertyDeclaration ||
			node.Kind == ast.KindPropertyAssignment
		if own {
			name := ""
			if node.Kind == ast.KindConstructor {
				name = "constructor"
			} else if node.Name() != nil {
				named := node.Name()
				if named.Kind == ast.KindIdentifier ||
					named.Kind == ast.KindPrivateIdentifier ||
					named.Kind == ast.KindStringLiteral ||
					named.Kind == ast.KindNumericLiteral {
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
		lowering.sourceIdentities[node] = ir.SourceIdentity{Module: lowering.modulePath(module), Declaration: path}
		counter := anonymous
		if own && node.Kind != ast.KindPropertyAssignment {
			counter = new(int)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child, path, counter); return false })
	}
	walk(module.AsNode(), nil, new(int))
	return lowering.sourceIdentities[node]
}

// functionSource includes concrete type spellings, never checker IDs or lowering IDs.
func (lowering *lowering) functionSource(node *ast.Node) ir.SourceIdentity {
	identity := lowering.sourceIdentity(node)
	parts := []string{}
	if lowering.function != nil && lowering.function.Source.Specialization != "" {
		parent := lowering.function.Source
		nested := parent.Module == identity.Module && len(parent.Declaration) < len(identity.Declaration)
		for i := 0; nested && i < len(parent.Declaration); i++ {
			nested = parent.Declaration[i] == identity.Declaration[i]
		}
		if nested {
			parts = append(parts, parent.Specialization)
		}
	}
	if lowering.classType != nil && lowering.declaredInCurrentClass(node) {
		parts = append(parts, lowering.sourceTypeKey(lowering.classType, map[*checker.Type]bool{}))
	}
	for _, parameter := range node.TypeParameters() {
		concrete := lowering.concrete(lowering.checker.GetTypeAtLocation(parameter.Name()))
		// The runtime representation merges distinct object layouts into ir.Object.
		// Specializations need the concrete checker type, as generic.go does.
		parts = append(parts, parameter.Name().Text()+"="+lowering.sourceTypeKey(concrete, map[*checker.Type]bool{}))
	}
	identity.Specialization = strings.Join(parts, ";")
	return identity
}

// Class lowering remains active while it calls module generics. Only lexical
// descendants of that class own its specialization; callers cannot name callees.
func (lowering *lowering) declaredInCurrentClass(node *ast.Node) bool {
	if lowering.classNode == nil {
		return false
	}
	// Instance and static class lowering store declaration.Name() in classNode.
	// Its parent is the class declaration itself, not the containing source file.
	declaration := lowering.classNode.Parent
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor == declaration {
			return true
		}
	}
	return false
}

// sourceTypeKey qualifies identically spelled imported types without checker IDs.
func (lowering *lowering) sourceTypeKey(proven *checker.Type, visiting map[*checker.Type]bool) string {
	if proven == nil {
		return "unread"
	}
	key := lowering.checker.TypeToString(proven)
	if symbol := proven.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		source := lowering.sourceIdentity(symbol.Declarations[0])
		key += fmt.Sprintf("@%q/%q", source.Module, source.Declaration)
	}
	if visiting[proven] {
		return key
	}
	visiting[proven] = true
	defer delete(visiting, proven)
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		for _, argument := range lowering.checker.GetTypeArguments(proven) {
			key += "[" + lowering.sourceTypeKey(argument, visiting) + "]"
		}
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, part := range proven.Types() {
			key += "[" + lowering.sourceTypeKey(part, visiting) + "]"
		}
	}
	return key
}
