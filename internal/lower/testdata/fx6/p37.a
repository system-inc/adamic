// A valid program: an unrelated, never viewed object has a string | number field whose name a
// view checks elsewhere.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly name: string; }
interface Other { readonly name: string | number; }
function show(holder: Other): string {
	const name = holder.name;
	return typeof name === 'string' ? `string ${name}` : `number ${name}`;
}
const raw = { kind: 'Holder' as const, name: 'abc' };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.name.length}`);
console.log(show({ name: 5 }));
