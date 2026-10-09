package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The generated framework registry, tested against the engine's own measurements.
//
// Two claims are separable here and both are tested. The first is that the checked-in table says
// what the engine says, which is a differential against `variant_fixtures.json` — the same fixture
// `variant_test.go` uses, captured by a different generator, so agreement between them is agreement
// between two independent readings of the engine rather than a tool checking itself. The second is
// that `LoadDesignSystem` wires the table in so that the behaviour downstream changes, which is
// what the blocker on #4q5dsn3 measured as broken and is not implied by the first.

// frameworkVariantFixtureEntry mirrors one registry entry in variant_fixtures.json.
type frameworkVariantFixtureEntry struct {
	Name  string            `json:"name"`
	Order int               `json:"order"`
	Kind  ParsedVariantKind `json:"kind"`
}

// frameworkVariantFixture is the slice of variant_fixtures.json this file reads.
//
// Only the registry is read. The behavioural half of that fixture belongs to variant_test.go, and
// reading it here would mean two files asserting the same thing and drifting apart about which one
// owns it.
type frameworkVariantFixture struct {
	TailwindVersion string `json:"tailwindVersion"`
	Cases           []struct {
		Name     string `json:"name"`
		Registry *struct {
			Entries         []frameworkVariantFixtureEntry `json:"entries"`
			CompareFnOrders []int                          `json:"compareFnOrders"`
		} `json:"registry"`
	} `json:"cases"`
}

