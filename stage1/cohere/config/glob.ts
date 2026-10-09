// cohere/internal/lint/configuration/glob.go, matching bytes as the Go does.
import { utf8At, utf8Length } from 'adamic';

function splitPath(value: string): string[] {
	let start = 0;
	let end = value.length;
	while (start < end && value[start] === '/') { start++; }
	while (end > start && value[end - 1] === '/') { end--; }
	return start === end ? [] : value.slice(start, end).split('/');
}

function splitAlternatives(body: string): string[] {
	const alternatives: string[] = [];
	let depth = 0;
	let start = 0;
	for (let index = 0; index < body.length; index++) {
		switch (body[index]) {
			case '{': depth++; break;
			case '}': depth--; break;
			case ',':
				if (depth === 0) { alternatives.push(body.slice(start, index)); start = index + 1; }
				break;
		}
	}
	alternatives.push(body.slice(start));
	return alternatives;
}

function expandBraces(pattern: string): string[] {
	const open = pattern.indexOf('{');
	if (open < 0) { return [pattern]; }
	let depth = 0;
	let closing = -1;
	for (let index = open; index < pattern.length; index++) {
		switch (pattern[index]) {
			case '{': depth++; break;
			case '}': depth--; if (depth === 0) { closing = index; } break;
		}
		if (closing >= 0) { break; }
	}
	if (closing < 0) { return [pattern]; }
	const expanded: string[] = [];
	for (const alternative of splitAlternatives(pattern.slice(open + 1, closing))) {
		for (const value of expandBraces(pattern.slice(0, open) + alternative + pattern.slice(closing + 1))) {
			expanded.push(value);
		}
	}
	return expanded;
}

function matchSegment(pattern: string, segment: string): boolean {
	let patternIndex = 0;
	let segmentIndex = 0;
	let starPattern = -1;
	let starSegment = 0;
	const patternLength = utf8Length(pattern);
	const segmentLength = utf8Length(segment);
	while (segmentIndex < segmentLength) {
		if (patternIndex < patternLength && (utf8At(pattern, patternIndex) === 63 || utf8At(pattern, patternIndex) === utf8At(segment, segmentIndex))) {
			patternIndex++;
			segmentIndex++;
		} else if (patternIndex < patternLength && utf8At(pattern, patternIndex) === 42) {
			starPattern = patternIndex;
			starSegment = segmentIndex;
			patternIndex++;
		} else if (starPattern >= 0) {
			starSegment++;
			segmentIndex = starSegment;
			patternIndex = starPattern + 1;
		} else { return false; }
	}
	while (patternIndex < patternLength && utf8At(pattern, patternIndex) === 42) { patternIndex++; }
	return patternIndex === patternLength;
}

function matchSegments(pattern: readonly string[], path: readonly string[], patternStart: number, pathStart: number): boolean {
	let patternIndex = patternStart;
	let pathIndex = pathStart;
	while (patternIndex < pattern.length) {
		const segment = pattern[patternIndex] ?? '';
		if (segment === '**') {
			if (patternIndex + 1 === pattern.length) { return true; }
			for (let index = pathIndex; index <= path.length; index++) {
				if (matchSegments(pattern, path, patternIndex + 1, index)) { return true; }
			}
			return false;
		}
		if (pathIndex === path.length || !matchSegment(segment, path[pathIndex] ?? '')) { return false; }
		patternIndex++;
		pathIndex++;
	}
	return pathIndex === path.length;
}

export class LintGlob {
	readonly #alternatives: readonly (readonly string[])[];
	constructor(pattern: string) {
		this.#alternatives = expandBraces(pattern).map((alternative) => splitPath(alternative));
	}
	matches(path: string): boolean {
		const segments = splitPath(path);
		for (const alternative of this.#alternatives) {
			if (matchSegments(alternative, segments, 0, 0)) { return true; }
		}
		return false;
	}
}

export function matchAny(patterns: readonly LintGlob[], path: string): boolean {
	for (const pattern of patterns) { if (pattern.matches(path)) { return true; } }
	return false;
}
