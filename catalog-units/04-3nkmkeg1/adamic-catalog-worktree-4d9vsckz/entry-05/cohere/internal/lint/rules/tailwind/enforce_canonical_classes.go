package tailwind

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

func messageCanonicalCollapse(inputs []string, output string) rule.Message {
	return rule.Message{
		Id: "canonicalCollapse",
		Description: "The classes \"" + strings.Join(inputs, "\", \"") + "\" say the same thing as \"" +
			output + "\". Class names are the one part of this codebase with no compiler, so the only " +
			"way anyone finds them is by reading and searching text, and two files that mean the same " +
			"layout should say it the same way. Otherwise a search for one spelling misses the other " +
			"and the design system's vocabulary quietly forks.",
	}
}

// EnforceCanonicalClassesOptions lets a project name the surfaces that carry class strings, and
// exempt classes it does not want rewritten.
type EnforceCanonicalClassesOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
	// Ignore lists regular expressions for classes to leave alone, matching the option upstream
	// takes. The oxlint configuration supplies one, and a rule that declared no options at all made
	// oxlint refuse the whole plugin with "does not accept options", which failed as a silent
	// zero-finding run rather than a crash.
	Ignore []string `json:"ignore"`
	// Collapse is upstream's `collapse`, on by default: whether classes that merge into one are
	// reported. Off, upstream still rewrites single classes, and so this rule, which reports only
	// collapses (single-class rewrites are out of its scope, see below), reports nothing.
	Collapse *bool `json:"collapse"`
	// Logical is upstream's `logical`, on by default: whether the canonicalizer reads logical
	// properties as their physical longhands when it compares classes. Off, the families that only
	// merge through that reading are not reported. Which families those are is measured by the
	// generator, per family, as RequiresLogicalToPhysical.
	Logical *bool `json:"logical"`
}

