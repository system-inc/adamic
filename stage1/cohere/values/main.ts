// The port's driver: it reads the cases file named by its first argument, parses every case's value
// with { loose: true } (Prettier's options) and then { loose: false } (the library's default), and
// prints each tree, so its output can be held byte for byte to Go cohere's (values_test.go, whose Go
// side prints the Go's trees the same way) and to the library's own (testdata/library.cjs).
//
// A cases file has one case per line: the value, with a backslash, a tab, a newline and a carriage
// return written \\, \t, \n and \r.
//
// The output, for each case and each mode, is `case <number> loose` (or `strict`), then the tree one
// node per line, indented two spaces a level: the node's type, then each of its own fields but type,
// parent and nodes, sorted, as `key=value`, then ` nodes=<count>` for a container. A string is quoted
// with \\, \", \n, \r, \t, and \u00XX for any other control character; a number is written as
// JavaScript writes it (NaN included); and raws and source are written as {key:value,...}, keys sorted.
// Where the library throws, it is `error <String(error)>`.
//
// With a second argument, `count`, it prints only how many parses succeeded and how many nodes the
// values hold, which is the parse alone, to time against Go cohere's Parse over the same cases.
//
//	node oracle/node.mjs stage1/cohere/values/main.ts stage1/cohere/values/sample-cases.txt

import { panic, programArguments, readTextFile } from 'adamic';
import { parse } from './values.ts';
import { hasNodes, keysOf, type Position, type Source, type ValueNode, type ValueTree } from './nodes.ts';

const hexDigits = '0123456789abcdef';

// escapeOf is how quote writes a character, or '' for a character written as itself. It would say
// undefined for that, but stage 0 writes `return undefined` from a function returning string | undefined
// as C that clang refuses (mediaquery's GAPS.md, gap 2).
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

// quote is a string in the output's quotes. It copies the runs between escapes whole.
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

function position(at: Position): string {
	return `{column:${at.column},line:${at.line}}`;
}

function source(of: Source): string {
	return `{end:${position(of.end)},start:${position(of.start)}}`;
}

// field is the node's field key, as the output writes it.
function field(node: ValueNode, key: string): string {
	switch (key) {
		case 'inline':
			return String(node.inline);
		case 'quoted':
			return String(node.quoted);
		case 'isColor':
			return node.isColor ? 'true' : 'false';
		case 'isHex':
			return node.isHex ? 'true' : 'false';
		case 'parenType':
			return '""';
		case 'raws': {
			const raws = node.raws;
			const quoteField = raws.quote === undefined ? '' : `,quote:${quote(raws.quote)}`;
			return `{after:${quote(raws.after)},before:${quote(raws.before)}${quoteField}}`;
		}
		case 'source': {
			const of = node.source;
			return of === undefined ? '<undefined>' : source(of);
		}
		case 'sourceIndex':
			return `${node.sourceIndex}`;
		case 'unbalanced':
			return `${node.unbalanced}`;
		case 'unit':
			return quote(node.unit);
		case 'value':
			return quote(node.value);
	}
	// return panic(...), but stage 0 doesn't lower that yet (gitignore's GAPS.md, gap 9).
	panic(`no field ${key}`);
}

// render writes a node and its children at depth.
function render(tree: ValueTree, depth: number, lines: string[]): void {
	const node = tree.node;
	let line = `${'  '.repeat(depth)}${node.type}`;
	for (const key of keysOf(node)) {
		line += ` ${key}=${field(node, key)}`;
	}
	if (hasNodes(node.shape)) {
		line += ` nodes=${tree.nodes.length}`;
	}
	lines.push(line);
	for (const child of tree.nodes) {
		render(child, depth + 1, lines);
	}
}

// count is how many nodes a tree holds, itself included.
function count(tree: ValueTree): number {
	let total = 1;
	for (const child of tree.nodes) {
		total += count(child);
	}
	return total;
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
	const value = unescape(lines[index] ?? '');
	for (const loose of [true, false]) {
		const parsed = parse(value, loose);
		if (countOnly) {
			if (parsed.kind === 'Ok') {
				parsedCount++;
				nodeCount += count(parsed.root);
			}
			continue;
		}
		console.log(`case ${index} ${loose ? 'loose' : 'strict'}`);
		if (parsed.kind === 'Error') {
			console.log(`error ${parsed.message}`);
			continue;
		}
		const rendered: string[] = [];
		render(parsed.root, 0, rendered);
		console.log(rendered.join('\n'));
	}
}
if (countOnly) {
	console.log(`${parsedCount} of ${lines.length * 2} parses succeeded, ${nodeCount} nodes`);
}
