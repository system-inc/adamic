// A port of cohere's internal/format/css/values/tokenize.go to Adamic 0.1: postcss-values-parser
// 2.0.1, lib/tokenize.js. Each piece names the Go it reads as.
//
// The Go walks UTF-16 code units held in a []uint16, as upstream's charCodeAt and slice do. The port's
// strings are UTF-16 already, so it walks them as upstream does, and the Go's stand-ins for JavaScript
// go back to being JavaScript:
//
//   - The Go's undefinedNext (math.MinInt, with an endColumnIsNaN flag) is upstream's `next` before
//     anything assigns it. Here `next` starts as NaN, which is what `undefined - offset` is, so a
//     bracket or paren token's end column comes out NaN, or stale, as upstream's does.
//   - The Go's charCodeAt returns -1 past either end, standing in for NaN; here it is NaN.
//   - A throw is a Result, since 0.1 has no exceptions. Its message is upstream's String(error).
//   - indexOf from a position is written with slice, since stage 0 doesn't lower indexOf's second
//     argument yet (gap 1 in GAPS.md).

// The token kinds are upstream's token[0].
export type TokenKind =
	| 'space'
	| 'colon'
	| 'comma'
	| '{'
	| '}'
	| '('
	| ')'
	| 'string'
	| 'atword'
	| 'word'
	| 'operator'
	| 'comment'
	| '#'
	| 'unicoderange';

// tokenize.go: token, upstream's token array: [type, value, startLine, startColumn, endLine, endColumn,
// index]. endColumn may be NaN.
export interface Token {
	readonly kind: TokenKind;
	readonly value: string;
	readonly startLine: number;
	readonly startColumn: number;
	readonly endLine: number;
	readonly endColumn: number;
	readonly index: number;
}

// What tokenize gives: the tokens, or the TokenizeError's String(error).
export type Tokenized = { readonly kind: 'Ok'; readonly tokens: Token[] } | { readonly kind: 'Error'; readonly message: string };

const openBracket = 0x7b; // {
const closeBracket = 0x7d; // }
const openParen = 0x28; // (
const closeParen = 0x29; // )
const singleQuote = 0x27; // '
const doubleQuote = 0x22; // "
const backslash = 0x5c; // \
const slash = 0x2f; // /
const period = 0x2e; // .
const comma = 0x2c; // ,
const colon = 0x3a; // :
const asterisk = 0x2a; // *
const minus = 0x2d; // -
const plus = 0x2b; // +
const pound = 0x23; // #
const newline = 0x0a;
const space = 0x20;
const feed = 0x0c;
const tab = 0x09;
const cr = 0x0d;
const at = 0x40; // @
const lowerE = 0x65; // e
const upperE = 0x45; // E
const digit0 = 0x30;
const digit9 = 0x39;
const lowerU = 0x75; // u
const upperU = 0x55; // U

// tokenize.go: atEnd, upstream's /[ \n\t\r\{\(\)'"\\;,/]/g at one index.
function atEnd(css: string, index: number): boolean {
	switch (css.charCodeAt(index)) {
		case 0x20:
		case 0x0a:
		case 0x09:
		case 0x0d:
		case 0x7b:
		case 0x28:
		case 0x29:
		case 0x27:
		case 0x22:
		case 0x5c:
		case 0x3b:
		case 0x2c:
		case 0x2f:
			return true;
	}
	return false;
}

// tokenize.go: wordEnd, upstream's /[ \n\t\r\(\)\{\}\*:;@!&'"\+\|~>,\[\]\\]|\/(?=\*)/g at one index.
function wordEnd(css: string, index: number): boolean {
	switch (css.charCodeAt(index)) {
		case 0x20:
		case 0x0a:
		case 0x09:
		case 0x0d:
		case 0x28:
		case 0x29:
		case 0x7b:
		case 0x7d:
		case 0x2a:
		case 0x3a:
		case 0x3b:
		case 0x40:
		case 0x21:
		case 0x26:
		case 0x27:
		case 0x22:
		case 0x2b:
		case 0x7c:
		case 0x7e:
		case 0x3e:
		case 0x2c:
		case 0x5b:
		case 0x5d:
		case 0x5c:
			return true;
		case 0x2f:
			return css.charCodeAt(index + 1) === asterisk;
	}
	return false;
}

// tokenize.go: wordEndNum, upstream's /[ \n\t\r\(\)\{\}\*:;@!&'"\-\+\|~>,\[\]\\]|\//g at one index.
function wordEndNum(css: string, index: number): boolean {
	switch (css.charCodeAt(index)) {
		case 0x20:
		case 0x0a:
		case 0x09:
		case 0x0d:
		case 0x28:
		case 0x29:
		case 0x7b:
		case 0x7d:
		case 0x2a:
		case 0x3a:
		case 0x3b:
		case 0x40:
		case 0x21:
		case 0x26:
		case 0x27:
		case 0x22:
		case 0x2d:
		case 0x2b:
		case 0x7c:
		case 0x7e:
		case 0x3e:
		case 0x2c:
		case 0x5b:
		case 0x5d:
		case 0x5c:
		case 0x2f:
			return true;
	}
	return false;
}

