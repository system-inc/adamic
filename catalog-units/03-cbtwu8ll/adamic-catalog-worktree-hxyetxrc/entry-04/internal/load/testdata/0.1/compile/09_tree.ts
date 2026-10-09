interface Tree {
	readonly value: number;
	readonly left: Tree | undefined;
	readonly right: Tree | undefined;
}

function insert(tree: Tree | undefined, value: number): Tree {
	if (tree === undefined) {
		return { value, left: undefined, right: undefined };
	}
	if (value < tree.value) {
		return { ...tree, left: insert(tree.left, value) };
	}
	if (value > tree.value) {
		return { ...tree, right: insert(tree.right, value) };
	}
	return tree;
}

function collect(tree: Tree | undefined, into: number[]): void {
	if (tree !== undefined) {
		collect(tree.left, into);
		into.push(tree.value);
		collect(tree.right, into);
	}
}

function height(tree: Tree | undefined): number {
	return tree === undefined ? 0 : 1 + Math.max(height(tree.left), height(tree.right));
}

let tree: Tree | undefined = undefined;
for (const value of [50, 30, 70, 20, 40, 60, 80, 30, 65]) {
	tree = insert(tree, value);
}
const before = tree;
tree = insert(tree, 10);

const values: number[] = [];
collect(tree, values);
console.log(values.join(' '));
console.log(`height ${height(tree)}, was ${height(before)}`);
console.log(`right subtree shared: ${tree.right === before?.right}`);
