// A valid program: an unrelated object, never viewed, has a field with a checked name and
// changes its representation.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly name: string; }
interface Other { name: string | number; }
function change(holder: Other): void {
	holder.name = 5;
}
function show(holder: Other): string {
	const name = holder.name;
	return typeof name === 'string' ? `string ${name}` : `number ${name}`;
}
const raw = { kind: 'Holder' as const, name: 'abc' };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.name.length}`);
const other: Other = { name: 'x' };
change(other);
console.log(show(other));