// tokenize.go: alphaNum, upstream's /^[a-z0-9]/i. JavaScript's case-insensitive class without the u
// flag folds only within ASCII here, so this is a plain ASCII test. NaN is none of them.
export function alphaNum(code: number): boolean {
	return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39);
}

// tokenize.go: unicodeRange, upstream's /^[a-f0-9?\-]/i.
function unicodeRange(code: number): boolean {
	return (code >= 0x61 && code <= 0x66) || (code >= 0x41 && code <= 0x46) || (code >= 0x30 && code <= 0x39) || code === 0x3f || code === minus;
}

// tokenize.go: regexLastIndex, `regex.lastIndex = from; regex.test(css); regex.lastIndex` for the
// one-character global regexes above: one past the match, or 0 when there is none. The regex is a
// function, as the Go passes it.
function regexLastIndex(css: string, from: number, matches: (css: string, index: number) => boolean): number {
	for (let index = from; index < css.length; index++) {
		if (matches(css, index)) {
			return index + 1;
		}
	}
	return 0;
}

// tokenize.go: indexOfUnits, css.indexOf(search, from), for a search that isn't empty. Written with
// slice, since stage 0 doesn't lower indexOf's second argument yet (gap 1 in GAPS.md).
function indexOfFrom(css: string, search: string, from: number): number {
	const start = Math.max(from, 0);
	const found = css.slice(start).indexOf(search);
	return found === -1 ? -1 : found + start;
}

