// Variant ordering: the position a variant holds in Tailwind's sort, and the bitmask the class
// comparator's first dimension is built from.
//
// Ported from `src/variants.ts` and the sort in `src/compile.ts` at Tailwind 4.3.3, read at the
// pinned tag rather than from the minified bundle, so a disagreement is a finding rather than
// version skew.
//
// # What was deliberately not ported, and why
//
// Almost all of it. `variants.ts` is 1,317 lines and roughly 1,000 of them build CSS: the `applyFn`
// of every registered variant, which rewrites a rule into a `@media` wrapper or a `:hover`
// selector, plus `substituteAtSlot`, `substituteAtVariant` and the selector machinery around them.
// cohere emits no CSS. What it needs from this file is the variant's sort position and nothing
// downstream of it, so the `applyFn` bodies are absent rather than stubbed. A stub would read as
// supported.
//
// Also absent: `suggest` and `getCompletions`, which exist for IntelliSense; `compoundsWith` and
// `compoundsForSelectors`, which the candidate parser owns in candidate.go and which this file
// consumes through the registry rather than reimplementing.
//
// # The position is not a property of the variant
//
// This is the finding that decides the component, and it is the opposite of what the source
// suggests. Reading `variants.ts` shows every registered variant holding an `order` number, which
// invites a static table of variant to position. cohere has shipped exactly that table, 145
// entries, generated once.
//
// `getVariantOrder` in `design-system.ts` does not read those numbers out to consumers. It sorts
// the variants that were *actually parsed during this run*, then assigns them dense indices from
// zero, and collapses ties onto a shared index:
//
//	for (let variant of variants) {
//	  if (prevVariant !== undefined && this.variants.compare(prevVariant, variant) !== 0) index++
//	  order.set(variant, index)
//	}
//
// So `hover` is index 0 in a run that parsed `hover` and `focus`, and index 1 in a run that also
// parsed `group-hover`. The registration numbers survive only as the comparison key inside
// `compare`; the index a consumer sees is a rank within one run's population. A static table cannot
// express that, and one built from a default build agrees with the engine on any corpus whose
// variant set happens to match it and diverges silently on every other. That is why this file
// names its own type RunVariantOrder and builds one per class list, rather than exporting a
// package-level map. The generated `VariantOrder` table in property_order_table.go is the shape
// this component argues against, and the two names sitting in one package is deliberate: the
// difference between them is the finding.
//
// # Registration order is a partial order, not a total one
//
// `Variants.group` assigns every variant registered inside it the *same* order number and may
// attach a comparison function to that number. In a default build `sm`, `md`, `lg`, `xl`, `2xl` and
// `min` all hold order 64, and four order numbers (63, 64, 65, 66) carry comparison functions.
// The five breakpoints therefore do not order each other by registration at all: they order by
// their resolved `--breakpoint` values, through the theme. A port that gave each breakpoint its own
// position agrees on a default theme and diverges on any repository that redefines one, which is
// the same per-repository defect this whole port exists to remove.
//
// # A `@custom-variant` that reuses a framework name does not move
//
// `Variants.set` does `Object.assign(existing, { kind, applyFn, compounds })` when the name is
// already registered, and never assigns `order`. So `@custom-variant dark (...)` replaces the
// selector `dark:` produces while leaving its sort position exactly where the framework put it.
// Both repositories in this component's corpus do precisely that, and their variant registries are
// byte-identical to the framework's despite emitting a completely different `dark:` selector. Only
// a `@custom-variant` under a *new* name appends a position, past the framework's last.
package tailwind

import (
	"math/big"
	"sort"
	"strings"
)

// VariantRegistration is what the engine records about one registered variant root.
//
// Order is upstream's `order` field. It is a comparison key rather than a position: several roots
// share one, and a shared order may carry a comparison function that separates them. See the file
// comment.
type VariantRegistration struct {
	// Name is the registered root, as written before the colon: `hover`, `group`, `@max`.
	Name string
	// Order is the registration number. Shared across every root registered inside one
	// `Variants.group`.
	Order int
	// Kind is the registration kind, which decides which branch of Compare applies.
	Kind ParsedVariantKind
}

