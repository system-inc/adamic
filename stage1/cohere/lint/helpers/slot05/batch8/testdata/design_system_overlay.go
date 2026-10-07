// The seam: one repository's whole design system, assembled once and then read-only.
//
// Nine components landed before this file and each answers one question well. None of them knows
// how to be assembled, and that is deliberate: a component that reached for a repository path would
// be untestable without one. This file is where the reaching happens, exactly once, and it is the
// only file in the package that reads the filesystem outside a test.
//
// # What "run-scoped" has to mean, and why it is a test rather than a comment
//
// `LoadDesignSystem` walks an `@import` graph, parses every stylesheet in it, and fills a Theme.
// That is cheap once and ruinous per file. The precedent is in this repository and it is not
// hypothetical: `boundary-no-project-theme-value` memoized nothing and rescanned the program per
// file, measured at 2,329ms of setup against 0.54ms of listening across 1,862 files — 64.5% of all
// rule time for a rule that reported nothing. Nothing failed. It was just slow, and its own comment
// said the scan happened once.
//
// Measured here: a full load of this repository's `theme.css` graph is 0.68ms warm, 9 stylesheets
// resolving 744 theme entries and 37 `@utility` blocks. Per-file across 1,862 files that is 1.27
// seconds of pure waste, and the shape of the failure is that every test still passes. So
// `BuildCount` exists on the loaded value and
// `internal/rules/tailwind` asserts it, because a counter is the only form of the build-once claim
// that cannot decay into a lie without something turning red.
//
// # Immutable after build
//
// `internal/program/walk.go` runs up to 16 goroutines striding over files with one mutex for the
// final merge only. A DesignSystem is therefore shared across all of them with no synchronization
// at all, which is safe only because nothing here is written after LoadDesignSystem returns. Every
// table this composes is already immutable — Theme, VariantRegistry and UtilityEvaluator each say
// so in their own doc comments — and this file keeps that property rather than introducing the
// package's first mutable one. `MarkUsedVariable` is the one method on Theme that writes, and the
// Theme accessor's own comment says so rather than the type enforcing it, because Go has no way to
// hand out a read-only view of a struct with pointer methods.
//
// # What is composed, and what is honestly still generated
//
// The end state of #twgo is a design system built entirely from the repository. This is not that
// yet, and the split is named rather than blurred:
//
//	from the repository   the Theme, its prefix, every `@utility` block, every `@custom-variant`
//	                      name. All of it parsed out of the repository's own stylesheet graph by
//	                      the components that landed tonight.
//	still generated       which roots the framework itself registers, carried as `KnownRoots` and
//	                      `KnownStatics`, and their readings, carried as `baseDescriptors` in
//	                      descriptor_base_table.go. Framework-invariant, and that was re-measured
//	                      rather than inherited: byte-identical `roots` arrays across the two corpus
//	                      repositories is *not* evidence, because they share the Structure submodule
//	                      and so share its `@theme` and its `@utility` blocks. Against a third,
//	                      synthetic design system sharing neither, 258 of 301 shared roots differ,
//	                      and every one differs only in its namespace maps. With those removed all
//	                      three agree exactly, which is the seam descriptor_live.go composes across.
//	also generated        the framework's own variant registrations, carried as
//	                      `FrameworkVariantRegistrations` and keyed on registration roots. Verified
//	                      repository-invariant across two installs and both corpus repositories
//	                      before being checked in, so a table is the correct shape for it. The
//	                      comparison functions those registrations reference are *not* generated:
//	                      they resolve breakpoint widths out of the live theme, because a repository
//	                      that redefines `--breakpoint-sm` reorders its own responsive classes.
//	                      `VariantOrder` in property_order_table.go is a different table, keyed on
//	                      composed prefixes (`group-hover:`) rather than on registration roots,
//	                      which is the shape variant.go's file comment argues against.
//
// A reader should be able to see that boundary without running anything, which is why
// `RepositoryContributions` reports what the repository added on top of the framework rather than
// leaving it to be inferred.
//
// # Failing safe
//
// A design system that will not build is no opinion, never a clean tree. LoadDesignSystem returns
// an error and the caller declines; it never returns a half-built value. cohere exists to eliminate
// confident green over unchecked work, and a rule reporting zero findings because the CSS could not
// be read is exactly that failure wearing the right colour.
package tailwind

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// LoadedDesignSystem is one repository's resolved design system, immutable once built.
//
// It satisfies the DesignSystem interface in candidate.go, which is the interface ParseCandidate
// asks its four questions through. That interface is deliberately narrow and this type is
// deliberately wider: the parser needs four answers, and a rule computing a sort key needs the
// theme, the descriptor tables and the variant registry as well.
//
// Safe for concurrent reads and never written after LoadDesignSystem returns. See the file comment
// for why that is a hard requirement rather than a nicety.
type LoadedDesignSystem struct {
	// EntryPoint is the stylesheet the load started from, absolute.
	EntryPoint string

	// TailwindVersion is the engine version the generated tables were produced against.
	//
	// Carried so a mismatch between the tables and an upgraded `tailwindcss` is a loud failure
	// rather than a subtly wrong sort. Nothing verifies it yet; #4q5dsn3 is where the sort starts
	// depending on it.
	TailwindVersion string

	// Stylesheets is every file the `@import` graph reached, absolute, in the order they were read.
	//
	// Reported rather than discarded because it is the invalidation set. A findings cache keyed on
	// a `.tsx` file's hash is stale the moment any of these changes, and a caller that cannot see
	// the list cannot know that. It is also what makes a load auditable: a repository whose theme
	// looks wrong is usually a repository whose graph reached a file nobody expected.
	Stylesheets []string

	// SkippedDirectives is every `@config` and `@plugin` the graph held.
	//
	// A `@config` can contribute theme values through JavaScript, which cohere does not run, so a
	// non-empty list means the theme may be short by however many keys that config declared. Kept
	// so a caller can decide, rather than dropped so a caller cannot.
	SkippedDirectives []SkippedDirective

	// BuildCount is how many times a design system has been built in this process.
	//
	// The whole point of this field is that it is observable from a test. See the file comment: the
	// per-file rebuild is a failure mode with no symptom other than time, and the rule that already
	// suffered it had a comment claiming it did not. A comment cannot fail; this can.
	BuildCount int

	theme    *Theme
	variants *VariantRegistry
	utility  *UtilityEvaluator

	// staticUtilityNodes is the body of every static `@utility` block the repository declared,
	// keyed by root.
	//
	// Kept because a static utility's reading is a constant PropertySort answers from its body, and
	// there is nowhere else to get it: the framework's static table does not know a name the
	// repository invented. This is the whole of what NewTable adds to `Statics`.
	staticUtilityNodes map[string][]*Node
	// utilityRoots is every `@utility` root the repository declared, mapped to the kinds it was
	// declared as, so HasUtility is a map lookup rather than a walk.
	//
	// A set of kinds rather than one kind, because a root can be declared both ways and sixteen in
	// this repository are: `@utility fade-in` and `@utility fade-in-*` are two blocks naming one
	// root, a static form taking no value and a functional form taking one. Storing a single kind
	// let the second block overwrite the first, which lost the static form silently and left
	// `fade-in` answerable by neither the table nor the evaluator.
	utilityRoots map[string]map[UtilityKind]bool
	// customVariants is every `@custom-variant` name the repository declared.
	customVariants map[string]bool
	// frameworkVariants is the set of names the framework registered, so Contributions can tell an
	// appended registration from an in-place update without re-deriving the split.
	frameworkVariants map[string]bool
}

