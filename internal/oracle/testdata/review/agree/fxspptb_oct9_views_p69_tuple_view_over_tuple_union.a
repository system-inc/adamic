// A fixed-tuple view in an untagged union with an object member.
interface Named { readonly text: string; }
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number] | Named; }
const pair: [string, number] = ['first', 1];
const raw = { kind: 'Holder' as const, value: pair };
const base: Root = raw;
const viewed = base as View;
const value = viewed.value;
console.log(`${value === pair}`);
