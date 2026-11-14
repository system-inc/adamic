// array_copy: copy.ts's copy as the brief first drew it, into a fresh array, fifty times. Built only
// for its count line: an array's elements aren't a region's, which is why copy.ts copies into a tree.
interface Item {
	readonly x: number;
	readonly y: number;
}

const empty: Item = { x: 0, y: 0 };

function copyArray(source: readonly Item[]): number {
	const copied: Item[] = [];
	for (let index = 0; index < source.length; index += 1) {
		const item = source[index] ?? empty;
		copied.push({ x: item.x, y: item.y });
	}
	let total = 0;
	for (const item of copied) {
		total += item.x * 3 + item.y;
	}
	return total;
}

const source: Item[] = [];
for (let index = 0; index < 200000; index += 1) {
	source.push({ x: index % 1000, y: (index * 7) % 13 });
}
let total = 0;
for (let round = 0; round < 50; round += 1) {
	total += copyArray(source);
}
console.log(`array copy checksum ${total}`);
