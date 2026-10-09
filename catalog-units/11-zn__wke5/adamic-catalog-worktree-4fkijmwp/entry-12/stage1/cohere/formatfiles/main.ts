// The port's driver: it reads the cases file named by its one argument, walks every tree in it as
// formatfiles.Enumerate and NestedRepositoriesBelow do, asks the questions after it, and prints what it
// found, so its output can be held byte for byte to Go cohere's (formatfiles_test.go, whose Go side
// walks the same trees on disk and prints the same way).
//
// A cases file is lines of fields separated by tabs, the first naming the record:
//
//	handles  <extension>...                        the extensions the formatter takes, lowercased
//	tree     <name> <root>                         a tree on disk, by its absolute root
//	house    <declared 0 or 1> <source> <line>...  formatoptions.Resolve's answer for the tree; walks it
//	                                               (its ignorePatterns are none: the Go side keeps them so)
//	has      <directory>                           HasOwnRepository
//	contains <file>                                NestedRepositoryContaining, from the tree's root
//	boundary <root> <settings directory>           repositoryBoundaryBetween
//
// A field writes a backslash, a tab, a newline and a carriage return as \\, \t, \n and \r.
//
// The output, for each tree, is `tree <name>`; then Enumerate's answer, `enumerate error <message>`, or
// its counts and lists: `walked`, each `layer <name> <count>` and `declined <extension> <count>` in
// Go's order (by code point), `unhandled`, `symbolic-links`, and each `nested`, `directory`,
// `ignore-file`, `file` and `adamic` in the order the walk found them; then NestedRepositoriesBelow's, each
// `nested-below <path>` or `nested-below error <message>`; and an answer for each question, in order.
// Names and paths are quoted with \\, \", \n, \r, \t, and \u00XX for any other control character.
//
//	node oracle/node.mjs stage1/cohere/formatfiles/main.ts <cases file>

import { panic, programArguments, readTextFile } from 'adamic';
import {
	enumerate,
	hasOwnRepository,
	nestedRepositoriesBelow,
	nestedRepositoryContaining,
	repositoryBoundaryBetween,
	type Resolution,
} from './enumerate.ts';
import { compareCodePoints, ext, goToLower } from './golang.ts';

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

// needsEscape is whether any character of text is one escapeOf writes otherwise: a control character,
// DEL, a quote or a backslash. One pass over the units, which a path, the usual text here, passes.
function needsEscape(text: string): boolean {
	for (let index = 0; index < text.length; index++) {
		const unit = text.charCodeAt(index);
		if (unit < 0x20 || unit === 0x7f || unit === 0x22 || unit === 0x5c) {
			return true;
		}
	}
	return false;
}

// quote is a string in the output's quotes. A string with nothing to escape is written whole; otherwise
// the runs between escapes are copied whole.
function quote(text: string): string {
	if (!needsEscape(text)) {
		return `"${text}"`;
	}
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

// unescapeOf is what an escape's letter stands for.
function unescapeOf(letter: string, field: string): string {
	switch (letter) {
		case '\\':
			return '\\';
		case 't':
			return '\t';
		case 'n':
			return '\n';
		case 'r':
			return '\r';
	}
	// return panic(...), but stage 0 doesn't lower that yet (gitignore's GAPS.md, gap 9).
	panic(`an unknown escape \\${letter} in ${field}`);
}

// unescape reads a field's escapes, copying the runs between them whole.
function unescape(field: string): string {
	let text = '';
	let runStart = 0;
	for (let index = 0; index < field.length; index++) {
		if (field.charCodeAt(index) === 0x5c) {
			text += field.slice(runStart, index) + unescapeOf(field.slice(index + 1, index + 2), field);
			index++;
			runStart = index + 1;
		}
	}
	return text + field.slice(runStart);
}

// field is a record's field at index, unescaped.
function field(fields: readonly string[], index: number, line: string): string {
	return unescape(fields[index] ?? panic(`a record without field ${index}: ${line}`));
}

// sortedKeys is a Map's keys in Go's order.
function sortedKeys(counts: ReadonlyMap<string, number>): string[] {
	return [...counts.keys()].sort(compareCodePoints);
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file>');
const read = readTextFile(casesPath);
if (read.kind === 'Error') {
	panic(read.message);
}

let handled: readonly string[] = [];
let root = '';
for (const line of read.text.split('\n')) {
	if (line === '') {
		continue;
	}
	const fields = line.split('\t');
	switch (fields[0] ?? '') {
		case 'handles':
			handled = fields.slice(1).map((extension) => unescape(extension));
			break;
		case 'tree':
			console.log(`tree ${field(fields, 1, line)}`);
			root = field(fields, 2, line);
			break;
		case 'house': {
			const extensions = handled;
			const resolution: Resolution = {
				houseIgnoreDeclared: field(fields, 1, line) === '1',
				source: field(fields, 2, line),
				houseIgnore: fields.slice(3).map((houseLine) => unescape(houseLine)),
				ignorePatterns: [],
			};
			const enumerated = enumerate(root, resolution, (fileName) => extensions.includes(goToLower(ext(fileName))));
			if (enumerated.kind === 'Error') {
				console.log(`enumerate error ${enumerated.message}`);
			} else {
				const enumeration = enumerated.enumeration;
				console.log(`walked ${enumeration.walked}`);
				for (const name of sortedKeys(enumeration.ignoredByLayer)) {
					console.log(`layer ${quote(name)} ${enumeration.ignoredByLayer.get(name) ?? 0}`);
				}
				console.log(`unhandled ${enumeration.unhandled}`);
				console.log(`symbolic-links ${enumeration.symbolicLinks}`);
				for (const extension of sortedKeys(enumeration.declinedExtensions)) {
					console.log(`declined ${quote(extension)} ${enumeration.declinedExtensions.get(extension) ?? 0}`);
				}
				for (const nested of enumeration.nestedRepositories) {
					console.log(`nested ${quote(nested)}`);
				}
				for (const directory of enumeration.directories) {
					console.log(`directory ${quote(directory)}`);
				}
				for (const ignoreFile of enumeration.ignoreFiles) {
					console.log(`ignore-file ${quote(ignoreFile)}`);
				}
				for (const file of enumeration.files) {
					console.log(`file ${quote(file)}`);
				}
				for (const file of enumeration.adamic) {
					console.log(`adamic ${quote(file)}`);
				}
			}
			const below = nestedRepositoriesBelow(root);
			if (below.kind === 'Error') {
				console.log(`nested-below error ${below.message}`);
			} else {
				for (const nested of below.nested) {
					console.log(`nested-below ${quote(nested)}`);
				}
			}
			break;
		}
		case 'has': {
			const directory = field(fields, 1, line);
			console.log(`has ${quote(directory)} ${hasOwnRepository(directory) ? '1' : '0'}`);
			break;
		}
		case 'contains': {
			const file = field(fields, 1, line);
			console.log(`contains ${quote(file)} ${quote(nestedRepositoryContaining(root, file))}`);
			break;
		}
		case 'boundary': {
			const from = field(fields, 1, line);
			const to = field(fields, 2, line);
			console.log(`boundary ${quote(from)} ${quote(to)} ${repositoryBoundaryBetween(from, to) ? '1' : '0'}`);
			break;
		}
		default:
			panic(`an unknown record: ${line}`);
	}
}
