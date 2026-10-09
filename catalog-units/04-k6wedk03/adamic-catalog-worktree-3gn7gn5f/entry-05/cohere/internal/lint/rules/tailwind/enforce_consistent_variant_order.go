package tailwind

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// messageVariantOrder names both spellings, because the reader has to see the difference.
//
// A finding that says only "the variants are out of order" leaves the reader to work out which
// permutation is wanted, and the permutation is the whole content of the finding.
func messageVariantOrder(written string, sorted string) rule.Message {
	return rule.Message{
		Id: "variantOrder",
		Description: "The variants on \"" + written + "\" are written in an order Tailwind does not " +
			"emit them in; it compiles to the same CSS as \"" + sorted + "\", which is the spelling " +
			"the engine's own sort produces. Nothing breaks either way, and that is why the two " +
			"spellings accumulate side by side until a reader cannot tell whether a difference " +
			"between two class lists is meaningful. Write it as \"" + sorted + "\".",
	}
}

// EnforceConsistentVariantOrderOptions lets a project name the surfaces that carry class strings.
type EnforceConsistentVariantOrderOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
}

// EnforceConsistentVariantOrder reports stacked variants written in an order Tailwind would not.
//
//	valid:   <div className="md:dark:flex" />
//	valid:   <div className="hover:focus:underline" />
//	invalid: <div className="dark:md:flex" />
//	invalid: <div className="hover:sm:flex" />
//
// # Why this is not a formatting rule
//
// Two variants on one class compile to nested rules, and the nesting order is decided by the
// engine's own sort rather than by the order they were written in. So `dark:md:flex` and
// `md:dark:flex` produce byte-identical CSS. Nothing catches the difference: not the compiler, not
// the browser, not a screenshot. What it costs is the same thing every unenforced spelling costs,
// which is that the next reader cannot tell whether two class lists differing only in variant order
// differ for a reason.
//
// # The comparison is deliberately weaker than the engine's full sort
//
// This is the part a reader is most likely to get wrong, and it was measured against
// eslint-plugin-better-tailwindcss 4.7.0 rather than inferred. Upstream's `compareVariantOrder`
// returns 0 when BOTH variants sort below its `GLOBAL` flag, so two element-scoped variants are
// left exactly as written: `hover:focus:` and `focus:hover:` are both accepted, and reporting
// either would be a divergence.
//
// A variant carries the global flag when every selector it generates is free of `&`, which is
// upstream's `hasGlobalSelector`. Those are the media and at-rule variants, and they must sort
// ahead of anything scoped to an element because that is the nesting the engine emits. Among two
// globals the sort is descending by registration order, which is what makes `sm:lg:` report as
// `lg:sm:` rather than the other way around.
//
// cohere cannot compute the flag the way upstream does. The variant registry carries registration
// order and kind but not the selector bodies, because `variant.go` deliberately did not port the
// thousand lines of `variants.ts` that build CSS. So the set is named here instead, and it was
// enumerated by measurement rather than by reading: every framework variant in
// `framework_variant_table.go` was stacked under `hover:` and run through upstream, and the ones
// that moved ahead are exactly these. See globalSelectorVariants.
//
// # Not fixable, for the same reason ordering is not
//
// The rewrite is mechanical and the rule knows the answer. It is left out because a class literal
// in this codebase wraps across lines with indentation that carries intent, and a fixer that
// rewrote one class in place would still be the first thing to reflow the literal around it. The
// sibling `enforce-consistent-class-order` declines a fixer on the same grounds.
//
// The name matches the upstream rule's, without a `tailwind-` prefix, so it answers to the
// `better-tailwindcss/enforce-consistent-variant-order` line a project already has.
var EnforceConsistentVariantOrder = rule.Rule{
	Name: "better-tailwindcss/enforce-consistent-variant-order",
	// The design system: its stylesheets, read through the recording file system (DesignSystemFS).
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Read before the design system is resolved: reading is pure, and a run that declines for want
		// of a Tailwind entry point must still reach it, or the registry's decoder agreement check
		// never looks at this rule (#zwd43jn).
		configured, isConfigured := rule.OptionsAs[EnforceConsistentVariantOrderOptions](options)

		designSystem := DesignSystemForProgramAt(ctx.Program, configured.Location())
		if designSystem.Err != nil {
			if ctx.Program == nil {
				return nil
			}
			// No Tailwind entry point is silence with nothing wrong in a project without Tailwind, and
			// a skip --coverage names, so a project that enables this rule without one can see it (#pa7k7zv).
			if skipWithoutTailwind(ctx, designSystem.Err) {
				return nil
			}
			return declineListeners(ctx, "enforce-consistent-variant-order", designSystem)
		}

		surfaces := DefaultClassLiteralSurfaces()
		if isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				reportVariantOrder(ctx, literal, designSystem)
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// reportVariantOrder checks each class in one literal against the engine's variant sort.
func reportVariantOrder(ctx rule.Context, literal ClassLiteral, designSystem DesignSystemResult) {
	system := designSystem.System
	if system == nil {
		return
	}

	for _, className := range SplitClasses(literal.Text) {
		sorted, needsReorder := sortedVariantSpelling(className, system)
		if !needsReorder {
			continue
		}
		ctx.ReportRange(literal.Range, messageVariantOrder(className, sorted))
	}
}

// sortedVariantSpelling returns the class rewritten with its variants in the engine's order.
//
// Reports false when the class has fewer than two variants, when the parser declined it, or when
// the written order already matches. Upstream bails on each for the same reason: a class it cannot
// fully resolve is one it has no opinion about.
//
// # The rank is per-run, not per-registration, and that is the whole subtlety
//
// A first version of this read `VariantRegistration.Order` straight out of the registry and was
// wrong on the case that matters most. `sm` and `lg` share registration order 64, because the
// framework registers every named breakpoint inside one `Variants.group`, so a comparison on
// registration numbers answers "equal" and leaves `sm:lg:` alone. Upstream reports it.
//
// The number upstream compares is `getVariantOrder()`, which sorts the variants this class list
// parsed and assigns dense indices from zero, breaking the group tie through the comparison
// function attached to order 64: the breakpoints order by their resolved `--breakpoint` widths.
// `BuildVariantOrder` is cohere's port of exactly that, so this asks it rather than the registry.
//
// The population is this one class's variants plus their nested sub-variants, which is narrower
// than `enforce-consistent-class-order`'s (the whole literal). That is deliberate and matches
// upstream: `getVariantOrder` is handed every class in the literal, but the comparison only ever
// runs between two variants of the same class, and a dense rank's relative order within a subset is
// the relative order it has in the whole. A wider population shifts indices without reordering them.
func sortedVariantSpelling(className string, system *tailwindengine.LoadedDesignSystem) (string, bool) {
	prefix, root, hasVariants := splitVariantPrefix(className)
	if !hasVariants {
		return "", false
	}

	written := splitVariantSegments(prefix)
	if len(written) < 2 {
		return "", false
	}

	parsed := tailwindengine.ParseCandidate(className, system)
	if len(parsed) == 0 {
		return "", false
	}
	candidate := parsed[0]

	// The parser reports variants innermost-first while they are written outermost-first, so the
	// two lists are reverses of each other. A mismatch in length means the class held a colon this
	// file split on that the parser read as part of a value, and sorting one list by the other's
	// indices would then rename variants rather than reorder them.
	if len(candidate.Variants) != len(written) {
		return "", false
	}

	variantOrder := system.Variants().BuildVariantOrder(candidate.Variants)

	orders := make([]int, len(written))
	for index := range written {
		variant := candidate.Variants[len(written)-1-index]
		rank, hasRank := variantOrder.IndexOf(variant)
		if !hasRank {
			return "", false
		}
		if isGlobalVariant(variant, system) {
			rank |= globalVariantFlag
		}
		orders[index] = rank
	}

	sorted := make([]string, len(written))
	copy(sorted, written)
	sortedOrders := make([]int, len(orders))
	copy(sortedOrders, orders)

	// Insertion sort, because the comparison is not a total order and `sort.Slice` is free to call a
	// non-transitive comparison in an order that produces a different permutation than upstream's
	// `toSorted`. Upstream sorts stably over a comparison that answers 0 for two non-global
	// variants, so equal elements keep their written order; an insertion sort reproduces that, and a
	// class carries a handful of variants at most.
	for outer := 1; outer < len(sorted); outer++ {
		variant, order := sorted[outer], sortedOrders[outer]
		inner := outer - 1
		for inner >= 0 && compareVariantOrder(sortedOrders[inner], order) > 0 {
			sorted[inner+1], sortedOrders[inner+1] = sorted[inner], sortedOrders[inner]
			inner--
		}
		sorted[inner+1], sortedOrders[inner+1] = variant, order
	}

	for index := range written {
		if written[index] != sorted[index] {
			return strings.Join(sorted, ":") + ":" + root, true
		}
	}
	return "", false
}

// compareVariantOrder is upstream's function of the same name.
//
// The two branches that return 0 are the rule's whole restraint: two variants below the global flag
// are left in the order they were written, whatever their registration numbers say. Above it the
// sort is DESCENDING, which is upstream's `orderB > orderA` returning +1.
func compareVariantOrder(left int, right int) int {
	if left == right {
		return 0
	}
	if left < globalVariantFlag && right < globalVariantFlag {
		return 0
	}
	if right > left {
		return +1
	}
	return -1
}

// globalVariantFlag is upstream's `VARIANT_ORDER_FLAGS.GLOBAL`, 1 << 16.
//
// The value is upstream's rather than an arbitrary sentinel, so a registration order and a flagged
// order can never collide: the framework's highest registration is under a hundred.
const globalVariantFlag = 65536

// globalSelectorVariants is upstream's `hasGlobalSelector` as a set, because cohere cannot compute it.
//
// Upstream asks each variant for the selectors it generates and calls it global when every one of
// them is free of `&`. That question needs the `applyFn` bodies, and `internal/lint/rules/tailwind/collapse/variant.go`
// deliberately did not port them: they are the thousand lines of `variants.ts` that build CSS, and
// cohere emits none.
//
// So the set is enumerated instead, and it was measured rather than reasoned about. Every framework
// variant in `framework_variant_table.go` was written as `hover:<variant>:flex` and run through
// eslint-plugin-better-tailwindcss 4.7.0; the ones it reordered ahead of `hover` are exactly these
// and nothing else. The shape of the answer is its own check: every member is a media or at-rule
// variant, and the near misses are instructive. `supports-[x]` is not here, because a support query
// wraps a rule that still carries `&`. Neither is `dark`, whose default registration is a selector
// rather than a media query, and neither is `max-md` or `@lg`, which are functional variants whose
// generated rule keeps its element scope.
//
// The named breakpoints are deliberately ABSENT from this map even though they are global, and that
// is the one place a static set would have been wrong rather than merely incomplete. `sm` and `md`
// are not special names: they are whatever `--breakpoint-*` keys the theme declares, and
// `registerThemeBreakpointVariants` registers each of them at the framework breakpoint group's
// order. Upstream's own test suite has a `--breakpoint-desktop: 80rem` fixture expecting
// `hover:desktop:` to report, and a map naming `sm`, `md`, `lg`, `xl`, `2xl` would have been silent
// on it. So breakpoints are recognised by their registration order instead, which answers for every
// name a repository chooses. See variantSortOrder.
//
// A repository that redefines one of the remaining names with `@custom-variant` changes its
// selectors and so could change its answer. That is a real divergence and it is bounded: it would
// have to redefine a framework media variant as a selector-based one, which inverts the variant's
// meaning rather than tuning it. The alternative, porting the selector machinery to ask the
// question honestly, is the whole CSS half of the engine.
var globalSelectorVariants = map[string]bool{
	"any-pointer-coarse": true,
	"any-pointer-fine":   true,
	"any-pointer-none":   true,
	"contrast-less":      true,
	"contrast-more":      true,
	"forced-colors":      true,
	"inverted-colors":    true,
	"landscape":          true,
	"motion-reduce":      true,
	"motion-safe":        true,
	"noscript":           true,
	"pointer-coarse":     true,
	"pointer-fine":       true,
	"pointer-none":       true,
	"portrait":           true,
	"print":              true,
	"starting":           true,
}

// isGlobalVariant is upstream's `hasGlobalSelector`, answered without the selector machinery.
//
// Upstream asks a variant for the selectors it generates and calls it global when every one is free
// of `&`. That needs the `applyFn` bodies, and `internal/lint/rules/tailwind/collapse/variant.go` deliberately did not
// port them: they are the thousand lines of `variants.ts` that build CSS, and cohere emits none.
//
// Two sources answer it here instead. Named breakpoints come from the theme, because they are the
// repository's own names rather than a fixed five: `--breakpoint-desktop: 80rem` makes `desktop` a
// breakpoint and it is global for the reason `md` is. Upstream's suite carries that exact fixture,
// which is what caught a first version of this file naming the framework breakpoints in a static
// set. Everything else comes from globalSelectorVariants.
//
// A compound or arbitrary variant is never global: `group-hover` wraps a selector, and an arbitrary
// `[@media(hover)]` has no registration to look up. Both are still ORDERED, through the dense rank
// the caller already has; they simply carry no flag, which is what puts them after every media
// variant and matches upstream on both.
func isGlobalVariant(variant tailwindengine.ParsedVariant, system *tailwindengine.LoadedDesignSystem) bool {
	if variant.Kind != tailwindengine.ParsedVariantKindStatic {
		return false
	}
	if globalSelectorVariants[variant.Root] {
		return true
	}
	theme := system.Theme()
	if theme == nil {
		return false
	}
	for _, name := range theme.KeysInNamespaces([]string{"--breakpoint"}) {
		if name == variant.Root {
			return true
		}
	}
	return false
}

// splitVariantSegments splits a variant prefix on the colons between variants.
//
// `strings.Split` is wrong here and its failure is quiet: `has-[:checked]:md` holds two variants and
// three colon-delimited pieces, so a naive split reports a length the parser disagrees with and the
// caller declines a class it should have reported. Found by differential run against upstream on
// exactly that class, which was the one miss in fifteen.
func splitVariantSegments(prefix string) []string {
	segments := []string{}
	depth := 0
	start := 0
	for index := 0; index < len(prefix); index++ {
		switch prefix[index] {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ':':
			if depth == 0 {
				segments = append(segments, prefix[start:index])
				start = index + 1
			}
		}
	}
	return append(segments, prefix[start:])
}

// splitVariantPrefix separates the variant segments from the utility they modify.
//
// Splits on the LAST colon that is not inside brackets, because an arbitrary variant and an
// arbitrary value both carry colons of their own: `[&:hover]:flex` has one variant, not two.
func splitVariantPrefix(className string) (string, string, bool) {
	depth := 0
	lastColon := -1
	for index := 0; index < len(className); index++ {
		switch className[index] {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ':':
			if depth == 0 {
				lastColon = index
			}
		}
	}
	if lastColon <= 0 || lastColon == len(className)-1 {
		return "", "", false
	}
	return className[:lastColon], className[lastColon+1:], true
}
