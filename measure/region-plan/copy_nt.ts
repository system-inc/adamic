// copy_nt (copy.ts without its timer, #ahnc64w): a 200,000-element array of small objects copied element by element as new objects, fifty
// times, each copy checksummed and let go (#cgn2cwt). Today's regions hold no array (an array's
// elements grow by realloc, docs/memory.md), so each copy is a balanced tree of the new objects in
// the array's order, made fresh, and the statement that checksums it is a region (array_copy.ts is
// the copy into a fresh array, for its count line).
import { readFileSync } from 'node:fs';

interface Item {
	readonly x: number;
	readonly y: number;
}

interface Copy {
	readonly x: number;
	readonly y: number;
	readonly left: Copy | undefined;
	readonly right: Copy | undefined;
}

const empty: Item = { x: 0, y: 0 };

// copyRange copies count elements from start: the first here, then the rest halved between two subtrees.
function copyRange(source: readonly Item[], start: number, count: number): Copy {
	const item = source[start] ?? empty;
	if (count === 1) {
		return { x: item.x, y: item.y, left: undefined, right: undefined };
	}
	if (count === 2) {
		return { x: item.x, y: item.y, left: copyRange(source, start + 1, 1), right: undefined };
	}
	const half = Math.floor((count - 1) / 2);
	return { x: item.x, y: item.y, left: copyRange(source, start + 1, half), right: copyRange(source, start + 1 + half, count - 1 - half) };
}

function checksum(copy: Copy): number {
	let total = copy.x * 3 + copy.y;
	if (copy.left !== undefined) {
		total += checksum(copy.left);
	}
	if (copy.right !== undefined) {
		total += checksum(copy.right);
	}
	return total;
}

function copyAll(source: readonly Item[]): number {
	const total = checksum(copyRange(source, 0, source.length));
	return total;
}

function resident(): string {
	for (const line of readFileSync('/proc/self/status', 'utf8').split('\n')) {
		if (line.startsWith('VmRSS:')) {
			return line.slice(6).trim();
		}
	}
	return 'unknown';
}

const source: Item[] = [];
for (let index = 0; index < 200000; index += 1) {
	source.push({ x: index % 1000, y: (index * 7) % 13 });
}
let total = 0;
for (let round = 0; round < 50; round += 1) {
	total += copyAll(source);
}
console.log(`copy checksum ${total}`);
console.error(`vmrss last ${resident()}`);