// A LoadedDesignSystem is what ParseCandidate asks its four questions through, so a signature that
// drifts from the interface must be a compile error rather than something the first rule to wire in
// discovers. Cheap to state and it fails at exactly the right moment: nothing calls ParseCandidate
// with one of these yet, so without this the drift would sit undetected until #4q5dsn3.
var _ DesignSystem = (*LoadedDesignSystem)(nil)

// Theme is the resolved theme, for reading.
//
// Returned by value-of-pointer because Theme's read methods are all on the pointer, and copying it
// would copy a map header rather than the map. The contract is the doc comment rather than the
// type: a caller that calls Add or ClearNamespace on this has broken the run for every other
// goroutine, and Go has no way to say so in the signature.
//
// `MarkUsedVariable` is the one read-shaped method that writes, and it is the reason this is named
// rather than embedded: a caller reaching for it has to type the reach and can be found by grep.
func (system *LoadedDesignSystem) Theme() *Theme { return system.theme }

// Variants is the variant registry, for reading.
func (system *LoadedDesignSystem) Variants() *VariantRegistry { return system.variants }

// Utilities is the `@utility` evaluator over this repository's own blocks.
//
// Nil when the repository declared no `@utility` blocks, which is a real state rather than a
// failure: most repositories declare none, and a caller must handle the nil rather than assume a
// block exists to evaluate.
func (system *LoadedDesignSystem) Utilities() *UtilityEvaluator { return system.utility }

