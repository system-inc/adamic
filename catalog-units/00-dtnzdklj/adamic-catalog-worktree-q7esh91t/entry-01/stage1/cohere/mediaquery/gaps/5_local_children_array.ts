// Gap 5: the same rule refuses a tree built bottom-up, its children collected in a local mutable
// array that is then handed to a readonly field and never written again. The finder reads types, so
// the local TreeNode[] is cycle-capable whatever happens to it after.
class TreeNode {
	readonly name: string;
	readonly nodes: readonly TreeNode[];
	constructor(name: string, nodes: readonly TreeNode[]) {
		this.name = name;
		this.nodes = nodes;
	}
}
function build(): TreeNode {
	const children: TreeNode[] = [];
	children.push(new TreeNode('a', []));
	children.push(new TreeNode('b', []));
	return new TreeNode('root', children);
}
console.log(`${build().nodes.length}`);
