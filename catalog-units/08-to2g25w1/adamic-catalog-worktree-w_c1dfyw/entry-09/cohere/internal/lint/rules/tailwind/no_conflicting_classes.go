package tailwind

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

func messageConflictingClasses(className string, conflicting []string, properties []string) rule.Message {
	return rule.Message{
		Id: "conflictingClasses",
		Description: "The class \"" + className + "\" and \"" + strings.Join(conflicting, "\", \"") +
			"\" set the same CSS " + propertyWord(properties) + " (" + strings.Join(properties, ", ") +
			"), so only one of them takes effect and which one depends on the order Tailwind emitted " +
			"them in rather than the order they are written here. Keep the one that was meant and " +
			"remove the rest.",
	}
}

func propertyWord(properties []string) string {
	if len(properties) == 1 {
		return "property"
	}
	return "properties"
}

// NoConflictingClassesOptions lets a project name the surfaces that carry class strings.
type NoConflictingClassesOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
}

// NoConflictingClasses reports two classes in one literal that set the same CSS property.
//
//	valid:   <div className="flex items-center" />
//	valid:   <div className="w-8 h-8" />
//	valid:   <div className="p-4 px-8" />
//	valid:   <div className="flex hover:block" />
//	invalid: <div className="flex block" />
//	invalid: <div className="px-4 px-8" />
//
// Only one of two conflicting classes takes effect, and which one is decided by the order Tailwind
// emitted its stylesheet in rather than the order the author wrote them. So the markup reads as
// though it says one thing and renders as though it says another, and editing the order of the
// class list does not change the outcome, which is what makes this genuinely confusing to debug.
//
// Three distinctions decide what conflicts, all measured against the real plugin rather than
// reasoned about, and each one inverts an answer if it is lost:
//
// Property names, not values. `w-8` and `h-8` declare the same value under `width` and `height` and
// do not conflict; they collapse into `size-8`, which is a different rule's finding. `px-4` and
// `px-8` declare different values under one property and do conflict.
//
// Shorthands are not normalised. `p-4` and `px-8` visually overlap and upstream reports no conflict,
// because `padding` and `padding-inline` are different property names. Normalising them here would
// invent a finding upstream does not report.
//
// Variants are part of the identity. `flex hover:block` is not a conflict, because the two apply in
// different states and both take effect in the state they belong to. `hover:flex hover:block` is a
// conflict. A port that compared bare class names would report the first and be wrong on correct
// code, which is how a rule gets disabled.
//
// Reported symmetrically, matching upstream: `flex block` produces two findings, one anchored at
// each class. A port emitting one finding per pair under-reports by half.
//
// The fix is a Suggestion rather than a Fix, deliberately. Removing either class changes what
// renders, and only the author knows which one was meant. The edit engine never applies
// suggestions, so this is the difference between a repair a human chooses and one that silently
// picks for them.
var NoConflictingClasses = rule.Rule{
	Name: "better-tailwindcss/no-conflicting-classes",
	// The design system: its stylesheets, read through the recording file system (DesignSystemFS).
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Read before the design system is resolved: reading is pure, and a run that declines for want
		// of a Tailwind entry point must still reach it, or the registry's decoder agreement check
		// never looks at this rule (#zwd43jn).
		configured, isConfigured := rule.OptionsAs[NoConflictingClassesOptions](options)

		// Resolved once per file and before the listeners are built, matching the other migrated
		// rules. A project with no Tailwind is silence with nothing wrong; a project whose CSS will
		// not parse is reported once per file rather than swallowed.
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
			return declineListeners(ctx, "no-conflicting-classes", designSystem)
		}

		surfaces := DefaultClassLiteralSurfaces()
		if isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				reportConflicts(ctx, literal, designSystem)
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// classFacts is one class resolved to everything that decides whether it can collide.
//
// Four things, and each was found by a false positive on the real tree rather than by reading
// upstream: the state it applies in, the element it lands on, what it sets, and whether it layers.
type classFacts struct {
	ClassName string
	// Variants decides which state the class applies in. `flex hover:block` is not a conflict.
	Variants string
	// SelectorShape decides which element it lands on. `divide-*` emits under
	// `:where(.CLASS > :not(:last-child))` and targets children, so it never collides with
	// `border-*` on the element itself, even though both declare `border-color`.
	SelectorShape string
	// Properties must match as a whole set, not merely overlap. `border` sets style plus width and
	// `border-dashed` sets style alone; they compose rather than collide.
	Properties []string
	// Composes marks a root whose utilities layer through custom properties rather than overwrite.
	// Every `shadow-*` and `ring-*` emits the same `box-shadow` var chain, so two of them never
	// conflict with each other.
	Composes bool
}

