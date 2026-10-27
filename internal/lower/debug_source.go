package lower

import (
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) debugSourceLine(node *ast.Node) ir.SourceLine {
	return ir.SourceLine{File: l.debugModulePath(ast.GetSourceFileOfNode(node)), Line: l.program.Line(node)}
}

// modulePath now names import routes. Debuggers need the actual entry-relative
// file instead; keep this rule separate from runtime declaration identities.
func (l *lowering) debugModulePath(file *ast.SourceFile) string {
	identity := l.modulePath(file)
	if file == l.entry || strings.HasPrefix(identity, "bundled://") || strings.HasPrefix(identity, "adamic-prelude://") {
		return identity
	}
	path, err := filepath.Rel(filepath.Dir(l.program.FileName(l.entry)), l.program.FileName(file))
	if err != nil {
		panic("lower: relative debug source path: " + err.Error())
	}
	return filepath.ToSlash(path)
}
