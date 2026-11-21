// #s03ndgw: an object with no literal tag, so its every field is matched, a string static among them.
interface Root { readonly kind: 'Holder'; }
interface Labeled { readonly label: string; readonly count: number; }
interface View extends Root { readonly value: number | Labeled; }
class Klass {
	static readonly label = 'hello';
	static readonly count = 7;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
const read = viewed.value;
if (typeof read !== 'number') {
	console.log(read.label + ' ' + String(read.count));
}
