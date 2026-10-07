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
// `é` is two bytes. A Go string is its bytes, indexed as text[index]; an Adamic string's UTF-8 is read
// the same way, in place, with utf8At and utf8Length from 'adamic', so the glob reads its pattern and
// its text exactly as the Go does, a byte a number from 0 to 255. Matching walks the text once while
// tracking every place in the pattern it could have reached, so no pattern backtracks: `*a*a*a*a*b`
// against a long run of `a` costs one pass.

import { panic, utf8At, utf8Length } from 'adamic';

// glob.go: [256]bool, a set of bytes.
type ByteSet = readonly boolean[];

// glob.go: tokenKind and token. The Go's one struct, whose set is nil except on a setToken, is a
// discriminated union here, so a set is there exactly when the kind says it is.
type Token =
	| { readonly kind: 'literalToken'; readonly literal: number } // one byte, itself
	| { readonly kind: 'anyByteToken' } // `?`
	| { readonly kind: 'setToken'; readonly set: ByteSet } // `[...]`
	| { readonly kind: 'starToken' } // `*`: any run, within one segment against a path
	| { readonly kind: 'everythingToken' } // a trailing `**` segment: any run, slashes included
	| { readonly kind: 'directoriesToken' }; // `**/` as a segment: nothing, or any run that ends in `/`

// glob.go: glob. never marks one that can match nothing (a set left open, an unknown class, a trailing
// backslash); literal holds the bytes of one with no wildcard at all, so matching it is one comparison.
// isSuffix marks a base-name glob that is one `*` and then literal text, `*.log`: matching it is a
// suffix comparison. It is the commonest shape after a plain name.
export class Glob {
	readonly tokens: readonly Token[];
	readonly never: boolean;
	readonly isLiteral: boolean;
	readonly literal: readonly number[];
	readonly isSuffix: boolean;

	constructor(tokens: readonly Token[], never: boolean, isLiteral: boolean, literal: readonly number[], isSuffix: boolean) {
		this.tokens = tokens;
		this.never = never;
		this.isLiteral = isLiteral;
		this.literal = literal;
		this.isSuffix = isSuffix;
	}

