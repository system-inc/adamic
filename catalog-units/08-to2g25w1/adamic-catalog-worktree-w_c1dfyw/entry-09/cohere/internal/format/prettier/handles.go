// Package prettier formats source by running our own Prettier fork inside a JavaScript engine.
//
// The fork at ~/Projects/system/prettier is what formats this codebase today, so running it is the
// only way to be byte-identical with the gate we are replacing rather than merely close to it. The
// alternative that was measured and rejected was typescript-go's own formatter: FormatCodeSettings
// has no printWidth field and no line-breaking engine at all, and 21.8% of tracked files diverge
// from our Prettier after every settings-reachable fix.
package prettier

import (
	"path/filepath"
	"strings"

	"github.com/system-inc/cohere/internal/types/sourcename"
)

// parserForExtension maps a file extension to the Prettier parser that handles it.
//
// This is the whole of what the engine can format, and it is deliberately a table rather than a
// try-and-see. A caller needs to know whether a file is a candidate before formatting it, and
// learning by attempting means running the formatter to discover it should not have run.
//
// package.json is not in here because it is not an extension question: Prettier infers the
// json-stringify parser for that one filename specifically, which preserves npm's own array
// formatting. Keying off the extension alone gives it the plain json parser and reformats the file
// away from what the existing gate produces. That is measured, not theoretical -- it is the one
// disagreement in 1,556 files when the probe got it wrong.
var parserForExtension = map[string]string{
	".ts":      "typescript",
	".tsx":     "typescript",
	".js":      "babel",
	".jsx":     "babel",
	".mjs":     "babel",
	".cjs":     "babel",
	".json":    "json",
	".css":     "css",
	".scss":    "scss",
	".less":    "less",
	".md":      "markdown",
	".graphql": "graphql",
	".gql":     "graphql",
	".yaml":    "yaml",
	".yml":     "yaml",
}

// parserFor returns the Prettier parser for a file, and whether this engine handles it at all.
//
// The filename matters and not only its extension: Prettier's own file-info inference picks
// json-stringify for package.json, and a formatter that used plain json there would rewrite npm's
// formatting on every run.
//
// The extension is the one the name is treated as, so an Adamic `.a` file parses with typescript, as the
// same file named `.ts` would (#6mhafvb).
func parserFor(fileName string) (string, bool) {
	if filepath.Base(fileName) == "package.json" {
		return "json-stringify", true
	}
	parser, ok := parserForExtension[strings.ToLower(filepath.Ext(sourcename.TreatedAs(fileName)))]
	return parser, ok
}
