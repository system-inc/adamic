// tokenizer: a lexer over a large made-up source text, walking it a UTF-16 code unit at a time with
// charCodeAt, the way a compiler's scanner does. The text has some non-ASCII in it, as real source
// does in identifiers and strings.

const words = ['let', 'value', 'count', 'résumé', 'naïve', 'index', 'total', 'mapping', 'the', 'λ', 'größe', 'reduce'];
const operators = ['+', '-', '*', '/', '=', '==', '(', ')', '{', '}', ';', ','];
const literals = ['"hello world"', "'résumé'", '"line\\nnext"', '"quote: \\"ok\\""', "'λ größe'"];

// Park-Miller: the largest product is below 2^46, hence exact in a JS number.
// The old multiplier exceeded 2^53 and lost the low bits used to choose tokens.
let seed = 42;
function next(limit: number): number {
	seed = (seed * 16807) % 2147483647;
	return seed % limit;
}

const pieces: string[] = [];
for (let index = 0; index < 600000; index += 1) {
	// About 3/8 identifiers, 1/4 numbers, 1/4 punctuation and 1/8 strings.
	const kind = next(8);
	if (kind < 2) {
		pieces.push(`${next(100000)}`);
	} else if (kind < 4) {
		pieces.push(operators[next(operators.length)] ?? '+');
	} else if (kind === 4) {
		pieces.push(literals[next(literals.length)] ?? '"text"');
	} else {
		pieces.push(words[next(words.length)] ?? 'x');
	}
	const space = next(16);
	pieces.push(space === 0 ? '\n' : space === 1 ? '\t' : space === 2 ? '  ' : ' ');
}
const text = pieces.join('');

function isLetter(code: number): boolean {
	return (code >= 97 && code <= 122) || (code >= 65 && code <= 90) || code === 95 || code > 127;
}

function isDigit(code: number): boolean {
	return code >= 48 && code <= 57;
}

let identifiers = 0;
let numbers = 0;
let strings = 0;
let symbols = 0;
let lines = 1;
let longest = '';
let sum = 0;
let position = 0;
const length = text.length;
while (position < length) {
	const code = text.charCodeAt(position);
	if (code === 10) {
		lines += 1;
		position += 1;
	} else if (code === 32 || code === 9) {
		position += 1;
	} else if (code === 34 || code === 39) {
		const quote = code;
		position += 1;
		while (position < length && text.charCodeAt(position) !== quote) {
			// Escaped quotes and backslashes belong to the string token.
			position += text.charCodeAt(position) === 92 ? 2 : 1;
		}
		position += 1;
		strings += 1;
	} else if (isLetter(code)) {
		const start = position;
		while (position < length && (isLetter(text.charCodeAt(position)) || isDigit(text.charCodeAt(position)))) {
			position += 1;
		}
		const name = text.slice(start, position);
		identifiers += 1;
		if (name.length > longest.length) {
			longest = name;
		}
	} else if (isDigit(code)) {
		let value = 0;
		while (position < length && isDigit(text.charCodeAt(position))) {
			value = value * 10 + (text.charCodeAt(position) - 48);
			position += 1;
		}
		numbers += 1;
		sum += value;
	} else {
		symbols += 1;
		position += 1;
	}
}
console.log(`${length} code units, ${lines} lines: ${identifiers} identifiers, ${numbers} numbers summing to ${sum}, ${strings} strings, ${symbols} symbols, longest ${longest}`);