	// glob.go: (*glob).matches. It reports whether the glob matches all of text. path keeps `?`, `*` and
	// sets out of `/`, as it does for a pattern matched against a path; the `**` tokens exist only in a
	// glob compiled for a path.
	matches(text: string, path: boolean): boolean {
		if (this.never) {
			return false;
		}
		const length = utf8Length(text);
		if (this.isLiteral) {
			return length === this.literal.length && endsWith(text, length, this.literal);
		}
		if (this.isSuffix) {
			return length >= this.literal.length && endsWith(text, length, this.literal);
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

		for (let textIndex = 0; textIndex < length; textIndex++) {
			const character = utf8At(text, textIndex);
			clearPositions(nextPositions);
			clearPositions(nextInside);
			const crossesNothing = path && character === 0x2f; // '/'
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
						if (reached && tokenEntry.set[character] === true && !crossesNothing) {
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
						if (character === 0x2f) {
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

	// glob.go: (*glob).reach. It marks position, and every position after it that a token matching
	// nothing passes straight on to.
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

}

// endsWith reports whether text, length bytes, ends with bytes: the Go's strings.HasSuffix, and with
// equal lengths its ==.
function endsWith(text: string, length: number, bytes: readonly number[]): boolean {
	const start = length - bytes.length;
	for (let index = 0; index < bytes.length; index++) {
		if (utf8At(text, start + index) !== bytes[index]) {
			return false;
		}
	}
	return true;
}

// glob.go: glob{never: true}.
function neverGlob(): Glob {
	return new Glob([], true, false, [], false);
}

// glob.go: compileGlob. It reads pattern. path says whether it will be matched against a path, which is
// what gives `**` its meaning; against a base name every run of asterisks is one `*`.
export function compileGlob(pattern: string, path: boolean): Glob {
	const length = utf8Length(pattern);
	const tokens: Token[] = [];
	const literal: number[] = [];
	let isLiteral = true;
	for (let index = 0; index < length; index++) {
		const character = utf8At(pattern, index);
		switch (character) {
			case 0x5c: {
				// '\\'
				index++;
				if (index >= length) {
					return neverGlob();
				}
				tokens.push({ kind: 'literalToken', literal: utf8At(pattern, index) });
				literal.push(utf8At(pattern, index));
				continue;
			}

			case 0x3f: // '?'
				tokens.push({ kind: 'anyByteToken' });
				break;

			case 0x5b: {
				// '['
				const compiledSet = compileSet(pattern, length, index);
				if (compiledSet === undefined) {
					return neverGlob();
				}
				tokens.push({ kind: 'setToken', set: compiledSet.set });
				index = compiledSet.end;
				break;
			}

			case 0x2a: {
				// '*'
				let runEnd = index;
				while (runEnd + 1 < length && utf8At(pattern, runEnd + 1) === 0x2a) {
					runEnd++;
				}
				let kind: 'starToken' | 'everythingToken' | 'directoriesToken' = 'starToken';
				if (path && runEnd > index) {
					const segmentStart = index === 0 || utf8At(pattern, index - 1) === 0x2f;
					const followedBySlash = runEnd + 1 < length && utf8At(pattern, runEnd + 1) === 0x2f;
					const followedByEscapedSlash = runEnd + 2 < length && utf8At(pattern, runEnd + 1) === 0x5c && utf8At(pattern, runEnd + 2) === 0x2f;
					if (segmentStart && runEnd + 1 === length) {
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
				literal.push(character);
				continue;
		}
		isLiteral = false;
	}
	if (isLiteral) {
		return new Glob(tokens, false, true, literal, false);
	}
	const first = tokens[0];
	if (!path && first !== undefined && first.kind === 'starToken') {
		const suffix: number[] = [];
		for (const rest of tokens.slice(1)) {
			if (rest.kind !== 'literalToken') {
				return new Glob(tokens, false, false, [], false);
			}
			suffix.push(rest.literal);
		}
		return new Glob(tokens, false, false, suffix, true);
	}
	return new Glob(tokens, false, false, [], false);
}

// glob.go: compileSet's results. The Go returns (set, end, ok); here undefined is not ok.
interface CompiledSet {
	readonly set: ByteSet;
	readonly end: number;
}

// glob.go: compileSet. It reads the bracket expression that opens at start in pattern, length bytes,
// returning its members and the index of its closing `]`, or undefined for one that never closes or
// names an unknown class.
function compileSet(pattern: string, length: number, start: number): CompiledSet | undefined {
	const members = Array.from({ length: 256 }, () => false);
	let index = start + 1;
	let negated = false;
	if (index < length && (utf8At(pattern, index) === 0x21 || utf8At(pattern, index) === 0x5e)) {
		// '!' or '^'
		negated = true;
		index++;
	}

	// rangeStart is the byte a following `-` would extend, or -1 after a class or a range, which
	// nothing extends.
	let rangeStart = -1;
	let first = true;
	for (;;) {
		if (index >= length) {
			return undefined;
		}
		const character = utf8At(pattern, index);
		if (character === 0x5d && !first) {
			// ']'
			break;
		}
		first = false;

		if (character === 0x5c) {
			// '\\'
			index++;
			if (index >= length) {
				return undefined;
			}
			members[utf8At(pattern, index)] = true;
			rangeStart = utf8At(pattern, index);
		} else if (character === 0x2d && rangeStart >= 0 && index + 1 < length && utf8At(pattern, index + 1) !== 0x5d) {
			// '-', not before ']'
			index++;
			let upper = utf8At(pattern, index);
			if (upper === 0x5c) {
				index++;
				if (index >= length) {
					return undefined;
				}
				upper = utf8At(pattern, index);
			}
			for (let member = rangeStart; member <= upper; member++) {
				members[member] = true;
			}
			rangeStart = -1;
		} else if (character === 0x5b && index + 1 < length && utf8At(pattern, index + 1) === 0x3a) {
			// '[:'
			let closing = index + 2;
			while (closing < length && utf8At(pattern, closing) !== 0x5d) {
				closing++;
			}
			if (closing >= length) {
				return undefined;
			}
			if (closing - 1 < index + 2 || utf8At(pattern, closing - 1) !== 0x3a) {
				// No `:]` before the next `]`, so this `[` is an ordinary member.
				members[0x5b] = true;
				rangeStart = 0x5b;
			} else {
				if (!addClass(members, className(pattern, index + 2, closing - 1))) {
					return undefined;
				}
				index = closing;
				rangeStart = -1;
			}
		} else {
			members[character] = true;
			rangeStart = character;
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

// className is the bytes from start to end of pattern as text: a class's name, which fnmatch's names
// are ASCII, so each byte is its own character, and any other byte makes a name no class has.
function className(pattern: string, start: number, end: number): string {
	let name = '';
	for (let index = start; index < end; index++) {
		name += String.fromCharCode(utf8At(pattern, index));
	}
	return name;
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
			inClass = isLetter;
			break;
		case 'blank':
			inClass = (character) => character === 0x20 || character === 0x09;
			break;
		case 'cntrl':
			inClass = (character) => character < 0x20 || character === 0x7f;
			break;
		case 'digit':
			inClass = isDigit;
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
			inClass = (character) => isDigit(character) || ((character | 0x20) >= 0x61 && (character | 0x20) <= 0x66);
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

// glob.go: isLetter and isDigit, over a byte's value.
function isLetter(character: number): boolean {
	return (character | 0x20) >= 0x61 && (character | 0x20) <= 0x7a;
}

function isDigit(character: number): boolean {
	return character >= 0x30 && character <= 0x39;
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
