interface TreeNode { name: string; parent: TreeNode | undefined; }
function adopt(child: TreeNode, parent: TreeNode): void { child.parent = parent; }
function path(node: TreeNode): string { const up = node.parent; return up === undefined ? node.name : `${path(up)}/${node.name}`; }
const root: TreeNode = { name: 'root', parent: undefined };
const leaf: TreeNode = { name: 'leaf', parent: undefined };
adopt(leaf, root);
console.log(path(leaf));
leaf.parent = undefined;
console.log(`${path(leaf)} ${leaf.parent === undefined}`);
