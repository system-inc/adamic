interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: boolean; }
const pair: [number, number] = [7, 1];
const raw = {kind: 'Holder' as const, value: pair};
const base: Root = raw;
const viewed = base as View;
console.log(String(viewed.value));
