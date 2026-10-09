package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/types/sourcename"
)

// formatScope is which files the format phase considers, and how that set was decided.
//
// The scope began as a necessity: the goja formatter ran 83 to 100ms per file, so ahra's whole tree
// was minutes against a lint phase of 392ms. The native printers format that tree in about 10
// seconds, so the scope is now about relevance rather than survival. Actual churn on this tree is
// one file per commit and 35 across five, so scoping to changed files puts the common case in
// milliseconds and formats only what the caller touched.
//
// The set is named as well as computed, because named paths, the whole tree, and the files not on
// record as formatted are three different answers to "what is formatted" and a reader cannot tell
// which one they got from the file count alone. The last is the default; see format_record.go.
type formatScope struct {
	// FileNames is the absolute paths in scope, sorted. Nil means every file is in scope.
	FileNames []string

	// index is FileNames as a set, so membership is a map lookup rather than a scan.
	//
	// The transform asks about every project file, so a linear scan here would be quadratic across
	// the tree: 3,407 files against a few hundred changed ones is a million comparisons to answer a
	// question a map answers in one.
	index map[string]struct{}

	// Description says how the set was decided, in words that go straight into the coverage line.
	Description string

	// Everything is whether the scope is the whole tree rather than a subset.
	Everything bool

	// DependentCount is how many files were added because they import something named, so the
	// coverage line can say why more was checked than was asked for.
	DependentCount int

	// RequestDescription says what the caller asked for, as opposed to what was found. A named
	// directory enumerates every file under it, most of which the program never contained, so the
	// enumerated count is not a number anybody wants beside the count actually checked.
	RequestDescription string

	// recorded, when set, says a file outside the scope is out only because the format record vouches
	// for its bytes, and answers whether a given text is still those bytes. A fix that rewrites such a
	// file hands the transform text the record never saw, so it is formatted after all: the record
	// proves the bytes it hashed, not whatever a fixer makes of them.
	recorded func(fileName string, text string) bool

	// failure is why the scope holds nothing when the walk that should have drawn it failed. A run whose
	// verdict is formatting alone must not read that empty scope as a clean tree.
	failure error

	// adamic is the walk's held-back Adamic `.a` files (formatfiles.Enumeration.Adamic): admitted by the
	// ignore layers, and in no scope until the program claims them, since `.a` is also a static library's
	// extension (#6mhafvb). See claimAdamic.
	adamic []string

	// adamicJoins is whether a claimed `.a` file joins the scope: true for a scope drawn from the walk, the
	// default and `--format-all`, and false for named paths, which already hold the `.a` files named.
	adamicJoins bool

	// adamicUnrecorded, when set, is the default scope's own filter: of the `.a` files given, those whose
	// bytes are not on record as formatted.
	adamicUnrecorded func(files []string) []string
}

// claimAdamic is the scope once the program is built, and the `.a` files it claims (#6mhafvb). An Adamic file
// is in the scope only when the program holds it and the walk admitted it: a scope drawn from the walk takes
// each such file (the default scope, those not on record as formatted), and a named `.a` the program does not
// hold, or the walk did not admit, leaves the named set, so a libfoo.a is never handed to the TypeScript
// printer or the fix engine's parse guard. claimed is every `.a` file the program holds that the walk
// admitted, for the format record to keep their entries.
func (s formatScope) claimAdamic(held map[string]struct{}) (formatScope, []string) {
	admitted := make(map[string]struct{}, len(s.adamic))
	var claimed []string
	for _, fileName := range s.adamic {
		admitted[fileName] = struct{}{}
		if _, isHeld := held[fileName]; isHeld {
			claimed = append(claimed, fileName)
		}
	}
	if s.Everything {
		return s, claimed
	}

	fileNames := make([]string, 0, len(s.FileNames))
	for _, fileName := range s.FileNames {
		_, isHeld := held[fileName]
		_, isAdmitted := admitted[fileName]
		if sourcename.IsAdamic(fileName) && !(isHeld && isAdmitted) {
			continue
		}
		fileNames = append(fileNames, fileName)
	}
	joining := 0
	if s.adamicJoins {
		joiners := claimed
		if s.adamicUnrecorded != nil {
			joiners = s.adamicUnrecorded(claimed)
		}
		fileNames = append(fileNames, joiners...)
		joining = len(joiners)
		sort.Strings(fileNames)
	}
	if len(fileNames) == len(s.FileNames) && joining == 0 {
		return s, claimed
	}

	index := make(map[string]struct{}, len(fileNames))
	for _, fileName := range fileNames {
		index[fileName] = struct{}{}
	}
	s.FileNames, s.index = fileNames, index
	if joining > 0 {
		s.Description = fmt.Sprintf("%s · and %s the program holds", s.Description, counted(joining, "Adamic .a file", "Adamic .a files"))
	}
	return s, claimed
}

