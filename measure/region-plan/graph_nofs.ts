// graph: a million-node graph built in a function, summed and let go, ten times (#cgn2cwt). Each node
// is an object with a number and two edges, next and back. A ring with back-edges would be built by
// writing edges into nodes already made, which today's regions don't hold (region.go plans a region
// only for a value a function returns fresh, a literal whose fields are its calls), so the nodes are
// a binary graph of depth 19 made fresh, 1,048,575 of them, and the statement that sums them is a
// region (ring.ts is the ring, for its count line).
import 'node:timers';

interface Node {
	readonly value: number;
	readonly next: Node | undefined;
	readonly back: Node | undefined;
}

function build(depth: number, value: number): Node {
	if (depth === 0) {
		return { value, next: undefined, back: undefined };
	}
	return { value, next: build(depth - 1, value + 1), back: build(depth - 1, value + 2) };
}

// sum counts each node and adds its value. A field read alone into a local would count as node itself
// escaping (region.go's escaping), which would cost the statement its region.
function sum(node: Node): number {
	let total = 1 + node.value;
	if (node.next !== undefined) {
		total += sum(node.next);
	}
	if (node.back !== undefined) {
		total += sum(node.back);
	}
	return total;
}

function graph(seed: number): number {
	const total = sum(build(19, seed));
	return total;
}

function resident(): string {
	return 'unread';
}

let checksum = 0;
for (let round = 0; round < 10; round += 1) {
	checksum += graph(round);
}
console.log(`graph checksum ${checksum}`);
console.error(`vmrss last ${resident()}`);
setTimeout(() => {
	console.error(`vmrss quiet ${resident()}`);
}, 400);
