package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/bridge/tsgo/checker/regex_pattern/comments"
)

func (p *Program) sourceComments(out *fields, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "source-comments" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-comments requires a SourceFile")
	}
	ranges := comments.All(source)
	out.number(uint64(len(ranges)))
	for _, comment := range ranges {
		out.number(uint64(comment.Range.Pos()))
		out.number(uint64(comment.Range.End()))
	}
	return out.String(), nil
}
