// TypeScript vfsmatch, not lint globs: no braces, Unicode scalar ? and hidden/package guards.
import { clean } from '../gitignore/path.ts';
import { join } from '../formatfiles/golang.ts';
function packageFolder(name: string): boolean { return ['node_modules', 'bower_components', 'jspm_packages'].includes(name.toLowerCase()); }
function segmentMatches(pattern: string, name: string): boolean {
	const characters: string[] = [];
	for (const character of name) { characters.push(character); }
	const patternCharacters: string[] = [];
	for (const character of pattern) { patternCharacters.push(character); }
	let at = 0;
	let index = 0;
	let star = -1;
	let consumed = 0;
	while (index < characters.length) {
		if (at < patternCharacters.length && (patternCharacters[at] === '?' || patternCharacters[at] === characters[index])) { at++; index++; }
		else if (patternCharacters[at] === '*') { star = at; consumed = index; at++; }
		else if (star >= 0) { consumed++; index = consumed; at = star + 1; }
		else { return false; }
	}
	while (patternCharacters[at] === '*') { at++; }
	return at === patternCharacters.length;
}
export function absolutePath(base: string, path: string): string { const normalized = path.split('\\').join('/'); return normalized.startsWith('/') ? clean(normalized) : join(base, normalized); }
export class TsGlob {
	readonly parts: readonly string[];
	readonly exclude: boolean;
	readonly valid: boolean;
	constructor(spec: string, base: string, exclude: boolean) {
		const parts = absolutePath(base, spec).split('/');
		const last = parts[parts.length - 1] ?? '';
		this.valid = exclude || last !== '**';
		if (!last.includes('.') && !last.includes('*') && !last.includes('?')) { parts.push('**'); parts.push('*'); }
		this.parts = parts;
		this.exclude = exclude;
	}
	matches(path: string, prefix: boolean): boolean { return this.valid && this.inner(path.split('/'), 0, 0, prefix); }
	inner(path: readonly string[], at: number, index: number, prefix: boolean): boolean {
		if (index === path.length) {
			if (prefix) { return true; }
			for (let rest = at; rest < this.parts.length; rest++) { if (this.parts[rest] !== '**') { return false; } }
			return true;
		}
		if (at === this.parts.length) { return this.exclude && !prefix; }
		const pattern = this.parts[at] ?? '';
		const name = path[index] ?? '';
		if (pattern === '**') {
			if (this.inner(path, at + 1, index, prefix)) { return true; }
			if (!this.exclude && (name.startsWith('.') || packageFolder(name))) { return false; }
			return this.inner(path, at, index + 1, prefix);
		}
		if (pattern.includes('*') || pattern.includes('?')) {
			if (!this.exclude && (packageFolder(name) || (name.startsWith('.') && (pattern.startsWith('*') || pattern.startsWith('?'))))) { return false; }
			if (!segmentMatches(pattern, name)) { return false; }
			if (!this.exclude && !prefix && name.endsWith('.min.js') && !pattern.includes('.min.') && !pattern.includes('.min.js')) { return false; }
		} else if (pattern !== name) { return false; }
		return this.inner(path, at + 1, index + 1, prefix);
	}
}
