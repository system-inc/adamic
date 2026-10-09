// A good untagged union view whose shared member is stored by its producer as string | number.
interface Left { readonly name: string; readonly text: string; }
interface Right { readonly name: string; readonly count: number; }
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: Left | Right; }
const inner = { name: 'abc' as string | number, text: 'b' };
const raw = { kind: 'Holder' as const, value: inner };
const base: Root = raw;
const value = (base as View).value;
console.log(`${value.name.length}`);
inner.name = 9;
console.log(`${value.name.length}`);
