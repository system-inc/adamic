// A port of cohere's internal/format/graphql/lexer.go to Adamic 0.1: graphql-js 17.0.2,
// language/lexer.js, with error/syntaxError.js and language/location.js's getLocation, which give a
// lexing or parsing failure its message and position.
//
// The Go walks UTF-8 bytes where graphql-js walks UTF-16 code units; this port walks code units with
// charCodeAt, as graphql-js does, so its surrogate pair branches are live here where the Go's are not.
// NaN past the end of the text fails every test the lexer makes, as in graphql-js.
//
// graphql-js throws a GraphQLError where lexing or parsing fails, and the Go panics with a
// *SyntaxError that Parse recovers. Here each failure is `throw new Error(syntaxError(...))`, caught in
// parser.ts's parse: syntaxError gives the message, since stage 0 throws only an Error made where it is
// thrown (GAPS.md, gap 2). The message is Prettier's, GraphQLError's message and then its first
// location, " (line:column)".
//
// Every string the lexer gives, a token's value or a message, is written in the test's output form,
// `escaped` below, rather than as the text itself: a String token's \u escapes can't be cooked into
// characters without String.fromCodePoint (gitignore's GAPS.md, gap 1; GAPS.md here, "Cooking a string
// without String.fromCodePoint"), and every other string is written the same way so that the output
// is one form.

import { panic } from 'adamic';
import { dedentBlockStringLines } from './blockString.ts';
import { isDigit, isNameContinue, isNameStart } from './characterClasses.ts';
import { Token, type TokenKind } from './token.ts';

// The printable ASCII characters, from U+0020 to U+007E, by code point less 0x20: a character from its
// code, which String.fromCodePoint would give (gap 1).
const printable = ' !"#$%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~';

// written is how the output form writes one code point: printable ASCII as itself, but a backslash as
// two, and anything else as \u{HEX}. A lone surrogate, which only a slice through a surrogate pair can
// make, is written as U+FFFD, which is what Node writes for one when the message leaves as UTF-8 and
// what the Go's sliceUnits makes of it.
export function written(code: number): string {
	if (code === 0x5c) {
		return '\\\\';
	}
	if (code >= 0x20 && code <= 0x7e) {
		return printable[code - 0x20] ?? panic(`no printable character ${code}`);
	}
	if (code >= 0xd800 && code <= 0xdfff) {
		return '\\u{FFFD}';
	}
	return `\\u{${code.toString(16).toUpperCase()}}`;
}

// escaped is a text in the output form, each code point as written writes it. It copies the runs that
// need no escape whole.
export function escaped(text: string): string {
	let result = '';
	let runStart = 0;
	let index = 0;
	while (index < text.length) {
		const code = text.codePointAt(index) ?? panic(`no code point at ${index}`);
		const size = code > 0xffff ? 2 : 1;
		if (code < 0x20 || code > 0x7e || code === 0x5c) {
			result += text.slice(runStart, index) + written(code);
			runStart = index + size;
		}
		index += size;
	}
	return result + text.slice(runStart);
}

interface Location {
	readonly line: number;
	readonly column: number;
}

// getLocation takes a Source and a UTF-16 offset, and returns the corresponding line and column as a
// SourceLocation. Lines are split by /\r\n|[\n\r]/g, as graphql-js splits them.
function getLocation(body: string, position: number): Location {
	let lastLineStart = 0;
	let line = 1;
	for (let index = 0; index < body.length; index++) {
		const code = body.charCodeAt(index);
		if (code !== 0x000a && code !== 0x000d) {
			continue;
		}
		if (index >= position) {
			break;
		}
		const length = code === 0x000d && body.charCodeAt(index + 1) === 0x000a ? 2 : 1;
		lastLineStart = index + length;
		line++;
		index += length - 1;
	}
	return { line, column: position + 1 - lastLineStart };
}

// syntaxError produces the message of a GraphQLError representing a syntax error, containing useful
// descriptive information about the syntax error's position in the source, as Prettier reports it. The
// description is in the output form already.
export function syntaxError(body: string, position: number, description: string): string {
	const location = getLocation(body, position);
	return `Syntax Error: ${description} (${location.line}:${location.column})`;
}

