package tailwind

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"github.com/system-inc/cohere/internal/types/program"
)

// The suite that makes "built once per run" a fact rather than a comment.
//
// Every helper here is prefixed `designSystem` on purpose. Several agents work in this tree at once
// and two of them independently defining `describeCandidate` blocked the whole package's tests for
// a while tonight, so a helper whose name does not say which component owns it is a collision
// waiting to happen.

const designSystemTestConfig = `{
    "compilerOptions": {
        "target": "ES2022",
        "module": "esnext",
        "moduleResolution": "bundler",
        "jsx": "react-jsx",
        "strict": true,
        "noEmit": true
    },
    "include": ["./**/*.ts", "./**/*.tsx"]
}`

// designSystemStylesheet is a miniature `@import` graph exercising every directive the collector
// handles: a nested import, an `@theme` with a prefix-free block, an `@utility` of each arity, a
// `@custom-variant` that reuses a framework name and one that does not, and a `@config` that must
// be reported rather than swallowed.
const designSystemStylesheet = `@import "tailwindcss";
@import "./tokens.css";
@config "./TailwindConfiguration.ts";

@custom-variant dark (&:where(.dark *));
@custom-variant pointer-coarse (@media (pointer: coarse));

@utility scrollbar-none {
    scrollbar-width: none;
}

@utility fade-in-* {
    --enter-opacity: --value(--percentage-*, integer);
}
`

const designSystemTokensStylesheet = `@theme {
    --color-brand: oklch(0.7 0.2 250);
    --color-brand-muted: oklch(0.5 0.1 250);
    --percentage-0: 0;
    --spacing-gutter: 1.25rem;
}
`

// designSystemFrameworkStylesheet stands in for the installed `tailwindcss` package's index.css.
//
// Small on purpose. This suite is about when the design system is built, not about what the
// framework's own theme resolves to, and that question already has 22,521 engine-compared answers
// in `internal/tailwind`. A test that pulled in the real framework stylesheet here would be slower
// and would fail for reasons belonging to another component.
//
// The `@layer theme { @theme default { ... } }` nesting is not decoration. It is exactly how the
// framework writes its own index.css, and a collector that handled only a top-level `@theme` would
// drop the entire framework theme on every real repository while passing a fixture that wrote the
// block flat. That defect escaped an earlier version of this constant, which is why the nesting is
// here and why this paragraph is.
const designSystemFrameworkStylesheet = `@layer theme {
    @theme default {
        --color-red-500: oklch(0.637 0.237 25.331);
        --spacing: 0.25rem;
    }
}
`

// designSystemProject lays a project on disk and returns its root.
//
// A real directory rather than an in-memory host, because the whole component under test reaches
// the filesystem: entry-point discovery, package-root discovery and the `@import` walk are three of
// the four things that can go wrong, and none of them is reachable through a fake.
func designSystemProject(t *testing.T, sourceFileCount int) string {
	t.Helper()

	root := t.TempDir()
	write := func(relativePath, contents string) {
		full := filepath.Join(root, relativePath)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
			t.Fatalf("writing %s: %v", full, err)
		}
	}

	write("tsconfig.json", designSystemTestConfig)
	write(filepath.Join("app", "_theme", "styles", "theme.css"), designSystemStylesheet)
	write(filepath.Join("app", "_theme", "styles", "tokens.css"), designSystemTokensStylesheet)
	write(filepath.Join("node_modules", "tailwindcss", "index.css"), designSystemFrameworkStylesheet)

	// Enough files that a per-file rebuild would be unmistakable in the counter rather than a
	// coincidence, and enough that the walk actually uses more than one worker.
	for index := range sourceFileCount {
		write(
			filepath.Join("app", "components", "Component"+string(rune('A'+index%26))+string(rune('0'+index/26))+".tsx"),
			"export function Component() {\n    return <div className=\"flex items-center\" />;\n}\n",
		)
	}

	return root
}

// designSystemProbeRule is a rule that asks for the design system on every file it sees.
//
// It reports nothing and records what it saw. Reporting would make this test depend on a rule's
// findings, which is a different question owned by a different task; what is under test is how many
// times the build ran and whether every file got the same value.
type designSystemProbe struct {
	mutex   sync.Mutex
	results []DesignSystemResult
	files   int
}

