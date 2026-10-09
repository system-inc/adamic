package edit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/replace"
	"github.com/system-inc/cohere/internal/types/sourcename"
)

// Parses reports whether text parses cleanly as the script kind implied by its file name.
//
// This is the refusal guard, and it is the single reason autofix can be trusted at all. A rule
// proposes bytes; nothing about a byte range guarantees the result is a program. A range computed
// one token off, a replacement missing a closing brace, two fixes that are individually fine and
// jointly nonsense — all of them produce text that looks plausible in a diff and does not compile.
// Parsing before writing turns every one of those into a refusal instead of a corrupted file.
//
// The check is syntax only, deliberately. A fix that leaves the file parseable but ill-typed is a
// bad fix and will be caught by the types phase on the same run, in the working tree, before
// commit. A fix that leaves the file unparseable destroys the ability of every later phase to say
// anything at all, which is why it is the one thing that must never reach disk.
//
// Measured: an unbalanced brace reports TS1005, a stray paren TS1005, a truncated statement
// TS1109, an unterminated string TS1002, and a line of garbage six diagnostics. Clean source and
// an empty file report none. The guard has been shown to fire, which is the only way to know it is
// a guard and not a decoration.
// TypeScriptParsable reports whether the parse guard can say anything about a file at all.
//
// The guard is a TypeScript parser, so it answers a question about TypeScript. Pointed at css, json,
// markdown or yaml it reports TS1128 and TS1434 and refuses the file, which is a correct answer to
// a question nobody asked: those files are not malformed, they are not TypeScript.
//
// This matters now that the format phase's universe is the working tree rather than the type graph.
// Before that split every file reaching this package was TypeScript by construction, so the guard's
// scope was adequate by accident, in exactly the way the type graph was an adequate format universe
// by accident. Both assumptions were incidental and both looked like architecture.
//
// A file this returns false for is passed through unguarded, and that is the honest trade: the
// engine cannot check a language it cannot parse, and refusing every such file would mean the
// formatter can never touch css or markdown. The formatter's own parser is the guard for those, and
// a transform that produces unparseable output in its own language fails inside the engine rather
// than here.
//
// An Adamic `.a` file is TypeScript (#6mhafvb), so the name is read through sourcename.TreatedAs. Only the
// program's `.a` files reach the engine: the format walk holds every `.a` back until the program claims
// it, so a static library named libfoo.a is never handed to this parser.
func TypeScriptParsable(fileName string) bool {
	treatedAs := sourcename.TreatedAs(fileName)
	for _, extension := range []string{".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs"} {
		if strings.HasSuffix(treatedAs, extension) {
			return true
		}
	}
	return false
}

func Parses(fileName string, text string) (bool, string) {
	parses, reason, _ := parsesWithTree(fileName, text)
	return parses, reason
}

// parsesWithTree is Parses, also returning the tree it built: nil for a file the TypeScript parser does not
// own, or one that does not parse. The fix engine hands it to the transform, which then parses nothing
// again (Transform).
func parsesWithTree(fileName string, text string) (bool, string, *ast.SourceFile) {
	// A file the TypeScript parser does not own is not something this guard can judge. Reporting it
	// as unparseable would refuse every css and markdown file in the tree with a TypeScript syntax
	// error, which is a true diagnostic about the wrong question.
	if !TypeScriptParsable(fileName) {
		return true, "", nil
	}

	sourceFile := parseText(fileName, text)
	if sourceFile == nil {
		return false, "the parser produced nothing", nil
	}
	return verdictOf(sourceFile)
}

// reasonOf is a refusal's reason, from the parse diagnostics of a file that does not parse.
func reasonOf(diagnostics []*ast.Diagnostic) string {
	// Localized rather than MessageText, which a parse diagnostic leaves empty: every refusal used to read
	// "TS1135: " with the message missing (#v1ah2qq).
	first := diagnostics[0]
	reason := fmt.Sprintf("TS%d: %s", first.Code(), first.Localize(locale.Locale{}))
	if len(diagnostics) > 1 {
		reason = fmt.Sprintf("%s (and %d more)", reason, len(diagnostics)-1)
	}
	return reason
}

// seedFor reports whether seed is the tree parseText would build for text: the same bytes, the script kind
// this file's extension gives the guard, and a name that is a declaration file exactly when this one is.
// Those are all the parser reads besides the module-indicator options, which set a property of the file and
// no node or diagnostic, so the seed's parse diagnostics are the guard's.
func seedFor(fileName string, text string, seed *ast.SourceFile) bool {
	return seed != nil && TypeScriptParsable(fileName) && seed.Text() == text && seed.ScriptKind == guardScriptKind(fileName) &&
		tspath.IsDeclarationFileName(seed.FileName().AsString()) == tspath.IsDeclarationFileName(fileName)
}

