package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) emitModuleKind(out *fields, node *ast.Node) error {
	if node.Kind != ast.KindSourceFile {
		return fmt.Errorf("emit-module-kind requires a SourceFile")
	}
	out.number(uint64(p.Compiler.Options().GetEmitModuleKind()))
	return nil
}

func init() {
	additionalQuestions["emit-module-kind"] = func(p *Program, out *fields, c *checker.Checker, node *ast.Node) error {
		return p.emitModuleKind(out, node)
	}
}