// VariantComparison separates two variants that share a registration order.
//
// Upstream's `CompareFn`, stored in `Variants.compareFns` against the order number a `group`
// registered under. Returning zero means the two are genuinely equal in the sort and will share a
// bitmask index, which is observable: their classes then interleave by property rather than being
// grouped.
type VariantComparison func(left ParsedVariant, right ParsedVariant) int

// VariantRegistry is the engine's variant registry: every registered root and its order.
//
// The zero value is not usable; build one with NewVariantRegistry. A registry is not safe for
// concurrent mutation and is safe for concurrent reads once built, matching Theme and for the same
// reason: it is built once per run and shared read-only across the worker pool.
type VariantRegistry struct {
	registrations map[string]VariantRegistration
	// comparisons is keyed by order number rather than by root, because that is how upstream keys
	// `compareFns`: the function belongs to the group, not to any one member of it.
	comparisons map[int]VariantComparison
	// lastOrder mirrors upstream's `lastOrder`, so a root registered after the framework's appends
	// rather than colliding.
	lastOrder int
	// groupOrder mirrors upstream's `groupOrder`: non-nil while inside Group, which is what makes
	// every root registered in that window share one number.
	groupOrder *int
}

// NewVariantRegistry returns an empty registry.
func NewVariantRegistry() *VariantRegistry {
	return &VariantRegistry{
		registrations: map[string]VariantRegistration{},
		comparisons:   map[int]VariantComparison{},
	}
}

// Register adds or updates a variant root, reproducing upstream's `Variants.set`.
//
// The update branch is load-bearing and is the one a reader is most likely to get wrong. When the
// name already exists, upstream assigns kind, applyFn and compounds onto the existing record and
// leaves `order` untouched. So re-registering never moves a variant. That is what makes
// `@custom-variant dark (...)` change the selector `dark:` emits without changing where `dark:`
// sorts, which is exactly what both repositories in the corpus do.
func (registry *VariantRegistry) Register(name string, kind ParsedVariantKind) {
	if existing, isRegistered := registry.registrations[name]; isRegistered {
		existing.Kind = kind
		registry.registrations[name] = existing
		return
	}

	order := registry.nextOrder()
	if registry.groupOrder == nil {
		registry.lastOrder = order
	}
	registry.registrations[name] = VariantRegistration{Name: name, Order: order, Kind: kind}
}

// Group registers everything inside register under a single shared order number.
//
// Upstream's `Variants.group`. The comparison, when non-nil, is attached to that shared number and
// is what separates roots the order number cannot. Passing nil registers a group whose members tie,
// which is a real configuration rather than a degenerate one: tied variants share a bitmask index.
func (registry *VariantRegistry) Group(register func(), comparison VariantComparison) {
	order := registry.nextOrder()
	registry.groupOrder = &order
	if comparison != nil {
		registry.comparisons[order] = comparison
	}
	register()
	registry.groupOrder = nil
	// Upstream leaves `lastOrder` alone through a group and lets `nextOrder` compute
	// `lastOrder + 1` afterwards. Because `groupOrder` was `lastOrder + 1` itself, the next
	// registration outside the group would otherwise reuse the group's own number. Advancing here
	// is what keeps a post-group registration past the group rather than inside it.
	registry.lastOrder = order
}

// nextOrder ports upstream's `nextOrder`.
func (registry *VariantRegistry) nextOrder() int {
	if registry.groupOrder != nil {
		return *registry.groupOrder
	}
	return registry.lastOrder + 1
}

