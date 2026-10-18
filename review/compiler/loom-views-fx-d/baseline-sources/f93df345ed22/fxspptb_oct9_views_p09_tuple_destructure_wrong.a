// A wrong fixed-tuple view read by array destructuring.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number]; }
const pair: [number, number] = [7, 1];
const raw = { kind: 'Holder' as const, value: pair };
const base: Root = raw;
const viewed = base as View;
const [first, second] = viewed.value;
console.log(`${first.length} ${second}`);
