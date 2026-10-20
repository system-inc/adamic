// a-check: refused an unproven value assigned to a numeric literal or enum member slot Kind.Text
// A class whose tag is a getter: comparing it proves nothing about the fields.
enum Kind {
	Text,
	Count,
}
class TextNode {
	readonly text: string = 'ahra';
	private reads = 0;
	private readonly later: number = 1;
	get kind(): Kind.Text {
		this.reads += 1;
		return this.reads > 1 ? this.later : Kind.Text;
	}
}
class CountNode {
	readonly kind = Kind.Count;
	readonly count: number = 1;
}
function show(node: TextNode | CountNode): string {
	if (node.kind === Kind.Text) {
		return 'text';
	}
	return `${node.count + 1}`;
}
const text = new TextNode();
console.log(text.kind === Kind.Text ? 'first' : 'other');
console.log(show(text));