// EnforceCanonicalClasses reports class sets that collapse into a shorter, equivalent set.
//
//	valid:   <div className="px-4 py-2" />
//	valid:   <div className="w-8 h-4" />
//	valid:   <div className="px-4 sm:py-4" />
//	invalid: <div className="px-4 py-4" />
//	invalid: <div className="w-8 h-8" />
//	invalid: <div className="mt-1 mb-1 ml-1 mr-1" />
//
// This is the rule the collapse table exists for, and the one that cost the most to make portable.
// Tailwind's own answer comes from `canonicalizeCandidates`, a signature-equivalence search over the
// entire utility registry: it compiles each candidate to CSS, builds a property-to-value signature,
// enumerates every other root, reprints under each, intersects, then enumerates subsets. Porting
// that search was rejected twice, at eight to ten thousand lines that change every Tailwind minor.
//
// Porting its answers is a different proposition. 327 functional roots produce 53,301 pairs, of
// which 44 collapse, and each family is a fact about roots rather than a computation per class:
// `px + py => p` holds for every value the two roots share. So this rule reads a generated table and
// needs no JavaScript engine at lint time.
//
// # What it can and cannot see
//
// Measured by diffing this rule against the real engine over 2,693 literals with planted
// violations: 13 of the engine's 14 findings, and zero false positives.
//
// The one miss is `text-sm + leading-relaxed => text-sm/relaxed`, which merges by folding one
// class's value into the other's slash modifier rather than producing a third root. That is a
// different mechanism from a root family and no root table can express it.
//
// The error direction is deliberate and is the important part. This rule under-reports rather than
// over-reports, so its failure is a finding that does not appear, which someone notices, rather than
// noise on correct code, which is how a rule gets turned off.
//
// Single-class rewrites are also out of scope: `z-[1]` is `z-1` and `start-4` is `inset-s-4`, and
// both are properties of one class rather than of a pair. The table deliberately excludes them,
// because a generator that counted them as families reported 1,326 families including
// `-bg-conic + -end => -inset-e`, where one class rewrote alone and the other was an unrelated
// neighbour.
//
// # Why the preconditions are the whole rule
//
// Two classes only merge when everything except their root already agrees: their variants, their
// importance, and their value. `px-4 py-4` becomes `p-4`, but `px-4 py-2` collapses to nothing and
// so does `px-4 sm:py-4`. Getting that wrong in the permissive direction reports correct code.
//
// No fix, deliberately. The collapse is mechanical but the rewrite is not: the replacement has to
// be spliced into a class list whose order is `enforce-consistent-class-order`'s concern, and a
// rewrite that also reordered would fight it. Reporting the shorter spelling is enough for an
// author to act on.
var EnforceCanonicalClasses = rule.Rule{
	Name: "better-tailwindcss/enforce-canonical-classes",
	// The design system: its stylesheets, read through the recording file system (DesignSystemFS).
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Read before the design system is resolved: reading is pure, and a run that declines for want
		// of a Tailwind entry point must still reach it, or the registry's decoder agreement check
		// never looks at this rule (#zwd43jn).
		configured, isConfigured := rule.OptionsAs[EnforceCanonicalClassesOptions](options)

		// Resolved once per file and before the listeners are built, matching the other two migrated
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
			return declineListeners(ctx, "enforce-canonical-classes", designSystem)
		}

		surfaces := DefaultClassLiteralSurfaces()
		var ignore []string
		collapse := canonicalCollapseOptions{logical: true}
		if isConfigured {
			if configured.Collapse != nil && !*configured.Collapse {
				// Nothing this rule reports survives `collapse: false`, so it declines the file.
				return nil
			}
			if configured.Logical != nil {
				collapse.logical = *configured.Logical
			}
			surfaces = configured.ClassLiteralSurfaces()
			ignore = configured.Ignore
		}

		reader := surfaces.ReaderFor(ctx.FileCache)
		ignored := compileIgnorePatterns(ignore)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				reportCollapses(ctx, literal, ignored, designSystem, collapse)
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// candidateParts is a class split into the pieces that decide whether it can merge.
type candidateParts struct {
	ClassName string
	// Prefix is everything before the root: the variants, in source order.
	Prefix string
	Root   string
	// Value is the text after the root, empty for a class like `border-l`.
	//
	// Decoded rather than verbatim, because it decides equality: `parseCandidate` resolves
	// `bg-(--x)` and `bg-[var(--x)]` to the same value, and they ARE the same class, so the merge
	// precondition must see them as equal. SourceValue is the half that gets printed.
	Value string
	// SourceValue is the same value as the author spelled it, for printing back.
	//
	// Two fields rather than one because the decoded form is unwritable. `grid-cols-[1fr_auto]`
	// decodes to `1fr auto`, and a rewrite suggesting `grid-cols-[1fr auto]` names a class with a
	// space in it, which cannot appear in a class attribute at all. Measured on the corpus: 16
	// classes decode to something different from what was written, all of them arbitrary values
	// carrying `_` or the `(--x)` custom-property shorthand.
	//
	// Empty means "same as Value", which is the common case and keeps every caller that builds these
	// by hand correct by default.
	SourceValue string
	Important   bool
}

// mergeKey is everything except the root. Two classes can only merge when this matches.
func (c candidateParts) mergeKey() string {
	importance := ""
	if c.Important {
		importance = "!"
	}
	return c.Prefix + "|" + importance + "|" + c.Value
}

// printedValue is the value to write into a rebuilt class name.
//
// SourceValue when the decoded form differs from what the author wrote, Value otherwise. See the
// field comment: the decoded form is what decides equality and is not always writable.
func (c candidateParts) printedValue() string {
	if c.SourceValue != "" {
		return c.SourceValue
	}
	return c.Value
}