// designSystemProgram is a program as a rule calling DesignSystemForProgram sees it, declaring what the
// design system reads.
func designSystemProgram(program *compiler.Program) rule.Program {
	return rule.ViewProgram(program, nil, rule.Rule{Name: "test/design-system", ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem})
}

func (probe *designSystemProbe) rule() rule.Rule {
	return rule.Rule{
		Name: "tailwind-design-system-probe",
		// What DesignSystemForProgram reads. ctx.Program refuses anything undeclared.
		ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem,
		Run: func(ctx rule.Context, _ any) rule.Listeners {
			result := DesignSystemForProgram(ctx.Program)
			probe.mutex.Lock()
			probe.results = append(probe.results, result)
			probe.files++
			probe.mutex.Unlock()
			return nil
		},
	}
}

// designSystemWalk builds a program over a project and walks every file in it with one rule.
//
// The real `program.Graph.Walk` rather than a hand-rolled traversal, because the property under
// test is a concurrency property: the walk runs up to 16 goroutines striding over files, and a
// single-threaded harness would pass against a cache with no mutex at all.
func designSystemWalk(t *testing.T, root string, subject rule.Rule) *program.Graph {
	t.Helper()

	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: root})
	if err != nil {
		t.Fatalf("building the program: %v", err)
	}

	files := graph.ProjectFiles()
	if len(files) == 0 {
		t.Fatal("the program holds no project files, so the walk would prove nothing")
	}

	if _, err := graph.Walk(context.Background(), files, []rule.Rule{subject}); err != nil {
		t.Fatalf("walking: %v", err)
	}
	return graph
}

// TestDesignSystemIsBuiltOncePerProgram is the assertion this whole component exists to make.
//
// A sibling rule in this tree measured 2,329ms of setup against 0.54ms of listening across 1,862
// files because its scan was redone per file, while its own comment claimed it happened once. That
// failure has no symptom other than time: every finding is correct, every test passes, and the run
// is twenty-five times slower. So the claim is a counter here rather than a sentence.
// Not parallel: it resets the package designSystemCache and asserts the process-wide
// tailwindengine.BuildsSoFar moved by exactly one, so a parallel test building a design system or walking
// another program (which evicts the cache's one slot) would add builds and hand later files a different
// pointer.
func TestDesignSystemIsBuiltOncePerProgram(t *testing.T) {
	resetDesignSystemCacheForTest()

	root := designSystemProject(t, 40)
	probe := &designSystemProbe{}

	before := tailwindengine.BuildsSoFar()
	graph := designSystemWalk(t, root, probe.rule())
	built := tailwindengine.BuildsSoFar() - before

	if probe.files < 40 {
		t.Fatalf("the probe saw %d files; fewer than the 40 laid down means a per-file rebuild "+
			"could hide inside the count", probe.files)
	}
	if built != 1 {
		t.Fatalf("the design system was built %d times across %d files; it must be built exactly "+
			"once per run. At the measured 0.68ms per build that is %v of pure waste on this tree "+
			"alone, and nothing else would fail",
			built, probe.files, time.Duration(built-1)*680*time.Microsecond)
	}

	// Same value on every file, not merely the same count. A cache that built once and then handed
	// out copies would pass the count and still break the immutability the parallel walk depends on.
	first := probe.results[0]
	if first.Err != nil {
		t.Fatalf("the fixture project should produce a design system and produced: %v", first.Err)
	}
	for index, result := range probe.results {
		if result.System != first.System {
			t.Fatalf("file %d received a different design system pointer than file 0; the cache is "+
				"handing out per-file values", index)
		}
		// The descriptor table rides the same cache entry, so it gets the same assertion. The
		// counter cannot cover it — it counts design system builds — and a table rebuilt per file
		// would be invisible to every check above while costing 959µs a file, which is more than the
		// design system build it sits behind.
		if result.Table != first.Table {
			t.Fatalf("file %d received a different descriptor table pointer than file 0; the table "+
				"is being rebuilt per file", index)
		}
	}
	if first.System.BuildCount == 0 {
		t.Fatal("BuildCount is zero, so the counter is not being set and every assertion above is vacuous")
	}
	if first.Table == nil {
		t.Fatal("the result carries no descriptor table, so a rule computing a sort key has nothing to look up in")
	}

	_ = graph
}

