// Live conflict resolution: what a class declares, asked of this repository's own `@utility` blocks
// first and of the framework's generated tables second.
//
// This is `no-conflicting-classes`'s half of the seam #3r6cxrb crosses, and it is the one where the
// answer is a split rather than a swap. The seam is drawn where it is because it was measured, and
// the measurement is below rather than in a commit message so the next person does not have to
// re-run it to know whether to trust the boundary.
//
// # What moved
//
// The class is now dissected by `ParseCandidate` against the live design system rather than by a
// prefix walk over the framework's root table. That fixes the same defect
// `enforce-canonical-classes` had: the walk re-derived where a root ends, and which roots exist is a
// question about this repository's `@utility` blocks as much as about the framework.
//
// The repository's own utilities are then answered by compiling them. `UtilityEvaluator.Compile`
// returns the declaration list of an `@utility` block, and its properties are the real CSS property
// names the class emits. Measured on the corpus: 85 classes are answered this way, every one an
// ahra or www-connected-app theme utility such as `background--0` or `content--placeholder`, and
// every one of them a class the framework tables can only answer because the generator was run
// against a repository that declared it.
//
// # What moved next, and why a reading was never the substitute
//
// The framework's own utilities were on `RootDeclaredProperties` and three companions until
// #mz0m6k8, and the reason given here was that nothing in shipped Go could produce the answer:
// compiling `px-4` needs the framework's utility handlers, which is `utilities.ts`, 6,827 lines,
// and porting it had been rejected twice.
//
// That is no longer true. The handle bodies are ported across three files in the engine package,
// and `DeclaredPropertiesFor` runs them, so a class's declared properties are computed the way the
// engine computes them. The four tables, 1,183 entries and 62 KB, are deleted.
//
// The shortcut this comment warned against is still the wrong one, and naming it is still worth the
// space because it is the obvious move for anyone reading the emitters for the first time.
// `Table.Lookup` returns `Order`, indices into `PropertyOrder`, invertible to property names. It got
// 1,355 of the corpus's 1,530 resolvable classes right and 75 wrong. The port does not use it, and
// the 75 say why:
//
//   - **The two table families answer deliberately different questions.**
//     `OrderingPropertiesByRoot`'s own comment states it: the ordering tables keep `--tw-*` custom
//     properties because they are what separates `shadow-lg` from `ring-1` in the sort, and a
//     conflict answer strips them because two classes both setting `--tw-border-style` are not in
//     conflict about anything an author sees. Stripping them from a reading recovers `[box-shadow]`
//     for both, which is right, and leaves `outline-none` and `ring-neutral-800` with nothing at all.
//   - **A `--tw-sort` override replaces the reading entirely.** `divide-y` reads `divide-y-width`,
//     which is not a CSS property and is not what the class declares; the class emits
//     `border-top-width`, `border-top-style` and their bottom pair. `PropertySort`'s own comment
//     records the latch that produces this. 8 corpus classes, all `divide-*`.
//   - **A reading is a sorted set of positions, not a declaration list.** `backdrop-blur-sm`
//     declares `-webkit-backdrop-filter` and `backdrop-filter`; only one has a position.
//
// Every one of those is a fact the declaration list has and the reading has thrown away, which is
// why `DeclaredPropertiesFor` asks the emitters for the list and never touches a reading.
//
// # What the deletion fixed rather than preserved
//
// A generated table knows the repository it was generated from. Measured: 29 of ahra's roots and 27
// of www-connected-app's were absent from `RootDeclaredProperties`, and 8 of ahra's own `@utility`
// blocks were captured INTO `StaticDeclaredProperties` as though they were Tailwind's. Computing the
// framework half and compiling the repository half puts each answer with the source that owns it.
//
// One answer changed, and it is a correction. `RootDeclaredProperties` stored
// `{font-size, line-height}` for root `text`; upstream emits the line height only when a modifier is
// written, so `text-[10px]` declares one property and `text-[10px]/6` declares two. 11 corpus
// occurrences moved, and the rule got stricter rather than quieter: a class claiming `line-height`
// it does not set is a class that fails to conflict with `leading-6` when it should.
//
// # The honest state of the boundary
//
// Two of this rule's tables remain, counted rather than estimated. `ComposingRoots` decides whether
// two values of a root layer or overwrite, which a declaration list cannot see because it carries no
// values; it is task #mz0m6k8's sibling P4 and is deliberately untouched here so the two changes do
// not race in one file.
//
// `ColorNames` is the other, and it is genuinely per-repository and still generated, which makes it a
// real gap rather than a settled seam: a repository that defines its own palette has different color names,
// and this rule reads Tailwind's. It is narrowed rather than closed here, since the theme is asked
// first so a repository's own colors are recognised, and what remains is the framework's default
// palette as a fallback for a repository that did not override it.
//
// # RootSelectorShapes, and why it stays without being invariant
//
// The eighth table is `RootSelectorShapes`, read by `selectorShapeOf`. Measured two ways rather than
// assumed, because "reads like framework data" is what was said about `CollapseFamilies` and the
// variant registrations before a measurement that could not have detected a repository token.
//
// First, generated against `independent_theme.css` and diffed. That is two observations rather than
// one wearing two names: 327 roots and 53,301 pairs on ahra against 302 and 45,451 on the
// independent system. Both produce the same 8 entries, same roots, same shapes, none on either side
// alone.
//
// That measurement is necessary and NOT sufficient, and the second one says why. A repository
// `@utility` block can declare a nested selector, and a synthetic one that does contributes a ninth
// entry: `@utility gutter-*` wrapping `& > :not(:last-child)` enumerates as
// `gutter -> .CLASS > :not(:last-child)`. So the table is repository-invariant across the systems
// that exist rather than by construction, and a repository that wrote such a block would have a root
// whose shape this table does not carry.
//
// It stays generated anyway, because that class cannot reach the comparison the shape decides.
// `repositoryClassFacts` counts only top-level declarations, and a nested block compiles to a single
// `rule` node carrying no property at all, so the class resolves to false and is skipped before any
// pairing. Measured on exactly that fixture: `gutter-small` beside `me-4` and beside `mr-4`, both
// margin-inline-end on the element itself, produces zero findings.
//
// So a missing shape costs a finding this rule was never going to make rather than producing a wrong
// one, which is the safe direction and the same direction every other gap here fails in.
// TestSelectorShapeCannotProduceAWrongFinding holds that, so it cannot decay into a claim.
package tailwind

