package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/bridge/tsgo/checker/executiongraph"
	"strconv"
	"strings"
)

// astContext exposes structural compiler metadata, never lint judgments.
func (p *Program) astContext(out *fields, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if parts[0] != "ast-context" || (len(parts) != 1 && len(parts) != 5) {
		return "", fmt.Errorf("invalid AST context selector")
	}
	if len(parts) == 5 {
		source := p.Compiler.GetSourceFile(parts[1])
		if source == nil {
			return "", fmt.Errorf("AST context source is not in this program")
		}
		start, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil || strconv.FormatUint(start, 10) != parts[2] {
			return "", fmt.Errorf("invalid AST context start")
		}
		end, err := strconv.ParseUint(parts[3], 10, 64)
		if err != nil || strconv.FormatUint(end, 10) != parts[3] || start > end || end > uint64(len(source.Text())) {
			return "", fmt.Errorf("invalid AST context end")
		}
		var selected *ast.Node
		var visit func(*ast.Node) bool
		visit = func(candidate *ast.Node) bool {
			if uint64(candidate.Pos()) == start && uint64(candidate.End()) == end && candidate.Kind.String() == "Kind"+parts[4] {
				selected = candidate
				return true
			}
			return candidate.ForEachChild(visit)
		}
		visit(source.AsNode())
		if selected == nil {
			return "", fmt.Errorf("no exact AST context node")
		}
		node = selected
	}

	file := ast.GetSourceFileOfNode(node)
	out.yes(file.IsDeclarationFile)
	out.yes(ast.IsExternalModule(file))
	out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.Path()))
	root := executiongraph.RootOf(node)
	out.yes(root != nil)
	if root != nil {
		out.text(strings.TrimPrefix(root.Kind.String(), "Kind"))
		out.number(uint64(root.Pos()))
		out.number(uint64(root.End()))
	}
	for current, depth := node, 0; depth < 6; depth++ {
		out.yes(current != nil)
		if current != nil {
			out.text(strings.TrimPrefix(current.Kind.String(), "Kind"))
			out.number(uint64(current.Pos()))
			out.number(uint64(current.End()))
			out.number(uint64(current.Flags))
			name := ""
			if declared := current.Name(); declared != nil {
				switch declared.Kind {
				case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
					name = declared.Text()
				}
			}
			out.text(name)
			out.yes(ast.IsGlobalScopeAugmentation(current))
			out.yes(current.Flags&ast.NodeFlagsConst != 0)
			out.yes(current.Flags&ast.NodeFlagsAmbient != 0)
			body := current.Body()
			out.yes(body != nil)
			if body != nil {
				out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
				out.number(uint64(body.Pos()))
				out.number(uint64(body.End()))
			}
			flags := ast.GetFunctionFlags(current)
			out.yes(flags&ast.FunctionFlagsGenerator != 0)
			out.yes(flags&ast.FunctionFlagsAsync != 0)
			current = current.Parent
		}
	}
	return out.String(), nil
}
