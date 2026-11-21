// #s03ndgw: an object with no literal tag, so its every field is matched, a string static among them. The
// view's field is Labeled itself, so nothing narrows between the read and its use.
interface Root { readonly kind: 'Holder'; }
interface Labeled { readonly label: string; readonly count: number; }
interface Viewed extends Root { readonly value: Labeled; }
class Klass {
	static readonly label = 'hello';
	static readonly count = 7;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log(viewed.value.label + ' ' + String(viewed.value.count));
