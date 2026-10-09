package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageInvalidRegexp = rule.Message{
	Id: "invalidRegexp",
	Description: "This RegExp constructor is handed a pattern or flags the engine cannot parse, so " +
		"the line throws a SyntaxError when it runs. A regex literal with the same mistake is caught " +
		"by the parser and never ships; written as a string it survives compilation, type checking " +
		"and review, and fails at the moment the constructor is reached.",
}

// validRegExpFlags are the flags the language defines. A flag outside this set is a SyntaxError
// rather than a no-op, which is why an unknown one is worth reporting rather than ignoring.
var validRegExpFlags = map[rune]bool{
	'd': true, 'g': true, 'i': true, 'm': true,
	's': true, 'u': true, 'v': true, 'y': true,
}

// NoInvalidRegexpOptions is the whole option surface. ESLint's schema is one object holding an
// `allowConstructorFlags` array of unique strings and `additionalProperties: false`.
type NoInvalidRegexpOptions struct {
	// AllowConstructorFlags names extra flag characters a constructor call may carry. Each string
	// contributes every character in it, as upstream joins them before reading.
	AllowConstructorFlags []string `json:"allowConstructorFlags"`
}

// DecodeNoInvalidRegexpOptions decodes the option object strictly and holds the schema's
// `uniqueItems`, which the generic decoder has no way to know about. A repeated string allows
// nothing a single one does not, but upstream refuses the config outright, and a config cohere loads
// where ESLint would not is a config that only works in one of them.
func DecodeNoInvalidRegexpOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[NoInvalidRegexpOptions]()(raw)
	if err != nil {
		return NoInvalidRegexpOptions{}, err
	}
	options, _ := decoded.(NoInvalidRegexpOptions)

	seen := make(map[string]bool, len(options.AllowConstructorFlags))
	for _, flags := range options.AllowConstructorFlags {
		if seen[flags] {
			return NoInvalidRegexpOptions{}, fmt.Errorf("allowConstructorFlags lists %q more than once, and its items must be unique", flags)
		}
		seen[flags] = true
	}
	return options, nil
}

// allowedConstructorFlags is the set of extra flags the options license, as upstream builds it: every
// character of every string, less the flags the language already defines. Dropping those is what
// keeps an allowed `u` from being struck twice, which would let `uu` through.
func allowedConstructorFlags(options any) map[rune]bool {
	allowed := map[rune]bool{}
	parsed, isParsed := rule.OptionsAs[NoInvalidRegexpOptions](options)
	if !isParsed {
		return allowed
	}
	for _, flags := range parsed.AllowConstructorFlags {
		for _, flag := range flags {
			if !validRegExpFlags[flag] {
				allowed[flag] = true
			}
		}
	}
	return allowed
}

// NoInvalidRegexp flags a RegExp constructor whose pattern or flags cannot be parsed.
//
//	valid:   new RegExp('[a-z]', 'i')
//	valid:   new RegExp(somePattern)          // not a literal, nothing to check
//	valid:   RegExp()                         // produces /(?:)/
//	invalid: new RegExp('[')
//	invalid: new RegExp('.', 'z')
//	invalid: new RegExp('.', 'ii')
//	invalid: new RegExp('.', 'uv')
//	valid:   new RegExp('.', 'a')             // with allowConstructorFlags: ["a"]
//
// Only a string literal is checked. `new RegExp(pattern)` where pattern is a variable could be
// anything at runtime, and reporting it would flag correct code; that is a deliberate narrowing
// rather than a gap, and it matches what ESLint does.
//
// The flags are validated before the pattern because a flag changes what the pattern means. `u`
// makes `\p{L}` a property escape and a lone `{` a syntax error, so validating a pattern against
// flags that are themselves invalid asks a question with no answer.
//
// When the flags are not a literal the pattern is only reported if it fails under every flag
// combination that could reach it. A pattern valid under `u` and invalid without it is not a defect
// here, because the flags argument may well supply the `u`.
//
// # A local named RegExp, and a pattern under `v`
//
// The constructor is checked only when `RegExp` is the global, asked of the checker, since
// `function foo(RegExp) { RegExp('['); }` calls whatever the caller passed. ESLint's corpus pins three
// such rows as clean, and they read as extra until the global was asked (#jjfa7qb).
//
// A pattern under `v` is not checked, only its flags. `v` has its own grammar (set operations like
// `[A--B]`, nested classes, `\q{...}`), and the engine this compiles with does not speak it, so
// reading the pattern as `u` reported valid `v` patterns, three of ESLint's rows, and passed invalid
// ones. Silence is the honest answer when cohere cannot parse what it is asked about. The full answer
// is an ECMAScript pattern validator in place of the engine (#4bgr3h2 tracks it).
//
// # allowConstructorFlags
//
// Upstream's one option names flag characters a constructor call may carry beyond the language's
// own, for an engine that defines extra ones. With `{"allowConstructorFlags": ["a"]}`,
// `new RegExp('.', 'a')` is clean and `new RegExp('.', 'aa')` still reports, as a duplicate. The
// allowed flags are threaded through the flag walk exactly as the language's are, struck once each,
// so a second copy is the same defect a second `g` is.
//
// Upstream's reading is followed to the character, because each step is observable in its corpus.
// The strings are joined and read character by character, so `["az"]` allows both `a` and `z`. The
// match is case-sensitive, so allowing `a` leaves `A` unknown. A flag the language already defines
// is dropped from the list, so allowing `u` does not license a second `u` and does not lift the rule
// that `u` and `v` cannot meet. With nothing configured the list is empty and every flag outside
// `validRegExpFlags` reports. The schema's `uniqueItems` is held at decode, so `["a", "a"]` is refused
// as upstream refuses it.
var NoInvalidRegexp = rule.Rule{
	Name: "no-invalid-regexp",

	// Whether `RegExp` is the global
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowedFlags := allowedConstructorFlags(options)

		check := func(node *ast.Node, callee *ast.Node, arguments *ast.NodeList) {
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "RegExp" {
				return
			}
			if ctx.TypeChecker == nil || !rule.IsDeclaredOnlyInDeclarationFiles(ctx.TypeChecker.GetSymbolAtLocation(callee)) {
				return
			}
			// `RegExp()` with no arguments produces the empty pattern and cannot fail.
			if arguments == nil || len(arguments.Nodes) == 0 {
				return
			}

			// A nil flags pointer means the second argument is not a literal, so what the flags
			// will be is unknowable here. That is different from the empty string, which is what a
			// one-argument call passes, and the two lead to different checks below.
			var flags *string
			if len(arguments.Nodes) >= 2 {
				if flagsNode := arguments.Nodes[1]; flagsNode.Kind == ast.KindStringLiteral {
					text := flagsNode.AsStringLiteral().Text
					flags = &text
				}
			} else {
				empty := ""
				flags = &empty
			}

			if flags != nil {
				if message := invalidFlagsMessage(*flags, allowedFlags); message != "" {
					ctx.ReportNode(node, rule.Message{Id: messageInvalidRegexp.Id, Description: message})
					return
				}
			}

			patternNode := arguments.Nodes[0]
			if patternNode.Kind != ast.KindStringLiteral {
				return
			}
			pattern := patternNode.AsStringLiteral().Text

			if flags != nil {
				// Under `v` the pattern has a grammar the engine does not parse; see the doc above
				if strings.ContainsRune(*flags, 'v') {
					return
				}
				if message := invalidPatternMessage(pattern, *flags); message != "" {
					ctx.ReportNode(node, rule.Message{Id: messageInvalidRegexp.Id, Description: message})
				}
				return
			}

			// Unknown flags: report only when no reachable flag combination accepts the pattern.
			// `v` could also reach it, and is unknowable here, so a pattern is reported only when
			// both readings the engine can check refuse it and `v` is the one left unasked
			withUnicode := invalidPatternMessage(pattern, "u")
			withNeither := invalidPatternMessage(pattern, "")
			if withUnicode != "" && withNeither != "" {
				ctx.ReportNode(node, rule.Message{Id: messageInvalidRegexp.Id, Description: withNeither})
			}
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				check(node, call.Expression, call.Arguments)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				expression := node.AsNewExpression()
				check(node, expression.Expression, expression.Arguments)
			},
		}
	},
}

