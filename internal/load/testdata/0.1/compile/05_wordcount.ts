type Entry = readonly [string, number];

function compareEntries(left: Entry, right: Entry): number {
	if (left[1] !== right[1]) {
		return right[1] - left[1];
	}
	return left[0] < right[0] ? -1 : left[0] > right[0] ? 1 : 0;
}

const text = 'the quick brown fox jumps over the lazy dog and the fox naps';
const counts = new Map<string, number>();
for (const word of text.split(' ')) {
	counts.set(word, (counts.get(word) ?? 0) + 1);
}

const entries: Entry[] = [...counts];
entries.sort(compareEntries);
for (const [word, count] of entries.slice(0, 4)) {
	console.log(`${word.padEnd(6, '.')}${count}`);
}
console.log(`${counts.size} distinct words`);
