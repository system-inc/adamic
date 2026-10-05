// The library's own answers, in main.ts's words: postcss-media-query-parser 0.2.3, the JavaScript both
// cohere's Go and the port were written from, run on every case of a cases file. mediaquery_test.go runs
// it when ADAMIC_MEDIA_QUERY_LIBRARY names a directory it can require the library from:
//
//	npm install postcss-media-query-parser@0.2.3
//
// Where the library would never return (an unclosed url(, whose parentheses loop reads past the end
// forever), this prints the Go's error instead of running it, so the comparison doesn't hang.
//
//	node library.mjs <library directory> <cases file>

import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const [libraryDirectory, casesPath] = process.argv.slice(2);
const require = createRequire(join(libraryDirectory, 'package.json'));
const parseMedia = require('postcss-media-query-parser').default;

const errorUnclosedUrl = 'postcss-media-query-parser never returns on an unclosed url(: its parentheses loop reads past the end forever';

// neverReturns is whether parseMediaList's url( loop would run past the end: the library's own match and
// loop, stopped at the end.
function neverReturns(params) {
	const match = /^(\s*)url\s*\(/.exec(params);
	if (match === null) {
		return false;
	}
	let level = 1;
	for (let index = match[0].length; level > 0; index++) {
		if (index >= params.length) {
			return true;
		}
		if (params[index] === '(') {
			level++;
		}
		if (params[index] === ')') {
			level--;
		}
	}
	return false;
}

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

function render(node, depth, lines) {
	let line = `${'  '.repeat(depth)}${node.type ?? '<undefined>'} ${quote(node.value)} @${node.sourceIndex} before=${quote(node.before)} after=${quote(node.after)}`;
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
	const params = unescape(line);
	output.push(`case ${index}`);
	if (neverReturns(params)) {
		output.push(`error ${errorUnclosedUrl}`);
		return;
	}
	try {
		render(parseMedia(params), 0, output);
	} catch (error) {
		output.push(`error ${error.message}`);
	}
});
process.stdout.write(output.join('\n') + '\n');