// Prefix is the configured utility prefix, empty when there is none.
//
// Part of the DesignSystem interface candidate.go declares.
func (system *LoadedDesignSystem) Prefix() string { return system.theme.Prefix }

// HasUtility reports whether root is registered as a utility of the given kind.
//
// The repository's own `@utility` blocks are asked first, then the framework's registry. That order
// is upstream's: `@utility` registration runs after the framework's and `Utilities.set` overwrites,
// so a repository redefining a framework root wins. Both are consulted because neither alone is
// right — the framework tables do not know about `slide-in-from-top`, and the repository's blocks
// do not know about `bg`.
//
// A repository `@utility` block is always functional or static according to whether its name
// carries the `-*` suffix, which is what collectStylesheets recorded when it parsed the block.
func (system *LoadedDesignSystem) HasUtility(root string, kind UtilityKind) bool {
	if declared, isDeclared := system.utilityRoots[root]; isDeclared {
		return declared[kind]
	}
	switch kind {
	case UtilityKindStatic:
		_, registered := FrameworkStaticDeclarations[root]
		return registered
	case UtilityKindFunctional:
		return frameworkFunctionalRoots[root]
	default:
		return false
	}
}

// frameworkFunctionalRoots is every functional root the framework registers.
//
// Built once from the ported registration tables plus the descriptor rows that carry the roots no
// registration shape expresses, which is the same union `Lookup` answers from. It replaces
// `KnownRoots` in collapse_table.go, and the reason is that the generated table was not describing
// Tailwind.
//
// Measured: `KnownStatics` held 895 names against the 890 the framework registers, and the 27 it
// carried beyond them are ahra's own `@utility` blocks. `markdown-content`, `prose`,
// `scrollbar-hide`, the `slide-in-from-*` family, `fade-in` and `zoom-in`. `fade-in` and `zoom-in`
// are declared by ahra and not by www-connected-app, so a table under a "Source: Tailwind 4.3.3"
// header was telling every other repository that two of ahra's animations were framework utilities.
//
// The repository's own blocks are answered above this switch, from `system.utilityRoots`, so nothing
// is lost by dropping them here: a repository that declares `prose` still gets it, and one that does
// not no longer inherits ahra's.
var frameworkFunctionalRoots = func() map[string]bool {
	roots := make(map[string]bool, len(baseDescriptors)+len(FrameworkFunctionalUtilities)+len(FrameworkMultiDeclarationUtilities))
	for root := range baseDescriptors {
		roots[root] = true
	}
	// A root that supports negation registers under both names, which is what upstream does.
	//
	// `functionalUtility` calls `utilities.functional('-' + classRoot, ...)` when `supportsNegative`
	// is set, so `-m` is a registered root rather than a sign applied to `m`. The parser reads
	// `-skew-3` as root `-skew`, and without the dashed name here it does not parse at all.
	for root, utility := range FrameworkFunctionalUtilities {
		roots[root] = true
		if utility.SupportsNegative {
			roots["-"+root] = true
		}
	}
	for root, utility := range FrameworkMultiDeclarationUtilities {
		roots[root] = true
		if utility.SupportsNegative {
			roots["-"+root] = true
		}
	}
	return roots
}()

// RepositoryStaticUtilityNames is every static utility the repository itself declares.
//
// Exposed for the upstream check, which compares the installed engine's registry against the
// checked-in framework tables and must subtract the repository's own blocks first. Reading them from
// `utilityRoots` rather than from the engine is what makes that subtraction exact: this is the set
// `collectStylesheets` recorded while parsing the repository's `@utility` blocks, so it cannot
// accidentally include a framework name.
func (system *LoadedDesignSystem) RepositoryStaticUtilityNames() []string {
	names := make([]string, 0, len(system.staticUtilityNodes))
	for name, kinds := range system.utilityRoots {
		if kinds[UtilityKindStatic] {
			names = append(names, name)
		}
	}
	return names
}

// DeclaresFunctionalUtility reports whether the repository itself declared this functional root.
//
// Narrower than `HasUtility`, and the difference is the whole point: `HasUtility` answers whether a
// root exists at all, framework roots included, so it says yes to `text` and `border`. This answers
// only what `collectStylesheets` recorded while parsing the repository's own `@utility` blocks.
//
// Needed by any caller that must treat a repository root differently from a framework one.
// `ClassValueResolvesIn` is the case: a framework root has a checked-in description saying how its
// values resolve, and a repository root has an `@utility` block the evaluator compiles instead, so
// asking a framework description about `shadow--3` would report a working class.
func (system *LoadedDesignSystem) DeclaresFunctionalUtility(root string) bool {
	return system.utilityRoots[root][UtilityKindFunctional]
}

