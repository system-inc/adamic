// Package graphql is Prettier's GraphQL language: graphql-js's parser, ported here, and Prettier's
// printer-graphql, ported beside it.
package graphql

// graphql-js 17.0.2, language/characterClasses.js.
//
// The codes are what the lexer reads: a byte for ASCII, -1 past the end of the text (graphql-js's
// charCodeAt gives NaN there, which every one of these tests rejects), or a decoded rune where the
// lexer needs the whole character. Every class here is ASCII, so a byte answers it exactly.

// isWhiteSpace is the spec's WhiteSpace without the byte order mark, which the lexer handles itself.
//
//	WhiteSpace ::
//	  - "Horizontal Tab (U+0009)"
//	  - "Space (U+0020)"
func isWhiteSpace(code int) bool {
	return code == 0x0009 || code == 0x0020
}

// isDigit is the spec's Digit.
//
//	Digit :: one of
//	  - `0` `1` `2` `3` `4` `5` `6` `7` `8` `9`
func isDigit(code int) bool {
	return code >= 0x0030 && code <= 0x0039
}

// isLetter is the spec's Letter.
//
//	Letter :: one of
//	  - `A` `B` `C` `D` `E` `F` `G` `H` `I` `J` `K` `L` `M`
//	  - `N` `O` `P` `Q` `R` `S` `T` `U` `V` `W` `X` `Y` `Z`
//	  - `a` `b` `c` `d` `e` `f` `g` `h` `i` `j` `k` `l` `m`
//	  - `n` `o` `p` `q` `r` `s` `t` `u` `v` `w` `x` `y` `z`
func isLetter(code int) bool {
	return (code >= 0x0061 && code <= 0x007a) || // A-Z
		(code >= 0x0041 && code <= 0x005a) // a-z
}

// isNameStart is the spec's NameStart.
//
//	NameStart ::
//	  - Letter
//	  - `_`
func isNameStart(code int) bool {
	return isLetter(code) || code == 0x005f
}

// isNameContinue is the spec's NameContinue.
//
//	NameContinue ::
//	  - Letter
//	  - Digit
//	  - `_`
func isNameContinue(code int) bool {
	return isLetter(code) || isDigit(code) || code == 0x005f
}
