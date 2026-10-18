interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: boolean | undefined | string; }
function pick(n: number): boolean | undefined {
	return n > 10 ? true : undefined;
}
const source = { kind: 'Holder' as const, value: 1234.5678 };
const raw = { ...source, value: pick(source.value > 0 ? 1 : 20) };
const base: Root = raw;
const viewed = base as View;
const read = viewed.value;
console.log(`${read === undefined}`);
console.log(typeof read);