export function isPunctuatorTokenKind(kind: TokenKind): boolean {
	return (
		kind === '!' ||
		kind === '$' ||
		kind === '&' ||
		kind === '(' ||
		kind === ')' ||
		kind === '.' ||
		kind === '...' ||
		kind === ':' ||
		kind === '=' ||
		kind === '@' ||
		kind === '[' ||
		kind === ']' ||
		kind === '{' ||
		kind === '|' ||
		kind === '}'
	);
}

// isUnicodeScalarValue: a Unicode scalar value is any Unicode code point except surrogate code points.
// In other words, the inclusive ranges of values 0x0000 to 0xD7FF and 0xE000 to 0x10FFFF.
//
// SourceCharacter ::
//   - "Any Unicode scalar value"
function isUnicodeScalarValue(code: number): boolean {
	return (code >= 0x0000 && code <= 0xd7ff) || (code >= 0xe000 && code <= 0x10ffff);
}

function isLeadingSurrogate(code: number): boolean {
	return code >= 0xd800 && code <= 0xdbff;
}

function isTrailingSurrogate(code: number): boolean {
	return code >= 0xdc00 && code <= 0xdfff;
}

// isSupplementaryCodePoint: the GraphQL specification defines source text as a sequence of unicode
// scalar values (which Unicode defines to exclude surrogate code points). However JavaScript defines
// strings as a sequence of UTF-16 code units which may include surrogates. A surrogate pair is a valid
// source character as it encodes a supplementary code point (above U+FFFF), but unpaired surrogate code
// points are not valid source characters.
function isSupplementaryCodePoint(body: string, location: number): boolean {
	return isLeadingSurrogate(body.charCodeAt(location)) && isTrailingSurrogate(body.charCodeAt(location + 1));
}

// printCodePointAt prints the code point (or end of file reference) at a given location in a source for
// use in error messages.
//
// Printable ASCII is printed quoted, while other points are printed in Unicode code point form (ie.
// U+1234).
function printCodePointAt(lexer: Lexer, location: number): string {
	const code = lexer.body.codePointAt(location);
	if (code === undefined) {
		return '<EOF>';
	} else if (code >= 0x0020 && code <= 0x007e) {
		const character = written(code);
		return character === '"' ? `'"'` : `"${character}"`;
	}
	return 'U+' + code.toString(16).toUpperCase().padStart(4, '0');
}

// readHexDigit reads a hexadecimal character and returns its positive integer value (0-15).
//
// '0' becomes 0, '9' becomes 9
// 'A' becomes 10, 'F' becomes 15
// 'a' becomes 10, 'f' becomes 15
//
// Returns -1 if the provided character code was not a valid hexadecimal digit.
//
// HexDigit :: one of
//   - `0` `1` `2` `3` `4` `5` `6` `7` `8` `9`
//   - `A` `B` `C` `D` `E` `F`
//   - `a` `b` `c` `d` `e` `f`
function readHexDigit(code: number): number {
	return code >= 0x0030 && code <= 0x0039
		? code - 0x0030 // 0-9
		: code >= 0x0041 && code <= 0x0046
			? code - 0x0037 // A-F
			: code >= 0x0061 && code <= 0x0066
				? code - 0x0057 // a-f
				: -1;
}

// read16BitHexCode reads four hexadecimal characters and returns the positive integer that 16bit
// hexadecimal string represents. For example, "000f" will return 15, and "dead" will return 57005.
//
// Returns a negative number if any char was not a valid hexadecimal digit. graphql-js ORs the digits
// shifted into place, where a -1 makes the whole negative; without the bitwise operators (gitignore's
// GAPS.md, gap 2) that is a test for a -1, then arithmetic.
function read16BitHexCode(body: string, position: number): number {
	const first = readHexDigit(body.charCodeAt(position));
	const second = readHexDigit(body.charCodeAt(position + 1));
	const third = readHexDigit(body.charCodeAt(position + 2));
	const fourth = readHexDigit(body.charCodeAt(position + 3));
	if (first < 0 || second < 0 || third < 0 || fourth < 0) {
		return -1;
	}
	return first * 0x1000 + second * 0x100 + third * 0x10 + fourth;
}

