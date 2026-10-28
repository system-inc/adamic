package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Only program modules reach refuse. The compiler's library, prelude and generated
// platform bindings remain the trusted declarations of values supplied by the runtime.
func (l *lowering) referenceRefusal(module *ast.SourceFile) error {
	var first *ast.FileReference
	var kind string
	for _, group := range []struct {
		kind string
		refs []*ast.FileReference
	}{
		{"path", module.ReferencedFiles},
		{"types", module.TypeReferenceDirectives},
		{"lib", module.LibReferenceDirectives},
	} {
		for _, reference := range group.refs {
			if first == nil || reference.Pos() < first.Pos() {
				first, kind = reference, group.kind
			}
		}
	}
	if first == nil {
		return nil
	}
	line, column := scanner.GetLineAndCharacterOfPosition(module, first.Pos())
	return &Refused{
		Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1),
		What:  "a triple-slash reference " + kind + " directive",
		Fix:   "remove the directive and import a module that defines its values; use Adamic's library or generated platform bindings for host values",
	}
}

func ambientDeclaration(node *ast.Node) bool {
	return node.Flags&ast.NodeFlagsAmbient != 0 || ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
}

// Type-only namespaces do not promise a runtime object. Look through dotted and
// nested namespaces so the refusal names the outermost declaration holding values.
func ambientNamespaceValues(node *ast.Node) bool {
	for _, member := range namespaceStatements(node) {
		switch member.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		case ast.KindModuleDeclaration:
			if ambientNamespaceValues(member) {
				return true
			}
		default:
			return true
		}
	}
	return false
}

func (l *lowering) ambientRefusal(node *ast.Node) error {
	what := ""
	if node.Kind == ast.KindModuleDeclaration {
		switch {
		case ast.IsGlobalScopeAugmentation(node):
			what = "a global augmentation"
		case !ast.IsIdentifier(node.Name()):
			what = "a module augmentation or ambient external module"
		case ambientDeclaration(node) && ambientNamespaceValues(node):
			what = "an ambient namespace or module holding values"
		}
	} else if ambientDeclaration(node) {
		switch node.Kind {
		case ast.KindVariableStatement:
			what = "an ambient value declaration"
		case ast.KindFunctionDeclaration:
			what = "an ambient function declaration"
		case ast.KindClassDeclaration:
			what = "an ambient class declaration"
		case ast.KindEnumDeclaration:
			what = "an ambient enum declaration"
		}
	}
	if what == "" {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: what, Fix: "define the values in a module and import them; use Adamic-generated platform bindings for host values"}
}

func init() { refusalRules.register(2200, "ambientRefusal", (*lowering).ambientRefusal) }
