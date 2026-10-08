interface Base { readonly kind: 'items' | 'other'; }
interface Entry { readonly value: number; readonly text: string; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
function entry(value: number): Entry { return {value, text: `item${value}`}; }
const source: Entry[] = [];
const raw = {kind: 'items' as const, values: source};
const values = items(raw).values;
values.push(entry(2));
console.log(`${values.length}:${values[0]!.text}`);