// tokenize.go: tokenize.
export function tokenize(input: string, loose: boolean): Tokenized {
	const tokens: Token[] = [];
	const css = input;
	const length = css.length;
	let offset = -1;
	let line = 1;
	let pos = 0;
	let parentCount = 0;
	let isURLArg = false;

	let code = 0;
	// upstream's `next` is undefined until assigned, and undefined - offset is NaN.
	let next = NaN;
	let quote = '';
	let escaped = false;
	let escapePos = 0;

	const push = (kind: TokenKind, value: string, startLine: number, startColumn: number, endLine: number, endColumn: number, index: number): void => {
		tokens.push({ kind, value, startLine, startColumn, endLine, endColumn, index });
	};

	const unclosed = (what: string): Tokenized => ({
		kind: 'Error',
		message: `TokenizeError: Unclosed ${what} at line: ${line}, column: ${pos - offset}, token: ${pos}`,
	});

	while (pos < length) {
		code = css.charCodeAt(pos);

		if (code === newline) {
			offset = pos;
			line += 1;
		}

		// Each case is written as its code, with the constant's name beside it: stage 0 doesn't lower a case
		// that names a constant yet (gap 3 in GAPS.md).
		switch (code) {
			case 0x0a: // newline
			case 0x20: // space
			case 0x09: // tab
			case 0x0d: // cr
			case 0x0c: // feed
				next = pos;
				do {
					next += 1;
					code = css.charCodeAt(next);
					if (code === newline) {
						offset = next;
						line += 1;
					}
				} while (code === space || code === newline || code === tab || code === cr || code === feed);

				push('space', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

				pos = next - 1;
				break;

			case 0x3a: // colon
				next = pos + 1;
				push('colon', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

				pos = next - 1;
				break;

			case 0x2c: // comma
				next = pos + 1;
				push('comma', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

				pos = next - 1;
				break;

			case 0x7b: // openBracket
				push('{', '{', line, pos - offset, line, next - offset, pos);
				break;

			case 0x7d: // closeBracket
				push('}', '}', line, pos - offset, line, next - offset, pos);
				break;

			case 0x28: { // openParen
				parentCount++;
				const previous = tokens.at(-1);
				isURLArg = !isURLArg && parentCount === 1 && previous !== undefined && previous.kind === 'word' && previous.value === 'url';
				push('(', '(', line, pos - offset, line, next - offset, pos);
				break;
			}

			case 0x29: // closeParen
				parentCount--;
				isURLArg = isURLArg && parentCount > 0;
				push(')', ')', line, pos - offset, line, next - offset, pos);
				break;

			case 0x27: // singleQuote
			case 0x22: // doubleQuote
				quote = code === singleQuote ? "'" : '"';
				next = pos;
				do {
					escaped = false;
					next = indexOfFrom(css, quote, next + 1);
					if (next === -1) {
						return unclosed('quote');
					}
					escapePos = next;
					while (css.charCodeAt(escapePos - 1) === backslash) {
						escapePos -= 1;
						escaped = !escaped;
					}
				} while (escaped);

				push('string', css.slice(pos, next + 1), line, pos - offset, line, next - offset, pos);
				pos = next;
				break;

			case 0x40: { // at
				const lastIndex = regexLastIndex(css, pos + 1, atEnd);

				if (lastIndex === 0) {
					next = length - 1;
				} else {
					next = lastIndex - 2;
				}

				push('atword', css.slice(pos, next + 1), line, pos - offset, line, next - offset, pos);
				pos = next;
				break;
			}

			case 0x5c: // backslash
				next = pos;
				code = css.charCodeAt(next + 1);

				// Upstream tests an `escape` variable here that nothing assigns, so the backslash is always
				// a one-character word and the branch that would take the escaped character is dead.

				push('word', css.slice(pos, next + 1), line, pos - offset, line, next - offset, pos);

				pos = next;
				break;

			case 0x2b: // plus
			case 0x2d: // minus
			case 0x2a: { // asterisk
				next = pos + 1;
				const nextChar = css.charCodeAt(pos + 1);

				// Upstream also computes a prevChar here that it never reads.

				// if the operator is immediately followed by a word character, then we
				// have a prefix of some kind, and should fall-through. eg. -webkit

				// look for --* for custom variables
				if (code === minus && nextChar === minus) {
					next++;

					push('word', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

					pos = next - 1;
					break;
				}

				push('operator', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

				pos = next - 1;
				break;
			}

			default:
				if (code === slash && (css.charCodeAt(pos + 1) === asterisk || (loose && !isURLArg && css.charCodeAt(pos + 1) === slash))) {
					const isStandardComment = css.charCodeAt(pos + 1) === asterisk;

					if (isStandardComment) {
						next = indexOfFrom(css, '*/', pos + 2) + 1;
						if (next === 0) {
							return unclosed('comment');
						}
					} else {
						const newlinePos = indexOfFrom(css, '\n', pos + 2);

						if (newlinePos !== -1) {
							next = newlinePos - 1;
						} else {
							next = length;
						}
					}

					const content = css.slice(pos, next + 1);
					const lines = content.split('\n');
					const last = lines.length - 1;

					let nextLine = 0;
					let nextOffset = 0;
					if (last > 0) {
						nextLine = line + last;
						nextOffset = next - (lines[last] ?? '').length;
					} else {
						nextLine = line;
						nextOffset = offset;
					}

					push('comment', content, line, pos - offset, nextLine, next - nextOffset, pos);

					offset = nextOffset;
					line = nextLine;
					pos = next;
				} else if (code === pound && !alphaNum(css.charCodeAt(pos + 1))) {
					next = pos + 1;

					push('#', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

					pos = next - 1;
				} else if ((code === lowerU || code === upperU) && css.charCodeAt(pos + 1) === plus) {
					next = pos + 2;

					do {
						next += 1;
						code = css.charCodeAt(next);
					} while (next < length && unicodeRange(code));

					push('unicoderange', css.slice(pos, next), line, pos - offset, line, next - offset, pos);
					pos = next - 1;
				} else if (code === slash) {
					// catch a regular slash, that isn't a comment
					next = pos + 1;

					push('operator', css.slice(pos, next), line, pos - offset, line, next - offset, pos);

					pos = next - 1;
				} else {
					let regex = wordEnd;
					let isWordEndNum = false;

					// we're dealing with a word that starts with a number
					// those get treated differently
					if (code >= digit0 && code <= digit9) {
						regex = wordEndNum;
						isWordEndNum = true;
					}

					const lastIndex = regexLastIndex(css, pos + 1, regex);

					if (lastIndex === 0) {
						next = length - 1;
					} else {
						next = lastIndex - 2;
					}

					// Exponential number notation with minus or plus: 1e-10, 1e+10
					if (isWordEndNum || code === period) {
						const ncode = css.charCodeAt(next);
						const ncode1 = css.charCodeAt(next + 1);
						const ncode2 = css.charCodeAt(next + 2);

						if ((ncode === lowerE || ncode === upperE) && (ncode1 === minus || ncode1 === plus) && ncode2 >= digit0 && ncode2 <= digit9) {
							const exponentLastIndex = regexLastIndex(css, next + 2, wordEndNum);

							if (exponentLastIndex === 0) {
								next = length - 1;
							} else {
								next = exponentLastIndex - 2;
							}
						}
					}

					push('word', css.slice(pos, next + 1), line, pos - offset, line, next - offset, pos);
					pos = next;
				}
				break;
		}

		pos++;
	}

	return { kind: 'Ok', tokens };
}
