// The library's own answers, in main.ts's words: graphql-js 17.0.2, the JavaScript both cohere's Go and
// the port were written from, called as Prettier's parser-graphql.js calls it, on every case of a cases
// file. graphql_test.go runs it when ADAMIC_GRAPHQL_LIBRARY names a directory it can require the
// library from:
//
//	npm install graphql@17.0.2
//
//	node library.mjs <library directory> <cases file>

import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const [libraryDirectory, casesPath] = process.argv.slice(2);
const require = createRequire(join(libraryDirectory, 'package.json'));
const { parse, version } = require('graphql');
if (version !== '17.0.2') {
	throw new Error(`graphql ${version}, where the port is of 17.0.2`);
}

// written is main.ts's output form: printable ASCII as itself but a backslash as two, and any other code
// point as \u{HEX}, a lone surrogate as U+FFFD, which is what Node writes for one.
function written(text) {
	let result = '';
	for (const character of text) {
		const code = character.codePointAt(0);
		if (code === 0x5c) {
			result += '\\\\';
		} else if (code >= 0x20 && code <= 0x7e) {
			result += character;
		} else if (code >= 0xd800 && code <= 0xdfff) {
			result += '\\u{FFFD}';
		} else {
			result += `\\u{${code.toString(16).toUpperCase()}}`;
		}
	}
	return result;
}

// outline writes a node as main.ts does: its fields are its own properties but kind and loc, in the
// order graphql-js wrote them.
function outline(value) {
	if (value === undefined) {
		return 'undefined';
	}
	if (Array.isArray(value)) {
		return `(${value.map(outline).join(' ')})`;
	}
	switch (typeof value) {
		case 'string':
			return `"${written(value)}"`;
		case 'boolean':
			return String(value);
		case 'object': {
			const keys = Object.keys(value).filter((key) => key !== 'kind' && key !== 'loc');
			const head = `${value.kind}[${value.loc.start},${value.loc.end}]`;
			return keys.length === 0 ? head : `${head}{${keys.map((key) => `${key}=${outline(value[key])}`).join(' ')}}`;
		}
	}
	throw new Error(`a ${typeof value} in a tree`);
}

function unescape(line) {
	return line.replace(/\\(.)/g, (_, letter) => ({ '\\': '\\', t: '\t', n: '\n', r: '\r' })[letter]);
}

const lines = [];
let caseNumber = 0;
for (const line of readFileSync(casesPath, 'utf8').split('\n')) {
	if (line === '') {
		continue;
	}
	lines.push(`case ${caseNumber}`);
	caseNumber++;
	const text = unescape(line.slice(1));
	let document;
	try {
		document = parse(text, { experimentalFragmentArguments: true });
	} catch (error) {
		// Prettier's createParseError: the message, and the first location.
		const location = error.locations?.[0];
		if (location === undefined) {
			throw error;
		}
		lines.push(`error ${written(`${error.message} (${location.line}:${location.column})`)}`);
		continue;
	}
	lines.push(outline(document));
	for (let token = document.loc.startToken; token !== document.loc.endToken; token = token.next) {
		if (token.kind === 'Comment') {
			lines.push(`comment Comment[${token.start},${token.end}]{value="${written(token.value)}"}`);
		}
	}
}
process.stdout.write(lines.join('\n') + '\n');
