// #s03ndgw: a read of Inner's fields through the view reads the class's statics, as Node reads Klass.count
// and Klass.tag. The view's field is Inner itself, so nothing narrows between the read and its use.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log(String(viewed.value.count));
console.log(viewed.value.tag);