// RegisterFrameworkVariants replays a generated registry: names, kinds, and the order numbers the
// engine itself assigned.
//
// Separate from Register, and the separation is the point. Register computes the next order, which
// is right for a `@custom-variant` appending past the framework and wrong for replaying the
// framework: the engine shares order numbers across the roots registered inside one
// `Variants.group`, and 88 registrations hold only 82 distinct numbers in a default build. Feeding
// those 88 through Register in sequence would give each its own number, which reads as a harmless
// off-by-a-few and is not. A shared order is the only place a comparison function is consulted, so
// splitting the six roots at order 64 apart means `sm` through `2xl` never reach
// `CompareBreakpoints` at all and fall through to a root-name comparison, where `2xl` precedes
// `sm`. Every responsive class in the tree would sort wrongly, from a table that looked correct.
//
// `lastOrder` is advanced past the highest replayed order, so a repository's `@custom-variant`
// under a new name appends after the framework rather than colliding with it. That is upstream's
// own sequencing: the framework registers first, the stylesheet's own variants after.
//
// Re-registering a name already present keeps its order, matching Register and upstream's
// `Variants.set`. Calling this twice is therefore idempotent in position rather than shifting
// everything, though nothing in the shipped path calls it twice.
func (registry *VariantRegistry) RegisterFrameworkVariants(registrations []FrameworkVariantRegistration) {
	for _, registration := range registrations {
		if existing, isRegistered := registry.registrations[registration.Name]; isRegistered {
			existing.Kind = registration.Kind
			registry.registrations[registration.Name] = existing
			continue
		}
		registry.registrations[registration.Name] = VariantRegistration{
			Name:  registration.Name,
			Order: registration.Order,
			Kind:  registration.Kind,
		}
		if registration.Order > registry.lastOrder {
			registry.lastOrder = registration.Order
		}
	}
}

// AttachComparison registers a comparison function against one order number.
//
// Upstream reaches this state only through `Variants.group`, which cannot be used here: replaying a
// generated registry supplies the order numbers rather than computing them, so there is no group
// window to register inside. The two paths write the same map, and Compare cannot tell which put an
// entry there.
//
// An order with no registration behind it is accepted rather than rejected, because the caller that
// would notice is the generator's validation, which already refuses to render one. Rejecting here
// would mean a second, quieter copy of that rule in a place with no way to report it.
func (registry *VariantRegistry) AttachComparison(order int, comparison VariantComparison) {
	if comparison == nil {
		return
	}
	registry.comparisons[order] = comparison
}

// Has reports whether a root is registered.
func (registry *VariantRegistry) Has(name string) bool {
	_, isRegistered := registry.registrations[name]
	return isRegistered
}

// Get returns a root's registration.
func (registry *VariantRegistry) Get(name string) (VariantRegistration, bool) {
	registration, isRegistered := registry.registrations[name]
	return registration, isRegistered
}

// Registrations returns every registration, ordered by name.
//
// Sorted rather than in map order so that two runs over the same registry produce the same slice.
// A Go map iterates in a deliberately randomized order, and a caller diffing two registries would
// otherwise see a different diff on every run.
func (registry *VariantRegistry) Registrations() []VariantRegistration {
	registrations := make([]VariantRegistration, 0, len(registry.registrations))
	for _, registration := range registry.registrations {
		registrations = append(registrations, registration)
	}
	sort.Slice(registrations, func(left int, right int) bool {
		return registrations[left].Name < registrations[right].Name
	})
	return registrations
}