func loadFrameworkVariantFixture(t *testing.T) frameworkVariantFixture {
	t.Helper()

	path := filepath.Join("testdata", "variant_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var fixture frameworkVariantFixture
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if fixture.TailwindVersion != TailwindVersion {
		t.Fatalf(
			"the fixture was captured from Tailwind %s and this package targets %s",
			fixture.TailwindVersion, TailwindVersion,
		)
	}
	return fixture
}

// frameworkVariantFixtureRegistry returns one named case's registry.
func frameworkVariantFixtureRegistry(t *testing.T, name string) ([]frameworkVariantFixtureEntry, []int) {
	t.Helper()

	fixture := loadFrameworkVariantFixture(t)
	for _, aCase := range fixture.Cases {
		if aCase.Name == name && aCase.Registry != nil {
			return aCase.Registry.Entries, aCase.Registry.CompareFnOrders
		}
	}
	t.Fatalf("the fixture holds no registry for %q", name)
	return nil, nil
}

// TestFrameworkVariantTableMatchesTheEngine is the differential: every registration the engine
// reported, reproduced by the checked-in table.
//
// Names, order numbers and kinds are all compared. Comparing names alone would pass a table that
// had every root and had renumbered them, which is the failure that costs the most: the order is a
// sort key, and a renumbering that preserved relative sequence would even look right until it hit
// the two shared orders.
func TestFrameworkVariantTableMatchesTheEngine(t *testing.T) {
	t.Parallel()
	entries, _ := frameworkVariantFixtureRegistry(t, "framework-default")

	if len(entries) == 0 {
		t.Fatal("the fixture's framework registry is empty, so this differential would pass vacuously")
	}
	if len(FrameworkVariantRegistrations) != len(entries) {
		t.Fatalf(
			"the table holds %d registrations and the engine reported %d",
			len(FrameworkVariantRegistrations), len(entries),
		)
	}

	generated := make(map[string]FrameworkVariantRegistration, len(FrameworkVariantRegistrations))
	for _, registration := range FrameworkVariantRegistrations {
		if _, isDuplicate := generated[registration.Name]; isDuplicate {
			t.Fatalf("%q appears twice in the table, so one of its two orders is unreachable", registration.Name)
		}
		generated[registration.Name] = registration
	}

	for _, entry := range entries {
		registration, isPresent := generated[entry.Name]
		if !isPresent {
			t.Errorf("the engine registers %q and the table does not, so `%s:` would not parse", entry.Name, entry.Name)
			continue
		}
		if registration.Order != entry.Order {
			t.Errorf("%q holds order %d in the table and %d on the engine", entry.Name, registration.Order, entry.Order)
		}
		if registration.Kind != entry.Kind {
			t.Errorf("%q is %q in the table and %q on the engine", entry.Name, registration.Kind, entry.Kind)
		}
	}

	t.Logf("tailwind %s: %d framework registrations compared, names, orders and kinds", TailwindVersion, len(entries))
}

// TestFrameworkVariantComparisonGroupsMatchTheEngine pins which orders carry a comparison function.
//
// A missing group is the quietest failure in this component. The registrations would still be
// complete, every class would still parse, and the six roots sharing order 64 would fall through to
// `compare`'s root-name branch, where `2xl` precedes `sm` alphabetically. Every responsive class in
// the tree would sort wrongly and nothing would be missing from the table a reader inspected.
func TestFrameworkVariantComparisonGroupsMatchTheEngine(t *testing.T) {
	t.Parallel()
	entries, compareFnOrders := frameworkVariantFixtureRegistry(t, "framework-default")

	if len(compareFnOrders) == 0 {
		t.Fatal("the fixture reports no comparison orders, so this test would pass vacuously")
	}
	if len(FrameworkVariantComparisonGroups) != len(compareFnOrders) {
		t.Fatalf(
			"the table holds %d comparison groups and the engine reported %d",
			len(FrameworkVariantComparisonGroups), len(compareFnOrders),
		)
	}

	generated := make(map[int]FrameworkVariantComparisonGroup, len(FrameworkVariantComparisonGroups))
	for _, group := range FrameworkVariantComparisonGroups {
		generated[group.Order] = group
	}
	for _, order := range compareFnOrders {
		if _, isPresent := generated[order]; !isPresent {
			t.Errorf("the engine registers a comparison at order %d and the table does not", order)
		}
	}

	// The members recorded against a group must be exactly the roots holding that order, or the
	// group is describing a set the registrations do not contain and a reader inspecting it would
	// be misled about which variants the comparison reaches.
	byOrder := map[int][]string{}
	for _, entry := range entries {
		byOrder[entry.Order] = append(byOrder[entry.Order], entry.Name)
	}
	for _, group := range FrameworkVariantComparisonGroups {
		holders := map[string]bool{}
		for _, name := range byOrder[group.Order] {
			holders[name] = true
		}
		if len(holders) != len(group.Members) {
			t.Errorf(
				"order %d is held by %d roots on the engine and the group records %d members",
				group.Order, len(holders), len(group.Members),
			)
		}
		for _, member := range group.Members {
			if !holders[member] {
				t.Errorf("order %d records member %q, which does not hold that order on the engine", group.Order, member)
			}
		}
	}
}

// TestFrameworkVariantTableHasSharedOrders is the guard against a table that looks right and was
// renumbered.
//
// This is the mutation the differential above would catch and this one names. Replaying 88
// registrations through `Register` in sequence produces 88 distinct orders, every relative position
// preserved, every name present. The only visible difference is that no two roots share a number,
// and a shared number is the sole trigger for a comparison function. So the whole failure is
// contained in this one count, which is why it is asserted directly rather than left implied.
func TestFrameworkVariantTableHasSharedOrders(t *testing.T) {
	t.Parallel()
	byOrder := map[int][]string{}
	for _, registration := range FrameworkVariantRegistrations {
		byOrder[registration.Order] = append(byOrder[registration.Order], registration.Name)
	}

	shared := 0
	for _, names := range byOrder {
		if len(names) > 1 {
			shared++
		}
	}
	if shared == 0 {
		t.Fatalf(
			"every one of the %d registrations holds its own order, so no comparison function is ever "+
				"consulted and the breakpoints order alphabetically: `2xl` before `sm`",
			len(FrameworkVariantRegistrations),
		)
	}
	if len(byOrder) >= len(FrameworkVariantRegistrations) {
		t.Fatalf(
			"%d registrations over %d distinct orders: the table was renumbered rather than replayed",
			len(FrameworkVariantRegistrations), len(byOrder),
		)
	}

	// Every comparison group must sit on an order more than one root holds, or the comparison has
	// nothing to separate. `max` and `@max` are the exception and are functional roots whose
	// *values* separate, so they legitimately hold an order alone.
	for _, group := range FrameworkVariantComparisonGroups {
		if len(byOrder[group.Order]) == 0 {
			t.Errorf("comparison group at order %d has no registrations holding that order", group.Order)
		}
	}

	t.Logf("%d registrations over %d distinct orders, %d shared", len(FrameworkVariantRegistrations), len(byOrder), shared)
}

// TestRegisterFrameworkVariantsPreservesSharedOrders pins the replay method against the sequential
// one it exists to replace.
//
// Both paths produce a registry holding every name. Only the replay produces one where `sm` and
// `md` share a number, and that difference is invisible in any assertion about membership.
func TestRegisterFrameworkVariantsPreservesSharedOrders(t *testing.T) {
	t.Parallel()
	replayed := NewVariantRegistry()
	replayed.RegisterFrameworkVariants(FrameworkVariantRegistrations)

	sequential := NewVariantRegistry()
	for _, registration := range FrameworkVariantRegistrations {
		sequential.Register(registration.Name, registration.Kind)
	}

	if len(replayed.Registrations()) != len(FrameworkVariantRegistrations) {
		t.Fatalf("replaying registered %d of %d", len(replayed.Registrations()), len(FrameworkVariantRegistrations))
	}

	for _, registration := range FrameworkVariantRegistrations {
		got, isRegistered := replayed.Get(registration.Name)
		if !isRegistered {
			t.Fatalf("%q was not registered by the replay", registration.Name)
		}
		if got.Order != registration.Order {
			t.Errorf("%q replayed at order %d, table says %d", registration.Name, got.Order, registration.Order)
		}
		if got.Kind != registration.Kind {
			t.Errorf("%q replayed as %q, table says %q", registration.Name, got.Kind, registration.Kind)
		}
	}

	// The known-dirty control: the sequential path is what a reader would reach for, and it is
	// wrong in exactly one way. Asserting that it *is* wrong keeps the replay's value stated rather
	// than assumed, and would fail loudly if Register ever started sharing orders.
	small, _ := sequential.Get("sm")
	medium, _ := sequential.Get("md")
	if small.Order == medium.Order {
		t.Fatal("sequential registration shared an order, so this control no longer distinguishes the two paths")
	}
	replayedSmall, _ := replayed.Get("sm")
	replayedMedium, _ := replayed.Get("md")
	if replayedSmall.Order != replayedMedium.Order {
		t.Errorf(
			"`sm` and `md` should share one order after a replay, got %d and %d",
			replayedSmall.Order, replayedMedium.Order,
		)
	}
}

// TestRegisterFrameworkVariantsLeavesRoomForCustomVariants pins the append point.
//
// A repository's `@custom-variant` under a new name registers after the framework's and must land
// past the framework's highest order, or it collides with a framework variant and the two share a
// position they were never measured to share. The replay advances `lastOrder` for exactly this
// reason, and nothing else in the package would notice if it stopped.
func TestRegisterFrameworkVariantsLeavesRoomForCustomVariants(t *testing.T) {
	t.Parallel()
	registry := NewVariantRegistry()
	registry.RegisterFrameworkVariants(FrameworkVariantRegistrations)

	highest := 0
	for _, registration := range FrameworkVariantRegistrations {
		if registration.Order > highest {
			highest = registration.Order
		}
	}

	registry.Register("sidebar", ParsedVariantKindStatic)
	sidebar, isRegistered := registry.Get("sidebar")
	if !isRegistered {
		t.Fatal("a custom variant should register after a replay")
	}
	if sidebar.Order <= highest {
		t.Fatalf(
			"a new `@custom-variant` took order %d, at or below the framework's highest at %d, so it "+
				"collides with a framework variant instead of appending past it",
			sidebar.Order, highest,
		)
	}

	// Re-registering a framework name keeps its order, which is what makes `@custom-variant dark`
	// restyle `dark:` without moving where it sorts. Both corpus repositories do exactly this.
	dark, _ := registry.Get("dark")
	registry.Register("dark", ParsedVariantKindStatic)
	darkAgain, _ := registry.Get("dark")
	if darkAgain.Order != dark.Order {
		t.Errorf("re-registering `dark` moved it from order %d to %d", dark.Order, darkAgain.Order)
	}
}

// frameworkVariantStylesheet writes a stylesheet plus a stand-in tailwindcss package, and returns
// the entry point and package root.
//
// A real `@theme` block in a real file rather than an in-memory graph, because what is under test
// is the theme reaching the comparison, and a resolver stub that returned canned text would let the
// theme half be mocked out of the very path being measured.
func frameworkVariantStylesheet(t *testing.T, entry string) (string, string) {
	t.Helper()

	directory := t.TempDir()
	packageRoot := filepath.Join(directory, "node_modules", "tailwindcss")
	if err := os.MkdirAll(packageRoot, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// The framework's own breakpoint and container scales, which is what `@import "tailwindcss"`
	// contributes and what an unmodified repository inherits.
	framework := `@theme {
  --breakpoint-sm: 40rem;
  --breakpoint-md: 48rem;
  --breakpoint-lg: 64rem;
  --breakpoint-xl: 80rem;
  --breakpoint-2xl: 96rem;
  --container-sm: 24rem;
  --container-md: 28rem;
  --container-lg: 32rem;
}
`
	if err := os.WriteFile(filepath.Join(packageRoot, "index.css"), []byte(framework), 0o644); err != nil {
		t.Fatalf("write framework: %v", err)
	}

	entryPoint := filepath.Join(directory, "theme.css")
	if err := os.WriteFile(entryPoint, []byte(entry), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	return entryPoint, packageRoot
}

// frameworkVariantIndices builds a run order over static variant roots and returns their indices.
func frameworkVariantIndices(t *testing.T, system *LoadedDesignSystem, roots []string) map[string]int {
	t.Helper()

	variants := make([]ParsedVariant, 0, len(roots))
	for _, root := range roots {
		variants = append(variants, ParsedVariant{Kind: ParsedVariantKindStatic, Root: root})
	}
	order := system.Variants().BuildVariantOrder(variants)

	indices := map[string]int{}
	for _, variant := range variants {
		index, isPlaced := order.IndexOf(variant)
		if !isPlaced {
			t.Fatalf("%q was not placed in the run order", variant.Root)
		}
		indices[variant.Root] = index
	}
	return indices
}

// TestLoadedDesignSystemRegistersTheFrameworkVariants is the wiring test, and it is the one the
// blocker on #4q5dsn3 would have failed.
//
// The registry differential above passes on a table nobody passes to `LoadDesignSystem`. That was
// the measured state before this landed: 88 registrations sitting in testdata, one registration in
// the loaded system, and every variant-bearing class returning zero candidates from
// `ParseCandidate`. So the count and the parse are asserted through a loaded system rather than
// through the table.
func TestLoadedDesignSystemRegistersTheFrameworkVariants(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@custom-variant dark (&:where(.dark *));
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if got := len(system.Variants().Registrations()); got != len(FrameworkVariantRegistrations) {
		t.Fatalf(
			"a loaded system holds %d registrations and the framework table holds %d: the table is not "+
				"reaching the registry, which is the state that made every variant-bearing class unparseable",
			got, len(FrameworkVariantRegistrations),
		)
	}

	// `dark` is the repository's own redefinition and must not have appended: it is already a
	// framework name. Its presence proves nothing on its own, which is why `hover` is checked
	// beside it — before this landed `dark` was the only variant a loaded system knew.
	for _, root := range []string{"hover", "focus", "md", "group", "peer", "dark", "@max"} {
		if !system.HasVariant(root) {
			t.Errorf("a loaded system does not know %q, so `%s:` classes would not parse", root, root)
		}
	}

	contributions := system.Contributions()
	for _, name := range contributions.NewVariantNames {
		if name == "dark" {
			t.Error("`dark` was reported as a new repository variant; it redefines a framework registration in place")
		}
	}
}

// TestLoadedDesignSystemParsesVariantBearingClasses measures the consequence rather than the cause.
//
// `ParseVariant` returns nil for an unregistered root and `ParseCandidate` then yields zero
// candidates for the whole class, so a missing registration is not a ranking difference, it is a
// class the engine reads and cohere cannot. This asserts the population directly, because a rule
// reading a system that cannot parse would decline on every one of these and a differential would
// score the silence as agreement.
func TestLoadedDesignSystemParsesVariantBearingClasses(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@custom-variant dark (&:where(.dark *));
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	// Every one of these was measured unparseable against the real ahra design system before this
	// landed, and they are the failures the blocker report named rather than a fresh selection.
	classes := []string{
		"hover:flex", "md:flex", "focus:flex", "group-hover:disabled:flex",
		"dark:hover:flex", "sm:hover:flex", "2xl:flex", "max-sm:flex",
		"@lg:flex", "not-hover:flex", "has-checked:flex", "print:flex",
	}
	var unparsed []string
	for _, class := range classes {
		if len(ParseCandidate(class, system)) == 0 {
			unparsed = append(unparsed, class)
		}
	}
	if len(unparsed) > 0 {
		t.Errorf("%d of %d variant-bearing classes did not parse: %v", len(unparsed), len(classes), unparsed)
	}
}

// TestFrameworkBreakpointsOrderByResolvedWidth pins the comparison functions.
//
// Without them the six roots sharing order 64 fall through to `compare`'s root-name branch, and
// this is the assertion that catches it: alphabetically `2xl` precedes `sm`, so a registry with the
// registrations and without the comparisons produces exactly the reverse of what the engine does at
// the low end.
func TestFrameworkBreakpointsOrderByResolvedWidth(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	indices := frameworkVariantIndices(t, system, []string{"sm", "md", "lg", "xl", "2xl"})
	// Measured on the engine: sm 0, md 1, lg 2, xl 3, 2xl 4, for a run that parsed exactly these
	// five. Ascending by resolved width, not by name.
	for _, ordered := range [][2]string{{"sm", "md"}, {"md", "lg"}, {"lg", "xl"}, {"xl", "2xl"}} {
		if indices[ordered[0]] >= indices[ordered[1]] {
			t.Errorf(
				"`%s` should precede `%s`, got indices %d and %d: the breakpoints are ordering by root "+
					"name rather than by resolved width",
				ordered[0], ordered[1], indices[ordered[0]], indices[ordered[1]],
			)
		}
	}
	if indices["2xl"] < indices["sm"] {
		t.Error("`2xl` sorted before `sm`, which is alphabetical order: no comparison function was consulted")
	}
}

// TestBreakpointComparisonReadsTheLiveTheme is why the comparison is not generated.
//
// A repository that redefines `--breakpoint-sm` reorders its own responsive classes, and the engine
// does exactly that: measured, redefining it to `200rem` moves `sm` from dense index 0 to index 4.
// A comparison closed over a hardcoded default scale passes every other test in this file and fails
// only here, on the one case that is the whole reason the port exists.
func TestBreakpointComparisonReadsTheLiveTheme(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@theme {
  --breakpoint-sm: 200rem;
}
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	indices := frameworkVariantIndices(t, system, []string{"sm", "md", "lg", "xl", "2xl"})
	// The engine's answer for this exact stylesheet: md 0, lg 1, xl 2, 2xl 3, sm 4.
	expected := map[string]int{"md": 0, "lg": 1, "xl": 2, "2xl": 3, "sm": 4}
	for root, want := range expected {
		if indices[root] != want {
			t.Errorf(
				"with `--breakpoint-sm: 200rem`, `%s` should take index %d and took %d: the comparison is "+
					"reading a hardcoded scale rather than this repository's theme",
				root, want, indices[root],
			)
		}
	}
}

// TestContainerQueriesReadTheContainerNamespace pins the namespace split.
//
// `sm` is 40rem as a breakpoint and 24rem as a container, so resolving `@sm` through `--breakpoint`
// returns a real number and orders the container queries against the wrong scale. That failure is
// invisible on the default scales, where both namespaces happen to ascend in the same name order,
// which is why this fixture gives the two namespaces opposite orderings.
func TestContainerQueriesReadTheContainerNamespace(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@theme {
  --container-sm: 90rem;
  --container-lg: 10rem;
}
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	// `@sm` and `@lg` are functional variants over the `@` root, not static ones.
	variants := []ParsedVariant{
		{Kind: ParsedVariantKindFunctional, Root: "@", Value: &ParsedVariantValue{Kind: ParsedValueKindNamed, Value: "sm"}},
		{Kind: ParsedVariantKindFunctional, Root: "@", Value: &ParsedVariantValue{Kind: ParsedValueKindNamed, Value: "lg"}},
	}
	order := system.Variants().BuildVariantOrder(variants)
	smallIndex, _ := order.IndexOf(variants[0])
	largeIndex, _ := order.IndexOf(variants[1])

	// `--container-sm` is 90rem and `--container-lg` is 10rem, so `@lg` is the narrower one and
	// ascends first. Reading `--breakpoint` instead would give sm 40rem against lg 64rem and
	// reverse this.
	if largeIndex >= smallIndex {
		t.Errorf(
			"`@lg` resolves to 10rem and `@sm` to 90rem, so `@lg` should precede `@sm`; got indices %d "+
				"and %d, which is the `--breakpoint` scale rather than `--container`",
			largeIndex, smallIndex,
		)
	}
}

// TestThemeDefinedBreakpointSortsAmongTheFrameworkOnes is a repository contribution the framework
// table cannot hold, and this test is how it was found.
//
// The first version of this file assumed a `@theme { --breakpoint-tablet: 50rem }` registered
// nothing and only needed resolving. It failed, and asking the engine said why: a new
// `--breakpoint-*` key becomes a static variant registration, at the breakpoint group's own shared
// order, so `tablet:` parses and sorts by its width. A default build reports 88 registrations and
// this stylesheet reports 89.
//
// Neither corpus repository defines one, so the invariance measurement that justified checking the
// table in could not have surfaced this: it is invariant across the two repositories that exist and
// would break on the third. That is why the registration happens at load time from the live theme
// rather than being folded into the generated table.
func TestThemeDefinedBreakpointSortsAmongTheFrameworkOnes(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@theme {
  --breakpoint-tablet: 50rem;
}
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	// The registration itself, which is the half a resolver alone would miss.
	tablet, isRegistered := system.Variants().Get("tablet")
	if !isRegistered {
		t.Fatal("`--breakpoint-tablet` should register `tablet` as a variant, so that `tablet:flex` parses")
	}
	small, _ := system.Variants().Get("sm")
	if tablet.Order != small.Order {
		t.Errorf(
			"`tablet` took order %d and `sm` holds %d: a theme breakpoint must join the breakpoint group, "+
				"or it is separated by registration rather than by width and sorts past every framework variant",
			tablet.Order, small.Order,
		)
	}
	if tablet.Kind != ParsedVariantKindStatic {
		t.Errorf("`tablet` registered as %q; the engine registers a theme breakpoint as static", tablet.Kind)
	}
	if len(ParseCandidate("tablet:flex", system)) == 0 {
		t.Error("`tablet:flex` did not parse, so the class the theme just defined is unreadable")
	}

	// And the ordering, which is what the registration is for.
	indices := frameworkVariantIndices(t, system, []string{"sm", "tablet", "lg"})
	if !(indices["sm"] < indices["tablet"] && indices["tablet"] < indices["lg"]) {
		t.Errorf(
			"`tablet` resolves to 50rem and should sort between `sm` (40rem) and `lg` (64rem); got sm=%d "+
				"tablet=%d lg=%d",
			indices["sm"], indices["tablet"], indices["lg"],
		)
	}
}

// TestThemeContainerKeysRegisterNothing is the known-dirty control for the test above.
//
// Without it, "register a variant for every theme key in a size-ish namespace" passes every
// assertion here and is wrong: measured on the engine, a stylesheet adding `--container-huge`
// reports 88 registrations, not 89. Container queries reach the theme through the already-registered
// `@` root, so `@huge:` parses without `huge` ever being a registration, and registering one would
// make the bare class `huge:flex` parse when the engine rejects it.
func TestThemeContainerKeysRegisterNothing(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@theme {
  --container-huge: 200rem;
}
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if system.HasVariant("huge") {
		t.Error("`--container-huge` should not register a `huge` variant; the engine reaches it through the `@` root")
	}
	if got := len(system.Variants().Registrations()); got != len(FrameworkVariantRegistrations) {
		t.Errorf(
			"a container-only theme produced %d registrations against the framework's %d",
			got, len(FrameworkVariantRegistrations),
		)
	}
}

// TestRedefiningAFrameworkBreakpointDoesNotRegisterAgain pins the other branch of the same walk.
//
// `--breakpoint-sm` is a name the framework already registers. Re-registering it would be harmless
// for position, since Register's update branch keeps the order, but it would count as a repository
// contribution and it would make the registration count drift by one per redefined breakpoint. The
// engine reports 88 for this stylesheet.
func TestRedefiningAFrameworkBreakpointDoesNotRegisterAgain(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";
@theme {
  --breakpoint-sm: 200rem;
}
`)

	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if got := len(system.Variants().Registrations()); got != len(FrameworkVariantRegistrations) {
		t.Errorf(
			"redefining `--breakpoint-sm` produced %d registrations against the framework's %d",
			got, len(FrameworkVariantRegistrations),
		)
	}
}

// TestExplicitFrameworkVariantsStillOverrideTheTable keeps the option honest.
//
// A non-nil slice replaces the table entirely, which is what the rules package's fixtures rely on:
// they build three-variant registries over stand-in stylesheets and would be swamped by 88. An
// explicitly empty slice is honoured rather than read as absent, because "no framework variants"
// is a state a caller can legitimately ask for and silently substituting 88 would make the loud
// wrong answer quiet again.
func TestExplicitFrameworkVariantsStillOverrideTheTable(t *testing.T) {
	t.Parallel()
	entryPoint, packageRoot := frameworkVariantStylesheet(t, `@import "tailwindcss";`)

	explicit, err := LoadDesignSystem(LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
		FrameworkVariants: []FrameworkVariant{
			{Name: "hover", Kind: ParsedVariantKindStatic},
			{Name: "focus", Kind: ParsedVariantKindStatic},
		},
	})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}
	if got := len(explicit.Variants().Registrations()); got != 2 {
		t.Errorf("an explicit list of 2 produced %d registrations", got)
	}
	if explicit.HasVariant("md") {
		t.Error("an explicit list should replace the table, not extend it")
	}

	empty, err := LoadDesignSystem(LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
		FrameworkVariants:   []FrameworkVariant{},
	})
	if err != nil {
		t.Fatalf("loading: %v", err)
	}
	if got := len(empty.Variants().Registrations()); got != 0 {
		t.Errorf("an explicitly empty list produced %d registrations, so it was read as absent", got)
	}
}
