interface Base { readonly kind: 'items' | 'other'; }
interface Entry { readonly value: number; readonly text: string; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
function entry(value: number): Entry { return {value, text: `item${value}`}; }
const raw = {kind: 'items' as const, values: [{value: 1, text: `item${1}`}]};
const incoming = {value: 2, text: `item${2}`};
const values = items(raw).values;
values[0] = incoming;
incoming.value = 3;
console.log(`${raw.values[0]!.value}:${values[0]!.text}`);
values[0] = {value: 4, text: `item${4}`};
console.log(`${incoming.value}:${incoming.text}:${values[0]!.value}`);