// Compare orders two parsed variants, porting upstream's `Variants.compare`.
//
// The branch order is upstream's and every step of it is observable:
//
//   - A nil variant sorts first. Upstream's `a === z` identity check collapses to a deep equality
//     here, because Go has no object identity for a value type and two separately parsed `hover`
//     variants must compare equal the way two references to one cached object do.
//   - Arbitrary variants sort after every registered one, and against each other by selector text.
//   - Registration order decides next, which is where a shared group order ties.
//   - Two compounds recurse into their inner variants before considering their modifiers.
//   - A group's comparison function, when one is registered for the shared order.
//   - Root name, then value kind (named before arbitrary), then value text.
func (registry *VariantRegistry) Compare(left *ParsedVariant, right *ParsedVariant) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return -1
	}
	if right == nil {
		return 1
	}
	if variantsAreEqual(*left, *right) {
		return 0
	}

	leftIsArbitrary := left.Kind == ParsedVariantKindArbitrary
	rightIsArbitrary := right.Kind == ParsedVariantKindArbitrary
	switch {
	case leftIsArbitrary && rightIsArbitrary:
		// Upstream returns -1 or 1 without an equality branch, relying on the list being deduped
		// before it is sorted. The equality check above already handled identical selectors, so
		// reaching here with equal text is impossible for the same reason.
		if left.Selector < right.Selector {
			return -1
		}
		return 1
	case leftIsArbitrary:
		return 1
	case rightIsArbitrary:
		return -1
	}

	leftRegistration, leftIsRegistered := registry.registrations[left.Root]
	rightRegistration, rightIsRegistered := registry.registrations[right.Root]
	// Upstream indexes the registry with a non-null assertion, because a candidate that named an
	// unregistered variant never parses and so never reaches the sort. cohere reaches this function
	// from a differential harness that may hold a class the parser rejected, so an unregistered root
	// is ranked past every registered one instead of panicking. Both unregistered roots then fall
	// through to the name comparison below, which keeps each one's classes grouped.
	leftOrder := unregisteredVariantOrder
	if leftIsRegistered {
		leftOrder = leftRegistration.Order
	}
	rightOrder := unregisteredVariantOrder
	if rightIsRegistered {
		rightOrder = rightRegistration.Order
	}
	if orderedByVariant := leftOrder - rightOrder; orderedByVariant != 0 {
		return orderedByVariant
	}

	if left.Kind == ParsedVariantKindCompound && right.Kind == ParsedVariantKindCompound {
		if order := registry.Compare(left.Variant, right.Variant); order != 0 {
			return order
		}
		switch {
		case left.Modifier != nil && right.Modifier != nil:
			if left.Modifier.Value < right.Modifier.Value {
				return -1
			}
			return 1
		case left.Modifier != nil:
			return 1
		case right.Modifier != nil:
			return -1
		default:
			return 0
		}
	}

	if comparison, hasComparison := registry.comparisons[leftOrder]; hasComparison {
		return comparison(*left, *right)
	}

	if left.Root != right.Root {
		if left.Root < right.Root {
			return -1
		}
		return 1
	}

	// Upstream's comment records that both variants are functional by this point, static ones
	// having been deduped and caught by the identity check. A static variant reaching here would
	// have a nil Value and take the null branches below, which is the same answer upstream's
	// undefined would produce rather than a crash.
	leftValue, rightValue := left.Value, right.Value
	if leftValue == nil {
		return -1
	}
	if rightValue == nil {
		return 1
	}

	leftValueIsArbitrary := leftValue.Kind == ParsedValueKindArbitrary
	rightValueIsArbitrary := rightValue.Kind == ParsedValueKindArbitrary
	if leftValueIsArbitrary && !rightValueIsArbitrary {
		return 1
	}
	if !leftValueIsArbitrary && rightValueIsArbitrary {
		return -1
	}

	if leftValue.Value < rightValue.Value {
		return -1
	}
	return 1
}

// unregisteredVariantOrder ranks a root the registry has never seen, past every registered one.
//
// Upstream cannot reach this state, so the value is cohere's own and is chosen to be far past any
// plausible registration count rather than to mean anything.
const unregisteredVariantOrder = 1 << 24

