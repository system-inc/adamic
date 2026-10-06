// parallel_files: cohere's shape without filesystem timing. Generate 4,096 source files,
// tokenize and summarize each independently, then merge summaries in input order.
import { parallelMap } from 'adamic';

interface Summary {
	readonly tokens: number;
	readonly distinct: number;
	readonly units: number;
	readonly digest: number;
}

function wordUnit(code: number): boolean {
	return (code >= 48 && code <= 57) || (code >= 65 && code <= 90) || (code >= 97 && code <= 122) || code === 95 || code > 127;
}

function tokenize(text: string): readonly string[] {
	const tokens: string[] = [];
	let position = 0;
	while (position < text.length) {
		if (!wordUnit(text.charCodeAt(position))) {
			position += 1;
		} else {
			const start = position;
			while (position < text.length && wordUnit(text.charCodeAt(position))) { position += 1; }
			tokens.push(text.slice(start, position));
		}
	}
	return tokens;
}

function summarize(text: string, index: number): Summary {
	const tokens = tokenize(text);
	const frequencies = new Map<string, number>();
	let digest = index;
	for (const token of tokens) {
		frequencies.set(token, (frequencies.get(token) ?? 0) + 1);
		for (let unit = 0; unit < token.length; unit += 1) {
			digest = (digest * 31 + token.charCodeAt(unit)) % 1000000007;
		}
	}
	return { tokens: tokens.length, distinct: frequencies.size, units: text.length, digest };
}

const generated: string[] = [];
for (let file = 0; file < 4096; file += 1) {
	const lines: string[] = [];
	for (let line = 0; line < 64; line += 1) {
		const symbol = (file * 37 + line * 17) % 997;
		lines.push(`export const value_${symbol} = source_${file % 101} + ${line}; // résumé λ: deterministic file ${file}\n`);
	}
	generated.push(lines.join(''));
}
const files: readonly string[] = generated;
const summaries = parallelMap(files, summarize);
let tokens = 0;
let distinct = 0;
let units = 0;
let digest = 0;
for (const summary of summaries) {
	tokens += summary.tokens;
	distinct += summary.distinct;
	units += summary.units;
	digest = (digest * 131 + summary.digest) % 1000000007;
}
console.log(`${summaries.length} files, ${units} UTF-16 units, ${tokens} tokens, ${distinct} distinct per file, ordered digest ${digest}`);
