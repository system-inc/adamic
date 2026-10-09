package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/format/printing"
)

// formatEngine formats one file's text, or says it could not.
//
// This is the narrowest possible surface onto whichever formatter the pipeline runs, and it is an
// interface rather than a direct call for a reason that is about this tree rather than about
// abstraction taste: the engine that will implement it does not exist in-tree yet. Naming the shape
// lets the seam, the outcome mapping, and every fixture around it be written and tested now, and
// lets the real engine drop in as one line when it lands. The alternative was inferring the
// concrete package's signature from a message, which is exactly what produced `internal/differential`.
//
// Three outcomes, which is what @system_cohere_format reports and fewer than the four the native
// formatter reports:
//
//   - formatted text, no error   the transform's result
//   - an error                   the formatter could not produce output, whether because the source
//     did not parse or because the formatter itself failed
//   - Handles returning false    this engine does not format that file type at all
//
// The engine never signals "nothing to do" by returning its input unchanged. That is
// indistinguishable from "already correctly formatted", so a formatter that silently handled
// nothing would report a whole tree as clean. Same class as the `\p{...}` bug that returned false
// for every character without throwing: the dangerous failures do not announce themselves.
type formatEngine interface {
	// Format returns the formatted text, or an error.
	Format(fileName string, text string) (string, error)

	// Handles reports whether this engine formats a file's type at all.
	//
	// Separate from Format rather than folded into it as an error, because the caller needs the
	// answer before deciding whether a file belongs in the candidate set, and asking by attempting a
	// format would mean running the formatter to learn it should not have.
	Handles(fileName string) bool

	// Enumerate walks a project root and returns every file the engine would format, with an account
	// of what the walk removed at each step.
	//
	// It belongs on the engine rather than in the pipeline because it is extension and ignore-file
	// knowledge, the same knowledge Handles already encodes. The pipeline decides which of those
	// files are in scope; the engine decides which files are formattable at all.
	Enumerate(root string) (formatfiles.Enumeration, error)

	// OptionsFingerprint names the options a file formats with, so the format record can tell bytes
	// formatted under one config from the same bytes under another. A config edit changes the
	// fingerprint of every file it reaches, and each of those files is formatted again.
	OptionsFingerprint(fileName string) (string, error)
}

// treeFormatter is a format engine that can take a tree of the text the fix engine's guard already parsed,
// rather than parsing the same bytes again (edit.Transform). The native engine is one; an engine that is
// not is handed text alone, as before.
type treeFormatter interface {
	FormatParsed(fileName string, text string, parsed *ast.SourceFile) (string, error)
}

// formatTransform adapts a format engine to the edit engine's whole-text transform.
//
// The edit engine owns what lands on disk, so formatting arrives as a transform rather than writing
// on its own. Two components writing the same file in one run is precisely the failure the edit
// engine exists to prevent, and building that into the pipeline that contains it would be a poor
// joke.
//
// It runs after fixes have converged, which is not arbitrary: fixes are semantic and change what the
// correct formatting is, so formatting first would leave every later fix mis-formatted and want a
// re-format. Formatting is total and idempotent, so running it last converges in one shot regardless
// of what the fixes did.
//
// The whole job here is keeping outcomes distinct that a careless adapter would collapse. A skip and
// a failure mean opposite things to a reader: one says the formatter chose not to look, the other
// says it looked and broke. Collapsing either into "returned the text unchanged" makes a formatter
// that never ran indistinguishable from a tree that was already correct, which is the ambiguity the
// coverage line exists to destroy.
func formatTransform(engine formatEngine) edit.Transform {
	return func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		if engine == nil {
			// Said as what happened. A nil engine is a run that did not ask to format, and every repository
			// cohere gates has a format block, so "no formatter is configured" was false on every plain
			// --no-fix run: `11 not formatted (11 no formatter is configured)` on ahra, which has one.
			return "", fmt.Errorf("%w: not requested", edit.ErrSkipped)
		}

		if !engine.Handles(fileName) {
			return "", fmt.Errorf("%w: %s is not a file type the formatter handles", edit.ErrSkipped, extensionOf(fileName))
		}

		// Formatted until the text stops changing, not once. Prettier is not idempotent on every input,
		// and the native printers match it pass for pass, so one pass can leave a file a second pass
		// still rewrites. A writing run that formats once then fails its own `--no-fix` check on the
		// file it just wrote: 7872fceb did exactly that to two of ahra's files. An already formatted
		// file is the common case and still costs one pass, because its first pass changes nothing.
		//
		// The guard's tree is of the text as given, so only the first pass is offered it.
		current := text
		withTree, takesTree := engine.(treeFormatter)
		for pass := 1; pass <= formatPassLimit; pass++ {
			var formatted string
			var err error
			if takesTree && pass == 1 && parsed != nil {
				formatted, err = withTree.FormatParsed(fileName, current, parsed)
			} else {
				formatted, err = engine.Format(fileName, current)
			}
			if err != nil {
				// A failure is reported as a failure rather than downgraded to a skip. The edit engine
				// keeps the fixes that already converged, so a formatter falling over on one file does
				// not cost the tree its correctness fixes, and the refusal is counted rather than
				// swallowed.
				//
				// The exception is a file that does not parse, which the engine may report as an error
				// and which is genuinely a skip: the edit engine already refuses to fix a file in that
				// state, and the types phase reports it in a form a reader can act on. A second
				// complaint from the formatter adds noise rather than information. Only on the first
				// pass, though: the formatter's own output failing to parse is the formatter breaking.
				if pass == 1 && isUnparseable(err) {
					return "", fmt.Errorf("%w: %s", edit.ErrSkipped, unparseableSkipReason)
				}
				return "", err
			}
			if formatted == current {
				return current, nil
			}
			current = formatted
		}

		// Still changing at the bound. Writing the last pass would call a file formatted that the next
		// run rewrites again, so it is a failure naming the file, and the edit engine keeps the fixes
		// and leaves the formatting undone.
		return "", fmt.Errorf("%s is not idempotent under the formatter: pass %d still changed it", fileName, formatPassLimit)
	}
}

