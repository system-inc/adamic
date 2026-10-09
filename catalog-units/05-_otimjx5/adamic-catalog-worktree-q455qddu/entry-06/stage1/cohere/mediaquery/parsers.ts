// A port of cohere's internal/format/css/mediaquery/parsers.go to Adamic 0.1: postcss-media-query-parser
// 0.2.3, dist/parsers.js. Each piece names the Go it reads as.
//
// Where the port differs from the Go, and why:
//
//   - Offsets. The Go walks bytes, and runes where it asks about whitespace, and its sourceIndex is in
//     bytes. The port's strings are UTF-16, as the library's are, so it walks UTF-16 units as the library
//     does, and its sourceIndex is the library's. The Go's own comment says why both walks take the same
//     decisions: every test on a character is against ASCII or whitespace, all of it in the Basic
//     Multilingual Plane. The test converts the Go's byte offsets to UTF-16 indexes with cohere's own
//     utf16OffsetsOf, as cohere's oracle test does before comparing with the library.
//   - Failures. The Go's (value, error) is a Result, since 0.1 has no exceptions and no multiple results.
//     The messages are the Go's.
//   - The declarations come callee first, as the gitignore port's do: stage 0 takes a call to a function
//     declared further down for a void one (gitignore's GAPS.md, gap 3).
//   - The lists of nodes are readonly and grow by copying (appended), where the Go appends to a slice:
//     a node keeps its children's list, so a mutable one could be made to hold its own node.
//   - The Go's (int, string) from matchUrlStart is a UrlStart, since stage 0 doesn't lower a tuple as a
//     value yet (gitignore's GAPS.md, gap 8).

import { panic } from 'adamic';
import { MediaNode, newContainer, newNode } from './nodes.ts';
import { leadingWhitespace, startsWithWhitespace, trailingWhitespace, trim } from './whitespace.ts';

// What a parse gives: its value, or the Go's error message.
export type Result<Value> = { readonly kind: 'Ok'; readonly value: Value } | { readonly kind: 'Error'; readonly message: string };

// parsers.go: errorReadingMode, the TypeError V8 throws on modesEntered[lastModeIndex].mode with
// lastModeIndex -1.
export const errorReadingMode = "Cannot read properties of undefined (reading 'mode')";

// parsers.go: errorUnclosedUrl, which stands for the infinite loop on an unclosed `url(`, where the
// library never returns.
export const errorUnclosedUrl = 'postcss-media-query-parser never returns on an unclosed url(: its parentheses loop reads past the end forever';

// nodeAt and modeAt are list[index] where the Go's index is in range by construction, checked as a Go
// index is. The Go needs no helper; the port's would be one generic function, but stage 0 doesn't
// instantiate a generic function's return type yet (gap 1 in GAPS.md).
function nodeAt(list: readonly MediaNode[], index: number): MediaNode {
	return list[index] ?? panic(`index out of range [${index}] with length ${list.length}`);
}

// appended is result.push(node) as a new list: the lists of nodes are readonly, since a node keeps its
// children's list, and stage 0 refuses a mutable list of nodes that can reach a list like it, a cycle
// reference counting couldn't free (adamic/cycle-capable; gaps 4 and 5 in GAPS.md). Each
// list holds a handful of nodes, so copying it costs little.
function appended(list: readonly MediaNode[], node: MediaNode): readonly MediaNode[] {
	return [...list, node];
}

function modeAt(list: readonly Mode[], index: number): Mode {
	return list[index] ?? panic(`index out of range [${index}] with length ${list.length}`);
}

// parsers.go: mode, one entry of modesEntered. character is the quote a string mode was entered with,
// '' for none (the Go's 0).
interface Mode {
	readonly mode: string;
	readonly isCalculationEnabled: boolean;
	readonly character: string;
}

/**
 * parsers.go: parseMediaFeature.
 *
 * Parses a media feature expression, e.g. `max-width: 10px`, `(color)`
 *
 * @param {string} string - the source expression string, can be inside parens
 * @param {Number} index - the index of `string` in the overall input
 *
 * @return {Array} an array of Nodes, the first element being a media feature,
 *    the secont - its value (may be missing)
 */
