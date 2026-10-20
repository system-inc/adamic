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
const kinds = new Map<string, Kind.Text>();
kinds.set('a', raw);
const kind = kinds.get('a');
if (kind !== undefined) {
	console.log(show({ kind, text: 'ahra' }));
}