// variantsAreEqual reports whether two parsed variants are the same variant.
//
// This stands in for upstream's `a === z`. Upstream compares object identity and gets away with it
// because every variant reaching the sort came out of a `DefaultMap` keyed by the variant string,
// so one spelling is one object. Go has no such identity for a value type, and two `hover` variants
// parsed from two class strings are distinct values that must still compare equal, or the sort
// would rank a variant against itself and the dense-index assignment would give one spelling two
// indices.
func variantsAreEqual(left ParsedVariant, right ParsedVariant) bool {
	if left.Kind != right.Kind ||
		left.Root != right.Root ||
		left.Selector != right.Selector ||
		left.Relative != right.Relative {
		return false
	}
	if (left.Value == nil) != (right.Value == nil) {
		return false
	}
	if left.Value != nil && (left.Value.Kind != right.Value.Kind || left.Value.Value != right.Value.Value) {
		return false
	}
	if (left.Modifier == nil) != (right.Modifier == nil) {
		return false
	}
	if left.Modifier != nil && (left.Modifier.Kind != right.Modifier.Kind || left.Modifier.Value != right.Modifier.Value) {
		return false
	}
	if (left.Variant == nil) != (right.Variant == nil) {
		return false
	}
	if left.Variant != nil {
		return variantsAreEqual(*left.Variant, *right.Variant)
	}
	return true
}

// RunVariantOrder is the dense index assigned to each variant for one run.
//
// Built by BuildVariantOrder over exactly the variants a run parsed, never package-level, because
// the index is a rank within that population rather than a fact about the variant. See the file
// comment.
type RunVariantOrder struct {
	// entries holds each distinct variant and its index, in sorted order.
	entries []variantOrderEntry
}

type variantOrderEntry struct {
	variant ParsedVariant
	index   int
}

// BuildVariantOrder ports `DesignSystem.getVariantOrder`.
//
// Deduplicates, sorts by Compare, then walks assigning dense indices from zero, advancing only when
// Compare says the next variant differs from the previous one. The tie collapse is the part that
// matters: two variants that compare equal share an index, so their classes carry the same bit and
// interleave by property rather than being separated. Upstream gets deduplication for free from the
// `DefaultMap` its variants are cached in; here it is explicit, and it has to be, because the dense
// index would otherwise advance once per duplicate and leave gaps that shift every later variant.
func (registry *VariantRegistry) BuildVariantOrder(variants []ParsedVariant) *RunVariantOrder {
	// Every nested sub-variant joins the population, not only the variants the caller named.
	//
	// This is the behaviour a reading of `getVariantOrder` does not predict, and it changes every
	// mask. Upstream ranks `Array.from(parsedVariants.values())`, which is the design system's
	// whole parse cache, and parsing `group-hover` populates that cache with the compound AND with
	// the standalone `hover` it wraps. So a run that named two variants can rank four, and those
	// extra entries are not inert: they consume dense indices and shift every later variant's bit
	// upward.
	//
	// Measured on the engine, the class list `dark:flex group-hover:disabled:flex` ranks four
	// variants, of which the standalone `hover` was never written by either class. Without it
	// `disabled` takes index 1 instead of 2 and `dark` takes 2 instead of 3, so both masks come out
	// half their true value. The relative order survives that particular shift, which is exactly
	// what makes the omission dangerous: it agrees on small lists and diverges once a phantom lands
	// between two real variants.
	population := make([]ParsedVariant, 0, len(variants))
	for _, variant := range variants {
		population = appendVariantAndNested(population, variant)
	}

	distinct := make([]ParsedVariant, 0, len(population))
	for _, variant := range population {
		isDuplicate := false
		for _, existing := range distinct {
			if variantsAreEqual(existing, variant) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			distinct = append(distinct, variant)
		}
	}

	// SliceStable rather than Slice, though the choice is defensive rather than load-bearing and it
	// is worth saying which. Tied variants receive the SAME index, so the order within a tie never
	// reaches a mask and an unstable sort produces identical masks; mutating this to Slice does not
	// fail the suite. What stability buys is that Variants() reports a reproducible sequence, which
	// a caller diffing two runs depends on and the mask does not.
	sort.SliceStable(distinct, func(left int, right int) bool {
		return registry.Compare(&distinct[left], &distinct[right]) < 0
	})

	order := &RunVariantOrder{entries: make([]variantOrderEntry, 0, len(distinct))}
	index := 0
	for position := range distinct {
		if position > 0 {
			previous := distinct[position-1]
			if registry.Compare(&previous, &distinct[position]) != 0 {
				index++
			}
		}
		order.entries = append(order.entries, variantOrderEntry{variant: distinct[position], index: index})
	}
	return order
}

