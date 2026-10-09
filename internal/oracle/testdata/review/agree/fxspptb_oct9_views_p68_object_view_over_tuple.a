// A good view whose field is an object type with index-named fields, over a real tuple.
interface Pair { readonly '0': string; readonly '1': number; }
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: Pair; }
const pair: [string, number] = ['first', 1];
const raw = { kind: 'Holder' as const, value: pair };
const base: Root = raw;
const viewed = base as View;
const value = viewed.value;
console.log(`${value[0]}:${value[1]}`);
