package lower

import (
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A member a program declares as any holds whatever was stored in it, so every read of it through a
// typed view is unchecked (#tmdhz1d). #bw3xg7c rules any refused, written and flowing inferred alike, so a
// property the program declares (an interface's or a type literal's property signature, a class's field)
// whose type is any is refused where it's declared, written (text: any, or an alias of any) or inferred
// from its initializer (x = JSON.parse(source)). A member of a lib or prelude declaration isn't the
// program's, so it's left to the relation rule (refuseAnyRelation) and the axiom audit. A member whose
// any is nested (items: any[]) is a relation's to judge where it's read as a typed member, which
// refuseAnyRelation does.
func (l *lowering) refuseAnyMember(node *ast.Node) error {
	if node.Kind != ast.KindPropertySignature && node.Kind != ast.KindPropertyDeclaration {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(node.Name())
	if symbol == nil {
		return nil
	}
	if l.checker.GetTypeOfSymbol(symbol).Flags()&checker.TypeFlagsAny == 0 {
		return nil
	}
	if site := anyMemberSite(node); anyMemberAllowance[site] {
		return l.notYet(node, "a member typed any ("+site+") that step 3's JSON design types (#3f4rr76): tsc's public JSON API takes any here")
	}
	written := "declared"
	if node.Type() == nil {
		written = "inferred from its initializer"
	}
	return &Refused{Where: l.program.Where(node), What: "a member typed any (" + scannedMemberName(node) + ", " + written + "): any has no proven type, so every read of it through a typed view is unchecked", Fix: "type it unknown and narrow it where it's read, or give it its real type (no-explicit-any)"}
}

// anyMemberAllowance are tsc's four members typed any whose type is fixed by its public JSON API's any
// inputs and returns, which step 3's JSON design types (#3f4rr76): each is NotYet, naming that, rather
// than Refused, and every other member typed any is Refused (Adamic's ruling on #tmdhz1d, Nov 1). A NotYet
// trusts nothing: each still stops. The list goes in the change that lands step 3's JSON design.
var anyMemberAllowance = map[string]bool{
	"types.ts ParsedCommandLine.raw":                     true,
	"commandLineParser.ts readConfigFile.config":            true,
	"commandLineParser.ts parseConfigFileTextToJson.config": true,
	"commandLineParser.ts ParsedTsconfig.raw":              true,
}

// anyMemberSite is a member's file, its enclosing declaration and its name, as anyMemberAllowance keys it.
func anyMemberSite(node *ast.Node) string {
	container := "?"
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindInterfaceDeclaration || parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindTypeAliasDeclaration || parent.Kind == ast.KindFunctionDeclaration {
			if name := parent.Name(); name != nil {
				container = name.Text()
			}
			break
		}
	}
	return filepath.Base(string(ast.GetSourceFileOfNode(node).FileName())) + " " + container + "." + scannedMemberName(node)
}

// scannedMemberName is a member's name as written.
func scannedMemberName(node *ast.Node) string {
	if name := node.Name(); name != nil {
		if ast.IsIdentifier(name) || ast.IsStringLiteral(name) || ast.IsNumericLiteral(name) {
			return name.Text()
		}
	}
	return "a computed name"
}

func init() { refusalRules.register(3280, "refuseAnyMember", (*lowering).refuseAnyMember) }
