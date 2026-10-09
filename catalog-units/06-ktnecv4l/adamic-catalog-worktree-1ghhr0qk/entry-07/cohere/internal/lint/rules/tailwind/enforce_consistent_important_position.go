package tailwind

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageImportantPosition names both spellings, because the replacement is the finding's content.
func messageImportantPosition(written string, corrected string) rule.Message {
	return rule.Message{
		Id: "importantPosition",
		Description: "The important marker on \"" + written + "\" is written where Tailwind v3 put " +
			"it. Tailwind v4 moved it to the end of the utility, so this is spelled \"" + corrected +
			"\" now. Both still compile, which is exactly why a v3 habit survives a paste and then " +
			"sits next to the v4 spelling with nothing to say which one the file means. Write it " +
			"as \"" + corrected + "\".",
	}
}

// EnforceConsistentImportantPositionOptions lets a project name the surfaces that carry classes.
//
// Position selects which spelling is canonical: "recommended" is v4's trailing marker and is the
// default, "legacy" is v3's leading one. Named rather than inferred from the installed Tailwind,
// because a repository mid-migration wants to pick a target rather than have one read off its
// node_modules.
type EnforceConsistentImportantPositionOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
	Position string `json:"position"`
}

// EnforceConsistentImportantPosition reports an important marker on the wrong side of a utility.
//
//	valid:   <div className="text-red-500!" />
//	valid:   <div className="hover:flex!" />
//	invalid: <div className="!text-red-500" />
//	invalid: <div className="hover:!flex" />
//
// # Why this is worth a rule
//
// Tailwind v3 wrote the important marker as a prefix and v4 writes it as a suffix. v4 still accepts
// both, so a class pasted out of a v3 answer compiles and renders and reviews clean. The cost is
// that the two spellings then sit side by side in one codebase with nothing to say which one is
// current, and the next person to write an important class picks whichever they saw last.
//
// # This rule does not need the design system
//
// Unlike its siblings here it asks nothing of the theme: the marker's position is a property of the
// string. `!unknown-class` reports, measured against upstream, which is the tell that membership is
// not part of the question. So this rule declares no ProgramReads and does not pay for a
// design system it would not read.
//
// Variants are stripped before the marker is looked for, because `hover:!flex` carries its marker on
// the utility rather than on the class, and a check that read the whole string would miss it.
//
// The negative sign is handled as upstream handles it, which is a case a reader would get wrong:
// `-!mt-4` and `!-mt-4` both report, and both correct to `-mt-4!`. Upstream strips the leading `-`
// FIRST and looks for `!` after it, so a marker written on either side of the sign is found, and the
// rebuild always puts the sign outside. That is why the two spellings collapse onto one answer.
//
// A class carrying the marker on both ends is left alone, because the trailing one already
// satisfies the recommended position and upstream's condition is an early return on exactly that.
//
// # One deliberate divergence, on a class that is nothing but the marker
//
// Upstream reports a bare `!` with the replacement `!`, a finding whose fix changes nothing.
// Measured, not inferred: it strips the marker, finds an empty base, and rebuilds a string identical
// to the input. This declines instead. A finding a reader cannot act on is worse than no finding,
// because acting on it is a no-op and the rule then reports it again forever. `no-unknown-classes`
// already reports the same class with something a reader can do about it.
//
// Not fixable, for the reason its siblings are not: a class literal here wraps across lines with
// indentation that carries intent, and a fixer would be the first thing to reflow it.
var EnforceConsistentImportantPosition = rule.Rule{
	Name: "better-tailwindcss/enforce-consistent-important-position",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		surfaces := DefaultClassLiteralSurfaces()
		position := importantPositionRecommended
		if configured, isConfigured := rule.OptionsAs[EnforceConsistentImportantPositionOptions](options); isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
			if configured.Position == importantPositionLegacy {
				position = importantPositionLegacy
			}
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				for _, className := range SplitClasses(literal.Text) {
					corrected, isMisplaced := correctedImportantPosition(className, position)
					if !isMisplaced {
						continue
					}
					ctx.ReportRange(literal.Range, messageImportantPosition(className, corrected))
				}
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

const (
	// importantPositionRecommended is v4's trailing marker: `flex!`.
	importantPositionRecommended = "recommended"
	// importantPositionLegacy is v3's leading marker: `!flex`.
	importantPositionLegacy = "legacy"
)

// correctedImportantPosition returns the class with its marker on the configured side.
//
// Reports false when the class carries no marker, when the marker is already on the configured
// side, or when it carries one on both ends. The last is upstream's behaviour rather than an
// omission: its condition returns early when the marker is present at the position being enforced,
// whatever else is true, so `!flex!` is accepted under both settings.
func correctedImportantPosition(className string, position string) (string, bool) {
	prefix, utility, hasVariants := splitVariantPrefix(className)
	if !hasVariants {
		utility = className
		prefix = ""
	}
	if utility == "" {
		return "", false
	}

	// Upstream strips the negative sign before looking for the marker, so a marker written inside
	// the sign is still found. Both `-!mt-4` and `!-mt-4` reach here as negative with a leading
	// marker, which is why they collapse onto one answer.
	isNegative := strings.HasPrefix(utility, "-")
	base := strings.TrimPrefix(utility, "-")

	atStart := strings.HasPrefix(base, "!")
	base = strings.TrimPrefix(base, "!")
	atEnd := strings.HasSuffix(base, "!")
	base = strings.TrimSuffix(base, "!")

	if base == "" {
		return "", false
	}
	switch {
	case !atStart && !atEnd:
		return "", false
	case position == importantPositionLegacy && atStart:
		return "", false
	case position == importantPositionRecommended && atEnd:
		return "", false
	}

	rebuilt := base
	if isNegative {
		rebuilt = "-" + rebuilt
	}
	if position == importantPositionLegacy {
		rebuilt = "!" + rebuilt
	} else {
		rebuilt = rebuilt + "!"
	}

	if prefix != "" {
		rebuilt = prefix + ":" + rebuilt
	}
	if rebuilt == className {
		return "", false
	}
	return rebuilt, true
}