// unparseableSkipReason is the skip a file that does not parse gets. A run whose verdict is formatting
// alone counts it as a file it could not check (see formatOnlyUnchecked), where every other skip is a file
// the formatter was never asked about.
const unparseableSkipReason = "the file does not parse, so there is nothing to format"

// formatPassLimit bounds how many times one file is formatted while its text keeps changing. Every
// non-idempotent input measured on ahra settled on the second pass, so a third that still changes
// something is a printer defect to report, not a file to keep formatting.
const formatPassLimit = 3

// isUnparseable reports whether a format error was caused by source that does not parse.
//
// Matched on the message because the engine is behind an interface and its error types are its own.
// A false negative here costs a refusal where a skip was meant, which is the safe direction: it is
// reported and counted rather than hidden.
func isUnparseable(err error) bool {
	// The native printers mark their parse failures, so their wording never has to be guessed.
	if printing.IsSyntax(err) {
		return true
	}
	message := strings.ToLower(err.Error())
	for _, marker := range []string{"does not parse", "parse error", "syntaxerror", "unexpected token"} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

// extensionOf names a file's type for a skip reason, so the summary groups skips by what was
// skipped rather than listing one line per file.
func extensionOf(fileName string) string {
	lastDot := strings.LastIndex(fileName, ".")
	lastSlash := strings.LastIndexAny(fileName, "/\\")
	if lastDot <= lastSlash || lastDot < 0 {
		return "a file with no extension"
	}
	return fileName[lastDot:]
}

// configuredFormatter returns the formatter the pipeline runs, or nil when formatting was not asked
// for.
//
// A nil engine skips with "not requested" and that reaches the coverage line, so a run that did not
// format reports as one that did not format rather than as a perfectly formatted tree.
//
// The formatter is cohere's native printers. They replaced the goja Prettier fork when they matched it
// on every file of ahra, www-phi-health and api-phi-health, and when a whole-tree run of each engine
// over the same ahra snapshot wrote the same bytes. The fork stays in internal/format/prettier as the
// differential's oracle, and is not linked into this binary.
func configuredFormatter(enabled bool) (formatEngine, error) {
	if !enabled {
		return nil, nil
	}

	// Resolving, not one formatter built from formatoptions.Default: each file formats with the config
	// its own directory resolves to. Default is ahra's block, and running it over api-phi-health
	// rewrote every multi-line JSX opening tag, because that repository sets bracketSameLine.
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("finding where to resolve the format options from: %w", err)
	}
	engine, err := native.NewResolving(workingDirectory)
	if err != nil {
		// Not a nil engine. Nil already means "nobody asked for a formatter", and the coverage line
		// reports that as a deliberate absence. A formatter that was asked for and could not resolve its
		// options is a different fact, and collapsing the two would print a failure as a choice.
		return nil, fmt.Errorf("loading the formatter: %w", err)
	}
	return engine, nil
}
