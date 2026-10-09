package rule_testing

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// defaultTsConfig is a minimal, self-contained project.
//
// `types: []` is load-bearing: without it the program tries to resolve @types packages from disk
// and a fixture picks up whatever happens to be installed near the test's temp directory, which
// makes the same fixture pass or fail depending on the machine. lib ES2022 gives Promise and the
// other built-ins a type-aware rule needs to reason about, without pulling in the DOM.
const defaultTsConfig = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"moduleDetection": "force",
		"types": []
	},
	"sourceExtensions": [".a"],
	"include": ["**/*.ts", "**/*.tsx", "**/*.a"]
}`

// An Adamic `.a` fixture is TypeScript source under its own name (#6mhafvb), so the fixture program
// lists ".a" in "sourceExtensions" and includes it, and a rule's fixture can show `.a` reports as `.ts`
// does (#kwt1htp).

// The include pattern recurses, and it did not until a fixture needed a nested path.
//
// `*.ts` matches only the temp root, so a multi-file fixture laying its files out on realistic
// paths silently produced a program with no inputs. That surfaced as `TS18003` from the config
// reader rather than as a wrong finding, which is the good direction, but it also meant any rule
// deciding on directory structure could only be fixtured against flat filenames it would never see
// in the tree. `boundary-no-project-theme-value` gates on `/libraries/structure/source/` and cannot
// be proven at all without real paths.

// FixtureText is the bytes a typed fixture is written as: the source with leading and trailing ASCII
// whitespace removed, and one newline. A test that slices the file by the rule's offsets reads this.
//
// ASCII only, and that is load-bearing. `strings.TrimSpace` also strips Unicode spaces, U+2028 and
// U+2029, vertical tab and form feed, which are exactly what `no-irregular-whitespace` reports, so a
// fixture of `foo\u2028` was written as `foo` and the rule was blamed for missing it: 30 of ESLint's
// corpus rows read as missing for that reason alone (#jjfa7qb). Indentation and blank lines around a
// backquoted fixture are all this trim exists for.
func FixtureText(contents string) string {
	return strings.Trim(contents, " \t\r\n") + "\n"
}

// RunTyped runs a rule against a real type graph, so a rule that asks the checker questions can
// actually be proven.
//
// Run and RunWithOptions parse a file and hand the rule a Context whose Program and TypeChecker are
// nil. That is right for the syntactic majority and wrong in a specific, dangerous way for a
// type-aware rule: every fixture exercises the nil-checker path instead of the logic under test,
// and the suite goes green having proven nothing. It is the vacuous-probe shape one level up, where
// the probe runs perfectly against something that is not the question.
//
// So a rule that reads ctx.TypeChecker must be tested through here. The cost is real, roughly a
// second per call to build the program, which is why this is a separate entry point rather than the
// default: a microsecond parse is the right harness for the rules that only need one.
//
// Identified by @system_cohere_lint_rules, which stopped porting a 1,208-line rule on discovering
// it could be written but not proven, rather than writing fixtures that would have passed
// vacuously.
func RunTyped(t *testing.T, subject rule.Rule, fileName string, sourceText string) Result {
	t.Helper()
	return RunTypedWithOptions(t, subject, fileName, sourceText, nil)
}

// RunTypedWithOptions is RunTyped for a rule that reads configuration.
func RunTypedWithOptions(t *testing.T, subject rule.Rule, fileName string, sourceText string, options any) Result {
	t.Helper()
	return RunTypedFilesWithOptions(t, subject, map[string]string{fileName: sourceText}, fileName, options)
}

// RunTypedVerbatimWithOptions is RunTypedWithOptions writing the source exactly as given, with no trim.
//
// For a source whose offsets are someone else's: an ESLint corpus row records where ESLint reported
// in the row's own bytes, and FixtureText's trim moved a finding at the start or end of the file, in
// both directions (#jjfa7qb). Hand-written fixtures keep the trim, which is what lets them be indented
// backquoted literals.
func RunTypedVerbatimWithOptions(t *testing.T, subject rule.Rule, fileName string, sourceText string, options any) Result {
	t.Helper()
	files := map[string]string{fileName: sourceText}
	return runTypedFiles(t, subject, files, fileName, fixedOptions(options), nil, true)
}

// RunTypedFiles runs a rule against one file of a multi-file program.
//
// Several type questions can only be asked across a module boundary: whether an imported symbol is
// a class, what a re-exported alias resolves to, whether a factory in another file returns a
// Thenable. A single-file harness cannot pose those questions at all, so a rule that gets them
// wrong would pass every fixture.
//
// subjectFileName selects which file the rule is run against; the rest are there to be resolved.
func RunTypedFiles(t *testing.T, subject rule.Rule, files map[string]string, subjectFileName string) Result {
	t.Helper()
	return RunTypedFilesWithOptions(t, subject, files, subjectFileName, nil)
}

// RunTypedFilesWithSetup is RunTypedFiles with a hook that runs against the fixture directory.
//
// The hook is called after the fixture files are written and before the program is built, and it is
// handed the temp directory they were written into. It exists for the one thing a map of file
// contents cannot express: something on disk that is not a file with text in it.
//
// The case that motivated it is a symlink. A rule reading a Tailwind design system finds the
// installed package by walking UP from the stylesheet looking for `node_modules/tailwindcss`, and a
// fixture in `t.TempDir()` has nothing above it, so the rule declines before it reads a line of the
// fixture's CSS. Writing an absolute import into the stylesheet does not help, because the rule never
// resolves the import itself. Linking a real package into the fixture does, and it makes the rule
// take the same path it takes on a repository rather than one arranged for the test.
func RunTypedFilesWithSetup(
	t *testing.T,
	subject rule.Rule,
	files map[string]string,
	subjectFileName string,
	setup func(directory string),
) Result {
	t.Helper()
	return runTypedFiles(t, subject, files, subjectFileName, fixedOptions(nil), setup, false)
}

// RunTypedFilesWithSetupAndOptions is RunTypedFilesWithSetup for a rule that also reads
// configuration.
//
// Both axes at once, because a rule can need both and the two existing entry points each drop one.
// `no-unknown-classes` is the case: it needs the symlinked package RunTypedFilesWithSetup exists for
// AND its `ignore` option, and without this its ignore-list fixtures would have to run through a
// harness that leaves the Program nil, where the rule declines and every assertion passes vacuously.
func RunTypedFilesWithSetupAndOptions(
	t *testing.T,
	subject rule.Rule,
	files map[string]string,
	subjectFileName string,
	options any,
	setup func(directory string),
) Result {
	t.Helper()
	return runTypedFiles(t, subject, files, subjectFileName, fixedOptions(options), setup, false)
}

// RunTypedFilesWithOptionsFor is RunTypedFilesWithSetup for a rule whose options name the fixture's
// own directory, which no caller knows until the files are written.
//
// A repository root is the case: `import-require-path-alias` makes every path relative to one, and a
// fixture that reads the aliases out of its own tsconfig needs that root to be the directory the
// tsconfig sits in. Passing a fixed root instead would put every fixture file outside it, and the
// rule would decline each one, so the clean cases would pass for the wrong reason.
//
// The fixture is built the setup way, uncached, because a fixture of this kind brings its own
// tsconfig and the cached build writes the default one over it.
func RunTypedFilesWithOptionsFor(
	t *testing.T,
	subject rule.Rule,
	files map[string]string,
	subjectFileName string,
	optionsFor func(directory string) any,
) Result {
	t.Helper()
	return runTypedFiles(t, subject, files, subjectFileName, optionsFor, func(string) {}, false)
}

// fixedOptions is an options value that does not depend on where the fixture was written.
func fixedOptions(options any) func(directory string) any {
	return func(string) any {
		return options
	}
}

// RunTypedFilesWithOptions is the full form the others delegate to.
func RunTypedFilesWithOptions(
	t *testing.T,
	subject rule.Rule,
	files map[string]string,
	subjectFileName string,
	options any,
) Result {
	t.Helper()
	return runTypedFiles(t, subject, files, subjectFileName, fixedOptions(options), nil, false)
}

// runTypedFiles is the body both public forms share.
func runTypedFiles(
	t *testing.T,
	subject rule.Rule,
	files map[string]string,
	subjectFileName string,
	optionsFor func(directory string) any,
	setup func(directory string),
	verbatim bool,
) Result {
	t.Helper()

	if len(files) == 0 {
		t.Fatal("no fixture files were given, so this test would prove nothing")
	}
	if _, isSubject := files[subjectFileName]; !isSubject {
		t.Fatalf("the subject file %q is not among the fixture files %v", subjectFileName, keysOf(files))
	}

	// A fixture with a setup hook is built the old way, uncached.
	//
	// The hook is handed the directory and may do anything to it, including linking a real package in
	// from outside. Two callers passing identical file maps can therefore still want different
	// programs, so the file bytes are no longer a complete cache key and sharing the build would be
	// wrong rather than merely conservative.
	var graph *program.Graph
	var directory string
	if setup != nil {
		directory = t.TempDir()
		for name, contents := range files {
			path := filepath.Join(directory, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("creating the fixture directory for %s: %v", name, err)
			}
			if err := os.WriteFile(path, []byte(FixtureText(contents)), 0o644); err != nil {
				t.Fatalf("writing the fixture %s: %v", name, err)
			}
		}

		// After the files exist and before the program is built, so a hook can add something to the
		// directory that the program or a rule will then find.
		setup(directory)

		// A hook that wrote its own tsconfig keeps it. A rule whose verdict depends on a compiler
		// option the default does not set, `emitDecoratorMetadata` for one, has no other way to reach
		// a program built under that option.
		configPath := filepath.Join(directory, "tsconfig.json")
		if _, statError := os.Stat(configPath); os.IsNotExist(statError) {
			if err := os.WriteFile(configPath, []byte(defaultTsConfig), 0o644); err != nil {
				t.Fatalf("writing the tsconfig: %v", err)
			}
		}

		built, err := program.Build(program.Options{
			ConfigFileName:   configPath,
			CurrentDirectory: directory,
			LibraryParses:    libraryParses,
			// One checker rather than several: a fixture is one file, and the parallel path adds
			// scheduling nondeterminism to a test whose whole value is being deterministic.
			SingleThreaded: true,
		})
		if err != nil {
			t.Fatalf("building the type graph: %v", err)
		}
		graph = built
	} else {
		built, cachedDirectory, err := buildCachedProgram(files, subjectFileName, verbatim)
		if err != nil {
			t.Fatalf("%v", err)
		}
		graph, directory = built, cachedDirectory
	}

	projectFiles := graph.ProjectFiles()
	if len(projectFiles) == 0 {
		// An empty file set makes every assertion downstream pass vacuously, which is the failure
		// this harness exists to prevent rather than reproduce.
		t.Fatal("the program contains no project files, so this test would prove nothing")
	}

	subjectPath := filepath.ToSlash(filepath.Join(directory, subjectFileName))
	var sourceFile = findSourceFile(projectFiles, subjectPath)
	if sourceFile == nil {
		t.Fatalf("the subject file %s is not in the built program", subjectFileName)
	}

	fileChecker, release := graph.CheckerForFile(context.Background(), sourceFile)
	defer release()

	if fileChecker == nil {
		// The entire point of this harness. A nil checker here means a type-aware rule would take
		// its nil path and the fixture would prove nothing, silently.
		t.Fatal("the program produced no type checker, so a type-aware rule could not be proven")
	}

	var diagnostics []rule.Diagnostic
	ruleContext := rule.Context{
		SourceFile:  sourceFile,
		Program:     rule.ViewProgram(graph.Program, sourceFile, subject),
		TypeChecker: fileChecker,
		FileCache:   rule.NewFileCache(),
		Report: func(diagnostic rule.Diagnostic) {
			diagnostic.RuleName = subject.Name
			if diagnostic.SourceFile == nil {
				diagnostic.SourceFile = sourceFile
			}
			diagnostics = append(diagnostics, diagnostic)
		},
	}

	options := optionsFor(directory)
	listeners := subject.Run(ruleContext, options)
	if listeners != nil {
		walk(sourceFile.AsNode(), listeners)
	}

	expectEachFixParses(t, sourceFile, diagnostics)
	return Result{
		Diagnostics: diagnostics,
		SourceFile:  sourceFile,
		capture:     newCapturedRun(subject, subjectFileName, len(files)-1, options),
	}
}

// findSourceFile locates the subject file in the built program by normalized path.
//
// Matched on the path rather than by index, because the compiler orders its file list by resolution
// rather than by the order files were written, and a fixture that silently ran against the wrong
// file would still report findings and still look correct.
func findSourceFile(files []*ast.SourceFile, wantPath string) *ast.SourceFile {
	for _, candidate := range files {
		if filepath.ToSlash(candidate.FileName().AsString()) == wantPath {
			return candidate
		}
	}
	return nil
}

func keysOf(files map[string]string) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