// TestDesignSystemCacheKeyIsTheProgram is the mutation the task asked for, run as a test.
//
// The precedent this file copies records the original bug: a module-level memo with no key, which
// is correct for a process that lints once and exits and wrong for anything that lints twice,
// because the second tree reads the first's data. Here that is executed rather than described. Two
// programs over two different projects are walked in sequence, and the second must see its own
// design system.
//
// Mutating the key to nothing — deleting the `designSystemCache.program == ctx.Program` comparison
// so any cached result is returned — makes this test fail on the entry-point assertion below, which
// is what makes the key load-bearing rather than decorative.
// Not parallel: it resets the package designSystemCache, the one slot every Tailwind rule's design system
// is cached in, and then asserts on what that slot holds, so it walks its two programs in turn with no
// other test's program taking the slot between them.
func TestDesignSystemCacheKeyIsTheProgram(t *testing.T) {
	resetDesignSystemCacheForTest()

	firstRoot := designSystemProject(t, 2)
	secondRoot := designSystemProject(t, 2)

	firstProbe := &designSystemProbe{}
	designSystemWalk(t, firstRoot, firstProbe.rule())

	secondProbe := &designSystemProbe{}
	designSystemWalk(t, secondRoot, secondProbe.rule())

	if len(firstProbe.results) == 0 || len(secondProbe.results) == 0 {
		t.Fatal("one of the two walks produced no results, so nothing is being compared")
	}

	firstEntry := firstProbe.results[0].EntryPoint
	secondEntry := secondProbe.results[0].EntryPoint
	if firstEntry == "" || secondEntry == "" {
		t.Fatalf("both walks must find an entry point; got %q and %q", firstEntry, secondEntry)
	}
	if firstEntry == secondEntry {
		t.Fatalf("both programs resolved the same entry point %q, so the second tree is reading the "+
			"first's data. This is the exact bug the program-pointer key exists to prevent",
			firstEntry)
	}
	if !strings.HasPrefix(firstEntry, firstRoot) {
		t.Fatalf("the first program's entry point %q is not inside its own project %q", firstEntry, firstRoot)
	}
	if !strings.HasPrefix(secondEntry, secondRoot) {
		t.Fatalf("the second program's entry point %q is not inside its own project %q", secondEntry, secondRoot)
	}
}

// TestDesignSystemBuildsAgainForASecondProgram is the other half of the key.
//
// The test above proves a second program does not read the first's data. This proves it actually
// gets one of its own, which a cache that returned an error on every miss would also pass the first
// test while failing every rule.
// Not parallel: it resets the package designSystemCache and asserts the process-wide
// tailwindengine.BuildsSoFar moved by exactly two, which any parallel test building a design system would
// inflate.
func TestDesignSystemBuildsAgainForASecondProgram(t *testing.T) {
	resetDesignSystemCacheForTest()

	before := tailwindengine.BuildsSoFar()

	firstProbe := &designSystemProbe{}
	designSystemWalk(t, designSystemProject(t, 2), firstProbe.rule())

	secondProbe := &designSystemProbe{}
	designSystemWalk(t, designSystemProject(t, 2), secondProbe.rule())

	if built := tailwindengine.BuildsSoFar() - before; built != 2 {
		t.Fatalf("two programs should produce exactly two builds, got %d", built)
	}
	if firstProbe.results[0].System == secondProbe.results[0].System {
		t.Fatal("the two programs share one design system pointer, so the second never built its own")
	}
}

// TestDesignSystemDeclinesRatherThanReportingClean is the failing-safe contract.
//
// A project with no Tailwind entry point must produce an error a rule can decline on, never a
// zero-valued design system. The distinction is the whole reason cohere exists: a rule reporting no
// findings because the CSS could not be read is indistinguishable from a clean tree, and it is the
// one failure mode a linter must not have.
// Not parallel: it resets the package designSystemCache, the one slot every Tailwind rule's design system
// is cached in, and then asserts on what that slot holds.
func TestDesignSystemDeclinesRatherThanReportingClean(t *testing.T) {
	resetDesignSystemCacheForTest()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(designSystemTestConfig), 0o644); err != nil {
		t.Fatalf("writing tsconfig: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "Component.tsx"),
		[]byte("export const value = 1;\n"), 0o644,
	); err != nil {
		t.Fatalf("writing source: %v", err)
	}

	probe := &designSystemProbe{}
	designSystemWalk(t, root, probe.rule())

	if len(probe.results) == 0 {
		t.Fatal("the probe never ran, so the decline path was never reached")
	}
	result := probe.results[0]
	if result.System != nil {
		t.Fatal("a project with no stylesheet produced a design system, which means rules would " +
			"silently lint against an empty theme and report a clean tree")
	}
	if !errors.Is(result.Err, ErrNoTailwindEntryPoint) {
		t.Fatalf("expected ErrNoTailwindEntryPoint so a caller can tell 'no Tailwind here' from "+
			"'the CSS would not parse'; got %v", result.Err)
	}

	message := DesignSystemDeclineMessage("tailwind-enforce-consistent-class-order", result)
	if !strings.Contains(message, "reporting nothing rather than reporting a clean tree") {
		t.Fatalf("the decline message must say the rule is silent rather than clean; got %q", message)
	}
}