// narrowTo reports how many of the scope's files are in a given population, and re-describes the
// scope to say both numbers.
//
// The scope is the paths a caller named and the population is the files the program actually
// contains, and those differ: a named file can be excluded by the tsconfig or by ignorePatterns and
// never reach the format phase. Reporting only the named number leaves a reader doing arithmetic that
// does not work out, which is what happened when 72 changed files produced 3 formatter candidates and
// the output explained neither.
//
// Both numbers, in one clause, so nobody has to reconstruct the difference.
func (s formatScope) narrowTo(population map[string]struct{}) formatScope {
	if s.Everything {
		return s
	}

	inPopulation := 0
	for _, fileName := range s.FileNames {
		if _, present := population[fileName]; present {
			inPopulation++
		}
	}

	// An empty scope keeps the description its constructor built, because that one names the
	// directory it looked in and this one cannot: narrowing knows the population but not where the
	// question was posed.
	//
	// Overwriting it unconditionally is the bug this branch exists to prevent, and it shipped. The
	// empty-scope naming was correct in the constructor and discarded here two function calls later,
	// so the guard was present, tested, and dead. Its fixture only ever called the constructor
	// directly and never the composition, which is a check that verifies a fragment read as a check
	// on the behavior.
	if len(s.FileNames) == 0 {
		return s
	}

	// The scope says how it was decided; this function only adds how much of it is in the program.
	//
	// Four defects of one shape have now passed through these three lines, each a description that
	// was correct where it was written and wrong once another kind of scope started arriving. The
	// last was inferring `named` from a request description being present, which made `--changed`
	// report `7 named` for files nobody named, one heartbeat after that field was added to fix an
	// empty pair of parentheses.
	//
	// So the wording is no longer decided here. The description a constructor already wrote says how
	// its files were chosen, and this function appends to it rather than choosing between spellings
	// it cannot know the caller meant.
	s.Description = fmt.Sprintf("%s, %d of them in the program", s.Description, inPopulation)
	return s
}

// includes reports whether a file is in a narrow scope.
//
// It deliberately does not consult Everything. That check lives in scopedTransform, which returns
// the unwrapped transform for a whole-tree scope so the filter is never reached at all. Testing it
// in both places looked defensive and was worse than useless: it made the bypass unobservable, so a
// mutant that deleted the fast path survived every fixture because the redundant check produced the
// same answer. Two guards where one is load-bearing means neither can be shown to work.
//
// A whole-tree scope never reaches this method. If one does, it has no index and answers false,
// which is a wrong answer arriving loudly rather than a right answer arriving by accident.
func (s formatScope) includes(fileName string) bool {
	_, inScope := s.index[fileName]
	return inScope
}

// wholeTreeScope is every file, for a deliberate normalization pass or a continuous-integration
// check that wants the whole corpus.
//
// It is a flag rather than a default, and it is allowed to cost five minutes, because someone who
// asks for the whole tree has said what they want. The default costing five minutes is what gets a
// gate bypassed.
func wholeTreeScope() formatScope {
	return formatScope{Everything: true, Description: "every file (--format-all)"}
}

// scopedTransform narrows a transform to the files a scope admits.
//
// Files outside the scope are skipped with the scope's own description as the reason, so they are
// counted and named rather than quietly passed over. That is the difference between "formatted 12
// files" with no denominator and a line that says what was in scope and why, which is the same
// discipline the skip channel already carries for unhandled file types.
func scopedTransform(inner edit.Transform, scope formatScope) edit.Transform {
	if inner == nil {
		return nil
	}
	if scope.Everything {
		return inner
	}

	return func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		if !scope.includes(fileName) && (scope.recorded == nil || scope.recorded(fileName, text)) {
			return "", fmt.Errorf("%w: outside the format scope, %s", edit.ErrSkipped, scope.Description)
		}
		return inner(fileName, text, parsed)
	}
}

