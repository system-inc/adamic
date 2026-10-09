package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/format/formatfiles"
)

// A scoped transform must format what is in scope and skip what is not, and the skip must carry the
// scope's own description.
//
// The skip channel rather than silence is the whole point: a file passed over without a word is
// indistinguishable from a file that was already correctly formatted, and a run that formatted 12 of
// 3,407 files must not print like a run that formatted everything.
func TestAScopedTransformFormatsOnlyWhatIsInScope(t *testing.T) {
	t.Parallel()
	inner := func(_ string, text string, _ *ast.SourceFile) (string, error) {
		return strings.ToUpper(text), nil
	}
	scope := formatScope{
		FileNames:   []string{"/repo/in.ts"},
		index:       map[string]struct{}{"/repo/in.ts": {}},
		Description: "1 named paths",
	}

	transform := scopedTransform(inner, scope)

	formatted, err := transform("/repo/in.ts", "hello\n", nil)
	if err != nil {
		t.Fatalf("a file in scope was not formatted: %v", err)
	}
	if formatted != "HELLO\n" {
		t.Fatalf("the inner transform did not run: %q", formatted)
	}

	_, err = transform("/repo/out.ts", "hello\n", nil)
	if !errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("a file outside scope should skip, got %v", err)
	}
	if !strings.Contains(err.Error(), "outside the format scope") {
		t.Fatalf("the skip did not say it was a scope decision: %v", err)
	}
	if !strings.Contains(err.Error(), scope.Description) {
		t.Fatalf("the skip did not carry the scope description: %v", err)
	}
}

// A whole-tree scope must not wrap the transform at all, so the flag costs nothing beyond the
// formatting it asks for.
func TestAWholeTreeScopeDoesNotFilter(t *testing.T) {
	t.Parallel()
	transform := scopedTransform(func(_ string, text string, _ *ast.SourceFile) (string, error) {
		return text + "!", nil
	}, wholeTreeScope())

	for _, fileName := range []string{"/anywhere/a.ts", "/elsewhere/b.tsx", "/deep/c.md"} {
		formatted, err := transform(fileName, "x", nil)
		if err != nil {
			t.Fatalf("a whole-tree scope skipped %s: %v", fileName, err)
		}
		if formatted != "x!" {
			t.Fatalf("the inner transform did not run for %s", fileName)
		}
	}
}

// A nil transform stays nil through scoping. Wrapping nothing would produce a transform that
// reports skips for a formatter that does not exist, which is a different and less useful statement
// than "not requested".
func TestScopingANilTransformStaysNil(t *testing.T) {
	t.Parallel()
	if scopedTransform(nil, wholeTreeScope()) != nil {
		t.Fatalf("a nil transform became non-nil")
	}
	if scopedTransform(nil, formatScope{Description: "some scope"}) != nil {
		t.Fatalf("a nil transform became non-nil under a narrow scope")
	}
}

// An empty scope skips everything, which is what a failed walk must produce.
//
// The alternative, falling back to the whole tree, would turn a failed walk into a five-minute
// formatting run nobody asked for. Withholding formatting and saying why is the safe
// direction.
func TestAnEmptyScopeSkipsEverythingAndSaysWhy(t *testing.T) {
	t.Parallel()
	transform := scopedTransform(func(_ string, text string, _ *ast.SourceFile) (string, error) {
		t.Fatalf("the inner transform ran under an empty scope")
		return text, nil
	}, formatScope{Description: "nothing (could not enumerate the tree: the walk failed)"})

	_, err := transform("/repo/a.ts", "x", nil)
	if !errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("expected a skip, got %v", err)
	}
	if !strings.Contains(err.Error(), "could not enumerate the tree") {
		t.Fatalf("the reason did not survive: %v", err)
	}
}