// EscapeSequence is the string value and lexed size of an escape sequence; the value is in the output
// form.
interface EscapeSequence {
	readonly value: string;
	readonly size: number;
}

function readEscapedUnicodeVariableWidth(lexer: Lexer, position: number): EscapeSequence {
	const body = lexer.body;
	let point = 0;
	let size = 3;
	// Cannot be larger than 12 chars (\u{00000000}).
	while (size < 12) {
		const code = body.charCodeAt(position + size);
		size++;
		// Closing Brace (})
		if (code === 0x007d) {
			// Must be at least 5 chars (\u{0}) and encode a Unicode scalar value.
			if (size < 5 || !isUnicodeScalarValue(point)) {
				break;
			}
			return { value: written(point), size };
		}
		// Append this hex digit to the code point. graphql-js's point is a JavaScript bitwise result, a
		// signed 32-bit integer, `(point << 4) | readHexDigit(code)`: negative for a digit that isn't one,
		// and negative when the shift reaches the sign bit, which the loop's eight digits at most can do
		// only from below it (gap 2).
		const digit = readHexDigit(code);
		point = digit < 0 || point * 16 + digit > 0x7fffffff ? -1 : point * 16 + digit;
		if (point < 0) {
			break;
		}
	}

	throw new Error(syntaxError(body, position, `Invalid Unicode escape sequence: "${escaped(body.slice(position, position + size))}".`));
}

function readEscapedUnicodeFixedWidth(lexer: Lexer, position: number): EscapeSequence {
	const body = lexer.body;
	const code = read16BitHexCode(body, position + 2);

	if (isUnicodeScalarValue(code)) {
		return { value: written(code), size: 6 };
	}

	// GraphQL allows JSON-style surrogate pair escape sequences, but only when a valid pair is formed.
	if (isLeadingSurrogate(code)) {
		// \u
		if (body.charCodeAt(position + 6) === 0x005c && body.charCodeAt(position + 7) === 0x0075) {
			const trailingCode = read16BitHexCode(body, position + 8);
			if (isTrailingSurrogate(trailingCode)) {
				// JavaScript defines strings as a sequence of UTF-16 code units and encodes Unicode code
				// points above U+FFFF using a surrogate pair of code units. Since this is a surrogate pair
				// escape sequence, graphql-js includes both codes in its string value; written out, they are
				// the one code point they spell.
				return { value: written((code - 0xd800) * 0x400 + (trailingCode - 0xdc00) + 0x10000), size: 12 };
			}
		}
	}

	throw new Error(syntaxError(body, position, `Invalid Unicode escape sequence: "${escaped(body.slice(position, position + 6))}".`));
}

// readEscapedCharacter:
//
// | Escaped Character | Code Point | Character Name               |
// | ----------------- | ---------- | ---------------------------- |
// | `"`               | U+0022     | double quote                 |
// | `\`               | U+005C     | reverse solidus (back slash) |
// | `/`               | U+002F     | solidus (forward slash)      |
// | `b`               | U+0008     | backspace                    |
// | `f`               | U+000C     | form feed                    |
// | `n`               | U+000A     | line feed (new line)         |
// | `r`               | U+000D     | carriage return              |
// | `t`               | U+0009     | horizontal tab               |
function readEscapedCharacter(lexer: Lexer, position: number): EscapeSequence {
	const body = lexer.body;
	const code = body.charCodeAt(position + 1);
	switch (code) {
		case 0x0022: // "
			return { value: written(0x0022), size: 2 };
		case 0x005c: // \
			return { value: written(0x005c), size: 2 };
		case 0x002f: // /
			return { value: written(0x002f), size: 2 };
		case 0x0062: // b
			return { value: written(0x0008), size: 2 };
		case 0x0066: // f
			return { value: written(0x000c), size: 2 };
		case 0x006e: // n
			return { value: written(0x000a), size: 2 };
		case 0x0072: // r
			return { value: written(0x000d), size: 2 };
		case 0x0074: // t
			return { value: written(0x0009), size: 2 };
	}
	throw new Error(syntaxError(body, position, `Invalid character escape sequence: "${escaped(body.slice(position, position + 2))}".`));
}

