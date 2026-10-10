interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number]; }
const pair: [number, number] = [7, 1];
const raw = { kind: 'Holder' as const, value: pair };
const base: Root = raw;
const viewed = base as View;
console.log(viewed.value[0]);