function parseMediaFeature(text: string, index: number): Result<readonly MediaNode[]> {
	const modesEntered: Mode[] = [{ mode: 'normal', isCalculationEnabled: false, character: '' }];
	let result: readonly MediaNode[] = [];
	let lastModeIndex = 0;
	let mediaFeature = '';
	let colon: MediaNode | undefined = undefined;
	let mediaFeatureValue: MediaNode | undefined = undefined;
	let indexLocal = index;

	let stringNormalized = text;
	// Strip trailing parens (if any), and correct the starting index
	if (text.length > 0 && text.startsWith('(') && text.endsWith(')')) {
		stringNormalized = text.slice(1, text.length - 1);
		indexLocal++;
	}

	for (let i = 0; i < stringNormalized.length; i++) {
		const character = stringNormalized[i] ?? panic(`index ${i} out of range`);

		// If entering/exiting a string
		if (character === "'" || character === '"') {
			const current = modeAt(modesEntered, lastModeIndex);
			if (current.isCalculationEnabled) {
				modesEntered.push({ mode: 'string', isCalculationEnabled: false, character });
				lastModeIndex++;
			} else if (current.mode === 'string' && current.character === character && (i === 0 || stringNormalized[i - 1] !== '\\')) {
				modesEntered.pop();
				lastModeIndex--;
			}
		}

		// If entering/exiting interpolation
		if (character === '{') {
			modesEntered.push({ mode: 'interpolation', isCalculationEnabled: true, character: '' });
			lastModeIndex++;
		} else if (character === '}') {
			modesEntered.pop();
			lastModeIndex--;
		}

		// The next line reads modesEntered[lastModeIndex].mode, which throws once a "}" has popped the
		// last mode. lastModeIndex is always modesEntered.length - 1: they move together until this throw.
		if (lastModeIndex < 0) {
			return { kind: 'Error', message: errorReadingMode };
		}

		// If a : is met outside of a string, function call or interpolation, than
		// this : separates a media feature and a value
		if (modeAt(modesEntered, lastModeIndex).mode === 'normal' && character === ':') {
			const mediaFeatureValueString = stringNormalized.slice(i + 1);
			const mediaFeatureValueBefore = leadingWhitespace(mediaFeatureValueString);
			mediaFeatureValue = newNode(
				trailingWhitespace(mediaFeatureValueString),
				mediaFeatureValueBefore,
				'value',
				trim(mediaFeatureValueString),
				// +1 for the colon
				mediaFeatureValueBefore.length + i + 1 + indexLocal,
			);
			// The Go sets colon's before below, once the feature's after is known.
			colon = newNode(mediaFeatureValueBefore, '', 'colon', ':', i + indexLocal);
			break;
		}

		mediaFeature += character;
	}

	// Forming a media feature node
	const mediaFeatureBefore = leadingWhitespace(mediaFeature);
	const mediaFeatureAfter = trailingWhitespace(mediaFeature);
	result = appended(result, newNode(mediaFeatureAfter, mediaFeatureBefore, 'media-feature', trim(mediaFeature), mediaFeatureBefore.length + indexLocal));

	if (colon !== undefined) {
		colon.before = mediaFeatureAfter;
		result = appended(result, colon);
	}

	if (mediaFeatureValue !== undefined) {
		result = appended(result, mediaFeatureValue);
	}

	return { kind: 'Ok', value: result };
}

// parsers.go: mediaQueryElement, the plain object parseMediaQuery builds an element in (resetNode's shape
// plus the type and sourceIndex it may gain) before handing it to Node or Container. The nodes it may
// gain are elementNodes in parseMediaQuery, not a field: with them this interface has every field a
// MediaNode has, so a MediaNode can be seen as one, and a mutable nodes field could then be given the
// node itself, a cycle stage 0 refuses (adamic/cycle-capable; gap 4 in GAPS.md).
interface MediaQueryElement {
	before: string;
	after: string;
	value: string;
	type: string;
	sourceIndex: number;
}

/**
 * parsers.go: parseMediaQuery.
 *
 * Parses a media query, e.g. `screen and (color)`, `only tv`
 *
 * @param {string} string - the source media query string
 * @param {Number} index - the index of `string` in the overall input
 *
 * @return {Array} an array of Nodes and Containers
 */
