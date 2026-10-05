// A port of cohere's internal/gitignore/glob.go to Adamic 0.1. The Go is beside it in the cohere
// submodule (cohere/internal/gitignore/glob.go); each piece here names the Go it reads as.
//
// A glob is one pattern's wildcard part, compiled once when its ignore file is read. It is written from
// gitignore(5) and fnmatch(3)'s descriptions:
//
//   - `?` is any one byte, `*` any run of bytes, and `[...]` one byte from a set: ranges, `!` or `^` to
//     negate, the `[:name:]` classes, and `]` as a member when it comes first.
//   - A backslash makes the next byte literal, and a backslash ending the pattern makes it match nothing.
//   - Against a path (anchored patterns), none of those crosses a `/`, and `**` as a whole segment does:
//     `**/` at the start or after a `/` is zero or more directories, and a trailing `/**` everything
//     below. Any other run of asterisks is one `*`.
//   - A set that never closes, or a class name fnmatch does not define, makes the pattern match nothing.
//
// Bytes, not UTF-16 code units, as git and the Go both match bytes: `?` against `é` is false, because
// `é` is two bytes. A Go string is its bytes; an Adamic string is UTF-16, so the glob first spells its
// pattern and its text as byte strings (bytes.ts), one character per byte, and then reads exactly as
// the Go does. Matching walks the text once while tracking every place in the pattern it could have
// reached, so no pattern backtracks: `*a*a*a*a*b` against a long run of `a` costs one pass.
//
// The declarations come callee first, not in the Go's order: stage 0 takes a call to a function or method
// declared further down for a void one (gap 3 in GAPS.md). Each still names the Go it reads as.

import { panic } from 'adamic';
import { byteAt, utf8Bytes } from './bytes.ts';

// glob.go: [256]bool, a set of bytes.
type ByteSet = readonly boolean[];

// glob.go: tokenKind and token. The Go's one struct, whose set is nil except on a setToken, is a
// discriminated union here, so a set is there exactly when the kind says it is.
type Token =
	| { readonly kind: 'literalToken'; readonly literal: string } // one byte, itself
	| { readonly kind: 'anyByteToken' } // `?`
	| { readonly kind: 'setToken'; readonly set: ByteSet } // `[...]`
	| { readonly kind: 'starToken' } // `*`: any run, within one segment against a path
	| { readonly kind: 'everythingToken' } // a trailing `**` segment: any run, slashes included
	| { readonly kind: 'directoriesToken' }; // `**/` as a segment: nothing, or any run that ends in `/`

// glob.go: isLetter and isDigit, over a byte's value. The Go folds case with character|0x20; stage 0 does
// not lower the bitwise operators yet (gap 2 in GAPS.md), so here, as for xdigit above, each case is its
// own range.
function isLetter(character: number): boolean {
	return (character >= 0x41 && character <= 0x5a) || (character >= 0x61 && character <= 0x7a);
}

function isDigit(character: number): boolean {
	return character >= 0x30 && character <= 0x39;
}

// glob.go: addClass. It adds a `[:name:]` class's bytes, reporting false for a name fnmatch does not
// define. The Go's variable is named in, which is a keyword here.
function addClass(members: boolean[], name: string): boolean {
	let inClass: (character: number) => boolean;
	switch (name) {
		case 'alnum':
			inClass = (character) => isLetter(character) || isDigit(character);
			break;
		case 'alpha':
			// The Go assigns isLetter itself; stage 0 does not lower a declared function as a value yet
			// (gap 4 in GAPS.md), so here, and for digit, an arrow calls it.
			inClass = (character) => isLetter(character);
			break;
		case 'blank':
			inClass = (character) => character === 0x20 || character === 0x09;
			break;
		case 'cntrl':
			inClass = (character) => character < 0x20 || character === 0x7f;
			break;
		case 'digit':
			inClass = (character) => isDigit(character);
			break;
		case 'graph':
			inClass = (character) => character > 0x20 && character < 0x7f;
			break;
		case 'lower':
			inClass = (character) => character >= 0x61 && character <= 0x7a;
			break;
		case 'print':
			inClass = (character) => character >= 0x20 && character < 0x7f;
			break;
		case 'punct':
			inClass = (character) => character > 0x20 && character < 0x7f && !isLetter(character) && !isDigit(character);
			break;
		case 'space':
			inClass = (character) => character === 0x20 || character === 0x09 || character === 0x0a || character === 0x0d;
			break;
		case 'upper':
			inClass = (character) => character >= 0x41 && character <= 0x5a;
			break;
		case 'xdigit':
			inClass = (character) => isDigit(character) || (character >= 0x41 && character <= 0x46) || (character >= 0x61 && character <= 0x66);
			break;
		default:
			return false;
	}
	for (let member = 0; member < 256; member++) {
		if (inClass(member)) {
			members[member] = true;
		}
	}
	return true;
}

// glob.go: positionSet, a set over a glob's positions, and its add, has and clear. The Go's is a bit set
// of words; here it is an array of booleans.
function newPositionSet(size: number): boolean[] {
	return Array.from({ length: size }, () => false);
}