// narrowToEnumeration replaces the type-graph narrowing for the format phase.
//
// The two narrowings answer different questions and the ruling split them. A proposed fix comes from
// a rule that ran over the program, so its candidate must be in the program: `narrowTo` is correct
// for that and stays. A format candidate comes from the disk, and intersecting it with the type
// graph is what made css, markdown, json and yaml invisible, because a tsconfig enumerates
// TypeScript by construction.
//
// The formatter could only speak TypeScript until the goja engine landed, so the type graph was an
// adequate universe by accident. That constraint was incidental and the pipeline mistook it for the
// architecture.
//
// The description states what the walk enumerated rather than what it intended: the root, the count
// before and after the engine's filter, and every declined extension by name. A file the engine
// cannot handle has to be a named zero rather than an absence, which is the one place in this
// pipeline where a file could previously be silently missing.
func (s formatScope) narrowToEnumeration(enumeration formatfiles.Enumeration) formatScope {
	if s.Everything {
		// A whole-tree scope becomes the enumeration itself: every file the walk found and the engine
		// handles, rather than every file in the type graph.
		index := make(map[string]struct{}, len(enumeration.Files))
		for _, fileName := range enumeration.Files {
			index[fileName] = struct{}{}
		}
		return formatScope{
			FileNames:   enumeration.Files,
			index:       index,
			Description: describeEnumeration(enumeration, len(enumeration.Files)),
			adamic:      enumeration.Adamic,
			adamicJoins: true,
		}
	}

	// A narrow scope keeps its own file set, which is the named set, and reports how much of it the
	// formatter will actually see. A named `.a` file stays in it only if the walk admitted it and the
	// program holds it (claimAdamic).
	s.adamic = enumeration.Adamic
	handled := make(map[string]struct{}, len(enumeration.Files))
	for _, fileName := range enumeration.Files {
		handled[fileName] = struct{}{}
	}

	formattable := 0
	for _, fileName := range s.FileNames {
		if _, present := handled[fileName]; present {
			formattable++
		}
	}

	// An empty scope keeps the description that names where it looked, for the same reason narrowTo
	// keeps it: this function knows the population but not where the question was posed.
	//
	// The Everything guard is not redundant with the branch above. A whole-tree scope also has zero
	// FileNames, so without it a whole-tree scope that reached here would take this early return and
	// come back with Everything still set, and scopedTransform would then bypass filtering entirely
	// and offer the engine files the walk never found. That is a real hole rather than a hypothetical:
	// it is what a mutation deleting the branch above produced, and no fixture could see it because
	// includes is never consulted for a scope that claims to be everything.
	if !s.Everything && len(s.FileNames) == 0 {
		return s
	}
	if s.Everything {
		// Unreachable through the branch above, and stated rather than assumed. A whole-tree scope that
		// arrives here has lost its enumeration, and answering with an empty narrow scope is a loud
		// wrong answer rather than a silent permissive one.
		return formatScope{
			index:       map[string]struct{}{},
			Description: fmt.Sprintf("nothing (a whole-tree scope reached narrowing without its enumeration, walked %s)", enumeration.Root),
		}
	}

	s.Description = fmt.Sprintf(
		"%s, %d the formatter handles · %s",
		s.Description, formattable, describeEnumeration(enumeration, formattable),
	)
	return s
}

// describeEnumeration renders what the walk actually did.
//
// Stated rather than summarized, because the requirement is that this line cannot become correct by
// accident. When the universe changes, this string has to be rewritten deliberately, so it names the
// root, the ignore layers that removed anything, the nested repositories it refused, and the
// declined extensions.
func describeEnumeration(enumeration formatfiles.Enumeration, formattable int) string {
	// One format string rather than two. An earlier version branched between a form that named the
	// formattable count and one that did not, and the branch made the root unobservable: a mutation
	// deleting the root from one string left the other correct, so no fixture could see it. Two
	// spellings of one sentence means neither can be shown to work.
	description := fmt.Sprintf(
		"walked %s, %d files, %d the formatter handles",
		enumeration.Root, enumeration.Walked, formattable,
	)

	// Ignore layers are named in the order they ran, and only when they removed something. A layer
	// that removed nothing is either correct or broken, and printing a zero for it every run trains
	// the reader to stop looking.
	layers := make([]string, 0, len(enumeration.IgnoredByLayer))
	for layer := range enumeration.IgnoredByLayer {
		if enumeration.IgnoredByLayer[layer] > 0 {
			layers = append(layers, fmt.Sprintf("%d by %s", enumeration.IgnoredByLayer[layer], layer))
		}
	}
	sort.Strings(layers)
	if len(layers) > 0 {
		description += ", ignored " + strings.Join(layers, ", ")
	}

	// Nested repositories are named rather than counted. "We skipped a repo" is a fact somebody may
	// want to argue with, and a number gives them nothing to argue with.
	if len(enumeration.NestedRepositories) > 0 {
		names := append([]string(nil), enumeration.NestedRepositories...)
		sort.Strings(names)
		description += ", skipped nested repositories " + strings.Join(names, ", ")
	}

	// A symbolic link is never followed or read (see Enumeration.SymbolicLinks), and says so.
	if enumeration.SymbolicLinks > 0 {
		description += fmt.Sprintf(", skipped %d symbolic links, which are never followed", enumeration.SymbolicLinks)
	}

	// Declines are the requirement this whole change exists for: a file the engine cannot handle is a
	// named zero rather than an absence.
	if len(enumeration.DeclinedExtensions) > 0 {
		extensions := make([]string, 0, len(enumeration.DeclinedExtensions))
		for extension := range enumeration.DeclinedExtensions {
			extensions = append(extensions, extension)
		}
		sort.Strings(extensions)

		declined := make([]string, 0, len(extensions))
		for _, extension := range extensions {
			declined = append(declined, fmt.Sprintf("%d %s", enumeration.DeclinedExtensions[extension], extension))
		}
		description += ", declined " + strings.Join(declined, ", ")
	}

	return description
}