// The scope description must distinguish the three cases a reader needs to tell apart, since it is
// the only thing in the output that says what was formatted and why.
func TestScopeDescriptionsAreDistinct(t *testing.T) {
	t.Parallel()
	whole := wholeTreeScope().Description
	narrow := formatScope{Description: "12 named paths"}.Description
	failed := formatScope{Description: "nothing (could not enumerate the tree: the walk failed)"}.Description

	if whole == narrow || narrow == failed || whole == failed {
		t.Fatalf("two scopes describe themselves identically: %q %q %q", whole, narrow, failed)
	}
	if !strings.Contains(whole, "--format-all") {
		t.Fatalf("the whole-tree description does not name the flag that produced it: %q", whole)
	}
}

// The whole-tree fast path must actually bypass the filter, not merely produce the same answer.
//
// TestAWholeTreeScopeDoesNotFilter passes either way, because a scope with Everything set also makes
// includes return true, so removing the bypass changes nothing it can see. Two guards and one
// assertion between them. This one watches whether the filter was consulted at all, by giving the
// whole-tree scope an index that would exclude the file if anything looked at it.
func TestAWholeTreeScopeBypassesTheFilterEntirely(t *testing.T) {
	t.Parallel()
	inner := func(_ string, text string, _ *ast.SourceFile) (string, error) { return text + "!", nil }

	// Everything is set, and the index deliberately does not contain the file. If the returned
	// transform consults the index at all, this skips.
	scope := wholeTreeScope()
	scope.index = map[string]struct{}{"/some/other/file.ts": {}}

	formatted, err := scopedTransform(inner, scope)("/not/in/the/index.ts", "x", nil)
	if err != nil {
		t.Fatalf("a whole-tree scope consulted its index: %v", err)
	}
	if formatted != "x!" {
		t.Fatalf("the inner transform did not run: %q", formatted)
	}
}

// The scope must state both counts: what was named and how many of those the program contains.
//
// The two differ because a path knows nothing about the tsconfig, so a named file can be excluded and
// never reach the format phase. Reporting only the named number left a reader doing arithmetic that
// does not work out: 72 changed files producing 3 formatter candidates, with the output explaining
// neither. That is the same shape as the summary-line defect from earlier tonight, where every
// number was right and the composition was misleading.
func TestTheScopeStatesBothCounts(t *testing.T) {
	t.Parallel()
	scope := formatScope{
		FileNames: []string{"/repo/a.ts", "/repo/b.ts", "/repo/excluded.ts"},
		index: map[string]struct{}{
			"/repo/a.ts": {}, "/repo/b.ts": {}, "/repo/excluded.ts": {},
		},
		Description: "3 named paths",
	}

	narrowed := scope.narrowTo(map[string]struct{}{
		"/repo/a.ts": {}, "/repo/b.ts": {}, "/repo/never-changed.ts": {},
	})

	want := "3 named paths, 2 of them in the program"
	if narrowed.Description != want {
		t.Fatalf("the scope description does not state both counts:\n  want %q\n  got  %q", want, narrowed.Description)
	}

	// Narrowing describes; it must not silently shrink what is in scope, or the two numbers would
	// describe the same population and the distinction would be lost.
	if len(narrowed.FileNames) != 3 {
		t.Fatalf("narrowing changed the scope rather than describing it: %v", narrowed.FileNames)
	}
}

// A whole-tree scope has no git count to reconcile, so narrowing must leave it alone.
func TestNarrowingAWholeTreeScopeChangesNothing(t *testing.T) {
	t.Parallel()
	before := wholeTreeScope()
	after := before.narrowTo(map[string]struct{}{"/repo/a.ts": {}})

	if after.Description != before.Description {
		t.Fatalf("a whole-tree scope was re-described: %q", after.Description)
	}
	if !after.Everything {
		t.Fatalf("narrowing turned a whole-tree scope into a narrow one")
	}
}

