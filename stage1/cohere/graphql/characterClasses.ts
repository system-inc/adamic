// A port of cohere's internal/format/graphql/character_classes.go to Adamic 0.1: graphql-js 17.0.2,
// language/characterClasses.js.
//
// The codes are what the lexer reads with charCodeAt: a UTF-16 code unit, or NaN past the end of the
// text, which every one of these tests rejects.

// isWhiteSpace is the spec's WhiteSpace without the byte order mark, which the lexer handles itself.
//
//	WhiteSpace ::
//	  - "Horizontal Tab (U+0009)"
//	  - "Space (U+0020)"
export function isWhiteSpace(code: number): boolean {
	return code === 0x0009 || code === 0x0020;
}

// isDigit is the spec's Digit.
//
//	Digit :: one of
//	  - `0` `1` `2` `3` `4` `5` `6` `7` `8` `9`
export function isDigit(code: number): boolean {
	return code >= 0x0030 && code <= 0x0039;
}

// isLetter is the spec's Letter.
//
//	Letter :: one of
//	  - `A` `B` `C` `D` `E` `F` `G` `H` `I` `J` `K` `L` `M`
//	  - `N` `O` `P` `Q` `R` `S` `T` `U` `V` `W` `X` `Y` `Z`
//	  - `a` `b` `c` `d` `e` `f` `g` `h` `i` `j` `k` `l` `m`
//	  - `n` `o` `p` `q` `r` `s` `t` `u` `v` `w` `x` `y` `z`
export function isLetter(code: number): boolean {
	return (
		(code >= 0x0061 && code <= 0x007a) || // a-z
		(code >= 0x0041 && code <= 0x005a) // A-Z
	);
}

// isNameStart is the spec's NameStart.
//
//	NameStart ::
//	  - Letter
//	  - `_`
export function isNameStart(code: number): boolean {
	return isLetter(code) || code === 0x005f;
}

// isNameContinue is the spec's NameContinue.
//
//	NameContinue ::
//	  - Letter
//	  - Digit
//	  - `_`
export function isNameContinue(code: number): boolean {
	return isLetter(code) || isDigit(code) || code === 0x005f;
}