// reportConflicts finds classes in one literal that collide on the same element in the same state.
func reportConflicts(ctx rule.Context, literal ClassLiteral, designSystem DesignSystemResult) {
	classes := SplitClasses(literal.Text)
	if len(classes) < 2 {
		return
	}

	// A repeated class is `no-duplicate-classes`'s finding; pairing it with itself would report
	// every duplicate in the tree as a conflict.
	seen := make(map[string]bool, len(classes))
	distinct := make([]string, 0, len(classes))
	for _, className := range classes {
		if seen[className] {
			continue
		}
		seen[className] = true
		distinct = append(distinct, className)
	}

	for _, finding := range conflictFindingsIn(distinct, designSystem) {
		// Reported per class rather than per pair, which is what makes the output symmetric and
		// matches upstream: `flex block` produces two findings, one anchored at each class.
		ctx.ReportRange(literal.Range, messageConflictingClasses(
			finding.ClassName, finding.Conflicting, finding.Properties))
	}
}

// conflictFinding is one class that collides, and what it collides with.
//
// Split out of `reportConflicts` so the corpus measurement in live_placement_test.go can count what
// the rule would report without building a Context and walking a tree. A measurement that
// reimplemented the pairing would be measuring its own copy, which is how a regression check comes
// to agree with a rule that changed.
type conflictFinding struct {
	ClassName   string
	Conflicting []string
	Properties  []string
}

// conflictFindingsIn resolves a class list and returns every collision in it.
//
// The classes must already be deduplicated: a repeated class is `no-duplicate-classes`'s finding,
// and pairing one with itself would report every duplicate in the tree as a conflict.
func conflictFindingsIn(distinct []string, designSystem DesignSystemResult) []conflictFinding {
	resolved := make([]classFacts, 0, len(distinct))
	for _, className := range distinct {
		facts, canResolve := resolveClassFactsIn(className, designSystem)
		if !canResolve {
			continue
		}
		resolved = append(resolved, facts)
	}

	findings := make([]conflictFinding, 0, len(resolved))
	for index, subject := range resolved {
		if subject.Composes {
			continue
		}

		conflictingNames := make([]string, 0, len(resolved))
		conflictingProperties := map[string]bool{}

		for otherIndex, other := range resolved {
			if otherIndex == index || other.Composes {
				continue
			}
			if other.Variants != subject.Variants || other.SelectorShape != subject.SelectorShape {
				continue
			}

			// The whole property set must match, not merely overlap. Upstream bails when the two
			// declare different numbers of properties and again when any is missing from the other.
			shared := sharedProperties(subject.Properties, other.Properties)
			if len(shared) == 0 || len(shared) != len(subject.Properties) || len(shared) != len(other.Properties) {
				continue
			}

			conflictingNames = append(conflictingNames, other.ClassName)
			for _, property := range shared {
				conflictingProperties[property] = true
			}
		}

		if len(conflictingNames) == 0 {
			continue
		}

		properties := make([]string, 0, len(conflictingProperties))
		for property := range conflictingProperties {
			properties = append(properties, property)
		}
		sort.Strings(properties)

		findings = append(findings, conflictFinding{
			ClassName:   subject.ClassName,
			Conflicting: conflictingNames,
			Properties:  properties,
		})
	}
	return findings
}

// selectorShapeOf returns the selector a utility emits under, defaulting to a bare class.
//
// Read off the ported handle bodies rather than from a table. Six framework roots wrap their
// declarations in a nested rule upstream, `divide`, `divide-x`, `divide-y`, `space-x` and `space-y`
// under `:where(& > :not(:last-child))` and `placeholder` under `&::placeholder`, and
// `SelectorShapeForRoot` returns what the emitter produced.
//
// # Why this stopped being a table
//
// `RootSelectorShapes` held those eight entries and its own comment worked out why that was not
// enough: the invariance measurement, 327 roots on ahra against 302 on an independent system both
// yielding the same 8, was "necessary and insufficient", because a repository `@utility` block can
// declare a nested selector that no corpus writes. The argument for keeping it was that such a class
// resolves to false before any pairing, so the gap cannot produce a wrong finding.
//
// That is true and it is a reason the defect is survivable rather than a reason the table is right.
// The handle bodies are ported, the wrapper is in them, so the shape is computable and the question
// of what a table might be missing does not arise.
//
// A repository `@utility` block still answers `.CLASS`, and that is upstream's answer rather than a
// default: `@utility` takes a declaration list, not a selector, so a block has no way to declare a
// wrapper. A nested selector written inside one compiles to a rule node carrying no top-level
// declaration, which `repositoryClassFacts` declines before this is consulted.
func selectorShapeOf(rootOrName string) string {
	return tailwindengine.SelectorShapeForRoot(rootOrName)
}

// looksLikeLength reports whether an arbitrary value is a measurement rather than a color.
func looksLikeLength(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '#' {
		return false
	}
	// Leading digit or sign covers `3px`, `0`, `-2rem`, `50%`.
	first := value[0]
	return first == '-' || first == '.' || (first >= '0' && first <= '9')
}

// sharedProperties returns the properties two classes both declare.
func sharedProperties(left []string, right []string) []string {
	shared := make([]string, 0, len(left))
	for _, property := range left {
		for _, other := range right {
			if property == other {
				shared = append(shared, property)
				break
			}
		}
	}
	return shared
}
