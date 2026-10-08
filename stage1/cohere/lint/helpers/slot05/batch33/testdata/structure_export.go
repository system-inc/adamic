package structure

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func AdamicObserve(mode string, node *ast.Node) string {
	if mode == "type" {
		return typeReferenceName(node)
	}
	answer := false
	if mode == "network" {
		answer = isNetworkServiceHookCall(node)
	} else {
		answer = IsLikelyReactComponent(node)
	}
	if answer {
		return "true"
	}
	return "false"
}