// The empty-scope naming must survive narrowing, which is where it died.
//
// The constructor names the directory when it found nothing, and `narrowTo` rebuilt the description
// unconditionally two calls later, discarding it. The guard was present, its fixture passed, and the
// behavior was gone: the fixture called the constructor directly and never the composition the
// pipeline actually runs.
//
// That is the same defect the guard exists to catch, one layer up. A check that verifies a fragment
// gets read as a check on the behavior, and the gap between them is where this lived. Caught by
// running the real binary and reading its output, not by any assertion.
func TestTheEmptyScopeNamingSurvivesNarrowing(t *testing.T) {
	t.Parallel()
	empty := formatScope{
		FileNames:   nil,
		index:       map[string]struct{}{},
		Description: "0 files under the named paths in /some/where",
	}

	narrowed := empty.narrowTo(map[string]struct{}{"/repo/a.ts": {}})

	if !strings.Contains(narrowed.Description, "/some/where") {
		t.Fatalf("narrowing discarded the directory an empty scope had named: %q", narrowed.Description)
	}
}

// enumerationOf builds an Enumeration the way the real walk would report one.
func enumerationOf(root string, files []string, walked int, declined map[string]int, nested []string, ignored map[string]int) formatfiles.Enumeration {
	return formatfiles.Enumeration{
		Root:               root,
		Walked:             walked,
		Files:              files,
		DeclinedExtensions: declined,
		NestedRepositories: nested,
		IgnoredByLayer:     ignored,
	}
}

// A whole-tree scope must become the enumeration, not the type graph.
//
// This is the ruling's core: the format phase's universe is what the walk found and the engine
// handles. Intersecting with the type graph is what made css, markdown, json and yaml invisible,
// because a tsconfig enumerates TypeScript by construction.
func TestAWholeTreeScopeBecomesTheEnumeration(t *testing.T) {
	t.Parallel()
	files := []string{"/repo/a.ts", "/repo/b.css", "/repo/c.md", "/repo/d.json", "/repo/e.yaml"}
	enumeration := enumerationOf("/repo", files, 12, map[string]int{".rb": 3, "(none)": 4}, nil, nil)

	scope := wholeTreeScope().narrowToEnumeration(enumeration)

	// Every language the engine handles is in scope, which is the thing that was broken.
	for _, fileName := range files {
		if !scope.includes(fileName) {
			t.Fatalf("%s was enumerated and handled but is not in scope: %v", fileName, scope.FileNames)
		}
	}
	if scope.Everything {
		t.Fatalf("the scope stayed whole-tree instead of becoming the enumeration")
	}

	// The file set is the enumeration's, exactly. Falling through to the narrow branch would keep
	// whatever the whole-tree scope had, which is nothing, and every assertion above would still pass
	// because a whole-tree scope's includes returns true for everything it is asked about.
	if len(scope.FileNames) != len(files) {
		t.Fatalf("the scope did not take the enumeration's file set: %v", scope.FileNames)
	}

	// And a file the walk did not find must not be in scope. This is the assertion that catches a
	// whole-tree scope that ignored the enumeration: it would answer true here.
	if scope.includes("/repo/never-walked.ts") {
		t.Fatalf("a file the enumeration never found is in scope, so the enumeration was ignored")
	}
}

// A file the engine cannot handle must be a named zero, not an absence.
//
// This is the requirement the ruling attached and the one place in this pipeline where a file could
// previously go missing without a word: a `.json` in a TypeScript project was never a candidate, so
// the coverage line could not even report it as skipped.
func TestDeclinedExtensionsAreNamedRatherThanOmitted(t *testing.T) {
	t.Parallel()
	enumeration := enumerationOf("/repo", []string{"/repo/a.ts"}, 61, map[string]int{".rb": 3, ".txt": 57, "(none)": 1}, nil, nil)

	description := wholeTreeScope().narrowToEnumeration(enumeration).Description

	for _, want := range []string{"3 .rb", "57 .txt", "1 (none)"} {
		if !strings.Contains(description, want) {
			t.Fatalf("a declined extension is missing from the coverage line: want %q in %q", want, description)
		}
	}
	if !strings.Contains(description, "declined") {
		t.Fatalf("declines are not labelled as declines: %q", description)
	}
}

