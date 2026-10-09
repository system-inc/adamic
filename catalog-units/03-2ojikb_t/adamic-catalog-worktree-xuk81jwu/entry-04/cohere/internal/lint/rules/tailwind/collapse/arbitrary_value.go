// decodeArbitraryValue and isValidArbitrary: the two functions every arbitrary bracket in a class
// string is routed through before the candidate parser will look at it.
//
// Ports of `src/utils/decode-arbitrary-value.ts`, `src/utils/is-valid-arbitrary.ts` and the
// `addWhitespaceAroundMathOperators` half of `src/utils/math-operators.ts`, all at Tailwind 4.3.3.
// The `hasMathFn` half of that last file already lives in segment.go, next to the data-type checks
// that were its only caller until now.
//
// # Why these live with the candidate parser and not with the value parser
//
// They read like value-parser utilities and they are not. `decodeArbitraryValue` is the function
// that turns the on-the-wire spelling of a class into the CSS it means: `w-[calc(100%-1rem)]` is
// written without spaces because a class attribute cannot contain them, and `_` stands in for a
// space because the same constraint applies. Nothing outside candidate parsing has a reason to
// perform that substitution, and performing it anywhere else would corrupt real CSS, where `_` is a
// legal identifier character rather than an escape.
//
// # The order of the two transformations is load-bearing
//
// Underscores are decoded through the value AST, so a `_` inside `url(…)` survives and a `_` inside
// `var(…)`'s first argument survives, because a custom-property name may legitimately contain one.
// Only then are spaces added around math operators, working on the already-decoded text. Running
// them the other way would add spaces around the `-` in a custom-property name that had not yet
// been recognised as one.
package tailwind

import "strings"

// isValidArbitrary reports whether a string may be used as an arbitrary value.
//
// Unbalanced parens and brackets are rejected, and so is a top-level `;`. A top-level `}` is
// rejected as well, which is the part worth stating plainly: `{` deliberately does not push onto
// the stack, so a `}` can never be balanced and `[&{color:red}]:flex` is invalid however it is
// spelled. Upstream annotates that as intentional and it is the whole reason the two brackets are
// treated asymmetrically.
//
// Upstream calls this "very similar to segment" and explains why segment cannot be reused: segment
// splits on a separator and this needs to split on a bracket-stack character. The same reasoning
// applies to this port, so this is a second small scanner rather than a parameter on the first.
func isValidArbitrary(input string) bool {
	// Upstream uses a shared 256-byte buffer and never bounds-checks it, on the reasoning that the
	// buffer may be left dirty because reading and writing always start at position zero. A slice
	// here costs an allocation only on nested input, and removes the fixed depth limit along with
	// the data race a shared buffer would be in a package the linter runs concurrently.
	var closingBracketStack []byte

	for index := 0; index < len(input); index++ {
		character := input[index]

		switch character {
		case '\\':
			// The next byte is escaped, so skip it.
			index++
		case '\'', '"':
			// Strings run as-is to their close; no bracket balancing happens inside one.
			for index+1 < len(input) {
				index++
				next := input[index]
				if next == '\\' {
					index++
					continue
				}
				if next == character {
					break
				}
			}
		case '(':
			closingBracketStack = append(closingBracketStack, ')')
		case '[':
			closingBracketStack = append(closingBracketStack, ']')
		case '{':
			// Intentionally does not push. See the doc comment: this is what makes a top-level `}`
			// unbalanceable rather than an oversight.
		case ']', '}', ')':
			if len(closingBracketStack) == 0 {
				return false
			}
			if character == closingBracketStack[len(closingBracketStack)-1] {
				closingBracketStack = closingBracketStack[:len(closingBracketStack)-1]
			}
		case ';':
			if len(closingBracketStack) == 0 {
				return false
			}
		}
	}

	return true
}

// decodeArbitraryValue turns the class-string spelling of an arbitrary value into the CSS it means.
//
// Two transformations, in this order: underscores become spaces (except where they must not), then
// whitespace is added around math operators. The early return when the input holds no `(` is
// upstream's and is not just an optimization: it is the path that keeps a value with no functions
// from being round-tripped through the value parser at all.
func decodeArbitraryValue(input string) string {
	if !strings.ContainsRune(input, '(') {
		return convertUnderscoresToWhitespace(input, false)
	}

	ast := ParseValue(input)
	recursivelyDecodeArbitraryValues(ast)
	input = ValueToCss(ast)

	return addWhitespaceAroundMathOperators(input)
}

