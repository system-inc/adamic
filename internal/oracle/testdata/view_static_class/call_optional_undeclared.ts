// #s03ndgw, a stop: Inner's optional call is a name Function gives every constructor, and Klass doesn't declare
// it, so Node reads Function's call where an absent field would read undefined. Lowering admits no class for
// such a contract (StaticsAdmitted stays false), so the read of viewed.value stops.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; readonly call?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.count) + ' ' + typeof viewed.value.call);
