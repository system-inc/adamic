package tailwind

import (
	"context"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"github.com/system-inc/cohere/internal/types/program"
)

// Build-once, asserted against the repository the tool actually lints rather than against a fixture.
//
// `TestDesignSystemIsBuiltOncePerProgram` in design_system_test.go already counts builds across a
// real `program.Graph.Walk`, and that test is the one that fails fastest and runs everywhere. This
// file is the second half of the same claim, and it exists because a 40-file temporary directory
// cannot reach three things the shipped run does:
//
//   - **The real design system.** The fixture's stylesheet is six directives and a two-token
//     stand-in for the framework. This repository's is a nine-file `@import` graph resolving 744
//     theme entries and 37 `@utility` blocks. A rebuild that costs 78µs against the fixture costs
//     0.68ms here, and a build path that is cheap enough to hide in the fixture's noise is not
//     cheap enough to hide in this one.
//   - **All nine rules asking at once.** The fixture walks a single probe rule. The shipped run has
//     nine rules calling `DesignSystemForProgram` on every file, which is the arrangement where a
//     per-rule cache rather than a per-run one would still pass a single-rule test. Nine rules over
//     3,481 files is 31,329 calls to a function that must build exactly once.
//   - **The full worker pool on a tree big enough to use it.** `Workers()` clamps to the file count,
//     so the 40-file fixture is not obviously short but the real tree is unambiguously at the
//     16-checker ceiling this machine allows, striding across files for long enough that a racing
//     second build has room to happen.
//
// The counter is process-scoped, so what is asserted is a delta across the walk rather than an
// absolute. That is deliberate: `go test` runs this package's tests in one process, and an absolute
// assertion here would be a test that passes or fails on which other tests ran first.
//
// Every helper is prefixed `liveWalk` for the reason design_system_test.go states: several agents
// work in this package at once, and a helper whose name does not say which component owns it is a
// collision waiting to happen.

// liveWalkRules is every rule in this package, as the shipped run registers them.
//
// Listed rather than read back out of the catalog, because the catalog is global and another
// package's rules landing in it would change what this test walks without anyone editing this file.
// If a tenth Tailwind rule is added and not added here, this test keeps passing while covering less,
// which `TestLiveWalkCoversEveryRegisteredTailwindRule` below turns into a failure.
func liveWalkRules() []rule.Rule {
	return []rule.Rule{
		EnforceConsistentClassOrder,
		EnforceConsistentVariantOrder,
		EnforceConsistentImportantPosition,
		EnforceConsistentVariableSyntax,
		EnforceShorthandClasses,
		EnforceCanonicalClasses,
		NoConcatenatedClasses,
		NoConflictingClasses,
		NoDeprecatedClasses,
		NoDuplicateClasses,
		NoUnknownClasses,
		NoUnnecessaryWhitespace,
		NoPhysicalDirection,
	}
}

// liveWalkMeasuredBuildCost is what one build of this repository's design system costs.
//
// Measured at 1.66ms mean over 30 cold builds against ~/Projects/ahra on 2026-08-25, covering the
// design system and the descriptor table that rides the same cache entry. Higher than the 0.68ms
// this component's file comment quotes, and the gap is the table: that figure was taken before the
// table moved onto this path, so it now understates the waste by more than half.
//
// It appears only inside a failure message, so drift costs an inaccurate estimate rather than a
// wrong verdict. The cached lookup that 3,480 of the 3,481 files pay instead measures 5ns.
const liveWalkMeasuredBuildCost = 1660 * time.Microsecond