// TestDesignSystemFailedBuildIsCachedToo asserts the failure path is not the slow path.
//
// A repository with no Tailwind would otherwise re-run entry-point discovery on every file, which
// is the per-file cost this component exists to refuse, in the case where there is not even a
// design system to show for it.
// Not parallel: it resets the package designSystemCache and asserts every file got the same cached error
// value, and a parallel test walking another program would evict the cache's one slot between files and
// force a fresh error.
func TestDesignSystemFailedBuildIsCachedToo(t *testing.T) {
	resetDesignSystemCacheForTest()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(designSystemTestConfig), 0o644); err != nil {
		t.Fatalf("writing tsconfig: %v", err)
	}
	for index := range 8 {
		if err := os.WriteFile(
			filepath.Join(root, "Component"+string(rune('A'+index))+".tsx"),
			[]byte("export const value = 1;\n"), 0o644,
		); err != nil {
			t.Fatalf("writing source: %v", err)
		}
	}

	probe := &designSystemProbe{}
	designSystemWalk(t, root, probe.rule())

	if probe.files < 8 {
		t.Fatalf("expected at least 8 files, saw %d", probe.files)
	}
	first := probe.results[0]
	for index, result := range probe.results {
		if result.Err != first.Err {
			t.Fatalf("file %d received a freshly-built error rather than the cached one, so "+
				"discovery is running per file", index)
		}
	}
}

// TestDesignSystemComposesTheRepositorysOwnContributions asserts the seam actually assembled
// something, rather than caching an empty value once and passing every timing assertion.
//
// This is the control for the whole suite. Every test above would pass against a design system that
// built nothing at all, which is the shape of a harness that has never returned a positive.
// Not parallel: it resets the package designSystemCache, the one slot every Tailwind rule's design system
// is cached in, and then asserts on what that slot holds.
func TestDesignSystemComposesTheRepositorysOwnContributions(t *testing.T) {
	resetDesignSystemCacheForTest()

	root := designSystemProject(t, 2)
	probe := &designSystemProbe{}
	designSystemWalk(t, root, probe.rule())

	result := probe.results[0]
	if result.Err != nil {
		t.Fatalf("the fixture project should build: %v", result.Err)
	}
	system := result.System

	contributions := system.Contributions()
	// Four tokens in tokens.css, two in the stand-in framework stylesheet.
	if contributions.ThemeEntries != 6 {
		t.Errorf("expected 6 theme entries from the fixture graph, got %d", contributions.ThemeEntries)
	}
	if contributions.UtilityBlocks != 2 {
		t.Errorf("expected the 2 `@utility` blocks the fixture declares, got %d", contributions.UtilityBlocks)
	}
	if contributions.CustomVariants != 2 {
		t.Errorf("expected the 2 `@custom-variant` names the fixture declares, got %d", contributions.CustomVariants)
	}
	if contributions.StylesheetCount != 3 {
		t.Errorf("expected the graph to reach theme.css, tokens.css and the framework's index.css, got %d",
			contributions.StylesheetCount)
	}

	// The `@config` is reported rather than swallowed, which is what keeps a theme that is quietly
	// short by a JavaScript config's worth of keys from looking complete.
	if len(system.SkippedDirectives) != 1 || system.SkippedDirectives[0].Name != "@config" {
		t.Errorf("expected exactly the one `@config` to be reported as skipped, got %#v", system.SkippedDirectives)
	}

	// Both arities of `@utility` are registered, under the kind their name declares.
	if !system.HasUtility("scrollbar-none", tailwindengine.UtilityKindStatic) {
		t.Error("`@utility scrollbar-none` should register a static utility")
	}
	if !system.HasUtility("fade-in", tailwindengine.UtilityKindFunctional) {
		t.Error("`@utility fade-in-*` should register a functional utility rooted at `fade-in`")
	}
	if system.HasUtility("fade-in", tailwindengine.UtilityKindStatic) {
		t.Error("`@utility fade-in-*` is functional; reading it as static would let `fade-in` bare compile")
	}

	// The framework's roots still answer, since the repository's blocks are consulted first and
	// then fall through rather than replacing the registry.
	if !system.HasUtility("bg", tailwindengine.UtilityKindFunctional) {
		t.Error("a framework root stopped resolving, so the repository's blocks are shadowing the registry")
	}

	// Only the functional block reaches the evaluator: a static `@utility` has no `--value()` to
	// resolve and would be a definition that can never compile.
	if system.Utilities() == nil {
		t.Fatal("the fixture declares a functional `@utility`, so an evaluator should exist")
	}
	if !system.Utilities().Has("fade-in") {
		t.Error("the evaluator should hold the functional block")
	}
	if system.Utilities().Has("scrollbar-none") {
		t.Error("the evaluator holds a static block, which can never satisfy Compile's first post-condition")
	}

	// The theme is the repository's, not the framework's alone.
	if value, found := system.Theme().Get([]string{"--color-brand"}); !found || value == "" {
		t.Errorf("the repository's own `--color-brand` did not resolve; got %q found=%v", value, found)
	}

	// Stylesheet reporting is project-relative, so a failure names a repository rather than a machine.
	relative := stylesheetsUnder(root, system)
	if len(relative) != 3 {
		t.Fatalf("expected 3 stylesheets, got %d: %v", len(relative), relative)
	}
	if filepath.IsAbs(relative[0]) {
		t.Errorf("the entry point should be reported relative to the project root, got %q", relative[0])
	}
}

