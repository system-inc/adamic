package checker

import (
	"context"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) inspectWave24(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	questions := map[string]func(*fields, *checker.Checker, *ast.Node, string) error{
		"class-this-types":         p.classThisTypes,
		"then-callback-signatures": p.thenCallbackSignatures,
		"handler-parameter-types":  p.handlerParameterTypes,
		"program-module-edges":     p.programModuleEdges,
		"resolved-call-origin":     p.resolvedCallOrigin,
		"symbol-provenance":        p.symbolProvenance,
		"symbol-provenance-alias":  p.symbolProvenance,
	}
	ask := questions[strings.Split(question, "\n")[0]]
	if ask == nil {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	if err := ask(out, c, node, question); err != nil {
		return "", err
	}
	return out.String(), nil
}

// InspectWave24 routes new raw questions while leaving the original dispatcher
// and its refusal mutation anchors unchanged. The C ABI registers this once.
func (p *Program) InspectWave24(file string, start, end uint64, kind, question string) (string, error) {
	mode := strings.Split(question, "\n")[0]
	if mode != "class-this-types" && mode != "then-callback-signatures" && mode != "handler-parameter-types" && mode != "symbol-provenance" && mode != "symbol-provenance-alias" && mode != "resolved-call-origin" && mode != "program-module-edges" {
		return p.Inspect(file, start, end, kind, question)
	}
	source, node, err := p.exact(file, start, end, kind)
	if err != nil {
		return "", err
	}
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	out := &fields{}
	out.number(1)
	out.text(mode)
	return p.inspectWave24(out, c, node, question)
}
