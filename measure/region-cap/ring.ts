// ring: graph.ts's nodes as the brief first drew them, a ring of a million with back-edges written in
// after the nodes are made, built in a function and summed, twice. Built only for its count line: its
// nodes aren't a region's, which is why graph.ts is a fresh binary graph instead.
interface RingNode {
	value: number;
	next: RingNode | undefined;
	back: RingNode | undefined;
}

function ring(seed: number): number {
	const count = 1000000;
	const nodes: RingNode[] = [];
	for (let index = 0; index < count; index += 1) {
		nodes.push({ value: (seed + index) % 1000, next: undefined, back: undefined });
	}
	for (let index = 0; index < count; index += 1) {
		const node = nodes[index];
		if (node !== undefined) {
			node.next = nodes[(index + 1) % count];
			node.back = nodes[(index + count - 7) % count];
		}
	}
	let total = 0;
	let at = nodes[0];
	for (let step = 0; step < count; step += 1) {
		if (at === undefined) {
			break;
		}
		total += at.value;
		const back = at.back;
		if (back !== undefined) {
			total += back.value;
		}
		at = at.next;
	}
	return total;
}

let checksum = 0;
for (let round = 0; round < 2; round += 1) {
	checksum += ring(round);
}
console.log(`ring checksum ${checksum}`);
