// #s03ndgw: an optional field no class on Klass's chain declares, and no class in the program names as a static
// method, is absent, and reads undefined, as Node reads it (#24kvb6k).
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.count) + ' ' + String(viewed.value.extra));
