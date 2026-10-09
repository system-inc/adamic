// tokenizer: a lexer over a large made-up source text, walking it a UTF-16 code unit at a time with
// charCodeAt, the way a compiler's scanner does. The text has some non-ASCII in it, as real source
// does in comments and strings.

const words = ['let', 'value', 'count', 'résumé', 'naïve', 'index', 'total', 'mapping', 'the', 'λ', 'größe', 'reduce'];
const operators = ['+', '-', '*', '/', '=', '==', '(', ')', '{', '}', ';', ','];

// A linear congruential generator, so every runtime makes the same text.
let seed = 42;
function next(limit: number): number {
	seed = (seed * 1103515245 + 12345) % 2147483648;
	return seed % limit;
}

const pieces: string[] = [];
for (let index = 0; index < 600000; index += 1) {
	const kind = next(4);
	if (kind === 0) {
		pieces.push(`${next(100000)}`);
	} else if (kind === 1) {
		pieces.push(operators[next(operators.length)] ?? '+');
	} else {
		pieces.push(words[next(words.length)] ?? 'x');
	}
	pieces.push(next(8) === 0 ? '\n' : ' ');
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
	} else if (code === 32) {
		position += 1;
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
console.log(`${length} code units, ${lines} lines: ${identifiers} identifiers, ${numbers} numbers summing to ${sum}, ${symbols} symbols, longest ${longest}`);
