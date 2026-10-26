// A valid program: an unrelated, never viewed object has a field of an object union type whose
// name a view checks elsewhere.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly name: string; }
interface Other { readonly name: { readonly a: number } | { readonly b: string }; }
function show(holder: Other): string {
	return `${holder.name !== undefined}`;
}
const raw = { kind: 'Holder' as const, name: 'abc' };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.name.length}`);
console.log(show({ name: { a: 1 } }));