// appendVariantAndNested adds a variant and every variant nested inside it.
//
// A compound's inner variant is appended after the compound itself, matching the order the engine's
// parse cache receives them: `parseVariant('group-hover')` inserts the compound, and the recursive
// parse of its argument inserts the inner `hover`. The sort that follows makes the arrival order
// irrelevant to the result, but keeping it faithful means a future reader diffing the population
// against the engine's cache sees the same sequence rather than a reordering they have to explain.
func appendVariantAndNested(population []ParsedVariant, variant ParsedVariant) []ParsedVariant {
	population = append(population, variant)
	if variant.Variant != nil {
		population = appendVariantAndNested(population, *variant.Variant)
	}
	return population
}

// Len reports how many distinct variants the order holds.
func (order *RunVariantOrder) Len() int {
	return len(order.entries)
}

// IndexOf returns the dense index assigned to a variant, and whether it was assigned one.
//
// Looked up by value equality rather than by identity, for the reason variantsAreEqual exists.
func (order *RunVariantOrder) IndexOf(variant ParsedVariant) (int, bool) {
	for _, entry := range order.entries {
		if variantsAreEqual(entry.variant, variant) {
			return entry.index, true
		}
	}
	return 0, false
}

// Variants returns each variant and its index, in sorted order.
func (order *RunVariantOrder) Variants() []ParsedVariant {
	variants := make([]ParsedVariant, 0, len(order.entries))
	for _, entry := range order.entries {
		variants = append(variants, entry.variant)
	}
	return variants
}

// VariantBitmask builds the mask `compile.ts` sorts on, for one candidate's variants.
//
// Upstream is three lines:
//
//	let variantOrder = 0n
//	for (let variant of candidate.variants) variantOrder |= 1n << BigInt(variantOrderMap.get(variant)!)
//
// A `big.Int` rather than a `uint64` because the shift is by the variant's dense index and a class
// list using more than 64 distinct variants is ordinary in a large repository. The corpus this is
// tested against reaches 143 distinct variants in one repository's lists, so a 64-bit mask would
// have silently wrapped rather than overflowed loudly, producing a mask that compares wrong against
// exactly the variants a small test never uses.
//
// Reports false when any variant has no index, which happens only when the order was built from a
// different population than the candidate came from. Returning a partial mask instead would be a
// mask missing a bit, and a missing high bit reorders the class against every other.
func (order *RunVariantOrder) VariantBitmask(variants []ParsedVariant) (*big.Int, bool) {
	mask := new(big.Int)
	for _, variant := range variants {
		index, hasIndex := order.IndexOf(variant)
		if !hasIndex {
			return nil, false
		}
		mask.SetBit(mask, index, 1)
	}
	return mask, true
}

// CompareVariantBitmasks orders two masks the way `compile.ts` does.
//
// Upstream subtracts the two BigInts and returns the sign, which is a numeric comparison of the
// whole mask rather than anything positional. That is the mechanism this component exists to state
// plainly, because it is not the rule it looks like:
//
// The mask is a *set* of bits, and comparing two sets numerically is decided by the highest bit on
// which they differ. So a class's position is governed by its single highest-ranked variant, and
// only ties on that are broken by the next one down. It follows that a stacked variant does NOT
// sort after every single variant. `group-hover:disabled:` sets bits for `group-hover` and
// `disabled`, and `dark:` sets one much higher bit; the stacked class sorts FIRST. Measured on the
// engine, `group-hover:disabled:flex` precedes `dark:flex`, and in a four-class list it lands
// between `hover:flex` and `dark:flex` rather than after both.
//
// A depth-first rule, "every single variant precedes every stacked one", agrees with this on the
// common case where the stacked variant's outermost variant already outranks the other class's, and
// disagrees whenever it does not. See variant_test.go, which measures how often.
func CompareVariantBitmasks(left *big.Int, right *big.Int) int {
	return left.Cmp(right)
}

