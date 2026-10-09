package tailwind

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageVariableSyntax names the class and the spelling it should carry.
//
// Upstream's message names only the class. The replacement is added here for the reason the sibling
// rules add it: a reader told a class is wrong without being told what right looks like has to go
// read the rule to act on the finding.
func messageVariableSyntax(written string, corrected string) rule.Message {
	return rule.Message{
		Id: "variableSyntax",
		Description: "The custom property in \"" + written + "\" is written as an arbitrary value. " +
			"Tailwind v4 reads parentheses as a custom property and square brackets as an arbitrary " +
			"value that happens to look like one, so \"" + corrected + "\" is the spelling that says " +
			"what this means. Both resolve today, which is why the two forms accumulate side by side " +
			"until a reader cannot tell whether a difference between two classes is deliberate. " +
			"Write it as \"" + corrected + "\".",
	}
}

// EnforceConsistentVariableSyntaxOptions lets a project name the surfaces and pick the target form.
//
// Syntax selects which spelling is canonical: "shorthand" is v4's parentheses and is the default,
// "variable" is the bracketed `var()` form. Named rather than inferred, so a repository can pick a
// target instead of having one read off its node_modules.
type EnforceConsistentVariableSyntaxOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
	Syntax string `json:"syntax"`
}

// EnforceConsistentVariableSyntax reports a custom property written as an arbitrary value.
//
//	valid:   <div className="text-(--brand)" />
//	invalid: <div className="text-[--brand]" />
//	invalid: <div className="text-[var(--brand)]" />
//
// # What the two forms actually mean
//
// Tailwind v4 added `(--x)` as the spelling for a custom property, and reads `[--x]` as an arbitrary
// value whose text happens to begin with two dashes. Both resolve to the same declaration today, so
// nothing surfaces the difference. What it costs is that a codebase ends up holding all three
// spellings of one idea, and the next person writing one picks whichever they saw last.
//
// Both bracketed forms collapse onto the parenthesised one, which is upstream's behaviour and the
// part a reader is most likely to predict wrongly: `text-[--x]` and `text-[var(--x)]` are both
// reported and both corrected to `text-(--x)`.
//
// # Two shapes are deliberately left alone
//
// A class whose base holds a colon is skipped, which is upstream's "skip variable definitions"
// guard. `[color:red]` names a property rather than reads one, and rewriting its brackets would
// change what it means rather than how it is spelled.
//
// A bracket whose contents are a function CALL other than `var` is not a custom property at all.
// Upstream's own test is `/^_*--(?![\w-]+\()/`, whose lookahead exists to decline exactly that, so
// `w-[calc(var(--x)*2)]` is silent: the outer brackets hold arithmetic, not a property reference.
//
// This rule asks nothing of the design system: the question is entirely about how the class is
// spelled, so it declares no ProgramReads.
//
// Not fixable, for the reason its siblings are not: a class literal here wraps across lines with
// indentation that carries intent, and a fixer would be the first thing to reflow it.
var EnforceConsistentVariableSyntax = rule.Rule{
	Name: "better-tailwindcss/enforce-consistent-variable-syntax",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		surfaces := DefaultClassLiteralSurfaces()
		syntax := variableSyntaxShorthand
		if configured, isConfigured := rule.OptionsAs[EnforceConsistentVariableSyntaxOptions](options); isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
			if configured.Syntax == variableSyntaxVariable {
				syntax = variableSyntaxVariable
			}
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				for _, className := range SplitClasses(literal.Text) {
					corrected, isMisspelled := correctedVariableSyntax(className, syntax)
					if !isMisspelled {
						continue
					}
					ctx.ReportRange(literal.Range, messageVariableSyntax(className, corrected))
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
	// variableSyntaxShorthand is v4's parenthesised custom property: `text-(--x)`.
	variableSyntaxShorthand = "shorthand"
	// variableSyntaxVariable is the bracketed call form: `text-[var(--x)]`.
	variableSyntaxVariable = "variable"
)

// correctedVariableSyntax returns the class with its custom property in the configured form.
func correctedVariableSyntax(className string, syntax string) (string, bool) {
	prefix, utility, hasVariants := splitVariantPrefix(className)
	if !hasVariants {
		utility = className
		prefix = ""
	}

	// The modifier is separated first and put back untouched. `text-[--v]/50` carries its opacity
	// outside the brackets, and folding it into the base would rewrite the wrong span.
	base, modifier := splitClassModifier(utility)

	// Upstream's "skip variable definitions": a colon in the base means the brackets name a property
	// rather than read one, so rewriting them would change meaning rather than spelling.
	if strings.Contains(base, ":") {
		return "", false
	}

	rebuilt, changed := rewriteVariableForm(base, syntax)
	if !changed {
		return "", false
	}

	rebuilt += modifier
	if prefix != "" {
		rebuilt = prefix + ":" + rebuilt
	}
	if rebuilt == className {
		return "", false
	}
	return rebuilt, true
}

// rewriteVariableForm swaps one balanced group between the bracketed and parenthesised spellings.
func rewriteVariableForm(base string, syntax string) (string, bool) {
	if syntax == variableSyntaxShorthand {
		before, contents, after, found := extractBalanced(base, '[', ']')
		if !found {
			return "", false
		}
		switch {
		case isArbitraryVariable(contents):
			// `[var(--x)]` holds a call; its own parentheses carry the property name.
			_, inner, trailing, hasInner := extractBalanced(contents, '(', ')')
			if !hasInner || text.TrimWhitespace(strings.ReplaceAll(trailing, "_", " ")) != "" {
				return "", false
			}
			return before + "(" + inner + ")" + after, true
		case isArbitraryShorthand(contents):
			// `[--x]` is the bare property in brackets.
			return before + "(" + contents + ")" + after, true
		}
		return "", false
	}

	before, contents, after, found := extractBalanced(base, '[', ']')
	if found && isArbitraryVariable(contents) {
		return "", false
	}
	if found && isArbitraryShorthand(contents) {
		return before + "[var(" + contents + ")]" + after, true
	}

	beforeParen, parenContents, afterParen, hasParen := extractBalanced(base, '(', ')')
	if hasParen && isArbitraryShorthand(parenContents) {
		return beforeParen + "[var(" + parenContents + ")]" + afterParen, true
	}
	return "", false
}

// isArbitraryVariable is upstream's `/^_*var\(_*--/`.
func isArbitraryVariable(contents string) bool {
	rest := strings.TrimLeft(contents, "_")
	if !strings.HasPrefix(rest, "var(") {
		return false
	}
	rest = strings.TrimLeft(rest[len("var("):], "_")
	return strings.HasPrefix(rest, "--")
}

// isArbitraryShorthand is upstream's `/^_*--(?![\w-]+\()/`.
//
// The lookahead is the whole point and is easy to drop: it declines a bracket whose contents are a
// function call rather than a property reference, which is what keeps `w-[calc(var(--x)*2)]` silent.
func isArbitraryShorthand(contents string) bool {
	rest := strings.TrimLeft(contents, "_")
	if !strings.HasPrefix(rest, "--") {
		return false
	}
	for index := len("--"); index < len(rest); index++ {
		character := rest[index]
		if character == '(' {
			return false
		}
		isNameCharacter := character == '-' || character == '_' ||
			(character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9')
		if !isNameCharacter {
			return true
		}
	}
	return true
}

// extractBalanced returns the text around the first balanced group of the given delimiters.
func extractBalanced(text string, open byte, close byte) (string, string, string, bool) {
	start := strings.IndexByte(text, open)
	if start < 0 {
		return "", "", "", false
	}
	depth := 0
	for index := start; index < len(text); index++ {
		switch text[index] {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return text[:start], text[start+1 : index], text[index+1:], true
			}
		}
	}
	return "", "", "", false
}

// splitClassModifier separates a trailing `/opacity` from the utility it modifies.
//
// Split on the last slash outside brackets, so an arbitrary value holding one keeps it.
func splitClassModifier(utility string) (string, string) {
	depth := 0
	for index := len(utility) - 1; index >= 0; index-- {
		switch utility[index] {
		case ']', ')':
			depth++
		case '[', '(':
			depth--
		case '/':
			if depth == 0 {
				return utility[:index], utility[index:]
			}
		}
	}
	return utility, ""
}