function parseMediaQuery(text: string, index: number): Result<readonly MediaNode[]> {
	let result: readonly MediaNode[] = [];

	// How many timies the parser entered parens/curly braces
	let localLevel = 0;
	// Has any keyword, media type, media feature expression or interpolation
	// ('element' hereafter) started
	let insideSomeValue = false;

	const resetNode = (): MediaQueryElement => ({ before: '', after: '', value: '', type: '', sourceIndex: 0 });

	let node = resetNode();
	let elementNodes: readonly MediaNode[] | undefined = undefined;

	for (let i = 0; i < text.length; i++) {
		const character = text[i] ?? panic(`index ${i} out of range`);
		// If not yet entered any element
		if (!insideSomeValue) {
			if (startsWithWhitespace(text, i)) {
				// A whitespace
				// Don't form 'after' yet; will do it later
				node.before += character;
			} else {
				// Not a whitespace - entering an element
				// Expression start
				if (character === '(') {
					node.type = 'media-feature-expression';
					localLevel++;
				}
				node.value = character;
				node.sourceIndex = index + i;
				insideSomeValue = true;
			}
		} else {
			// Already in the middle of some alement
			node.value += character;

			// Here parens just increase localLevel and don't trigger a start of
			// a media feature expression (since they can't be nested)
			// Interpolation start
			if (character === '{' || character === '(') {
				localLevel++;
			}
			// Interpolation/function call/media feature expression end
			if (character === ')' || character === '}') {
				localLevel--;
			}
		}

		// If exited all parens/curlies and the next symbol
		if (insideSomeValue && localLevel === 0 && (character === ')' || i === text.length - 1 || startsWithWhitespace(text, i + 1))) {
			if (node.value === 'not' || node.value === 'only' || node.value === 'and') {
				node.type = 'keyword';
			}
			// if it's an expression, parse its contents
			if (node.type === 'media-feature-expression') {
				const parsed = parseMediaFeature(node.value, node.sourceIndex);
				if (parsed.kind === 'Error') {
					return parsed;
				}
				elementNodes = parsed.value;
			}
			if (elementNodes !== undefined) {
				result = appended(result, newContainer(node.after, node.before, node.type, node.value, node.sourceIndex, elementNodes));
			} else {
				result = appended(result, newNode(node.after, node.before, node.type, node.value, node.sourceIndex));
			}
			node = resetNode();
			elementNodes = undefined;
			insideSomeValue = false;
		}
	}

	// Now process the result array - to specify undefined types of the nodes
	// and specify the `after` prop
	for (let i = 0; i < result.length; i++) {
		const current = nodeAt(result, i);
		if (i > 0) {
			nodeAt(result, i - 1).after = current.before;
		}

		// Node types. Might not be set because contains interpolation/function
		// calls or fully consists of them
		if (current.type === '') {
			if (i > 0) {
				const previous = nodeAt(result, i - 1);
				// only `and` can follow an expression
				if (previous.type === 'media-feature-expression') {
					current.type = 'keyword';
					continue;
				}
				// Anything after 'only|not' is a media type
				if (previous.value === 'not' || previous.value === 'only') {
					current.type = 'media-type';
					continue;
				}
				// Anything after 'and' is an expression
				if (previous.value === 'and') {
					current.type = 'media-feature-expression';
					continue;
				}

				if (previous.type === 'media-type') {
					// if it is the last element - it might be an expression
					// or 'and' depending on what is after it
					if (i + 1 >= result.length) {
						current.type = 'media-feature-expression';
					} else if (nodeAt(result, i + 1).type === 'media-feature-expression') {
						current.type = 'keyword';
					} else {
						current.type = 'media-feature-expression';
					}
				}
			}

			if (i === 0) {
				// `screen`, `fn( ... )`, `#{ ... }`. Not an expression, since then
				// its type would have been set by now
				if (i + 1 >= result.length) {
					current.type = 'media-type';
					continue;
				}

				// `screen and` or `#{...} (max-width: 10px)`
				const next = nodeAt(result, i + 1);
				if (next.type === 'media-feature-expression' || next.type === 'keyword') {
					current.type = 'media-type';
					continue;
				}
				if (i + 2 < result.length) {
					// `screen and (color) ...`
					if (nodeAt(result, i + 2).type === 'media-feature-expression') {
						current.type = 'media-type';
						next.type = 'keyword';
						continue;
					}
					// `only screen and ...`
					if (nodeAt(result, i + 2).type === 'keyword') {
						current.type = 'keyword';
						next.type = 'media-type';
						continue;
					}
				}
				if (i + 3 < result.length) {
					// `screen and (color) ...`
					if (nodeAt(result, i + 3).type === 'media-feature-expression') {
						current.type = 'keyword';
						next.type = 'media-type';
						nodeAt(result, i + 2).type = 'keyword';
						continue;
					}
				}
			}
		}
	}
	return { kind: 'Ok', value: result };
}

