package tailwind

import "github.com/microsoft/TypeScript/tsc/shim/core"

// classToken is one class, or one run of whitespace between classes, at its own range in the source.
//
// The four rules in this package that fix (`no-duplicate-classes`, `no-unnecessary-whitespace`,
// `no-deprecated-classes` and `enforce-consistent-class-order`) all used to rewrite a literal whole.
// The edit engine lands one edit per range per pass and refuses the rest as overlapping, so a
// shuffled, duplicated, badly spaced literal took a pass per rule, and every pass but the last
// counted refusals. Measured by @system_cohere_format_markdown on the perturbed corpus (#vf1hd6j):
// 6,847 overlap refusals in ahra and 8,517 in www-phi-health, every one of them landing a pass
// later. Nothing was lost, and a run that refuses thousands of fixes to land them reads as a run
// that is fighting itself.
//
// So each fixer edits only its own bytes, and the bytes are divided up here:
//
//   - a class belongs to the rule changing that class: `no-deprecated-classes` renames it,
//     `enforce-consistent-class-order` writes the class that belongs in its slot, and
//     `no-duplicate-classes` deletes a repeat
//   - the first character of the separator before a repeat belongs to `no-duplicate-classes`, which
//     deletes it along with the repeat so the list closes up behind it
//   - every other whitespace byte belongs to `no-unnecessary-whitespace`, which deletes the excess
//     past a separator's first character and the padding at either edge
//
// The rules still decide alone: none of them reads another's verdict. What they share is the claim
// on each byte, and two claims never cross, so every fix lands in the pass that proposes it.
type classToken struct {
	Text      string
	Range     core.TextRange
	Separator bool
}

// classTokensIn splits a run of class text into its classes and separators, at source offsets.
//
// False when the source bytes in the range are not the value the parser decoded. An escape such as
// ` ` decodes to a space, so the decoded text's offsets do not line up with the source's, and a
// token range computed from one would cut the other in the wrong place. A caller that gets false
// falls back to whatever it did before tokens existed, or reports without a fix.
//
// Separators are `isSpace`, the set `no-unnecessary-whitespace` has always used, so a separator here
// is one that rule would have seen.
func classTokensIn(sourceText string, textRange core.TextRange, value string) ([]classToken, bool) {
	start, end := textRange.Pos(), textRange.End()
	if start < 0 || end > len(sourceText) || start > end || sourceText[start:end] != value {
		return nil, false
	}
	return classTokensOf(value, start), true
}

// classTokensOf splits a run of class text into its classes and separators, with ranges offset by
// start. A caller that has no trustworthy source offset passes the text alone and uses only the
// tokens' text, never their ranges.
func classTokensOf(value string, start int) []classToken {
	tokens := []classToken{}
	index := 0
	for index < len(value) {
		tokenStart := index
		separator := isSpace(rune(value[index]))
		for index < len(value) && isSpace(rune(value[index])) == separator {
			index++
		}
		tokens = append(tokens, classToken{
			Text:      value[tokenStart:index],
			Range:     core.NewTextRange(start+tokenStart, start+index),
			Separator: separator,
		})
	}
	return tokens
}

// classesOf is the class tokens alone, in source order.
func classesOf(tokens []classToken) []classToken {
	classes := make([]classToken, 0, len(tokens)/2+1)
	for _, token := range tokens {
		if !token.Separator {
			classes = append(classes, token)
		}
	}
	return classes
}