import (
	"strings"

	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// resolveClassFactsIn reads a class's four deciding facts against a live design system.
//
// Two sources, in the order upstream registers them: the repository's `@utility` blocks first,
// because `Utilities.set` overwrites and a repository redefining a framework root wins, then the
// framework's tables. That order is the same one `LoadedDesignSystem.HasUtility` and
// `class_order_key.go`'s `readingFor` use, so all three agree about who answers for a root.
//
// Returns false for a class neither source can read, which is a decline rather than a failure:
// `reportConflicts` skips it, so it participates in no pair. Measured on the corpus, 16 classes
// reach that state — `container`, the gradient stops, `[font:inherit]` and the two
// `data-[show=*]:fade-*` classes — each one a class whose properties no per-root table can state,
// and each unchanged by this migration.
func resolveClassFactsIn(className string, designSystem DesignSystemResult) (classFacts, bool) {
	if designSystem.System == nil {
		return classFacts{}, false
	}

	variants, base, _ := dissectClass(className)

	// The repository's own blocks, compiled. This is the half that could not be a table: an
	// `@utility` block is this repository's, and a generated table only knows the one it was
	// generated from.
	if facts, isRepositoryUtility := repositoryClassFacts(className, variants, designSystem); isRepositoryUtility {
		return facts, true
	}

	// The framework's half, computed rather than looked up.
	//
	// This was four generated tables until #mz0m6k8: a static list, a root list, thirteen per-class
	// overrides and fifteen colour arms. `DeclaredPropertiesFor` runs the ported handle bodies
	// instead, so the answer follows from the value the way the engine's does and the three override
	// tables have nothing left to patch. The measurement that justified deleting them ran while both
	// sides existed: 13 of 13 class overrides and 15 of 15 colour arms reproduced, and 839 of 839
	// statics, with 8 declines that were this repository's own `@utility` blocks captured into a
	// framework table.
	//
	// What this fixes rather than preserves is the blindness the tables had to a repository's own
	// roots. A generated table knows the repository it was generated from, and 29 of ahra's roots
	// and 27 of www-connected-app's were absent from it.
	parsed := tailwindengine.ParseCandidate(className, designSystem.System)
	if len(parsed) == 0 {
		return classFacts{}, false
	}
	// The first reading, matching every other consumer of the parser in this package. See
	// `gate.go:227` for the measurement that choice rests on.
	candidate := parsed[0]

	properties, canCompute := tailwindengine.DeclaredPropertiesFor(
		&candidate,
		valueResolutionIn(base, &candidate, designSystem.System),
	)
	if !canCompute {
		return classFacts{}, false
	}

	// The name the shape and the composition flag are keyed on. A static is keyed on its whole name
	// and a functional class on its root, which is the same split the deleted tables made: `flex` is
	// the static `display: flex` AND the root of `flex-4`, and one key cannot serve both.
	shapeKey := base
	if candidate.Kind == tailwindengine.ParsedCandidateKindFunctional {
		shapeKey = candidate.Root
	}

	return classFacts{
		ClassName:     className,
		Variants:      variants,
		SelectorShape: selectorShapeOf(shapeKey),
		Properties:    properties,
		Composes:      composesForRoot(shapeKey),
	}, true
}

// composesForRoot answers whether a root's utilities layer rather than overwrite.
//
// Computed from the ported handle bodies across all three emitter slices. `ComposesFor` emits the
// root twice with two different values and compares the declarations an author can see, which is
// upstream's own definition rather than a proxy for it: a family that layers writes its value into
// its own `--tw-*` variable and a constant into the property they share, so two of its utilities
// produce identical visible text and stack.
//
// Measured against the generated table this replaced, class for class: 81 of 82 rows answered and
// all 81 agreed, with 0 disagreements. 171 roots the table omits are computed as conflicting with 0
// false positives. The one row not answered was `shadow-`, which is not a utility at all: it is one
// of the eight trailing-dash repository tokens `4f23e9f` found sitting in tables headed with a
// Tailwind version, and `DeclaredPropertiesFor` declines it before this is ever consulted.
//
// Returns false for a root no emitter answers, which is the same decline the property computation
// makes for it, so the two stay consistent about which roots exist.
func composesForRoot(root string) bool {
	composes, _ := tailwindengine.ComposesFor(root)
	return composes
}

// valueResolutionIn asks this repository's theme what a class's value resolved through.
//
// The emitters need this and deliberately do not compute it: which keys are in `--color` or `--font`
// is a fact about the repository's `@theme`, and answering it inside the engine's framework tables
// would put one repository's tokens in a file describing Tailwind. So the question is asked here,
// where a live design system is in hand.
//
// Two namespaces are consulted because two decide an arm. The colour namespaces separate
// `border-red-500` from `border-4`, which is most of these roots. `--font` separates `font-mono` from
// `font-medium`, which is neither a colour nor an inferable type, and is what the three
// `ClassDeclaredProperties` font overrides recorded.
//
// Flattening either half is loud rather than silent, which was measured: forcing `IsColor` to false
// makes every colour class report its root's width properties, and the corpus goes from 0 findings to
// 206 across both repositories.
func valueResolutionIn(
	base string,
	candidate *tailwindengine.ParsedCandidate,
	system *tailwindengine.LoadedDesignSystem,
) tailwindengine.ValueResolution {
	if candidate.Kind != tailwindengine.ParsedCandidateKindFunctional {
		return tailwindengine.ValueResolution{}
	}

	resolution := tailwindengine.ValueResolution{
		IsColor: valueIsColorIn(base, candidate.Root, system),
	}

	// `--font` is consulted only for the root that branches on it. Asking for every root would make
	// a value that happens to name a font key change an unrelated root's answer, which is the shape
	// of defect the bare-value precedence in `descriptor.go` exists to prevent.
	if candidate.Root == "font" && candidate.Value != nil &&
		candidate.Value.Kind == tailwindengine.ParsedValueKindNamed && system != nil {
		if _, isFontKey := system.Theme().Get([]string{"--font-" + candidate.Value.Value}); isFontKey {
			resolution.Namespace = "--font"
		}
	}
	return resolution
}

// repositoryClassFacts answers for a class the repository's own `@utility` blocks declare.
//
// Compiled rather than looked up, which is the whole reason this half can be live at all: the
// evaluator has the block's body and can say what it emits. `Compile` returns the declaration list,
// and its properties are real CSS property names.
//
// Custom properties are stripped, matching what the conflict tables do and for the same reason their
// doc comment gives: two classes both setting `--tw-border-style` are not in conflict about anything
// an author sees. A block that declares nothing else is therefore not resolvable here and falls
// through to the framework half, which is correct — `fade-in` is exactly that, an `@utility` block
// setting only `--enter-opacity`, and it has no visible property to collide on.
func repositoryClassFacts(
	className string,
	variants string,
	designSystem DesignSystemResult,
) (classFacts, bool) {
	evaluator := designSystem.System.Utilities()
	if evaluator == nil {
		return classFacts{}, false
	}

	parsed := tailwindengine.ParseCandidate(className, designSystem.System)
	if len(parsed) == 0 {
		return classFacts{}, false
	}
	// The first reading, matching every other consumer of the parser in this package. See
	// `gate.go:227` for the measurement that choice rests on.
	candidate := parsed[0]

	if !evaluator.Has(candidate.Root) {
		return classFacts{}, false
	}

	nodes, didCompile := evaluator.Compile(&candidate)
	if !didCompile {
		return classFacts{}, false
	}

	properties := make([]string, 0, len(nodes))
	seen := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if node.Property == "" || strings.HasPrefix(node.Property, "--") {
			continue
		}
		if seen[node.Property] {
			continue
		}
		seen[node.Property] = true
		properties = append(properties, node.Property)
	}
	if len(properties) == 0 {
		return classFacts{}, false
	}

	return classFacts{
		ClassName: className,
		Variants:  variants,
		// A repository `@utility` block emits under a bare class selector. The non-default shapes
		// the framework has — `divide-*` under `:where(.CLASS > :not(:last-child))` — come from its
		// own handlers, and a repository block has no way to declare one: `@utility` takes a
		// declaration list, not a selector.
		SelectorShape: ".CLASS",
		Properties:    properties,
		Composes:      composesForRoot(candidate.Root),
	}, true
}

