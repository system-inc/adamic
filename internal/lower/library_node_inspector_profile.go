package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// The registry's stable names omit namespace identifiers. A namespace constraint
// keeps Runtime.CallFrame.url separate from Debugger.CallFrame.url, in direct
// accesses and host-object views alike. Existing registrations remain unscoped.
var nodeLibraryMemberNamespaces = map[string]string{}

func registerNodeLibraryMemberInNamespace(name, namespace string) {
	RegisterNodeLibraryMembers(name)
	nodeLibraryMemberNamespaces[name] = namespace
}

func nodeLibraryMemberImplemented(symbol *ast.Symbol) bool {
	name := nodeLibraryName(symbol)
	if !implementedNodeMembers[name] {
		return false
	}
	namespace, scoped := nodeLibraryMemberNamespaces[name]
	if !scoped {
		return true
	}
	if len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		found := ""
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil {
				if parent.Name().Kind == ast.KindStringLiteral {
					break
				}
				found = parent.Name().Text()
				break
			}
		}
		if found != namespace {
			return false
		}
	}
	return true
}

func init() {
	// These are structural program data, never inspector handles or operations.
	registerNodeLibraryMemberInNamespace("node:inspector.Profile.nodes", "Profiler")
	registerNodeLibraryMemberInNamespace("node:inspector.ProfileNode.callFrame", "Profiler")
	registerNodeLibraryMemberInNamespace("node:inspector.CallFrame.url", "Runtime")
}
