package tailwind

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageDuplicateClass names the class rather than the literal.
//
// A finding that says only "this literal has a duplicate" makes the reader scan a forty-class
// string to find which one, which is exactly the moment a rule stops being worth having on.
func messageDuplicateClass(className string) rule.Message {
	return rule.Message{
		Id: "duplicateClass",
		Description: "The class \"" + className + "\" is listed more than once. The repeat has no " +
			"effect, so it is either a leftover from an edit or a sign that two people added the " +
			"same class from different ends of the string. Remove the later one.",
	}
}

// NoDuplicateClassesOptions lets a project name the surfaces that carry class strings.
//
// The zero value is not "read nothing": an unconfigured rule falls back to the defaults, because a
// rule that declines every file when unconfigured reports a clean tree and is indistinguishable
// from one with nothing to say.
type NoDuplicateClassesOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
}

// NoDuplicateClasses reports a class name written twice in the same string.
//
//	valid:   <div className="flex items-center" />
//	valid:   <div className="px-4 py-2" />
//	invalid: <div className="flex items-center flex" />
//	invalid: mergeClassNames('px-4 py-2 px-4')
//
// The defect is small and the reason to catch it is not. A repeated class does nothing at runtime,
// so nothing ever surfaces it: it survives every review, renders correctly, and accumulates. Its
// cost is that it makes the string a worse record of intent. The second `flex` is evidence that two
// edits met in the middle without either author seeing the other, and the next person to change
// that literal inherits the ambiguity about which one is load-bearing.
//
// Reads all three surfaces the configuration names, not just JSX attributes. On the ahra tree that
// is 7,773 attribute literals against 2,192 in callees and variables, so an attribute-only rule
// would cover 78% of the class surface while printing green over the rest.
//
// Fixable, and safe to fix unattended: removing a later duplicate cannot change what the string
// means, because CSS class order carries no meaning of its own. Ordering is
// `enforce-consistent-class-order`'s concern and is left alone here.
//
// The name deliberately matches the upstream rule's, without a `tailwind-` prefix. Config lookup
// resolves a bare name against a namespaced entry on a `/` boundary, so `no-duplicate-classes`
// answers to the `better-tailwindcss/no-duplicate-classes` line the project already has. A prefixed
// name would not match it, and an unmatched rule is not an error: it runs over zero files and
// reports zero findings, which reads exactly like a clean tree. That was the first thing this rule
// did, and only the coverage line's "listened to no files" note caught it.
var NoDuplicateClasses = rule.Rule{
	Name: "better-tailwindcss/no-duplicate-classes",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		surfaces := DefaultClassLiteralSurfaces()
		if configured, isConfigured := rule.OptionsAs[NoDuplicateClassesOptions](options); isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				reportDuplicates(ctx, literal)
			}
			for _, segments := range reader.ClassTemplateSegmentsIn(node) {
				reportTemplateDuplicates(ctx, segments)
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// reportDuplicates finds repeats in one literal and proposes the rewritten string.
//
// The fix is computed once for the whole literal rather than as one deletion per duplicate. Two
// deletions inside the same string would each be computed against the original offsets, and
// applying the first invalidates the second. That is the class of bug the edit engine exists to
// prevent, and not one a rule should be handing it in the first place.
func reportDuplicates(ctx rule.Context, literal ClassLiteral) {
	classes := SplitClasses(literal.Text)
	if len(classes) < 2 {
		return
	}

	seen := make(map[string]bool, len(classes))
	var duplicated []string
	for _, className := range classes {
		if seen[className] {
			// Report each repeated class once, however many times it repeats: the reader's job is
			// to delete the extras, and saying `flex` three times does not help them do it.
			if !containsString(duplicated, className) {
				duplicated = append(duplicated, className)
			}
			continue
		}
		seen[className] = true
	}

	if len(duplicated) == 0 {
		return
	}

	// A literal holding a template hole is only partly known, so its classes can be reported but its
	// text must not be rewritten: the fix is built from the fragments the reader could see, and
	// would drop the interpolation.
	canFix := !strings.Contains(literal.Text, "${")
	var tokens []classToken
	tokenized := false
	if canFix && ctx.SourceFile != nil {
		tokens, tokenized = classTokensIn(ctx.SourceFile.Text(), literal.Range, literal.Text)
	}

	for _, className := range duplicated {
		message := messageDuplicateClass(className)
		if !canFix {
			ctx.ReportRange(literal.Range, message)
			continue
		}

		// Each class's diagnostic carries the deletions for its own repeats and nothing else. The
		// whole-literal rewrite this replaced was attached to every duplicated class's diagnostic,
		// so a literal with two duplicated classes proposed the same edit twice and the engine
		// refused the second as an overlap.
		// A literal whose source is not its decoded value is reported without a fix, for the reason
		// no-unnecessary-whitespace gives: decoded text written over an escape can stop the file
		// parsing, and then nothing in it is fixed.
		var fixes []rule.Fix
		if tokenized {
			fixes = repeatDeletions(tokens, className)
		}

		ctx.Report(rule.Diagnostic{
			Range:      literal.Range,
			Message:    message,
			SourceFile: ctx.SourceFile,
			Fixes:      fixes,
		})
	}
}

// reportTemplateDuplicates finds repeats in the runs of a template with holes.
//
// A repeat inside one run is fixed exactly as in a literal, which is what Prettier's Tailwind plugin
// does to the same run before sorting it: `flex flex ${size}` becomes `flex ${size}`. A repeat split
// across a hole, `flex ${size} flex`, is reported and not fixed. Both classes apply whatever the hole
// holds, so it is a real repeat, but the plugin dedupes each run on its own and leaves this one as
// written, and a fix here would be the one rewrite of a class string the plugin never makes.
//
// A class glued to a hole is not a class: `px-` in `px-${size}` is half of one the hole completes,
// so it is neither counted nor compared, matching the runs `enforce-consistent-class-order` leaves
// in place. Until this read templates, a repeat in a template run was invisible to every rule here,
// and the order rule, which leaves a run holding a repeat unordered, left it unordered in silence.
func reportTemplateDuplicates(ctx rule.Context, segments []ClassSegment) {
	sourceText := ""
	if ctx.SourceFile != nil {
		sourceText = ctx.SourceFile.Text()
	}

	type repeat struct {
		segment  ClassSegment
		fixes    []rule.Fix
		fixable  bool
		reported bool
	}
	repeats := map[string]*repeat{}
	order := []string{}
	seenInTemplate := map[string]bool{}

	for index, segment := range segments {
		if strings.Contains(segment.Text, "${") {
			continue
		}
		tokens, tokenized := classTokensIn(sourceText, segment.Range, segment.Text)
		if !tokenized {
			tokens = classTokensOf(segment.Text, 0)
		}
		if len(tokens) == 0 {
			continue
		}

		// The glued classes at either end, by the same test the plugin's ignoreFirst and ignoreLast
		// make.
		first, last := 0, len(tokens)
		if index > 0 && !tokens[0].Separator {
			first = 1
		}
		if index < len(segments)-1 && !tokens[len(tokens)-1].Separator {
			last = len(tokens) - 1
		}

		seenInRun := map[string]bool{}
		for tokenIndex := first; tokenIndex < last; tokenIndex++ {
			token := tokens[tokenIndex]
			if token.Separator {
				continue
			}
			className := token.Text
			if !seenInRun[className] && !seenInTemplate[className] {
				seenInRun[className] = true
				seenInTemplate[className] = true
				continue
			}

			found, exists := repeats[className]
			if !exists {
				found = &repeat{segment: segment, fixable: true}
				repeats[className] = found
				order = append(order, className)
			}
			if !seenInRun[className] || !tokenized {
				// Across a hole, or a run whose source is not its value: reported, not fixed.
				found.fixable = false
				seenInRun[className] = true
				continue
			}
			separator := tokens[tokenIndex-1].Range
			if separator.End() == separator.Pos()+1 {
				found.fixes = append(found.fixes, rule.ReplaceRange(core.NewTextRange(separator.Pos(), token.Range.End()), ""))
			} else {
				found.fixes = append(found.fixes,
					rule.ReplaceRange(core.NewTextRange(separator.Pos(), separator.Pos()+1), ""),
					rule.ReplaceRange(token.Range, ""),
				)
			}
		}
	}

	for _, className := range order {
		found := repeats[className]
		var fixes []rule.Fix
		if found.fixable {
			fixes = found.fixes
		}
		ctx.Report(rule.Diagnostic{
			Range:      found.segment.Range,
			Message:    messageDuplicateClass(className),
			SourceFile: ctx.SourceFile,
			Fixes:      fixes,
		})
	}
}

// repeatDeletions deletes every occurrence of a class after its first, each with the first byte of
// the separator before it.
//
// The first occurrence is kept, in its original position, which is what makes the fix invisible to
// `enforce-consistent-class-order`: the surviving classes are in the order the author wrote them,
// minus the ones that were saying nothing.
//
// Only the separator's first byte, because the rest of a long run is `no-unnecessary-whitespace`'s
// to delete, and the two edits then meet end to start rather than overlapping (class_tokens.go). A
// repeat always has a separator before it, since the occurrence it repeats came first.
func repeatDeletions(tokens []classToken, className string) []rule.Fix {
	fixes := []rule.Fix{}
	seen := false
	for index, token := range tokens {
		if token.Separator || token.Text != className {
			continue
		}
		if !seen {
			seen = true
			continue
		}
		separator := tokens[index-1].Range
		if separator.End() == separator.Pos()+1 {
			// Adjacent, so one edit rather than two that touch.
			fixes = append(fixes, rule.ReplaceRange(core.NewTextRange(separator.Pos(), token.Range.End()), ""))
			continue
		}
		fixes = append(fixes,
			rule.ReplaceRange(core.NewTextRange(separator.Pos(), separator.Pos()+1), ""),
			rule.ReplaceRange(token.Range, ""),
		)
	}
	return fixes
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
