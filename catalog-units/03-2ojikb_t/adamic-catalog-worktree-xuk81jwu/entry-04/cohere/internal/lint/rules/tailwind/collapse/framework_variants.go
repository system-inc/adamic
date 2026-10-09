// The framework's variant registrations, attached to a design system over its live theme.
//
// `framework_variant_table.go` holds the static half — 88 names, the order numbers the engine
// assigned them, and which of those orders carry a comparison function. This file is the half that
// cannot be generated: the comparison itself, which upstream closes over the design system's theme.
//
// # Why the comparison cannot be a table
//
// Four order numbers carry a comparison in a default build, and all four are breakpoint
// comparisons. What they compare is not the variant's name but the width it resolves to, read out
// of the theme's `--breakpoint` and `--container` namespaces. Measured against the engine:
// redefining `--breakpoint-sm` to `200rem` moves `sm` from dense index 0 to index 4, and a
// `@theme { --breakpoint-tablet: 50rem }` registers a new static variant that sorts between `sm`
// and `lg`. A comparison closed over a hardcoded default scale agrees with the engine on a default
// theme and is wrong on exactly the repository that customized one, which is the per-repository
// defect the whole port exists to remove.
//
// The test suite's differential resolves these from a hardcoded scale, deliberately, so that it
// does not take a dependency on the theme port while testing the variant port. That is right for a
// test and would be wrong here.
//
// # Why the direction is generated and the function is not
//
// `max` and `@max` descend; the breakpoint group and `@`/`@min` ascend. That is four booleans, and
// they are measured by the generator probing the engine rather than read off `variants.ts`, so a
// release that flipped one produces a diff in a checked-in file. The function is the same function
// in all four cases, `CompareBreakpoints`, parameterized by that boolean.
package tailwind

import "strings"

// breakpointNamespace and containerNamespace are the theme namespaces the two families of
// breakpoint variant resolve through.
//
// `sm`, `min-sm` and `max-sm` read `--breakpoint-sm`; `@sm`, `@min-sm` and `@max-sm` read
// `--container-sm`. Upstream splits them the same way, and the two scales genuinely differ: `sm` is
// 40rem as a breakpoint and 24rem as a container, so resolving a container query through the
// breakpoint namespace would order every `@` variant against the wrong numbers.
const (
	breakpointNamespace = "--breakpoint"
	containerNamespace  = "--container"
)

// registerThemeBreakpointVariants registers a static variant for every `--breakpoint-*` key the
// theme defines that the framework has not already registered.
//
// This is a repository contribution the framework table cannot hold, and it was found by a test
// rather than assumed: `@theme { --breakpoint-tablet: 50rem }` makes the engine register `tablet`
// as a static variant, at the breakpoints' own shared order, so `tablet:flex` parses and sorts
// between `sm:` and `lg:` by its width. Measured on the engine, which reports 89 registrations for
// that stylesheet against 88 for a default build.
//
// Three properties of the engine's behaviour are reproduced here and each was measured:
//
//   - The order is the breakpoint group's shared one, not an appended position. A new breakpoint
//     joins the group and is separated from its members by the comparison function, which is what
//     makes it sort by width rather than after every framework variant.
//   - Only `--breakpoint-*` does this. `--container-*` contributes no registration: a stylesheet
//     adding `--container-huge` still reports 88, because `@huge:` reaches the theme through the
//     already-registered `@` root rather than through a root of its own.
//   - A name the framework already registers is left alone, which falls out of Register's update
//     branch. Redefining `--breakpoint-sm` changes what `sm:` resolves to and never moves it.
//
// Neither corpus repository defines one today, which is exactly why this could not be inferred from
// the invariance measurement and had to be measured against the engine directly.
func registerThemeBreakpointVariants(registry *VariantRegistry, theme *Theme) {
	if theme == nil {
		return
	}

	order, hasGroup := breakpointGroupOrder()
	if !hasGroup {
		return
	}

	for _, name := range theme.KeysInNamespaces([]string{breakpointNamespace}) {
		if name == "" || registry.Has(name) {
			continue
		}
		registry.RegisterFrameworkVariants([]FrameworkVariantRegistration{
			{Name: name, Order: order, Kind: ParsedVariantKindStatic},
		})
	}
}

// breakpointGroupOrder is the order number the framework's named breakpoints hold.
//
// Read out of the generated table rather than written as a constant, so a Tailwind release that
// renumbers the group is picked up by regenerating one file instead of by remembering that a second
// place holds the same number. `sm` is the probe because it is the group's first member and the one
// the generator's own direction probe uses; a build where `sm` is absent is a build this package
// has no map for, and registering a theme breakpoint at a guessed order would put it in a group
// with nothing to compare against.
func breakpointGroupOrder() (int, bool) {
	for _, registration := range FrameworkVariantRegistrations {
		if registration.Name == "sm" {
			return registration.Order, true
		}
	}
	return 0, false
}

