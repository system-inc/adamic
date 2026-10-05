// Gap 4: a tree with mutable child arrays is refused as cycle-capable. Any node's children could be
// given the node itself (root.nodes.push(root)), a cycle reference counting can't free.
class TreeNode {
	readonly name: string;
	readonly nodes: TreeNode[] = [];
	constructor(name: string) {
		this.name = name;
	}
}
const root = new TreeNode('root');
root.nodes.push(new TreeNode('a'));
root.nodes.push(new TreeNode('b'));
console.log(`${root.nodes.length}`);