// createToken creates a token with line and column location information.
function createToken(lexer: Lexer, kind: TokenKind, start: number, end: number, value: string | undefined): Token {
	const line = lexer.line;
	const column = 1 + start - lexer.lineStart;
	return new Token(kind, start, end, line, column, value);
}

// readComment reads a comment token from the source file.
//
//	Comment :: # CommentChar* [lookahead != CommentChar]
//
//	CommentChar :: SourceCharacter but not LineTerminator
function readComment(lexer: Lexer, start: number): Token {
	const body = lexer.body;
	const bodyLength = body.length;
	let position = start + 1;

	while (position < bodyLength) {
		const code = body.charCodeAt(position);

		// LineTerminator (\n | \r)
		if (code === 0x000a || code === 0x000d) {
			break;
		}

		// SourceCharacter
		if (isUnicodeScalarValue(code)) {
			position++;
		} else if (isSupplementaryCodePoint(body, position)) {
			position += 2;
		} else {
			break;
		}
	}

	return createToken(lexer, 'Comment', start, position, escaped(body.slice(start + 1, position)));
}

// readDigits returns the new position in the source after reading one or more digits.
function readDigits(lexer: Lexer, start: number, firstCode: number): number {
	const body = lexer.body;
	if (!isDigit(firstCode)) {
		throw new Error(syntaxError(body, start, `Invalid number, expected digit but got: ${printCodePointAt(lexer, start)}.`));
	}

	let position = start + 1; // +1 to skip first firstCode

	while (isDigit(body.charCodeAt(position))) {
		position++;
	}

	return position;
}

// readNumber reads a number token from the source file, either a FloatValue or an IntValue depending on
// whether a FractionalPart or ExponentPart is encountered.
//
//	IntValue :: IntegerPart [lookahead != {Digit, `.`, NameStart}]
//
//	IntegerPart ::
//	  - NegativeSign? 0
//	  - NegativeSign? NonZeroDigit Digit*
//
//	NegativeSign :: -
//
//	NonZeroDigit :: Digit but not `0`
//
//	FloatValue ::
//	  - IntegerPart FractionalPart ExponentPart [lookahead != {Digit, `.`, NameStart}]
//	  - IntegerPart FractionalPart [lookahead != {Digit, `.`, NameStart}]
//	  - IntegerPart ExponentPart [lookahead != {Digit, `.`, NameStart}]
//
//	FractionalPart :: . Digit+
//
//	ExponentPart :: ExponentIndicator Sign? Digit+
//
//	ExponentIndicator :: one of `e` `E`
//
//	Sign :: one of + -
function readNumber(lexer: Lexer, start: number, firstCode: number): Token {
	const body = lexer.body;
	let position = start;
	let code = firstCode;
	let isFloat = false;

	// NegativeSign (-)
	if (code === 0x002d) {
		position++;
		code = body.charCodeAt(position);
	}

	// Zero (0)
	if (code === 0x0030) {
		position++;
		code = body.charCodeAt(position);
		if (isDigit(code)) {
			throw new Error(syntaxError(body, position, `Invalid number, unexpected digit after 0: ${printCodePointAt(lexer, position)}.`));
		}
	} else {
		position = readDigits(lexer, position, code);
		code = body.charCodeAt(position);
	}

	// Full stop (.)
	if (code === 0x002e) {
		isFloat = true;

		position++;
		code = body.charCodeAt(position);
		position = readDigits(lexer, position, code);
		code = body.charCodeAt(position);
	}

	// E e
	if (code === 0x0045 || code === 0x0065) {
		isFloat = true;

		position++;
		code = body.charCodeAt(position);
		// + -
		if (code === 0x002b || code === 0x002d) {
			position++;
			code = body.charCodeAt(position);
		}
		position = readDigits(lexer, position, code);
		code = body.charCodeAt(position);
	}

	// Numbers cannot be followed by . or NameStart
	if (code === 0x002e || isNameStart(code)) {
		throw new Error(syntaxError(body, position, `Invalid number, expected digit but got: ${printCodePointAt(lexer, position)}.`));
	}

	return createToken(lexer, isFloat ? 'Float' : 'Int', start, position, body.slice(start, position));
}

