package lower

import (
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (lowering *lowering) debugSourceLine(node *ast.Node) ir.SourceLine {
	return ir.SourceLine{
		File: lowering.debugModulePath(ast.GetSourceFileOfNode(node)),
		Line: lowering.program.Line(node),
	}
}

// modulePath now names import routes. Debuggers need the actual entry-relative
// file instead; keep this rule separate from runtime declaration identities.
func (lowering *lowering) debugModulePath(file *ast.SourceFile) string {
	identity := lowering.modulePath(file)
	if file == lowering.entry ||
		strings.HasPrefix(identity, "bundled://") ||
		strings.HasPrefix(identity, "adamic-prelude://") {
		return identity
	}
	path, relativePathError := filepath.Rel(
		filepath.Dir(lowering.program.FileName(lowering.entry)),
		lowering.program.FileName(file),
	)
	if relativePathError != nil {
		panic("lower: relative debug source path: " + relativePathError.Error())
	}
	return filepath.ToSlash(path)
}
