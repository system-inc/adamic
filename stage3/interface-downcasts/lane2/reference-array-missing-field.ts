interface Base { readonly kind: 'items' | 'other'; }
interface Entry { readonly value: number; readonly text: string; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
function entry(value: number): Entry { return {value, text: `item${value}`}; }
const raw = {kind: 'items' as const, values: [{value: 1, text: `item${1}`, secret: 99}]};
const values = items(raw).values;
values[0] = entry(2);
console.log('written');
console.log(`${raw.values[0]!.secret}`);
