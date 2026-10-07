// A port of cohere's internal/format/css/mediaquery/whitespace.go to Adamic 0.1. The Go is beside it in
// the cohere submodule; each piece here names the Go it reads as.
//
// The library's whitespace is JavaScript's: the regular expression class \s, and String.prototype.trim,
// which strip the same set (ECMAScript WhiteSpace and LineTerminator). It is not Go's unicode.IsSpace,
// which takes U+0085 and leaves out U+FEFF. Every member is in the Basic Multilingual Plane, so one
// UTF-16 unit is one character here, and testing units answers as the library does. The Go tests runes
// for the same reason; the port tests units, because its strings are the library's, UTF-16.

// whitespace.go: isWhitespace, `character.search(/\s/) !== -1` for one character, here one UTF-16 unit.
export function isWhitespace(unit: number): boolean {
	switch (unit) {
		case 0x09:
		case 0x0a:
		case 0x0b:
		case 0x0c:
		case 0x0d:
		case 0x20:
		case 0xa0:
		case 0x1680:
		case 0x2028:
		case 0x2029:
		case 0x202f:
		case 0x205f:
		case 0x3000:
		case 0xfeff:
			return true;
	}
	return unit >= 0x2000 && unit <= 0x200a;
}

// whitespace.go: leadingWhitespace, `/^(\s*)/.exec(text)[1]`.
export function leadingWhitespace(text: string): string {
	let end = 0;
	while (end < text.length && isWhitespace(text.charCodeAt(end))) {
		end++;
	}
	return text.slice(0, end);
}

// whitespace.go: trailingWhitespace, `/(\s*)$/.exec(text)[1]`: the first position where \s* reaches the
// end, which is where the trailing run of whitespace starts.
export function trailingWhitespace(text: string): string {
	let start = text.length;
	while (start > 0 && isWhitespace(text.charCodeAt(start - 1))) {
		start--;
	}
	return text.slice(start);
}

// whitespace.go: trim, which the Go writes as strings.TrimFunc with isWhitespace because Go has no
// String.prototype.trim. The port has it, so it calls it, as the library does: stage 0's trim is then
// held to Node's on every whitespace character above.
export function trim(text: string): string {
	return text.trim();
}

// whitespace.go: startsWithWhitespace, `text[0].search(/\s/) !== -1` for a non-empty text. The port asks
// it of the unit at an index inside the text, since it never needs the rest.
export function startsWithWhitespace(text: string, index: number): boolean {
	return isWhitespace(text.charCodeAt(index));
}
