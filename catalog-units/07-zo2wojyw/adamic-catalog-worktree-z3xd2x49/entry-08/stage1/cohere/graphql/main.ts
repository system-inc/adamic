// The port's driver: it reads the cases file named by its one argument, parses each text in it, and
// prints the tree and the comments, or the syntax error, so its output can be held byte for byte to Go
// cohere's (graphql_test.go, whose Go side prints the Go's the same way).
//
// A cases file is one text per line, after a `>` (so that the empty text is a line too), with a
// backslash, a tab, a newline and a carriage return written as \\, \t, \n and \r.
//
// The output, for each text, is `case <number>`, then either `error <message>`, or the tree on one line
// in the outline form of cohere's parser_test.go (Kind[start,end]{key=value ...}, a list in
// parentheses, a field written as undefined as `undefined`, one never written left out) and then one
// line `comment Comment[start,end]{value="..."}` for each comment. Offsets are UTF-16 indexes. Every
// string, and every message, is in the output form lexer.ts's `escaped` writes: printable ASCII as
// itself but a backslash as two, and any other code point as \u{HEX}.
//
//	node oracle/node.mjs stage1/cohere/graphql/main.ts stage1/cohere/graphql/sample-cases.txt

import { panic, programArguments, readTextFile } from 'adamic';
import { type GraphNode, parse } from './parser.ts';

// unescapeOf is what an escape's letter stands for.
function unescapeOf(letter: string, line: string): string {
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
	panic(`an unknown escape \\${letter} in ${line}`);
}

// unescape reads a line's escapes. It splits the line at its backslashes and joins the pieces once:
// after a backslash, a piece starts with its escape's letter, and an empty piece is the first of an
// escaped backslash, after which the next piece is text. Appending each piece to the text so far, as
// this once did, copied a long line once per escape in it.
function unescape(line: string): string {
	const pieces = line.split('\\');
	const parts: string[] = [pieces[0] ?? ''];
	for (let index = 1; index < pieces.length; index++) {
		const piece = pieces[index] ?? '';
		if (piece === '') {
			parts.push('\\');
			index++;
			parts.push(pieces[index] ?? '');
			continue;
		}
		parts.push(unescapeOf(piece.slice(0, 1), line));
		parts.push(piece.slice(1));
	}
	return parts.join('');
}

// outline writes a node in the outline form into parts, and its children with it.
function outline(nodes: readonly GraphNode[], index: number, parts: string[]): void {
	const node = nodes[index] ?? panic(`no node ${index}`);
	parts.push(`${node.kind}[${node.start},${node.end}]`);
	if (node.fields.length === 0) {
		return;
	}
	parts.push('{');
	for (let position = 0; position < node.fields.length; position++) {
		const field = node.fields[position] ?? panic(`no field ${position}`);
		if (position > 0) {
			parts.push(' ');
		}
		parts.push(`${field.key}=`);
		const value = field.value;
		switch (value.kind) {
			case 'Undefined':
				parts.push('undefined');
				break;
			case 'Text':
				parts.push(`"${value.text}"`);
				break;
			case 'Flag':
				parts.push(value.flag ? 'true' : 'false');
				break;
			case 'Node':
				outline(nodes, value.node, parts);
				break;
			case 'List':
				parts.push('(');
				for (let item = 0; item < value.nodes.length; item++) {
					if (item > 0) {
						parts.push(' ');
					}
					outline(nodes, value.nodes[item] ?? panic(`no item ${item}`), parts);
				}
				parts.push(')');
				break;
		}
	}
	parts.push('}');
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file>');
const read = readTextFile(casesPath);
if (read.kind === 'Error') {
	panic(read.message);
}

let caseNumber = 0;
for (const line of read.text.split('\n')) {
	if (line === '') {
		continue;
	}
	if (!line.startsWith('>')) {
		panic(`a case that doesn't start with >: ${line}`);
	}
	console.log(`case ${caseNumber}`);
	caseNumber++;
	const result = parse(unescape(line.slice(1)));
	if (result.kind === 'Refused') {
		console.log(`error ${result.message}`);
		continue;
	}
	const document = result.document;
	const parts: string[] = [];
	outline(document.nodes, document.root, parts);
	console.log(parts.join(''));
	for (const comment of document.comments) {
		console.log(`comment Comment[${comment.start},${comment.end}]{value="${comment.value}"}`);
	}
}
