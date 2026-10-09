package lower

import (
	"fmt"
	"path/filepath"
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
	root := filepath.Dir(l.program.FileName(l.program.Files()[0]))
	relative, err := filepath.Rel(root, l.program.FileName(module))
	if err != nil {
		panic("lower: source module path: " + err.Error())
	}
	return filepath.ToSlash(relative)
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
	if l.classType != nil {
		parts = append(parts, l.sourceTypeKey(l.classType, map[*checker.Type]bool{}))
	}
	for _, parameter := range node.TypeParameters() {
		concrete := l.concrete(l.checker.GetTypeAtLocation(parameter.Name()))
		if held, ok := l.substitution[l.checker.GetTypeAtLocation(parameter.Name())]; ok {
			parts = append(parts, fmt.Sprint(held))
		} else {
			parts = append(parts, parameter.Name().Text()+"="+l.sourceTypeKey(concrete, map[*checker.Type]bool{}))
		}
	}
	identity.Specialization = strings.Join(parts, ";")
	return identity
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
