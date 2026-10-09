package tailwind

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
)

// The measurement `#twany` opened with, re-run at the end of it.
//
// That task recorded a design system sharing nothing with the corpus, and this measurement:
//
//	declares 2 utility roots
//	2 of 2 absent from RootDeclaredProperties
//	2 of 2 absent from ComposingRoots
//
// Both tables are deleted, so the question is no longer whether a row exists but whether the
// computation answers those roots. It does, and this holds it rather than leaving the claim in a
// commit message where a reader pointing the tool at a new repository will not find it.
//
// The corpus cannot substitute for this file. ahra and www-connected-app vendor the same Structure
// submodule, so a fact they agree on may be a fact about the submodule, which is how five generated
// tables carried one repository's tokens as framework facts for as long as they did.
func TestTheUnseenRepositorysRootsAreAnswered(t *testing.T) {
	t.Parallel()
	system := unseenDesignSystem(t)

	// The two `@utility` roots the file declares. Spelled with a value for the functional one,
	// because a functional root without one is a different question.
	answers := map[string]bool{
		"synthetic-static":   true,
		"synthetic-fn-small": true,
	}

	for className, wantAnswered := range answers {
		candidates := ParseCandidate(className, system)
		if len(candidates) == 0 {
			t.Errorf("%s does not parse against the system that declares it", className)
			continue
		}
		if got := ClassValueResolvesIn(&candidates[0], system); got != wantAnswered {
			t.Errorf("%s: the unseen system declares it and the port answers %v", className, got)
		}
	}
	t.Logf("the two roots #twany measured as invisible: %d of %d answered", len(answers), len(answers))
}

// A framework class resolves against a theme the port has never seen, and a typo does not.
//
// The half that makes the test above mean something. A port that answered everything would pass it,
// so this asserts the other direction on the same system: `text-nonsense` and `bg-weird` name
// framework roots whose values this file's theme does not hold, and `--color-weird` is deliberately
// a namespace no framework root consumes.
func TestTheUnseenRepositorySeparatesClassesFromTypos(t *testing.T) {
	t.Parallel()
	system := unseenDesignSystem(t)

	answers := map[string]bool{
		"text-lg":       true,
		"border-2":      true,
		"shadow-lg":     true,
		"px-4":          true,
		"text-nonsense": false,
		"bg-weird":      false,
	}

	var resolved, refused int
	for className, wantResolves := range answers {
		candidates := ParseCandidate(className, system)
		if len(candidates) == 0 {
			t.Errorf("%s does not parse", className)
			continue
		}
		got := ClassValueResolvesIn(&candidates[0], system)
		if got != wantResolves {
			t.Errorf("%s: want resolves=%v against the unseen system, got %v", className, wantResolves, got)
			continue
		}
		if got {
			resolved++
		} else {
			refused++
		}
	}

	t.Logf("unseen system: %d framework classes resolved, %d typos refused", resolved, refused)
	if resolved == 0 || refused == 0 {
		t.Fatal("one side of the comparison was empty, so this test measured nothing")
	}
}

// A repository functional root is declined by the table and answered by the evaluator, deliberately.
//
// This is the gap the closing claim names rather than hides. `Table.Lookup` returns false for
// `synthetic-fn-small` because a measured descriptor row needs the engine to probe 527 value shapes
// and nothing in shipped Go can do that. `addRepositoryFunctionalRoots` gives every repository root
// a present-but-declining row so the decline reads as "this table is not what answers it" rather
// than as "this root does not exist".
//
// Pinned because it is the difference between a known limit and a silent wrong answer, and because a
// future change that made the table answer these would need to delete this deliberately.
func TestTheUnseenRepositorysFunctionalRootIsDeclinedByTheTable(t *testing.T) {
	t.Parallel()
	system := unseenDesignSystem(t)
	table := NewTable(system)

	candidates := ParseCandidate("synthetic-fn-small", system)
	if len(candidates) == 0 {
		t.Fatal("synthetic-fn-small does not parse against the system that declares it")
	}

	if _, answered := table.Lookup(&candidates[0]); answered {
		t.Error("the table now answers a repository functional root, which the closing claim in " +
			"descriptor_base_table.go says it cannot; that claim needs re-measuring rather than this " +
			"assertion being deleted")
	}
	if !ClassValueResolvesIn(&candidates[0], system) {
		t.Error("the evaluator's half must still answer it, or the decline above is a real gap")
	}
}

// unseenDesignSystem loads the design system that shares nothing with the corpus.
func unseenDesignSystem(t *testing.T) *LoadedDesignSystem {
	t.Helper()

	entryPoint, err := filepath.Abs(filepath.Join("..", "tools", "generate_descriptor_base", "testdata", "independent_theme.css"))
	if err != nil {
		t.Fatalf("resolving the unseen system's path: %v", err)
	}
	if _, err := os.Stat(entryPoint); err != nil {
		t.Skipf("the unseen design system is not present at %s", entryPoint)
	}

	// The vendored snapshot of tailwindcss, so the unseen system resolves on any machine (#sycrdr6).
	packageRoot := vendored.TailwindPackageRoot()

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading the unseen design system: %v", err)
	}
	return system
}

// A root declared both as a static and as a functional `@utility` keeps both.
//
// `@utility fade-in` and `@utility fade-in-*` are two blocks naming one root: a static form taking
// no value and a functional form taking one. Sixteen roots in the corpus have that shape, every
// `fade-*`, `slide-*` and `zoom-*` animation.
//
// `utilityRoots` mapped a root to one kind, so the second block overwrote the first. The static form
// was lost silently, and since the parser reads `fade-in` as functional while the evaluator holds
// only functional definitions, the class was answerable by neither: `readingFor` declined it after
// consulting both. Measured on the corpus, `fade-in` and `fade-out` were the only two classes of 799
// that nothing could answer.
//
// Asserted on the live system rather than a list, so a repository adding such a pair joins this test.
func TestARootDeclaredBothWaysKeepsBothKinds(t *testing.T) {
	t.Parallel()
	system, _ := liveTableFor(t, corpusRepositories[0].spelling)
	if system == nil {
		t.Skip("no design system loaded")
	}

	var bothKinds int
	for _, root := range []string{"fade-in", "fade-out", "zoom-in", "zoom-out"} {
		staticForm := system.HasUtility(root, UtilityKindStatic)
		functionalForm := system.HasUtility(root, UtilityKindFunctional)

		if !staticForm || !functionalForm {
			t.Errorf("`@utility %s` and `@utility %s-*` are both declared and the system reports static=%v functional=%v",
				root, root, staticForm, functionalForm)
			continue
		}
		bothKinds++
	}

	t.Logf("roots declared both ways: %d checked", bothKinds)
	if bothKinds == 0 {
		t.Skip("this repository declares no root both ways")
	}
}
