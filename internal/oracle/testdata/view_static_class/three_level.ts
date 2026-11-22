// #s03ndgw: C owns its tag and inherits count through B, which declares no static, from A, so the read walks
// two parents, as Node walks C's prototype chain, and reads A's 3.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class A {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
class B extends A {}
class C extends B {
	static override readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: C };
const base: Root = raw;
const viewed = base as Viewed;
console.log(String(viewed.value.count));