// The description must say what it walked, so a small number is distinguishable from a wrong root.
func TestTheEnumerationDescriptionNamesTheRootAndTheCounts(t *testing.T) {
	t.Parallel()
	enumeration := enumerationOf("/repo/deep", []string{"/repo/deep/a.ts"}, 900, nil, nil, nil)

	description := wholeTreeScope().narrowToEnumeration(enumeration).Description

	for _, want := range []string{"/repo/deep", "900 files", "1 the formatter handles"} {
		if !strings.Contains(description, want) {
			t.Fatalf("want %q in %q", want, description)
		}
	}
}

// Ignore layers and nested repositories are reported when they removed something.
//
// Nested repositories are named rather than counted because "we skipped a repo" is a fact somebody
// may want to argue with, and a number gives them nothing to argue with. Tonight a whole-tree run
// wrote into a submodule; four nested repositories exist on this tree and three had not been
// mentioned by anyone.
func TestIgnoreLayersAndNestedRepositoriesAreReported(t *testing.T) {
	t.Parallel()
	enumeration := enumerationOf(
		"/repo", []string{"/repo/a.ts"}, 500, nil,
		[]string{"libraries/structure", "projects/ahraos-macos"},
		map[string]int{".gitignore": 400, formatfiles.HouseIgnoreLayer: 28, formatfiles.IgnorePatternsLayer: 0},
	)

	description := wholeTreeScope().narrowToEnumeration(enumeration).Description

	for _, want := range []string{"400 by .gitignore", "28 by format.ignore", "libraries/structure", "projects/ahraos-macos"} {
		if !strings.Contains(description, want) {
			t.Fatalf("want %q in %q", want, description)
		}
	}
	// A layer that removed nothing is not printed. Printing a zero every run trains the reader to
	// stop looking at the line.
	if strings.Contains(description, "0 by ignorePatterns") {
		t.Fatalf("a layer that removed nothing was printed: %q", description)
	}
}

// A named scope keeps its own set and reports how much of it the formatter will see.
func TestANarrowScopeReportsHowMuchIsFormattable(t *testing.T) {
	t.Parallel()
	scope := formatScope{
		FileNames: []string{"/repo/a.ts", "/repo/b.rb", "/repo/c.css"},
		index: map[string]struct{}{
			"/repo/a.ts": {}, "/repo/b.rb": {}, "/repo/c.css": {},
		},
		Description: "3 named paths",
	}
	// The engine handles the ts and the css; the rb was declined during the walk.
	enumeration := enumerationOf("/repo", []string{"/repo/a.ts", "/repo/c.css"}, 40, map[string]int{".rb": 1}, nil, nil)

	narrowed := scope.narrowToEnumeration(enumeration)

	if !strings.Contains(narrowed.Description, "3 named paths") {
		t.Fatalf("the named count was lost: %q", narrowed.Description)
	}
	if !strings.Contains(narrowed.Description, "2 the formatter handles") {
		t.Fatalf("the formattable count is wrong or missing: %q", narrowed.Description)
	}
	// The scope keeps all three, so the two numbers describe two populations rather than collapsing.
	if len(narrowed.FileNames) != 3 {
		t.Fatalf("narrowing shrank the scope instead of describing it: %v", narrowed.FileNames)
	}
}

// An empty scope keeps the description that names where it looked, exactly as narrowTo does.
func TestAnEmptyScopeSurvivesEnumerationNarrowing(t *testing.T) {
	t.Parallel()
	empty := formatScope{
		index:       map[string]struct{}{},
		Description: "0 files under the named paths in /some/where",
	}

	narrowed := empty.narrowToEnumeration(enumerationOf("/repo", []string{"/repo/a.ts"}, 10, nil, nil, nil))

	if !strings.Contains(narrowed.Description, "/some/where") {
		t.Fatalf("enumeration narrowing discarded the directory an empty scope had named: %q", narrowed.Description)
	}
}