// attachFrameworkVariantComparisons registers the generated comparison groups against a registry,
// resolving widths through the given theme.
//
// The theme is captured by the closures rather than passed per call, because upstream's is too: the
// comparison belongs to the design system that built it. A registry outliving its theme would be a
// use-after-free of a sort, and it cannot happen here because both are built inside
// LoadDesignSystem and shared read-only thereafter.
func attachFrameworkVariantComparisons(registry *VariantRegistry, theme *Theme) {
	for _, group := range FrameworkVariantComparisonGroups {
		ascending := group.Ascending
		registry.AttachComparison(group.Order, func(left ParsedVariant, right ParsedVariant) int {
			return compareBreakpointVariants(theme, left, right, ascending)
		})
	}
}

// compareBreakpointVariants ports upstream's `compareBreakpointVariants` closure.
//
// The unresolvable branches are upstream's own and are measured rather than assumed. A variant
// whose width the theme cannot answer sorts *first* when ascending and *last* when descending,
// which reads backwards until you see what it produces: in both directions the unresolvable one
// lands at the narrow end of the run. Confirmed on the engine, where `min-[var(--x)]` takes index 0
// against `sm` and `lg`, and `max-[var(--x)]` takes the last index against `max-sm` and `max-lg`.
//
// Both unresolvable is a genuine tie rather than an arbitrary pick. Upstream returns 0 and the two
// share a dense index, which is observable: their classes interleave by property instead of being
// grouped. Returning a stable non-zero here would look tidier and would split a bit the engine
// shares, shifting every higher variant's bit.
func compareBreakpointVariants(theme *Theme, left ParsedVariant, right ParsedVariant, ascending bool) int {
	leftValue, leftResolves := resolveBreakpointWidth(theme, left)
	rightValue, rightResolves := resolveBreakpointWidth(theme, right)

	switch {
	case !leftResolves && !rightResolves:
		return 0
	case !leftResolves:
		if ascending {
			return -1
		}
		return 1
	case !rightResolves:
		if ascending {
			return 1
		}
		return -1
	}

	return CompareBreakpoints(leftValue, rightValue, ascending)
}

// resolveBreakpointWidth answers what width a breakpoint or container variant resolves to.
//
// Three shapes reach here and each resolves differently:
//
//   - A static variant is a bare breakpoint name: `sm:`, and `tablet:` when the theme defines
//     `--breakpoint-tablet`. The root is the theme key.
//   - A functional variant with a named value is `min-sm:` or `@lg:`. The value is the theme key,
//     read from whichever namespace the root selects.
//   - A functional variant with an arbitrary value is `min-[40rem]:`, which carries its own width
//     and consults no theme at all.
//
// An arbitrary value containing `var(` reports unresolved rather than comparing the literal text.
// Upstream cannot know what a custom property holds, and comparing `var(--x)` as a string would
// bucket it under `var(--x)` and order it against nothing meaningfully. Measured: the engine puts
// `min-[var(--x)]` at the narrow end rather than sorting it by text.
func resolveBreakpointWidth(theme *Theme, variant ParsedVariant) (string, bool) {
	if theme == nil {
		return "", false
	}

	switch variant.Kind {
	case ParsedVariantKindStatic:
		return theme.ResolveValue(variant.Root, true, []string{breakpointNamespace})

	case ParsedVariantKindFunctional:
		if variant.Value == nil {
			return "", false
		}
		if variant.Value.Kind == ParsedValueKindArbitrary {
			if strings.Contains(variant.Value.Value, "var(") {
				return "", false
			}
			return variant.Value.Value, true
		}
		return theme.ResolveValue(variant.Value.Value, true, []string{namespaceForVariantRoot(variant.Root)})

	default:
		// A compound or arbitrary variant never holds one of these order numbers, so this is
		// unreachable through Compare. Answering false rather than panicking keeps a caller that
		// reached it by another path ordered rather than crashed.
		return "", false
	}
}

// namespaceForVariantRoot picks the theme namespace a functional breakpoint variant reads.
//
// The `@` prefix is the whole signal: `@`, `@min` and `@max` are container queries and read
// `--container`; `min` and `max` are media queries and read `--breakpoint`. Upstream registers them
// under separate groups for the same reason, and the two scales assign different widths to the same
// names.
func namespaceForVariantRoot(root string) string {
	if strings.HasPrefix(root, "@") {
		return containerNamespace
	}
	return breakpointNamespace
}