// readString reads a single-quote string token from the source file.
//
//	StringValue ::
//	  - `""` [lookahead != `"`]
//	  - `"` StringCharacter+ `"`
//
//	StringCharacter ::
//	  - SourceCharacter but not `"` or `\` or LineTerminator
//	  - `\u` EscapedUnicode
//	  - `\` EscapedCharacter
//
//	EscapedUnicode ::
//	  - `{` HexDigit+ `}`
//	  - HexDigit HexDigit HexDigit HexDigit
//
//	EscapedCharacter :: one of `"` `\` `/` `b` `f` `n` `r` `t`
function readString(lexer: Lexer, start: number): Token {
	const body = lexer.body;
	const bodyLength = body.length;
	let position = start + 1;
	let chunkStart = position;
	let value = '';

	while (position < bodyLength) {
		const code = body.charCodeAt(position);

		// Closing Quote (")
		if (code === 0x0022) {
			value += escaped(body.slice(chunkStart, position));
			return createToken(lexer, 'String', start, position + 1, value);
		}

		// Escape Sequence (\)
		if (code === 0x005c) {
			value += escaped(body.slice(chunkStart, position));
			const escape =
				body.charCodeAt(position + 1) === 0x0075 // u
					? body.charCodeAt(position + 2) === 0x007b // {
						? readEscapedUnicodeVariableWidth(lexer, position)
						: readEscapedUnicodeFixedWidth(lexer, position)
					: readEscapedCharacter(lexer, position);
			value += escape.value;
			position += escape.size;
			chunkStart = position;
			continue;
		}

		// LineTerminator (\n | \r)
		if (code === 0x000a || code === 0x000d) {
			break;
		}

		// SourceCharacter
		if (isUnicodeScalarValue(code)) {
			position++;
		} else if (isSupplementaryCodePoint(body, position)) {
			position += 2;
		} else {
			throw new Error(syntaxError(body, position, `Invalid character within String: ${printCodePointAt(lexer, position)}.`));
		}
	}

	throw new Error(syntaxError(body, position, 'Unterminated string.'));
}

// readBlockString reads a block string token from the source file.
//
//	StringValue ::
//	  - `"""` BlockStringCharacter* `"""`
//
//	BlockStringCharacter ::
//	  - SourceCharacter but not `"""` or `\"""`
//	  - `\"""`
function readBlockString(lexer: Lexer, start: number): Token {
	const body = lexer.body;
	const bodyLength = body.length;
	let lineStart = lexer.lineStart;

	let position = start + 3;
	let chunkStart = position;
	let currentLine = '';

	const blockLines: string[] = [];
	while (position < bodyLength) {
		const code = body.charCodeAt(position);

		// Closing Triple-Quote (""")
		if (code === 0x0022 && body.charCodeAt(position + 1) === 0x0022 && body.charCodeAt(position + 2) === 0x0022) {
			currentLine += body.slice(chunkStart, position);
			blockLines.push(currentLine);

			const token = createToken(
				lexer,
				'BlockString',
				start,
				position + 3,
				// Return a string of the lines joined with U+000A.
				escaped(dedentBlockStringLines(blockLines).join('\n')),
			);

			lexer.line += blockLines.length - 1;
			lexer.lineStart = lineStart;
			return token;
		}

		// Escaped Triple-Quote (\""")
		if (code === 0x005c && body.charCodeAt(position + 1) === 0x0022 && body.charCodeAt(position + 2) === 0x0022 && body.charCodeAt(position + 3) === 0x0022) {
			currentLine += body.slice(chunkStart, position);
			chunkStart = position + 1; // skip only slash
			position += 4;
			continue;
		}

		// LineTerminator
		if (code === 0x000a || code === 0x000d) {
			currentLine += body.slice(chunkStart, position);
			blockLines.push(currentLine);

			if (code === 0x000d && body.charCodeAt(position + 1) === 0x000a) {
				position += 2;
			} else {
				position++;
			}

			currentLine = '';
			chunkStart = position;
			lineStart = position;
			continue;
		}

		// SourceCharacter
		if (isUnicodeScalarValue(code)) {
			position++;
		} else if (isSupplementaryCodePoint(body, position)) {
			position += 2;
		} else {
			throw new Error(syntaxError(body, position, `Invalid character within String: ${printCodePointAt(lexer, position)}.`));
		}
	}

	throw new Error(syntaxError(body, position, 'Unterminated string.'));
}

