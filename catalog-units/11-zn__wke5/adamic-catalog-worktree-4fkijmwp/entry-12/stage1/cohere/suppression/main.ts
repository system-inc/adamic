// The port's driver: it reads the cases file named by its one argument, builds the suppression index
// of every source in it, asks it every question the file asks, in order, and prints what it found and
// answered, so its output can be held byte for byte to Go cohere's (suppression_test.go, whose Go side
// prints the Go's the same way).
//
// A cases file is lines of fields separated by tabs, the first naming the record:
//
//	subject <rule>            a rule that reports on directives (directives.RegisterSubject)
//	source  <text>            a source to build an index of; the records after it ask about it
//	query   <rule> <offset>   whether a finding for the rule at the offset is suppressed (Suppresses)
//	lineof  <offset>          the line an offset falls on (LineOf)
//
// A field writes a backslash, a tab, a newline and a carriage return as \\, \t, \n and \r. Offsets are
// UTF-16 indexes.
//
// The output, for each source, is `case <number>`; then each comment the scanner found, `comment <pos>
// <end> <scope word>`, with `-` for a comment that is no directive (directives.Recognize); each
// directive, `directive <index> <kind> rules=<names> reason=<reason> line=<line> start=<start line>
// end=<end line> pos=<pos> end=<end>`; each rule reference, `reference <name> <pos> <end>`; each answer,
// `query <rule> <offset> <1 or 0>` and `lineof <offset> <line>`, in the order asked; then `applied`
// with each directive's count, `total`, and the indexes of the `unused` directives and of those
// `without-reason`. Names and reasons are quoted with \\, \", \n, \r, \t, and \u00XX for any other
// control character.
//
//	node oracle/node.mjs stage1/cohere/suppression/main.ts stage1/cohere/suppression/sample-cases.txt

import { panic, programArguments, readTextFile } from 'adamic';
import { recognize, registerSubject } from './directives.ts';
import { scanComments } from './scan.ts';
import { build, type Directive, type Index } from './suppression.ts';

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

// unescape reads a field's escapes. It copies the runs between them whole, since a source is long and
// appending it a character at a time copied it once per character.
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

// indexesOf is the indexes, in all, of the directives in some, joined by spaces.
function indexesOf(all: readonly Directive[], some: readonly Directive[]): string {
	const indexes: string[] = [];
	for (const directive of some) {
		indexes.push(`${all.indexOf(directive)}`);
	}
	return indexes.join(' ');
}

// describe prints what an index found, before it is asked anything.
function describe(source: string, index: Index): void {
	for (const comment of scanComments(source)) {
		const scopeWord = recognize(source.slice(comment.pos, comment.end));
		console.log(`comment ${comment.pos} ${comment.end} ${scopeWord === '' ? '-' : scopeWord}`);
	}
	const directives = index.directives();
	for (let position = 0; position < directives.length; position++) {
		const directive = directives[position] ?? panic(`no directive ${position}`);
		const rules = directive.rules.map((rule) => quote(rule)).join(',');
		console.log(
			`directive ${position} ${directive.kind} rules=${rules} reason=${quote(directive.reason)} line=${directive.line} start=${directive.startLine} end=${directive.endLine} pos=${directive.pos} end=${directive.end}`,
		);
	}
	for (const reference of index.ruleReferences()) {
		console.log(`reference ${quote(reference.name)} ${reference.pos} ${reference.end}`);
	}
}

// summarize prints what an index's directives did, once it has been asked everything.
function summarize(index: Index): void {
	const directives = index.directives();
	const counts: string[] = [];
	for (let position = 0; position < directives.length; position++) {
		counts.push(`${index.appliedCount(position)}`);
	}
	console.log(`applied ${counts.join(' ')}`);
	console.log(`total ${index.totalApplied()}`);
	console.log(`unused ${indexesOf(directives, index.unused())}`);
	console.log(`without-reason ${indexesOf(directives, index.withoutReason())}`);
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file>');
const read = readTextFile(casesPath);
if (read.kind === 'Error') {
	panic(read.message);
}

let current: Index | undefined = undefined;
let caseNumber = 0;
for (const line of read.text.split('\n')) {
	if (line === '') {
		continue;
	}
	const fields = line.split('\t');
	const record = fields[0] ?? '';
	switch (record) {
		case 'subject':
			registerSubject(unescape(fields[1] ?? panic(`a subject without a rule: ${line}`)));
			break;
		case 'source': {
			if (current !== undefined) {
				summarize(current);
			}
			const source = unescape(fields[1] ?? panic(`a source without text: ${line}`));
			console.log(`case ${caseNumber}`);
			caseNumber++;
			current = build(source);
			describe(source, current);
			break;
		}
		case 'query': {
			const index = current ?? panic('a query before any source');
			const rule = unescape(fields[1] ?? panic(`a query without a rule: ${line}`));
			const offset = Number.parseInt(fields[2] ?? panic(`a query without an offset: ${line}`), 10);
			console.log(`query ${rule} ${offset} ${index.suppresses(rule, offset) ? '1' : '0'}`);
			break;
		}
		case 'lineof': {
			const index = current ?? panic('a lineof before any source');
			const offset = Number.parseInt(fields[1] ?? panic(`a lineof without an offset: ${line}`), 10);
			console.log(`lineof ${offset} ${index.lineOf(offset)}`);
			break;
		}
		default:
			panic(`an unknown record: ${line}`);
	}
}
if (current !== undefined) {
	summarize(current);
}