// invalidFlagsMessage reports why a flags string is not accepted, or the empty string when it is.
//
// The order matters and follows the language rather than convenience: each known flag, the
// language's or one the options allow, is struck once, so what remains is either a duplicate or an
// unknown one. Checking `u` and `v` together first is what makes `uv` report as mutually exclusive
// rather than as two fine flags. The walk is by character rather than byte because an allowed flag
// is whatever character the config names, and upstream reads them as characters.
func invalidFlagsMessage(flags string, allowed map[rune]bool) string {
	known := func(flag rune) bool { return validRegExpFlags[flag] || allowed[flag] }

	struck := map[rune]bool{}
	var remaining []rune
	for _, flag := range flags {
		if known(flag) && !struck[flag] {
			struck[flag] = true
			continue
		}
		remaining = append(remaining, flag)
	}

	if strings.Contains(flags, "u") && strings.Contains(flags, "v") {
		return fmt.Sprintf("Invalid regular expression flags: %s. The u and v flags cannot both be set", flags)
	}
	for _, flag := range remaining {
		if known(flag) {
			return fmt.Sprintf("Invalid regular expression flags: %s. Duplicate flag %q", flags, flag)
		}
	}
	if len(remaining) > 0 {
		return fmt.Sprintf("Invalid regular expression flags: %s. Unknown flag %q", flags, remaining[0])
	}
	return ""
}

// invalidPatternMessage reports why a pattern is not accepted under the given flags.
//
// Only the flags that change what a pattern means are passed to the parser: the rest affect
// matching rather than parsing, and handing them over would ask the parser about something it does
// not decide. `v` extends `u` rather than replacing it, so a pattern under `v` is parsed as `u`.
func invalidPatternMessage(pattern string, flags string) string {
	var parseFlags strings.Builder
	for _, flag := range "imsu" {
		if strings.ContainsRune(flags, flag) {
			parseFlags.WriteRune(flag)
		}
	}
	compileFlags := parseFlags.String()
	if !strings.Contains(compileFlags, "u") && strings.Contains(flags, "v") {
		compileFlags += "u"
	}

	if _, err := esregexp.Compile(pattern, compileFlags); err != nil {
		text := err.Error()
		// The underlying engine prefixes "error parsing regexp: ", which repeats what the message
		// already says. Trimming to the first ": " drops it. Our parser's own refusals put the
		// refused construct after that same separator, so trimming there would keep the construct
		// and throw away the explanation, which is why they are excluded.
		if !errors.Is(err, esregexp.ErrUnsupportedSyntax) && !errors.Is(err, esregexp.ErrUnsupportedFlag) {
			if separator := strings.Index(text, ": "); separator != -1 {
				text = text[separator+2:]
			}
		}
		return fmt.Sprintf("Invalid regular expression: /%s/: %s", pattern, text)
	}
	return ""
}