// HasVariant reports whether root is a registered variant.
//
// The registry holds the framework's registrations and every `@custom-variant` the repository
// added, because LoadDesignSystem registered both into it. A `@custom-variant` reusing a framework
// name does not appear twice: `Register` updates in place and leaves the order untouched, which is
// upstream's behaviour and the reason both corpus repositories have registries byte-identical to
// the framework's despite redefining `dark`.
func (system *LoadedDesignSystem) HasVariant(root string) bool {
	return system.variants.Has(root)
}

// VariantKind returns the registration kind of a variant root.
//
// Only called for roots HasVariant accepted, per the interface's contract. An unregistered root
// answers static rather than panicking, because a parser asking about a root it was told exists is
// a bug in the caller and a panic inside a lint run costs a file rather than surfacing it.
func (system *LoadedDesignSystem) VariantKind(root string) ParsedVariantKind {
	registration, isRegistered := system.variants.Get(root)
	if !isRegistered {
		return ParsedVariantKindStatic
	}
	return registration.Kind
}

// VariantCompoundsWith reports whether the compound variant parent accepts the rules child
// produces.
//
// **The one method here that approximates rather than composes, and the child argument is ignored.**
// Upstream answers this from a bitwise-and of two `compounds` bitmasks: the parent's, assigned at
// registration in `variants.ts`, and the child's, which for an arbitrary variant is computed from
// its selector instead of looked up. Neither half is reachable from shipped Go. The masks are not in
// any generated table, and `compoundsForSelectors` — the selector computation — is ported only
// inside `candidate_test.go`, where the shipped path cannot call it.
//
// So this answers the parent's half alone: a compound variant accepts a child, anything else
// accepts nothing. That is upstream's answer whenever the child produces style rules, which is the
// overwhelming majority, and it is wrong in one direction for one shape — a compound whose parent
// accepts only style rules wrapping an at-rule-only child, `group-[@media(print)]` and its kin,
// where upstream refuses and this accepts.
//
// Written out rather than left as a bare `return true` because a caller should find the gap by
// reading the method, not by chasing a wrong sort. Closing it needs a `compounds` column on the
// variant table and `compoundsForSelectors` moved into the package proper, both of which are
// variant work rather than seam work. Nothing reads this yet: #4q5dsn3 and #3r6cxrb are where rules
// begin to.
func (system *LoadedDesignSystem) VariantCompoundsWith(parent string, _ ParsedVariant) bool {
	registration, isRegistered := system.variants.Get(parent)
	if !isRegistered {
		return false
	}
	// A compound variant is the only kind that can wrap another at all.
	return registration.Kind == ParsedVariantKindCompound
}

// RepositoryContributions is what this repository added on top of the framework.
//
// Exists so the boundary in the file comment is measurable rather than described. A repository
// whose contributions come back empty is a repository whose theme.css did not resolve the way its
// author thought, and that is indistinguishable from a correct load in every other observable.
// #j2aac50 is the task that asserts these are non-empty on our own tree.
type RepositoryContributions struct {
	// ThemeEntries is how many keys the resolved theme holds in total.
	ThemeEntries int
	// UtilityBlocks is how many `@utility` blocks the graph declared.
	UtilityBlocks int
	// CustomVariants is how many `@custom-variant` names the graph declared.
	CustomVariants int
	// NewVariantNames is the `@custom-variant` names that were not already framework variants, so
	// they appended a registration order rather than updating one in place.
	NewVariantNames []string
	// StylesheetCount is how many files the `@import` graph reached.
	StylesheetCount int
}

// Contributions reports what the repository added, for tests and for `--timing`-style reporting.
func (system *LoadedDesignSystem) Contributions() RepositoryContributions {
	newNames := []string{}
	for name := range system.customVariants {
		if !system.frameworkVariants[name] {
			newNames = append(newNames, name)
		}
	}
	sort.Strings(newNames)

	return RepositoryContributions{
		ThemeEntries:    system.theme.Size(),
		UtilityBlocks:   len(system.utilityRoots),
		CustomVariants:  len(system.customVariants),
		NewVariantNames: newNames,
		StylesheetCount: len(system.Stylesheets),
	}
}