// TestNamedPathsScopeResolvesFilesAndDirectories holds the four answers a named path can produce.
//
// A positional path was parsed and discarded before this scope existed, so `cohere --lint OneFile.ts`
// checked all 3,542 files and printed every finding in the tree: 3.042s and 5,201 findings for one
// named file, against 3.003s and 5,201 for the whole tree. Nothing in either run said the argument
// had been ignored, which is why the assertions below are about membership rather than about a
// count: a scope that silently held everything would satisfy a count.
func TestNamedPathsScopeResolvesFilesAndDirectories(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()

	write := func(relative string) string {
		t.Helper()
		path := filepath.Join(directory, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("const a = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	named := write("source/Named.ts")
	sibling := write("source/Sibling.ts")
	nested := write("source/deep/Nested.ts")
	elsewhere := write("other/Elsewhere.ts")

	t.Run("one file is one file", func(t *testing.T) {
		t.Parallel()
		scope, err := namedPathsScope(directory, directory, []string{"source/Named.ts"})
		if err != nil {
			t.Fatal(err)
		}
		if scope.Everything {
			t.Fatal("a named file produced a whole-tree scope, which is the defect this replaces")
		}
		if !scope.includes(named) {
			t.Errorf("the named file was not in scope: %v", scope.FileNames)
		}
		// The control. Without it a scope that included everything would pass the assertion above.
		if scope.includes(sibling) || scope.includes(elsewhere) {
			t.Errorf("an unnamed file was in scope: %v", scope.FileNames)
		}
	})

	t.Run("a directory is everything under it", func(t *testing.T) {
		t.Parallel()
		scope, err := namedPathsScope(directory, directory, []string{"source"})
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{named, sibling, nested} {
			if !scope.includes(want) {
				t.Errorf("a file under the named directory was not in scope: %s", want)
			}
		}
		if scope.includes(elsewhere) {
			t.Errorf("a file outside the named directory was in scope: %s", elsewhere)
		}
	})

	// `cohere --lint .` has always meant the whole tree and has to keep meaning it. A prefix match
	// would make it a subset of one directory entry, silently.
	t.Run("the working directory itself is the whole tree", func(t *testing.T) {
		t.Parallel()
		scope, err := namedPathsScope(directory, directory, []string{"."})
		if err != nil {
			t.Fatal(err)
		}
		if !scope.Everything {
			t.Fatalf("`.` narrowed instead of meaning everything: %v", scope.FileNames)
		}
	})

	// A typo must fail loudly. An empty scope reporting success is the green-over-zero-files failure,
	// and this scope reached it once during development by resolving against an empty directory.
	t.Run("a path that does not exist is an error", func(t *testing.T) {
		t.Parallel()
		if _, err := namedPathsScope(directory, directory, []string{"source/Missing.ts"}); err == nil {
			t.Fatal("a nonexistent path resolved without error")
		}
	})

	// The working directory defaults to empty at the flag, meaning the process's own. Resolving
	// against empty leaves a relative path, which never matches an absolute source file name, so
	// every named path fell out of scope and the run checked nothing while reporting success.
	t.Run("an empty working directory resolves against the process", func(t *testing.T) {
		t.Parallel()
		scope, err := namedPathsScope("", "", []string{"scope.go"})
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range scope.FileNames {
			if !filepath.IsAbs(name) {
				t.Fatalf("a relative path reached the scope index and would match nothing: %s", name)
			}
		}
	})
}

// TestNarrowToUsesTheWholeProgramNotTheLintScope holds that the format phase keeps its own universe
// when the lint phase has been scoped to a named path.
//
// The two narrowings are different questions and the file already says so: a proposed fix comes from
// a rule that ran over the program, so its candidate must be in the program; a format candidate
// comes from the disk. Scoping the lint walk to one named file and then narrowing the format scope
// against that same slice reported `7 changed files, 0 of them in the program` on a tree where three
// of them were.
//
// This asserts the composition rather than `narrowTo` alone, because `narrowTo` was correct at both
// readings. What changed was the population handed to it, which is precisely the shape of mistake
// the comment above that function was written about.
func TestNarrowToUsesTheWholeProgramNotTheLintScope(t *testing.T) {
	t.Parallel()
	changed := formatScope{
		FileNames: []string{"/repository/a.ts", "/repository/b.ts", "/repository/notes.md"},
		index: map[string]struct{}{
			"/repository/a.ts":     {},
			"/repository/b.ts":     {},
			"/repository/notes.md": {},
		},
		Description: "3 named paths",
	}

	wholeProgram := map[string]struct{}{
		"/repository/a.ts": {},
		"/repository/b.ts": {},
		"/repository/c.ts": {},
	}
	if narrowed := changed.narrowTo(wholeProgram); !strings.Contains(narrowed.Description, "2 of them in the program") {
		t.Fatalf("the whole program should have found two of the changed files: %s", narrowed.Description)
	}

	// The regression. A lint scope of one named file is not the population this question is about,
	// and narrowing against it reports zero on a tree that has two.
	lintScopedToOneFile := map[string]struct{}{"/repository/c.ts": {}}
	if narrowed := changed.narrowTo(lintScopedToOneFile); !strings.Contains(narrowed.Description, "0 of them in the program") {
		t.Fatalf("this fixture no longer reproduces the defect it was written for: %s", narrowed.Description)
	}
}

// TestNarrowToAppendsToTheScopesOwnWording holds that this function does not choose the wording.
//
// Four defects of one shape have passed through these three lines. The empty-scope branch above
// carries the first, the format phase's population the second, inferring `changed files` for a
// named scope the third, and inferring `named` from a request description the fourth, which made
// `--changed` report `7 named` for files nobody named one heartbeat after that field was added for
// an unrelated reason.
//
// The repair is that each constructor writes a sentence about itself and this function appends to
// it. The assertions below are that both kinds survive unchanged, which is the property the four
// defects each broke in a different direction.
func TestNarrowToAppendsToTheScopesOwnWording(t *testing.T) {
	t.Parallel()
	population := map[string]struct{}{"/repository/a.ts": {}}

	named := formatScope{
		FileNames:   []string{"/repository/a.ts"},
		index:       map[string]struct{}{"/repository/a.ts": {}},
		Description: "1 named path",
	}
	if described := named.narrowTo(population).Description; described != "1 named path, 1 of them in the program" {
		t.Errorf("a named scope's own wording did not survive: %s", described)
	}

	fromGit := formatScope{
		FileNames:   []string{"/repository/a.ts"},
		index:       map[string]struct{}{"/repository/a.ts": {}},
		Description: "1 named paths",
	}
	described := fromGit.narrowTo(population).Description
	if described != "1 named paths, 1 of them in the program" {
		t.Errorf("a changed-files scope's own wording did not survive: %s", described)
	}

	// The control that matters: the two must differ. A function that returned one spelling for both
	// would satisfy either assertion alone, and that is exactly what the third and fourth defects
	// did.
	if described == named.narrowTo(population).Description {
		t.Error("both kinds of scope described themselves identically")
	}
}

// TestTheScopeLineNamesSkippedSymbolicLinks: a walk that skipped links says so, so a link that held a
// source file is a stated skip rather than an absence.
func TestTheScopeLineNamesSkippedSymbolicLinks(t *testing.T) {
	t.Parallel()
	description := describeEnumeration(formatfiles.Enumeration{Root: "/project", Walked: 4, SymbolicLinks: 2, Files: []string{"/project/a.ts", "/project/b.ts"}}, 2)
	if !strings.Contains(description, "skipped 2 symbolic links") {
		t.Fatalf("description %q does not name the skipped links", description)
	}
	if description := describeEnumeration(formatfiles.Enumeration{Root: "/project", Walked: 1, Files: []string{"/project/a.ts"}}, 1); strings.Contains(description, "symbolic") {
		t.Fatalf("a walk with no links names them: %q", description)
	}
}