// verdictOf is parsesWithTree's answer for a tree in hand.
func verdictOf(sourceFile *ast.SourceFile) (bool, string, *ast.SourceFile) {
	diagnostics := sourceFile.Diagnostics()
	if len(diagnostics) == 0 {
		return true, "", sourceFile
	}
	return false, reasonOf(diagnostics), nil
}

// parseText parses source text under a rooted, normalized file name.
//
// typescript-go stores the file name as an identity and panics on a relative path, so the name is
// rooted here rather than at each call site. The script kind comes from the extension because .tsx
// and .ts disagree about whether a `<T>` is a type argument or an element, and parsing one as the
// other invents syntax errors in a file that was fine.
func parseText(fileName string, text string) *ast.SourceFile {
	rooted := fileName
	if !tspath.IsRootedDiskPath(rooted) {
		rooted = "/" + strings.TrimPrefix(rooted, "/")
	}
	rooted = tspath.NormalizePath(rooted)

	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(rooted),
		PathKey:  tspath.PathKey(rooted),
	}, text, guardScriptKind(fileName))
}

// guardScriptKind is the script kind the guard parses a file as, chosen by extension.
func guardScriptKind(fileName string) core.ScriptKind {
	switch {
	case strings.HasSuffix(fileName, ".tsx"):
		return core.ScriptKindTSX
	case strings.HasSuffix(fileName, ".jsx"):
		return core.ScriptKindJSX
	case strings.HasSuffix(fileName, ".js"), strings.HasSuffix(fileName, ".mjs"), strings.HasSuffix(fileName, ".cjs"):
		return core.ScriptKindJS
	}
	return core.ScriptKindTS
}

// WriteAtomically replaces a file's contents so that no reader ever observes a partial write.
//
// Write the new text to a temp file in the same directory, then rename over the target. Rename
// within a filesystem is atomic, so a concurrent reader sees the old bytes or the new bytes and
// never a truncated file. The same directory matters: a temp file in the system temp directory can
// be on a different filesystem, where rename degrades to copy-and-delete and the atomicity is
// silently lost.
//
// Five agents work in this tree at once, and a half-written file is worse than an unfixed one — an
// unfixed file still compiles. The failure this prevents is not theoretical here: a half-written
// file failing to build and then healing two minutes later was observed in this repo tonight.
//
// The original file's permission bits are preserved, because a fresh temp file is created at 0600
// and renaming it over a source file would quietly change its mode.
func WriteAtomically(fileName string, text string) error {
	directory := filepath.Dir(fileName)

	mode := os.FileMode(0o644)
	if info, err := os.Stat(fileName); err == nil {
		mode = info.Mode().Perm()
	}

	temporary, err := os.CreateTemp(directory, "."+filepath.Base(fileName)+".cohere-*")
	if err != nil {
		return fmt.Errorf("creating a temp file beside %s: %w", fileName, err)
	}
	temporaryName := temporary.Name()

	// Any failure from here leaves a temp file behind unless it is removed, and a stale dotfile in
	// a source directory is its own small mess.
	cleanup := func() {
		temporary.Close()
		os.Remove(temporaryName)
	}

	if _, err := temporary.WriteString(text); err != nil {
		cleanup()
		return fmt.Errorf("writing %s: %w", temporaryName, err)
	}

	// Sync before rename. Without it the rename can be durable while the contents are not, which on
	// a crash produces a file that exists, has the right name, and holds zeros.
	//
	// This is the one guard in this package with no fixture behind it, and it is called out rather
	// than left quiet. Removing this call passes every test here, because the failure it prevents
	// needs the machine to lose power between the rename and the writeback — which no in-process
	// test can induce. Every other guard was verified by deliberately breaking it and watching a
	// fixture go red; this one rests on the argument alone. Treat it accordingly: it is load-bearing
	// and unverified, so do not remove it because the tests stay green.
	if err := temporary.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("syncing %s: %w", temporaryName, err)
	}
	if err := temporary.Close(); err != nil {
		os.Remove(temporaryName)
		return fmt.Errorf("closing %s: %w", temporaryName, err)
	}
	if err := os.Chmod(temporaryName, mode); err != nil {
		os.Remove(temporaryName)
		return fmt.Errorf("setting the mode on %s: %w", temporaryName, err)
	}
	// replace.File rather than os.Rename: on Windows a rename over a file another process is reading is
	// refused for as long as the read takes, and is retried through that.
	if err := replace.File(temporaryName, fileName); err != nil {
		os.Remove(temporaryName)
		return fmt.Errorf("renaming %s over %s: %w", temporaryName, fileName, err)
	}

	return nil
}