// FrameworkVariant is one variant the framework itself registers, before the repository's
// `@custom-variant` blocks are applied.
//
// A separate type from VariantRegistration even though the fields coincide, because this is an
// *input* and that is a *record*. VariantRegistration carries an Order the registry assigned;
// these carry no order at all, since the registry assigns it by registration sequence and taking
// one from the caller would let a caller produce a registry the engine could never build.
type FrameworkVariant struct {
	// Name is the registered root, as written before the colon: `hover`, `group`, `@max`.
	Name string
	// Kind is the registration kind, which decides which branch of Compare applies.
	Kind ParsedVariantKind
}

// LoadOptions is where a design system is loaded from.
type LoadOptions struct {
	// EntryPoint is the repository's root stylesheet, the file that `@import`s tailwindcss.
	EntryPoint string
	// TailwindPackageRoot is the installed `tailwindcss` package directory, which is what
	// `@import "tailwindcss"` resolves against.
	TailwindPackageRoot string
	// Resolve overrides import resolution entirely. Nil means NodeStylesheetResolver over
	// TailwindPackageRoot, which is what a real repository wants; a test that wants an in-memory
	// graph supplies its own.
	Resolve StylesheetResolver

	// FrameworkVariants overrides the framework's variant registrations with a caller-supplied set.
	//
	// Nil, which is what every shipped caller passes, means the generated
	// FrameworkVariantRegistrations: the installed Tailwind's own 88 roots with the order numbers
	// the engine assigned them, plus the four comparison groups that order breakpoints by their
	// resolved widths. That table was verified repository-invariant before being checked in; see
	// framework_variant_table.go's header for the measurement.
	//
	// A non-nil value replaces the table entirely and is registered in sequence, so each root takes
	// its own order and no comparison groups are attached. That is the right shape for a test
	// building a small registry over a fixture stylesheet, and the wrong shape for a repository:
	// sequential registration cannot express the six roots that share order 64, and splitting them
	// apart means the breakpoints never reach a comparison function and fall through to a root-name
	// comparison, where `2xl` precedes `sm`.
	//
	// An explicitly empty non-nil slice is honoured rather than treated as absent, and it is not
	// silently degraded: HasVariant then answers false for `hover`, `ParseVariant` returns nil for
	// every framework root, and `Contributions().NewVariantNames` reports every repository variant
	// as new. That is a loud wrong answer rather than a quiet one, which is what a caller asking for
	// no framework variants at all should get.
	FrameworkVariants []FrameworkVariant
}

// LoadDesignSystem builds one repository's design system from its stylesheet graph.
//
// One walk of the graph fills everything: the theme, the `@utility` blocks and the
// `@custom-variant` names all come out of the same parse rather than three. That is not only
// cheaper, it is the only way they can agree — a second walk that resolved imports differently
// would produce a theme and a utility set describing different graphs.
//
// Returns an error and no value when the graph cannot be read. Never a partial system: see the
// failing-safe note in the file comment.
func LoadDesignSystem(options LoadOptions) (*LoadedDesignSystem, error) {
	if options.EntryPoint == "" {
		return nil, fmt.Errorf("no entry point given: a design system cannot be built without a stylesheet to build it from")
	}

	resolve := options.Resolve
	if resolve == nil {
		if options.TailwindPackageRoot == "" {
			return nil, fmt.Errorf(
				"%s: no tailwindcss package root and no resolver, so `@import \"tailwindcss\"` cannot resolve",
				options.EntryPoint,
			)
		}
		resolve = adamicLoaderResolver(options.TailwindPackageRoot)
	}

	entryPoint, err := adamicLoaderAbsolute(options.EntryPoint)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", options.EntryPoint, err)
	}

	collector := &stylesheetCollector{
		theme:              adamicLoaderTheme(),
		resolve:            resolve,
		visiting:           map[string]bool{},
		staticUtilityNodes: map[string][]*Node{},
		utilityRoots:       map[string]map[UtilityKind]bool{},
		customVariants:     map[string]bool{},
	}
	if err := adamicLoaderFile(collector, entryPoint); err != nil {
		return nil, err
	}

	variants := adamicLoaderVariants()
	// The framework registers first and the repository's `@custom-variant` names after, which is
	// upstream's sequencing and the reason a redefinition keeps its original position.
	//
	// The default path replays the generated table, carrying the engine's own order numbers rather
	// than recomputing them. That distinction is load-bearing: 88 registrations hold 82 distinct
	// orders, and a shared order is the only place a comparison function is consulted. See
	// RegisterFrameworkVariants.
	frameworkNames := map[string]bool{}
	if options.FrameworkVariants == nil {
		adamicLoaderFramework(variants)
		// A `@theme { --breakpoint-tablet: ... }` registers `tablet` as a static variant on the
		// engine, joining the breakpoint group rather than appending past it. Registered before the
		// repository's `@custom-variant` names for the same reason the framework is: it is
		// breakpoint machinery rather than a stylesheet-declared variant, and it must not consume
		// an appended position.
		adamicLoaderBreakpoints(variants, collector.theme)
		adamicLoaderComparisons(variants, collector.theme)
		for _, registration := range FrameworkVariantRegistrations {
			frameworkNames[registration.Name] = true
		}
	} else {
		for _, variant := range options.FrameworkVariants {
			adamicLoaderRegister(variants, variant.Name, variant.Kind)
			frameworkNames[variant.Name] = true
		}
	}
	// The repository's `@custom-variant` names register after the framework's, which is upstream's
	// order and the reason a redefinition keeps its original position: Register updates in place
	// and never reassigns Order. Sorted so a run is reproducible — a map range would assign
	// appended orders in a different sequence every run, and those orders are a sort key.
	for _, name := range sortedKeys(collector.customVariants) {
		adamicLoaderRegister(variants, name, ParsedVariantKindStatic)
	}

	var evaluator *UtilityEvaluator
	if len(collector.utilityDefinitions) > 0 {
		evaluator = adamicLoaderEvaluator(collector.theme, collector.utilityDefinitions)
	}

	return &LoadedDesignSystem{
		EntryPoint:         entryPoint,
		TailwindVersion:    TailwindVersion,
		Stylesheets:        collector.stylesheets,
		SkippedDirectives:  collector.skipped,
		BuildCount:         adamicLoaderCount(),
		theme:              collector.theme,
		variants:           variants,
		utility:            evaluator,
		staticUtilityNodes: collector.staticUtilityNodes,
		utilityRoots:       collector.utilityRoots,
		customVariants:     collector.customVariants,
		frameworkVariants:  frameworkNames,
	}, nil
}