// TestDesignSystemCustomVariantsRegisterWithoutMovingFrameworkOnes pins the one behaviour a reader
// is most likely to get wrong.
//
// Upstream's `Variants.set` assigns kind onto an existing record and never reassigns `order`, so a
// `@custom-variant dark (...)` changes the selector `dark:` emits and leaves its sort position
// exactly where the framework put it. Both corpus repositories do precisely that. Only a name the
// framework never registered appends a position.
func TestDesignSystemCustomVariantsRegisterWithoutMovingFrameworkOnes(t *testing.T) {
	t.Parallel()
	resolve := designSystemFixtureResolver(map[string]string{
		"tailwindcss": designSystemFrameworkStylesheet,
	})

	entryPoint := designSystemWriteStylesheet(t, `@import "tailwindcss";
@custom-variant dark (&:where(.dark *));
@custom-variant pointer-coarse (@media (pointer: coarse));
`)

	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint: entryPoint,
		Resolve:    resolve,
		FrameworkVariants: []tailwindengine.FrameworkVariant{
			{Name: "hover", Kind: tailwindengine.ParsedVariantKindStatic},
			{Name: "dark", Kind: tailwindengine.ParsedVariantKindStatic},
			{Name: "group", Kind: tailwindengine.ParsedVariantKindCompound},
		},
	})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	darkBefore, isRegistered := system.Variants().Get("dark")
	if !isRegistered {
		t.Fatal("`dark` should be registered by the framework")
	}
	hover, _ := system.Variants().Get("hover")
	if darkBefore.Order <= hover.Order {
		// `dark` registers after `hover` in the injected list, so its order must be higher. If the
		// re-registration had appended, it would be higher still and past `pointer-coarse`.
		t.Fatalf("`dark` order %d should follow `hover` order %d", darkBefore.Order, hover.Order)
	}

	coarse, isRegistered := system.Variants().Get("pointer-coarse")
	if !isRegistered {
		t.Fatal("a `@custom-variant` under a new name should append a registration")
	}
	if coarse.Order <= darkBefore.Order {
		t.Fatalf("a new `@custom-variant` should append past the framework's last, got order %d "+
			"against `dark` at %d", coarse.Order, darkBefore.Order)
	}

	contributions := system.Contributions()
	if len(contributions.NewVariantNames) != 1 || contributions.NewVariantNames[0] != "pointer-coarse" {
		t.Fatalf("only `pointer-coarse` is new; `dark` updated a framework registration in place. Got %v",
			contributions.NewVariantNames)
	}
}

