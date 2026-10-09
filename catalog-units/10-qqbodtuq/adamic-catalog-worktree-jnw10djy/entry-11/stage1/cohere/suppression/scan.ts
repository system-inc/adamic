// A port of cohere's internal/lint/suppression/scan.go to Adamic 0.1: where the comments are in a
// file, as TypeScript lexes it. Each piece names the Go it reads as.
//
// The Go walks bytes; the port walks UTF-16 units, and its offsets are UTF-16 indexes where the Go's are
// byte offsets. Every test the scanner makes is on an ASCII character, and a multi-byte character is
// one or two units here and two to four bytes there, none of them ASCII, so both take the same
// decisions; the test converts the Go's offsets to UTF-16 indexes before comparing.

import { parseEnable } from './directives.ts';

// scan.go: commentSpan, one comment's range in a file.
export interface CommentSpan {
	readonly pos: number;
	readonly end: number;
}

// scan.go: isSpace.
function isSpace(character: string): boolean {
	return character === ' ' || character === '\t' || character === '\n' || character === '\r';
}

// scan.go: opensRegularExpression decides whether a `/` starts a pattern or divides. After a value, a
// slash divides; after an operator, a comma, or an opening bracket, it opens a pattern. previousToken
// is '' at the start, the Go's 0.
function opensRegularExpression(previousToken: string): boolean {
	switch (previousToken) {
		case '':
		case '(':
		case ',':
		case '=':
		case ':':
		case '[':
		case '!':
		case '&':
		case '|':
		case '?':
		case '{':
		case '}':
		case ';':
		case '+':
		case '-':
		case '*':
		case '~':
		case '^':
		case '<':
		case '>':
		case '%':
			return true;
	}
	return false;
}

// scan.go: skipQuoted advances past a single- or double-quoted string, honoring backslash escapes. An
// unterminated one ends at its line's end, so the scan still terminates.
function skipQuoted(text: string, start: number, quote: string): number {
	let index = start + 1;
	while (index < text.length) {
		const character = text[index];
		if (character === '\\') {
			index += 2;
			continue;
		}
		if (character === quote) {
			return index + 1;
		}
		if (character === '\n') {
			return index + 1;
		}
		index++;
	}
	return index;
}

// scan.go: skipRegularExpression advances past a regex literal, honoring escapes and character classes.
// A `/` inside `[...]` does not end the pattern. A pattern that reaches its line's end was a division
// after all, and the scan backs out there.
function skipRegularExpression(text: string, start: number): number {
	let index = start + 1;
	let inClass = false;
	while (index < text.length) {
		switch (text[index]) {
			case '\\':
				index += 2;
				continue;
			case '[':
				inClass = true;
				break;
			case ']':
				inClass = false;
				break;
			case '/':
				if (!inClass) {
					return index + 1;
				}
				break;
			case '\n':
				return index;
		}
		index++;
	}
	return index;
}

// scan.go: scanCode lexes code from index, appending every comment it passes, and returns where it
// stopped: the end of the text, or, inside an interpolation, just past the `}` that closes the `${`.
function scanCode(text: string, start: number, inInterpolation: boolean, comments: CommentSpan[]): number {
	let index = start;
	// previousToken is the last significant character seen, the only way to tell a regular expression
	// from a division. An interpolation starts an expression, so it starts as the file does.
	let previousToken = '';

	// depth counts the braces opened inside an interpolation, so `${ {a: 1}.a }` closes on its own
	// brace rather than the object literal's.
	let depth = 0;

	while (index < text.length) {
		const character = text[index] ?? '';

		if (character === '/' && text[index + 1] === '/') {
			const commentStart = index;
			index += 2;
			while (index < text.length && text[index] !== '\n') {
				index++;
			}
			comments.push({ pos: commentStart, end: index });
		} else if (character === '/' && text[index + 1] === '*') {
			const commentStart = index;
			index += 2;
			while (index + 1 < text.length && !(text[index] === '*' && text[index + 1] === '/')) {
				index++;
			}
			if (index + 1 < text.length) {
				index += 2;
			} else {
				index = text.length;
			}
			comments.push({ pos: commentStart, end: index });
		} else if (character === '"' || character === "'") {
			index = skipQuoted(text, index, character);
			previousToken = character;
		} else if (character === '`') {
			// scan.go: skipTemplate advances past a template literal, lexing each `${...}` as code,
			// because a comment, a string or a regular expression can live inside an interpolation. It
			// is written here rather than as its own function: it and scanCode call each other, and
			// stage 0 lowers no pair of functions that do (gap 1 in GAPS.md).
			index++;
			while (index < text.length) {
				const inTemplate = text[index];
				if (inTemplate === '\\') {
					index += 2;
					continue;
				}
				if (inTemplate === '`') {
					index++;
					break;
				}
				if (inTemplate === '$' && text[index + 1] === '{') {
					index = scanCode(text, index + 2, true, comments);
					continue;
				}
				index++;
			}
			previousToken = character;
		} else if (character === '/' && opensRegularExpression(previousToken)) {
			index = skipRegularExpression(text, index);
			previousToken = '/';
		} else if (character === '}' && inInterpolation && depth === 0) {
			return index + 1;
		} else {
			if (character === '{') {
				depth++;
			} else if (character === '}') {
				depth--;
			}
			if (!isSpace(character)) {
				previousToken = character;
			}
			index++;
		}
	}

	return index;
}

// scan.go: scanComments finds every comment in a file, skipping string literals, template literals and
// regular expressions rather than searching them, so source that writes a directive into a string
// isn't read as one.
export function scanComments(text: string): CommentSpan[] {
	const comments: CommentSpan[] = [];
	scanCode(text, 0, false, comments);
	return comments;
}

// scan.go: enableSpan, one enable directive, resolved to the line it sits on and the rules it names.
export interface EnableSpan {
	readonly line: number;
	readonly rules: readonly string[];
}

// scan.go: scanEnables finds the comments that close a block suppression.
export function scanEnables(text: string, lineOf: (offset: number) => number): EnableSpan[] {
	const found: EnableSpan[] = [];

	for (const comment of scanComments(text)) {
		const enable = parseEnable(text.slice(comment.pos, comment.end));
		if (enable === undefined) {
			continue;
		}
		found.push({ line: lineOf(comment.pos), rules: enable.rules });
	}

	return found;
}

// scan.go: buildLineIndex returns a function from offset to zero-based line, built once per file rather
// than counting newlines per lookup.
export function buildLineIndex(text: string): (offset: number) => number {
	const starts: number[] = [0];
	for (let index = 0; index < text.length; index++) {
		if (text[index] === '\n') {
			starts.push(index + 1);
		}
	}
	return (offset: number): number => {
		let low = 0;
		let high = starts.length - 1;
		while (low < high) {
			const middle = Math.floor((low + high + 1) / 2);
			if ((starts[middle] ?? 0) <= offset) {
				low = middle;
			} else {
				high = middle - 1;
			}
		}
		return low;
	};
}
