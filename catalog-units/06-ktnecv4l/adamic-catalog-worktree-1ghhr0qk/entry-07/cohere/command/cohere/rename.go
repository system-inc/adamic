package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/rename"
	"github.com/system-inc/cohere/internal/types/program"
)

// renameVerbName is the subcommand that triggers a program-wide rename.
//
// A verb rather than a flag, and that separation is the design rather than a naming preference. Every
// other write in this tool is a fix: a rule found something wrong, the repair is scoped to the file
// the rule was looking at, and `--fix` runs by DEFAULT because a fix is meaning-preserving by
// construction. A rename is none of those things. Nothing is wrong with the code, the user is
// asserting an intention rather than accepting a repair, and the blast radius is wherever the symbol
// lives. Folding it into `--fix` would mean a flag whose default is on could rewrite a symbol across
// six files because of an argument somewhere else on the command line.
const renameVerbName = "rename"

// runRenameVerb handles `cohere rename <target> <newName>`.
//
// # Why the dry run is the default here when --fix is the default everywhere else
//
// The two defaults point opposite ways on purpose, and the asymmetry is the argument. `--fix`
// defaults to on because a fix is reactive and bounded: a rule already decided the edit was safe,
// the change is confined to a file the rule examined, and the engine's parse guard plus the type
// phase on the same run catch what is left. Nothing about running the gate is a statement of
// intent, so the tool doing the obvious repair is a service rather than a surprise.
//
// A rename inverts every one of those. It is imperative rather than reactive, so the tool has no
// independent judgment that the edit is correct — the user's assertion IS the justification. It is
// unbounded, reaching every file that mentions the symbol. And its characteristic failure is silent:
// `internal/rule/rule.go` records a shipped case where a rename updated a usage, missed its
// declaration, and produced a file that parsed while failing to compile. A rename that is subtly
// wrong is not caught by the parse guard, and the half-renamed tree is exactly the state this
// project exists to make impossible.
//
// So the safe default is inverted: the plan is computed and printed in full, and `--write` is the
// separate affirmative act that applies it. The cost is one extra command on a rename someone
// already knew they wanted. The alternative cost is a tree rewritten by a command whose preview
// nobody read, and those are not comparable.
func runRenameVerb(arguments []string) error {
	flags := flag.NewFlagSet(renameVerbName, flag.ContinueOnError)
	configFileName := flags.String("tsconfig", "tsconfig.json", "the tsconfig that defines the program, relative to --directory")
	directory := flags.String("directory", "", "the working directory paths resolve against (default: the process's own)")
	singleThreaded := flags.Bool("single-threaded", false, "use one checker instead of several")
	write := flags.Bool("write", false, "apply the rename; without it nothing is written")
	dryRun := flags.Bool("dry-run", false, "print the plan and write nothing (the default; accepted so it can be stated explicitly)")
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: cohere %s <file>:<line>:<column> <newName>\n", renameVerbName)
		fmt.Fprintf(os.Stderr, "       cohere %s <name> <newName>\n\n", renameVerbName)
		fmt.Fprintf(os.Stderr, "Renames a symbol everywhere the type graph says it is used.\n")
		fmt.Fprintf(os.Stderr, "The position form is primary; a bare name is refused when it is ambiguous.\n")
		fmt.Fprintf(os.Stderr, "Nothing is written without --write.\n\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	positional := flags.Args()
	if len(positional) != 2 {
		flags.Usage()
		return fmt.Errorf("expected a target and a new name, got %d arguments", len(positional))
	}
	target, newName := positional[0], positional[1]

	// Naming both is a contradiction the user should hear about rather than have silently resolved,
	// exactly as `--fix` and `--no-fix` are handled in the main pipeline.
	if *write && *dryRun {
		return fmt.Errorf("--write and --dry-run contradict each other: --dry-run writes nothing, --write applies the rename")
	}

	if !rename.IsValidIdentifier(newName) {
		return fmt.Errorf("%q cannot be written as an identifier, so the rename would produce source that does not parse", newName)
	}

	graph, err := program.Build(program.Options{
		ConfigFileName:   *configFileName,
		CurrentDirectory: *directory,
		SingleThreaded:   *singleThreaded,
	})
	if err != nil {
		// A program that fails to build is a loud failure and never an empty result, for the same
		// reason the main pipeline says so: an empty file list is indistinguishable from a tree with
		// nothing to rename.
		return fmt.Errorf("building the type graph: %w", err)
	}

	ctx := context.Background()

	position, isPosition := rename.ParsePosition(target)
	if isPosition {
		resolved, err := absolutePath(position.FileName, *directory)
		if err != nil {
			return err
		}
		// In the compiler's spelling, which is how the program names its files and what the anchor is
		// matched against. On Windows Clean leaves `C:\work\a.ts`, which no program file's
		// `C:/work/a.ts` equals, and every position was refused as not in the program.
		position.FileName = tspath.NormalizePath(resolved)
	} else {
		// The bare-name form is sugar, and it refuses rather than guesses. A bare name is not a
		// symbol: if three files each declare a local `err`, this has three valid answers and the
		// cursor that would disambiguate in an editor does not exist on a command line.
		candidates, symbols, err := rename.ResolveBareName(ctx, graph, target)
		if err != nil {
			return err
		}
		if len(symbols) == 0 {
			return fmt.Errorf(
				"no declaration named %q in this project — name a position instead if the symbol is declared elsewhere",
				target,
			)
		}
		if len(symbols) > 1 {
			rename.WriteCandidates(os.Stdout, target, candidates)
			return fmt.Errorf("%q is ambiguous, so nothing was renamed", target)
		}
		position = rename.Position{
			FileName: candidates[0].FileName,
			Line:     candidates[0].Line,
			Column:   candidates[0].Column,
		}
	}

	plan, err := rename.Build(ctx, graph, position, newName)
	if err != nil {
		return err
	}

	if len(plan.Refusals) > 0 {
		rename.Write(os.Stdout, plan, false)
		return fmt.Errorf("the rename was refused, so nothing was written")
	}

	if !*write {
		rename.Write(os.Stdout, plan, false)
		return nil
	}

	// Printed BEFORE applying, so the record of what was about to happen exists even if the write
	// fails partway. A summary printed only on success cannot explain a tree that was half-written.
	rename.Write(os.Stdout, plan, true)

	written, err := rename.Apply(plan)
	if err != nil {
		return fmt.Errorf("applying the rename: %w", err)
	}
	fmt.Printf("\nwrote %d files.\n", written)
	return nil
}

// absolutePath resolves a path the user typed against the directory the run is rooted at.
//
// The program's own file names are absolute, so a relative path from the command line would match
// nothing and produce a "not in the program" error that names the wrong cause.
func absolutePath(fileName string, directory string) (string, error) {
	if filepath.IsAbs(fileName) {
		return filepath.Clean(fileName), nil
	}
	base := directory
	if base == "" {
		working, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolving %s: %w", fileName, err)
		}
		base = working
	}
	return filepath.Clean(filepath.Join(base, fileName)), nil
}

// isRenameVerb reports whether the first argument names the rename verb.
//
// Matched positionally at the front, before flag.Parse runs, because the flag package stops at the
// first argument it does not recognize and a subcommand is exactly that. The dispatcher already
// forwards unknown arguments untouched, so the verb reaches this binary intact.
func isRenameVerb(arguments []string) bool {
	return len(arguments) > 0 && !strings.HasPrefix(arguments[0], "-") && arguments[0] == renameVerbName
}
