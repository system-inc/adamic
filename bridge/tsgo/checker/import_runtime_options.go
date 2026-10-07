package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func (p *Program) importRuntimeOptions(out *fields, node *ast.Node, question string) (string, error) {
	if question != "import-runtime-options" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("import-runtime-options requires a SourceFile")
	}
	options := p.Compiler.Options()
	out.yes(options.EmitDecoratorMetadata.IsTrue())
	out.text(options.JsxFactory)
	out.text(options.JsxFragmentFactory)
	return out.String(), nil
}