// readName reads an alphanumeric + underscore name from the source.
//
//	Name ::
//	  - NameStart NameContinue* [lookahead != NameContinue]
function readName(lexer: Lexer, start: number): Token {
	const body = lexer.body;
	const bodyLength = body.length;
	let position = start + 1;

	while (position < bodyLength) {
		const code = body.charCodeAt(position);
		if (isNameContinue(code)) {
			position++;
		} else {
			break;
		}
	}

	return createToken(lexer, 'Name', start, position, body.slice(start, position));
}

// readNextToken gets the next token from the source starting at the given position.
//
// This skips over whitespace until it finds the next lexable token, then lexes punctuators immediately
// or calls the appropriate helper function for more complicated tokens.
function readNextToken(lexer: Lexer, start: number): Token {
	const body = lexer.body;
	const bodyLength = body.length;
	let position = start;

	while (position < bodyLength) {
		const code = body.charCodeAt(position);

		// SourceCharacter
		switch (code) {
			// Ignored ::
			//   - UnicodeBOM
			//   - WhiteSpace
			//   - LineTerminator
			//   - Comment
			//   - Comma
			//
			// UnicodeBOM :: "Byte Order Mark (U+FEFF)"
			//
			// WhiteSpace ::
			//   - "Horizontal Tab (U+0009)"
			//   - "Space (U+0020)"
			//
			// Comma :: ,
			case 0xfeff: // <BOM>
			case 0x0009: // \t
			case 0x0020: // <space>
			case 0x002c: // ,
				position++;
				continue;
			// LineTerminator ::
			//   - "New Line (U+000A)"
			//   - "Carriage Return (U+000D)" [lookahead != "New Line (U+000A)"]
			//   - "Carriage Return (U+000D)" "New Line (U+000A)"
			case 0x000a: // \n
				position++;
				lexer.line++;
				lexer.lineStart = position;
				continue;
			case 0x000d: // \r
				if (body.charCodeAt(position + 1) === 0x000a) {
					position += 2;
				} else {
					position++;
				}
				lexer.line++;
				lexer.lineStart = position;
				continue;
			// Comment
			case 0x0023: // #
				return readComment(lexer, position);
			// Token ::
			//   - Punctuator
			//   - Name
			//   - IntValue
			//   - FloatValue
			//   - StringValue
			//
			// Punctuator :: one of ! $ & ( ) ... : = @ [ ] { | }
			case 0x0021: // !
				return createToken(lexer, '!', position, position + 1, undefined);
			case 0x0024: // $
				return createToken(lexer, '$', position, position + 1, undefined);
			case 0x0026: // &
				return createToken(lexer, '&', position, position + 1, undefined);
			case 0x0028: // (
				return createToken(lexer, '(', position, position + 1, undefined);
			case 0x0029: // )
				return createToken(lexer, ')', position, position + 1, undefined);
			case 0x002e: {
				// .
				const nextCode = body.charCodeAt(position + 1);
				if (nextCode === 0x002e && body.charCodeAt(position + 2) === 0x002e) {
					return createToken(lexer, '...', position, position + 3, undefined);
				}
				if (nextCode === 0x002e) {
					throw new Error(syntaxError(body, position, 'Unexpected "..", did you mean "..."?'));
				} else if (isDigit(nextCode)) {
					const digits = body.slice(position + 1, readDigits(lexer, position + 1, nextCode));
					throw new Error(syntaxError(body, position, `Invalid number, expected digit before ".", did you mean "0.${digits}"?`));
				}
				break;
			}
			case 0x003a: // :
				return createToken(lexer, ':', position, position + 1, undefined);
			case 0x003d: // =
				return createToken(lexer, '=', position, position + 1, undefined);
			case 0x0040: // @
				return createToken(lexer, '@', position, position + 1, undefined);
			case 0x005b: // [
				return createToken(lexer, '[', position, position + 1, undefined);
			case 0x005d: // ]
				return createToken(lexer, ']', position, position + 1, undefined);
			case 0x007b: // {
				return createToken(lexer, '{', position, position + 1, undefined);
			case 0x007c: // |
				return createToken(lexer, '|', position, position + 1, undefined);
			case 0x007d: // }
				return createToken(lexer, '}', position, position + 1, undefined);
			// StringValue
			case 0x0022: // "
				if (body.charCodeAt(position + 1) === 0x0022 && body.charCodeAt(position + 2) === 0x0022) {
					return readBlockString(lexer, position);
				}
				return readString(lexer, position);
		}

		// IntValue | FloatValue (Digit | -)
		if (isDigit(code) || code === 0x002d) {
			return readNumber(lexer, position, code);
		}

		// Name
		if (isNameStart(code)) {
			return readName(lexer, position);
		}

		throw new Error(
			syntaxError(
				body,
				position,
				code === 0x0027
					? 'Unexpected single quote character (\'), did you mean to use a double quote (")?'
					: isUnicodeScalarValue(code) || isSupplementaryCodePoint(body, position)
						? `Unexpected character: ${printCodePointAt(lexer, position)}.`
						: `Invalid character: ${printCodePointAt(lexer, position)}.`,
			),
		);
	}

	return createToken(lexer, '<EOF>', bodyLength, bodyLength, undefined);
}

