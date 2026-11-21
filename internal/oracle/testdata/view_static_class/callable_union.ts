// #s03ndgw, a stop: a callable could stand in the union too, so a class constructor, a function, keeps its stop.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface View extends Root { readonly value: number | Inner | (() => number); }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
console.log(typeof read === 'number' ? 'number' : typeof read);