// reportCollapses finds mergeable sets in one literal and reports the shorter spelling.
//
// Applied repeatedly, because collapses compose: `mt-1 mb-1 ml-1 mr-1` becomes `m-1` only through
// `mt+mb => my`, then `ml+mr => mx`, then `my+mx => m`. Verified against the engine, which produces
// the same single finding for the four-class case.
func reportCollapses(
	ctx rule.Context,
	literal ClassLiteral,
	ignored []*ignorePattern,
	designSystem DesignSystemResult,
	collapse canonicalCollapseOptions,
) {
	classes := SplitClasses(literal.Text)
	if len(classes) < 2 {
		return
	}

	remaining := make([]string, 0, len(classes))
	seen := make(map[string]bool, len(classes))
	for _, className := range classes {
		if seen[className] || isIgnored(className, ignored) {
			continue
		}
		seen[className] = true
		remaining = append(remaining, className)
	}

	// Bounded rather than `for {}`: a family whose output re-entered its own inputs would loop
	// forever, and a rule that hangs is worse than one that misses. Four passes covers the deepest
	// real chain, which is the three steps of `m-1` plus one to notice there is nothing left.
	const maximumPasses = 6
	for pass := 0; pass < maximumPasses; pass++ {
		inputs, output, didMerge := mergeOnce(remaining, designSystem, collapse)
		if !didMerge {
			return
		}

		ctx.ReportRange(literal.Range, messageCanonicalCollapse(inputs, output))

		next := make([]string, 0, len(remaining))
		merged := map[string]bool{inputs[0]: true, inputs[1]: true}
		for _, className := range remaining {
			if merged[className] {
				continue
			}
			next = append(next, className)
		}
		remaining = append(next, output)
	}
}

// canonicalCollapseOptions is what a collapse is allowed to use, from the rule's options.
type canonicalCollapseOptions struct {
	// logical is upstream's `logical`: whether a family that merges only through the canonicalizer's
	// logical-to-physical reading counts.
	logical bool
}

// mergeOnce finds the first mergeable pair and returns what it becomes.
func mergeOnce(classes []string, designSystem DesignSystemResult, collapse canonicalCollapseOptions) ([]string, string, bool) {
	parsed := make([]candidateParts, 0, len(classes))
	for _, className := range classes {
		parts, canParse := splitCandidateIn(className, designSystem.System)
		if !canParse {
			continue
		}
		parsed = append(parsed, parts)
	}

	for i := 0; i < len(parsed); i++ {
		for j := i + 1; j < len(parsed); j++ {
			left, right := parsed[i], parsed[j]

			// Everything except the root must already agree. This is the precondition the whole
			// rule rests on, and relaxing it reports `px-4 py-2` and `px-4 sm:py-4`, both correct.
			if left.mergeKey() != right.mergeKey() {
				continue
			}

			outputRoot, hasFamily := collapseOutputFor(left.Root, right.Root, collapse)
			if !hasFamily {
				continue
			}

			return []string{left.ClassName, right.ClassName}, rebuildClass(left, outputRoot), true
		}
	}

	return nil, "", false
}

// collapseOutputFor looks up two roots in the generated table, in either order.
func collapseOutputFor(left string, right string, collapse canonicalCollapseOptions) (string, bool) {
	for _, family := range tailwindengine.CollapseFamilies {
		if family.RequiresLogicalToPhysical && !collapse.logical {
			continue
		}
		if family.First == left && family.Second == right {
			return family.Output, true
		}
		if family.First == right && family.Second == left {
			return family.Output, true
		}
	}
	return "", false
}

// rebuildClass reassembles a class around a new root, keeping everything the merge required to
// match.
func rebuildClass(parts candidateParts, root string) string {
	rebuilt := parts.Prefix + root
	if value := parts.printedValue(); value != "" {
		rebuilt += "-" + value
	}
	if parts.Important {
		rebuilt += "!"
	}
	return rebuilt
}

// ignorePattern is one compiled entry from the rule's `ignore` option.
type ignorePattern struct {
	pattern *regexp.Regexp
}

// compileIgnorePatterns turns the configured expressions into matchers.
//
// An unparseable expression is skipped rather than fatal, matching how the class-literal reader
// handles its own patterns: a rule that refuses to run because one regular expression in a config
// file is malformed reports a clean tree, which is the failure this whole tool exists to remove.
func compileIgnorePatterns(patterns []string) []*ignorePattern {
	compiled := make([]*ignorePattern, 0, len(patterns))
	for _, pattern := range patterns {
		expression, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		compiled = append(compiled, &ignorePattern{pattern: expression})
	}
	return compiled
}

// isIgnored reports whether a class was exempted by configuration.
//
// Applied before the merge rather than after, exactly as upstream does, so an ignored class costs
// one regular expression test rather than participating in a collapse that then has to be discarded.
func isIgnored(className string, ignored []*ignorePattern) bool {
	for _, entry := range ignored {
		if entry.pattern.MatchString(className) {
			return true
		}
	}
	return false
}