function addPosition(positions: boolean[], position: number): void {
	positions[position] = true;
}

function hasPosition(positions: readonly boolean[], position: number): boolean {
	return positions[position] === true;
}

function clearPositions(positions: boolean[]): void {
	positions.fill(false);
}

// glob.go: compileSet's results. The Go returns (set, end, ok); here undefined is not ok.
interface CompiledSet {
	readonly set: ByteSet;
	readonly end: number;
}

// glob.go: compileSet. It reads the bracket expression that opens at start, returning its members and
// the index of its closing `]`, or undefined for one that never closes or names an unknown class.
function compileSet(pattern: string, start: number): CompiledSet | undefined {
	const members = Array.from({ length: 256 }, () => false);
	let index = start + 1;
	let negated = false;
	if (index < pattern.length && (byteAt(pattern, index) === '!' || byteAt(pattern, index) === '^')) {
		negated = true;
		index++;
	}

	// rangeStart is the byte a following `-` would extend, or -1 after a class or a range, which
	// nothing extends.
	let rangeStart = -1;
	let first = true;
	for (;;) {
		if (index >= pattern.length) {
			return undefined;
		}
		const character = byteAt(pattern, index);
		if (character === ']' && !first) {
			break;
		}
		first = false;

		if (character === '\\') {
			index++;
			if (index >= pattern.length) {
				return undefined;
			}
			members[byteAt(pattern, index).charCodeAt(0)] = true;
			rangeStart = byteAt(pattern, index).charCodeAt(0);
		} else if (character === '-' && rangeStart >= 0 && index + 1 < pattern.length && byteAt(pattern, index + 1) !== ']') {
			index++;
			let upper = byteAt(pattern, index);
			if (upper === '\\') {
				index++;
				if (index >= pattern.length) {
					return undefined;
				}
				upper = byteAt(pattern, index);
			}
			for (let member = rangeStart; member <= upper.charCodeAt(0); member++) {
				members[member] = true;
			}
			rangeStart = -1;
		} else if (character === '[' && index + 1 < pattern.length && byteAt(pattern, index + 1) === ':') {
			let closing = pattern.slice(index + 2).indexOf(']');
			if (closing < 0) {
				return undefined;
			}
			closing += index + 2;
			if (closing - 1 < index + 2 || byteAt(pattern, closing - 1) !== ':') {
				// No `:]` before the next `]`, so this `[` is an ordinary member.
				members['['.charCodeAt(0)] = true;
				rangeStart = '['.charCodeAt(0);
			} else {
				if (!addClass(members, pattern.slice(index + 2, closing - 1))) {
					return undefined;
				}
				index = closing;
				rangeStart = -1;
			}
		} else {
			members[character.charCodeAt(0)] = true;
			rangeStart = character.charCodeAt(0);
		}
		index++;
	}
	if (negated) {
		for (let member = 0; member < members.length; member++) {
			members[member] = members[member] !== true;
		}
	}
	return { set: members, end: index };
}

// glob.go: glob. never marks one that can match nothing (a set left open, an unknown class, a trailing
// backslash); literal holds the text of one with no wildcard at all, so matching it is one comparison.
// isSuffix marks a base-name glob that is one `*` and then literal text, `*.log`: matching it is a
// suffix comparison. It is the commonest shape after a plain name.
export class Glob {
	readonly tokens: readonly Token[];
	readonly never: boolean;
	readonly isLiteral: boolean;
	readonly literal: string;
	readonly isSuffix: boolean;

	constructor(tokens: readonly Token[], never: boolean, isLiteral: boolean, literal: string, isSuffix: boolean) {
		this.tokens = tokens;
		this.never = never;
		this.isLiteral = isLiteral;
		this.literal = literal;
		this.isSuffix = isSuffix;
	}

	// glob.go: (*glob).reach, ahead of matches because of gap 3 in GAPS.md. It marks position, and every
	// position after it that a token matching nothing passes straight on to.
	reach(positions: boolean[], start: number): void {
		let position = start;
		for (;;) {
			addPosition(positions, position);
			const tokenEntry = this.tokens[position];
			if (tokenEntry === undefined) {
				return;
			}
			switch (tokenEntry.kind) {
				case 'starToken':
				case 'everythingToken':
				case 'directoriesToken':
					position++;
					break;
				default:
					return;
			}
		}
	}

