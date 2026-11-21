// #s03ndgw, a stop pinned for its own sake (Language's attack): Klass declares count wider than Inner's
// field, and holds a string there, which Node reads as 'x'. Lowering stops the static today, as one without
// a native slot. When wide statics get native slots, this road must still keep a stop, the view's check,
// and never read the string as a number.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static count: number | string = 'x';
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.count));
