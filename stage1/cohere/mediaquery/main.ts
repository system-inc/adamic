// The port's driver: it reads the cases file named by its one argument, parses every case's params, and
// prints each tree, so its output can be held byte for byte to Go cohere's (mediaquery_test.go, whose Go
// side prints the Go's trees the same way).
//
// A cases file has one case per line: the params, with a backslash, a tab, a newline and a carriage
// return written \\, \t, \n and \r.
//
// The output, for each case, is `case <number>`, then the tree one node per line, indented two spaces
// a level:
//
//	<type> "<value>" @<sourceIndex> before="<before>" after="<after>"
//
// with ` nodes=<count>` after a container's, <undefined> for a type the library leaves undefined, and
// sourceIndex in UTF-16 units; or `error <message>` where the library throws or would never return.
// Strings are quoted with \\, \", \n, \r, \t, and \u00XX for any other control character.
//
// With a second argument, `count`, it prints only how many cases parsed and how many top-level nodes
// their trees have, which is the parse alone, to time against Go cohere's Parse over the same cases.
//
//	node oracle/node.mjs stage1/cohere/mediaquery/main.ts stage1/cohere/mediaquery/sample-cases.txt

import { panic, programArguments, readTextFile } from 'adamic';
import { parse } from './index.ts';
import type { MediaNode } from './nodes.ts';

const hexDigits = '0123456789abcdef';

// escapeOf is how quote writes a character, or '' for a character written as itself. It would say
// undefined for that, but stage 0 writes `return undefined` from a function returning string | undefined
// as C that clang refuses (gap 2 in GAPS.md).
function escapeOf(character: string): string {
	switch (character) {
		case '\\':
			return '\\\\';
		case '"':
			return '\\"';
		case '\n':
			return '\\n';
		case '\r':
			return '\\r';
		case '\t':
			return '\\t';
	}
	const unit = character.charCodeAt(0);
	if (unit < 0x20 || unit === 0x7f) {
		return `\\u00${hexDigits[Math.floor(unit / 16)] ?? '?'}${hexDigits[unit % 16] ?? '?'}`;
	}
	return '';
}

// quote is a string in the output's quotes. It copies the runs between escapes whole, since most
// strings have none.
function quote(text: string): string {
	let quoted = '"';
	let runStart = 0;
	let index = 0;
	for (const character of text) {
		const escape = escapeOf(character);
		if (escape !== '') {
			quoted += text.slice(runStart, index) + escape;
			runStart = index + character.length;
		}
		index += character.length;
	}
	return `${quoted}${text.slice(runStart)}"`;
}

// render writes a node and its children at depth.
function render(node: MediaNode, depth: number, lines: string[]): void {
	const type = node.type === '' ? '<undefined>' : node.type;
	let line = `${'  '.repeat(depth)}${type} ${quote(node.value)} @${node.sourceIndex} before=${quote(node.before)} after=${quote(node.after)}`;
	const children = node.nodes;
	if (children !== undefined) {
		line += ` nodes=${children.length}`;
	}
	lines.push(line);
	if (children !== undefined) {
		for (const child of children) {
			render(child, depth + 1, lines);
		}
	}
}

// unescape reads a case's escapes.
function unescape(field: string): string {
	if (!field.includes('\\')) {
		return field;
	}
	let text = '';
	let escaped = false;
	for (const character of field) {
		if (!escaped) {
			if (character === '\\') {
				escaped = true;
			} else {
				text += character;
			}
			continue;
		}
		escaped = false;
		switch (character) {
			case '\\':
				text += '\\';
				break;
			case 't':
				text += '\t';
				break;
			case 'n':
				text += '\n';
				break;
			case 'r':
				text += '\r';
				break;
			default:
				panic(`an unknown escape \\${character} in ${field}`);
		}
	}
	return text;
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file> [count]');
const countOnly = programArguments()[1] === 'count';
const read = readTextFile(casesPath);
if (read.kind === 'Error') {
	panic(read.message);
}
const lines = read.text.split('\n');
// The file ends with a newline, so the last field is empty and is no case.
if (lines.at(-1) === '') {
	lines.pop();
}
let parsedCount = 0;
let nodeCount = 0;
for (let index = 0; index < lines.length; index++) {
	const params = unescape(lines[index] ?? '');
	if (countOnly) {
		const parsed = parse(params);
		if (parsed.kind === 'Ok') {
			parsedCount++;
			// parsed.value.nodes?.length ?? 0, but stage 0 doesn't lower ?. on an array yet (gap 3 in GAPS.md).
			const children = parsed.value.nodes;
			nodeCount += children === undefined ? 0 : children.length;
		}
		continue;
	}
	console.log(`case ${index}`);
	const parsed = parse(params);
	if (parsed.kind === 'Error') {
		console.log(`error ${parsed.message}`);
		continue;
	}
	const rendered: string[] = [];
	render(parsed.value, 0, rendered);
	console.log(rendered.join('\n'));
}
if (countOnly) {
	console.log(`${parsedCount} of ${lines.length} cases parsed, ${nodeCount} nodes`);
}
