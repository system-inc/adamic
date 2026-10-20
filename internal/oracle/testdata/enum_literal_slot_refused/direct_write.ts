// a-check: refused an unproven value assigned to a numeric literal or enum member slot Kind.Text
// After narrowing, a number written into a singleton enum-literal tag; a later comparison then proves the wrong layout.
enum Kind {
	Text,
	Count,
}
type Node = { kind: Kind.Text; text: string } | { kind: Kind.Count; count: number };
function retag(node: Node, value: number): void {
	if (node.kind === Kind.Text) {
		node.kind = value;
	}
}
function show(node: Node): string {
	if (node.kind === Kind.Count) {
		return `${node.count + 1}`;
	}
	return node.text.toUpperCase();
}
const node: Node = { kind: Kind.Text, text: 'ahra' };
retag(node, 1);
console.log(show(node));
