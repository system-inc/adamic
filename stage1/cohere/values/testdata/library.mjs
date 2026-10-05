// The library's own answers, in main.ts's words: postcss-values-parser 2.0.1, the JavaScript both
// cohere's Go and the port were written from, loaded as Prettier's parse-value.js loads it
// (lib/parser.js) and run on every case of a cases file with { loose: true } and { loose: false }.
// values_test.go runs it when ADAMIC_VALUES_LIBRARY names a directory it can require the library from:
//
//	npm install postcss-values-parser@2.0.1
//
// Every node is written with all of its own fields but type, parent and nodes, sorted, whatever they
// are, so a field the port leaves out or adds shows as a difference.
//
//	node library.mjs <library directory> <cases file>

import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const [libraryDirectory, casesPath] = process.argv.slice(2);
const require = createRequire(join(libraryDirectory, 'package.json'));
const Parser = require('postcss-values-parser/lib/parser.js');

function quote(text) {
	let quoted = '"';
	for (let index = 0; index < text.length; index++) {
		const unit = text.charCodeAt(index);
		if (unit === 0x5c) {
			quoted += '\\\\';
		} else if (unit === 0x22) {
			quoted += '\\"';
		} else if (unit === 0x0a) {
			quoted += '\\n';
		} else if (unit === 0x0d) {
			quoted += '\\r';
		} else if (unit === 0x09) {
			quoted += '\\t';
		} else if (unit < 0x20 || unit === 0x7f) {
			quoted += '\\u' + unit.toString(16).padStart(4, '0');
		} else {
			quoted += text[index];
		}
	}
	return quoted + '"';
}

function value(of) {
	switch (typeof of) {
		case 'undefined':
			return '<undefined>';
		case 'string':
			return quote(of);
		case 'number':
		case 'boolean':
			return `${of}`;
		case 'object':
			if (of !== null && !Array.isArray(of)) {
				return `{${Object.keys(of).sort().map((key) => `${key}:${value(of[key])}`).join(',')}}`;
			}
	}
	return `<a ${typeof of}>`;
}

function render(node, depth, lines) {
	let line = `${'  '.repeat(depth)}${node.type}`;
	for (const key of Object.keys(node).sort()) {
		if (key === 'type' || key === 'parent' || key === 'nodes') {
			continue;
		}
		line += ` ${key}=${value(node[key])}`;
	}
	if (node.nodes !== undefined) {
		line += ` nodes=${node.nodes.length}`;
	}
	lines.push(line);
	for (const child of node.nodes ?? []) {
		render(child, depth + 1, lines);
	}
}

const unescape = (field) => field.replace(/\\(.)/g, (_, character) => ({ '\\': '\\', t: '\t', n: '\n', r: '\r' })[character]);

const lines = readFileSync(casesPath, 'utf8').split('\n');
lines.pop();
const output = [];
lines.forEach((line, index) => {
	const text = unescape(line);
	for (const loose of [true, false]) {
		output.push(`case ${index} ${loose ? 'loose' : 'strict'}`);
		try {
			render(new Parser(text, { loose }).parse(), 0, output);
		} catch (error) {
			output.push(`error ${String(error)}`);
		}
	}
});
process.stdout.write(output.join('\n') + '\n');