// convertUnderscoresToWhitespace turns `_` into a space and `\_` into a literal `_`.
//
// skipUnderscoreToSpace suppresses only the first half: an escaped underscore is still unescaped.
// That combination has exactly one caller, the first argument of `var(…)`, where `--my_var` must
// keep its underscore because a custom-property name may contain one, while `--my\_var` must still
// lose its backslash.
func convertUnderscoresToWhitespace(input string, skipUnderscoreToSpace bool) string {
	// Nothing to do is the common case for a value that reached here through the AST, where most
	// nodes hold no underscore at all.
	if !strings.ContainsAny(input, "_\\") {
		return input
	}

	var output strings.Builder
	output.Grow(len(input))
	for index := 0; index < len(input); index++ {
		character := input[index]

		switch {
		case character == '\\' && index+1 < len(input) && input[index+1] == '_':
			output.WriteByte('_')
			index++
		case character == '_' && !skipUnderscoreToSpace:
			output.WriteByte(' ')
		default:
			output.WriteByte(character)
		}
	}
	return output.String()
}

// recursivelyDecodeArbitraryValues walks the value AST applying the underscore substitution, with
// the three exemptions upstream carves out.
//
// The exemptions are not cosmetic. A `url()` keeps its underscores because a URL path may contain
// one and there is no other way to write it. `var()` and `theme()` keep the underscores of their
// first argument because that argument names a custom property. Both tests are `name == x ||
// strings.HasSuffix(name, "_"+x)`, matching upstream, because a function name that itself arrived
// with an escaped underscore has not been decoded yet at the moment it is tested.
func recursivelyDecodeArbitraryValues(ast []ValueNode) {
	for index := range ast {
		node := &ast[index]
		switch node.Kind {
		case ValueNodeKindFunction:
			if node.Value == "url" || strings.HasSuffix(node.Value, "_url") {
				// The name is decoded, the arguments are left alone entirely.
				node.Value = convertUnderscoresToWhitespace(node.Value, false)
				continue
			}

			if node.Value == "var" || strings.HasSuffix(node.Value, "_var") ||
				node.Value == "theme" || strings.HasSuffix(node.Value, "_theme") {
				node.Value = convertUnderscoresToWhitespace(node.Value, false)
				for argument := range node.Nodes {
					// The first argument keeps its underscores when it is a word, because it names
					// a custom property. A first argument that is a separator gets the ordinary
					// treatment, which is upstream's behaviour and not an oversight: the guard
					// tests the kind, not just the position.
					if argument == 0 && node.Nodes[argument].Kind == ValueNodeKindWord {
						node.Nodes[argument].Value = convertUnderscoresToWhitespace(node.Nodes[argument].Value, true)
						continue
					}
					recursivelyDecodeArbitraryValues(node.Nodes[argument : argument+1])
				}
				continue
			}

			node.Value = convertUnderscoresToWhitespace(node.Value, false)
			recursivelyDecodeArbitraryValues(node.Nodes)

		case ValueNodeKindSeparator, ValueNodeKindWord:
			node.Value = convertUnderscoresToWhitespace(node.Value, false)
		}
	}
}