// builds counts how many design systems this process has built.
//
// A package-level counter under its own mutex rather than an atomic, because the read in
// LoadDesignSystem is a read-modify-write and an atomic add returning the new value would be
// equivalent but less obvious about it. Contention is nil by construction: this is incremented once
// per run in the shipped path, and the only caller that increments it in a loop is the test whose
// entire purpose is to count.
//
// Process-scoped rather than per-cache, and that is what makes it useful. A per-cache counter
// resets when a second cache is created, which is exactly the situation the run-scoped cache exists
// to get right, so a test asserting build-once could pass against a cache that rebuilt on every
// file as long as each file got its own counter.
var builds struct {
	sync.Mutex
	count int
}

func nextBuildCount() int {
	builds.Lock()
	defer builds.Unlock()
	builds.count++
	return builds.count
}

// BuildsSoFar is how many design systems this process has built.
//
// Exported for the assertion in internal/rules/tailwind. The build-once property is the one this
// component is most likely to lose silently — the sibling rule that lost it kept a comment claiming
// it had not, while costing 2,329ms — so the claim is a counter a test reads rather than a sentence
// a reader believes.
func BuildsSoFar() int {
	builds.Lock()
	defer builds.Unlock()
	return builds.count
}

// stylesheetCollector walks one `@import` graph and gathers everything a design system needs.
//
// Deliberately a second walker rather than a change to themeLoader. That type's file comment states
// its own boundary — theme only, with `@utility` and `@custom-variant` named as sibling tasks — and
// widening it would make a component that documents what it does not do into one that quietly does
// more. The duplicated part is import resolution and cycle detection, about thirty lines, and the
// alternative was a shared walker with a visitor interface that both callers would have had to be
// written against before either existed.
type stylesheetCollector struct {
	theme   *Theme
	resolve StylesheetResolver
	skipped []SkippedDirective

	// visiting is the import stack, so a cycle is an error rather than a stack overflow. Not a
	// "seen" set: a file imported twice from different places is processed twice, matching the
	// engine, and that matters because the second pass can overwrite what the first set.
	visiting map[string]bool

	stylesheets        []string
	utilityDefinitions []*UtilityDefinition
	staticUtilityNodes map[string][]*Node
	utilityRoots       map[string]map[UtilityKind]bool
	customVariants     map[string]bool
}

