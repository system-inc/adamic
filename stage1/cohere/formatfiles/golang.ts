// What the formatfiles Go reaches for in Go's standard library and has no JavaScript twin, written here
// as Go answers it: strings.ToLower, the order sort.Strings and Go's map printing put strings in, and
// path/filepath's Ext and Rel for the absolute, slash-separated paths a walk on Linux makes. Clean, Dir
// and Base come from the gitignore slice's path.ts.

import { clean } from '../gitignore/path.ts';

// strings.ToLower maps each character on its own, with Unicode's simple lowercase mapping. JavaScript's
// toLowerCase uses the full mapping in context, which differs in two places: U+0130 (İ) lowercases to
// two characters, i and a combining dot, where Go's is i alone; and Σ before a word's end is final ς,
// where Go's is always σ. Lowering one character at a time leaves only the first, and both sides read
// Unicode 17.0 (Go's unicode.Version and Node's process.versions.unicode, measured), so the tables agree.
export function goToLower(text: string): string {
	let lowered = '';
	for (const character of text) {
		lowered += character === 'İ' ? 'i' : character.toLowerCase();
	}
	return lowered;
}

// compareCodePoints orders two strings as Go's sort.Strings does, by their UTF-8 bytes, which is their
// code points' order. JavaScript's < compares UTF-16 units, and a supplementary character's high
// surrogate (U+D800 to U+DBFF) sorts before U+E000 to U+FFFF there, after them in UTF-8.
export function compareCodePoints(left: string, right: string): number {
	let leftIndex = 0;
	let rightIndex = 0;
	while (leftIndex < left.length && rightIndex < right.length) {
		const leftPoint = left.codePointAt(leftIndex) ?? 0;
		const rightPoint = right.codePointAt(rightIndex) ?? 0;
		if (leftPoint !== rightPoint) {
			return leftPoint < rightPoint ? -1 : 1;
		}
		leftIndex += leftPoint > 0xffff ? 2 : 1;
		rightIndex += rightPoint > 0xffff ? 2 : 1;
	}
	const leftDone = leftIndex >= left.length;
	const rightDone = rightIndex >= right.length;
	if (leftDone && rightDone) {
		return 0;
	}
	return leftDone ? -1 : 1;
}

// filepath.Ext: the suffix from the final dot of the last element, or '' when it has none.
export function ext(path: string): string {
	for (let index = path.length - 1; index >= 0 && path[index] !== '/'; index--) {
		if (path[index] === '.') {
			return path.slice(index);
		}
	}
	return '';
}

// filepath.Join of a directory and a name holding no separator.
export function join(directory: string, name: string): string {
	return clean(`${directory}/${name}`);
}

// What filepath.Rel gives, as far as its callers here read it: '.' for the same path, the path below
// base without the base, and '..' for any path outside it, where Go spells out the way there ('../x');
// every caller only asks whether it is '..' or starts '../'. Go's error, when one path is absolute and
// the other isn't, is '', which Rel never answers otherwise. It would be undefined, but stage 0 writes
// `return undefined` from a function returning string | undefined as C that clang refuses
// (mediaquery's GAPS.md, gap 2).
export function rel(base: string, target: string): string {
	const cleanBase = clean(base);
	const cleanTarget = clean(target);
	if (cleanBase.startsWith('/') !== cleanTarget.startsWith('/')) {
		return '';
	}
	if (cleanBase === cleanTarget) {
		return '.';
	}
	const prefix = cleanBase === '/' ? '/' : `${cleanBase}/`;
	if (cleanTarget.startsWith(prefix)) {
		return cleanTarget.slice(prefix.length);
	}
	return '..';
}