// functionalRootIn returns the root a class reads as, when the design system reads it functionally.
//
// This replaces `functionalRootOf`'s longest-prefix walk over the deleted root table. The walk's
// own comment records why longest-wins was needed — `border-l-4` has root `border-l` rather than
// `border`, and the shorter match gives it `border-width` instead of `border-left-width`, which
// changes what it conflicts with — and the parser answers the same question without re-deriving it.
//
// Returns false when the class reads as static or arbitrary, which the caller handles separately:
// a static's properties are keyed on its whole name rather than on a root.
func functionalRootIn(className string, system *tailwindengine.LoadedDesignSystem) (string, bool) {
	parsed := tailwindengine.ParseCandidate(className, system)
	if len(parsed) == 0 {
		return "", false
	}
	candidate := parsed[0]
	if candidate.Kind != tailwindengine.ParsedCandidateKindFunctional {
		return "", false
	}
	return candidate.Root, true
}

// valueIsColorIn reports whether a class's value names a color, asking this repository's theme first.
//
// The shipped version read `ColorNames`, generated from one repository, and its own comment already
// said why that is the wrong shape: "a project that customises its colors has different names and
// hardcoding Tailwind's defaults would be wrong for it." It then hardcoded a generated set, which is
// the same defect one step removed.
//
// The table is gone. A repository that defines no `--color-*` of its own still inherits Tailwind's,
// and it inherits them into its theme, which is what this asks. Measured across both corpus systems
// before deleting: 558 of the table's 561 names are answered by the live theme, the 3 that are not
// are `current`, `inherit` and `transparent`, which are CSS keywords in no namespace and are what
// `colorKeywords` already holds, and 0 of 561 were left uncovered by the two together.
//
// The other direction is why it had to go rather than merely could. www-connected-app's theme
// declares `--color-brand`, `--color-brand-hover` and `--color-brand-active`, and the table knows
// none of them, because it was generated from ahra. A generated palette is wrong for every
// repository except the one it came from, which is the defect `KnownStatics` and
// `StaticDeclaredProperties` were deleted for.
func valueIsColorIn(base string, root string, system *tailwindengine.LoadedDesignSystem) bool {
	if len(base) <= len(root) {
		return false
	}
	value := strings.TrimPrefix(base[len(root):], "-")

	// An opacity modifier is part of the color, not of the name: `white/30` is `white`.
	if slash := strings.Index(value, "/"); slash >= 0 {
		value = value[:slash]
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inner := strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
		// A length carries a unit or is bare digits; anything else in brackets reads as a color.
		return !looksLikeLength(inner)
	}

	// This repository's own palette, which is the half a table cannot carry.
	if system != nil {
		if _, isThemeColor := system.Theme().Get([]string{"--color-" + value}); isThemeColor {
			return true
		}
	}

	// Tailwind's own colour keywords, which sit in no theme namespace and so cannot be asked for.
	return tailwindengine.IsColorKeyword(value)
}