func (collector *stylesheetCollector) loadFile(path string) error {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", path, err)
	}

	if collector.visiting[absolutePath] {
		return fmt.Errorf("circular @import: %s imports itself", absolutePath)
	}
	collector.visiting[absolutePath] = true
	defer delete(collector.visiting, absolutePath)

	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", absolutePath, err)
	}

	nodes, err := ParseCSS(string(content))
	if err != nil {
		return fmt.Errorf("parse %s: %w", absolutePath, err)
	}

	collector.stylesheets = append(collector.stylesheets, absolutePath)
	return collector.ingest(nodes, absolutePath)
}

// ingest walks a parsed stylesheet, following imports at the point they appear.
//
// Explicit rather than a call to Walk for the same reason themeloader.go gives: a theme is
// order-sensitive, so an imported file's entries have to land between the entries written before
// and after the `@import` line, and Walk's visitor cannot express "load another file here, then
// continue" without the loader reentering itself mid-traversal.
func (collector *stylesheetCollector) ingest(nodes []*Node, path string) error {
	for _, node := range nodes {
		if node.Kind != KindAtRule {
			if node.IsContainer() {
				if err := collector.ingest(node.Nodes, path); err != nil {
					return err
				}
			}
			continue
		}

		switch node.Name {
		case "@import":
			resolved, err := collector.resolveImport(node, path)
			if err != nil {
				return err
			}
			if err := collector.loadFile(resolved); err != nil {
				return err
			}

		case "@theme":
			if err := collector.ingestThemeBlock(node, path); err != nil {
				return err
			}

		case "@utility":
			if err := collector.ingestUtilityBlock(node, path); err != nil {
				return err
			}

		case "@custom-variant":
			collector.ingestCustomVariant(node)

		case "@config", "@plugin":
			collector.skipped = append(collector.skipped, SkippedDirective{
				Name:   node.Name,
				Params: node.Params,
				Path:   path,
			})

		default:
			// `@layer theme { @theme default { ... } }` is how the framework's own index.css is
			// written, so descending into container at-rules is required rather than defensive.
			if node.IsContainer() {
				if err := collector.ingest(node.Nodes, path); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// resolveImport turns one `@import` into the path it names.
//
// Every modifier other than `source(...)` changes what the imported file means, so an unrecognized
// one is an error rather than a silently dropped qualifier — a `theme(reference)` import quietly
// treated as plain produces entries that all lack the reference bit, which is a wrong answer that
// looks exactly like a right one. Same judgement as themeloader.go's followImport, and deliberately
// the same error text so a repository hitting it finds one explanation rather than two.
func (collector *stylesheetCollector) resolveImport(node *Node, path string) (string, error) {
	parts := segment(strings.TrimSpace(node.Params), ' ')
	if len(parts) == 0 || parts[0] == "" {
		return "", fmt.Errorf("%s: @import with no specifier", path)
	}

	specifier := strings.Trim(parts[0], `"'`)

	for _, modifier := range parts[1:] {
		modifier = strings.TrimSpace(modifier)
		if modifier == "" {
			continue
		}
		if strings.HasPrefix(modifier, "source(") {
			continue
		}
		return "", fmt.Errorf(
			"%s: unsupported @import modifier %q on %q; see the boundary note in themeloader.go",
			path, modifier, specifier,
		)
	}

	resolved, err := collector.resolve(specifier, filepath.Dir(path))
	if err != nil {
		return "", fmt.Errorf("%s: resolve @import %q: %w", path, specifier, err)
	}
	return resolved, nil
}

// ingestThemeBlock records every custom property in one `@theme` block.
//
// The `@keyframes` skip and the reject-anything-else branch are upstream's, reproduced for the
// reason themeloader.go states: a stylesheet the engine rejects must not quietly produce a theme
// here, since any theme built from input the engine refuses is one no repository can be using.
func (collector *stylesheetCollector) ingestThemeBlock(node *Node, path string) error {
	options, prefix := parseThemeOptions(node.Params)

	if prefix != "" {
		if !isValidThemePrefix(prefix) {
			return fmt.Errorf(
				"%s: the prefix %q is invalid. Prefixes must be lowercase ASCII letters (a-z) only",
				path, prefix,
			)
		}
		collector.theme.Prefix = prefix
	}

	var walkErr error
	Walk(node.Nodes, func(child *Node) WalkAction {
		if child.Kind == KindAtRule && child.Name == "@keyframes" {
			return WalkSkip
		}
		if child.Kind == KindComment {
			return WalkContinue
		}
		if child.Kind == KindDeclaration && strings.HasPrefix(child.Property, "--") {
			if err := collector.theme.Add(unescapeCSSIdentifier(child.Property), child.Value, options); err != nil {
				walkErr = fmt.Errorf("%s: %w", path, err)
				return WalkStop
			}
			return WalkContinue
		}
		if child.IsContainer() {
			return WalkContinue
		}

		walkErr = fmt.Errorf(
			"%s: `@theme` blocks must only contain custom properties or `@keyframes`, found %s %q",
			path, child.Kind, child.Property,
		)
		return WalkStop
	})

	return walkErr
}

// ingestUtilityBlock records one `@utility` block as a definition the evaluator can compile.
//
// The name carries the arity. Upstream's `@utility foo-*` registers a functional utility rooted at
// `foo` and `@utility foo` registers a static one, and the `-*` suffix is the only thing that
// distinguishes them. That is why the suffix is stripped for the Name and remembered as the kind:
// UtilityDefinition.Name is documented as the root without its suffix, and HasUtility has to be
// able to answer which kind a root was declared under.
//
// A block whose name is neither shape is an error rather than a skip. Upstream throws on it, and a
// silently dropped `@utility` is a repository utility every rule then treats as an unknown class.
func (collector *stylesheetCollector) ingestUtilityBlock(node *Node, path string) error {
	name := strings.TrimSpace(node.Params)
	if name == "" {
		return fmt.Errorf("%s: `@utility` with no name", path)
	}

	kind := UtilityKindStatic
	root := name
	if strings.HasSuffix(name, "-*") {
		kind = UtilityKindFunctional
		root = strings.TrimSuffix(name, "-*")
	}
	if root == "" || strings.Contains(root, "*") {
		return fmt.Errorf(
			"%s: `@utility %s` is not a valid utility name; expected `name` or `name-*`",
			path, name,
		)
	}

	if collector.utilityRoots[root] == nil {
		collector.utilityRoots[root] = make(map[UtilityKind]bool, 2)
	}
	collector.utilityRoots[root][kind] = true
	// Only functional blocks reach the evaluator. A static `@utility` takes no value, so it has no
	// `--value()` to resolve and Compile would reject it by its own first post-condition; its
	// reading is a constant that PropertySort answers directly from the block body. Registering it
	// here anyway would put a definition in the evaluator that can never compile, which reads to a
	// caller as a defect in the evaluator rather than as a shape it does not handle.
	//
	// The body is kept either way, and for the static case that is the whole point: the constant
	// PropertySort answers is the only place a repository's own static utility can get a reading
	// from. `markdown-content` and `typing-dots` are ahra's, they are not in the framework's static
	// table, and NewTable reads them from here.
	//
	// The two kinds are recorded independently, and they must be. `@utility fade-in` and `@utility
	// fade-in-*` are both declared in this repository and they are two different utilities that
	// share a name, not a redefinition: the static reads `[]#1` and the functional resolves a value.
	// Sixteen roots here are that shape. Letting the functional block clear the static body would
	// lose the static reading entirely, which is a class an author can write going unanswered.
	//
	// `utilityRoots` records both kinds for such a root. It used to map a root to one kind, so the
	// second block overwrote the first and the static form was lost: `fade-in` read as functional
	// only, which left it answerable by neither the table nor the evaluator and unanswered on the
	// corpus.
	if kind == UtilityKindFunctional {
		collector.utilityDefinitions = append(collector.utilityDefinitions, &UtilityDefinition{
			Name:  root,
			Nodes: node.Nodes,
		})
		return nil
	}
	collector.staticUtilityNodes[root] = node.Nodes
	return nil
}

// ingestCustomVariant records one `@custom-variant` name.
//
// The name is the first space-separated word of the params: `@custom-variant dark (&:where(.dark
// *))` names `dark`. The selector is deliberately not recorded, because nothing in this port emits
// CSS and a variant's sort position does not depend on what it emits — see the file comment in
// variant.go, where both corpus repositories redefine `dark` and keep the framework's position.
func (collector *stylesheetCollector) ingestCustomVariant(node *Node) {
	parts := segment(strings.TrimSpace(node.Params), ' ')
	if len(parts) == 0 {
		return
	}
	name := strings.TrimSpace(parts[0])
	// A functional custom variant is written `@custom-variant foo-*`, and the root is the name
	// without the suffix, the same split `@utility` uses.
	name = strings.TrimSuffix(name, "-*")
	if name == "" {
		return
	}
	collector.customVariants[name] = true
}

// sortedKeys returns a map's keys in a stable order.
//
// Not a general helper reaching for a shelf: it exists because registration order is a sort key,
// and a map range would hand appended variants a different order on every run. That is the kind of
// nondeterminism that passes every test that sorts before comparing.
func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