// parsers.go: matchUrlStart's two results, the match's length (0 for no match, since a match is at least
// four characters long) and its first group.
interface UrlStart {
	readonly length: number;
	readonly before: string;
}

// parsers.go: matchUrlStart, `/^(\s*)url\s*\(/.exec(string)`.
function matchUrlStart(text: string): UrlStart {
	const before = leadingWhitespace(text);
	const rest = text.slice(before.length);
	if (!rest.startsWith('url')) {
		return { length: 0, before: '' };
	}
	const afterUrl = rest.slice(3 + leadingWhitespace(rest.slice(3)).length);
	if (!afterUrl.startsWith('(')) {
		return { length: 0, before: '' };
	}
	return { length: text.length - afterUrl.length + 1, before };
}

/**
 * parsers.go: parseMediaList.
 *
 * Parses a media query list. Takes a possible `url()` at the start into
 * account, and divides the list into media queries that are parsed separately
 *
 * @param {string} string - the source media query list string
 *
 * @return {Array} an array of Nodes/Containers
 */
export function parseMediaList(text: string): Result<readonly MediaNode[]> {
	let result: readonly MediaNode[] = [];
	let interimIndex = 0;
	let levelLocal = 0;

	// Check for a `url(...)` part (if it is contents of an @import rule)
	const urlStart = matchUrlStart(text);
	if (urlStart.length > 0) {
		let i = urlStart.length;
		let parenthesesLevel = 1;
		while (parenthesesLevel > 0) {
			// string[i] past the end is undefined, and the loop never ends.
			if (i >= text.length) {
				return { kind: 'Error', message: errorUnclosedUrl };
			}
			const character = text[i];
			if (character === '(') {
				parenthesesLevel++;
			}
			if (character === ')') {
				parenthesesLevel--;
			}
			i++;
		}
		// result.unshift into an empty result.
		result = appended(result, newNode(leadingWhitespace(text.slice(i)), urlStart.before, 'url', trim(text.slice(0, i)), urlStart.before.length));
		interimIndex = i;
	}

	// Start processing the media query list
	for (let i = interimIndex; i < text.length; i++) {
		const character = text[i];

		// Dividing the media query list into comma-separated media queries
		// Only count commas that are outside of any parens
		// (i.e., not part of function call params list, etc.)
		if (character === '(') {
			levelLocal++;
		}
		if (character === ')') {
			levelLocal--;
		}
		if (levelLocal === 0 && character === ',') {
			const mediaQueryString = text.slice(interimIndex, i);
			const spaceBefore = leadingWhitespace(mediaQueryString);
			const parsed = parseMediaQuery(mediaQueryString, interimIndex);
			if (parsed.kind === 'Error') {
				return parsed;
			}
			result = appended(
				result,
				newContainer(trailingWhitespace(mediaQueryString), spaceBefore, 'media-query', trim(mediaQueryString), interimIndex + spaceBefore.length, parsed.value),
			);
			interimIndex = i + 1;
		}
	}

	const mediaQueryString = text.slice(interimIndex);
	const spaceBefore = leadingWhitespace(mediaQueryString);
	const parsed = parseMediaQuery(mediaQueryString, interimIndex);
	if (parsed.kind === 'Error') {
		return parsed;
	}
	result = appended(
		result,
		newContainer(trailingWhitespace(mediaQueryString), spaceBefore, 'media-query', trim(mediaQueryString), interimIndex + spaceBefore.length, parsed.value),
	);

	return { kind: 'Ok', value: result };
}
