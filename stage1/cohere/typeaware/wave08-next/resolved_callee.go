package wave08next

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// ResolvedCalleeFields supplies raw checker identity and syntax for one resolved
// signature. It is not registered: the shared bridge dispatcher is outside this
// continuation's permitted territory. No lint decision is returned.
func ResolvedCalleeFields(c *checker.Checker, node *ast.Node) ([]string, error) {
	if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression && node.Kind != ast.KindTaggedTemplateExpression {
		return nil, fmt.Errorf("resolved callee requires call-like node")
	}
	out := []string{"1", "wave08-resolved-callee"}
	signature := c.GetResolvedSignature(node)
	if signature == nil || signature.Declaration() == nil {
		return append(out, "0"), nil
	}
	declaration := signature.Declaration()
	file := ast.GetSourceFileOfNode(declaration)
	if file == nil {
		return nil, fmt.Errorf("signature declaration has no source")
	}
	flag := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	out = append(out, "1", file.FileName(), strings.TrimPrefix(declaration.Kind.String(), "Kind"), strconv.Itoa(declaration.Pos()), strconv.Itoa(declaration.End()), flag(file.IsDeclarationFile), flag(ast.IsExternalModule(file)), strconv.FormatUint(uint64(ast.GetFunctionFlags(declaration)), 10))
	body := declaration.Body()
	out = append(out, flag(body != nil))
	if body != nil {
		out = append(out, strconv.Itoa(body.Pos()), strconv.Itoa(body.End()))
	}
	flags := uint64(0)
	if returned := c.GetReturnTypeOfSignature(signature); returned != nil {
		flags = uint64(returned.Flags())
	}
	return append(out, strconv.FormatUint(flags, 10)), nil
}