// TestDesignSystemVariantRegistrationOrderIsDeterministic pins the ordering a map range would
// destroy.
//
// Registration order is a sort key: a `@custom-variant` under a new name appends past the
// framework's last, and where it lands decides where its classes sort. Ranging a map to register
// them assigns those positions in a different sequence on every run, which is nondeterminism that
// passes every test comparing sorted output and produces a different class order on each lint.
//
// This exists because a weaker version of it did not catch the mutation. With the two variants the
// main fixture declares, a map range agrees with sorted order often enough to look stable; twelve
// names make disagreement essentially certain, and ten loads make a single lucky ordering not
// enough to pass.
func TestDesignSystemVariantRegistrationOrderIsDeterministic(t *testing.T) {
	t.Parallel()
	names := []string{
		"zulu", "yankee", "xray", "whiskey", "victor", "uniform",
		"tango", "sierra", "romeo", "quebec", "papa", "oscar",
	}

	stylesheet := "@import \"tailwindcss\";\n"
	for _, name := range names {
		stylesheet += "@custom-variant " + name + " (&:where(." + name + " *));\n"
	}

	resolve := designSystemFixtureResolver(map[string]string{
		"tailwindcss": designSystemFrameworkStylesheet,
	})
	entryPoint := designSystemWriteStylesheet(t, stylesheet)

	orderOf := func() map[string]int {
		system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
			EntryPoint: entryPoint,
			Resolve:    resolve,
			FrameworkVariants: []tailwindengine.FrameworkVariant{
				{Name: "hover", Kind: tailwindengine.ParsedVariantKindStatic},
			},
		})
		if err != nil {
			t.Fatalf("loading: %v", err)
		}
		orders := map[string]int{}
		for _, name := range names {
			registration, isRegistered := system.Variants().Get(name)
			if !isRegistered {
				t.Fatalf("`%s` should be registered", name)
			}
			orders[name] = registration.Order
		}
		return orders
	}

	// Alphabetical, because that is what sortedKeys guarantees and what a map range would not.
	first := orderOf()
	for index := 1; index < len(names); index++ {
		earlier := names[len(names)-1-(index-1)]
		later := names[len(names)-1-index]
		if first[earlier] >= first[later] {
			t.Fatalf("`%s` (order %d) should register before `%s` (order %d): registration walks "+
				"the names in sorted order, and a map range would not",
				earlier, first[earlier], later, first[later])
		}
	}

	// Ten loads, because a single map range can agree with sorted order by luck and a test that
	// ran once would call that stability.
	for attempt := range 10 {
		again := orderOf()
		for _, name := range names {
			if again[name] != first[name] {
				t.Fatalf("load %d gave `%s` order %d where the first gave %d; registration order is "+
					"a sort key, so this is a different class order on every lint",
					attempt+2, name, again[name], first[name])
			}
		}
	}
}

// TestDesignSystemRefusesRatherThanHalfBuilding covers the three ways a graph can be unreadable.
//
// Each must be an error and no value. A half-built design system is worse than none: it produces a
// theme that is quietly short, and every rule reading it reports confidently against tokens the
// repository does not have.
func TestDesignSystemRefusesRatherThanHalfBuilding(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		stylesheet string
		wants      string
	}{
		{
			name:       "unparseable css",
			stylesheet: "@theme {\n    --color-a: red;\n",
			wants:      "",
		},
		{
			name:       "an @import modifier that changes what the file means",
			stylesheet: "@import \"tailwindcss\" theme(reference);\n",
			wants:      "unsupported @import modifier",
		},
		{
			name:       "a non-custom-property inside @theme",
			stylesheet: "@theme {\n    color: red;\n}\n",
			wants:      "must only contain custom properties",
		},
		{
			name:       "an @utility with no name",
			stylesheet: "@utility {\n    display: flex;\n}\n",
			wants:      "`@utility` with no name",
		},
	}

	resolve := designSystemFixtureResolver(map[string]string{
		"tailwindcss": designSystemFrameworkStylesheet,
	})

	for _, aCase := range cases {
		t.Run(aCase.name, func(t *testing.T) {
			t.Parallel()
			entryPoint := designSystemWriteStylesheet(t, aCase.stylesheet)
			system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
				EntryPoint: entryPoint,
				Resolve:    resolve,
			})
			if err == nil {
				t.Fatalf("expected a refusal; got a design system with %d theme entries",
					system.Contributions().ThemeEntries)
			}
			if system != nil {
				t.Fatal("a failed load must return no value at all, never a partial one")
			}
			if aCase.wants != "" && !strings.Contains(err.Error(), aCase.wants) {
				t.Fatalf("expected the error to name %q, got %v", aCase.wants, err)
			}
		})
	}
}