// TestDesignSystemIsBuiltOnceUnderTheRealWalk is the end-to-end form of the build-once claim.
//
// The failure it guards has no symptom other than time. A sibling rule in this tree spent 2,329ms of
// setup against 0.54ms of listening across 1,862 files because its scan was redone per file, while
// its own comment said it happened once. Every finding stayed correct and every test stayed green.
// So the claim is a counter read across a real walk of a real repository, which is the only form of
// it that cannot decay into a sentence someone believes.
// Not parallel: it resets the package designSystemCache and asserts the process-wide
// tailwindengine.BuildsSoFar moved by exactly one across the walk, which any parallel test building a
// design system or evicting the cache's one slot would inflate.
//
// The tree is ahra, the corpus the rest of the live suites in this package read. Without it the test
// skips naming the variable, because what is under test is a property of the rules and not of the
// machine.
func TestDesignSystemIsBuiltOnceUnderTheRealWalk(t *testing.T) {
	liveWalkRepository := corpus.Ahra.Root(t)

	resetDesignSystemCacheForTest()

	graph, err := program.Build(program.Options{
		ConfigFileName:   "tsconfig.json",
		CurrentDirectory: liveWalkRepository,
	})
	if err != nil {
		t.Fatalf("building a program over %s: %v", liveWalkRepository, err)
	}

	files := graph.ProjectFiles()
	if len(files) < 1000 {
		t.Fatalf("the corpus repository yielded %d project files; this assertion is about a tree "+
			"large enough to use the worker pool, and a number this small means the program was "+
			"built over something other than the repository", len(files))
	}

	// The pool is asserted rather than assumed. A walk that silently ran single-threaded would pass
	// every count below while proving nothing about the concurrency the cache's mutex exists for.
	if workers := graph.Workers(); workers < 2 {
		t.Fatalf("the walk would run with %d worker(s), so this test cannot observe a race between "+
			"two builds and the property it claims to prove would be untested", workers)
	}

	rules := liveWalkRules()
	before := tailwindengine.BuildsSoFar()
	if _, err := graph.Walk(context.Background(), files, rules); err != nil {
		t.Fatalf("walking %s: %v", liveWalkRepository, err)
	}
	built := tailwindengine.BuildsSoFar() - before

	if built != 1 {
		// The counterfactual is stated in the failure rather than left for the reader to compute,
		// because the whole defect is that the number is invisible until someone multiplies it out.
		t.Fatalf("the design system was built %d times across %d files by %d rules on %d workers; "+
			"it must be built exactly once per run. At the measured %v per build on this "+
			"repository, %d extra builds is roughly %v of pure waste, and nothing else would fail",
			built, len(files), len(rules), graph.Workers(), liveWalkMeasuredBuildCost,
			built-1, time.Duration(built-1)*liveWalkMeasuredBuildCost)
	}
}

// TestLiveWalkCoversEveryRegisteredTailwindRule keeps the list above honest.
//
// `liveWalkRules` is written by hand, so the failure mode is a tenth rule that reads the design
// system, is never added here, and is therefore never covered by the build-once assertion. That is a
// test that keeps passing while measuring less, which is the quietest way for a guard to stop
// guarding. Counting the list against the package's own registrations turns it into a failure that
// names the missing rule.
func TestLiveWalkCoversEveryRegisteredTailwindRule(t *testing.T) {
	t.Parallel()
	covered := make(map[string]bool)
	for _, subject := range liveWalkRules() {
		covered[subject.Name] = true
	}

	// The catalog, which in this package's test binary holds this package's registrations and
	// nothing else: `rule.Register` is called from each rule package's own `init`, and a test binary
	// links only the packages under test and their imports.
	//
	// Guarded rather than trusted. If that ever stops being true the count grows, and a silent grow
	// would make this test start demanding coverage of rules that are not this package's business,
	// which is a failure that would read as this test being wrong rather than as the catalog moving.
	registered := rule.Registered()
	if len(registered) != len(liveWalkRules()) {
		t.Fatalf("the catalog holds %d rules and liveWalkRules lists %d; either a Tailwind rule was "+
			"added without being added to liveWalkRules, or another package's rules are now linked "+
			"into this test binary and this comparison is measuring the wrong set",
			len(registered), len(liveWalkRules()))
	}
	for _, registration := range registered {
		if !covered[registration.Rule.Name] {
			t.Errorf("rule %q is registered by this package but is not in liveWalkRules, so the "+
				"build-once assertion never walks it", registration.Rule.Name)
		}
	}
}
