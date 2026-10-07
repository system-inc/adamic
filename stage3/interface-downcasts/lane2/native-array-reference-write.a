interface Base { readonly kind: 'items' | 'other'; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
interface Entry { readonly value: number; }
const raw = {kind: 'items' as const, values: [{value: 7, secret: 99}]};
const values = items(raw).values;
values[0] = {value: 8};
console.log(`${raw.values[0]!.secret}`);