// VariantRootsOf lists the registered roots a variant reaches, outermost first.
//
// Not part of upstream, and used only by the differential harness in the tests: comparing two
// orderings needs a way to say which registered roots a class actually named, and reading them off
// the parsed variant is more honest than re-splitting the class string on colons. An arbitrary
// variant contributes its selector rather than a root, marked so it cannot be confused with one.
func VariantRootsOf(variant ParsedVariant) []string {
	if variant.Kind == ParsedVariantKindArbitrary {
		return []string{"[" + variant.Selector + "]"}
	}
	roots := []string{variant.Root}
	if variant.Variant != nil {
		roots = append(roots, VariantRootsOf(*variant.Variant)...)
	}
	return roots
}

// CompareBreakpoints ports `utils/compare-breakpoints.ts`.
//
// Registered as the comparison for the four grouped orders: `max` descending, the breakpoints and
// `min` ascending, `@max` descending and `@`/`@min` ascending. Without it the five breakpoints tie,
// because `Variants.group` gave them one order number, and their classes interleave.
//
// Two behaviours here are worth stating because a reader would not predict either. Bucketing strips
// every digit and dot from the value, so `40rem` and `1024px` bucket as `rem` and `px` and never
// compare numerically against each other at all: units sort alphabetically first. And when the
// numeric comparison yields NaN, which happens for `calc(...)` values whose parse produces nothing,
// upstream falls back to comparing the raw strings. Go's strconv returns an error rather than NaN,
// so the fallback is reached explicitly on a parse failure.
func CompareBreakpoints(left string, right string, ascending bool) int {
	if left == right {
		return 0
	}

	leftBucket := breakpointBucket(left)
	rightBucket := breakpointBucket(right)
	if leftBucket != rightBucket {
		if leftBucket < rightBucket {
			return -1
		}
		return 1
	}

	leftNumber, leftIsNumeric := leadingInteger(left)
	rightNumber, rightIsNumeric := leadingInteger(right)
	if !leftIsNumeric || !rightIsNumeric {
		// Upstream's `Number.isNaN(order)` branch. Reached for `calc(100%-1rem)` against
		// `calc(100%-2rem)`, where both bucket as `calc` and neither parses to a number.
		if left < right {
			return -1
		}
		return 1
	}

	if ascending {
		return leftNumber - rightNumber
	}
	return rightNumber - leftNumber
}

// breakpointBucket ports the bucketing half of compareBreakpoints.
//
// A value containing `(` buckets by the text before it, on upstream's stated assumption that a
// parenthesis means a CSS function. Everything else buckets by its unit, which upstream obtains by
// deleting every digit and dot with `/[\d.]+/g`.
func breakpointBucket(value string) string {
	if index := strings.IndexByte(value, '('); index != -1 {
		return value[:index]
	}
	var bucket strings.Builder
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character >= '0' && character <= '9') || character == '.' {
			continue
		}
		bucket.WriteByte(character)
	}
	return bucket.String()
}

// leadingInteger reads the integer JavaScript's `parseInt` would read from the front of a value.
//
// `parseInt` skips leading whitespace, accepts an optional sign, then consumes digits until a
// non-digit, and yields NaN when it consumed none. `strconv.Atoi` on the whole string would reject
// `40rem`, which is the ordinary case here rather than an edge one.
func leadingInteger(value string) (int, bool) {
	index := 0
	for index < len(value) && (value[index] == ' ' || value[index] == '\t' || value[index] == '\n' || value[index] == '\r') {
		index++
	}

	sign := 1
	if index < len(value) && (value[index] == '+' || value[index] == '-') {
		if value[index] == '-' {
			sign = -1
		}
		index++
	}

	start := index
	number := 0
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		number = number*10 + int(value[index]-'0')
		index++
	}
	if index == start {
		return 0, false
	}
	return sign * number, true
}