// addWhitespaceAroundMathOperators spaces out the operators inside math functions.
//
// `w-[calc(100%-1rem)]` has to be written without spaces and means `calc(100% - 1rem)`, which is a
// different expression: without the spaces `100%-1rem` is a single token to a CSS parser. This is
// the function that makes those two the same thing.
//
// It is a character loop with a stack of booleans rather than an AST walk, matching upstream, and
// the reason is visible in the branches: it needs to know whether the character before the operator
// was a digit, whether the one two back was, and whether the previous position ended a value plus
// its unit. Those are questions about the emitted text, not about the tree.
func addWhitespaceAroundMathOperators(input string) string {
	// The bail is on the bare function name appearing anywhere, not on `name(`, so it is wider than
	// hasMathFunction in segment.go. Two different tests, both upstream's, and they are not
	// interchangeable: this one must not bail on input a later branch would have formatted.
	found := false
	for _, function := range mathFunctions {
		if strings.Contains(input, function) {
			found = true
			break
		}
	}
	if !found {
		return input
	}

	var result strings.Builder
	result.Grow(len(input) + 16)

	// formattable is a stack, innermost first, matching upstream's `unshift`/`shift` on the front.
	// A slice used as a stack from the back would be the idiomatic Go shape and would invert the
	// indexing every branch does, so this keeps upstream's orientation and pays for it with the
	// prepend, which happens once per `(` rather than once per character.
	var formattable []bool

	// valuePosition tracks a run that looks like a number followed by a unit, so `1px` is known not
	// to be a function name. lastValuePosition remembers where the previous such run ended, which
	// is what one of the spacing conditions below tests.
	valuePosition := -1
	lastValuePosition := -1

	// resultEndsWith reports the last byte written, which several branches need. Reading back from
	// a strings.Builder is not possible, so the last byte is tracked as it is written.
	lastWritten := byte(0)
	write := func(text string) {
		if len(text) == 0 {
			return
		}
		result.WriteString(text)
		lastWritten = text[len(text)-1]
	}

	// trimmedTail returns the last two bytes of the result ignoring trailing spaces, which is what
	// upstream's `result.trimEnd()` is used for and the only thing it is used for.
	trimmedTail := func() (previous byte, previousPrevious byte) {
		text := strings.TrimRight(result.String(), " \t\n\r\f\v")
		if len(text) >= 1 {
			previous = text[len(text)-1]
		}
		if len(text) >= 2 {
			previousPrevious = text[len(text)-2]
		}
		return previous, previousPrevious
	}

	isDigit := func(character byte) bool { return character >= '0' && character <= '9' }
	isOperator := func(character byte) bool {
		return character == '+' || character == '-' || character == '*' || character == '/'
	}

	for index := 0; index < len(input); index++ {
		character := input[index]

		switch {
		case isDigit(character):
			valuePosition = index
		case valuePosition != -1 && (character == '%' ||
			(character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z')):
			valuePosition = index
		default:
			lastValuePosition = valuePosition
			valuePosition = -1
		}

		switch {
		case character == '(':
			write(string(character))

			// Scan backwards for the function name, which upstream assumes is lowercase
			// alphanumeric. A name with a dash or an uppercase letter stops the scan and therefore
			// reads as a different, unknown function, which is upstream's behaviour.
			start := index
			for back := index - 1; back >= 0; back-- {
				inner := input[back]
				if isDigit(inner) || (inner >= 'a' && inner <= 'z') {
					start = back
					continue
				}
				break
			}
			name := input[start:index]

			switch {
			case isMathFunctionName(name):
				formattable = append([]bool{true}, formattable...)
			case len(formattable) > 0 && formattable[0] && name == "":
				// A bare `(` nested inside a math function keeps formatting, so
				// `calc((1px+2px)*3)` spaces both levels.
				formattable = append([]bool{true}, formattable...)
			default:
				formattable = append([]bool{false}, formattable...)
			}

		case character == ')':
			write(string(character))
			if len(formattable) > 0 {
				formattable = formattable[1:]
			}

		case character == ',' && len(formattable) > 0 && formattable[0]:
			write(", ")

		case character == ' ' && len(formattable) > 0 && formattable[0] && lastWritten == ' ':
			// Consecutive whitespace inside a math function collapses.

		case isOperator(character) && len(formattable) > 0 && formattable[0]:
			previous, previousPrevious := trimmedTail()
			var next byte
			if index+1 < len(input) {
				next = input[index+1]
			}

			switch {
			// Scientific notation, e.g. `-3.4e-2`: the `-` is part of the number.
			case (previous == 'e' || previous == 'E') && isDigit(previousPrevious):
				write(string(character))
			// Preceded by an operator, so this one is a sign rather than a binary operator.
			case isOperator(previous):
				write(string(character))
			// At the start of an argument, likewise a sign.
			case previous == '(' || previous == ',':
				write(string(character))
			// Already spaced before, so only the trailing space is missing.
			case index > 0 && input[index-1] == ' ':
				write(string(character) + " ")
			case isDigit(previous) || isDigit(next) || previous == ')' || next == '(' ||
				isOperator(next) ||
				(lastValuePosition != -1 && lastValuePosition == index-1):
				write(" " + string(character) + " ")
			default:
				write(string(character))
			}

		default:
			write(string(character))
		}
	}

	return result.String()
}

// isMathFunctionName reports whether name is one of the math functions, matching upstream's
// `MATH_FUNCTIONS.includes(fn)` rather than the substring test hasMathFunction performs.
func isMathFunctionName(name string) bool {
	for _, function := range mathFunctions {
		if function == name {
			return true
		}
	}
	return false
}
