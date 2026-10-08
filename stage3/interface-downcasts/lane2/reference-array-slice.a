interface Base { readonly kind: 'items' | 'other'; }
interface Entry { readonly value: number; readonly text: string; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
function entry(value: number): Entry { return {value, text: `item${value}`}; }
const raw = {kind: 'items' as const, values: [entry(1), entry(2)]};
const values = items(raw).values.slice(1);
values[0] = entry(3);
values.push(entry(4));
console.log(`${values.length}:${values[0]!.text}:${raw.values[1]!.text}`);