	// glob.go: (*glob).matches. It reports whether the glob matches all of text. path keeps `?`, `*` and
	// sets out of `/`, as it does for a pattern matched against a path; the `**` tokens exist only in a
	// glob compiled for a path.
	matches(textCharacters: string, path: boolean): boolean {
		if (this.never) {
			return false;
		}
		const text = utf8Bytes(textCharacters);
		if (this.isLiteral) {
			return text === this.literal;
		}
		if (this.isSuffix) {
			return text.endsWith(this.literal);
		}

		// One place per position in the pattern, plus a second set for a directoriesToken that has
		// consumed bytes since its last `/`, which may not yet hand on to the token after it. The Go keeps
		// these as bit sets on the stack; here they are arrays of booleans.
		const places = this.tokens.length + 1;
		let current = newPositionSet(places);
		let inside = newPositionSet(places);
		let nextPositions = newPositionSet(places);
		let nextInside = newPositionSet(places);
		this.reach(current, 0);

		for (let textIndex = 0; textIndex < text.length; textIndex++) {
			const character = byteAt(text, textIndex);
			clearPositions(nextPositions);
			clearPositions(nextInside);
			const crossesNothing = path && character === '/';
			let advanced = false;
			for (let position = 0; position < this.tokens.length; position++) {
				const tokenEntry = this.tokens[position] ?? panic('glob: a position past the last token');
				const reached = hasPosition(current, position);
				if (!reached && !hasPosition(inside, position)) {
					continue;
				}
				switch (tokenEntry.kind) {
					case 'literalToken':
						if (reached && character === tokenEntry.literal) {
							this.reach(nextPositions, position + 1);
							advanced = true;
						}
						break;
					case 'anyByteToken':
						if (reached && !crossesNothing) {
							this.reach(nextPositions, position + 1);
							advanced = true;
						}
						break;
					case 'setToken':
						if (reached && tokenEntry.set[character.charCodeAt(0)] === true && !crossesNothing) {
							this.reach(nextPositions, position + 1);
							advanced = true;
						}
						break;
					case 'starToken':
						if (reached && !crossesNothing) {
							this.reach(nextPositions, position);
							advanced = true;
						}
						break;
					case 'everythingToken':
						if (reached) {
							this.reach(nextPositions, position);
							advanced = true;
						}
						break;
					case 'directoriesToken':
						addPosition(nextInside, position);
						advanced = true;
						if (character === '/') {
							this.reach(nextPositions, position + 1);
						}
						break;
				}
			}
			if (!advanced) {
				return false;
			}
			const swappedPositions = current;
			current = nextPositions;
			nextPositions = swappedPositions;
			const swappedInside = inside;
			inside = nextInside;
			nextInside = swappedInside;
		}
		return hasPosition(current, this.tokens.length);
	}

}

// glob.go: glob{never: true}.
function neverGlob(): Glob {
	return new Glob([], true, false, '', false);
}

// glob.go: compileGlob. It reads pattern. path says whether it will be matched against a path, which is
// what gives `**` its meaning; against a base name every run of asterisks is one `*`.
export function compileGlob(patternCharacters: string, path: boolean): Glob {
	const pattern = utf8Bytes(patternCharacters);
	const tokens: Token[] = [];
	let literal = '';
	let isLiteral = true;
	for (let index = 0; index < pattern.length; index++) {
		const character = byteAt(pattern, index);
		switch (character) {
			case '\\': {
				index++;
				if (index >= pattern.length) {
					return neverGlob();
				}
				tokens.push({ kind: 'literalToken', literal: byteAt(pattern, index) });
				literal += byteAt(pattern, index);
				continue;
			}

			case '?':
				tokens.push({ kind: 'anyByteToken' });
				break;

			case '[': {
				const compiledSet = compileSet(pattern, index);
				if (compiledSet === undefined) {
					return neverGlob();
				}
				tokens.push({ kind: 'setToken', set: compiledSet.set });
				index = compiledSet.end;
				break;
			}

			case '*': {
				let runEnd = index;
				while (runEnd + 1 < pattern.length && byteAt(pattern, runEnd + 1) === '*') {
					runEnd++;
				}
				let kind: 'starToken' | 'everythingToken' | 'directoriesToken' = 'starToken';
				if (path && runEnd > index) {
					const segmentStart = index === 0 || byteAt(pattern, index - 1) === '/';
					const followedBySlash = runEnd + 1 < pattern.length && byteAt(pattern, runEnd + 1) === '/';
					const followedByEscapedSlash =
						runEnd + 2 < pattern.length && byteAt(pattern, runEnd + 1) === '\\' && byteAt(pattern, runEnd + 2) === '/';
					if (segmentStart && runEnd + 1 === pattern.length) {
						kind = 'everythingToken';
					} else if (segmentStart && followedBySlash) {
						// The segment's slash belongs to the token: zero directories consume nothing at all.
						kind = 'directoriesToken';
						runEnd++;
					} else if (segmentStart && followedByEscapedSlash) {
						kind = 'directoriesToken';
						runEnd += 2;
					}
				}
				tokens.push({ kind });
				index = runEnd;
				break;
			}

			default:
				tokens.push({ kind: 'literalToken', literal: character });
				literal += character;
				continue;
		}
		isLiteral = false;
	}
	if (isLiteral) {
		return new Glob(tokens, false, true, literal, false);
	}
	const first = tokens[0];
	if (!path && first !== undefined && first.kind === 'starToken') {
		let suffix = '';
		for (const rest of tokens.slice(1)) {
			if (rest.kind !== 'literalToken') {
				return new Glob(tokens, false, false, '', false);
			}
			suffix += rest.literal;
		}
		return new Glob(tokens, false, false, suffix, true);
	}
	return new Glob(tokens, false, false, '', false);
}
