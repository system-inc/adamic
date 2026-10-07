package load

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Config parsing sees .a through exactly the same virtual .a.ts names as source checking.
// Only file specifications change, using the upstream JSON syntax tree so comments, extends,
// compiler options, and validation remain the checker's responsibility.
type projectFS struct{ vfs.FS }

func (s *projectFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	text, ok := s.FS.ReadFile(path)
	if !ok || !strings.HasSuffix(path.AsString(), ".json") {
		return text, ok
	}
	config := tsoptions.NewTsconfigSourceFileFromFilePath(path, s.FS.CaseSensitivity().PathKey(path.AsPath()), text)
	type replacement struct {
		start, end int
		text       string
	}
	var replacements []replacement
	depth := 0
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			depth++
			defer func() { depth-- }()
		}
		if node.Kind == ast.KindPropertyAssignment && depth == 1 {
			property := node.AsPropertyAssignment()
			name := property.Name().Text()
			if (name == "files" || name == "include" || name == "exclude") && property.Initializer != nil && property.Initializer.Kind == ast.KindArrayLiteralExpression {
				for _, element := range property.Initializer.AsArrayLiteralExpression().Elements.Nodes {
					if element.Kind == ast.KindStringLiteral && strings.HasSuffix(element.Text(), ".a") && !s.FS.DirectoryExists(path.Directory().ResolveDirectory(element.Text())) {
						// Marshaling a string cannot fail and preserves JSON escapes in file names.
						quoted, _ := json.Marshal(element.Text() + ".ts")
						replacements = append(replacements, replacement{scanner.GetTokenPosOfNode(element, config.SourceFile, false), element.End(), string(quoted)})
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	visit(config.SourceFile.AsNode())
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].start > replacements[j].start })
	for _, replacement := range replacements {
		text = text[:replacement.start] + replacement.text + text[replacement.end:]
	}
	return text, true
}
