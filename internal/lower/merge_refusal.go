package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func init() { refusalRules.register(2075, "classMergeRefusal", (*lowering).classMergeRefusal) }

func (l *lowering) classMergeRefusal(node *ast.Node) error {
	if node.Kind != ast.KindInterfaceDeclaration && node.Kind != ast.KindModuleDeclaration {
		return nil
	}
	if !ast.IsIdentifier(node.Name()) || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	if node.Kind == ast.KindInterfaceDeclaration && len(node.Members()) == 0 {
		return nil
	}
	symbol := l.symbol(node.Name())
	if symbol == nil || symbol.Flags&ast.SymbolFlagsClass == 0 {
		return nil
	}

	instanceNames, staticNames, indexKeys := l.classMergeDeclarations(symbol)
	var kind, member, fix string
	if node.Kind == ast.KindInterfaceDeclaration {
		kind = "an interface"
		for _, candidate := range node.Members() {
			if name := candidate.Name(); name != nil {
				if key := l.mergeMemberName(name); key != ast.InternalSymbolNameMissing && instanceNames[key] {
					continue
				}
				member = sourceExpression(name)
				fix = "declare " + member + " on the class itself"
			} else {
				if candidate.Kind == ast.KindIndexSignature && indexKeys[l.mergeIndexKey(candidate)] {
					continue
				}
				member = strings.TrimPrefix(candidate.Kind.String(), "Kind")
				fix = "declare the class's members on the class itself"
			}
			break
		}
	} else {
		kind = "a namespace"
		member = firstNamespaceMergeMember(node, symbol, staticNames)
		fix = "declare " + member + " as a static member on the class itself or export it separately"
	}
	if member == "" {
		return nil
	}
	return &Refused{
		Where: l.program.Where(node),
		What:  fmt.Sprintf("%s declaration merged into class %s, adding member %s", kind, node.Name().Text(), member),
		Fix:   fix,
	}
}

// Walk the class declarations' own extends clauses. The checker's merged
// instance type can also inherit members through an interface's extends clause.
func (l *lowering) classMergeDeclarations(symbol *ast.Symbol) (instanceNames, staticNames map[string]bool, indexKeys map[*checker.Type]bool) {
	instanceNames, staticNames = map[string]bool{}, map[string]bool{}
	indexKeys = map[*checker.Type]bool{}
	seen := map[*ast.Symbol]bool{}
	var visit func(*ast.Symbol)
	visit = func(class *ast.Symbol) {
		if class == nil || seen[class] {
			return
		}
		seen[class] = true
		for _, declaration := range class.Declarations {
			if !ast.IsClassLike(declaration) {
				continue
			}
			for _, member := range classMembersWithParameters(declaration) {
				switch {
				case ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic):
					if name := member.Name(); name != nil {
						staticNames[l.mergeMemberName(name)] = true
					}
				case member.Kind == ast.KindIndexSignature:
					indexKeys[l.mergeIndexKey(member)] = true
				default:
					if name := member.Name(); name != nil {
						instanceNames[l.mergeMemberName(name)] = true
					}
				}
			}
			for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
				if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
					base := clause.AsHeritageClause().Types.Nodes[0].Expression()
					visit(l.checker.GetTypeAtLocation(base).Symbol())
				}
			}
		}
	}
	visit(symbol)
	return instanceNames, staticNames, indexKeys
}

func (l *lowering) mergeMemberName(name *ast.Node) string {
	if name.Kind == ast.KindPrivateIdentifier {
		return ast.InternalSymbolNameMissing
	}
	key := ast.GetPropertyNameForPropertyNameNode(name)
	if key == ast.InternalSymbolNameMissing {
		keyType := l.checker.GetTypeAtLocation(name.Expression())
		if checker.IsTypeUsableAsPropertyName(keyType) {
			return checker.GetPropertyNameFromType(keyType)
		}
	}
	return key
}

func (l *lowering) mergeIndexKey(member *ast.Node) *checker.Type {
	return l.checker.GetTypeAtLocation(member.Parameters()[0])
}

// The merged export table also contains class statics and exports from other
// namespace blocks. Pick the first export declared by this block in source order.
func firstNamespaceMergeMember(namespace *ast.Node, symbol *ast.Symbol, staticNames map[string]bool) string {
	var first *ast.Node
	var name string
	for key, exported := range symbol.Exports {
		if staticNames[key] {
			continue
		}
		for _, declaration := range exported.Declarations {
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				if parent == namespace {
					if first == nil || declaration.Pos() < first.Pos() {
						first, name = declaration, ast.SymbolName(exported)
					}
					break
				}
			}
		}
	}
	return name
}
