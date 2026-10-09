package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/types/program"
)

// stdinRequest is one editor save: `cohere --fix [--format] --stdin-filepath <path>` with the buffer on
// stdin.
type stdinRequest struct {
	Location         projectLocation
	WorkingDirectory string
	FilePath         string
	NoFormat         bool
	MaxPasses        int
	SingleThreaded   bool
}

// runStdin answers what `cohere --fix <path>` (fixes and format, or fixes alone with --no-format) would
// write for one file, computed from the buffer on stdin instead of the file on disk, and writes it to
// stdout. Nothing is written to disk.
//
// It exists for the editor's save hook (editors/vscode), which formats the buffer before the save
// lands so the document never reloads and undo keeps working. The answer has to be the gate's answer,
// byte for byte, or the editor and the gate disagree and every save leaves work for the next run. So
// every decision is the gate's own, reached through the same functions: the type graph is built with
// the buffer in place of the file (program.Options.Overlay), rules and options come from
// configureLint, the fixpoint and the format transform from edit.FixAndTransformText, which is
// FixAndTransformFile without the write, and the format scope from the same ignore layers.
//
// The outcomes, for the hook to act on:
//
//   - The new text on stdout, exit 0, whether or not anything changed.
//   - The buffer unchanged on stdout, exit 0, for a file cohere declines: no printer for its type,
//     outside the format scope, or source that does not parse. A save is not the place to report a
//     typo; the gate reports it.
//   - An error on stderr and nothing on stdout, exit 1, for a formatter or fixer that broke. The hook
//     shows it and saves the buffer as typed.
func runStdin(ctx context.Context, request stdinRequest, input io.Reader, output io.Writer) error {
	buffer, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("reading the buffer from stdin: %w", err)
	}
	text := string(buffer)

	path := request.FilePath
	if !filepath.IsAbs(path) {
		path = filepath.Join(request.WorkingDirectory, path)
	}
	path = filepath.Clean(path)

	// The gate's own decision: formatting by default, left out by --no-format.
	formatter, err := configuredFormatter(!request.NoFormat)
	if err != nil {
		return err
	}

	propose, held, err := stdinProposals(ctx, request, path, text)
	if err != nil {
		return err
	}
	transform, err := stdinTransform(formatter, request.Location, path, held)
	if err != nil {
		return err
	}

	result, err := edit.FixAndTransformText(path, text, propose, transform, request.MaxPasses)
	if err != nil {
		// FixText refuses a buffer that does not parse before anything is applied, and that is a
		// decline rather than a failure: the gate leaves such a file alone too.
		if parses, _ := edit.Parses(path, text); !parses {
			_, writeError := io.WriteString(output, text)
			return writeError
		}
		return err
	}
	for _, rejection := range result.Rejected {
		// A transform that errored is a broken formatter, which the gate reports and the save must not
		// hide. Every other refusal (overlaps, a pass that broke the parse) is the gate's normal
		// bookkeeping and leaves a result the gate would write.
		if strings.HasPrefix(rejection.Reason, edit.ReasonTransformFailed) {
			return fmt.Errorf("formatting %s: %s", path, rejection.Reason)
		}
	}
	_, err = io.WriteString(output, result.Text)
	return err
}

// stdinTransform is the gate's format transform for one named file: the same engine, the same ignore
// layers, the same scope a run naming this path would have. A file outside it comes back skipped, and so
// does an Adamic `.a` file the program does not hold (held is the program's `.a` files; see claimAdamic).
func stdinTransform(formatter formatEngine, location projectLocation, path string, held map[string]struct{}) (edit.Transform, error) {
	if formatter == nil {
		return nil, nil
	}
	scope := formatScope{
		FileNames:          []string{path},
		index:              map[string]struct{}{path: {}},
		Description:        "1 named path",
		RequestDescription: path,
	}
	enumeration, err := formatter.Enumerate(location.Root)
	if err != nil {
		// The gate withholds formatting when the walk fails; a save fails loudly instead, because a
		// hook that silently stopped formatting would read as files that need none.
		return nil, fmt.Errorf("enumerating what the formatter handles under %s: %w", location.Root, err)
	}
	scope, _ = scope.narrowToEnumeration(enumeration).claimAdamic(held)
	return scopedTransform(formatTransform(formatter), scope), nil
}

// stdinProposals is the gate's fix proposals for one file, computed from the buffer: the first pass
// from a walk of a type graph that holds the buffer, later passes by re-linting the rewritten text,
// exactly as applyProposedFixes does. A file outside the program has no rules, so nothing is proposed
// and only the format transform applies, which is what the gate does with a .md or a .css. held is the
// program's Adamic `.a` files, for the format transform to know whether this one is source.
func stdinProposals(ctx context.Context, request stdinRequest, path string, text string) (edit.Propose, map[string]struct{}, error) {
	none := func(string, string) ([]edit.Proposal, error) { return nil, nil }
	if !edit.TypeScriptParsable(path) {
		return none, nil, nil
	}

	graph, err := program.Build(program.Options{
		ConfigFileName:   request.Location.ConfigFileName,
		CurrentDirectory: request.Location.Root,
		SingleThreaded:   request.SingleThreaded,
		Overlay:          map[string]string{path: text},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("building the type graph: %w", err)
	}
	held := adamicHeld(graph)

	var sourceFile *ast.SourceFile
	normalized := tspath.NormalizePath(path)
	for _, projectFile := range graph.ProjectFiles() {
		if projectFile.FileName().AsString() == normalized {
			sourceFile = projectFile
			break
		}
	}
	if sourceFile == nil {
		return none, held, nil
	}

	if _, err := configureLint(graph, request.Location, nil); err != nil {
		return nil, nil, err
	}
	rules := registry.All()
	walk, err := graph.Walk(ctx, []*ast.SourceFile{sourceFile}, rules)
	if err != nil {
		return nil, nil, fmt.Errorf("collecting proposals: %w", err)
	}
	var firstPass []edit.Proposal
	for _, diagnostic := range walk.Diagnostics {
		if diagnostic.SourceFile == nil || len(diagnostic.Fixes) == 0 || diagnostic.SourceFile.FileName().AsString() != normalized {
			continue
		}
		for _, proposed := range diagnostic.Fixes {
			firstPass = append(firstPass, edit.Proposal{RuleName: diagnostic.RuleName, Fix: proposed})
		}
	}

	used := false
	return func(fileName string, current string) ([]edit.Proposal, error) {
		if !used {
			used = true
			return firstPass, nil
		}
		return proposalsForText(fileName, current, graph, rules)
	}, held, nil
}

// errStdinNeedsFix is the refusal for --stdin-filepath without --fix: the mode answers what --fix
// writes, and any other phase reports rather than writes, so there is nothing for it to answer.
var errStdinNeedsFix = errors.New("--stdin-filepath answers what --fix writes for one file: pass --fix, and --format to format")
