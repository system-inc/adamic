package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The effective compiler enum, not the rule's file/module classification.
func (p *Program) emitModuleKind(out *fields, node *ast.Node, question string) (string, error) {
	if question != "emit-module-kind" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("emit-module-kind requires a SourceFile")
	}
	out.number(uint64(p.Compiler.Options().GetEmitModuleKind()))
	return out.String(), nil
}
