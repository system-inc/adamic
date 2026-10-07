package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// The merged Buffer unit implements these members. Register them with the shared
// declaration guard so byte reads can be consumed by sys.ts's actual BOM decoder.
func init() {
	RegisterNodeLibraryMembers("node:fs.native", "node:path.join", "node:path.resolve", "node:path.dirname", "node:path.relative", "node:path.PlatformPath.join", "node:path.PlatformPath.resolve", "node:path.PlatformPath.dirname", "node:path.PlatformPath.relative", "node:fs.readdirSync", "node:fs.realpathSync", "node:fs.realpathSync.native", "node:fs.Dirent.isFile", "node:fs.Dirent.isDirectory", "node:fs.Dirent.isSymbolicLink", "node:fs.Dirent.name")
	RegisterNodeLibraryMembers("node:buffer.Buffer", "node:buffer.BufferConstructor.from", "node:buffer.Buffer.toString",
		"node:crypto.createHash", "node:crypto.Hash.update", "node:crypto.Hash.digest")
}

// The readSync generic constraint is ArrayBufferView. Its proven Buffer argument
// is synchronously borrowed as bytes, never exposed as a different object view.
func (l *lowering) nodeFSFileBufferArgument(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := outer.Parent.AsCallExpression()
	if len(call.Arguments.Nodes) < 2 || call.Arguments.Nodes[1] != outer || !l.nodeBufferType(l.checker.GetTypeAtLocation(node), "Buffer") {
		return false
	}
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	if symbol == nil || (symbol.Name != "readSync" && symbol.Name != "writeFileSync") {
		return false
	}
	if symbol.Name == "readSync" && len(call.Arguments.Nodes) != 5 {
		return false
	}
	if symbol.Name == "writeFileSync" && len(call.Arguments.Nodes) > 3 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if load.IsNodeLibrary(file) && strings.HasSuffix(file.FileName().AsString(), "/fs.d.ts") {
			return true
		}
	}
	return false
}
