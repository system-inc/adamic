package structure

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func AdamicParameterNodes(parameters *ast.NodeList) []*ast.Node { return parameterNodes(parameters) }
