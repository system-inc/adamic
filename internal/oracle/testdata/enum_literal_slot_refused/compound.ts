// a-check: refused an unproven value assigned to a numeric literal or enum member slot Kind.Text
enum Kind {
	Text,
	Count,
}
type Node = { kind: Kind.Text; text: string } | { kind: Kind.Count; count: number };
function show(node: Node): string {
	if (node.kind === Kind.Count) {
		return `${node.count + 1}`;
	}
	return node.text.toUpperCase();
}
const raw: number = 1;
const flag: boolean = true;
const node: Node = { kind: Kind.Text, text: 'ahra' };
if (node.kind === Kind.Text) {
	node.kind += 1;
}
console.log(show(node));