// Lexer is given a Source object and creates a Lexer for that source. A Lexer is a stateful stream
// generator in that every time it is advanced, it returns the next token in the Source. Assuming the
// source lexes, the final Token emitted by the lexer will be of kind EOF, after which the lexer will
// repeatedly return the same EOF token whenever called.
export class Lexer {
	// Source document used to derive error locations.
	readonly body: string;

	// Most recent non-ignored token returned by the lexer.
	lastToken: Token;

	// Current non-ignored token at the lexer cursor.
	token: Token;

	// The (1-indexed) line containing the current token.
	line = 1;

	// Character offset where the current line starts.
	lineStart = 0;

	// Every token read, ignored ones included, in order: graphql-js's linked list of tokens, from <SOF>
	// (token.ts says why it is an array), with the places in it of token and lastToken.
	readonly tokens: Token[] = [];
	tokenIndex = 0;
	lastTokenIndex = 0;

	constructor(body: string) {
		const startOfFileToken = new Token('<SOF>', 0, 0, 0, 0, undefined);
		this.body = body;
		this.lastToken = startOfFileToken;
		this.token = startOfFileToken;
		this.tokens.push(startOfFileToken);
	}

	// lookaheadIndex is where in tokens the next non-ignored token is, read and added to the list when
	// it hasn't been.
	lookaheadIndex(): number {
		let index = this.tokenIndex;
		let token = this.token;
		if (token.kind !== '<EOF>') {
			do {
				index++;
				const next = this.tokens[index];
				if (next !== undefined) {
					token = next;
				} else {
					// Read the next token and form a link in the token linked-list.
					token = readNextToken(this, token.end);
					this.tokens.push(token);
				}
				// Comments are ignored.
			} while (token.kind === 'Comment');
		}
		return index;
	}

	// advance advances the token stream to the next non-ignored token.
	advance(): Token {
		this.lastToken = this.token;
		this.lastTokenIndex = this.tokenIndex;
		this.tokenIndex = this.lookaheadIndex();
		this.token = this.tokens[this.tokenIndex] ?? panic(`no token ${this.tokenIndex}`);
		return this.token;
	}

	// lookahead looks ahead and returns the next non-ignored token, but does not change the state of
	// Lexer.
	lookahead(): Token {
		return this.tokens[this.lookaheadIndex()] ?? panic('no token ahead');
	}
}
