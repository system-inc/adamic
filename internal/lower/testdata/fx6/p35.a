// A valid program: a view reads its string field, then the producer writes a number there and
// reads it back through its own honest type. The view is never read again.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly name: string; }
interface Honest { readonly kind: 'Holder'; name: string | number; }
function change(holder: Honest): void {
	holder.name = 5;
}
function show(holder: Honest): string {
	const name = holder.name;
	return typeof name === 'string' ? `string ${name}` : `number ${name}`;
}
const raw: Honest = { kind: 'Holder', name: 'abc' };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.name.length}`);
change(raw);
console.log(show(raw));
