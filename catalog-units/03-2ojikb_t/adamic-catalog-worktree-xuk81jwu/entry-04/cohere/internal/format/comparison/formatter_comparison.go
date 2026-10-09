// Package formatdiff measures typescript-go's formatter against the Prettier fork that formats this
// codebase today.
//
// This exists because the question it answers decides whether `cohere --format` can replace
// Prettier at all, and the answer cannot come from reading either implementation. Both are large,
// both are correct about different things, and a formatter that silently reflows a tree produces a
// diff nobody asked for that buries real changes in noise. So the deliverable is a measurement over
// real files, and the taxonomy of what differs rather than the count.
//
// The corpus is the ahra tree, which `s pnc` keeps formatted by @system-inc/prettier. Every file in
// it is a fixed point of that formatter, which is what makes it usable as an oracle: any change
// FormatDocument proposes against an already-formatted file is, by construction, a divergence.
package formatter_comparison

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/format"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// Settings returns the FormatCodeSettings tuned as close to our Prettier fork as the settings can
// reach, measured rather than assumed. Each non-default here closes a divergence class that showed
// up in the corpus sweep; each one is pinned by a test in this package.
//
// The Prettier config in ~/Projects/ahra is tabWidth 4, useTabs false, semi true, singleQuote true,
// printWidth 120. Two of those five have no field to map onto at all: `printWidth`, because
// FormatCodeSettings has no width of any kind, and `singleQuote`, because the formatter never
// rewrites a string literal. Those are the structural gaps, not omissions here.
func Settings(newLine string) format.FormatCodeSettings {
	settings := format.GetDefaultFormatCodeSettings()
	settings.NewLineCharacter = newLine

	settings.TabSize = 4
	settings.IndentSize = 4
	settings.ConvertTabsToSpaces = core.TSTrue

	// Our house style is `if(x)`, not `if (x)`. This is the single largest divergence class in the
	// corpus by file count, and it is one boolean.
	settings.InsertSpaceAfterKeywordsInControlFlowStatements = core.TSFalse

	// "insert" makes the formatter add semicolons after type members, turning `{ a: number }` into
	// `{ a: number; }`, which our Prettier does not do. It is also the setting that reaches the ASI
	// path that panics on some real files. "ignore" leaves existing semicolons alone, which is what
	// an already-semicolon-correct codebase wants.
	settings.Semicolons = format.SemicolonPreferenceIgnore

	// `function() {}` and `const x = {}` keep their empty braces tight rather than becoming `{ }`.
	settings.InsertSpaceAfterOpeningAndBeforeClosingEmptyBraces = core.TSFalse

	return settings
}

// Apply returns what the text becomes once the changes land.
//
// The changes come back in source order, and each one's range is stated against the original text,
// so they are applied back-to-front. Applying forward would invalidate every range after the first
// edit that changed a length, which is the same reason the edit engine sorts before it writes.
func Apply(text string, changes []core.TextChange) string {
	ordered := make([]core.TextChange, len(changes))
	copy(ordered, changes)
	sort.SliceStable(ordered, func(a, b int) bool {
		return ordered[a].TextRange.Pos() > ordered[b].TextRange.Pos()
	})

	var builder strings.Builder
	result := text
	for _, change := range ordered {
		start, end := change.TextRange.Pos(), change.TextRange.End()
		if start < 0 || end > len(result) || start > end {
			// A range outside the text means the change was computed against a different string.
			// Returning the input unchanged would report a false agreement, so this is loud.
			panic("formatdiff: text change out of range")
		}
		builder.Reset()
		builder.WriteString(result[:start])
		builder.WriteString(change.NewText)
		builder.WriteString(result[end:])
		result = builder.String()
	}
	return result
}

// FormatFile parses one file and returns what FormatDocument would make of it.
//
// A file that will not parse returns ok false rather than an empty string: an empty result is
// indistinguishable from a file that formatted to nothing, and that confusion is the exact failure
// this tool exists to eliminate.
func FormatFile(fileName string, text string) (formatted string, ok bool, panicked bool, panicMessage string) {
	scriptKind := core.ScriptKindTS
	switch {
	case strings.HasSuffix(fileName, ".tsx"):
		scriptKind = core.ScriptKindTSX
	case strings.HasSuffix(fileName, ".jsx"):
		scriptKind = core.ScriptKindJSX
	case strings.HasSuffix(fileName, ".js"), strings.HasSuffix(fileName, ".mjs"):
		scriptKind = core.ScriptKindJS
	}

	if !tspath.IsRootedDiskPath(fileName) {
		fileName = "/" + strings.TrimPrefix(fileName, "/")
	}
	fileName = tspath.NormalizePath(fileName)

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName),
		PathKey:  tspath.PathKey(fileName),
	}, text, scriptKind)
	if sourceFile == nil {
		return "", false, false, ""
	}

	ctx := format.WithFormatCodeSettings(context.Background(), Settings("\n"), "\n")

	// The formatter panics on some real files: its semicolon-insertion check walks to the next
	// token and asserts on the position it finds, and that assertion does not hold everywhere.
	// A panic here is a property of the formatter worth counting, not a reason to lose the sweep,
	// so it is recovered and reported as a distinct outcome rather than swallowed into "differs".
	defer func() {
		if recovered := recover(); recovered != nil {
			formatted, ok, panicked = "", false, true
			panicMessage = fmt.Sprint(recovered)
		}
	}()

	return Apply(text, format.FormatDocument(ctx, sourceFile)), true, false, ""
}
