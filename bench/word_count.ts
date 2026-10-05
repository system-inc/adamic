// word_count: words counted in a Map<string, number>, then the most frequent sorted out. Hashing
// strings, map lookups and updates, and the strings split out of a large text.

let seed = 7;
function next(limit: number): number {
	seed = (seed * 1103515245 + 12345) % 2147483648;
	return seed % limit;
}

const syllables = ['ka', 'ri', 'to', 'men', 'su', 'lo', 'vi', 'na', 'ze', 'qu'];
const vocabulary: string[] = [];
for (let index = 0; index < 20000; index += 1) {
	let word = '';
	const count = 1 + next(4);
	for (let syllable = 0; syllable < count; syllable += 1) {
		word += syllables[next(syllables.length)] ?? 'a';
	}
	vocabulary.push(word);
}

const chosen: string[] = [];
for (let index = 0; index < 1000000; index += 1) {
	// Zipf-like: small indexes far more often than large ones.
	const rank = next(next(vocabulary.length) + 1);
	chosen.push(vocabulary[rank] ?? 'a');
}
const text = chosen.join(' ');

const counts = new Map<string, number>();
for (const word of text.split(' ')) {
	counts.set(word, (counts.get(word) ?? 0) + 1);
}
const entries: [string, number][] = [...counts];
entries.sort((left, right) => (right[1] !== left[1] ? right[1] - left[1] : left[0] < right[0] ? -1 : left[0] > right[0] ? 1 : 0));
console.log(`${counts.size} distinct words`);
for (const [word, count] of entries.slice(0, 5)) {
	console.log(`${word} ${count}`);
}
