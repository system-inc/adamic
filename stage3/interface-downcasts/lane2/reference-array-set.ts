interface Base { readonly kind: 'items' | 'other'; }
interface Entry { readonly value: number; readonly text: string; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
function entry(value: number): Entry { return {value, text: `item${value}`}; }
const raw = {kind: 'items' as const, values: [entry(1)]};
const values = items(raw).values;
values[0] = entry(2);
console.log(`${raw.values[0]!.value}:${values[0]!.text}`);
