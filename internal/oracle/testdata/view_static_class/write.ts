// #s03ndgw, a stop: the view reads a class's statics and never writes them, so a write through it stops.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
viewed.value.count = 4;
console.log(String(Klass.count));