// formatCandidates is the files the format phase should visit.
//
// A whole-tree scope hands back everything the enumeration found; a narrow scope hands back its own
// set, the named paths or the files not on record as formatted. Either way these are candidates rather than a promise: a file here still passes
// through the transform's own Handles check and can still come back declined, which is what keeps a
// skip visible instead of turning it into an absence.
//
// A scope with no formatter configured yields nothing, so a fix-only run visits exactly the files
// that had proposals and costs what it did before the format phase existed.
func (s formatScope) formatCandidates() []string {
	return s.FileNames
}

// namedPathsScope resolves the paths a caller named on the command line.
//
// A positional path was parsed and discarded before this existed, so `cohere --lint OneFile.ts`
// checked all 3,542 files and reported every finding in the tree. Measured: 3.042s and 5,201
// findings for one named file, against 3.003s and 5,201 for the whole tree. Same cost, same output,
// and no way to tell from either that the argument had been ignored.
//
// A named directory expands to the files under it rather than matching by prefix, so membership
// stays one map lookup for every phase downstream. `.` is the whole tree and says so, which keeps
// `cohere --lint .` meaning what it has always meant instead of quietly becoming a subset of one
// directory entry.
//
// Resolution is against the working directory the caller gave, not the process's own, because
// `--directory` already redefines what a relative path means everywhere else in this binary and a
// second answer here would be a bug nobody could see from the output.
//
// The whole-tree answer belongs to the project root and not to the working directory, and the two
// stopped being the same directory once the root could be found by walking up. `cohere .` from
// `modules/tasks` means that directory, and reading it as the whole project would turn the
// narrowest thing a caller can type into the widest run there is.
func namedPathsScope(workingDirectory string, root string, names []string) (formatScope, error) {
	// `--directory` defaults to empty, meaning the process's own. Resolving against empty leaves a
	// relative path, which never matches a source file name because those are absolute, so every
	// named path fell out of scope and the run checked nothing while reporting success. That is the
	// green-over-zero-files failure this binary exists to have stopped making, and it was one
	// `filepath.Join` away from shipping.
	if workingDirectory == "" {
		current, err := os.Getwd()
		if err != nil {
			return formatScope{}, fmt.Errorf("resolving the working directory: %w", err)
		}
		workingDirectory = current
	}
	if root == "" {
		root = workingDirectory
	}

	absolute := make([]string, 0, len(names))

	for _, name := range names {
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(workingDirectory, name)
		}
		path = filepath.Clean(path)

		information, err := os.Stat(path)
		if err != nil {
			// A path that does not exist is a loud failure rather than an empty scope. Silently
			// checking nothing because of a typo is the green-over-zero-files defect, and it is the
			// one this binary exists to have stopped making.
			return formatScope{}, fmt.Errorf("resolving %s: %w", name, err)
		}

		if !information.IsDir() {
			absolute = append(absolute, path)
			continue
		}

		if path == filepath.Clean(root) {
			return wholeTreeScope(), nil
		}

		err = filepath.WalkDir(path, func(entry string, directoryEntry fs.DirEntry, walkError error) error {
			if walkError != nil {
				return walkError
			}
			if directoryEntry.IsDir() {
				return nil
			}
			absolute = append(absolute, filepath.Clean(entry))
			return nil
		})
		if err != nil {
			return formatScope{}, fmt.Errorf("walking %s: %w", name, err)
		}
	}

	sort.Strings(absolute)

	index := make(map[string]struct{}, len(absolute))
	for _, path := range absolute {
		index[path] = struct{}{}
	}

	description := fmt.Sprintf("%d named path%s", len(absolute), plural(len(absolute)))
	if len(absolute) == 0 {
		description = fmt.Sprintf("0 files under the named paths in %s", workingDirectory)
	}

	request := strings.Join(names, ", ")
	if len(names) > 3 {
		request = fmt.Sprintf("%s and %d more", strings.Join(names[:3], ", "), len(names)-3)
	}

	return formatScope{
		FileNames:          absolute,
		index:              index,
		Description:        description,
		RequestDescription: request,
	}, nil
}

// plural is the one-character suffix that keeps a count reading as English.
func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
