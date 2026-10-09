// trees: the Computer Language Benchmarks Game's binary-trees, without its threads. Allocating
// millions of small objects, walking them, and letting them go: with no garbage collector,
// reference counting does all the freeing.

interface Tree {
	readonly left: Tree | undefined;
	readonly right: Tree | undefined;
}

function build(depth: number): Tree {
	if (depth === 0) {
		return { left: undefined, right: undefined };
	}
	return { left: build(depth - 1), right: build(depth - 1) };
}

function check(tree: Tree): number {
	if (tree.left === undefined || tree.right === undefined) {
		return 1;
	}
	return 1 + check(tree.left) + check(tree.right);
}

const maximumDepth = 18;
const stretch = build(maximumDepth + 1);
console.log(`stretch tree of depth ${maximumDepth + 1}\t check: ${check(stretch)}`);
const longLived = build(maximumDepth);
for (let depth = 4; depth <= maximumDepth; depth += 2) {
	const iterations = 2 ** (maximumDepth - depth + 4);
	let total = 0;
	for (let index = 0; index < iterations; index += 1) {
		total += check(build(depth));
	}
	console.log(`${iterations}\t trees of depth ${depth}\t check: ${total}`);
}
console.log(`long lived tree of depth ${maximumDepth}\t check: ${check(longLived)}`);