// TestDesignSystemIsSafeUnderTheParallelWalk exercises the cache from many goroutines at once.
//
// `internal/program/walk.go` runs up to 16 workers striding over files with one mutex only for the
// final merge, so the cache is reached concurrently by construction. This is the assertion that
// holds under `-race`; without it the mutex is a claim rather than a tested property.
// Not parallel: it resets the package designSystemCache and asserts the process-wide
// tailwindengine.BuildsSoFar moved by exactly one across its 64 goroutines, which any parallel test
// building a design system would inflate.
func TestDesignSystemIsSafeUnderTheParallelWalk(t *testing.T) {
	resetDesignSystemCacheForTest()

	root := designSystemProject(t, 2)
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: root})
	if err != nil {
		t.Fatalf("building the program: %v", err)
	}

	before := tailwindengine.BuildsSoFar()

	var waitGroup sync.WaitGroup
	systems := make([]*tailwindengine.LoadedDesignSystem, 64)
	for index := range systems {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			systems[index] = DesignSystemForProgram(designSystemProgram(graph.Program)).System
		}()
	}
	waitGroup.Wait()

	if built := tailwindengine.BuildsSoFar() - before; built != 1 {
		t.Fatalf("64 concurrent asks produced %d builds; the mutex is not covering the whole "+
			"read-modify-write", built)
	}
	for index, system := range systems {
		if system == nil {
			t.Fatalf("goroutine %d got no design system", index)
		}
		if system != systems[0] {
			t.Fatalf("goroutine %d got a different design system pointer, so two builds raced", index)
		}
	}
}

// TestDesignSystemLoadCost measures what one build costs and prints it.
//
// Printed rather than bounded. A threshold here would fail on a loaded machine and teach everyone
// to rerun it, and the number that matters is the ratio against a per-file rebuild rather than any
// absolute. What is asserted is only that the measurement happened at all, since a timing test that
// measured nothing and a fast one look identical in a green line.
func TestDesignSystemLoadCost(t *testing.T) {
	t.Parallel()
	resolve := designSystemFixtureResolver(map[string]string{
		"tailwindcss": designSystemFrameworkStylesheet,
	})
	entryPoint := designSystemWriteStylesheet(t, designSystemStylesheet)
	tokensPath := filepath.Join(filepath.Dir(entryPoint), "tokens.css")
	if err := os.WriteFile(tokensPath, []byte(designSystemTokensStylesheet), 0o644); err != nil {
		t.Fatalf("writing tokens: %v", err)
	}

	const iterations = 50
	start := time.Now()
	for range iterations {
		if _, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
			EntryPoint: entryPoint,
			Resolve:    resolve,
		}); err != nil {
			t.Fatalf("loading: %v", err)
		}
	}
	perBuild := time.Since(start) / iterations

	if perBuild <= 0 {
		t.Fatal("the measurement came back at zero, so nothing was actually built")
	}
	t.Logf("one design system build: %v (%d iterations). Per-file across 1,862 files that would be %v",
		perBuild, iterations, perBuild*1862)
}

// designSystemWriteStylesheet lays one stylesheet in a temporary directory and returns its path.
func designSystemWriteStylesheet(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "theme.css")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

// designSystemFixtureResolver resolves named specifiers to stylesheets written on demand.
//
// Lets a test supply an in-memory `@import` graph without laying out a node_modules tree, which
// keeps the failure cases above readable. A specifier this map does not name resolves relative to
// the importing file, matching NodeStylesheetResolver's default arm.
func designSystemFixtureResolver(bySpecifier map[string]string) tailwindengine.StylesheetResolver {
	directory := ""
	return func(specifier, base string) (string, error) {
		contents, isNamed := bySpecifier[specifier]
		if !isNamed {
			path := filepath.Join(base, specifier)
			if !strings.HasSuffix(path, ".css") {
				path += ".css"
			}
			return path, nil
		}

		if directory == "" {
			created, err := os.MkdirTemp("", "design-system-fixture")
			if err != nil {
				return "", err
			}
			directory = created
		}
		path := filepath.Join(directory, strings.ReplaceAll(specifier, "/", "-")+".css")
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return "", err
		}
		return path, nil
	}
}
